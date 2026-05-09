package gorm

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/javiacuna/kinesio-backend/internal/diagnoses/domain"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

var _ domain.Repository = (*Repository)(nil)

func (r *Repository) SearchCIE10(ctx context.Context, query string, limit int) ([]domain.CIE10Code, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}

	q := r.db.WithContext(ctx).Model(&CIE10CodeModel{}).Where("active = true")
	query = strings.TrimSpace(query)
	if query != "" {
		pattern := "%" + query + "%"
		q = q.Where("code ILIKE ? OR description ILIKE ? OR chapter ILIKE ?", pattern, pattern, pattern)
	}

	var models []CIE10CodeModel
	if err := q.Order("code ASC").Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}

	out := make([]domain.CIE10Code, 0, len(models))
	for _, model := range models {
		out = append(out, toCIE10Domain(model))
	}
	return out, nil
}

func (r *Repository) GetCIE10ByCode(ctx context.Context, code string) (domain.CIE10Code, bool, error) {
	var model CIE10CodeModel
	err := r.db.WithContext(ctx).First(&model, "code = ? AND active = true", domain.NormalizeCode(code)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.CIE10Code{}, false, nil
		}
		return domain.CIE10Code{}, false, err
	}
	return toCIE10Domain(model), true, nil
}

func (r *Repository) ListByPatient(ctx context.Context, patientID uuid.UUID) ([]domain.PatientDiagnosis, error) {
	var models []PatientDiagnosisModel
	if err := r.db.WithContext(ctx).
		Preload("CIE10").
		Where("patient_id = ?", patientID).
		Order("diagnosed_at DESC, created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	out := make([]domain.PatientDiagnosis, 0, len(models))
	for _, model := range models {
		out = append(out, toDiagnosisDomain(model))
	}
	return out, nil
}

func (r *Repository) CreatePatientDiagnosis(ctx context.Context, diagnosis domain.PatientDiagnosis) (domain.PatientDiagnosis, error) {
	model := toDiagnosisModel(diagnosis)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		if isForeignKeyError(err) {
			return domain.PatientDiagnosis{}, domain.ErrNotFound
		}
		return domain.PatientDiagnosis{}, err
	}

	return r.getDiagnosisByID(ctx, model.ID)
}

func (r *Repository) UpdatePatientDiagnosis(ctx context.Context, diagnosis domain.PatientDiagnosis) (domain.PatientDiagnosis, error) {
	res := r.db.WithContext(ctx).
		Model(&PatientDiagnosisModel{}).
		Where("id = ?", diagnosis.ID).
		Updates(map[string]any{
			"cie10_code":       diagnosis.CIE10Code,
			"kind":             diagnosis.Kind,
			"status":           diagnosis.Status,
			"diagnosed_at":     diagnosis.DiagnosedAt,
			"notes":            diagnosis.Notes,
			"updated_by_email": diagnosis.UpdatedByEmail,
			"updated_by_role":  diagnosis.UpdatedByRole,
		})
	if res.Error != nil {
		if isForeignKeyError(res.Error) {
			return domain.PatientDiagnosis{}, domain.ErrNotFound
		}
		return domain.PatientDiagnosis{}, res.Error
	}
	if res.RowsAffected == 0 {
		return domain.PatientDiagnosis{}, domain.ErrNotFound
	}

	return r.getDiagnosisByID(ctx, diagnosis.ID)
}

func (r *Repository) DeletePatientDiagnosis(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&PatientDiagnosisModel{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) getDiagnosisByID(ctx context.Context, id uuid.UUID) (domain.PatientDiagnosis, error) {
	var model PatientDiagnosisModel
	err := r.db.WithContext(ctx).Preload("CIE10").First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.PatientDiagnosis{}, domain.ErrNotFound
		}
		return domain.PatientDiagnosis{}, err
	}
	return toDiagnosisDomain(model), nil
}

func toCIE10Domain(model CIE10CodeModel) domain.CIE10Code {
	return domain.CIE10Code{
		Code:        model.Code,
		Description: model.Description,
		Chapter:     model.Chapter,
		Active:      model.Active,
	}
}

func toDiagnosisDomain(model PatientDiagnosisModel) domain.PatientDiagnosis {
	return domain.PatientDiagnosis{
		ID:             model.ID,
		PatientID:      model.PatientID,
		CIE10Code:      model.CIE10Code,
		CIE10:          toCIE10Domain(model.CIE10),
		Kind:           model.Kind,
		Status:         model.Status,
		DiagnosedAt:    model.DiagnosedAt,
		Notes:          model.Notes,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
		CreatedByEmail: model.CreatedByEmail,
		CreatedByRole:  model.CreatedByRole,
		UpdatedByEmail: model.UpdatedByEmail,
		UpdatedByRole:  model.UpdatedByRole,
	}
}

func toDiagnosisModel(diagnosis domain.PatientDiagnosis) PatientDiagnosisModel {
	return PatientDiagnosisModel{
		ID:             diagnosis.ID,
		PatientID:      diagnosis.PatientID,
		CIE10Code:      diagnosis.CIE10Code,
		Kind:           diagnosis.Kind,
		Status:         diagnosis.Status,
		DiagnosedAt:    diagnosis.DiagnosedAt,
		Notes:          diagnosis.Notes,
		CreatedByEmail: diagnosis.CreatedByEmail,
		CreatedByRole:  diagnosis.CreatedByRole,
		UpdatedByEmail: diagnosis.UpdatedByEmail,
		UpdatedByRole:  diagnosis.UpdatedByRole,
	}
}

func isForeignKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
