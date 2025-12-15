package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/appointments/domain"
	"github.com/javiacuna/kinesio-backend/internal/appointments/ports"
)

type ListAppointmentsByPatientUseCase struct {
	repo ports.Repository
}

func NewListAppointmentsByPatientUseCase(repo ports.Repository) *ListAppointmentsByPatientUseCase {
	return &ListAppointmentsByPatientUseCase{repo: repo}
}

func (uc *ListAppointmentsByPatientUseCase) Execute(
	ctx context.Context,
	patientID string,
	from string,
	to string,
) ([]domain.Appointment, map[string]string, error) {

	errs := map[string]string{}

	pid, err := uuid.Parse(strings.TrimSpace(patientID))
	if err != nil {
		errs["patient_id"] = "UUID inválido"
	}

	start, err := time.Parse(time.RFC3339, strings.TrimSpace(from))
	if err != nil {
		errs["from"] = "Formato inválido (RFC3339)"
	}

	end, err := time.Parse(time.RFC3339, strings.TrimSpace(to))
	if err != nil {
		errs["to"] = "Formato inválido (RFC3339)"
	}

	if len(errs) > 0 {
		return nil, errs, domain.ErrValidation
	}

	items, err := uc.repo.ListByPatientAndRange(
		ctx,
		pid,
		start.UTC(),
		end.UTC(),
	)
	if err != nil {
		return nil, nil, err
	}

	return items, nil, nil
}
