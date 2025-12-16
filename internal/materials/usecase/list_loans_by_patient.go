package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/materials/domain"
)

type ListLoansByPatientUseCase struct {
	repo domain.Repository
}

func NewListLoansByPatientUseCase(repo domain.Repository) *ListLoansByPatientUseCase {
	return &ListLoansByPatientUseCase{repo: repo}
}

func (uc *ListLoansByPatientUseCase) Execute(ctx context.Context, patientID uuid.UUID, onlyActive bool, limit int) ([]domain.MaterialLoan, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return uc.repo.ListLoansByPatient(ctx, patientID, onlyActive, limit)
}
