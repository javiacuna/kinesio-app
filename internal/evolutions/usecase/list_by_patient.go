package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/evolutions/domain"
)

type ListEvolutionsByPatientUseCase struct {
	repo domain.Repository
}

func NewListEvolutionsByPatientUseCase(repo domain.Repository) *ListEvolutionsByPatientUseCase {
	return &ListEvolutionsByPatientUseCase{repo: repo}
}

func (uc *ListEvolutionsByPatientUseCase) Execute(ctx context.Context, patientID uuid.UUID, limit int) ([]domain.PatientEvolution, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return uc.repo.ListByPatient(ctx, patientID, limit)
}
