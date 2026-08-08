package usecase

import (
	"context"
	"errors"
	"time"

	"firebase.google.com/go/v4/auth"

	notificationDomain "github.com/javiacuna/kinesio-backend/internal/notifications/domain"
	patientsDomain "github.com/javiacuna/kinesio-backend/internal/patients/domain"
	patientsUC "github.com/javiacuna/kinesio-backend/internal/patients/usecase"
	"github.com/javiacuna/kinesio-backend/internal/patientsignups/domain"
	"github.com/javiacuna/kinesio-backend/internal/patientsignups/ports"
)

type ApproveSignupRequestUseCase struct {
	repo     ports.Repository
	patients patientMatcher
	register *patientsUC.RegisterPatientUseCase
	firebase *auth.Client
	notifier notificationCreator
}

func NewApproveSignupRequestUseCase(repo ports.Repository, patients patientMatcher, register *patientsUC.RegisterPatientUseCase, firebase *auth.Client, notifier notificationCreator) *ApproveSignupRequestUseCase {
	return &ApproveSignupRequestUseCase{repo: repo, patients: patients, register: register, firebase: firebase, notifier: notifier}
}

func (uc *ApproveSignupRequestUseCase) Execute(ctx context.Context, requestID, reviewerEmail string) (domain.SignupRequest, error) {
	if uc.firebase == nil {
		return domain.SignupRequest{}, errFirebaseNotConfigured
	}

	req, found, err := uc.repo.GetByID(ctx, requestID)
	if err != nil {
		return domain.SignupRequest{}, err
	}
	if !found {
		return domain.SignupRequest{}, domain.ErrRequestNotFound
	}
	if req.Status != domain.StatusPending {
		return domain.SignupRequest{}, domain.ErrAlreadyReviewed
	}

	matched, matchFound, err := uc.patients.FindByDNIAndEmail(ctx, req.DNI, req.Email)
	if err != nil {
		return domain.SignupRequest{}, err
	}

	var matchedPatient patientsDomain.Patient
	if matchFound {
		matchedPatient = matched
	} else {
		created, validationErrs, regErr := uc.register.Execute(ctx, patientsUC.RegisterPatientInput{
			DNI:       req.DNI,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
		})
		if regErr != nil {
			if errors.Is(regErr, patientsDomain.ErrDuplicateDNI) || errors.Is(regErr, patientsDomain.ErrDuplicateEmail) {
				retried, retryFound, retryErr := uc.patients.FindByDNIAndEmail(ctx, req.DNI, req.Email)
				if retryErr != nil {
					return domain.SignupRequest{}, retryErr
				}
				if !retryFound {
					return domain.SignupRequest{}, domain.ErrPatientConflict
				}
				matchedPatient = retried
			} else {
				_ = validationErrs
				return domain.SignupRequest{}, regErr
			}
		} else {
			matchedPatient = created
		}
	}

	if err := setRole(ctx, uc.firebase, req.FirebaseUID, "paciente"); err != nil {
		return domain.SignupRequest{}, err
	}

	now := time.Now().UTC()
	updated, err := uc.repo.UpdateStatus(ctx, requestID, domain.StatusApproved, &matchedPatient.ID, &reviewerEmail, now, nil)
	if err != nil {
		return domain.SignupRequest{}, err
	}

	entityType := "patient"
	_ = uc.notifier.Create(ctx, notificationDomain.NewNotification(notificationDomain.NewNotificationInput{
		RecipientEmail: req.Email,
		Type:           "patient_signup_approved",
		Title:          "Tu cuenta fue aprobada",
		Message:        "Ya podés ingresar al portal de pacientes con tu email y contraseña.",
		EntityType:     &entityType,
		EntityID:       &matchedPatient.ID,
	}))

	return updated, nil
}
