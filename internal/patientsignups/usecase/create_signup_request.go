package usecase

import (
	"context"
	"strings"
	"time"

	"firebase.google.com/go/v4/auth"

	sharedDomain "github.com/javiacuna/kinesio-backend/internal/domain"
	patientsDomain "github.com/javiacuna/kinesio-backend/internal/patients/domain"
	"github.com/javiacuna/kinesio-backend/internal/patientsignups/domain"
	"github.com/javiacuna/kinesio-backend/internal/patientsignups/ports"
)

type patientMatcher interface {
	FindByDNIAndEmail(ctx context.Context, dni, email string) (patientsDomain.Patient, bool, error)
}

type CreateSignupRequestInput struct {
	Email     string
	Password  string
	DNI       string
	FirstName string
	LastName  string
}

type CreateSignupRequestUseCase struct {
	repo     ports.Repository
	patients patientMatcher
	firebase *auth.Client
}

func NewCreateSignupRequestUseCase(repo ports.Repository, patients patientMatcher, firebase *auth.Client) *CreateSignupRequestUseCase {
	return &CreateSignupRequestUseCase{repo: repo, patients: patients, firebase: firebase}
}

func (uc *CreateSignupRequestUseCase) Execute(ctx context.Context, in CreateSignupRequestInput) (domain.SignupRequest, map[string]string, error) {
	errs := map[string]string{}

	in.Email = strings.TrimSpace(in.Email)
	in.DNI = strings.TrimSpace(in.DNI)
	in.FirstName = strings.TrimSpace(in.FirstName)
	in.LastName = strings.TrimSpace(in.LastName)

	if in.Email == "" {
		errs["email"] = "Campo obligatorio"
	} else if !strings.Contains(in.Email, "@") {
		errs["email"] = "Formato inválido"
	}
	if sharedDomain.PasswordPolicyViolation(in.Password) != "" {
		errs["password"] = "Debe tener al menos 8 caracteres, con mayúscula, minúscula, número y carácter especial"
	}
	if in.DNI == "" {
		errs["dni"] = "Campo obligatorio"
	}
	for _, ch := range in.DNI {
		if ch < '0' || ch > '9' {
			errs["dni"] = "Debe ser numérico (sin puntos ni guiones)"
			break
		}
	}
	if in.FirstName == "" {
		errs["first_name"] = "Campo obligatorio"
	}
	if in.LastName == "" {
		errs["last_name"] = "Campo obligatorio"
	}

	if len(errs) > 0 {
		return domain.SignupRequest{}, errs, domain.ErrValidation
	}

	if uc.firebase == nil {
		return domain.SignupRequest{}, nil, errFirebaseNotConfigured
	}

	user, err := uc.firebase.CreateUser(ctx, (&auth.UserToCreate{}).
		Email(in.Email).
		Password(in.Password).
		EmailVerified(false).
		Disabled(false))
	if err != nil {
		if auth.IsEmailAlreadyExists(err) {
			return domain.SignupRequest{}, nil, domain.ErrEmailInUse
		}
		return domain.SignupRequest{}, nil, err
	}

	req := domain.NewSignupRequest(user.UID, in.DNI, in.FirstName, in.LastName, in.Email)

	matched, found, err := uc.patients.FindByDNIAndEmail(ctx, in.DNI, in.Email)
	if err != nil {
		return domain.SignupRequest{}, nil, err
	}
	if found {
		now := time.Now().UTC()
		req.Status = domain.StatusApproved
		req.MatchedPatientID = &matched.ID
		req.ReviewedAt = &now
	}

	created, err := uc.repo.Create(ctx, req)
	if err != nil {
		return domain.SignupRequest{}, nil, err
	}

	if found {
		if err := setRole(ctx, uc.firebase, user.UID, "paciente"); err != nil {
			return domain.SignupRequest{}, nil, err
		}
	}

	return created, nil, nil
}
