package usecase

import (
	"context"

	"github.com/javiacuna/kinesio-backend/internal/patientsignups/domain"
	"github.com/javiacuna/kinesio-backend/internal/patientsignups/ports"
)

// GetMySignupStatusUseCase permite que un usuario ya autenticado (pero
// todavía sin rol asignado, porque su autorregistro sigue en revisión)
// consulte el estado de su propia solicitud, en vez de ver sólo la pantalla
// genérica de "acceso no configurado".
type GetMySignupStatusUseCase struct {
	repo ports.Repository
}

func NewGetMySignupStatusUseCase(repo ports.Repository) *GetMySignupStatusUseCase {
	return &GetMySignupStatusUseCase{repo: repo}
}

func (uc *GetMySignupStatusUseCase) Execute(ctx context.Context, email string) (domain.SignupRequest, bool, error) {
	return uc.repo.FindLatestByEmail(ctx, email)
}
