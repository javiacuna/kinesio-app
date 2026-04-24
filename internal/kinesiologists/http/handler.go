package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	"github.com/javiacuna/kinesio-backend/internal/kinesiologists/domain"
	"github.com/javiacuna/kinesio-backend/internal/kinesiologists/usecase"
)

type Handler struct {
	list *usecase.ListKinesiologistsUseCase
	save *usecase.SaveKinesiologistUseCase
}

func NewHandler(list *usecase.ListKinesiologistsUseCase, save *usecase.SaveKinesiologistUseCase) *Handler {
	return &Handler{list: list, save: save}
}

type saveReq struct {
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Email         string  `json:"email"`
	LicenseNumber *string `json:"license_number,omitempty"`
	Active        bool    `json:"active"`
}

type resp struct {
	ID            string  `json:"id"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Email         string  `json:"email"`
	LicenseNumber *string `json:"license_number,omitempty"`
	Active        bool    `json:"active"`
}

func (h *Handler) List(c *gin.Context) {
	onlyActive := true
	if v := strings.TrimSpace(c.Query("active")); v != "" {
		// active=true|false; si te pasan false, devolvemos todos
		onlyActive = strings.EqualFold(v, "true")
	}

	if user, ok := middleware.CurrentUser(c); ok && strings.EqualFold(user.Role, "kinesiologo") {
		item, found, err := h.list.FindByEmail(c.Request.Context(), user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
		if !found || !item.Active {
			c.JSON(http.StatusOK, []resp{})
			return
		}
		c.JSON(http.StatusOK, []resp{toResp(item)})
		return
	}

	items, err := h.list.Execute(c.Request.Context(), onlyActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]resp, 0, len(items))
	for _, k := range items {
		out = append(out, toResp(k))
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

	out, validation, err := h.save.Create(c.Request.Context(), usecase.SaveKinesiologistInput{
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Email:         req.Email,
		LicenseNumber: req.LicenseNumber,
		Active:        req.Active,
	})
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

	out, validation, err := h.save.Update(c.Request.Context(), usecase.SaveKinesiologistInput{
		ID:            c.Param("id"),
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Email:         req.Email,
		LicenseNumber: req.LicenseNumber,
		Active:        req.Active,
	})
	if err != nil {
		writeSaveError(c, validation, err)
		return
	}

	c.JSON(http.StatusOK, toResp(out))
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

func toResp(k domain.Kinesiologist) resp {
	return resp{
		ID:            k.ID.String(),
		FirstName:     k.FirstName,
		LastName:      k.LastName,
		Email:         k.Email,
		LicenseNumber: k.LicenseNumber,
		Active:        k.Active,
	}
}
