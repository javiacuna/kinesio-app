package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/appointments/domain"
	"github.com/javiacuna/kinesio-backend/internal/appointments/ports"
)

type UpdateAppointmentInput struct {
	StartAt         *string // RFC3339 (opcional)
	EndAt           *string // RFC3339 (opcional)
	Status          *string // "scheduled" | "cancelled" (opcional)
	CancelledReason *string // opcional
	Notes           *string // opcional
}

type UpdateAppointmentUseCase struct {
	repo ports.Repository
}

func NewUpdateAppointmentUseCase(repo ports.Repository) *UpdateAppointmentUseCase {
	return &UpdateAppointmentUseCase{repo: repo}
}

func (uc *UpdateAppointmentUseCase) Execute(ctx context.Context, id string, in UpdateAppointmentInput) (domain.Appointment, map[string]string, error) {
	errs := map[string]string{}

	aid, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		errs["id"] = "UUID inválido"
		return domain.Appointment{}, errs, domain.ErrValidation
	}

	current, found, err := uc.repo.GetByID(ctx, aid)
	if err != nil {
		return domain.Appointment{}, nil, err
	}
	if !found {
		return domain.Appointment{}, nil, domain.ErrNotFound
	}

	// Status
	if in.Status != nil {
		switch domain.Status(strings.TrimSpace(*in.Status)) {
		case domain.StatusScheduled, domain.StatusCancelled:
			current.Status = domain.Status(strings.TrimSpace(*in.Status))
		default:
			errs["status"] = "Valor inválido (scheduled|cancelled)"
		}
	}

	// Notes
	if in.Notes != nil {
		current.Notes = trimPtr(in.Notes)
	}
	// Cancelled reason
	if in.CancelledReason != nil {
		current.CancelledReason = trimPtr(in.CancelledReason)
	}

	// Reprogramación (si vienen start/end)
	var newStart = current.StartAt
	var newEnd = current.EndAt

	if in.StartAt != nil {
		tm, e := time.Parse(time.RFC3339, strings.TrimSpace(*in.StartAt))
		if e != nil {
			errs["start_at"] = "Formato inválido (RFC3339)"
		} else {
			newStart = tm.UTC()
		}
	}
	if in.EndAt != nil {
		tm, e := time.Parse(time.RFC3339, strings.TrimSpace(*in.EndAt))
		if e != nil {
			errs["end_at"] = "Formato inválido (RFC3339)"
		} else {
			newEnd = tm.UTC()
		}
	}
	if (in.StartAt != nil || in.EndAt != nil) && !newEnd.After(newStart) {
		errs["end_at"] = "Debe ser mayor a start_at"
	}

	if len(errs) > 0 {
		return domain.Appointment{}, errs, domain.ErrValidation
	}

	// Si se reprogramó, validar solapamiento (excluyéndose)
	if in.StartAt != nil || in.EndAt != nil {
		ex := current.ID
		overlap, err := uc.repo.HasOverlap(ctx, current.KinesiologistID, newStart, newEnd, &ex)
		if err != nil {
			return domain.Appointment{}, nil, err
		}
		if overlap {
			return domain.Appointment{}, nil, domain.ErrOverlap
		}
		current.StartAt = newStart
		current.EndAt = newEnd
	}

	updated, err := uc.repo.Update(ctx, current)
	if err != nil {
		return domain.Appointment{}, nil, err
	}
	return updated, nil, nil
}
