package gorm

import (
	"time"

	"github.com/google/uuid"
)

type PatientModel struct {
	ID                          uuid.UUID  `gorm:"type:uuid;primaryKey;column:id"`
	DNI                         string     `gorm:"column:dni;not null;uniqueIndex:ux_patients_dni"`
	FirstName                   string     `gorm:"column:first_name;not null"`
	LastName                    string     `gorm:"column:last_name;not null"`
	Email                       string     `gorm:"column:email;not null;uniqueIndex:ux_patients_email,expression:lower(email)"`
	Phone                       *string    `gorm:"column:phone"`
	BirthDate                   *time.Time `gorm:"column:birth_date"`
	ClinicalNotes               *string    `gorm:"column:clinical_notes"`
	ClinicalNotesUpdatedByEmail *string    `gorm:"column:clinical_notes_updated_by_email"`
	ClinicalNotesUpdatedByRole  *string    `gorm:"column:clinical_notes_updated_by_role"`
	ClinicalNotesUpdatedAt      *time.Time `gorm:"column:clinical_notes_updated_at"`
	FinancierID                 *uuid.UUID `gorm:"column:financier_id"`
	FinancierMemberNumber       *string    `gorm:"column:financier_member_number"`
	Active                      bool       `gorm:"column:active;not null"`
	CreatedAt                   time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt                   time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (PatientModel) TableName() string { return "patients" }
