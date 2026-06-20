package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/clinicalreports/domain"
	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
)

type Store interface {
	Create(ctx context.Context, patientID uuid.UUID, kinesiologistID uuid.UUID, periodFrom time.Time, periodTo time.Time, recommendations *string, actorEmail *string, actorRole *string) (domain.ClinicalReport, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit int) ([]domain.ClinicalReport, error)
}

type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

type createReportRequest struct {
	KinesiologistID string  `json:"kinesiologist_id"`
	PeriodFrom      string  `json:"period_from"`
	PeriodTo        string  `json:"period_to"`
	Recommendations *string `json:"recommendations"`
}

type reportResponse struct {
	ID                  string   `json:"id"`
	PatientID           string   `json:"patient_id"`
	KinesiologistID     string   `json:"kinesiologist_id"`
	PeriodFrom          string   `json:"period_from"`
	PeriodTo            string   `json:"period_to"`
	EvolutionCount      int      `json:"evolution_count"`
	AvgPainLevel        *float64 `json:"avg_pain_level,omitempty"`
	AvgMobilityScore    *float64 `json:"avg_mobility_score,omitempty"`
	AvgStrengthScore    *float64 `json:"avg_strength_score,omitempty"`
	AvgFunctionalScore  *float64 `json:"avg_functional_score,omitempty"`
	ActivePlanCount     int      `json:"active_plan_count"`
	ActivePlanItemCount int      `json:"active_plan_item_count"`
	Summary             string   `json:"summary"`
	Recommendations     *string  `json:"recommendations,omitempty"`
	CreatedByEmail      *string  `json:"created_by_email,omitempty"`
	CreatedByRole       *string  `json:"created_by_role,omitempty"`
	CreatedAt           string   `json:"created_at"`
	UpdatedAt           string   `json:"updated_at"`
}

func (h *Handler) CreateForPatient(c *gin.Context) {
	if !middleware.HasRole(c, "kinesiologo") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	patientID, err := uuid.Parse(strings.TrimSpace(c.Param("patient_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": gin.H{"patient_id": "UUID inválido"}})
		return
	}

	var req createReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	kinesiologistID, validationErrs := parseKinesiologistID(req.KinesiologistID)
	periodFrom, okFrom := parseDate(req.PeriodFrom)
	periodTo, okTo := parseDate(req.PeriodTo)
	if !okFrom {
		validationErrs["period_from"] = "Fecha inválida"
	}
	if !okTo {
		validationErrs["period_to"] = "Fecha inválida"
	}
	if okFrom && okTo && periodTo.Before(periodFrom) {
		validationErrs["period_to"] = "Debe ser posterior o igual al inicio"
	}
	if len(validationErrs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validationErrs})
		return
	}

	out, err := h.store.Create(c.Request.Context(), patientID, kinesiologistID, periodFrom, periodTo, trimOptional(req.Recommendations), actorEmail(c), actorRole(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(out))
}

func (h *Handler) ListByPatient(c *gin.Context) {
	if !middleware.HasRole(c, "recepcionista", "kinesiologo") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	patientID, err := uuid.Parse(strings.TrimSpace(c.Param("patient_id")))
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

	out := make([]reportResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toResponse(item))
	}
	c.JSON(http.StatusOK, out)
}

func parseKinesiologistID(value string) (uuid.UUID, map[string]string) {
	errs := map[string]string{}
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		errs["kinesiologist_id"] = "UUID inválido"
	}
	return id, errs
}

func parseDate(value string) (time.Time, bool) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	return parsed, err == nil
}

func actorEmail(c *gin.Context) *string {
	user, ok := middleware.CurrentUser(c)
	if !ok || strings.TrimSpace(user.Email) == "" {
		return nil
	}
	email := strings.TrimSpace(user.Email)
	return &email
}

func actorRole(c *gin.Context) *string {
	user, ok := middleware.CurrentUser(c)
	if !ok || strings.TrimSpace(user.Role) == "" {
		return nil
	}
	role := strings.TrimSpace(user.Role)
	return &role
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func toResponse(report domain.ClinicalReport) reportResponse {
	return reportResponse{
		ID:                  report.ID.String(),
		PatientID:           report.PatientID.String(),
		KinesiologistID:     report.KinesiologistID.String(),
		PeriodFrom:          report.PeriodFrom.Format("2006-01-02"),
		PeriodTo:            report.PeriodTo.Format("2006-01-02"),
		EvolutionCount:      report.EvolutionCount,
		AvgPainLevel:        report.AvgPainLevel,
		AvgMobilityScore:    report.AvgMobilityScore,
		AvgStrengthScore:    report.AvgStrengthScore,
		AvgFunctionalScore:  report.AvgFunctionalScore,
		ActivePlanCount:     report.ActivePlanCount,
		ActivePlanItemCount: report.ActivePlanItemCount,
		Summary:             report.Summary,
		Recommendations:     report.Recommendations,
		CreatedByEmail:      report.CreatedByEmail,
		CreatedByRole:       report.CreatedByRole,
		CreatedAt:           report.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:           report.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
