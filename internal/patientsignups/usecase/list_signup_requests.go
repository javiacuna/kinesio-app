package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"
	patientsDomain "github.com/javiacuna/kinesio-backend/internal/patients/domain"
	"github.com/javiacuna/kinesio-backend/internal/patientsignups/domain"
	"github.com/javiacuna/kinesio-backend/internal/patientsignups/ports"
)

// patientDNIMatcher busca un paciente existente por DNI, sin importar el
// email. Se usa sólo para sugerir una posible coincidencia al revisor; el
// match "fuerte" (DNI+email) que dispara la auto-aprobación sigue viviendo
// en CreateSignupRequestUseCase.
type patientDNIMatcher interface {
	FindByDNI(ctx context.Context, dni string) (patientsDomain.Patient, bool, error)
}

// PossibleMatch describe un paciente ya cargado por la clínica cuyo DNI
// coincide con el de una solicitud de autorregistro todavía sin resolver.
type PossibleMatch struct {
	PatientID    uuid.UUID
	FirstName    string
	LastName     string
	Email        string
	EmailMatches bool
}

// SignupRequestListItem es una solicitud de autorregistro más, opcionalmente,
// la sugerencia de con qué paciente existente podría corresponderse.
type SignupRequestListItem struct {
	domain.SignupRequest
	PossibleMatch *PossibleMatch
}

type ListSignupRequestsUseCase struct {
	repo     ports.Repository
	patients patientDNIMatcher
}

func NewListSignupRequestsUseCase(repo ports.Repository, patients patientDNIMatcher) *ListSignupRequestsUseCase {
	return &ListSignupRequestsUseCase{repo: repo, patients: patients}
}

func (uc *ListSignupRequestsUseCase) Execute(ctx context.Context, status string) ([]SignupRequestListItem, error) {
	items, err := uc.repo.List(ctx, status)
	if err != nil {
		return nil, err
	}

	out := make([]SignupRequestListItem, 0, len(items))
	for _, item := range items {
		entry := SignupRequestListItem{SignupRequest: item}

		// Si ya tiene un match confirmado (aprobada, o rechazada tras haber
		// matcheado), no hace falta sugerir nada: ya se sabe con certeza.
		if item.MatchedPatientID == nil && uc.patients != nil && strings.TrimSpace(item.DNI) != "" {
			patient, found, err := uc.patients.FindByDNI(ctx, item.DNI)
			if err == nil && found {
				entry.PossibleMatch = &PossibleMatch{
					PatientID:    patient.ID,
					FirstName:    patient.FirstName,
					LastName:     patient.LastName,
					Email:        patient.Email,
					EmailMatches: strings.EqualFold(strings.TrimSpace(patient.Email), strings.TrimSpace(item.Email)),
				}
			}
		}

		out = append(out, entry)
	}
	return out, nil
}
