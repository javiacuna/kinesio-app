package domain

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	SearchCIE10(ctx context.Context, query string, limit int) ([]CIE10Code, error)
	GetCIE10ByCode(ctx context.Context, code string) (CIE10Code, bool, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID) ([]PatientDiagnosis, error)
	CreatePatientDiagnosis(ctx context.Context, diagnosis PatientDiagnosis) (PatientDiagnosis, error)
	UpdatePatientDiagnosis(ctx context.Context, diagnosis PatientDiagnosis) (PatientDiagnosis, error)
	DeletePatientDiagnosis(ctx context.Context, id uuid.UUID) error
}
