package firebase

import (
	"context"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

func NewAuthClient(ctx context.Context, projectID string, credentialsFile string) (*auth.Client, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, nil
	}

	opts := []option.ClientOption{}
	credentialsFile = strings.TrimSpace(credentialsFile)
	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}

	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID}, opts...)
	if err != nil {
		return nil, err
	}

	return app.Auth(ctx)
}
