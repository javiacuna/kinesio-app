package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/exerciseplans/domain"
	"github.com/javiacuna/kinesio-backend/internal/exerciseplans/usecase"
)

type Handler struct {
	createUC *usecase.CreatePlanUseCase
	listUC   *usecase.ListPlansByPatientUseCase
	getUC    *usecase.GetPlanByIDUseCase
}

func NewHandler(createUC *usecase.CreatePlanUseCase, listUC *usecase.ListPlansByPatientUseCase, getUC *usecase.GetPlanByIDUseCase) *Handler {
	return &Handler{createUC: createUC, listUC: listUC, getUC: getUC}
}

type createPlanRequest struct {
	KinesiologistID string                        `json:"kinesiologist_id"`
	Frequency       string                        `json:"frequency"`
	DurationWeeks   int                           `json:"duration_weeks"`
	Observations    *string                       `json:"observations"`
	Items           []usecase.CreatePlanItemInput `json:"items"`
}

func (h *Handler) CreateForPatient(c *gin.Context) {
	patientID := c.Param("patient_id")

	var req createPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	out, validation, err := h.createUC.Execute(c.Request.Context(), usecase.CreatePlanInput{
		PatientID:       patientID,
		KinesiologistID: req.KinesiologistID,
		Frequency:       req.Frequency,
		DurationWeeks:   req.DurationWeeks,
		Observations:    req.Observations,
		Items:           req.Items,
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
	items, err := h.listUC.Execute(c.Request.Context(), pid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	out := make([]planResponse, 0, len(items))
	for _, p := range items {
		out = append(out, toResponse(p))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("plan_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_plan_id"})
		return
	}
	p, found, err := h.getUC.Execute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, toResponse(p))
}
