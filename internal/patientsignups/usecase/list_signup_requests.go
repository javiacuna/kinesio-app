package usecase

import (
	"context"

	"github.com/javiacuna/kinesio-backend/internal/patientsignups/domain"
	"github.com/javiacuna/kinesio-backend/internal/patientsignups/ports"
)

type ListSignupRequestsUseCase struct {
	repo ports.Repository
}

func NewListSignupRequestsUseCase(repo ports.Repository) *ListSignupRequestsUseCase {
	return &ListSignupRequestsUseCase{repo: repo}
}

func (uc *ListSignupRequestsUseCase) Execute(ctx context.Context, status string) ([]domain.SignupRequest, error) {
	return uc.repo.List(ctx, status)
}
