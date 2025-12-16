package domain

import (
	"time"

	"github.com/google/uuid"
)

type Frequency string

const (
	FrequencyDaily  Frequency = "daily"
	FrequencyWeekly Frequency = "weekly"
)

type PlanStatus string

const (
	PlanActive PlanStatus = "active"
	PlanClosed PlanStatus = "closed"
)

type ExercisePlan struct {
	ID              uuid.UUID
	PatientID       uuid.UUID
	KinesiologistID uuid.UUID

	Frequency     Frequency
	DurationWeeks int
	Observations  *string
	Status        PlanStatus

	Items     []ExercisePlanItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ExercisePlanItem struct {
	ID               uuid.UUID
	PlanID           uuid.UUID
	Name             string
	Description      *string
	VideoURL         *string
	GuideURL         *string
	EstimatedMinutes int
	Sets             *int
	Reps             *int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
