package gorm

import (
	"time"

	"github.com/google/uuid"
)

type NotificationModel struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;column:id"`
	RecipientEmail string     `gorm:"column:recipient_email"`
	RecipientRole  *string    `gorm:"column:recipient_role"`
	Type           string     `gorm:"column:type"`
	Title          string     `gorm:"column:title"`
	Message        string     `gorm:"column:message"`
	EntityType     *string    `gorm:"column:entity_type"`
	EntityID       *uuid.UUID `gorm:"type:uuid;column:entity_id"`
	ReadAt         *time.Time `gorm:"column:read_at"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (NotificationModel) TableName() string { return "notifications" }
