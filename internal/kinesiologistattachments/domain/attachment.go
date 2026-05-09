package domain

import (
	"time"

	"github.com/google/uuid"
)

type Attachment struct {
	ID              uuid.UUID
	KinesiologistID uuid.UUID
	FileName        string
	ContentType     string
	SizeBytes       int64
	StoragePath     string
	Kind            string
	Category        string
	Notes           *string
	UploadedByEmail *string
	UploadedByRole  *string
	CreatedAt       time.Time
}
