package videocalls

import (
	"context"
	"fmt"
	"strings"
)

type JitsiConfig struct {
	RoomNamePrefix string
}

type JitsiProvider struct {
	roomNamePrefix string
}

func NewJitsiProvider(cfg JitsiConfig) *JitsiProvider {
	prefix := strings.TrimSpace(cfg.RoomNamePrefix)
	if prefix == "" {
		prefix = "kinesio"
	}
	return &JitsiProvider{roomNamePrefix: prefix}
}

func (p *JitsiProvider) Name() string { return "jitsi" }

// CreateRoom builds a Jitsi Meet room URL deterministically from the
// appointment ID. Jitsi Meet (meet.jit.si) requires no API key or signup:
// any room name resolves to a working room the moment someone joins it.
// Using the appointment's UUID keeps the room name unguessable without
// needing an external API call to reserve it.
func (p *JitsiProvider) CreateRoom(_ context.Context, input CreateRoomInput) (CreatedRoom, error) {
	roomName := fmt.Sprintf("%s-%s", p.roomNamePrefix, strings.ReplaceAll(input.AppointmentID, "-", ""))
	url := "https://meet.jit.si/" + roomName

	return CreatedRoom{
		Provider: p.Name(),
		URL:      url,
	}, nil
}
