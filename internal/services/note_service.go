package services

import (
	"context"
	"fmt"

	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
	"github.com/jamesphm04/splose-clone-be/internal/repositories"
	"go.uber.org/zap"
)

type CreateNoteInput struct {
	PatientID string `json:"patientId" validate:"required,uuid"`
	UserID    string `json:"userId" validate:"required,uuid"`
	Title     string `json:"title"`
	Content   string `json:"content" validate:"required"`
}

type UpdateNoteInput struct {
	UserID  *string `json:"userId"`
	Title   *string `json:"title"`
	Content *string `json:"content"`
}

// Service

type NoteService struct {
	repo repositories.NoteRepository
	log  *zap.Logger
}

func NewNoteService(repo repositories.NoteRepository, log *zap.Logger) *NoteService {
	return &NoteService{
		repo: repo,
		log:  log.Named("note-service"),
	}
}

func (s *NoteService) Create(ctx context.Context, in CreateNoteInput) (*entities.Note, error) {
	note := &entities.Note{
		PatientID: in.PatientID,
		UserID:    in.UserID,
		Title:     in.Title,
		Content:   in.Content,
	}

	if err := s.repo.Create(ctx, note); err != nil {
		s.log.Error("note creation failed", zap.Error(err))
		return nil, fmt.Errorf("creating note: %w", err)
	}

	s.log.Info("note created", zap.String("title", in.Title))
	return note, nil
}

func (s *NoteService) GetByID(ctx context.Context, id string) (*entities.Note, error) {
	note, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.log.Error("note retrieval failed", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("retrieving note: %w", err)
	}

	return note, nil
}

func (s *NoteService) ListByPatientID(ctx context.Context, patientID string) ([]entities.Note, error) {
	notes, err := s.repo.FindByPatientID(ctx, patientID)
	if err != nil {
		s.log.Error("notes list failed", zap.String("patientID", patientID), zap.Error(err))
		return nil, fmt.Errorf("listing notes: %w", err)
	}
	return notes, nil
}

func (s *NoteService) List(ctx context.Context, offset, limit int) ([]entities.Note, int64, error) {
	notes, total, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		s.log.Error("notes list failed", zap.Error(err))
		return nil, 0, fmt.Errorf("listing notes: %w", err)
	}

	return notes, total, nil
}

func (s *NoteService) Update(ctx context.Context, id string, in UpdateNoteInput) (*entities.Note, error) {
	note, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.log.Error("note update failed", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("updating note: %w", err)
	}

	if in.UserID != nil {
		note.UserID = *in.UserID
	}

	if in.Title != nil {
		note.Title = *in.Title
	}

	if in.Content != nil {
		note.Content = *in.Content
	}

	if err := s.repo.Update(ctx, note); err != nil {
		s.log.Error("note update failed", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("updating note: %w", err)
	}

	s.log.Info("note updated", zap.String("id", id))
	return note, nil
}

func (s *NoteService) SoftDelete(ctx context.Context, id string) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		s.log.Error("note soft delete failed", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("soft deleting note: %w", err)
	}

	s.log.Info("note soft deleted", zap.String("id", id))
	return nil
}
