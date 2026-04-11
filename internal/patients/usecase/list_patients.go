package usecase

import (
	"context"

	"github.com/javiacuna/kinesio-backend/internal/patients/domain"
	"github.com/javiacuna/kinesio-backend/internal/patients/ports"
)

type ListPatientsUseCase struct {
	repo ports.Repository
}

func NewListPatientsUseCase(repo ports.Repository) *ListPatientsUseCase {
	return &ListPatientsUseCase{repo: repo}
}

func (uc *ListPatientsUseCase) Execute(ctx context.Context, limit int, offset int) ([]domain.Patient, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.repo.List(ctx, limit, offset)
}
