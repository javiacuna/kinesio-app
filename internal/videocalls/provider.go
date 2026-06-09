package videocalls

import (
	"context"
	"time"

	"github.com/javiacuna/kinesio-backend/internal/config"
)

type CreateRoomInput struct {
	AppointmentID string
	StartAt       time.Time
	EndAt         time.Time
}

type CreatedRoom struct {
	Provider  string
	MeetingID *string
	URL       string
}

type Provider interface {
	Name() string
	CreateRoom(ctx context.Context, input CreateRoomInput) (CreatedRoom, error)
}

func NewProvider(cfg config.Config) Provider {
	switch cfg.VideoCallProvider {
	case "daily":
		if cfg.DailyAPIKey == "" {
			return nil
		}
		return NewDailyProvider(DailyConfig{
			APIKey:          cfg.DailyAPIKey,
			ExpiryBufferMin: cfg.VideoCallExpiryMin,
		})
	case "whereby":
		if cfg.WherebyAPIKey == "" {
			return nil
		}
		return NewWherebyProvider(WherebyConfig{
			APIKey:          cfg.WherebyAPIKey,
			RoomNamePrefix:  cfg.WherebyRoomPrefix,
			ExpiryBufferMin: cfg.VideoCallExpiryMin,
		})
	default:
		return nil
	}
}
