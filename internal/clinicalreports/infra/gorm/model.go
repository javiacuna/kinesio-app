package gorm

import "time"

type ClinicalReportModel struct {
	ID                  string `gorm:"type:uuid;primaryKey"`
	PatientID           string `gorm:"type:uuid;not null"`
	KinesiologistID     string `gorm:"type:uuid;not null"`
	PeriodFrom          time.Time
	PeriodTo            time.Time
	EvolutionCount      int
	AvgPainLevel        *float64
	AvgMobilityScore    *float64
	AvgStrengthScore    *float64
	AvgFunctionalScore  *float64
	ActivePlanCount     int
	ActivePlanItemCount int
	Summary             string `gorm:"not null"`
	Recommendations     *string
	CreatedByEmail      *string
	CreatedByRole       *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (ClinicalReportModel) TableName() string { return "clinical_reports" }
