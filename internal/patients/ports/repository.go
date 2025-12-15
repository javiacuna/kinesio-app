package ports

import (
	"context"

	"github.com/javiacuna/kinesio-backend/internal/patients/domain"
)

type Repository interface {
	Create(ctx context.Context, p domain.Patient) (domain.Patient, error)
	ExistsByDNI(ctx context.Context, dni string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	GetByID(ctx context.Context, id string) (domain.Patient, bool, error)
	Search(ctx context.Context, query string, limit int) ([]domain.Patient, error)
}
