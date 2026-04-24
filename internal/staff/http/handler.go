package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	"github.com/javiacuna/kinesio-backend/internal/staff/domain"
	"github.com/javiacuna/kinesio-backend/internal/staff/usecase"
)

type Handler struct {
	list *usecase.ListStaffMembersUseCase
	save *usecase.SaveStaffMemberUseCase
}

func NewHandler(list *usecase.ListStaffMembersUseCase, save *usecase.SaveStaffMemberUseCase) *Handler {
	return &Handler{list: list, save: save}
}

type saveReq struct {
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	Phone       *string `json:"phone,omitempty"`
	Active      bool    `json:"active"`
	FirebaseUID *string `json:"firebase_uid,omitempty"`
}

type resp struct {
	ID          string  `json:"id"`
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	Phone       *string `json:"phone,omitempty"`
	Active      bool    `json:"active"`
	FirebaseUID *string `json:"firebase_uid,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func (h *Handler) List(c *gin.Context) {
	if !middleware.HasRole(c, "admin") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	includeInactive := strings.EqualFold(strings.TrimSpace(c.Query("active")), "false")
	items, err := h.list.Execute(c.Request.Context(), includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]resp, 0, len(items))
	for _, item := range items {
		out = append(out, toResp(item))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) Create(c *gin.Context) {
	if !middleware.HasRole(c, "admin") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req saveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	out, validation, err := h.save.Create(c.Request.Context(), toInput("", req))
	if err != nil {
		writeSaveError(c, validation, err)
		return
	}
	c.JSON(http.StatusCreated, toResp(out))
}

func (h *Handler) Update(c *gin.Context) {
	if !middleware.HasRole(c, "admin") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req saveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	out, validation, err := h.save.Update(c.Request.Context(), toInput(c.Param("id"), req))
	if err != nil {
		writeSaveError(c, validation, err)
		return
	}
	c.JSON(http.StatusOK, toResp(out))
}

func toInput(id string, req saveReq) usecase.SaveStaffMemberInput {
	return usecase.SaveStaffMemberInput{
		ID:          id,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		Role:        req.Role,
		Phone:       req.Phone,
		Active:      req.Active,
		FirebaseUID: req.FirebaseUID,
	}
}

func writeSaveError(c *gin.Context, validation map[string]string, err error) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
	case errors.Is(err, domain.ErrDuplicateEmail):
		c.JSON(http.StatusConflict, gin.H{"error": "email_duplicado"})
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}

func toResp(member domain.StaffMember) resp {
	return resp{
		ID:          member.ID.String(),
		FirstName:   member.FirstName,
		LastName:    member.LastName,
		Email:       member.Email,
		Role:        string(member.Role),
		Phone:       member.Phone,
		Active:      member.Active,
		FirebaseUID: member.FirebaseUID,
		CreatedAt:   member.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   member.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}
