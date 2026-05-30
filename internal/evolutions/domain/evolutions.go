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
	DiagnosisID     *uuid.UUID

	PainLevel       *int // 0..10
	MobilityScore   *int // 0..100
	StrengthScore   *int // 0..100
	FunctionalScore *int // 0..100
	Notes           string
	Photos          []PatientEvolutionPhoto

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
