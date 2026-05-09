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
	WorkDays      []int
	Active        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NormalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func NewKinesiologist(firstName, lastName, email string, licenseNumber *string, workStartTime, workEndTime string, workDays []int, active bool) Kinesiologist {
	return Kinesiologist{
		ID:            uuid.New(),
		FirstName:     strings.TrimSpace(firstName),
		LastName:      strings.TrimSpace(lastName),
		Email:         NormalizeEmail(email),
		LicenseNumber: trimOptional(licenseNumber),
		WorkStartTime: strings.TrimSpace(workStartTime),
		WorkEndTime:   strings.TrimSpace(workEndTime),
		WorkDays:      NormalizeWorkDays(workDays),
		Active:        active,
	}
}

func DefaultWorkDays() []int {
	return []int{1, 2, 3, 4, 5}
}

func NormalizeWorkDays(values []int) []int {
	if len(values) == 0 {
		return DefaultWorkDays()
	}

	seen := map[int]struct{}{}
	out := make([]int, 0, len(values))
	for _, value := range values {
		if value < 1 || value > 7 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	if len(out) == 0 {
		return DefaultWorkDays()
	}
	return out
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
