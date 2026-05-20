package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID             uuid.UUID
	RecipientEmail string
	RecipientRole  *string
	Type           string
	Title          string
	Message        string
	EntityType     *string
	EntityID       *uuid.UUID
	ReadAt         *time.Time
	CreatedAt      time.Time
}

type NewNotificationInput struct {
	RecipientEmail string
	RecipientRole  *string
	Type           string
	Title          string
	Message        string
	EntityType     *string
	EntityID       *uuid.UUID
}

func NewNotification(input NewNotificationInput) Notification {
	return Notification{
		ID:             uuid.New(),
		RecipientEmail: strings.ToLower(strings.TrimSpace(input.RecipientEmail)),
		RecipientRole:  trimOptional(input.RecipientRole),
		Type:           strings.TrimSpace(input.Type),
		Title:          strings.TrimSpace(input.Title),
		Message:        strings.TrimSpace(input.Message),
		EntityType:     trimOptional(input.EntityType),
		EntityID:       input.EntityID,
	}
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
