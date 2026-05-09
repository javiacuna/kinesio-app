package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/javiacuna/kinesio-backend/internal/diagnoses/domain"
	"github.com/javiacuna/kinesio-backend/internal/diagnoses/usecase"
	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
)

type Handler struct {
	searchCIE10 *usecase.SearchCIE10UseCase
	list        *usecase.ListPatientDiagnosesUseCase
	save        *usecase.SavePatientDiagnosisUseCase
	delete      *usecase.DeletePatientDiagnosisUseCase
}

func NewHandler(
	searchCIE10 *usecase.SearchCIE10UseCase,
	list *usecase.ListPatientDiagnosesUseCase,
	save *usecase.SavePatientDiagnosisUseCase,
	delete *usecase.DeletePatientDiagnosisUseCase,
) *Handler {
	return &Handler{
		searchCIE10: searchCIE10,
		list:        list,
		save:        save,
		delete:      delete,
	}
}

type cie10Response struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Chapter     *string `json:"chapter,omitempty"`
}

type diagnosisRequest struct {
	CIE10Code   string  `json:"cie10_code"`
	Kind        string  `json:"kind"`
	Status      string  `json:"status"`
	DiagnosedAt string  `json:"diagnosed_at"`
	Notes       *string `json:"notes"`
}

type diagnosisResponse struct {
	ID             string  `json:"id"`
	PatientID      string  `json:"patient_id"`
	CIE10Code      string  `json:"cie10_code"`
	CIE10Desc      string  `json:"cie10_description"`
	CIE10Chapter   *string `json:"cie10_chapter,omitempty"`
	Kind           string  `json:"kind"`
	Status         string  `json:"status"`
	DiagnosedAt    string  `json:"diagnosed_at"`
	Notes          *string `json:"notes,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	CreatedByEmail *string `json:"created_by_email,omitempty"`
	CreatedByRole  *string `json:"created_by_role,omitempty"`
	UpdatedByEmail *string `json:"updated_by_email,omitempty"`
	UpdatedByRole  *string `json:"updated_by_role,omitempty"`
}

func (h *Handler) SearchCIE10(c *gin.Context) {
	limit := 30
	if value := strings.TrimSpace(c.Query("limit")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			limit = parsed
		}
	}

	items, err := h.searchCIE10.Execute(c.Request.Context(), c.Query("query"), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]cie10Response, 0, len(items))
	for _, item := range items {
		out = append(out, toCIE10Response(item))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) ListByPatient(c *gin.Context) {
	items, validation, err := h.list.Execute(c.Request.Context(), c.Param("patient_id"))
	if err != nil {
		writeDiagnosisError(c, validation, err)
		return
	}

	out := make([]diagnosisResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toDiagnosisResponse(item))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) CreateForPatient(c *gin.Context) {
	var req diagnosisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	out, validation, err := h.save.Create(c.Request.Context(), usecase.SavePatientDiagnosisInput{
		PatientID:   c.Param("patient_id"),
		CIE10Code:   req.CIE10Code,
		Kind:        req.Kind,
		Status:      req.Status,
		DiagnosedAt: req.DiagnosedAt,
		Notes:       req.Notes,
		ActorEmail:  actorEmail(c),
		ActorRole:   actorRole(c),
	})
	if err != nil {
		writeDiagnosisError(c, validation, err)
		return
	}

	c.JSON(http.StatusCreated, toDiagnosisResponse(out))
}

func (h *Handler) Update(c *gin.Context) {
	var req diagnosisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	out, validation, err := h.save.Update(c.Request.Context(), usecase.SavePatientDiagnosisInput{
		ID:          c.Param("diagnosis_id"),
		PatientID:   c.Param("patient_id"),
		CIE10Code:   req.CIE10Code,
		Kind:        req.Kind,
		Status:      req.Status,
		DiagnosedAt: req.DiagnosedAt,
		Notes:       req.Notes,
		ActorEmail:  actorEmail(c),
		ActorRole:   actorRole(c),
	})
	if err != nil {
		writeDiagnosisError(c, validation, err)
		return
	}

	c.JSON(http.StatusOK, toDiagnosisResponse(out))
}

func (h *Handler) Delete(c *gin.Context) {
	validation, err := h.delete.Execute(c.Request.Context(), c.Param("diagnosis_id"))
	if err != nil {
		writeDiagnosisError(c, validation, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func writeDiagnosisError(c *gin.Context, validation map[string]string, err error) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}

func toCIE10Response(item domain.CIE10Code) cie10Response {
	return cie10Response{
		Code:        item.Code,
		Description: item.Description,
		Chapter:     item.Chapter,
	}
}

func toDiagnosisResponse(item domain.PatientDiagnosis) diagnosisResponse {
	return diagnosisResponse{
		ID:             item.ID.String(),
		PatientID:      item.PatientID.String(),
		CIE10Code:      item.CIE10Code,
		CIE10Desc:      item.CIE10.Description,
		CIE10Chapter:   item.CIE10.Chapter,
		Kind:           item.Kind,
		Status:         item.Status,
		DiagnosedAt:    item.DiagnosedAt.Format("2006-01-02"),
		Notes:          item.Notes,
		CreatedAt:      item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      item.UpdatedAt.UTC().Format(time.RFC3339),
		CreatedByEmail: item.CreatedByEmail,
		CreatedByRole:  item.CreatedByRole,
		UpdatedByEmail: item.UpdatedByEmail,
		UpdatedByRole:  item.UpdatedByRole,
	}
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
