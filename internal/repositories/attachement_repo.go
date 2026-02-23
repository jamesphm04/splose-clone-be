package repositories

import (
	"context"
	"errors"

	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AttachmentRepository interface {
	Create(ctx context.Context, attachment *entities.Attachment) error
}

type attachmentRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewAttachmentRepository returns a GORM-backed AttachmentRepository.
func NewAttachmentRepository(db *gorm.DB, log *zap.Logger) AttachmentRepository {
	return &attachmentRepo{
		db:  db,
		log: log.Named("attachment-repository"),
	}
}

func (r *attachmentRepo) Create(ctx context.Context, attachment *entities.Attachment) error {
	if err := r.db.WithContext(ctx).Create(attachment).Error; err != nil {
		r.log.Error("failed to create attachment", zap.String("attachmentID", attachment.ID), zap.Error(err))
		return err
	}

	r.log.Info("attachment created", zap.String("attachmentID", attachment.ID))
	return nil
}

func (r *attachmentRepo) FindByID(ctx context.Context, id string) (*entities.Attachment, error) {
	var a entities.Attachment
	err := r.db.WithContext(ctx).First(&a, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		r.log.Error("FindByID failed", zap.String("id", id), zap.Error(err))
	}
	return &a, nil
}

func (r *attachmentRepo) List(ctx context.Context, offset, limit int) ([]entities.Attachment, int64, error) {
	var attachments []entities.Attachment
	var total int64

	// count total
	if err := r.db.WithContext(ctx).Model(&entities.Attachment{}).Count(&total).Error; err != nil {
		r.log.Error("List count failed", zap.Error(err))
		return nil, 0, err
	}

	// list
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&attachments).Error; err != nil {
		r.log.Error("List query failed", zap.Error(err))
		return nil, 0, err
	}

	return attachments, total, nil
}

func (r *attachmentRepo) Update(ctx context.Context, attachment *entities.Attachment) error {
	if err := r.db.WithContext(ctx).Save(attachment).Error; err != nil {
		r.log.Error("Update failed", zap.String("attachmentID", attachment.ID), zap.Error(err))
		return err
	}

	return nil
}

func (r *attachmentRepo) SoftDelete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&entities.Attachment{}, "id = ?", id)
	if res.Error != nil {
		r.log.Error("SoftDelete failed", zap.String("attachmentID", id), zap.Error(res.Error))
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	r.log.Info("attachment soft-deleted", zap.String("attachmentID", id))
	return nil
}
