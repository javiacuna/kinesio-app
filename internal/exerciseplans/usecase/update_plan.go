package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/exerciseplans/domain"
)

type UpdatePlanInput struct {
	KinesiologistID string                `json:"kinesiologist_id"`
	DiagnosisID     *string               `json:"patient_diagnosis_id"`
	Frequency       string                `json:"frequency"`
	DurationWeeks   int                   `json:"duration_weeks"`
	Observations    *string               `json:"observations"`
	Status          string                `json:"status"`
	Items           []CreatePlanItemInput `json:"items"`
	ActorEmail      string                `json:"-"`
	ActorRole       string                `json:"-"`
}

type UpdatePlanUseCase struct {
	repo domain.Repository
}

func NewUpdatePlanUseCase(repo domain.Repository) *UpdatePlanUseCase {
	return &UpdatePlanUseCase{repo: repo}
}

func (uc *UpdatePlanUseCase) Execute(ctx context.Context, id string, in UpdatePlanInput) (domain.ExercisePlan, map[string]string, error) {
	validation := map[string]string{}

	planID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		validation["plan_id"] = "invalid_uuid"
	}
	kID, err := uuid.Parse(strings.TrimSpace(in.KinesiologistID))
	if err != nil {
		validation["kinesiologist_id"] = "invalid_uuid"
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

	freq := domain.Frequency(strings.TrimSpace(in.Frequency))
	if freq != domain.FrequencyDaily && freq != domain.FrequencyWeekly {
		validation["frequency"] = "must_be_daily_or_weekly"
	}
	if in.DurationWeeks <= 0 {
		validation["duration_weeks"] = "must_be_>=_1"
	}

	status := domain.PlanStatus(strings.TrimSpace(in.Status))
	if status == "" {
		status = domain.PlanActive
	}
	if status != domain.PlanActive && status != domain.PlanClosed {
		validation["status"] = "must_be_active_or_closed"
	}
	if len(in.Items) == 0 {
		validation["items"] = "must_have_at_least_one_item"
	}
	for i, it := range in.Items {
		if strings.TrimSpace(it.Name) == "" {
			validation[fmt.Sprintf("items[%d].name", i)] = "required"
		}
		if it.EstimatedMinutes <= 0 {
			validation[fmt.Sprintf("items[%d].estimated_minutes", i)] = "must_be_>_0"
		}
	}

	if len(validation) > 0 {
		return domain.ExercisePlan{}, validation, domain.ErrValidation
	}

	current, found, err := uc.repo.GetByID(ctx, planID)
	if err != nil {
		return domain.ExercisePlan{}, nil, err
	}
	if !found {
		return domain.ExercisePlan{}, nil, domain.ErrNotFound
	}

	now := time.Now().UTC()
	plan := domain.ExercisePlan{
		ID:              current.ID,
		PatientID:       current.PatientID,
		KinesiologistID: kID,
		DiagnosisID:     diagnosisID,
		Frequency:       freq,
		DurationWeeks:   in.DurationWeeks,
		Observations:    in.Observations,
		Status:          status,
		Items:           make([]domain.ExercisePlanItem, 0, len(in.Items)),
		CreatedAt:       current.CreatedAt,
		UpdatedAt:       now,
		CreatedByEmail:  current.CreatedByEmail,
		CreatedByRole:   current.CreatedByRole,
		UpdatedByEmail:  trimOptionalString(in.ActorEmail),
		UpdatedByRole:   trimOptionalString(in.ActorRole),
	}

	for _, it := range in.Items {
		plan.Items = append(plan.Items, domain.ExercisePlanItem{
			ID:               uuid.New(),
			PlanID:           plan.ID,
			Name:             strings.TrimSpace(it.Name),
			Description:      it.Description,
			VideoURL:         it.VideoURL,
			GuideURL:         it.GuideURL,
			EstimatedMinutes: it.EstimatedMinutes,
			Sets:             it.Sets,
			Reps:             it.Reps,
			CreatedAt:        now,
			UpdatedAt:        now,
		})
	}

	out, err := uc.repo.Update(ctx, plan)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ExercisePlan{}, nil, domain.ErrNotFound
		}
		return domain.ExercisePlan{}, nil, err
	}
	return out, nil, nil
}
