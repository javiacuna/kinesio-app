package gorm

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/javiacuna/kinesio-backend/internal/patientcheckins/domain"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, item domain.PatientCheckIn) (domain.PatientCheckIn, error) {
	model := toModel(item)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.PatientCheckIn{}, err
	}
	return toDomain(model), nil
}

func (r *Repository) ListByPatient(ctx context.Context, patientID uuid.UUID, limit int) ([]domain.PatientCheckIn, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var models []PatientCheckInModel
	if err := r.db.WithContext(ctx).
		Order("created_at desc").
		Limit(limit).
		Find(&models, "patient_id = ?", patientID.String()).
		Error; err != nil {
		return nil, err
	}

	out := make([]domain.PatientCheckIn, 0, len(models))
	for _, model := range models {
		out = append(out, toDomain(model))
	}
	return out, nil
}

func toModel(item domain.PatientCheckIn) PatientCheckInModel {
	return PatientCheckInModel{
		ID:              item.ID.String(),
		PatientID:       item.PatientID.String(),
		PainLevel:       item.PainLevel,
		MobilityScore:   item.MobilityScore,
		StrengthScore:   item.StrengthScore,
		FunctionalScore: item.FunctionalScore,
		Notes:           item.Notes,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

func toDomain(model PatientCheckInModel) domain.PatientCheckIn {
	return domain.PatientCheckIn{
		ID:              uuid.MustParse(model.ID),
		PatientID:       uuid.MustParse(model.PatientID),
		PainLevel:       model.PainLevel,
		MobilityScore:   model.MobilityScore,
		StrengthScore:   model.StrengthScore,
		FunctionalScore: model.FunctionalScore,
		Notes:           model.Notes,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}
