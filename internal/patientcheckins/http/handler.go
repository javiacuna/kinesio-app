package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	"github.com/javiacuna/kinesio-backend/internal/patientcheckins/domain"
	patientDomain "github.com/javiacuna/kinesio-backend/internal/patients/domain"
)

type Store interface {
	Create(ctx context.Context, item domain.PatientCheckIn) (domain.PatientCheckIn, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit int) ([]domain.PatientCheckIn, error)
}

type patientSearcher interface {
	Search(ctx context.Context, query string, limit int, includeInactive bool) ([]patientDomain.Patient, error)
}

type Handler struct {
	store    Store
	patients patientSearcher
}

func NewHandler(store Store, lookups ...any) *Handler {
	var patientRepo patientSearcher
	for _, lookup := range lookups {
		if repo, ok := lookup.(patientSearcher); ok {
			patientRepo = repo
		}
	}
	return &Handler{store: store, patients: patientRepo}
}

type createCheckInRequest struct {
	PainLevel       *int   `json:"pain_level,omitempty"`
	MobilityScore   *int   `json:"mobility_score,omitempty"`
	StrengthScore   *int   `json:"strength_score,omitempty"`
	FunctionalScore *int   `json:"functional_score,omitempty"`
	Notes           string `json:"notes"`
}

type checkInResponse struct {
	ID              string `json:"id"`
	PatientID       string `json:"patient_id"`
	PainLevel       *int   `json:"pain_level,omitempty"`
	MobilityScore   *int   `json:"mobility_score,omitempty"`
	StrengthScore   *int   `json:"strength_score,omitempty"`
	FunctionalScore *int   `json:"functional_score,omitempty"`
	Notes           string `json:"notes"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func (h *Handler) CreateForMe(c *gin.Context) {
	if !h.isCurrentPatient(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	patientID, ok := h.patientIDForCurrentPatient(c, "me")
	if !ok {
		return
	}

	var req createCheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	validation := validateRequest(req)
	if len(validation) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
		return
	}

	created, err := h.store.Create(c.Request.Context(), domain.NewPatientCheckIn(
		patientID,
		req.PainLevel,
		req.MobilityScore,
		req.StrengthScore,
		req.FunctionalScore,
		req.Notes,
	))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(created))
}

func (h *Handler) ListByPatient(c *gin.Context) {
	patientIDParam := strings.TrimSpace(c.Param("patient_id"))

	if strings.EqualFold(patientIDParam, "me") || h.isCurrentPatient(c) {
		resolved, ok := h.patientIDForCurrentPatient(c, patientIDParam)
		if !ok {
			return
		}
		patientIDParam = resolved.String()
	} else if !middleware.HasRole(c, "admin", "recepcionista", "kinesiologo") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	patientID, err := uuid.Parse(patientIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_patient_id"})
		return
	}

	limit := 20
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			limit = n
		}
	}

	items, err := h.store.ListByPatient(c.Request.Context(), patientID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]checkInResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toResponse(item))
	}
	c.JSON(http.StatusOK, out)
}

func validateRequest(req createCheckInRequest) map[string]string {
	validation := map[string]string{}
	if strings.TrimSpace(req.Notes) == "" {
		validation["notes"] = "required"
	}
	if req.PainLevel != nil && (*req.PainLevel < 0 || *req.PainLevel > 10) {
		validation["pain_level"] = "must_be_between_0_and_10"
	}
	validatePercent(validation, "mobility_score", req.MobilityScore)
	validatePercent(validation, "strength_score", req.StrengthScore)
	validatePercent(validation, "functional_score", req.FunctionalScore)
	return validation
}

func validatePercent(validation map[string]string, field string, value *int) {
	if value != nil && (*value < 0 || *value > 100) {
		validation[field] = "must_be_between_0_and_100"
	}
}

func (h *Handler) patientIDForCurrentPatient(c *gin.Context, requestedPatientID string) (uuid.UUID, bool) {
	user, _ := middleware.CurrentUser(c)
	if h.patients == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return uuid.Nil, false
	}

	email := strings.TrimSpace(user.Email)
	if email == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "patient_profile_not_found"})
		return uuid.Nil, false
	}

	patients, err := h.patients.Search(c.Request.Context(), email, 10, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return uuid.Nil, false
	}

	for _, patient := range patients {
		if strings.EqualFold(patient.Email, email) && patient.Active {
			requested := strings.TrimSpace(requestedPatientID)
			if requested != "" && !strings.EqualFold(requested, "me") && requested != patient.ID.String() {
				c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
				return uuid.Nil, false
			}
			return patient.ID, true
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "patient_profile_not_found"})
	return uuid.Nil, false
}

func (h *Handler) isCurrentPatient(c *gin.Context) bool {
	user, ok := middleware.CurrentUser(c)
	return ok && strings.EqualFold(user.Role, "paciente")
}

func toResponse(item domain.PatientCheckIn) checkInResponse {
	return checkInResponse{
		ID:              item.ID.String(),
		PatientID:       item.PatientID.String(),
		PainLevel:       item.PainLevel,
		MobilityScore:   item.MobilityScore,
		StrengthScore:   item.StrengthScore,
		FunctionalScore: item.FunctionalScore,
		Notes:           item.Notes,
		CreatedAt:       item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
