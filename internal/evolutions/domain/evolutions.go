package domain

import (
	"time"

	"github.com/google/uuid"
)

type PatientEvolution struct {
	ID              uuid.UUID
	PatientID       uuid.UUID
	KinesiologistID uuid.UUID
	AppointmentID   *uuid.UUID

	PainLevel *int   // 0..10
	Notes     string // requerido

	CreatedAt time.Time
	UpdatedAt time.Time
}
