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
	WorkStartTime string    `gorm:"column:work_start_time;not null"`
	WorkEndTime   string    `gorm:"column:work_end_time;not null"`
	WorkDays      string    `gorm:"column:work_days;not null"`
	Active        bool      `gorm:"column:active;not null"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (KinesiologistModel) TableName() string { return "kinesiologists" }

type SpecialtyModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;column:id"`
	Name      string    `gorm:"column:name;not null"`
	Active    bool      `gorm:"column:active;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (SpecialtyModel) TableName() string { return "specialties" }

type PracticeModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;column:id"`
	SpecialtyID uuid.UUID `gorm:"type:uuid;column:specialty_id;not null"`
	Name        string    `gorm:"column:name;not null"`
	Description *string   `gorm:"column:description"`
	Active      bool      `gorm:"column:active;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (PracticeModel) TableName() string { return "practices" }

type KinesiologistPracticeModel struct {
	KinesiologistID uuid.UUID `gorm:"type:uuid;primaryKey;column:kinesiologist_id"`
	PracticeID      uuid.UUID `gorm:"type:uuid;primaryKey;column:practice_id"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (KinesiologistPracticeModel) TableName() string { return "kinesiologist_practices" }
