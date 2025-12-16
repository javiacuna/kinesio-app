package domain

import (
	"time"

	"github.com/google/uuid"
)

type Material struct {
	ID           uuid.UUID
	Name         string
	Description  *string
	TotalQty     int
	AvailableQty int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type MaterialLoan struct {
	ID              uuid.UUID
	MaterialID      uuid.UUID
	PatientID       uuid.UUID
	KinesiologistID uuid.UUID
	Qty             int
	Notes           *string
	LoanedAt        time.Time
	ReturnedAt      *time.Time
}
