package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/evolutions/domain"
	"github.com/javiacuna/kinesio-backend/internal/evolutions/usecase"
	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	patientDomain "github.com/javiacuna/kinesio-backend/internal/patients/domain"
)

type Handler struct {
	createUC *usecase.CreateEvolutionUseCase
	listUC   *usecase.ListEvolutionsByPatientUseCase
	getUC    *usecase.GetEvolutionByIDUseCase
	patients patientSearcher
}

type patientSearcher interface {
	Search(ctx context.Context, query string, limit int, includeInactive bool) ([]patientDomain.Patient, error)
}

func NewHandler(createUC *usecase.CreateEvolutionUseCase, listUC *usecase.ListEvolutionsByPatientUseCase, getUC *usecase.GetEvolutionByIDUseCase, lookups ...any) *Handler {
	var patientRepo patientSearcher
	for _, lookup := range lookups {
		if repo, ok := lookup.(patientSearcher); ok {
			patientRepo = repo
		}
	}
	return &Handler{createUC: createUC, listUC: listUC, getUC: getUC, patients: patientRepo}
}

type createEvolutionRequest struct {
	KinesiologistID string                              `json:"kinesiologist_id"`
	AppointmentID   *string                             `json:"appointment_id,omitempty"`
	DiagnosisID     *string                             `json:"patient_diagnosis_id,omitempty"`
	PainLevel       *int                                `json:"pain_level,omitempty"`
	MobilityScore   *int                                `json:"mobility_score,omitempty"`
	StrengthScore   *int                                `json:"strength_score,omitempty"`
	FunctionalScore *int                                `json:"functional_score,omitempty"`
	Notes           string                              `json:"notes"`
	Photos          []usecase.CreateEvolutionPhotoInput `json:"photos,omitempty"`
}

type evolutionResponse struct {
	ID              string                   `json:"id"`
	PatientID       string                   `json:"patient_id"`
	KinesiologistID string                   `json:"kinesiologist_id"`
	AppointmentID   *string                  `json:"appointment_id,omitempty"`
	DiagnosisID     *string                  `json:"patient_diagnosis_id,omitempty"`
	PainLevel       *int                     `json:"pain_level,omitempty"`
	MobilityScore   *int                     `json:"mobility_score,omitempty"`
	StrengthScore   *int                     `json:"strength_score,omitempty"`
	FunctionalScore *int                     `json:"functional_score,omitempty"`
	Notes           string                   `json:"notes"`
	Photos          []evolutionPhotoResponse `json:"photos"`
	CreatedAt       string                   `json:"created_at"`
	UpdatedAt       string                   `json:"updated_at"`
}

type evolutionPhotoResponse struct {
	ID        string  `json:"id"`
	URL       string  `json:"url"`
	Caption   *string `json:"caption,omitempty"`
	CreatedAt string  `json:"created_at"`
}

func (h *Handler) CreateForPatient(c *gin.Context) {
	patientID := c.Param("patient_id")

	var req createEvolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	h.create(c, patientID, req)
}

func (h *Handler) create(c *gin.Context, patientID string, req createEvolutionRequest) {
	out, validation, err := h.createUC.Execute(c.Request.Context(), usecase.CreateEvolutionInput{
		PatientID:       patientID,
		KinesiologistID: req.KinesiologistID,
		AppointmentID:   req.AppointmentID,
		DiagnosisID:     req.DiagnosisID,
		PainLevel:       req.PainLevel,
		MobilityScore:   req.MobilityScore,
		StrengthScore:   req.StrengthScore,
		FunctionalScore: req.FunctionalScore,
		Notes:           req.Notes,
		Photos:          req.Photos,
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
	patientID := strings.TrimSpace(c.Param("patient_id"))
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

func toResponse(e domain.PatientEvolution) evolutionResponse {
	var appt *string
	if e.AppointmentID != nil {
		s := e.AppointmentID.String()
		appt = &s
	}
	var diagnosis *string
	if e.DiagnosisID != nil {
		s := e.DiagnosisID.String()
		diagnosis = &s
	}

	photos := make([]evolutionPhotoResponse, 0, len(e.Photos))
	for _, photo := range e.Photos {
		photos = append(photos, evolutionPhotoResponse{
			ID:        photo.ID.String(),
			URL:       photo.URL,
			Caption:   photo.Caption,
			CreatedAt: photo.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return evolutionResponse{
		ID:              e.ID.String(),
		PatientID:       e.PatientID.String(),
		KinesiologistID: e.KinesiologistID.String(),
		AppointmentID:   appt,
		DiagnosisID:     diagnosis,
		PainLevel:       e.PainLevel,
		MobilityScore:   e.MobilityScore,
		StrengthScore:   e.StrengthScore,
		FunctionalScore: e.FunctionalScore,
		Notes:           e.Notes,
		Photos:          photos,
		CreatedAt:       e.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       e.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
