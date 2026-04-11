package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/appointments/domain"
	"github.com/javiacuna/kinesio-backend/internal/appointments/ports"
)

type CancelAppointmentInput struct {
	Reason *string
}

type CancelAppointmentUseCase struct {
	repo ports.Repository
}

func NewCancelAppointmentUseCase(repo ports.Repository) *CancelAppointmentUseCase {
	return &CancelAppointmentUseCase{repo: repo}
}

func (uc *CancelAppointmentUseCase) Execute(ctx context.Context, id string, in CancelAppointmentInput) (domain.Appointment, map[string]string, error) {
	errs := map[string]string{}

	aid, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		errs["id"] = "UUID invalido"
		return domain.Appointment{}, errs, domain.ErrValidation
	}

	current, found, err := uc.repo.GetByID(ctx, aid)
	if err != nil {
		return domain.Appointment{}, nil, err
	}
	if !found {
		return domain.Appointment{}, nil, domain.ErrNotFound
	}

	current.Status = domain.StatusCancelled
	current.CancelledReason = trimPtr(in.Reason)

	cancelled, err := uc.repo.Update(ctx, current)
	if err != nil {
		return domain.Appointment{}, nil, err
	}
	return cancelled, nil, nil
}
