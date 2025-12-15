package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/javiacuna/kinesio-backend/internal/patients/domain"
	"github.com/javiacuna/kinesio-backend/internal/patients/ports"
)

type RegisterPatientInput struct {
	DNI           string
	FirstName     string
	LastName      string
	Email         string
	Phone         *string
	BirthDate     *string // YYYY-MM-DD (string) para parsear acá o en handler; acá lo parseamos.
	ClinicalNotes *string
}

type RegisterPatientUseCase struct {
	repo ports.Repository
}

func NewRegisterPatientUseCase(repo ports.Repository) *RegisterPatientUseCase {
	return &RegisterPatientUseCase{repo: repo}
}

func (uc *RegisterPatientUseCase) Execute(ctx context.Context, in RegisterPatientInput) (domain.Patient, map[string]string, error) {
	// Validación (reglas mínimas)
	errs := map[string]string{}

	in.DNI = strings.TrimSpace(in.DNI)
	in.FirstName = strings.TrimSpace(in.FirstName)
	in.LastName = strings.TrimSpace(in.LastName)
	in.Email = strings.TrimSpace(in.Email)

	if in.DNI == "" {
		errs["dni"] = "Campo obligatorio"
	}
	if in.FirstName == "" {
		errs["first_name"] = "Campo obligatorio"
	}
	if in.LastName == "" {
		errs["last_name"] = "Campo obligatorio"
	}
	if in.Email == "" {
		errs["email"] = "Campo obligatorio"
	} else if !strings.Contains(in.Email, "@") {
		errs["email"] = "Formato inválido"
	}

	for _, ch := range in.DNI {
		if ch < '0' || ch > '9' {
			errs["dni"] = "Debe ser numérico (sin puntos ni guiones)"
			break
		}
	}

	var birthDatePtr *time.Time
	if in.BirthDate != nil && strings.TrimSpace(*in.BirthDate) != "" {
		tm, e := time.Parse("2006-01-02", strings.TrimSpace(*in.BirthDate))
		if e != nil {
			errs["birth_date"] = "Formato inválido (usar YYYY-MM-DD)"
		} else {
			utc := tm.UTC()
			birthDatePtr = &utc
		}
	}

	if len(errs) > 0 {
		return domain.Patient{}, errs, domain.ErrValidation
	}

	// Chequeo unicidad
	exists, err := uc.repo.ExistsByDNI(ctx, in.DNI)
	if err != nil {
		return domain.Patient{}, nil, err
	}
	if exists {
		return domain.Patient{}, nil, domain.ErrDuplicateDNI
	}

	exists, err = uc.repo.ExistsByEmail(ctx, in.Email)
	if err != nil {
		return domain.Patient{}, nil, err
	}
	if exists {
		return domain.Patient{}, nil, domain.ErrDuplicateEmail
	}

	p := domain.NewPatient(in.DNI, in.FirstName, in.LastName, in.Email, in.Phone, birthDatePtr, in.ClinicalNotes)
	created, err := uc.repo.Create(ctx, p)
	if err != nil {
		return domain.Patient{}, nil, err
	}

	return created, nil, nil
}
