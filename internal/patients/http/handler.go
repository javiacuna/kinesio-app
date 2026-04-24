package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	"github.com/javiacuna/kinesio-backend/internal/patients/domain"
	"github.com/javiacuna/kinesio-backend/internal/patients/usecase"
)

type Handler struct {
	register *usecase.RegisterPatientUseCase
	update   *usecase.UpdatePatientUseCase
	delete   *usecase.DeletePatientUseCase
	getByID  *usecase.GetPatientByIDUseCase
	listUC   *usecase.ListPatientsUseCase
	searchUC *usecase.SearchPatientsUseCase
}

func NewHandler(register *usecase.RegisterPatientUseCase, update *usecase.UpdatePatientUseCase, deleteUC *usecase.DeletePatientUseCase,
	getByID *usecase.GetPatientByIDUseCase, listUC *usecase.ListPatientsUseCase,
	searchUC *usecase.SearchPatientsUseCase) *Handler {
	return &Handler{register: register, update: update, delete: deleteUC, getByID: getByID, listUC: listUC, searchUC: searchUC}
}

type registerPatientRequest struct {
	DNI           string  `json:"dni"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Email         string  `json:"email"`
	Phone         *string `json:"phone"`
	BirthDate     *string `json:"birth_date"` // YYYY-MM-DD
	ClinicalNotes *string `json:"clinical_notes"`
	Active        *bool   `json:"active,omitempty"`
}

type patientResponse struct {
	ID            string  `json:"id"`
	DNI           string  `json:"dni"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Email         string  `json:"email"`
	Phone         *string `json:"phone,omitempty"`
	BirthDate     *string `json:"birth_date,omitempty"`
	ClinicalNotes *string `json:"clinical_notes,omitempty"`
	Active        bool    `json:"active"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func (h *Handler) RegisterPatient(c *gin.Context) {
	if !middleware.HasRole(c, "recepcionista") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req registerPatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	out, validation, err := h.register.Execute(c.Request.Context(), usecase.RegisterPatientInput{
		DNI:           req.DNI,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Email:         req.Email,
		Phone:         req.Phone,
		BirthDate:     req.BirthDate,
		ClinicalNotes: req.ClinicalNotes,
	})

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
			return
		case errors.Is(err, domain.ErrDuplicateDNI):
			c.JSON(http.StatusConflict, gin.H{"error": "dni_duplicado"})
			return
		case errors.Is(err, domain.ErrDuplicateEmail):
			c.JSON(http.StatusConflict, gin.H{"error": "email_duplicado"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
	}

	resp := toResponse(out)
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) UpdatePatient(c *gin.Context) {
	if !middleware.HasRole(c, "recepcionista") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req registerPatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	out, validation, err := h.update.Execute(c.Request.Context(), usecase.UpdatePatientInput{
		ID:            c.Param("patient_id"),
		DNI:           req.DNI,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Email:         req.Email,
		Phone:         req.Phone,
		BirthDate:     req.BirthDate,
		ClinicalNotes: req.ClinicalNotes,
		Active:        req.Active,
	})

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
			return
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		case errors.Is(err, domain.ErrDuplicateDNI):
			c.JSON(http.StatusConflict, gin.H{"error": "dni_duplicado"})
			return
		case errors.Is(err, domain.ErrDuplicateEmail):
			c.JSON(http.StatusConflict, gin.H{"error": "email_duplicado"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
	}

	c.JSON(http.StatusOK, toResponse(out))
}

func (h *Handler) DeletePatient(c *gin.Context) {
	if !middleware.HasRole(c, "recepcionista") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	validation, err := h.delete.Execute(c.Request.Context(), c.Param("patient_id"))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
			return
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
	}

	c.Status(http.StatusNoContent)
}

func toResponse(p domain.Patient) patientResponse {
	var birth *string
	if p.BirthDate != nil {
		s := p.BirthDate.Format("2006-01-02")
		birth = &s
	}

	return patientResponse{
		ID:            p.ID.String(),
		DNI:           p.DNI,
		FirstName:     p.FirstName,
		LastName:      p.LastName,
		Email:         p.Email,
		Phone:         p.Phone,
		BirthDate:     birth,
		ClinicalNotes: p.ClinicalNotes,
		Active:        p.Active,
		CreatedAt:     p.CreatedAt.UTC().Format(timeRFC3339()),
		UpdatedAt:     p.UpdatedAt.UTC().Format(timeRFC3339()),
	}
}

func timeRFC3339() string { return "2006-01-02T15:04:05Z07:00" }

func (h *Handler) GetPatientByID(c *gin.Context) {
	// Por ahora lo dejo SIN auth (útil para debug y para frontend).
	// Lo cerramos por rol en el siguiente paso.
	id := c.Param("patient_id")

	p, found, err := h.getByID.Execute(c.Request.Context(), id)
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

func (h *Handler) Search(c *gin.Context) {
	if !middleware.HasRole(c, "recepcionista") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	q := strings.TrimSpace(c.Query("query"))
	limit := 20
	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		if parsedLimit, err := strconv.Atoi(rawLimit); err == nil {
			limit = parsedLimit
		}
	}
	includeInactive := strings.EqualFold(strings.TrimSpace(c.Query("active")), "false")

	if q == "" {
		offset := 0
		if rawOffset := strings.TrimSpace(c.Query("offset")); rawOffset != "" {
			if parsedOffset, err := strconv.Atoi(rawOffset); err == nil {
				offset = parsedOffset
			}
		}

		items, err := h.listUC.Execute(c.Request.Context(), limit, offset, includeInactive)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}

		out := make([]patientResponse, 0, len(items))
		for _, p := range items {
			out = append(out, toResponse(p))
		}

		c.JSON(http.StatusOK, out)
		return
	}

	items, err := h.searchUC.Execute(c.Request.Context(), q, limit, includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]patientResponse, 0, len(items))
	for _, p := range items {
		out = append(out, toResponse(p))
	}

	c.JSON(http.StatusOK, out)
}
