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

type DailyConfig struct {
	APIKey          string
	ExpiryBufferMin int
}

type DailyProvider struct {
	apiKey          string
	expiryBufferMin int
	client          *http.Client
}

func NewDailyProvider(cfg DailyConfig) *DailyProvider {
	expiryBufferMin := cfg.ExpiryBufferMin
	if expiryBufferMin <= 0 {
		expiryBufferMin = 60
	}
	return &DailyProvider{
		apiKey:          cfg.APIKey,
		expiryBufferMin: expiryBufferMin,
		client:          &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *DailyProvider) Name() string { return "daily" }

func (p *DailyProvider) CreateRoom(ctx context.Context, input CreateRoomInput) (CreatedRoom, error) {
	if strings.TrimSpace(p.apiKey) == "" {
		return CreatedRoom{}, errors.New("daily api key is empty")
	}

	body := map[string]any{
		"properties": map[string]any{
			"nbf": input.StartAt.Add(-15 * time.Minute).Unix(),
			"exp": input.EndAt.Add(time.Duration(p.expiryBufferMin) * time.Minute).Unix(),
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return CreatedRoom{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.daily.co/v1/rooms", bytes.NewReader(payload))
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
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return CreatedRoom{}, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return CreatedRoom{}, fmt.Errorf("daily create room failed: status %d", res.StatusCode)
	}
	if strings.TrimSpace(out.URL) == "" {
		return CreatedRoom{}, errors.New("daily response did not include room url")
	}

	var meetingID *string
	if strings.TrimSpace(out.Name) != "" {
		value := strings.TrimSpace(out.Name)
		meetingID = &value
	}
	return CreatedRoom{
		Provider:  p.Name(),
		MeetingID: meetingID,
		URL:       strings.TrimSpace(out.URL),
	}, nil
}
