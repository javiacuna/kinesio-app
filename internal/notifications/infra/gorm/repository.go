package gorm

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/notifications/domain"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, item domain.Notification) error {
	model := NotificationModel{
		ID:             item.ID,
		RecipientEmail: strings.ToLower(strings.TrimSpace(item.RecipientEmail)),
		RecipientRole:  item.RecipientRole,
		Type:           item.Type,
		Title:          item.Title,
		Message:        item.Message,
		EntityType:     item.EntityType,
		EntityID:       item.EntityID,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *Repository) ListByRecipient(ctx context.Context, email string, limit int, unreadOnly bool) ([]domain.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}

	var models []NotificationModel
	query := r.db.WithContext(ctx).
		Where("recipient_email = ?", strings.ToLower(strings.TrimSpace(email))).
		Order("created_at DESC").
		Limit(limit)
	if unreadOnly {
		query = query.Where("read_at IS NULL")
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	out := make([]domain.Notification, 0, len(models))
	for _, model := range models {
		out = append(out, toDomain(model))
	}
	return out, nil
}

func (r *Repository) CountUnread(ctx context.Context, email string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&NotificationModel{}).
		Where("recipient_email = ? AND read_at IS NULL", strings.ToLower(strings.TrimSpace(email))).
		Count(&count).
		Error
	return count, err
}

func (r *Repository) MarkRead(ctx context.Context, id uuid.UUID, email string) error {
	return r.db.WithContext(ctx).
		Model(&NotificationModel{}).
		Where("id = ? AND recipient_email = ?", id, strings.ToLower(strings.TrimSpace(email))).
		Where("read_at IS NULL").
		Update("read_at", time.Now().UTC()).
		Error
}

func (r *Repository) MarkAllRead(ctx context.Context, email string) error {
	return r.db.WithContext(ctx).
		Model(&NotificationModel{}).
		Where("recipient_email = ? AND read_at IS NULL", strings.ToLower(strings.TrimSpace(email))).
		Update("read_at", time.Now().UTC()).
		Error
}

func toDomain(model NotificationModel) domain.Notification {
	return domain.Notification{
		ID:             model.ID,
		RecipientEmail: model.RecipientEmail,
		RecipientRole:  model.RecipientRole,
		Type:           model.Type,
		Title:          model.Title,
		Message:        model.Message,
		EntityType:     model.EntityType,
		EntityID:       model.EntityID,
		ReadAt:         model.ReadAt,
		CreatedAt:      model.CreatedAt,
	}
}
