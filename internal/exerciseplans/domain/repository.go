package domain

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, p ExercisePlan) (ExercisePlan, error)
	GetByID(ctx context.Context, id uuid.UUID) (ExercisePlan, bool, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID) ([]ExercisePlan, error)
}
