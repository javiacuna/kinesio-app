package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	DiagnosisKindPrimary   = "primary"
	DiagnosisKindSecondary = "secondary"

	DiagnosisStatusSuspected = "suspected"
	DiagnosisStatusConfirmed = "confirmed"
	DiagnosisStatusResolved  = "resolved"
)

type CIE10Code struct {
	Code        string
	Description string
	Chapter     *string
	Active      bool
}

type PatientDiagnosis struct {
	ID             uuid.UUID
	PatientID      uuid.UUID
	CIE10Code      string
	CIE10          CIE10Code
	Kind           string
	Status         string
	DiagnosedAt    time.Time
	Notes          *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CreatedByEmail *string
	CreatedByRole  *string
	UpdatedByEmail *string
	UpdatedByRole  *string
}

func NormalizeCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func NormalizeKind(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == DiagnosisKindSecondary {
		return DiagnosisKindSecondary
	}
	return DiagnosisKindPrimary
}

func NormalizeStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case DiagnosisStatusConfirmed:
		return DiagnosisStatusConfirmed
	case DiagnosisStatusResolved:
		return DiagnosisStatusResolved
	default:
		return DiagnosisStatusSuspected
	}
}

func TrimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
