package services

import (
	"context"
	"fmt"

	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
	"github.com/jamesphm04/splose-clone-be/internal/repositories"
	"go.uber.org/zap"
)

type CreateConversationInput struct {
	NoteID string
}

type ConversationService struct {
	repo repositories.ConversationRepository
	log  *zap.Logger
}

func NewConversationService(repo repositories.ConversationRepository, log *zap.Logger) *ConversationService {
	return &ConversationService{
		repo: repo,
		log:  log.Named("conversation-service"),
	}
}

func (s *ConversationService) Create(ctx context.Context, in CreateConversationInput) (*entities.Conversation, error) {
	conv := &entities.Conversation{
		NoteID: in.NoteID,
	}

	if err := s.repo.Create(ctx, conv); err != nil {
		s.log.Error("conversation creatation failed", zap.Error(err))
		return nil, fmt.Errorf("creating conversation: %w", err)
	}

	s.log.Info("conversation created")
	return conv, nil
}

func (s *ConversationService) GetByNoteID(ctx context.Context, noteID string) (*entities.Conversation, error) {
	conv, err := s.repo.FindByNoteID(ctx, noteID)
	if err != nil {
		s.log.Error("conversation retrieval failed", zap.String("noteID", noteID), zap.Error(err))
		return nil, fmt.Errorf("retrieving conversation: %w", err)
	}
	return conv, nil
}
