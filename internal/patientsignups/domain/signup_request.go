package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

type SignupRequest struct {
	ID               uuid.UUID
	FirebaseUID      string
	DNI              string
	Email            string
	FirstName        string
	LastName         string
	Status           Status
	MatchedPatientID *uuid.UUID
	ReviewedByEmail  *string
	ReviewedAt       *time.Time
	RejectionReason  *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewSignupRequest(firebaseUID, dni, firstName, lastName, email string) SignupRequest {
	return SignupRequest{
		ID:          uuid.New(),
		FirebaseUID: strings.TrimSpace(firebaseUID),
		DNI:         strings.TrimSpace(dni),
		FirstName:   strings.TrimSpace(firstName),
		LastName:    strings.TrimSpace(lastName),
		Email:       strings.ToLower(strings.TrimSpace(email)),
		Status:      StatusPending,
	}
}
