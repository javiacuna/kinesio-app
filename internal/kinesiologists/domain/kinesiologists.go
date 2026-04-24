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

func NewKinesiologist(firstName, lastName, email string, licenseNumber *string, active bool) Kinesiologist {
	return Kinesiologist{
		ID:            uuid.New(),
		FirstName:     strings.TrimSpace(firstName),
		LastName:      strings.TrimSpace(lastName),
		Email:         NormalizeEmail(email),
		LicenseNumber: trimOptional(licenseNumber),
		Active:        active,
	}
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
