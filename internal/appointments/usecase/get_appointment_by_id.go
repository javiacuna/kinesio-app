package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/appointments/domain"
	"github.com/javiacuna/kinesio-backend/internal/appointments/ports"
)

type GetAppointmentByIDUseCase struct {
	repo ports.Repository
}

func NewGetAppointmentByIDUseCase(repo ports.Repository) *GetAppointmentByIDUseCase {
	return &GetAppointmentByIDUseCase{repo: repo}
}

func (uc *GetAppointmentByIDUseCase) Execute(
	ctx context.Context,
	id string,
) (domain.Appointment, bool, error) {

	aid, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return domain.Appointment{}, false, domain.ErrValidation
	}
	return uc.repo.GetByID(ctx, aid)
}
