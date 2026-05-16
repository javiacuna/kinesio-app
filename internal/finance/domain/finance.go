package domain

import (
	"time"

	"github.com/google/uuid"
)

type Financier struct {
	ID        uuid.UUID
	Name      string
	Kind      string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PracticeTariff struct {
	ID                uuid.UUID
	PracticeID        uuid.UUID
	FinancierID       uuid.UUID
	BillingValueCents int64
	CopayCents        int64
	Currency          string
	ValidFrom         time.Time
	ValidTo           *time.Time
	Active            bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type ProfessionalFeeRule struct {
	ID              uuid.UUID
	KinesiologistID uuid.UUID
	PracticeID      uuid.UUID
	RuleType        string
	FixedValueCents *int64
	Percentage      *float64
	Active          bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type FinancialMovement struct {
	ID                    uuid.UUID
	AppointmentID         uuid.UUID
	PatientID             uuid.UUID
	KinesiologistID       uuid.UUID
	PracticeID            uuid.UUID
	FinancierID           uuid.UUID
	TariffID              uuid.UUID
	FeeRuleID             *uuid.UUID
	BillingValueCents     int64
	CopayCents            int64
	PayerValueCents       int64
	ProfessionalFeeCents  int64
	CenterAmountCents     int64
	Status                string
	CollectionStatus      string
	ProfessionalPayStatus string
	CollectedAt           *time.Time
	ProfessionalPaidAt    *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type AppointmentSnapshot struct {
	ID              uuid.UUID
	PatientID       uuid.UUID
	KinesiologistID uuid.UUID
	PracticeID      *uuid.UUID
	FinancierID     *uuid.UUID
	Status          string
}
