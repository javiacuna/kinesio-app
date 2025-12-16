package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/exerciseplans/domain"
)

type GetPlanByIDUseCase struct {
	repo domain.Repository
}

func NewGetPlanByIDUseCase(repo domain.Repository) *GetPlanByIDUseCase {
	return &GetPlanByIDUseCase{repo: repo}
}

func (uc *GetPlanByIDUseCase) Execute(ctx context.Context, id uuid.UUID) (domain.ExercisePlan, bool, error) {
	return uc.repo.GetByID(ctx, id)
}
