package usecase

import (
	"context"

	"github.com/javiacuna/kinesio-backend/internal/materials/domain"
)

type ListLoansUseCase struct {
	repo domain.Repository
}

func NewListLoansUseCase(repo domain.Repository) *ListLoansUseCase {
	return &ListLoansUseCase{repo: repo}
}

func (uc *ListLoansUseCase) Execute(ctx context.Context, onlyActive bool, limit int) ([]domain.MaterialLoan, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return uc.repo.ListLoans(ctx, onlyActive, limit)
}
