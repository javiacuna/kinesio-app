package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	"github.com/javiacuna/kinesio-backend/internal/patients/domain"
	"github.com/javiacuna/kinesio-backend/internal/patients/usecase"
)

// kinesiologistPatientsResolver resuelve, a partir del email de un kinesiólogo, los IDs
// de los pacientes con los que alguna vez tuvo un turno agendado. Se usa para que el
// listado/búsqueda de pacientes le muestre a un kinesiólogo solo a sus propios pacientes.
type kinesiologistPatientsResolver interface {
	ListPatientIDsForKinesiologistEmail(ctx context.Context, email string) ([]uuid.UUID, error)
}

type Handler struct {
	register              *usecase.RegisterPatientUseCase
	update                *usecase.UpdatePatientUseCase
	delete                *usecase.DeletePatientUseCase
	getByID               *usecase.GetPatientByIDUseCase
	listUC                *usecase.ListPatientsUseCase
	searchUC              *usecase.SearchPatientsUseCase
	kinesiologistPatients kinesiologistPatientsResolver
}

func NewHandler(register *usecase.RegisterPatientUseCase, update *usecase.UpdatePatientUseCase, deleteUC *usecase.DeletePatientUseCase,
	getByID *usecase.GetPatientByIDUseCase, listUC *usecase.ListPatientsUseCase,
	searchUC *usecase.SearchPatientsUseCase, lookups ...any) *Handler {
	var kinesiologistPatients kinesiologistPatientsResolver
	for _, lookup := range lookups {
		if resolver, ok := lookup.(kinesiologistPatientsResolver); ok {
			kinesiologistPatients = resolver
		}
	}

	return &Handler{
		register:              register,
		update:                update,
		delete:                deleteUC,
		getByID:               getByID,
		listUC:                listUC,
		searchUC:              searchUC,
		kinesiologistPatients: kinesiologistPatients,
	}
}

type registerPatientRequest struct {
	DNI                   string  `json:"dni"`
	FirstName             string  `json:"first_name"`
	LastName              string  `json:"last_name"`
	Email                 string  `json:"email"`
	Phone                 *string `json:"phone"`
	BirthDate             *string `json:"birth_date"` // YYYY-MM-DD
	ClinicalNotes         *string `json:"clinical_notes"`
	FinancierID           *string `json:"financier_id"`
	FinancierMemberNumber *string `json:"financier_member_number"`
	Active                *bool   `json:"active,omitempty"`
}

type patientResponse struct {
	ID                          string  `json:"id"`
	DNI                         string  `json:"dni"`
	FirstName                   string  `json:"first_name"`
	LastName                    string  `json:"last_name"`
	Email                       string  `json:"email"`
	Phone                       *string `json:"phone,omitempty"`
	BirthDate                   *string `json:"birth_date,omitempty"`
	ClinicalNotes               *string `json:"clinical_notes,omitempty"`
	ClinicalNotesUpdatedByEmail *string `json:"clinical_notes_updated_by_email,omitempty"`
	ClinicalNotesUpdatedByRole  *string `json:"clinical_notes_updated_by_role,omitempty"`
	ClinicalNotesUpdatedAt      *string `json:"clinical_notes_updated_at,omitempty"`
	FinancierID                 *string `json:"financier_id,omitempty"`
	FinancierMemberNumber       *string `json:"financier_member_number,omitempty"`
	Active                      bool    `json:"active"`
	CreatedAt                   string  `json:"created_at"`
	UpdatedAt                   string  `json:"updated_at"`
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
		DNI:                   req.DNI,
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Email:                 req.Email,
		Phone:                 req.Phone,
		BirthDate:             req.BirthDate,
		ClinicalNotes:         req.ClinicalNotes,
		FinancierID:           req.FinancierID,
		FinancierMemberNumber: req.FinancierMemberNumber,
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
	if !middleware.HasRole(c, "recepcionista", "kinesiologo") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req registerPatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	out, validation, err := h.update.Execute(c.Request.Context(), usecase.UpdatePatientInput{
		ID:                    c.Param("patient_id"),
		DNI:                   req.DNI,
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Email:                 req.Email,
		Phone:                 req.Phone,
		BirthDate:             req.BirthDate,
		ClinicalNotes:         req.ClinicalNotes,
		FinancierID:           req.FinancierID,
		FinancierMemberNumber: req.FinancierMemberNumber,
		Active:                req.Active,
		ActorEmail:            actorEmail(c),
		ActorRole:             actorRole(c),
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
	var clinicalNotesUpdatedAt *string
	if p.ClinicalNotesUpdatedAt != nil {
		s := p.ClinicalNotesUpdatedAt.UTC().Format(timeRFC3339())
		clinicalNotesUpdatedAt = &s
	}
	var financierID *string
	if p.FinancierID != nil {
		s := p.FinancierID.String()
		financierID = &s
	}

	return patientResponse{
		ID:                          p.ID.String(),
		DNI:                         p.DNI,
		FirstName:                   p.FirstName,
		LastName:                    p.LastName,
		Email:                       p.Email,
		Phone:                       p.Phone,
		BirthDate:                   birth,
		ClinicalNotes:               p.ClinicalNotes,
		ClinicalNotesUpdatedByEmail: p.ClinicalNotesUpdatedByEmail,
		ClinicalNotesUpdatedByRole:  p.ClinicalNotesUpdatedByRole,
		ClinicalNotesUpdatedAt:      clinicalNotesUpdatedAt,
		FinancierID:                 financierID,
		FinancierMemberNumber:       p.FinancierMemberNumber,
		Active:                      p.Active,
		CreatedAt:                   p.CreatedAt.UTC().Format(timeRFC3339()),
		UpdatedAt:                   p.UpdatedAt.UTC().Format(timeRFC3339()),
	}
}

func timeRFC3339() string { return "2006-01-02T15:04:05Z07:00" }

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

func (h *Handler) GetPatientByID(c *gin.Context) {
	// El handler no valida rol por sí mismo: la ruta se registra en
	// router.go con middleware.RequireRole("recepcionista", "kinesiologo")
	// + patientAccessGuard, que son quienes exigen sesión y limitan el
	// acceso de un kinesiólogo a sus propios pacientes.
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
	if !middleware.HasRole(c, "recepcionista", "kinesiologo") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Un kinesiólogo solo puede ver/buscar entre sus propios pacientes (con los que
	// tuvo algun turno agendado), nunca la nomina completa.
	var allowedPatientIDs map[uuid.UUID]struct{}
	if user, ok := middleware.CurrentUser(c); ok && strings.EqualFold(strings.TrimSpace(user.Role), "kinesiologo") {
		if h.kinesiologistPatients == nil {
			c.JSON(http.StatusOK, []patientResponse{})
			return
		}
		ids, err := h.kinesiologistPatients.ListPatientIDsForKinesiologistEmail(c.Request.Context(), user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
		allowedPatientIDs = make(map[uuid.UUID]struct{}, len(ids))
		for _, id := range ids {
			allowedPatientIDs[id] = struct{}{}
		}
	}

	q := strings.TrimSpace(c.Query("query"))
	limit := 20
	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		if parsedLimit, err := strconv.Atoi(rawLimit); err == nil {
			limit = parsedLimit
		}
	}
	includeInactive := strings.EqualFold(strings.TrimSpace(c.Query("active")), "false")

	// Si estamos filtrando por kinesiólogo, traemos un lote amplio antes de filtrar
	// (paginar y despues filtrar podria dejar afuera pacientes suyos por el camino).
	queryLimit := limit
	if allowedPatientIDs != nil && queryLimit < 500 {
		queryLimit = 500
	}

	if q == "" {
		offset := 0
		if rawOffset := strings.TrimSpace(c.Query("offset")); rawOffset != "" {
			if parsedOffset, err := strconv.Atoi(rawOffset); err == nil {
				offset = parsedOffset
			}
		}
		if allowedPatientIDs != nil {
			offset = 0
		}

		items, err := h.listUC.Execute(c.Request.Context(), queryLimit, offset, includeInactive)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}

		out := make([]patientResponse, 0, len(items))
		for _, p := range items {
			if allowedPatientIDs != nil {
				if _, ok := allowedPatientIDs[p.ID]; !ok {
					continue
				}
			}
			out = append(out, toResponse(p))
			if len(out) >= limit {
				break
			}
		}

		c.JSON(http.StatusOK, out)
		return
	}

	items, err := h.searchUC.Execute(c.Request.Context(), q, queryLimit, includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if allowedPatientIDs != nil {
		filtered := make([]domain.Patient, 0, len(items))
		for _, p := range items {
			if _, ok := allowedPatientIDs[p.ID]; ok {
				filtered = append(filtered, p)
				if len(filtered) >= limit {
					break
				}
			}
		}
		items = filtered
	}

	out := make([]patientResponse, 0, len(items))
	for _, p := range items {
		out = append(out, toResponse(p))
	}

	c.JSON(http.StatusOK, out)
}
