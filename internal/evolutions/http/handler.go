package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/evolutions/domain"
	"github.com/javiacuna/kinesio-backend/internal/evolutions/usecase"
)

type Handler struct {
	createUC *usecase.CreateEvolutionUseCase
	listUC   *usecase.ListEvolutionsByPatientUseCase
	getUC    *usecase.GetEvolutionByIDUseCase
}

func NewHandler(createUC *usecase.CreateEvolutionUseCase, listUC *usecase.ListEvolutionsByPatientUseCase, getUC *usecase.GetEvolutionByIDUseCase) *Handler {
	return &Handler{createUC: createUC, listUC: listUC, getUC: getUC}
}

type createEvolutionRequest struct {
	KinesiologistID string  `json:"kinesiologist_id"`
	AppointmentID   *string `json:"appointment_id,omitempty"`
	PainLevel       *int    `json:"pain_level,omitempty"`
	Notes           string  `json:"notes"`
}

type evolutionResponse struct {
	ID              string  `json:"id"`
	PatientID       string  `json:"patient_id"`
	KinesiologistID string  `json:"kinesiologist_id"`
	AppointmentID   *string `json:"appointment_id,omitempty"`
	PainLevel       *int    `json:"pain_level,omitempty"`
	Notes           string  `json:"notes"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

func (h *Handler) CreateForPatient(c *gin.Context) {
	patientID := c.Param("patient_id")

	var req createEvolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	out, validation, err := h.createUC.Execute(c.Request.Context(), usecase.CreateEvolutionInput{
		PatientID:       patientID,
		KinesiologistID: req.KinesiologistID,
		AppointmentID:   req.AppointmentID,
		PainLevel:       req.PainLevel,
		Notes:           req.Notes,
	})

	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(out))
}

func (h *Handler) ListByPatient(c *gin.Context) {
	pid, err := uuid.Parse(c.Param("patient_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_patient_id"})
		return
	}

	limit := 50
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			limit = n
		}
	}

	items, err := h.listUC.Execute(c.Request.Context(), pid, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]evolutionResponse, 0, len(items))
	for _, e := range items {
		out = append(out, toResponse(e))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("evolution_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_evolution_id"})
		return
	}

	e, found, err := h.getUC.Execute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}

	c.JSON(http.StatusOK, toResponse(e))
}

func toResponse(e domain.PatientEvolution) evolutionResponse {
	var appt *string
	if e.AppointmentID != nil {
		s := e.AppointmentID.String()
		appt = &s
	}

	return evolutionResponse{
		ID:              e.ID.String(),
		PatientID:       e.PatientID.String(),
		KinesiologistID: e.KinesiologistID.String(),
		AppointmentID:   appt,
		PainLevel:       e.PainLevel,
		Notes:           e.Notes,
		CreatedAt:       e.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       e.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
