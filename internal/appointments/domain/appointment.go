package domain

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusScheduled Status = "scheduled"
	StatusCancelled Status = "cancelled"
	StatusCompleted Status = "completed"
)

type Appointment struct {
	ID                   uuid.UUID
	PatientID            uuid.UUID
	KinesiologistID      uuid.UUID
	PracticeID           *uuid.UUID
	FinancierID          *uuid.UUID
	PackageID            *uuid.UUID
	PackageSessionNumber *int
	StartAt              time.Time
	EndAt                time.Time
	Status               Status
	Notes                *string
	CancelledReason      *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type AppointmentPackage struct {
	ID              uuid.UUID
	PatientID       uuid.UUID
	KinesiologistID uuid.UUID
	PracticeID      *uuid.UUID
	FinancierID     *uuid.UUID
	SessionsCount   int
	DurationMin     int
	StartDate       time.Time
	StartTime       string
	WeekdaysOnly    bool
	WorkDays        []int
	Notes           *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
