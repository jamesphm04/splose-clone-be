package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jamesphm04/splose-clone-be/internal/models/dtos"
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

	// ─── Optional Attachment ──────────────────────────────
	file, header, err := c.Request.FormFile("attachment")
	if err != nil && err != http.ErrMissingFile {
		utils.BadRequest(c, fmt.Sprintf("failed to parse attachment: %v", err))
		return
	}

	assistantMsg, err := h.convSvc.SendMessage(c.Request.Context(), services.SendMessageInput{
		NoteID:     noteID,
		Message:    message,
		File:       file,
		FileHeader: header,
	})
	if err != nil {
		utils.BadRequest(c, fmt.Sprintf("failed to send message: %v", err))
		return
	}

	utils.OK(c, dtos.ToDTO(assistantMsg))
}

// ListMessages GET /api/v1/conversations/messages?noteID=xxx
func (h *ConversationHandler) ListMessagesByNoteID(c *gin.Context) {
	noteID := c.Query("noteID")

	if noteID == "" {
		utils.BadRequest(c, "noteID is required")
		return
	}

	messages, err := h.messageSvc.ListByNoteID(c.Request.Context(), noteID)
	if err != nil {
		utils.BadRequest(c, fmt.Sprintf("failed to get messages: %v", err))
		return
	}

	utils.OK(c, messages)
}
