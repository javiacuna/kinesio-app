package ports

import (
	"context"

	"github.com/javiacuna/kinesio-backend/internal/kinesiologists/domain"
)

type Repository interface {
	Create(ctx context.Context, k domain.Kinesiologist) (domain.Kinesiologist, error)
	Update(ctx context.Context, k domain.Kinesiologist) (domain.Kinesiologist, error)
	List(ctx context.Context, onlyActive bool) ([]domain.Kinesiologist, error)
	FindByEmail(ctx context.Context, email string) (domain.Kinesiologist, bool, error)
	GetByID(ctx context.Context, id string) (domain.Kinesiologist, bool, error)
	ListSpecialties(ctx context.Context, includeInactive bool) ([]domain.Specialty, error)
	SaveSpecialty(ctx context.Context, specialty domain.Specialty) (domain.Specialty, error)
	ListPractices(ctx context.Context, includeInactive bool) ([]domain.Practice, error)
	SavePractice(ctx context.Context, practice domain.Practice) (domain.Practice, error)
}
