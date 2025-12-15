package usecase

import (
	"context"

	"github.com/javiacuna/kinesio-backend/internal/patients/domain"
	"github.com/javiacuna/kinesio-backend/internal/patients/ports"
)

type GetPatientByIDUseCase struct {
	repo ports.Repository
}

func NewGetPatientByIDUseCase(repo ports.Repository) *GetPatientByIDUseCase {
	return &GetPatientByIDUseCase{repo: repo}
}

func (uc *GetPatientByIDUseCase) Execute(ctx context.Context, id string) (domain.Patient, bool, error) {
	return uc.repo.GetByID(ctx, id)
}
