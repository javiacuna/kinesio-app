package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/javiacuna/kinesio-backend/internal/appointments/domain"
	"github.com/javiacuna/kinesio-backend/internal/appointments/usecase"
)

type Handler struct {
	create        *usecase.CreateAppointmentUseCase
	listDay       *usecase.ListAppointmentsDayUseCase
	update        *usecase.UpdateAppointmentUseCase
	getByID       *usecase.GetAppointmentByIDUseCase
	listByPatient *usecase.ListAppointmentsByPatientUseCase
}

func NewHandler(
	create *usecase.CreateAppointmentUseCase,
	listDay *usecase.ListAppointmentsDayUseCase,
	update *usecase.UpdateAppointmentUseCase,
	getByID *usecase.GetAppointmentByIDUseCase,
	listByPatient *usecase.ListAppointmentsByPatientUseCase,
) *Handler {
	return &Handler{
		create:        create,
		listDay:       listDay,
		update:        update,
		getByID:       getByID,
		listByPatient: listByPatient,
	}
}

type createReq struct {
	PatientID       string  `json:"patient_id"`
	KinesiologistID string  `json:"kinesiologist_id"`
	StartAt         string  `json:"start_at"` // RFC3339
	EndAt           string  `json:"end_at"`   // RFC3339
	Notes           *string `json:"notes,omitempty"`
}

type updateReq struct {
	StartAt         *string `json:"start_at,omitempty"`
	EndAt           *string `json:"end_at,omitempty"`
	Status          *string `json:"status,omitempty"` // scheduled|cancelled
	CancelledReason *string `json:"cancelled_reason,omitempty"`
	Notes           *string `json:"notes,omitempty"`
}

type resp struct {
	ID              string  `json:"id"`
	PatientID       string  `json:"patient_id"`
	KinesiologistID string  `json:"kinesiologist_id"`
	StartAt         string  `json:"start_at"`
	EndAt           string  `json:"end_at"`
	Status          string  `json:"status"`
	Notes           *string `json:"notes,omitempty"`
	CancelledReason *string `json:"cancelled_reason,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

func (h *Handler) Create(c *gin.Context) {
	if !isReceptionist(c.GetHeader("Authorization")) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	out, details, err := h.create.Execute(c.Request.Context(), usecase.CreateAppointmentInput{
		PatientID:       req.PatientID,
		KinesiologistID: req.KinesiologistID,
		StartAt:         req.StartAt,
		EndAt:           req.EndAt,
		Notes:           req.Notes,
	})

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": details})
		case errors.Is(err, domain.ErrOverlap):
			c.JSON(http.StatusConflict, gin.H{"error": "overlap"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusCreated, toResp(out))
}

func (h *Handler) ListDay(c *gin.Context) {
	// Para agenda normalmente también debería estar autenticado; lo dejamos abierto si querés.
	kid := c.Query("kinesiologist_id")
	date := c.Query("date")

	items, details, err := h.listDay.Execute(c.Request.Context(), kid, date)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": details})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	resp := make([]resp, 0, len(items))
	for _, it := range items {
		resp = append(resp, toResp(it))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Update(c *gin.Context) {
	if !isReceptionist(c.GetHeader("Authorization")) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req updateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	id := c.Param("id")
	out, details, err := h.update.Execute(c.Request.Context(), id, usecase.UpdateAppointmentInput{
		StartAt:         req.StartAt,
		EndAt:           req.EndAt,
		Status:          req.Status,
		CancelledReason: req.CancelledReason,
		Notes:           req.Notes,
	})

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": details})
		case errors.Is(err, domain.ErrOverlap):
			c.JSON(http.StatusConflict, gin.H{"error": "overlap"})
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusOK, toResp(out))
}

func toResp(a domain.Appointment) resp {
	return resp{
		ID:              a.ID.String(),
		PatientID:       a.PatientID.String(),
		KinesiologistID: a.KinesiologistID.String(),
		StartAt:         a.StartAt.UTC().Format(timeRFC3339()),
		EndAt:           a.EndAt.UTC().Format(timeRFC3339()),
		Status:          string(a.Status),
		Notes:           a.Notes,
		CancelledReason: a.CancelledReason,
		CreatedAt:       a.CreatedAt.UTC().Format(timeRFC3339()),
		UpdatedAt:       a.UpdatedAt.UTC().Format(timeRFC3339()),
	}
}

func timeRFC3339() string { return "2006-01-02T15:04:05Z07:00" }

func isReceptionist(auth string) bool {
	return strings.EqualFold(strings.TrimSpace(auth), "Bearer demo-recepcionista-token")
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")

	out, found, err := h.getByID.Execute(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}

	c.JSON(http.StatusOK, toResp(out))
}

func (h *Handler) ListByPatient(c *gin.Context) {
	patientID := c.Query("patient_id")
	from := c.Query("from")
	to := c.Query("to")

	items, details, err := h.listByPatient.Execute(
		c.Request.Context(),
		patientID,
		from,
		to,
	)

	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_error",
				"details": details,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	resp := make([]resp, 0, len(items))
	for _, it := range items {
		resp = append(resp, toResp(it))
	}
	c.JSON(http.StatusOK, resp)
}
