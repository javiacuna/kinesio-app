package ports

import (
	"context"

	"github.com/javiacuna/kinesio-backend/internal/staff/domain"
)

type Repository interface {
	Create(ctx context.Context, member domain.StaffMember) (domain.StaffMember, error)
	Update(ctx context.Context, member domain.StaffMember) (domain.StaffMember, error)
	List(ctx context.Context, includeInactive bool) ([]domain.StaffMember, error)
}
