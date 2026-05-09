package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/kinesiologists/domain"
	"github.com/javiacuna/kinesio-backend/internal/kinesiologists/ports"
)

type ListSpecialtiesUseCase struct {
	repo ports.Repository
}

func NewListSpecialtiesUseCase(repo ports.Repository) *ListSpecialtiesUseCase {
	return &ListSpecialtiesUseCase{repo: repo}
}

func (uc *ListSpecialtiesUseCase) Execute(ctx context.Context, includeInactive bool) ([]domain.Specialty, error) {
	return uc.repo.ListSpecialties(ctx, includeInactive)
}

type SaveSpecialtyInput struct {
	ID     string
	Name   string
	Active bool
}

type SaveSpecialtyUseCase struct {
	repo ports.Repository
}

func NewSaveSpecialtyUseCase(repo ports.Repository) *SaveSpecialtyUseCase {
	return &SaveSpecialtyUseCase{repo: repo}
}

func (uc *SaveSpecialtyUseCase) Create(ctx context.Context, in SaveSpecialtyInput) (domain.Specialty, map[string]string, error) {
	specialty, errs := buildSpecialty(in)
	if len(errs) > 0 {
		return domain.Specialty{}, errs, domain.ErrValidation
	}
	specialty.ID = uuid.New()
	out, err := uc.repo.SaveSpecialty(ctx, specialty)
	return out, nil, err
}

func (uc *SaveSpecialtyUseCase) Update(ctx context.Context, in SaveSpecialtyInput) (domain.Specialty, map[string]string, error) {
	specialty, errs := buildSpecialty(in)
	id, err := uuid.Parse(strings.TrimSpace(in.ID))
	if err != nil {
		errs["id"] = "UUID inválido"
	}
	if len(errs) > 0 {
		return domain.Specialty{}, errs, domain.ErrValidation
	}
	specialty.ID = id
	out, err := uc.repo.SaveSpecialty(ctx, specialty)
	return out, nil, err
}

func buildSpecialty(in SaveSpecialtyInput) (domain.Specialty, map[string]string) {
	errs := map[string]string{}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		errs["name"] = "Campo obligatorio"
	}
	return domain.Specialty{Name: name, Active: in.Active}, errs
}

type ListPracticesUseCase struct {
	repo ports.Repository
}

func NewListPracticesUseCase(repo ports.Repository) *ListPracticesUseCase {
	return &ListPracticesUseCase{repo: repo}
}

func (uc *ListPracticesUseCase) Execute(ctx context.Context, includeInactive bool) ([]domain.Practice, error) {
	return uc.repo.ListPractices(ctx, includeInactive)
}

type SavePracticeInput struct {
	ID          string
	SpecialtyID string
	Name        string
	Description *string
	Active      bool
}

type SavePracticeUseCase struct {
	repo ports.Repository
}

func NewSavePracticeUseCase(repo ports.Repository) *SavePracticeUseCase {
	return &SavePracticeUseCase{repo: repo}
}

func (uc *SavePracticeUseCase) Create(ctx context.Context, in SavePracticeInput) (domain.Practice, map[string]string, error) {
	practice, errs := buildPractice(in)
	if len(errs) > 0 {
		return domain.Practice{}, errs, domain.ErrValidation
	}
	practice.ID = uuid.New()
	out, err := uc.repo.SavePractice(ctx, practice)
	return out, nil, err
}

func (uc *SavePracticeUseCase) Update(ctx context.Context, in SavePracticeInput) (domain.Practice, map[string]string, error) {
	practice, errs := buildPractice(in)
	id, err := uuid.Parse(strings.TrimSpace(in.ID))
	if err != nil {
		errs["id"] = "UUID inválido"
	}
	if len(errs) > 0 {
		return domain.Practice{}, errs, domain.ErrValidation
	}
	practice.ID = id
	out, err := uc.repo.SavePractice(ctx, practice)
	return out, nil, err
}

func buildPractice(in SavePracticeInput) (domain.Practice, map[string]string) {
	errs := map[string]string{}
	specialtyID, err := uuid.Parse(strings.TrimSpace(in.SpecialtyID))
	if err != nil {
		errs["specialty_id"] = "UUID inválido"
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		errs["name"] = "Campo obligatorio"
	}
	return domain.Practice{
		SpecialtyID: specialtyID,
		Name:        name,
		Description: trimOptionalString(in.Description),
		Active:      in.Active,
	}, errs
}

func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
