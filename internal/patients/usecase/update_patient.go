package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/patients/domain"
	"github.com/javiacuna/kinesio-backend/internal/patients/ports"
)

type UpdatePatientInput struct {
	ID            string
	DNI           string
	FirstName     string
	LastName      string
	Email         string
	Phone         *string
	BirthDate     *string
	ClinicalNotes *string
}

type UpdatePatientUseCase struct {
	repo ports.Repository
}

func NewUpdatePatientUseCase(repo ports.Repository) *UpdatePatientUseCase {
	return &UpdatePatientUseCase{repo: repo}
}

func (uc *UpdatePatientUseCase) Execute(ctx context.Context, in UpdatePatientInput) (domain.Patient, map[string]string, error) {
	errs := map[string]string{}

	id, err := uuid.Parse(strings.TrimSpace(in.ID))
	if err != nil {
		errs["id"] = "UUID invalido"
	}

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
		errs["email"] = "Formato invalido"
	}

	for _, ch := range in.DNI {
		if ch < '0' || ch > '9' {
			errs["dni"] = "Debe ser numerico (sin puntos ni guiones)"
			break
		}
	}

	var birthDatePtr *time.Time
	if in.BirthDate != nil && strings.TrimSpace(*in.BirthDate) != "" {
		tm, e := time.Parse("2006-01-02", strings.TrimSpace(*in.BirthDate))
		if e != nil {
			errs["birth_date"] = "Formato invalido (usar YYYY-MM-DD)"
		} else {
			utc := tm.UTC()
			birthDatePtr = &utc
		}
	}

	if len(errs) > 0 {
		return domain.Patient{}, errs, domain.ErrValidation
	}

	current, found, err := uc.repo.GetByID(ctx, id.String())
	if err != nil {
		return domain.Patient{}, nil, err
	}
	if !found {
		return domain.Patient{}, nil, domain.ErrNotFound
	}

	if current.DNI != in.DNI {
		exists, err := uc.repo.ExistsByDNI(ctx, in.DNI)
		if err != nil {
			return domain.Patient{}, nil, err
		}
		if exists {
			return domain.Patient{}, nil, domain.ErrDuplicateDNI
		}
	}

	if !strings.EqualFold(current.Email, in.Email) {
		exists, err := uc.repo.ExistsByEmail(ctx, in.Email)
		if err != nil {
			return domain.Patient{}, nil, err
		}
		if exists {
			return domain.Patient{}, nil, domain.ErrDuplicateEmail
		}
	}

	current.DNI = in.DNI
	current.FirstName = in.FirstName
	current.LastName = in.LastName
	current.Email = strings.ToLower(in.Email)
	current.Phone = domain.TrimOptionalString(in.Phone)
	current.BirthDate = birthDatePtr
	current.ClinicalNotes = domain.TrimOptionalString(in.ClinicalNotes)

	updated, err := uc.repo.Update(ctx, current)
	if err != nil {
		return domain.Patient{}, nil, err
	}

	return updated, nil, nil
}
