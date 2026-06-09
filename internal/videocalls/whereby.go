package videocalls

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type WherebyConfig struct {
	APIKey          string
	RoomNamePrefix  string
	ExpiryBufferMin int
}

type WherebyProvider struct {
	apiKey          string
	roomNamePrefix  string
	expiryBufferMin int
	client          *http.Client
}

func NewWherebyProvider(cfg WherebyConfig) *WherebyProvider {
	expiryBufferMin := cfg.ExpiryBufferMin
	if expiryBufferMin <= 0 {
		expiryBufferMin = 60
	}
	return &WherebyProvider{
		apiKey:          cfg.APIKey,
		roomNamePrefix:  strings.TrimSpace(cfg.RoomNamePrefix),
		expiryBufferMin: expiryBufferMin,
		client:          &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *WherebyProvider) Name() string { return "whereby" }

func (p *WherebyProvider) CreateRoom(ctx context.Context, input CreateRoomInput) (CreatedRoom, error) {
	if strings.TrimSpace(p.apiKey) == "" {
		return CreatedRoom{}, errors.New("whereby api key is empty")
	}

	body := map[string]any{
		"startDate":       input.StartAt.Format(time.RFC3339),
		"endDate":         input.EndAt.Add(time.Duration(p.expiryBufferMin) * time.Minute).Format(time.RFC3339),
		"roomNamePattern": "uuid",
	}
	if p.roomNamePrefix != "" {
		body["roomNamePrefix"] = p.roomNamePrefix
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return CreatedRoom{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.whereby.dev/v1/meetings", bytes.NewReader(payload))
	if err != nil {
		return CreatedRoom{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := p.client.Do(req)
	if err != nil {
		return CreatedRoom{}, err
	}
	defer res.Body.Close()

	var out struct {
		MeetingID string `json:"meetingId"`
		RoomURL   string `json:"roomUrl"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return CreatedRoom{}, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return CreatedRoom{}, fmt.Errorf("whereby create meeting failed: status %d", res.StatusCode)
	}
	if strings.TrimSpace(out.RoomURL) == "" {
		return CreatedRoom{}, errors.New("whereby response did not include room url")
	}

	var meetingID *string
	if strings.TrimSpace(out.MeetingID) != "" {
		value := strings.TrimSpace(out.MeetingID)
		meetingID = &value
	}
	return CreatedRoom{
		Provider:  p.Name(),
		MeetingID: meetingID,
		URL:       strings.TrimSpace(out.RoomURL),
	}, nil
}
