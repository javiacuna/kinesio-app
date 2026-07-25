package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/appointments/domain"
	"github.com/javiacuna/kinesio-backend/internal/appointments/ports"
)

type CancelFutureAppointmentsByPatientUseCase struct {
	repo ports.Repository
}

func NewCancelFutureAppointmentsByPatientUseCase(repo ports.Repository) *CancelFutureAppointmentsByPatientUseCase {
	return &CancelFutureAppointmentsByPatientUseCase{repo: repo}
}

// Execute cancela todos los turnos programados de un paciente desde el momento
// de la ejecución hacia adelante. Se usa al dar de baja (desactivar) un paciente,
// para que no queden turnos agendados "colgados" para alguien inactivo.
func (uc *CancelFutureAppointmentsByPatientUseCase) Execute(ctx context.Context, patientID string, reason string) error {
	pid, err := uuid.Parse(patientID)
	if err != nil {
		return domain.ErrValidation
	}

	now := time.Now().UTC()
	farFuture := now.AddDate(10, 0, 0)

	appointments, err := uc.repo.ListByPatientAndRange(ctx, pid, now, farFuture)
	if err != nil {
		return err
	}

	for _, appointment := range appointments {
		if appointment.Status != domain.StatusScheduled {
			continue
		}
		appointment.Status = domain.StatusCancelled
		cancelledReason := reason
		appointment.CancelledReason = &cancelledReason
		if _, err := uc.repo.Update(ctx, appointment); err != nil {
			return err
		}
	}

	return nil
}
