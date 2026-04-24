package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/patients/domain"
	"github.com/javiacuna/kinesio-backend/internal/patients/ports"
)

type DeletePatientUseCase struct {
	repo ports.Repository
}

func NewDeletePatientUseCase(repo ports.Repository) *DeletePatientUseCase {
	return &DeletePatientUseCase{repo: repo}
}

func (uc *DeletePatientUseCase) Execute(ctx context.Context, id string) (map[string]string, error) {
	trimmedID := strings.TrimSpace(id)
	if _, err := uuid.Parse(trimmedID); err != nil {
		return map[string]string{"id": "UUID invalido"}, domain.ErrValidation
	}

	if err := uc.repo.SetActive(ctx, trimmedID, false); err != nil {
		return nil, err
	}
	return nil, nil
}
