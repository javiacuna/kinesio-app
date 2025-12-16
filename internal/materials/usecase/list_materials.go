package usecase

import (
	"context"

	"github.com/javiacuna/kinesio-backend/internal/materials/domain"
)

type ListMaterialsUseCase struct {
	repo domain.Repository
}

func NewListMaterialsUseCase(repo domain.Repository) *ListMaterialsUseCase {
	return &ListMaterialsUseCase{repo: repo}
}

func (uc *ListMaterialsUseCase) Execute(ctx context.Context, limit int) ([]domain.Material, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return uc.repo.ListMaterials(ctx, limit)
}
