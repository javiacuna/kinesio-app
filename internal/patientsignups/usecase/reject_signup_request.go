package usecase

import (
	"context"
	"time"

	"firebase.google.com/go/v4/auth"

	notificationDomain "github.com/javiacuna/kinesio-backend/internal/notifications/domain"
	"github.com/javiacuna/kinesio-backend/internal/patientsignups/domain"
	"github.com/javiacuna/kinesio-backend/internal/patientsignups/ports"
)

type RejectSignupRequestUseCase struct {
	repo     ports.Repository
	firebase *auth.Client
	notifier notificationCreator
}

func NewRejectSignupRequestUseCase(repo ports.Repository, firebase *auth.Client, notifier notificationCreator) *RejectSignupRequestUseCase {
	return &RejectSignupRequestUseCase{repo: repo, firebase: firebase, notifier: notifier}
}

func (uc *RejectSignupRequestUseCase) Execute(ctx context.Context, requestID, reviewerEmail string, reason *string) (domain.SignupRequest, error) {
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

	if err := uc.firebase.DeleteUser(ctx, req.FirebaseUID); err != nil && !auth.IsUserNotFound(err) {
		return domain.SignupRequest{}, err
	}

	now := time.Now().UTC()
	updated, err := uc.repo.UpdateStatus(ctx, requestID, domain.StatusRejected, nil, &reviewerEmail, now, reason)
	if err != nil {
		return domain.SignupRequest{}, err
	}

	entityType := "patient_signup_request"
	_ = uc.notifier.Create(ctx, notificationDomain.NewNotification(notificationDomain.NewNotificationInput{
		RecipientEmail: req.Email,
		Type:           "patient_signup_rejected",
		Title:          "Tu solicitud fue rechazada",
		Message:        "No pudimos aprobar tu solicitud de acceso al portal. Podés volver a registrarte si creés que fue un error.",
		EntityType:     &entityType,
		EntityID:       &req.ID,
	}))

	return updated, nil
}
