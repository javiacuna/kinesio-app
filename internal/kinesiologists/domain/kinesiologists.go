package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Kinesiologist struct {
	ID            uuid.UUID
	FirstName     string
	LastName      string
	Email         string
	LicenseNumber *string
	Active        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NormalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
