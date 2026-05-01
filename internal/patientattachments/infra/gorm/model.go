package gorm

import "time"

type AttachmentModel struct {
	ID              string `gorm:"type:uuid;primaryKey"`
	PatientID       string `gorm:"type:uuid;not null"`
	FileName        string `gorm:"not null"`
	ContentType     string `gorm:"not null"`
	SizeBytes       int64  `gorm:"not null"`
	StoragePath     string `gorm:"not null"`
	Kind            string `gorm:"not null"`
	Category        string `gorm:"not null"`
	PatientVisible  bool   `gorm:"not null"`
	Notes           *string
	UploadedByEmail *string
	UploadedByRole  *string
	CreatedAt       time.Time
	UpdatedByEmail  *string
	UpdatedByRole   *string
	UpdatedAt       *time.Time
}

func (AttachmentModel) TableName() string { return "patient_attachments" }
