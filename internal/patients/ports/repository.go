package ports

import (
	"context"

	"github.com/javiacuna/kinesio-backend/internal/patients/domain"
)

type Repository interface {
	Create(ctx context.Context, p domain.Patient) (domain.Patient, error)
	Update(ctx context.Context, p domain.Patient) (domain.Patient, error)
	SetActive(ctx context.Context, id string, active bool) error
	ExistsByDNI(ctx context.Context, dni string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	GetByID(ctx context.Context, id string) (domain.Patient, bool, error)
	List(ctx context.Context, limit int, offset int, includeInactive bool) ([]domain.Patient, error)
	Search(ctx context.Context, query string, limit int, includeInactive bool) ([]domain.Patient, error)
}

// AppointmentCanceller cancela los turnos futuros de un paciente. La implementa
// el módulo de turnos (CancelFutureAppointmentsByPatientUseCase); se define acá
// como puerto para que patients no dependa directamente de appointments.
type AppointmentCanceller interface {
	Execute(ctx context.Context, patientID string, reason string) error
}
