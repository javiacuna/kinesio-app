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

// date: YYYY-MM-DD; retorna [date 00:00, next day 00:00) en el huso horario
// de la clínica (America/Argentina/Cordoba), no en UTC. Un turno a las 22:15
// hora local cae después de las 00:00 UTC del día siguiente, así que usar
// medianoche UTC como límite lo mostraba en el día equivocado de la agenda
// (y podía disparar un falso "sin solapamiento" contra el día real).
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

	loc := clinicLocation()
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc)
	end := start.Add(24 * time.Hour)

	items, err := uc.repo.ListByKinesiologistAndRange(ctx, kid, start, end)
	if err != nil {
		return nil, nil, err
	}
	return items, nil, nil
}
