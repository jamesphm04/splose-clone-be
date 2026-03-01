package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
	"github.com/jamesphm04/splose-clone-be/internal/repositories"
	"github.com/jamesphm04/splose-clone-be/pkg/storage"
	"go.uber.org/zap"
)

type FileUploadInput struct {
	NoteID     string                // FK → notes.id
	MessageID  string                // FK → messages.id
	File       multipart.File        // open file handle (caller must close)
	FileHeader *multipart.FileHeader // carries Name, Size, Header (MIME)
}

type AttachmentService struct {
	repo     repositories.AttachmentRepository
	s3Client *storage.Client
	log      *zap.Logger
}

func NewAttachmentService(
	repo repositories.AttachmentRepository,
	s3Client *storage.Client,
	log *zap.Logger,
) *AttachmentService {
	return &AttachmentService{
		repo:     repo,
		s3Client: s3Client,
		log:      log.Named("attachment_service"),
	}
}

func (s *AttachmentService) Create(ctx context.Context, in FileUploadInput) (*entities.Attachment, string, error) {
	s.log.Info("creating attachmenttttt", zap.String("input", fmt.Sprintf("%+v", in)))
	// S3
	safeName := filepath.Base(in.FileHeader.Filename)
	safeName = strings.ReplaceAll(safeName, " ", "_")

	s3Key := fmt.Sprintf("attachments/%s/%d_%s",
		in.NoteID,
		time.Now().UnixMilli(),
		safeName,
	)

	// Detect MIME type
	// Prefer the Content-Type the client sent; fall back to octet-stream.
	contentType := in.FileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	s.log.Info("uploading attachment to S3",
		zap.String("key", s3Key),
		zap.String("contentType", contentType),
		zap.Int64("size", in.FileHeader.Size),
	)

	// Upload to S3
	uploadOut, err := s.s3Client.Upload(ctx, storage.UploadInput{
		Key:         s3Key,
		Body:        in.File,
		ContentType: contentType,
		Size:        in.FileHeader.Size,
	})
	if err != nil {
		return nil, "", fmt.Errorf("uploading attachment to S3: %w", err)
	}

	// Save to DB
	att := &entities.Attachment{
		NoteID:    in.NoteID,
		MessageID: in.MessageID,
		URL:       uploadOut.URL,
		Name:      in.FileHeader.Filename, // keep original display name
		Type:      contentType,
		Size:      in.FileHeader.Size,
		S3Key:     s3Key, // stored so we can delete later
	}

	if err := s.repo.Create(ctx, att); err != nil {
		// DB write failed after a successful S3 upload.
		// Attempt to clean up the orphaned S3 object.
		s.log.Error("DB write failed after S3 upload – attempting S3 rollback",
			zap.String("key", s3Key),
			zap.Error(err),
		)
		if delErr := s.s3Client.Delete(ctx, s3Key); delErr != nil {
			s.log.Error("S3 rollback also failed – orphaned object",
				zap.String("key", s3Key),
				zap.Error(delErr),
			)
		}
		return nil, "", fmt.Errorf("saving attachment metadata: %w", err)
	}

	s.log.Info("attachment uploaded and recorded",
		zap.String("attachmentID", att.ID),
		zap.String("messageID", att.MessageID),
		zap.String("s3Key", s3Key),
		zap.Int64("size", att.Size),
	)
	return att, uploadOut.PresignedURL, nil
}
