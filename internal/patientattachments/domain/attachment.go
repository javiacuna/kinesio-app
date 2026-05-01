package domain

import (
	"time"

	"github.com/google/uuid"
)

type Attachment struct {
	ID              uuid.UUID
	PatientID       uuid.UUID
	FileName        string
	ContentType     string
	SizeBytes       int64
	StoragePath     string
	Kind            string
	Category        string
	PatientVisible  bool
	Notes           *string
	UploadedByEmail *string
	UploadedByRole  *string
	CreatedAt       time.Time
	UpdatedByEmail  *string
	UpdatedByRole   *string
	UpdatedAt       *time.Time
}
