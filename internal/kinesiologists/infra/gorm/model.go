package gorm

import (
	"time"

	"github.com/google/uuid"
)

type KinesiologistModel struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;column:id"`
	FirstName     string    `gorm:"column:first_name;not null"`
	LastName      string    `gorm:"column:last_name;not null"`
	Email         string    `gorm:"column:email;not null;uniqueIndex:ux_kinesiologists_email,expression:lower(email)"`
	LicenseNumber *string   `gorm:"column:license_number"`
	Active        bool      `gorm:"column:active;not null"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (KinesiologistModel) TableName() string { return "kinesiologists" }
