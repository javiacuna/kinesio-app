package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type PatientCheckIn struct {
	ID              uuid.UUID
	PatientID       uuid.UUID
	PainLevel       *int
	MobilityScore   *int
	StrengthScore   *int
	FunctionalScore *int
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewPatientCheckIn(patientID uuid.UUID, painLevel *int, mobilityScore *int, strengthScore *int, functionalScore *int, notes string) PatientCheckIn {
	now := time.Now().UTC()
	return PatientCheckIn{
		ID:              uuid.New(),
		PatientID:       patientID,
		PainLevel:       painLevel,
		MobilityScore:   mobilityScore,
		StrengthScore:   strengthScore,
		FunctionalScore: functionalScore,
		Notes:           strings.TrimSpace(notes),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
