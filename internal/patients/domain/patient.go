package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Patient struct {
	ID            uuid.UUID
	DNI           string
	FirstName     string
	LastName      string
	Email         string
	Phone         *string
	BirthDate     *time.Time
	ClinicalNotes *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewPatient(dni, firstName, lastName, email string, phone *string, birthDate *time.Time, notes *string) Patient {
	return Patient{
		ID:            uuid.New(),
		DNI:           strings.TrimSpace(dni),
		FirstName:     strings.TrimSpace(firstName),
		LastName:      strings.TrimSpace(lastName),
		Email:         strings.ToLower(strings.TrimSpace(email)),
		Phone:         trimPtr(phone),
		BirthDate:     birthDate,
		ClinicalNotes: trimPtr(notes),
	}
}

func trimPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}
