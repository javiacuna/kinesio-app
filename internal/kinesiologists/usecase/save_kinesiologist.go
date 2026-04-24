package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/kinesiologists/domain"
	"github.com/javiacuna/kinesio-backend/internal/kinesiologists/ports"
)

type SaveKinesiologistInput struct {
	ID            string
	FirstName     string
	LastName      string
	Email         string
	LicenseNumber *string
	Active        bool
}

type SaveKinesiologistUseCase struct {
	repo ports.Repository
}

func NewSaveKinesiologistUseCase(repo ports.Repository) *SaveKinesiologistUseCase {
	return &SaveKinesiologistUseCase{repo: repo}
}

func (uc *SaveKinesiologistUseCase) Create(ctx context.Context, in SaveKinesiologistInput) (domain.Kinesiologist, map[string]string, error) {
	k, errs := buildKinesiologist(in)
	if len(errs) > 0 {
		return domain.Kinesiologist{}, errs, domain.ErrValidation
	}

	created, err := uc.repo.Create(ctx, k)
	return created, nil, err
}

func (uc *SaveKinesiologistUseCase) Update(ctx context.Context, in SaveKinesiologistInput) (domain.Kinesiologist, map[string]string, error) {
	k, errs := buildKinesiologist(in)
	id, err := uuid.Parse(strings.TrimSpace(in.ID))
	if err != nil {
		errs["id"] = "UUID invalido"
	}
	if len(errs) > 0 {
		return domain.Kinesiologist{}, errs, domain.ErrValidation
	}

	k.ID = id
	updated, err := uc.repo.Update(ctx, k)
	return updated, nil, err
}

func buildKinesiologist(in SaveKinesiologistInput) (domain.Kinesiologist, map[string]string) {
	errs := map[string]string{}
	firstName := strings.TrimSpace(in.FirstName)
	lastName := strings.TrimSpace(in.LastName)
	email := domain.NormalizeEmail(in.Email)

	if firstName == "" {
		errs["first_name"] = "Campo obligatorio"
	}
	if lastName == "" {
		errs["last_name"] = "Campo obligatorio"
	}
	if email == "" {
		errs["email"] = "Campo obligatorio"
	} else if !strings.Contains(email, "@") {
		errs["email"] = "Formato invalido"
	}

	return domain.NewKinesiologist(firstName, lastName, email, in.LicenseNumber, in.Active), errs
}
