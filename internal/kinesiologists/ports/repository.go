package ports

import (
	"context"

	"github.com/javiacuna/kinesio-backend/internal/kinesiologists/domain"
)

type Repository interface {
	List(ctx context.Context, onlyActive bool) ([]domain.Kinesiologist, error)
}
