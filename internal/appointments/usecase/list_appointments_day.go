package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/appointments/domain"
	"github.com/javiacuna/kinesio-backend/internal/appointments/ports"
)

type ListAppointmentsDayUseCase struct {
	repo ports.Repository
}

func NewListAppointmentsDayUseCase(repo ports.Repository) *ListAppointmentsDayUseCase {
	return &ListAppointmentsDayUseCase{repo: repo}
}

// date: YYYY-MM-DD; retorna [date 00:00, next day 00:00) en UTC (simple para empezar)
func (uc *ListAppointmentsDayUseCase) Execute(ctx context.Context, kinesiologistID string, date string) ([]domain.Appointment, map[string]string, error) {
	errs := map[string]string{}

	kid, err := uuid.Parse(strings.TrimSpace(kinesiologistID))
	if err != nil {
		errs["kinesiologist_id"] = "UUID inválido"
	}

	day, err := time.Parse("2006-01-02", strings.TrimSpace(date))
	if err != nil {
		errs["date"] = "Formato inválido (usar YYYY-MM-DD)"
	}

	if len(errs) > 0 {
		return nil, errs, domain.ErrValidation
	}

	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	items, err := uc.repo.ListByKinesiologistAndRange(ctx, kid, start, end)
	if err != nil {
		return nil, nil, err
	}
	return items, nil, nil
}
