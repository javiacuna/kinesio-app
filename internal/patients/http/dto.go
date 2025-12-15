package http

import (
	"time"

	"github.com/javiacuna/kinesio-backend/internal/patients/domain"
)

type PatientResponse struct {
	ID            string     `json:"id"`
	DNI           string     `json:"dni"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	Email         string     `json:"email"`
	Phone         *string    `json:"phone,omitempty"`
	BirthDate     *time.Time `json:"birth_date,omitempty"`
	ClinicalNotes *string    `json:"clinical_notes,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func PatientResponseFromDomain(p domain.Patient) PatientResponse {
	return PatientResponse{
		ID:            p.ID.String(),
		DNI:           p.DNI,
		FirstName:     p.FirstName,
		LastName:      p.LastName,
		Email:         p.Email,
		Phone:         p.Phone,
		BirthDate:     p.BirthDate,
		ClinicalNotes: p.ClinicalNotes,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}
