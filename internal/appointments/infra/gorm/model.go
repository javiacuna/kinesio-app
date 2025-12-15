package gorm

import (
	"time"

	"github.com/google/uuid"
)

type AppointmentModel struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;column:id"`
	PatientID       uuid.UUID `gorm:"type:uuid;column:patient_id;not null"`
	KinesiologistID uuid.UUID `gorm:"type:uuid;column:kinesiologist_id;not null"`
	StartAt         time.Time `gorm:"column:start_at;not null"`
	EndAt           time.Time `gorm:"column:end_at;not null"`
	Status          string    `gorm:"column:status;not null"`
	Notes           *string   `gorm:"column:notes"`
	CancelledReason *string   `gorm:"column:cancelled_reason"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (AppointmentModel) TableName() string { return "appointments" }
