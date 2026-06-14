package gorm

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/javiacuna/kinesio-backend/internal/support/domain"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error) {
	model := TicketModel{
		ID:                    ticket.ID,
		UserEmail:             strings.ToLower(strings.TrimSpace(ticket.UserEmail)),
		UserRole:              ticket.UserRole,
		Subject:               ticket.Subject,
		Message:               ticket.Message,
		AttachmentFileName:    ticket.AttachmentFileName,
		AttachmentContentType: ticket.AttachmentContentType,
		AttachmentSizeBytes:   ticket.AttachmentSizeBytes,
		AttachmentStoragePath: ticket.AttachmentStoragePath,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Ticket{}, err
	}
	return toDomain(model), nil
}

func (r *Repository) ListByUser(ctx context.Context, email string, limit int) ([]domain.Ticket, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}

	var models []TicketModel
	if err := r.db.WithContext(ctx).
		Where("user_email = ?", strings.ToLower(strings.TrimSpace(email))).
		Order("created_at DESC").
		Limit(limit).
		Find(&models).Error; err != nil {
		return nil, err
	}

	out := make([]domain.Ticket, 0, len(models))
	for _, model := range models {
		out = append(out, toDomain(model))
	}
	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (domain.Ticket, bool, error) {
	var model TicketModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.Ticket{}, false, nil
		}
		return domain.Ticket{}, false, err
	}
	return toDomain(model), true, nil
}

func toDomain(model TicketModel) domain.Ticket {
	return domain.Ticket{
		ID:                    model.ID,
		UserEmail:             model.UserEmail,
		UserRole:              model.UserRole,
		Subject:               model.Subject,
		Message:               model.Message,
		AttachmentFileName:    model.AttachmentFileName,
		AttachmentContentType: model.AttachmentContentType,
		AttachmentSizeBytes:   model.AttachmentSizeBytes,
		AttachmentStoragePath: model.AttachmentStoragePath,
		CreatedAt:             model.CreatedAt,
	}
}
