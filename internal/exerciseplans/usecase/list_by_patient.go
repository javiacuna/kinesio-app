package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/exerciseplans/domain"
)

type ListPlansByPatientUseCase struct {
	repo domain.Repository
}

func NewListPlansByPatientUseCase(repo domain.Repository) *ListPlansByPatientUseCase {
	return &ListPlansByPatientUseCase{repo: repo}
}

func (uc *ListPlansByPatientUseCase) Execute(ctx context.Context, patientID uuid.UUID) ([]domain.ExercisePlan, error) {
	return uc.repo.ListByPatient(ctx, patientID)
}
