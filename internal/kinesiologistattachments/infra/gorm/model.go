package gorm

import "time"

type AttachmentModel struct {
	ID              string `gorm:"type:uuid;primaryKey"`
	KinesiologistID string `gorm:"type:uuid;not null"`
	FileName        string `gorm:"not null"`
	ContentType     string `gorm:"not null"`
	SizeBytes       int64  `gorm:"not null"`
	StoragePath     string `gorm:"not null"`
	Kind            string `gorm:"not null"`
	Category        string `gorm:"not null"`
	Notes           *string
	UploadedByEmail *string
	UploadedByRole  *string
	CreatedAt       time.Time
}

func (AttachmentModel) TableName() string { return "kinesiologist_attachments" }
