package gorm

import (
	"time"

	"github.com/google/uuid"
)

type CIE10CodeModel struct {
	Code        string  `gorm:"column:code;primaryKey"`
	Description string  `gorm:"column:description;not null"`
	Chapter     *string `gorm:"column:chapter"`
	Active      bool    `gorm:"column:active;not null"`
}

func (CIE10CodeModel) TableName() string { return "cie10_codes" }

type PatientDiagnosisModel struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;column:id"`
	PatientID      uuid.UUID      `gorm:"type:uuid;column:patient_id;not null"`
	CIE10Code      string         `gorm:"column:cie10_code;not null"`
	CIE10          CIE10CodeModel `gorm:"foreignKey:CIE10Code;references:Code"`
	Kind           string         `gorm:"column:kind;not null"`
	Status         string         `gorm:"column:status;not null"`
	DiagnosedAt    time.Time      `gorm:"column:diagnosed_at;not null"`
	Notes          *string        `gorm:"column:notes"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	CreatedByEmail *string        `gorm:"column:created_by_email"`
	CreatedByRole  *string        `gorm:"column:created_by_role"`
	UpdatedByEmail *string        `gorm:"column:updated_by_email"`
	UpdatedByRole  *string        `gorm:"column:updated_by_role"`
}

func (PatientDiagnosisModel) TableName() string { return "patient_diagnoses" }
