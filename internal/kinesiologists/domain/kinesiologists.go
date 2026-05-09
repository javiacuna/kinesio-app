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
	PracticeIDs   []uuid.UUID
	Practices     []Practice
	Active        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Specialty struct {
	ID        uuid.UUID
	Name      string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Practice struct {
	ID          uuid.UUID
	SpecialtyID uuid.UUID
	Name        string
	Description *string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NormalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func NewKinesiologist(firstName, lastName, email string, licenseNumber *string, workStartTime, workEndTime string, workDays []int, practiceIDs []uuid.UUID, active bool) Kinesiologist {
	return Kinesiologist{
		ID:            uuid.New(),
		FirstName:     strings.TrimSpace(firstName),
		LastName:      strings.TrimSpace(lastName),
		Email:         NormalizeEmail(email),
		LicenseNumber: trimOptional(licenseNumber),
		WorkStartTime: strings.TrimSpace(workStartTime),
		WorkEndTime:   strings.TrimSpace(workEndTime),
		WorkDays:      NormalizeWorkDays(workDays),
		PracticeIDs:   NormalizePracticeIDs(practiceIDs),
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

func NormalizePracticeIDs(values []uuid.UUID) []uuid.UUID {
	if len(values) == 0 {
		return nil
	}
	seen := map[uuid.UUID]struct{}{}
	out := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
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
