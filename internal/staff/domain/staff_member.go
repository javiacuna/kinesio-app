package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin         Role = "admin"
	RoleReceptionist  Role = "recepcionista"
	RoleKinesiologist Role = "kinesiologo"
)

type StaffMember struct {
	ID          uuid.UUID
	FirstName   string
	LastName    string
	Email       string
	Role        Role
	Phone       *string
	Active      bool
	FirebaseUID *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewStaffMember(firstName, lastName, email string, role Role, phone *string, active bool, firebaseUID *string) StaffMember {
	return StaffMember{
		ID:          uuid.New(),
		FirstName:   strings.TrimSpace(firstName),
		LastName:    strings.TrimSpace(lastName),
		Email:       NormalizeEmail(email),
		Role:        role,
		Phone:       trimOptional(phone),
		Active:      active,
		FirebaseUID: trimOptional(firebaseUID),
	}
}

func NormalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func IsValidRole(role Role) bool {
	switch role {
	case RoleAdmin, RoleReceptionist, RoleKinesiologist:
		return true
	default:
		return false
	}
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
