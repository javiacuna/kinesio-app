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

	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	"github.com/javiacuna/kinesio-backend/internal/materials/domain"
	"github.com/javiacuna/kinesio-backend/internal/materials/usecase"
	notificationDomain "github.com/javiacuna/kinesio-backend/internal/notifications/domain"
	staffDomain "github.com/javiacuna/kinesio-backend/internal/staff/domain"
)

type Handler struct {
	createMaterialUC   *usecase.CreateMaterialUseCase
	updateMaterialUC   *usecase.UpdateMaterialUseCase
	listMaterialsUC    *usecase.ListMaterialsUseCase
	loanUC             *usecase.LoanMaterialUseCase
	returnUC           *usecase.ReturnMaterialUseCase
	listAllLoansUC     *usecase.ListLoansUseCase
	listPatientLoansUC *usecase.ListLoansByPatientUseCase
	materials          materialGetter
	staff              staffLister
	notifier           notificationStore
}

type materialGetter interface {
	GetMaterialByID(ctx context.Context, id uuid.UUID) (domain.Material, bool, error)
}

type staffLister interface {
	List(ctx context.Context, includeInactive bool) ([]staffDomain.StaffMember, error)
}

type notificationStore interface {
	Create(ctx context.Context, item notificationDomain.Notification) error
	ListByRecipient(ctx context.Context, email string, limit int, unreadOnly bool) ([]notificationDomain.Notification, error)
}

func NewHandler(
	createMaterialUC *usecase.CreateMaterialUseCase,
	updateMaterialUC *usecase.UpdateMaterialUseCase,
	listMaterialsUC *usecase.ListMaterialsUseCase,
	loanUC *usecase.LoanMaterialUseCase,
	returnUC *usecase.ReturnMaterialUseCase,
	listAllLoansUC *usecase.ListLoansUseCase,
	listPatientLoansUC *usecase.ListLoansByPatientUseCase,
	lookups ...any,
) *Handler {
	var materials materialGetter
	var staff staffLister
	var notifier notificationStore
	for _, lookup := range lookups {
		if repo, ok := lookup.(materialGetter); ok {
			materials = repo
		}
		if repo, ok := lookup.(staffLister); ok {
			staff = repo
		}
		if repo, ok := lookup.(notificationStore); ok {
			notifier = repo
		}
	}

	return &Handler{
		createMaterialUC:   createMaterialUC,
		updateMaterialUC:   updateMaterialUC,
		listMaterialsUC:    listMaterialsUC,
		loanUC:             loanUC,
		returnUC:           returnUC,
		listAllLoansUC:     listAllLoansUC,
		listPatientLoansUC: listPatientLoansUC,
		materials:          materials,
		staff:              staff,
		notifier:           notifier,
	}
}

// ---------- Responses

type materialResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Description    *string `json:"description,omitempty"`
	TotalQty       int     `json:"total_qty"`
	AvailableQty   int     `json:"available_qty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	CreatedByEmail *string `json:"created_by_email,omitempty"`
	CreatedByRole  *string `json:"created_by_role,omitempty"`
	UpdatedByEmail *string `json:"updated_by_email,omitempty"`
	UpdatedByRole  *string `json:"updated_by_role,omitempty"`
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
	LoanedByEmail   *string `json:"loaned_by_email,omitempty"`
	LoanedByRole    *string `json:"loaned_by_role,omitempty"`
	ReturnedByEmail *string `json:"returned_by_email,omitempty"`
	ReturnedByRole  *string `json:"returned_by_role,omitempty"`
}

func toMaterialResp(m domain.Material) materialResponse {
	return materialResponse{
		ID:             m.ID.String(),
		Name:           m.Name,
		Description:    m.Description,
		TotalQty:       m.TotalQty,
		AvailableQty:   m.AvailableQty,
		CreatedAt:      m.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      m.UpdatedAt.UTC().Format(time.RFC3339),
		CreatedByEmail: m.CreatedByEmail,
		CreatedByRole:  m.CreatedByRole,
		UpdatedByEmail: m.UpdatedByEmail,
		UpdatedByRole:  m.UpdatedByRole,
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
		LoanedByEmail:   l.LoanedByEmail,
		LoanedByRole:    l.LoanedByRole,
		ReturnedByEmail: l.ReturnedByEmail,
		ReturnedByRole:  l.ReturnedByRole,
	}
}

// ---------- Handlers

func (h *Handler) CreateMaterial(c *gin.Context) {
	var req usecase.CreateMaterialInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	req.ActorEmail = actorEmail(c)
	req.ActorRole = actorRole(c)

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

	h.notifyLowStock(c, out)
	c.JSON(http.StatusCreated, toMaterialResp(out))
}

func (h *Handler) UpdateMaterial(c *gin.Context) {
	var req usecase.UpdateMaterialInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	req.ActorEmail = actorEmail(c)
	req.ActorRole = actorRole(c)

	out, validation, err := h.updateMaterialUC.Execute(c.Request.Context(), c.Param("material_id"), req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
			return
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		case errors.Is(err, domain.ErrDuplicateName):
			c.JSON(http.StatusConflict, gin.H{"error": "duplicate_name"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
	}

	h.notifyLowStock(c, out)
	c.JSON(http.StatusOK, toMaterialResp(out))
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
	req.ActorEmail = actorEmail(c)
	req.ActorRole = actorRole(c)

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

	h.notifyLowStockAfterLoan(c, out)
	c.JSON(http.StatusCreated, toLoanResp(out))
}

func (h *Handler) ReturnLoan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("loan_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_loan_id"})
		return
	}

	out, err := h.returnUC.Execute(c.Request.Context(), id, actorEmail(c), actorRole(c))
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

func (h *Handler) ListLoans(c *gin.Context) {
	onlyActive := strings.EqualFold(strings.TrimSpace(c.Query("active")), "true")

	limit := 100
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			limit = n
		}
	}

	items, err := h.listAllLoansUC.Execute(c.Request.Context(), onlyActive, limit)
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

	items, err := h.listPatientLoansUC.Execute(c.Request.Context(), pid, onlyActive, limit)
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

func actorEmail(c *gin.Context) string {
	if user, ok := middleware.CurrentUser(c); ok {
		return user.Email
	}
	return ""
}

func actorRole(c *gin.Context) string {
	if user, ok := middleware.CurrentUser(c); ok {
		return user.Role
	}
	return ""
}

const lowStockThreshold = 1

type notificationRecipient struct {
	email string
	role  string
}

func (h *Handler) notifyLowStockAfterLoan(c *gin.Context, loan domain.MaterialLoan) {
	if h.materials == nil {
		return
	}
	material, found, err := h.materials.GetMaterialByID(c.Request.Context(), loan.MaterialID)
	if err != nil || !found {
		return
	}
	h.notifyLowStock(c, material)
}

func (h *Handler) notifyLowStock(c *gin.Context, material domain.Material) {
	if h.notifier == nil || !isLowStock(material) {
		return
	}

	recipients := h.notificationRecipients(c)
	if len(recipients) == 0 {
		return
	}

	entityType := "material"
	notificationType := "material_low_stock"
	title := "Stock bajo de material"
	message := "El material " + strings.TrimSpace(material.Name) + " tiene " + strconv.Itoa(material.AvailableQty) + " disponibles de " + strconv.Itoa(material.TotalQty) + "."

	for _, recipient := range recipients {
		if h.hasUnreadLowStockNotification(c.Request.Context(), recipient.email, material.ID) {
			continue
		}

		role := strings.TrimSpace(recipient.role)
		var recipientRole *string
		if role != "" {
			recipientRole = &role
		}

		_ = h.notifier.Create(c.Request.Context(), notificationDomain.NewNotification(notificationDomain.NewNotificationInput{
			RecipientEmail: recipient.email,
			RecipientRole:  recipientRole,
			Type:           notificationType,
			Title:          title,
			Message:        message,
			EntityType:     &entityType,
			EntityID:       &material.ID,
		}))
	}
}

func (h *Handler) notificationRecipients(c *gin.Context) []notificationRecipient {
	byEmail := map[string]notificationRecipient{}

	if h.staff != nil {
		members, err := h.staff.List(c.Request.Context(), false)
		if err == nil {
			for _, member := range members {
				if !member.Active || !canReceiveStockAlert(string(member.Role)) {
					continue
				}
				email := strings.ToLower(strings.TrimSpace(member.Email))
				if email == "" {
					continue
				}
				byEmail[email] = notificationRecipient{email: email, role: string(member.Role)}
			}
		}
	}

	if user, ok := middleware.CurrentUser(c); ok && canReceiveStockAlert(user.Role) {
		email := strings.ToLower(strings.TrimSpace(user.Email))
		if email != "" {
			byEmail[email] = notificationRecipient{email: email, role: strings.TrimSpace(user.Role)}
		}
	}

	out := make([]notificationRecipient, 0, len(byEmail))
	for _, recipient := range byEmail {
		out = append(out, recipient)
	}
	return out
}

func (h *Handler) hasUnreadLowStockNotification(ctx context.Context, email string, materialID uuid.UUID) bool {
	items, err := h.notifier.ListByRecipient(ctx, email, 100, true)
	if err != nil {
		return false
	}
	for _, item := range items {
		if item.EntityID == nil || *item.EntityID != materialID {
			continue
		}
		if item.EntityType != nil && *item.EntityType == "material" && item.Type == "material_low_stock" {
			return true
		}
	}
	return false
}

func isLowStock(material domain.Material) bool {
	return material.TotalQty > 0 && material.AvailableQty <= lowStockThreshold
}

func canReceiveStockAlert(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin", "recepcionista", "kinesiologo":
		return true
	default:
		return false
	}
}
