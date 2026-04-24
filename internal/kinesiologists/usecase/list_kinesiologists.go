package usecase

import (
	"context"

	"github.com/javiacuna/kinesio-backend/internal/kinesiologists/domain"
	"github.com/javiacuna/kinesio-backend/internal/kinesiologists/ports"
)

type ListKinesiologistsUseCase struct {
	repo ports.Repository
}

func NewListKinesiologistsUseCase(repo ports.Repository) *ListKinesiologistsUseCase {
	return &ListKinesiologistsUseCase{repo: repo}
}

func (uc *ListKinesiologistsUseCase) Execute(ctx context.Context, onlyActive bool) ([]domain.Kinesiologist, error) {
	return uc.repo.List(ctx, onlyActive)
}

func (uc *ListKinesiologistsUseCase) FindByEmail(ctx context.Context, email string) (domain.Kinesiologist, bool, error) {
	return uc.repo.FindByEmail(ctx, email)
}
