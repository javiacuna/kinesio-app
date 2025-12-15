package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/javiacuna/kinesio-backend/internal/patients/domain"
	"github.com/javiacuna/kinesio-backend/internal/patients/usecase"
)

type Handler struct {
	register *usecase.RegisterPatientUseCase
	getByID  *usecase.GetPatientByIDUseCase
	searchUC *usecase.SearchPatientsUseCase
}

func NewHandler(register *usecase.RegisterPatientUseCase, getByID *usecase.GetPatientByIDUseCase,
	searchUC *usecase.SearchPatientsUseCase) *Handler {
	return &Handler{register: register, getByID: getByID, searchUC: searchUC}
}

type registerPatientRequest struct {
	DNI           string  `json:"dni"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Email         string  `json:"email"`
	Phone         *string `json:"phone"`
	BirthDate     *string `json:"birth_date"` // YYYY-MM-DD
	ClinicalNotes *string `json:"clinical_notes"`
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
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func (h *Handler) RegisterPatient(c *gin.Context) {
	// Auth demo (hasta integrar Firebase real): requiere header
	// Authorization: Bearer demo-recepcionista-token
	if !isReceptionist(c.GetHeader("Authorization")) {
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
		CreatedAt:     p.CreatedAt.UTC().Format(timeRFC3339()),
		UpdatedAt:     p.UpdatedAt.UTC().Format(timeRFC3339()),
	}
}

func timeRFC3339() string { return "2006-01-02T15:04:05Z07:00" }

func isReceptionist(auth string) bool {
	// Demo token
	auth = strings.TrimSpace(auth)
	return strings.EqualFold(auth, "Bearer demo-recepcionista-token")
}

func (h *Handler) GetPatientByID(c *gin.Context) {
	// Por ahora lo dejo SIN auth (útil para debug y para frontend).
	// Lo cerramos por rol en el siguiente paso.
	id := c.Param("id")

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
	// Auth demo (igual que Register): requiere header
	if !isReceptionist(c.GetHeader("Authorization")) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	q := strings.TrimSpace(c.Query("query"))
	limit := 20

	items, err := h.searchUC.Execute(c.Request.Context(), q, limit)
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
