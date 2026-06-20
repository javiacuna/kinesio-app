package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ClinicalReport struct {
	ID                  uuid.UUID
	PatientID           uuid.UUID
	KinesiologistID     uuid.UUID
	PeriodFrom          time.Time
	PeriodTo            time.Time
	EvolutionCount      int
	AvgPainLevel        *float64
	AvgMobilityScore    *float64
	AvgStrengthScore    *float64
	AvgFunctionalScore  *float64
	ActivePlanCount     int
	ActivePlanItemCount int
	Summary             string
	Recommendations     *string
	CreatedByEmail      *string
	CreatedByRole       *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type ReportStats struct {
	EvolutionCount      int
	AvgPainLevel        *float64
	AvgMobilityScore    *float64
	AvgStrengthScore    *float64
	AvgFunctionalScore  *float64
	ActivePlanCount     int
	ActivePlanItemCount int
}

func NewClinicalReport(patientID, kinesiologistID uuid.UUID, periodFrom, periodTo time.Time, stats ReportStats, recommendations *string, actorEmail *string, actorRole *string) ClinicalReport {
	now := time.Now().UTC()
	return ClinicalReport{
		ID:                  uuid.New(),
		PatientID:           patientID,
		KinesiologistID:     kinesiologistID,
		PeriodFrom:          truncateDate(periodFrom),
		PeriodTo:            truncateDate(periodTo),
		EvolutionCount:      stats.EvolutionCount,
		AvgPainLevel:        roundOptional(stats.AvgPainLevel),
		AvgMobilityScore:    roundOptional(stats.AvgMobilityScore),
		AvgStrengthScore:    roundOptional(stats.AvgStrengthScore),
		AvgFunctionalScore:  roundOptional(stats.AvgFunctionalScore),
		ActivePlanCount:     stats.ActivePlanCount,
		ActivePlanItemCount: stats.ActivePlanItemCount,
		Summary:             BuildSummary(stats),
		Recommendations:     trimOptional(recommendations),
		CreatedByEmail:      trimOptional(actorEmail),
		CreatedByRole:       trimOptional(actorRole),
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

func BuildSummary(stats ReportStats) string {
	var parts []string
	if stats.EvolutionCount == 0 {
		parts = append(parts, "No se registraron evoluciones clinicas en el periodo seleccionado.")
	} else if stats.EvolutionCount == 1 {
		parts = append(parts, "Se registro 1 evolucion clinica en el periodo seleccionado.")
	} else {
		parts = append(parts, fmt.Sprintf("Se registraron %d evoluciones clinicas en el periodo seleccionado.", stats.EvolutionCount))
	}

	if stats.AvgPainLevel != nil {
		parts = append(parts, fmt.Sprintf("Dolor promedio: %.1f/10.", *stats.AvgPainLevel))
	}
	if stats.AvgMobilityScore != nil {
		parts = append(parts, fmt.Sprintf("Movilidad promedio: %.1f/100.", *stats.AvgMobilityScore))
	}
	if stats.AvgStrengthScore != nil {
		parts = append(parts, fmt.Sprintf("Fuerza promedio: %.1f/100.", *stats.AvgStrengthScore))
	}
	if stats.AvgFunctionalScore != nil {
		parts = append(parts, fmt.Sprintf("Funcionalidad promedio: %.1f/100.", *stats.AvgFunctionalScore))
	}

	if stats.ActivePlanCount == 0 {
		parts = append(parts, "El paciente no registra planes activos al momento del reporte.")
	} else if stats.ActivePlanCount == 1 {
		parts = append(parts, fmt.Sprintf("El paciente registra 1 plan activo con %d ejercicios.", stats.ActivePlanItemCount))
	} else {
		parts = append(parts, fmt.Sprintf("El paciente registra %d planes activos con %d ejercicios en total.", stats.ActivePlanCount, stats.ActivePlanItemCount))
	}

	return strings.Join(parts, " ")
}

func truncateDate(value time.Time) time.Time {
	y, m, d := value.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func roundOptional(value *float64) *float64 {
	if value == nil {
		return nil
	}
	rounded := float64(int((*value)*100+0.5)) / 100
	return &rounded
}
