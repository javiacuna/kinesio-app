package usecase

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/evolutions/domain"
)

type CreateEvolutionInput struct {
	PatientID       string                      `json:"patient_id"`
	KinesiologistID string                      `json:"kinesiologist_id"`
	AppointmentID   *string                     `json:"appointment_id,omitempty"`
	DiagnosisID     *string                     `json:"patient_diagnosis_id,omitempty"`
	PainLevel       *int                        `json:"pain_level,omitempty"`
	MobilityScore   *int                        `json:"mobility_score,omitempty"`
	StrengthScore   *int                        `json:"strength_score,omitempty"`
	FunctionalScore *int                        `json:"functional_score,omitempty"`
	Notes           string                      `json:"notes"`
	Photos          []CreateEvolutionPhotoInput `json:"photos,omitempty"`
}

type CreateEvolutionPhotoInput struct {
	URL     string  `json:"url"`
	Caption *string `json:"caption,omitempty"`
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

	var diagnosisID *uuid.UUID
	if in.DiagnosisID != nil && strings.TrimSpace(*in.DiagnosisID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(*in.DiagnosisID))
		if err != nil {
			validation["patient_diagnosis_id"] = "invalid_uuid"
		} else {
			diagnosisID = &id
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
	validatePercentScore(validation, "mobility_score", in.MobilityScore)
	validatePercentScore(validation, "strength_score", in.StrengthScore)
	validatePercentScore(validation, "functional_score", in.FunctionalScore)
	for i, photo := range in.Photos {
		if strings.TrimSpace(photo.URL) == "" {
			validation["photos["+strconv.Itoa(i)+"].url"] = "required"
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
		DiagnosisID:     diagnosisID,
		PainLevel:       in.PainLevel,
		MobilityScore:   in.MobilityScore,
		StrengthScore:   in.StrengthScore,
		FunctionalScore: in.FunctionalScore,
		Notes:           notes,
		Photos:          make([]domain.PatientEvolutionPhoto, 0, len(in.Photos)),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	for _, photo := range in.Photos {
		e.Photos = append(e.Photos, domain.PatientEvolutionPhoto{
			ID:          uuid.New(),
			EvolutionID: e.ID,
			URL:         strings.TrimSpace(photo.URL),
			Caption:     trimPtr(photo.Caption),
			CreatedAt:   now,
		})
	}

	out, err := uc.repo.Create(ctx, e)
	if err != nil {
		return domain.PatientEvolution{}, nil, err
	}
	return out, nil, nil
}

func validatePercentScore(validation map[string]string, field string, value *int) {
	if value == nil {
		return
	}
	if *value < 0 || *value > 100 {
		validation[field] = "must_be_between_0_and_100"
	}
}

func trimPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}
