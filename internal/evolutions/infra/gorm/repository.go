package gorm

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/javiacuna/kinesio-backend/internal/evolutions/domain"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, e domain.PatientEvolution) (domain.PatientEvolution, error) {
	m := toModel(e)

	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return domain.PatientEvolution{}, err
	}

	out, _, err := r.GetByID(ctx, uuid.MustParse(m.ID))
	if err != nil {
		return domain.PatientEvolution{}, err
	}
	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (domain.PatientEvolution, bool, error) {
	var m PatientEvolutionModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id.String()).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.PatientEvolution{}, false, nil
		}
		return domain.PatientEvolution{}, false, err
	}
	return toDomain(m), true, nil
}

func (r *Repository) ListByPatient(ctx context.Context, patientID uuid.UUID, limit int) ([]domain.PatientEvolution, error) {
	var ms []PatientEvolutionModel
	if err := r.db.WithContext(ctx).
		Order("created_at desc").
		Limit(limit).
		Find(&ms, "patient_id = ?", patientID.String()).
		Error; err != nil {
		return nil, err
	}

	out := make([]domain.PatientEvolution, 0, len(ms))
	for _, m := range ms {
		out = append(out, toDomain(m))
	}
	return out, nil
}

func toModel(e domain.PatientEvolution) PatientEvolutionModel {
	var appt *string
	if e.AppointmentID != nil {
		s := e.AppointmentID.String()
		appt = &s
	}

	return PatientEvolutionModel{
		ID:              e.ID.String(),
		PatientID:       e.PatientID.String(),
		KinesiologistID: e.KinesiologistID.String(),
		AppointmentID:   appt,
		PainLevel:       e.PainLevel,
		Notes:           e.Notes,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}

func toDomain(m PatientEvolutionModel) domain.PatientEvolution {
	var appt *uuid.UUID
	if m.AppointmentID != nil && *m.AppointmentID != "" {
		id := uuid.MustParse(*m.AppointmentID)
		appt = &id
	}

	return domain.PatientEvolution{
		ID:              uuid.MustParse(m.ID),
		PatientID:       uuid.MustParse(m.PatientID),
		KinesiologistID: uuid.MustParse(m.KinesiologistID),
		AppointmentID:   appt,
		PainLevel:       m.PainLevel,
		Notes:           m.Notes,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}
