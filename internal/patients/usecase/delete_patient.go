package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/patients/domain"
	"github.com/javiacuna/kinesio-backend/internal/patients/ports"
)

const patientDeactivatedReason = "Paciente dado de baja"

type DeletePatientUseCase struct {
	repo                 ports.Repository
	appointmentCanceller ports.AppointmentCanceller
}

func NewDeletePatientUseCase(repo ports.Repository, appointmentCanceller ports.AppointmentCanceller) *DeletePatientUseCase {
	return &DeletePatientUseCase{repo: repo, appointmentCanceller: appointmentCanceller}
}

func (uc *DeletePatientUseCase) Execute(ctx context.Context, id string) (map[string]string, error) {
	trimmedID := strings.TrimSpace(id)
	if _, err := uuid.Parse(trimmedID); err != nil {
		return map[string]string{"id": "UUID invalido"}, domain.ErrValidation
	}

	if err := uc.repo.SetActive(ctx, trimmedID, false); err != nil {
		return nil, err
	}

	if uc.appointmentCanceller != nil {
		if err := uc.appointmentCanceller.Execute(ctx, trimmedID, patientDeactivatedReason); err != nil {
			return nil, err
		}
	}

	return nil, nil
}
