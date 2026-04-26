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
	WorkStartTime string
	WorkEndTime   string
	Active        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NormalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func NewKinesiologist(firstName, lastName, email string, licenseNumber *string, workStartTime, workEndTime string, active bool) Kinesiologist {
	return Kinesiologist{
		ID:            uuid.New(),
		FirstName:     strings.TrimSpace(firstName),
		LastName:      strings.TrimSpace(lastName),
		Email:         NormalizeEmail(email),
		LicenseNumber: trimOptional(licenseNumber),
		WorkStartTime: strings.TrimSpace(workStartTime),
		WorkEndTime:   strings.TrimSpace(workEndTime),
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
