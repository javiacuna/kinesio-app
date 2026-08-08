package usecase

import (
	"context"
	"errors"

	"firebase.google.com/go/v4/auth"

	notificationDomain "github.com/javiacuna/kinesio-backend/internal/notifications/domain"
)

var errFirebaseNotConfigured = errors.New("firebase auth not configured")

func setRole(ctx context.Context, client *auth.Client, uid, role string) error {
	user, err := client.GetUser(ctx, uid)
	if err != nil {
		return err
	}
	claims := map[string]any{}
	for k, v := range user.CustomClaims {
		claims[k] = v
	}
	claims["role"] = role
	return client.SetCustomUserClaims(ctx, uid, claims)
}

type notificationCreator interface {
	Create(ctx context.Context, item notificationDomain.Notification) error
}
