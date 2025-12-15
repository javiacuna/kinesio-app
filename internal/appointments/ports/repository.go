package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/appointments/domain"
)

type Repository interface {
	Create(ctx context.Context, a domain.Appointment) (domain.Appointment, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Appointment, bool, error)
	Update(ctx context.Context, a domain.Appointment) (domain.Appointment, error)

	// Solapamiento por kinesiólogo. excludeID sirve para reprogramar sin chocarse consigo mismo.
	HasOverlap(ctx context.Context, kinesiologistID uuid.UUID, startAt, endAt time.Time, excludeID *uuid.UUID) (bool, error)

	// Agenda del día (filtrada por kinesiólogo). startDay inclusive, endDay exclusive.
	ListByKinesiologistAndRange(ctx context.Context, kinesiologistID uuid.UUID, startDay, endDay time.Time) ([]domain.Appointment, error)
}
