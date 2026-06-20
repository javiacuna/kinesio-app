package gorm

import "time"

type PatientCheckInModel struct {
	ID              string `gorm:"type:uuid;primaryKey"`
	PatientID       string `gorm:"type:uuid;not null"`
	PainLevel       *int
	MobilityScore   *int
	StrengthScore   *int
	FunctionalScore *int
	Notes           string `gorm:"not null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (PatientCheckInModel) TableName() string { return "patient_check_ins" }
