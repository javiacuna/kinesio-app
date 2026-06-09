package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/appointments/domain"
	"github.com/javiacuna/kinesio-backend/internal/appointments/ports"
	"github.com/javiacuna/kinesio-backend/internal/videocalls"
)

type GenerateAppointmentVideoCallUseCase struct {
	repo     ports.Repository
	provider videocalls.Provider
}

func NewGenerateAppointmentVideoCallUseCase(
	repo ports.Repository,
	provider videocalls.Provider,
) *GenerateAppointmentVideoCallUseCase {
	return &GenerateAppointmentVideoCallUseCase{
		repo:     repo,
		provider: provider,
	}
}

func (uc *GenerateAppointmentVideoCallUseCase) Execute(
	ctx context.Context,
	id string,
) (domain.Appointment, map[string]string, error) {
	errs := map[string]string{}
	aid, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		errs["id"] = "UUID inválido"
		return domain.Appointment{}, errs, domain.ErrValidation
	}
	if uc.provider == nil {
		return domain.Appointment{}, nil, domain.ErrVideoProviderMissing
	}

	current, found, err := uc.repo.GetByID(ctx, aid)
	if err != nil {
		return domain.Appointment{}, nil, err
	}
	if !found {
		return domain.Appointment{}, nil, domain.ErrNotFound
	}
	if current.Status != domain.StatusScheduled {
		return domain.Appointment{}, map[string]string{
			"status": "Sólo se puede generar link para turnos programados",
		}, domain.ErrValidation
	}

	room, err := uc.provider.CreateRoom(ctx, videocalls.CreateRoomInput{
		AppointmentID: current.ID.String(),
		StartAt:       current.StartAt,
		EndAt:         current.EndAt,
	})
	if err != nil {
		return domain.Appointment{}, nil, domain.ErrVideoProviderFailed
	}

	provider := strings.TrimSpace(room.Provider)
	if provider == "" {
		provider = uc.provider.Name()
	}
	current.Modality = domain.ModalityVirtual
	current.VideoCallURL = &room.URL
	current.VideoProvider = &provider
	current.VideoMeetingID = room.MeetingID

	updated, err := uc.repo.Update(ctx, current)
	if err != nil {
		return domain.Appointment{}, nil, err
	}
	return updated, nil, nil
}
