package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/jamesphm04/splose-clone-be/internal/services"
	"github.com/jamesphm04/splose-clone-be/internal/utils"
)

// PatientHandler handles patient management endpoints.
type PatientHandler struct {
	patientSvc *services.PatientService
	validate   *validator.Validate
	log        *zap.Logger
}

func NewPatientHandler(patientSvc *services.PatientService, log *zap.Logger) *PatientHandler {
	v := validator.New()
	utils.RegisterCustomValidators(v)
	return &PatientHandler{
		patientSvc: patientSvc,
		validate:   v,
		log:        log.Named("patient_handler"),
	}
}

// Create  POST /api/v1/patients
func (h *PatientHandler) Create(c *gin.Context) {
	var in services.CreatePatientInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, "invalid request body")
		return
	}

	if err := h.validate.Struct(in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	patient, err := h.patientSvc.Create(c.Request.Context(), in)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Created(c, patient)
}

// GetByID  GET /api/v1/patients/:id
func (h *PatientHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	patient, err := h.patientSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.NotFound(c, "patient")
		return
	}
	utils.OK(c, patient)
}

// List  GET /api/v1/patients
func (h *PatientHandler) List(c *gin.Context) {
	page, pageSize, offset := utils.Pagination(c)
	patients, total, err := h.patientSvc.List(c.Request.Context(), offset, pageSize)
	if err != nil {
		utils.InternalError(c)
		return
	}
	utils.OKList(c, patients, utils.BuildMeta(page, pageSize, total))
}
