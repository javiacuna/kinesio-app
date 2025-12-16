package gorm

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/javiacuna/kinesio-backend/internal/exerciseplans/domain"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, p domain.ExercisePlan) (domain.ExercisePlan, error) {
	m := toModel(p)

	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return domain.ExercisePlan{}, err
	}

	out, _, err := r.GetByID(ctx, uuid.MustParse(m.ID))
	if err != nil {
		return domain.ExercisePlan{}, err
	}

	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (domain.ExercisePlan, bool, error) {
	var m ExercisePlanModel
	err := r.db.WithContext(ctx).
		Preload("Items").
		First(&m, "id = ?", id.String()).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.ExercisePlan{}, false, nil
		}
		return domain.ExercisePlan{}, false, err
	}
	return toDomain(m), true, nil
}

func (r *Repository) ListByPatient(ctx context.Context, patientID uuid.UUID) ([]domain.ExercisePlan, error) {
	var ms []ExercisePlanModel
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Order("created_at desc").
		Find(&ms, "patient_id = ?", patientID.String()).
		Error; err != nil {
		return nil, err
	}
	out := make([]domain.ExercisePlan, 0, len(ms))
	for _, m := range ms {
		out = append(out, toDomain(m))
	}
	return out, nil
}

func toModel(p domain.ExercisePlan) ExercisePlanModel {
	m := ExercisePlanModel{
		ID:              p.ID.String(),
		PatientID:       p.PatientID.String(),
		KinesiologistID: p.KinesiologistID.String(),
		Frequency:       string(p.Frequency),
		DurationWeeks:   p.DurationWeeks,
		Observations:    p.Observations,
		Status:          string(p.Status),
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		Items:           make([]ExercisePlanItemModel, 0, len(p.Items)),
	}
	for _, it := range p.Items {
		m.Items = append(m.Items, ExercisePlanItemModel{
			ID:               it.ID.String(),
			PlanID:           it.PlanID.String(),
			Name:             it.Name,
			Description:      it.Description,
			VideoURL:         it.VideoURL,
			GuideURL:         it.GuideURL,
			EstimatedMinutes: it.EstimatedMinutes,
			Sets:             it.Sets,
			Reps:             it.Reps,
			CreatedAt:        it.CreatedAt,
			UpdatedAt:        it.UpdatedAt,
		})
	}
	return m
}

func toDomain(m ExercisePlanModel) domain.ExercisePlan {
	p := domain.ExercisePlan{
		ID:              uuid.MustParse(m.ID),
		PatientID:       uuid.MustParse(m.PatientID),
		KinesiologistID: uuid.MustParse(m.KinesiologistID),
		Frequency:       domain.Frequency(m.Frequency),
		DurationWeeks:   m.DurationWeeks,
		Observations:    m.Observations,
		Status:          domain.PlanStatus(m.Status),
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		Items:           make([]domain.ExercisePlanItem, 0, len(m.Items)),
	}
	for _, it := range m.Items {
		p.Items = append(p.Items, domain.ExercisePlanItem{
			ID:               uuid.MustParse(it.ID),
			PlanID:           uuid.MustParse(it.PlanID),
			Name:             it.Name,
			Description:      it.Description,
			VideoURL:         it.VideoURL,
			GuideURL:         it.GuideURL,
			EstimatedMinutes: it.EstimatedMinutes,
			Sets:             it.Sets,
			Reps:             it.Reps,
			CreatedAt:        it.CreatedAt,
			UpdatedAt:        it.UpdatedAt,
		})
	}
	return p
}
