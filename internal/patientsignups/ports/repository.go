package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/patientsignups/domain"
)

type Repository interface {
	Create(ctx context.Context, r domain.SignupRequest) (domain.SignupRequest, error)
	GetByID(ctx context.Context, id string) (domain.SignupRequest, bool, error)
	List(ctx context.Context, status string) ([]domain.SignupRequest, error)
	UpdateStatus(ctx context.Context, id string, status domain.Status, matchedPatientID *uuid.UUID, reviewedByEmail *string, reviewedAt time.Time, rejectionReason *string) (domain.SignupRequest, error)
}
