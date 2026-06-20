package gorm

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/javiacuna/kinesio-backend/internal/clinicalreports/domain"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, patientID uuid.UUID, kinesiologistID uuid.UUID, periodFrom time.Time, periodTo time.Time, recommendations *string, actorEmail *string, actorRole *string) (domain.ClinicalReport, error) {
	stats, err := r.stats(ctx, patientID, periodFrom, periodTo)
	if err != nil {
		return domain.ClinicalReport{}, err
	}

	report := domain.NewClinicalReport(patientID, kinesiologistID, periodFrom, periodTo, stats, recommendations, actorEmail, actorRole)
	model := toModel(report)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.ClinicalReport{}, err
	}

	return toDomain(model), nil
}

func (r *Repository) ListByPatient(ctx context.Context, patientID uuid.UUID, limit int) ([]domain.ClinicalReport, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var models []ClinicalReportModel
	if err := r.db.WithContext(ctx).
		Order("created_at desc").
		Limit(limit).
		Find(&models, "patient_id = ?", patientID.String()).
		Error; err != nil {
		return nil, err
	}

	out := make([]domain.ClinicalReport, 0, len(models))
	for _, model := range models {
		out = append(out, toDomain(model))
	}
	return out, nil
}

func (r *Repository) stats(ctx context.Context, patientID uuid.UUID, periodFrom time.Time, periodTo time.Time) (domain.ReportStats, error) {
	endExclusive := periodTo.AddDate(0, 0, 1)

	var evolutionStats struct {
		EvolutionCount     int
		AvgPainLevel       *float64
		AvgMobilityScore   *float64
		AvgStrengthScore   *float64
		AvgFunctionalScore *float64
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*)::int AS evolution_count,
			AVG(pain_level)::float8 AS avg_pain_level,
			AVG(mobility_score)::float8 AS avg_mobility_score,
			AVG(strength_score)::float8 AS avg_strength_score,
			AVG(functional_score)::float8 AS avg_functional_score
		FROM patient_evolutions
		WHERE patient_id = ? AND created_at >= ? AND created_at < ?
	`, patientID.String(), periodFrom, endExclusive).Scan(&evolutionStats).Error; err != nil {
		return domain.ReportStats{}, err
	}

	var planStats struct {
		ActivePlanCount     int
		ActivePlanItemCount int
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(DISTINCT ep.id)::int AS active_plan_count,
			COUNT(epi.id)::int AS active_plan_item_count
		FROM exercise_plans ep
		LEFT JOIN exercise_plan_items epi ON epi.plan_id = ep.id
		WHERE ep.patient_id = ? AND ep.status = 'active'
	`, patientID.String()).Scan(&planStats).Error; err != nil {
		return domain.ReportStats{}, err
	}

	return domain.ReportStats{
		EvolutionCount:      evolutionStats.EvolutionCount,
		AvgPainLevel:        evolutionStats.AvgPainLevel,
		AvgMobilityScore:    evolutionStats.AvgMobilityScore,
		AvgStrengthScore:    evolutionStats.AvgStrengthScore,
		AvgFunctionalScore:  evolutionStats.AvgFunctionalScore,
		ActivePlanCount:     planStats.ActivePlanCount,
		ActivePlanItemCount: planStats.ActivePlanItemCount,
	}, nil
}

func toModel(report domain.ClinicalReport) ClinicalReportModel {
	return ClinicalReportModel{
		ID:                  report.ID.String(),
		PatientID:           report.PatientID.String(),
		KinesiologistID:     report.KinesiologistID.String(),
		PeriodFrom:          report.PeriodFrom,
		PeriodTo:            report.PeriodTo,
		EvolutionCount:      report.EvolutionCount,
		AvgPainLevel:        report.AvgPainLevel,
		AvgMobilityScore:    report.AvgMobilityScore,
		AvgStrengthScore:    report.AvgStrengthScore,
		AvgFunctionalScore:  report.AvgFunctionalScore,
		ActivePlanCount:     report.ActivePlanCount,
		ActivePlanItemCount: report.ActivePlanItemCount,
		Summary:             report.Summary,
		Recommendations:     report.Recommendations,
		CreatedByEmail:      report.CreatedByEmail,
		CreatedByRole:       report.CreatedByRole,
		CreatedAt:           report.CreatedAt,
		UpdatedAt:           report.UpdatedAt,
	}
}

func toDomain(model ClinicalReportModel) domain.ClinicalReport {
	return domain.ClinicalReport{
		ID:                  uuid.MustParse(model.ID),
		PatientID:           uuid.MustParse(model.PatientID),
		KinesiologistID:     uuid.MustParse(model.KinesiologistID),
		PeriodFrom:          model.PeriodFrom,
		PeriodTo:            model.PeriodTo,
		EvolutionCount:      model.EvolutionCount,
		AvgPainLevel:        model.AvgPainLevel,
		AvgMobilityScore:    model.AvgMobilityScore,
		AvgStrengthScore:    model.AvgStrengthScore,
		AvgFunctionalScore:  model.AvgFunctionalScore,
		ActivePlanCount:     model.ActivePlanCount,
		ActivePlanItemCount: model.ActivePlanItemCount,
		Summary:             model.Summary,
		Recommendations:     model.Recommendations,
		CreatedByEmail:      model.CreatedByEmail,
		CreatedByRole:       model.CreatedByRole,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}
}
