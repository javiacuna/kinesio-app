package domain

import (
	"time"

	"github.com/google/uuid"
)

type Ticket struct {
	ID                    uuid.UUID
	UserEmail             string
	UserRole              *string
	Subject               string
	Message               string
	AttachmentFileName    *string
	AttachmentContentType *string
	AttachmentSizeBytes   *int64
	AttachmentStoragePath *string
	CreatedAt             time.Time
}
