package services

import (
	"context"
	"fmt"

	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
	"github.com/jamesphm04/splose-clone-be/internal/repositories"
	"go.uber.org/zap"
)

type CreateMessageInput struct {
	ConversationID string
	Role           string
	Content        string
}

type ListByConversationIDInput struct {
	ConversationID string
}

type MessageService struct {
	repo repositories.MessageRepository
	log  *zap.Logger
}

func NewMessageService(repo repositories.MessageRepository, log *zap.Logger) *MessageService {
	return &MessageService{
		repo: repo,
		log:  log.Named("NewMessageService"),
	}
}

func (s *MessageService) Create(ctx context.Context, in CreateMessageInput) (*entities.Message, error) {
	msg := &entities.Message{
		ConversationID: in.ConversationID,
		Role:           entities.MessageRole(in.Role),
		Content:        in.Content,
	}

	if err := s.repo.Create(ctx, msg); err != nil {
		s.log.Error("message creation failed", zap.Error(err))
		return nil, fmt.Errorf("creating message: %w", err)
	}

	s.log.Info("message created")
	return msg, nil
}

func (s *MessageService) ListByConversationID(ctx context.Context, in ListByConversationIDInput) ([]entities.Message, int, error) {
	msges, err := s.repo.FindByConversationID(ctx, in.ConversationID)
	if err != nil {
		s.log.Error("messages list failed", zap.String("conversationID", in.ConversationID), zap.Error(err))
		return nil, 0, fmt.Errorf("listing messages: %w", err)
	}
	return msges, len(msges), nil
}
