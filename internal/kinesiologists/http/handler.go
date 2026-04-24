package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	"github.com/javiacuna/kinesio-backend/internal/kinesiologists/domain"
	"github.com/javiacuna/kinesio-backend/internal/kinesiologists/usecase"
)

type Handler struct {
	list *usecase.ListKinesiologistsUseCase
}

func NewHandler(list *usecase.ListKinesiologistsUseCase) *Handler {
	return &Handler{list: list}
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
