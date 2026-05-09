package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/diagnoses/domain"
)

type SearchCIE10UseCase struct {
	repo domain.Repository
}

func NewSearchCIE10UseCase(repo domain.Repository) *SearchCIE10UseCase {
	return &SearchCIE10UseCase{repo: repo}
}

func (uc *SearchCIE10UseCase) Execute(ctx context.Context, query string, limit int) ([]domain.CIE10Code, error) {
	return uc.repo.SearchCIE10(ctx, query, limit)
}

type ListPatientDiagnosesUseCase struct {
	repo domain.Repository
}

func NewListPatientDiagnosesUseCase(repo domain.Repository) *ListPatientDiagnosesUseCase {
	return &ListPatientDiagnosesUseCase{repo: repo}
}

func (uc *ListPatientDiagnosesUseCase) Execute(ctx context.Context, patientID string) ([]domain.PatientDiagnosis, map[string]string, error) {
	id, errs := parseUUID("patient_id", patientID)
	if len(errs) > 0 {
		return nil, errs, domain.ErrValidation
	}
	out, err := uc.repo.ListByPatient(ctx, id)
	return out, nil, err
}

type SavePatientDiagnosisInput struct {
	ID          string
	PatientID   string
	CIE10Code   string
	Kind        string
	Status      string
	DiagnosedAt string
	Notes       *string
	ActorEmail  *string
	ActorRole   *string
}

type SavePatientDiagnosisUseCase struct {
	repo domain.Repository
}

func NewSavePatientDiagnosisUseCase(repo domain.Repository) *SavePatientDiagnosisUseCase {
	return &SavePatientDiagnosisUseCase{repo: repo}
}

func (uc *SavePatientDiagnosisUseCase) Create(ctx context.Context, in SavePatientDiagnosisInput) (domain.PatientDiagnosis, map[string]string, error) {
	diagnosis, errs := buildDiagnosis(in)
	if len(errs) > 0 {
		return domain.PatientDiagnosis{}, errs, domain.ErrValidation
	}
	diagnosis.ID = uuid.New()
	diagnosis.CreatedByEmail = in.ActorEmail
	diagnosis.CreatedByRole = in.ActorRole
	diagnosis.UpdatedByEmail = in.ActorEmail
	diagnosis.UpdatedByRole = in.ActorRole

	if _, found, err := uc.repo.GetCIE10ByCode(ctx, diagnosis.CIE10Code); err != nil {
		return domain.PatientDiagnosis{}, nil, err
	} else if !found {
		return domain.PatientDiagnosis{}, map[string]string{"cie10_code": "Código CIE-10 no encontrado"}, domain.ErrValidation
	}

	out, err := uc.repo.CreatePatientDiagnosis(ctx, diagnosis)
	return out, nil, err
}

func (uc *SavePatientDiagnosisUseCase) Update(ctx context.Context, in SavePatientDiagnosisInput) (domain.PatientDiagnosis, map[string]string, error) {
	diagnosis, errs := buildDiagnosis(in)
	id, idErrs := parseUUID("id", in.ID)
	for key, value := range idErrs {
		errs[key] = value
	}
	if len(errs) > 0 {
		return domain.PatientDiagnosis{}, errs, domain.ErrValidation
	}
	diagnosis.ID = id
	diagnosis.UpdatedByEmail = in.ActorEmail
	diagnosis.UpdatedByRole = in.ActorRole

	if _, found, err := uc.repo.GetCIE10ByCode(ctx, diagnosis.CIE10Code); err != nil {
		return domain.PatientDiagnosis{}, nil, err
	} else if !found {
		return domain.PatientDiagnosis{}, map[string]string{"cie10_code": "Código CIE-10 no encontrado"}, domain.ErrValidation
	}

	out, err := uc.repo.UpdatePatientDiagnosis(ctx, diagnosis)
	return out, nil, err
}

type DeletePatientDiagnosisUseCase struct {
	repo domain.Repository
}

func NewDeletePatientDiagnosisUseCase(repo domain.Repository) *DeletePatientDiagnosisUseCase {
	return &DeletePatientDiagnosisUseCase{repo: repo}
}

func (uc *DeletePatientDiagnosisUseCase) Execute(ctx context.Context, id string) (map[string]string, error) {
	parsed, errs := parseUUID("id", id)
	if len(errs) > 0 {
		return errs, domain.ErrValidation
	}
	return nil, uc.repo.DeletePatientDiagnosis(ctx, parsed)
}

func buildDiagnosis(in SavePatientDiagnosisInput) (domain.PatientDiagnosis, map[string]string) {
	errs := map[string]string{}
	patientID, patientErrs := parseUUID("patient_id", in.PatientID)
	for key, value := range patientErrs {
		errs[key] = value
	}

	code := domain.NormalizeCode(in.CIE10Code)
	if code == "" {
		errs["cie10_code"] = "Campo obligatorio"
	}

	diagnosedAt := time.Now().UTC()
	if value := strings.TrimSpace(in.DiagnosedAt); value != "" {
		parsed, err := time.Parse("2006-01-02", value)
		if err != nil {
			errs["diagnosed_at"] = "Fecha inválida"
		} else {
			diagnosedAt = parsed
		}
	}

	return domain.PatientDiagnosis{
		PatientID:   patientID,
		CIE10Code:   code,
		Kind:        domain.NormalizeKind(in.Kind),
		Status:      domain.NormalizeStatus(in.Status),
		DiagnosedAt: diagnosedAt,
		Notes:       domain.TrimOptional(in.Notes),
	}, errs
}

func parseUUID(field string, value string) (uuid.UUID, map[string]string) {
	errs := map[string]string{}
	parsed, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		errs[field] = "UUID inválido"
	}
	return parsed, errs
}
