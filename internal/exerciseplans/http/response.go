package http

import (
	"time"

	"github.com/javiacuna/kinesio-backend/internal/exerciseplans/domain"
)

type planItemResponse struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Description      *string `json:"description,omitempty"`
	VideoURL         *string `json:"video_url,omitempty"`
	GuideURL         *string `json:"guide_url,omitempty"`
	EstimatedMinutes int     `json:"estimated_minutes"`
	Sets             *int    `json:"sets,omitempty"`
	Reps             *int    `json:"reps,omitempty"`
}

type planResponse struct {
	ID              string             `json:"id"`
	PatientID       string             `json:"patient_id"`
	KinesiologistID string             `json:"kinesiologist_id"`
	DiagnosisID     *string            `json:"patient_diagnosis_id,omitempty"`
	Frequency       string             `json:"frequency"`
	DurationWeeks   int                `json:"duration_weeks"`
	Observations    *string            `json:"observations,omitempty"`
	Status          string             `json:"status"`
	Items           []planItemResponse `json:"items"`
	CreatedAt       string             `json:"created_at"`
	UpdatedAt       string             `json:"updated_at"`
	CreatedByEmail  *string            `json:"created_by_email,omitempty"`
	CreatedByRole   *string            `json:"created_by_role,omitempty"`
	UpdatedByEmail  *string            `json:"updated_by_email,omitempty"`
	UpdatedByRole   *string            `json:"updated_by_role,omitempty"`
}

func toResponse(p domain.ExercisePlan) planResponse {
	var diagnosisID *string
	if p.DiagnosisID != nil {
		value := p.DiagnosisID.String()
		diagnosisID = &value
	}

	items := make([]planItemResponse, 0, len(p.Items))
	for _, it := range p.Items {
		items = append(items, planItemResponse{
			ID:               it.ID.String(),
			Name:             it.Name,
			Description:      it.Description,
			VideoURL:         it.VideoURL,
			GuideURL:         it.GuideURL,
			EstimatedMinutes: it.EstimatedMinutes,
			Sets:             it.Sets,
			Reps:             it.Reps,
		})
	}
	return planResponse{
		ID:              p.ID.String(),
		PatientID:       p.PatientID.String(),
		KinesiologistID: p.KinesiologistID.String(),
		DiagnosisID:     diagnosisID,
		Frequency:       string(p.Frequency),
		DurationWeeks:   p.DurationWeeks,
		Observations:    p.Observations,
		Status:          string(p.Status),
		Items:           items,
		CreatedAt:       p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       p.UpdatedAt.UTC().Format(time.RFC3339),
		CreatedByEmail:  p.CreatedByEmail,
		CreatedByRole:   p.CreatedByRole,
		UpdatedByEmail:  p.UpdatedByEmail,
		UpdatedByRole:   p.UpdatedByRole,
	}
}
