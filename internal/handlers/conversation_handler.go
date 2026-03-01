package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
	"github.com/jamesphm04/splose-clone-be/internal/services"
	"github.com/jamesphm04/splose-clone-be/internal/utils"
	"go.uber.org/zap"
)

type SendMessageResponse struct {
	Message      string `json:"message"`
	PresignedURL string `json:"presignedURL"`
}

type ConversationHandler struct {
	convSvc       *services.ConversationService
	messageSvc    *services.MessageService
	attachmentSvc *services.AttachmentService
	validate      *validator.Validate
	log           *zap.Logger
}

func NewConversationHandler(
	convSvc *services.ConversationService,
	messageSvc *services.MessageService,
	attachmentSvc *services.AttachmentService,
	log *zap.Logger,
) *ConversationHandler {
	v := validator.New()
	return &ConversationHandler{
		convSvc:       convSvc,
		messageSvc:    messageSvc,
		attachmentSvc: attachmentSvc,
		validate:      v,
		log:           log.Named("conversation_handler"),
	}
}

// Create POST /api/v1/conversations/send-message
func (h *ConversationHandler) SendMessage(c *gin.Context) {
	noteID := c.PostForm("noteID")
	message := c.PostForm("message")

	if noteID == "" {
		utils.BadRequest(c, "noteID is required")
		return
	}

	conversation, err := h.convSvc.GetByNoteID(c.Request.Context(), noteID)
	if err != nil {
		utils.BadRequest(c, fmt.Sprintf("failed to get conversation: %v", err))
		return
	}

	// Save user message
	userMsgIn := services.CreateMessageInput{
		ConversationID: conversation.ID,
		Role:           string(entities.RoleUser),
		Content:        message,
	}

	userMsg, err := h.messageSvc.Create(c.Request.Context(), userMsgIn)
	if err != nil {
		utils.BadRequest(c, fmt.Sprintf("failed to save message: %v", err))
		return
	}

	h.log.Info("user message created", zap.String("messageID", userMsg.ID))

	// ─── Optional Attachment ──────────────────────────────

	var presignedURL string

	file, header, err := c.Request.FormFile("attachment")
	if err != nil && err != http.ErrMissingFile {
		utils.BadRequest(c, fmt.Sprintf("failed to parse attachment: %v", err))
		return
	}

	if err == nil && file != nil && header != nil {
		createAttIn := services.FileUploadInput{
			NoteID:     noteID,
			MessageID:  userMsg.ID,
			File:       file,
			FileHeader: header,
		}

		h.log.Info("creating attachment", zap.String("filename", header.Filename))

		_, presignedURL, err = h.attachmentSvc.Create(c.Request.Context(), createAttIn)
		if err != nil {
			utils.BadRequest(c, fmt.Sprintf("failed to create attachment: %v", err))
			return
		}
	}

	// ─── MOCK AI Response ─────────────────────────────────

	assistantMsgIn := services.CreateMessageInput{
		ConversationID: conversation.ID,
		Role:           string(entities.RoleAssistant),
		Content:        "This is a mock AI response",
	}

	assistantMsg, err := h.messageSvc.Create(c.Request.Context(), assistantMsgIn)
	if err != nil {
		utils.BadRequest(c, fmt.Sprintf("failed to save message: %v", err))
		return
	}

	utils.OK(c, SendMessageResponse{
		Message:      assistantMsg.Content,
		PresignedURL: presignedURL, // empty string if no attachment
	})
}
