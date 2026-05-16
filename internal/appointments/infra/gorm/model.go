package gorm

import (
	"time"

	"github.com/google/uuid"
)

type AppointmentModel struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey;column:id"`
	PatientID            uuid.UUID  `gorm:"type:uuid;column:patient_id;not null"`
	KinesiologistID      uuid.UUID  `gorm:"type:uuid;column:kinesiologist_id;not null"`
	PracticeID           *uuid.UUID `gorm:"type:uuid;column:practice_id"`
	FinancierID          *uuid.UUID `gorm:"type:uuid;column:financier_id"`
	PackageID            *uuid.UUID `gorm:"type:uuid;column:package_id"`
	PackageSessionNumber *int       `gorm:"column:package_session_number"`
	StartAt              time.Time  `gorm:"column:start_at;not null"`
	EndAt                time.Time  `gorm:"column:end_at;not null"`
	Status               string     `gorm:"column:status;not null"`
	Notes                *string    `gorm:"column:notes"`
	CancelledReason      *string    `gorm:"column:cancelled_reason"`
	CreatedAt            time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt            time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (AppointmentModel) TableName() string { return "appointments" }

type AppointmentPackageModel struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;column:id"`
	PatientID       uuid.UUID  `gorm:"type:uuid;column:patient_id;not null"`
	KinesiologistID uuid.UUID  `gorm:"type:uuid;column:kinesiologist_id;not null"`
	PracticeID      *uuid.UUID `gorm:"type:uuid;column:practice_id"`
	FinancierID     *uuid.UUID `gorm:"type:uuid;column:financier_id"`
	SessionsCount   int        `gorm:"column:sessions_count;not null"`
	DurationMin     int        `gorm:"column:duration_min;not null"`
	StartDate       time.Time  `gorm:"column:start_date;not null"`
	StartTime       string     `gorm:"column:start_time;not null"`
	WeekdaysOnly    bool       `gorm:"column:weekdays_only;not null"`
	WorkDays        *string    `gorm:"column:work_days"`
	Notes           *string    `gorm:"column:notes"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (AppointmentPackageModel) TableName() string { return "appointment_packages" }
