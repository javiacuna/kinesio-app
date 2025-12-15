package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/appointments/domain"
	"github.com/javiacuna/kinesio-backend/internal/appointments/ports"
)

type CreateAppointmentInput struct {
	PatientID       string
	KinesiologistID string
	StartAt         string // RFC3339
	EndAt           string // RFC3339
	Notes           *string
}

type CreateAppointmentUseCase struct {
	repo ports.Repository
}

func NewCreateAppointmentUseCase(repo ports.Repository) *CreateAppointmentUseCase {
	return &CreateAppointmentUseCase{repo: repo}
}

func (uc *CreateAppointmentUseCase) Execute(ctx context.Context, in CreateAppointmentInput) (domain.Appointment, map[string]string, error) {
	errs := map[string]string{}

	pid, err := uuid.Parse(strings.TrimSpace(in.PatientID))
	if err != nil {
		errs["patient_id"] = "UUID inválido"
	}
	kid, err := uuid.Parse(strings.TrimSpace(in.KinesiologistID))
	if err != nil {
		errs["kinesiologist_id"] = "UUID inválido"
	}

	startAt, err := time.Parse(time.RFC3339, strings.TrimSpace(in.StartAt))
	if err != nil {
		errs["start_at"] = "Formato inválido (usar RFC3339)"
	}
	endAt, err := time.Parse(time.RFC3339, strings.TrimSpace(in.EndAt))
	if err != nil {
		errs["end_at"] = "Formato inválido (usar RFC3339)"
	}
	if err == nil && !endAt.After(startAt) {
		errs["end_at"] = "Debe ser mayor a start_at"
	}

	if len(errs) > 0 {
		return domain.Appointment{}, errs, domain.ErrValidation
	}

	overlap, err := uc.repo.HasOverlap(ctx, kid, startAt.UTC(), endAt.UTC(), nil)
	if err != nil {
		return domain.Appointment{}, nil, err
	}
	if overlap {
		return domain.Appointment{}, nil, domain.ErrOverlap
	}

	a := domain.Appointment{
		ID:              uuid.New(),
		PatientID:       pid,
		KinesiologistID: kid,
		StartAt:         startAt.UTC(),
		EndAt:           endAt.UTC(),
		Status:          domain.StatusScheduled,
		Notes:           trimPtr(in.Notes),
	}
	created, err := uc.repo.Create(ctx, a)
	if err != nil {
		return domain.Appointment{}, nil, err
	}
	return created, nil, nil
}

func trimPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}
