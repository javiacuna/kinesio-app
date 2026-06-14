package gorm

import (
	"time"

	"github.com/google/uuid"
)

type TicketModel struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey;column:id"`
	UserEmail             string    `gorm:"column:user_email"`
	UserRole              *string   `gorm:"column:user_role"`
	Subject               string    `gorm:"column:subject"`
	Message               string    `gorm:"column:message"`
	AttachmentFileName    *string   `gorm:"column:attachment_file_name"`
	AttachmentContentType *string   `gorm:"column:attachment_content_type"`
	AttachmentSizeBytes   *int64    `gorm:"column:attachment_size_bytes"`
	AttachmentStoragePath *string   `gorm:"column:attachment_storage_path"`
	CreatedAt             time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (TicketModel) TableName() string { return "support_tickets" }
