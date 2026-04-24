package gorm

import (
	"time"

	"github.com/google/uuid"
)

type StaffMemberModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;column:id"`
	FirstName   string    `gorm:"column:first_name;not null"`
	LastName    string    `gorm:"column:last_name;not null"`
	Email       string    `gorm:"column:email;not null;uniqueIndex:ux_staff_members_email,expression:lower(email)"`
	Role        string    `gorm:"column:role;not null"`
	Phone       *string   `gorm:"column:phone"`
	Active      bool      `gorm:"column:active;not null"`
	FirebaseUID *string   `gorm:"column:firebase_uid"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (StaffMemberModel) TableName() string { return "staff_members" }
