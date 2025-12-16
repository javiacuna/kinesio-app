package gorm

import "time"

type PatientEvolutionModel struct {
	ID              string  `gorm:"type:uuid;primaryKey"`
	PatientID       string  `gorm:"type:uuid;not null"`
	KinesiologistID string  `gorm:"type:uuid;not null"`
	AppointmentID   *string `gorm:"type:uuid"`
	PainLevel       *int
	Notes           string `gorm:"not null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (PatientEvolutionModel) TableName() string { return "patient_evolutions" }
