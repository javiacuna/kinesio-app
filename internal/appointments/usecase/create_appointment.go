package usecase

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/appointments/domain"
	"github.com/javiacuna/kinesio-backend/internal/appointments/ports"
)

type CreateAppointmentInput struct {
	PatientID       string
	KinesiologistID string
	PracticeID      *string
	FinancierID     *string
	StartAt         string // RFC3339
	EndAt           string // RFC3339
	Modality        *string
	VideoCallURL    *string
	Notes           *string
}

type CreateAppointmentUseCase struct {
	repo ports.Repository
}

func NewCreateAppointmentUseCase(repo ports.Repository) *CreateAppointmentUseCase {
	return &CreateAppointmentUseCase{repo: repo}
}

func (uc *CreateAppointmentUseCase) Execute(ctx context.Context, in CreateAppointmentInput) (domain.Appointment, map[string]string, error) {
	errs := map[string]string{}

	pid, err := uuid.Parse(strings.TrimSpace(in.PatientID))
	if err != nil {
		errs["patient_id"] = "UUID inválido"
	}
	kid, err := uuid.Parse(strings.TrimSpace(in.KinesiologistID))
	if err != nil {
		errs["kinesiologist_id"] = "UUID inválido"
	}
	var practiceID *uuid.UUID
	if in.PracticeID != nil && strings.TrimSpace(*in.PracticeID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(*in.PracticeID))
		if err != nil {
			errs["practice_id"] = "UUID inválido"
		} else {
			practiceID = &id
		}
	}
	var financierID *uuid.UUID
	if in.FinancierID != nil && strings.TrimSpace(*in.FinancierID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(*in.FinancierID))
		if err != nil {
			errs["financier_id"] = "UUID inválido"
		} else {
			financierID = &id
		}
	}

	startAt, startErr := time.Parse(time.RFC3339, strings.TrimSpace(in.StartAt))
	if startErr != nil {
		errs["start_at"] = "Formato inválido (usar RFC3339)"
	}
	endAt, endErr := time.Parse(time.RFC3339, strings.TrimSpace(in.EndAt))
	if endErr != nil {
		errs["end_at"] = "Formato inválido (usar RFC3339)"
	}
	if startErr == nil && !startAt.After(time.Now().UTC()) {
		errs["start_at"] = "No se pueden crear turnos en horarios pasados"
	}
	if startErr == nil && endErr == nil && !endAt.After(startAt) {
		errs["end_at"] = "Debe ser mayor a start_at"
	}
	modality := domain.ModalityInPerson
	if in.Modality != nil && strings.TrimSpace(*in.Modality) != "" {
		switch domain.Modality(strings.TrimSpace(*in.Modality)) {
		case domain.ModalityInPerson, domain.ModalityVirtual:
			modality = domain.Modality(strings.TrimSpace(*in.Modality))
		default:
			errs["modality"] = "Valor inválido (in_person|virtual)"
		}
	}
	videoCallURL := trimPtr(in.VideoCallURL)
	if modality == domain.ModalityInPerson {
		videoCallURL = nil
	}
	if videoCallURL != nil && !isValidVideoCallURL(*videoCallURL) {
		errs["video_call_url"] = "URL inválida"
	}

	if len(errs) > 0 {
		return domain.Appointment{}, errs, domain.ErrValidation
	}

	active, err := uc.repo.IsPatientActive(ctx, pid)
	if err != nil {
		return domain.Appointment{}, nil, err
	}
	if !active {
		return domain.Appointment{}, nil, domain.ErrPatientInactive
	}

	overlap, err := uc.repo.HasOverlap(ctx, kid, startAt.UTC(), endAt.UTC(), nil)
	if err != nil {
		return domain.Appointment{}, nil, err
	}
	if overlap {
		return domain.Appointment{}, nil, domain.ErrOverlap
	}

	a := domain.Appointment{
		ID:              uuid.New(),
		PatientID:       pid,
		KinesiologistID: kid,
		PracticeID:      practiceID,
		FinancierID:     financierID,
		StartAt:         startAt.UTC(),
		EndAt:           endAt.UTC(),
		Status:          domain.StatusScheduled,
		Modality:        modality,
		VideoCallURL:    videoCallURL,
		Notes:           trimPtr(in.Notes),
	}
	created, err := uc.repo.Create(ctx, a)
	if err != nil {
		return domain.Appointment{}, nil, err
	}
	return created, nil, nil
}

func isValidVideoCallURL(value string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
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
