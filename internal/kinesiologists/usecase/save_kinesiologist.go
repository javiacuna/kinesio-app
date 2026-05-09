package usecase

import (
	"context"
	"regexp"
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
	WorkStartTime string
	WorkEndTime   string
	WorkDays      []int
	PracticeIDs   []string
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
	workStartTime := defaultWorkTime(in.WorkStartTime, "08:00")
	workEndTime := defaultWorkTime(in.WorkEndTime, "20:00")
	practiceIDs := make([]uuid.UUID, 0, len(in.PracticeIDs))

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
	if !isHHmm(workStartTime) {
		errs["work_start_time"] = "Formato invalido"
	}
	if !isHHmm(workEndTime) {
		errs["work_end_time"] = "Formato invalido"
	}
	if isHHmm(workStartTime) && isHHmm(workEndTime) && workStartTime >= workEndTime {
		errs["work_end_time"] = "Debe ser mayor al horario de inicio"
	}

	workDays := domain.NormalizeWorkDays(in.WorkDays)
	if len(in.WorkDays) > 0 && len(workDays) == 0 {
		errs["work_days"] = "Seleccioná al menos un día laboral"
	}
	for _, day := range in.WorkDays {
		if day < 1 || day > 7 {
			errs["work_days"] = "Los días deben estar entre 1 y 7"
			break
		}
	}

	for _, id := range in.PracticeIDs {
		parsed, err := uuid.Parse(strings.TrimSpace(id))
		if err != nil {
			errs["practice_ids"] = "Hay prácticas con ID inválido"
			break
		}
		practiceIDs = append(practiceIDs, parsed)
	}

	return domain.NewKinesiologist(firstName, lastName, email, in.LicenseNumber, workStartTime, workEndTime, workDays, practiceIDs, in.Active), errs
}

var hhmmPattern = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

func defaultWorkTime(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func isHHmm(value string) bool {
	return hhmmPattern.MatchString(strings.TrimSpace(value))
}
