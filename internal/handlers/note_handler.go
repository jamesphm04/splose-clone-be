package handlers

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/jamesphm04/splose-clone-be/internal/services"
	"github.com/jamesphm04/splose-clone-be/internal/utils"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type CreateNoteResponse struct {
	NoteID         string `json:"noteId"`
	ConversationID string `json:"conversationId"`
}

type NoteHandler struct {
	noteSvc  *services.NoteService
	convSvc  *services.ConversationService
	validate *validator.Validate
	log      *zap.Logger
}

func NewNoteHandler(noteSvc *services.NoteService, convSvc *services.ConversationService, log *zap.Logger) *NoteHandler {
	v := validator.New()
	return &NoteHandler{
		noteSvc:  noteSvc,
		convSvc:  convSvc,
		validate: v,
		log:      log.Named("patient_handler"),
	}
}

// Create POST /api/v1/notes
func (h *NoteHandler) Create(c *gin.Context) {
	var in services.CreateNoteInput

	// must bind the request body
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, "invalid request body")
		return
	}

	h.log.Info("note create request", zap.String("request", fmt.Sprintf("%+v", in)))
	// then validate the request body
	if err := h.validate.Struct(in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// create the note
	note, err := h.noteSvc.Create(c.Request.Context(), in)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// create the conversation
	conv, err := h.convSvc.Create(c.Request.Context(), services.CreateConversationInput{
		NoteID: note.ID,
	})

	res := CreateNoteResponse{
		NoteID:         note.ID,
		ConversationID: conv.ID,
	}

	utils.Created(c, res)
}

// List GET /api/v1/notes
func (h *NoteHandler) List(c *gin.Context) {
	page, pageSize, offset := utils.Pagination(c)
	notes, total, err := h.noteSvc.List(c.Request.Context(), offset, pageSize)
	if err != nil {
		utils.InternalError(c)
	}
	utils.OKList(c, notes, utils.BuildMeta(page, pageSize, total))
}

// List GET /api/v1/notes/:id
func (h *NoteHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	note, err := h.noteSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.NotFound(c, "patient")
		return
	}
	utils.OK(c, note)
}

// ListByPatientID GET /api/v1/notes/patient/:patientID
func (h *NoteHandler) ListByPatientID(c *gin.Context) {
	patientID := c.Param("patientID")

	notes, err := h.noteSvc.ListByPatientID(c.Request.Context(), patientID)
	if err != nil {
		utils.NotFound(c, "patient")
		return
	}
	utils.OKList(c, notes, nil)
}

// Update PATCH /api/v1/notes/:id
func (h *NoteHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var in services.UpdateNoteInput

	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, "invalid request body")
		return
	}
	if err := h.validate.Struct(in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	note, err := h.noteSvc.Update(c.Request.Context(), id, in)
	if err != nil {
		utils.BadRequest(c, err.Error())
	}
	utils.OK(c, note)
}

// Delete DELETE /api/v1/notes/:id
func (h *NoteHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.noteSvc.SoftDelete(c.Request.Context(), id); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, gin.H{"message": "note deleted"})
}
