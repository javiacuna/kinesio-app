package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/evolutions/domain"
)

type CreateEvolutionInput struct {
	PatientID       string  `json:"patient_id"`
	KinesiologistID string  `json:"kinesiologist_id"`
	AppointmentID   *string `json:"appointment_id,omitempty"`
	PainLevel       *int    `json:"pain_level,omitempty"`
	Notes           string  `json:"notes"`
}

type CreateEvolutionUseCase struct {
	repo domain.Repository
}

func NewCreateEvolutionUseCase(repo domain.Repository) *CreateEvolutionUseCase {
	return &CreateEvolutionUseCase{repo: repo}
}

func (uc *CreateEvolutionUseCase) Execute(ctx context.Context, in CreateEvolutionInput) (domain.PatientEvolution, map[string]string, error) {
	validation := map[string]string{}

	patientID, err := uuid.Parse(strings.TrimSpace(in.PatientID))
	if err != nil {
		validation["patient_id"] = "invalid_uuid"
	}

	kID, err := uuid.Parse(strings.TrimSpace(in.KinesiologistID))
	if err != nil {
		validation["kinesiologist_id"] = "invalid_uuid"
	}

	var apptID *uuid.UUID
	if in.AppointmentID != nil && strings.TrimSpace(*in.AppointmentID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(*in.AppointmentID))
		if err != nil {
			validation["appointment_id"] = "invalid_uuid"
		} else {
			apptID = &id
		}
	}

	notes := strings.TrimSpace(in.Notes)
	if notes == "" {
		validation["notes"] = "required"
	}

	if in.PainLevel != nil {
		if *in.PainLevel < 0 || *in.PainLevel > 10 {
			validation["pain_level"] = "must_be_between_0_and_10"
		}
	}

	if len(validation) > 0 {
		return domain.PatientEvolution{}, validation, domain.ErrValidation
	}

	now := time.Now().UTC()

	e := domain.PatientEvolution{
		ID:              uuid.New(),
		PatientID:       patientID,
		KinesiologistID: kID,
		AppointmentID:   apptID,
		PainLevel:       in.PainLevel,
		Notes:           notes,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	out, err := uc.repo.Create(ctx, e)
	if err != nil {
		return domain.PatientEvolution{}, nil, err
	}
	return out, nil, nil
}
