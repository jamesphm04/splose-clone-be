package services

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/jamesphm04/splose-clone-be/internal/clients"
	"github.com/jamesphm04/splose-clone-be/internal/models/dtos/open_ai_client"
	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
	"github.com/jamesphm04/splose-clone-be/internal/repositories"
	"go.uber.org/zap"
)

type CreateConversationInput struct {
	NoteID string
}

type SendMessageInput struct {
	NoteID     string
	Message    string
	File       multipart.File
	FileHeader *multipart.FileHeader
}

type ConversationService struct {
	repo          repositories.ConversationRepository
	client        *clients.SploseCloneAIClient
	attachmentSvc *AttachmentService
	messageSvc    *MessageService
	noteSvc       *NoteService
	patientSvc    *PatientService
	log           *zap.Logger
}

func NewConversationService(
	repo repositories.ConversationRepository,
	client *clients.SploseCloneAIClient,
	messageSvc *MessageService,
	noteSvc *NoteService,
	patientSvc *PatientService,
	attachmentSvc *AttachmentService,
	log *zap.Logger,
) *ConversationService {
	return &ConversationService{
		repo:          repo,
		client:        client,
		messageSvc:    messageSvc,
		noteSvc:       noteSvc,
		patientSvc:    patientSvc,
		attachmentSvc: attachmentSvc,
		log:           log.Named("conversation-service"),
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

func (s *ConversationService) GetByNoteID(ctx context.Context, noteID string, offset *int, limit *int) (*entities.Conversation, error) {
	conv, err := s.repo.FindByNoteID(ctx, noteID, offset, limit)
	if err != nil {
		s.log.Error("conversation retrieval failed", zap.String("noteID", noteID), zap.Error(err))
		return nil, fmt.Errorf("retrieving conversation: %w", err)
	}
	return conv, nil
}

func buildAIConversation(messages []entities.Message) []open_ai_client.Message {
	result := make([]open_ai_client.Message, 0, len(messages))

	for _, m := range messages {
		aiMsg := open_ai_client.Message{
			Role:    string(m.Role),
			Content: m.Content,
		}

		if len(m.Attachments) > 0 {
			attachments := make([]open_ai_client.Attachment, 0, len(m.Attachments))
			for _, a := range m.Attachments {
				attachments = append(attachments, open_ai_client.Attachment{
					Name: a.Name,
					Type: a.Type,
					URL:  a.URL,
				})
			}
			aiMsg.Attachments = attachments
		}

		result = append(result, aiMsg)
	}

	return result
}

func (s *ConversationService) SendMessage(ctx context.Context, in SendMessageInput) (*entities.Message, error) {
	var presignedURL string

	offset := 1
	limit := 0
	currentConversation, err := s.GetByNoteID(ctx, in.NoteID, &offset, &limit)
	if err != nil {
		return nil, fmt.Errorf("retrieving conversation: %w", err)
	}

	note, err := s.noteSvc.GetByID(ctx, in.NoteID)
	if err != nil {
		return nil, fmt.Errorf("retrieving note: %w", err)
	}

	// Save user message
	userMsgIn := CreateMessageInput{
		ConversationID: currentConversation.ID,
		Role:           string(entities.RoleUser),
		Content:        in.Message,
	}
	userMsg, err := s.messageSvc.Create(ctx, userMsgIn)
	if err != nil {
		return nil, fmt.Errorf("creating user message: %w", err)
	}

	s.log.Info("user message created", zap.String("messageID", userMsg.ID))

	// Save attachment
	if in.File != nil && in.FileHeader != nil {
		s.log.Info("saving attachment", zap.String("filename", in.FileHeader.Filename))
		attachmentIn := FileUploadInput{
			NoteID:     in.NoteID,
			MessageID:  userMsg.ID,
			File:       in.File,
			FileHeader: in.FileHeader,
		}

		_, _, err = s.attachmentSvc.Create(ctx, attachmentIn)
		if err != nil {
			return nil, fmt.Errorf("creating attachment: %w", err)
		}
		s.log.Info("attachment saved", zap.String("presignedURL", presignedURL))
	}

	// Send message to AI
	// build conversation context

	// Fetch the 20 lastest after the user message
	offset = 0
	limit = 20
	conversation, err := s.GetByNoteID(ctx, in.NoteID, &offset, &limit)
	if err != nil {
		return nil, fmt.Errorf("retrieving updated conversation: %w", err)
	}
	// get patient
	patient, err := s.patientSvc.GetByID(ctx, conversation.Note.PatientID)
	if err != nil {
		return nil, fmt.Errorf("retrieving patient: %w", err)
	}

	aiConversation := buildAIConversation(conversation.Messages)
	req := &open_ai_client.SendMessageRequest{
		ConversationContext: open_ai_client.ConversationContext{
			Patient: open_ai_client.Patient{
				Email:       patient.Email,
				FirstName:   patient.FirstName,
				LastName:    patient.LastName,
				PhoneNumber: patient.PhoneNumber,
				DateOfBirth: patient.DateOfBirth,
				Gender:      patient.Gender,
				FullAddress: patient.FullAddress,
			},
			Note: open_ai_client.Note{
				Title:   note.Title,
				Content: note.Content,
			},
			Conversation: aiConversation,
		},
		Message: in.Message,
	}

	responseMsg, err := s.client.SendMessage(ctx, *req)
	if err != nil {
		return nil, fmt.Errorf("sending message to AI: %w", err)
	}

	// Save assistant message
	assistantMsgIn := CreateMessageInput{
		ConversationID: conversation.ID,
		Role:           string(entities.RoleAssistant),
		Content:        responseMsg,
	}

	assistantMsg, err := s.messageSvc.Create(ctx, assistantMsgIn)
	if err != nil {
		return nil, fmt.Errorf("creating assistant message: %w", err)
	}

	s.log.Info("assistant message created", zap.String("messageID", assistantMsg.ID))

	return assistantMsg, nil
}
