package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/materials/domain"
	"github.com/javiacuna/kinesio-backend/internal/materials/usecase"
)

type Handler struct {
	createMaterialUC *usecase.CreateMaterialUseCase
	listMaterialsUC  *usecase.ListMaterialsUseCase
	loanUC           *usecase.LoanMaterialUseCase
	returnUC         *usecase.ReturnMaterialUseCase
	listLoansUC      *usecase.ListLoansByPatientUseCase
}

func NewHandler(
	createMaterialUC *usecase.CreateMaterialUseCase,
	listMaterialsUC *usecase.ListMaterialsUseCase,
	loanUC *usecase.LoanMaterialUseCase,
	returnUC *usecase.ReturnMaterialUseCase,
	listLoansUC *usecase.ListLoansByPatientUseCase,
) *Handler {
	return &Handler{
		createMaterialUC: createMaterialUC,
		listMaterialsUC:  listMaterialsUC,
		loanUC:           loanUC,
		returnUC:         returnUC,
		listLoansUC:      listLoansUC,
	}
}

// ---------- Responses

type materialResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	TotalQty     int     `json:"total_qty"`
	AvailableQty int     `json:"available_qty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type loanResponse struct {
	ID              string  `json:"id"`
	MaterialID      string  `json:"material_id"`
	PatientID       string  `json:"patient_id"`
	KinesiologistID string  `json:"kinesiologist_id"`
	Qty             int     `json:"qty"`
	Notes           *string `json:"notes,omitempty"`
	LoanedAt        string  `json:"loaned_at"`
	ReturnedAt      *string `json:"returned_at,omitempty"`
}

func toMaterialResp(m domain.Material) materialResponse {
	return materialResponse{
		ID:           m.ID.String(),
		Name:         m.Name,
		Description:  m.Description,
		TotalQty:     m.TotalQty,
		AvailableQty: m.AvailableQty,
		CreatedAt:    m.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    m.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toLoanResp(l domain.MaterialLoan) loanResponse {
	var returned *string
	if l.ReturnedAt != nil {
		s := l.ReturnedAt.UTC().Format(time.RFC3339)
		returned = &s
	}
	return loanResponse{
		ID:              l.ID.String(),
		MaterialID:      l.MaterialID.String(),
		PatientID:       l.PatientID.String(),
		KinesiologistID: l.KinesiologistID.String(),
		Qty:             l.Qty,
		Notes:           l.Notes,
		LoanedAt:        l.LoanedAt.UTC().Format(time.RFC3339),
		ReturnedAt:      returned,
	}
}

// ---------- Handlers

func (h *Handler) CreateMaterial(c *gin.Context) {
	var req usecase.CreateMaterialInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	out, validation, err := h.createMaterialUC.Execute(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
			return
		case errors.Is(err, domain.ErrDuplicateName):
			c.JSON(http.StatusConflict, gin.H{"error": "duplicate_name"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
	}

	c.JSON(http.StatusCreated, toMaterialResp(out))
}

func (h *Handler) ListMaterials(c *gin.Context) {
	limit := 50
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			limit = n
		}
	}

	items, err := h.listMaterialsUC.Execute(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]materialResponse, 0, len(items))
	for _, m := range items {
		out = append(out, toMaterialResp(m))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) LoanMaterial(c *gin.Context) {
	var req usecase.LoanMaterialInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	out, validation, err := h.loanUC.Execute(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
			return
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "details": validation})
			return
		case errors.Is(err, domain.ErrInsufficientStock):
			c.JSON(http.StatusConflict, gin.H{"error": "insufficient_stock"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
	}

	c.JSON(http.StatusCreated, toLoanResp(out))
}

func (h *Handler) ReturnLoan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("loan_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_loan_id"})
		return
	}

	out, err := h.returnUC.Execute(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		case errors.Is(err, domain.ErrAlreadyReturned):
			c.JSON(http.StatusConflict, gin.H{"error": "already_returned"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
	}

	c.JSON(http.StatusOK, toLoanResp(out))
}

func (h *Handler) ListLoansByPatient(c *gin.Context) {
	pid, err := uuid.Parse(c.Param("patient_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_patient_id"})
		return
	}

	onlyActive := strings.EqualFold(strings.TrimSpace(c.Query("active")), "true")

	limit := 50
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			limit = n
		}
	}

	items, err := h.listLoansUC.Execute(c.Request.Context(), pid, onlyActive, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]loanResponse, 0, len(items))
	for _, l := range items {
		out = append(out, toLoanResp(l))
	}
	c.JSON(http.StatusOK, out)
}
