package domain

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, e PatientEvolution) (PatientEvolution, error)
	GetByID(ctx context.Context, id uuid.UUID) (PatientEvolution, bool, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit int) ([]PatientEvolution, error)
}
