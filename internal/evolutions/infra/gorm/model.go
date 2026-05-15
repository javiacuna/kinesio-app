package gorm

import "time"

type PatientEvolutionModel struct {
	ID              string  `gorm:"type:uuid;primaryKey"`
	PatientID       string  `gorm:"type:uuid;not null"`
	KinesiologistID string  `gorm:"type:uuid;not null"`
	AppointmentID   *string `gorm:"type:uuid"`
	DiagnosisID     *string `gorm:"type:uuid;column:patient_diagnosis_id"`
	PainLevel       *int
	Notes           string `gorm:"not null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time

	Photos []PatientEvolutionPhotoModel `gorm:"foreignKey:EvolutionID;constraint:OnDelete:CASCADE"`
}

func (PatientEvolutionModel) TableName() string { return "patient_evolutions" }

type PatientEvolutionPhotoModel struct {
	ID          string `gorm:"type:uuid;primaryKey"`
	EvolutionID string `gorm:"type:uuid;not null"`
	URL         string `gorm:"not null"`
	Caption     *string
	CreatedAt   time.Time
}

func (PatientEvolutionPhotoModel) TableName() string { return "patient_evolution_photos" }
