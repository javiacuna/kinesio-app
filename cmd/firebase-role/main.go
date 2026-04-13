package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"

	authFirebase "github.com/javiacuna/kinesio-backend/internal/auth/firebase"
	"github.com/javiacuna/kinesio-backend/internal/config"
)

var validRoles = map[string]struct{}{
	"admin":         {},
	"recepcionista": {},
	"kinesiologo":   {},
	"paciente":      {},
}

func main() {
	_ = godotenv.Load()

	email := flag.String("email", "", "Firebase user email")
	uid := flag.String("uid", "", "Firebase user uid")
	role := flag.String("role", "", "Role: admin, recepcionista, kinesiologo")
	flag.Parse()

	selectedRole := strings.ToLower(strings.TrimSpace(*role))
	if _, ok := validRoles[selectedRole]; !ok {
		exitf("invalid role %q; use admin, recepcionista, kinesiologo or paciente", *role)
	}

	cfg := config.MustLoad()
	if cfg.FirebaseProjectID == "" {
		exitf("FIREBASE_PROJECT_ID is required")
	}

	ctx := context.Background()
	client, err := authFirebase.NewAuthClient(ctx, cfg.FirebaseProjectID, cfg.FirebaseCredentialsFile)
	if err != nil {
		exitf("init firebase auth: %v", err)
	}
	if client == nil {
		exitf("firebase auth client is not configured")
	}

	targetUID := strings.TrimSpace(*uid)
	if targetUID == "" {
		targetEmail := strings.TrimSpace(*email)
		if targetEmail == "" {
			exitf("provide -email or -uid")
		}
		user, err := client.GetUserByEmail(ctx, targetEmail)
		if err != nil {
			exitf("get firebase user by email: %v", err)
		}
		targetUID = user.UID
	}

	if err := client.SetCustomUserClaims(ctx, targetUID, map[string]any{"role": selectedRole}); err != nil {
		exitf("set custom user claims: %v", err)
	}

	fmt.Printf("role %q assigned to uid %s\n", selectedRole, targetUID)
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
