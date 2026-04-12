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
	Photos    []PatientEvolutionPhoto

	CreatedAt time.Time
	UpdatedAt time.Time
}

type PatientEvolutionPhoto struct {
	ID          uuid.UUID
	EvolutionID uuid.UUID
	URL         string
	Caption     *string
	CreatedAt   time.Time
}
