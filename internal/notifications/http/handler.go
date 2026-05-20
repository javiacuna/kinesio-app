package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	"github.com/javiacuna/kinesio-backend/internal/notifications/domain"
)

type Store interface {
	ListByRecipient(ctx context.Context, email string, limit int, unreadOnly bool) ([]domain.Notification, error)
	CountUnread(ctx context.Context, email string) (int64, error)
	MarkRead(ctx context.Context, id uuid.UUID, email string) error
	MarkAllRead(ctx context.Context, email string) error
}

type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

type notificationResp struct {
	ID             string  `json:"id"`
	RecipientEmail string  `json:"recipient_email"`
	RecipientRole  *string `json:"recipient_role,omitempty"`
	Type           string  `json:"type"`
	Title          string  `json:"title"`
	Message        string  `json:"message"`
	EntityType     *string `json:"entity_type,omitempty"`
	EntityID       *string `json:"entity_id,omitempty"`
	ReadAt         *string `json:"read_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

func (h *Handler) List(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok || strings.TrimSpace(user.Email) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(strings.TrimSpace(c.Query("limit")))
	unreadOnly := strings.EqualFold(strings.TrimSpace(c.Query("unread_only")), "true")
	items, err := h.store.ListByRecipient(c.Request.Context(), user.Email, limit, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	unreadCount, err := h.store.CountUnread(c.Request.Context(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]notificationResp, 0, len(items))
	for _, item := range items {
		out = append(out, toResp(item))
	}
	c.JSON(http.StatusOK, gin.H{
		"items":        out,
		"unread_count": unreadCount,
	})
}

func (h *Handler) UnreadCount(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok || strings.TrimSpace(user.Email) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	count, err := h.store.CountUnread(c.Request.Context(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

func (h *Handler) MarkRead(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok || strings.TrimSpace(user.Email) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(strings.TrimSpace(c.Param("notification_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_notification_id"})
		return
	}
	if err := h.store.MarkRead(c.Request.Context(), id, user.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) MarkAllRead(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok || strings.TrimSpace(user.Email) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.store.MarkAllRead(c.Request.Context(), user.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.Status(http.StatusNoContent)
}

func toResp(item domain.Notification) notificationResp {
	var entityID *string
	if item.EntityID != nil {
		value := item.EntityID.String()
		entityID = &value
	}
	var readAt *string
	if item.ReadAt != nil {
		value := item.ReadAt.UTC().Format(time.RFC3339)
		readAt = &value
	}
	return notificationResp{
		ID:             item.ID.String(),
		RecipientEmail: item.RecipientEmail,
		RecipientRole:  item.RecipientRole,
		Type:           item.Type,
		Title:          item.Title,
		Message:        item.Message,
		EntityType:     item.EntityType,
		EntityID:       entityID,
		ReadAt:         readAt,
		CreatedAt:      item.CreatedAt.UTC().Format(time.RFC3339),
	}
}
