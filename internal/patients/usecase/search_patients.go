package usecase

import (
	"context"
	"strings"

	"github.com/javiacuna/kinesio-backend/internal/patients/domain"
	"github.com/javiacuna/kinesio-backend/internal/patients/ports"
)

type SearchPatientsUseCase struct {
	repo ports.Repository
}

func NewSearchPatientsUseCase(repo ports.Repository) *SearchPatientsUseCase {
	return &SearchPatientsUseCase{repo: repo}
}

func (uc *SearchPatientsUseCase) Execute(ctx context.Context, query string, limit int) ([]domain.Patient, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return []domain.Patient{}, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	return uc.repo.Search(ctx, q, limit)
}
