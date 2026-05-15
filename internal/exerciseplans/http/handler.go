package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/exerciseplans/domain"
	"github.com/javiacuna/kinesio-backend/internal/exerciseplans/usecase"
	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	patientDomain "github.com/javiacuna/kinesio-backend/internal/patients/domain"
)

type Handler struct {
	createUC *usecase.CreatePlanUseCase
	listUC   *usecase.ListPlansByPatientUseCase
	getUC    *usecase.GetPlanByIDUseCase
	updateUC *usecase.UpdatePlanUseCase
	patients patientSearcher
}

type patientSearcher interface {
	Search(ctx context.Context, query string, limit int, includeInactive bool) ([]patientDomain.Patient, error)
}

func NewHandler(createUC *usecase.CreatePlanUseCase, listUC *usecase.ListPlansByPatientUseCase, getUC *usecase.GetPlanByIDUseCase, lookups ...any) *Handler {
	var updateUC *usecase.UpdatePlanUseCase
	var patientRepo patientSearcher
	for _, lookup := range lookups {
		switch repo := lookup.(type) {
		case *usecase.UpdatePlanUseCase:
			updateUC = repo
		case patientSearcher:
			patientRepo = repo
		}
	}

	return &Handler{createUC: createUC, listUC: listUC, getUC: getUC, updateUC: updateUC, patients: patientRepo}
}

type createPlanRequest struct {
	PatientID       string                        `json:"patient_id"`
	KinesiologistID string                        `json:"kinesiologist_id"`
	DiagnosisID     *string                       `json:"patient_diagnosis_id,omitempty"`
	Frequency       string                        `json:"frequency"`
	DurationWeeks   int                           `json:"duration_weeks"`
	Observations    *string                       `json:"observations"`
	Items           []usecase.CreatePlanItemInput `json:"items"`
}

type updatePlanRequest struct {
	KinesiologistID string                        `json:"kinesiologist_id"`
	DiagnosisID     *string                       `json:"patient_diagnosis_id,omitempty"`
	Frequency       string                        `json:"frequency"`
	DurationWeeks   int                           `json:"duration_weeks"`
	Observations    *string                       `json:"observations"`
	Status          string                        `json:"status"`
	Items           []usecase.CreatePlanItemInput `json:"items"`
}

func (h *Handler) Create(c *gin.Context) {
	if !middleware.HasRole(c, "kinesiologo") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req createPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	h.create(c, req.PatientID, req)
}

func (h *Handler) CreateForPatient(c *gin.Context) {
	if !middleware.HasRole(c, "kinesiologo") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	patientID := c.Param("patient_id")

	var req createPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	h.create(c, patientID, req)
}

func (h *Handler) create(c *gin.Context, patientID string, req createPlanRequest) {
	out, validation, err := h.createUC.Execute(c.Request.Context(), usecase.CreatePlanInput{
		PatientID:       patientID,
		KinesiologistID: req.KinesiologistID,
		DiagnosisID:     req.DiagnosisID,
		Frequency:       req.Frequency,
		DurationWeeks:   req.DurationWeeks,
		Observations:    req.Observations,
		Items:           req.Items,
		ActorEmail:      actorEmail(c),
		ActorRole:       actorRole(c),
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
	patientID := c.Param("patient_id")
	if patientID == "" {
		patientID = c.Param("id")
	}
	patientID = strings.TrimSpace(patientID)

	if strings.EqualFold(patientID, "me") || h.isCurrentPatient(c) {
		resolvedPatientID, ok := h.patientIDForCurrentPatient(c, patientID)
		if !ok {
			return
		}
		patientID = resolvedPatientID
	} else if !middleware.HasRole(c, "recepcionista", "kinesiologo") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	pid, err := uuid.Parse(patientID)
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

func (h *Handler) Update(c *gin.Context) {
	if !middleware.HasRole(c, "kinesiologo") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if h.updateUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	var req updatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	out, validation, err := h.updateUC.Execute(c.Request.Context(), c.Param("plan_id"), usecase.UpdatePlanInput{
		KinesiologistID: req.KinesiologistID,
		DiagnosisID:     req.DiagnosisID,
		Frequency:       req.Frequency,
		DurationWeeks:   req.DurationWeeks,
		Observations:    req.Observations,
		Status:          req.Status,
		Items:           req.Items,
		ActorEmail:      actorEmail(c),
		ActorRole:       actorRole(c),
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusOK, toResponse(out))
}

func (h *Handler) GetByID(c *gin.Context) {
	if !middleware.HasRole(c, "recepcionista", "kinesiologo") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

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

func (h *Handler) patientIDForCurrentPatient(c *gin.Context, requestedPatientID string) (string, bool) {
	user, _ := middleware.CurrentUser(c)
	if h.patients == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return "", false
	}

	email := strings.TrimSpace(user.Email)
	if email == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "patient_profile_not_found"})
		return "", false
	}

	patients, err := h.patients.Search(c.Request.Context(), email, 10, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return "", false
	}

	for _, patient := range patients {
		if strings.EqualFold(patient.Email, email) && patient.Active {
			ownID := patient.ID.String()
			requested := strings.TrimSpace(requestedPatientID)
			if requested != "" && !strings.EqualFold(requested, "me") && requested != ownID {
				c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
				return "", false
			}
			return ownID, true
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "patient_profile_not_found"})
	return "", false
}

func (h *Handler) isCurrentPatient(c *gin.Context) bool {
	user, ok := middleware.CurrentUser(c)
	return ok && strings.EqualFold(user.Role, "paciente")
}

func actorEmail(c *gin.Context) string {
	if user, ok := middleware.CurrentUser(c); ok {
		return user.Email
	}
	if middleware.HasRole(c, "recepcionista") {
		return "demo@local"
	}
	return ""
}

func actorRole(c *gin.Context) string {
	if user, ok := middleware.CurrentUser(c); ok {
		return user.Role
	}
	if middleware.HasRole(c, "recepcionista") {
		return "recepcionista"
	}
	return ""
}
