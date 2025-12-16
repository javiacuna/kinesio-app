package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/exerciseplans/domain"
)

type CreatePlanItemInput struct {
	Name             string  `json:"name"`
	Description      *string `json:"description"`
	VideoURL         *string `json:"video_url"`
	GuideURL         *string `json:"guide_url"`
	EstimatedMinutes int     `json:"estimated_minutes"`
	Sets             *int    `json:"sets"`
	Reps             *int    `json:"reps"`
}

type CreatePlanInput struct {
	PatientID       string                `json:"patient_id"`
	KinesiologistID string                `json:"kinesiologist_id"`
	Frequency       string                `json:"frequency"`      // daily|weekly
	DurationWeeks   int                   `json:"duration_weeks"` // >=1
	Observations    *string               `json:"observations"`
	Items           []CreatePlanItemInput `json:"items"`
}

type CreatePlanUseCase struct {
	repo domain.Repository
}

func NewCreatePlanUseCase(repo domain.Repository) *CreatePlanUseCase {
	return &CreatePlanUseCase{repo: repo}
}

func (uc *CreatePlanUseCase) Execute(ctx context.Context, in CreatePlanInput) (domain.ExercisePlan, map[string]string, error) {
	validation := map[string]string{}

	patientID, err := uuid.Parse(strings.TrimSpace(in.PatientID))
	if err != nil {
		validation["patient_id"] = "invalid_uuid"
	}
	kID, err := uuid.Parse(strings.TrimSpace(in.KinesiologistID))
	if err != nil {
		validation["kinesiologist_id"] = "invalid_uuid"
	}

	freq := domain.Frequency(strings.TrimSpace(in.Frequency))
	if freq != domain.FrequencyDaily && freq != domain.FrequencyWeekly {
		validation["frequency"] = "must_be_daily_or_weekly"
	}
	if in.DurationWeeks <= 0 {
		validation["duration_weeks"] = "must_be_>=_1"
	}
	if len(in.Items) == 0 {
		validation["items"] = "must_have_at_least_one_item"
	}

	for i, it := range in.Items {
		if strings.TrimSpace(it.Name) == "" {
			validation[("items[" + string(rune(i)) + "].name")] = "required"
		}
		if it.EstimatedMinutes <= 0 {
			validation[("items[" + string(rune(i)) + "].estimated_minutes")] = "must_be_>_0"
		}
	}

	if len(validation) > 0 {
		return domain.ExercisePlan{}, validation, domain.ErrValidation
	}

	now := time.Now().UTC()

	plan := domain.ExercisePlan{
		ID:              uuid.New(),
		PatientID:       patientID,
		KinesiologistID: kID,
		Frequency:       freq,
		DurationWeeks:   in.DurationWeeks,
		Observations:    in.Observations,
		Status:          domain.PlanActive,
		Items:           make([]domain.ExercisePlanItem, 0, len(in.Items)),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	for _, it := range in.Items {
		name := strings.TrimSpace(it.Name)
		item := domain.ExercisePlanItem{
			ID:               uuid.New(),
			PlanID:           plan.ID,
			Name:             name,
			Description:      it.Description,
			VideoURL:         it.VideoURL,
			GuideURL:         it.GuideURL,
			EstimatedMinutes: it.EstimatedMinutes,
			Sets:             it.Sets,
			Reps:             it.Reps,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		plan.Items = append(plan.Items, item)
	}

	out, err := uc.repo.Create(ctx, plan)
	if err != nil {
		return domain.ExercisePlan{}, nil, err
	}
	return out, nil, nil
}
