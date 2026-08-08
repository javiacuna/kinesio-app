package gorm

import (
	"time"

	"github.com/google/uuid"
)

type SignupRequestModel struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;column:id"`
	FirebaseUID      string     `gorm:"column:firebase_uid;not null;uniqueIndex:ux_patient_signup_requests_firebase_uid"`
	DNI              string     `gorm:"column:dni;not null"`
	Email            string     `gorm:"column:email;not null"`
	FirstName        string     `gorm:"column:first_name;not null"`
	LastName         string     `gorm:"column:last_name;not null"`
	Status           string     `gorm:"column:status;not null"`
	MatchedPatientID *uuid.UUID `gorm:"column:matched_patient_id"`
	ReviewedByEmail  *string    `gorm:"column:reviewed_by_email"`
	ReviewedAt       *time.Time `gorm:"column:reviewed_at"`
	RejectionReason  *string    `gorm:"column:rejection_reason"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (SignupRequestModel) TableName() string { return "patient_signup_requests" }
