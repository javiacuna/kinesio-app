package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/evolutions/domain"
)

type GetEvolutionByIDUseCase struct {
	repo domain.Repository
}

func NewGetEvolutionByIDUseCase(repo domain.Repository) *GetEvolutionByIDUseCase {
	return &GetEvolutionByIDUseCase{repo: repo}
}

func (uc *GetEvolutionByIDUseCase) Execute(ctx context.Context, id uuid.UUID) (domain.PatientEvolution, bool, error) {
	return uc.repo.GetByID(ctx, id)
}
