package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/exerciseplans/domain"
	"github.com/javiacuna/kinesio-backend/internal/exerciseplans/usecase"
)

type exercisePlanRepo struct {
	created  domain.ExercisePlan
	listed   []domain.ExercisePlan
	listedID uuid.UUID
}

func (r *exercisePlanRepo) Create(ctx context.Context, p domain.ExercisePlan) (domain.ExercisePlan, error) {
	r.created = p
	return p, nil
}

func (r *exercisePlanRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.ExercisePlan, bool, error) {
	return domain.ExercisePlan{}, false, nil
}

func (r *exercisePlanRepo) ListByPatient(ctx context.Context, patientID uuid.UUID) ([]domain.ExercisePlan, error) {
	r.listedID = patientID
	return r.listed, nil
}

func TestCreatePlanEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patientID := uuid.New()
	kinesiologistID := uuid.New()

	tests := []struct {
		name       string
		route      string
		body       map[string]any
		wantStatus int
		wantError  string
	}{
		{
			name:  "creates plan from post plans",
			route: "/plans",
			body: map[string]any{
				"patient_id":       patientID.String(),
				"kinesiologist_id": kinesiologistID.String(),
				"frequency":        "weekly",
				"duration_weeks":   4,
				"items": []map[string]any{
					{
						"name":              "Sentadillas",
						"estimated_minutes": 10,
						"sets":              3,
						"reps":              12,
					},
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:  "creates and assigns plan from patient route",
			route: "/patients/" + patientID.String() + "/plans",
			body: map[string]any{
				"kinesiologist_id": kinesiologistID.String(),
				"frequency":        "daily",
				"duration_weeks":   2,
				"items": []map[string]any{
					{
						"name":              "Elongacion",
						"estimated_minutes": 8,
					},
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:  "validates body",
			route: "/plans",
			body: map[string]any{
				"patient_id":       "bad-patient",
				"kinesiologist_id": "bad-kine",
				"frequency":        "monthly",
				"duration_weeks":   0,
				"items": []map[string]any{
					{
						"name":              "",
						"estimated_minutes": 0,
					},
				},
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "validation_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &exercisePlanRepo{}
			createUC := usecase.NewCreatePlanUseCase(repo)
			handler := NewHandler(createUC, nil, nil)

			router := gin.New()
			router.POST("/plans", handler.Create)
			router.POST("/patients/:patient_id/plans", handler.CreateForPatient)

			payload, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("marshal body: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, tt.route, bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			var got map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}

			if tt.wantError != "" {
				if got["error"] != tt.wantError {
					t.Fatalf("error = %v, want %q", got["error"], tt.wantError)
				}
				return
			}

			if got["patient_id"] != patientID.String() {
				t.Fatalf("patient_id = %v, want %s", got["patient_id"], patientID.String())
			}
			if got["kinesiologist_id"] != kinesiologistID.String() {
				t.Fatalf("kinesiologist_id = %v, want %s", got["kinesiologist_id"], kinesiologistID.String())
			}
			if repo.created.PatientID != patientID {
				t.Fatalf("created patient id = %s, want %s", repo.created.PatientID, patientID)
			}
		})
	}
}

func TestListPatientPlansEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patientID := uuid.New()
	kinesiologistID := uuid.New()
	planID := uuid.New()
	now := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		route      string
		wantStatus int
		wantError  string
	}{
		{
			name:       "lists plans from short route",
			route:      "/patients/" + patientID.String() + "/plans",
			wantStatus: http.StatusOK,
		},
		{
			name:       "validates patient id",
			route:      "/patients/not-a-uuid/plans",
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid_patient_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &exercisePlanRepo{
				listed: []domain.ExercisePlan{
					{
						ID:              planID,
						PatientID:       patientID,
						KinesiologistID: kinesiologistID,
						Frequency:       domain.FrequencyWeekly,
						DurationWeeks:   4,
						Status:          domain.PlanActive,
						CreatedAt:       now,
						UpdatedAt:       now,
						Items: []domain.ExercisePlanItem{
							{
								ID:               uuid.New(),
								PlanID:           planID,
								Name:             "Sentadillas",
								EstimatedMinutes: 10,
								CreatedAt:        now,
								UpdatedAt:        now,
							},
						},
					},
				},
			}
			listUC := usecase.NewListPlansByPatientUseCase(repo)
			handler := NewHandler(nil, listUC, nil)

			router := gin.New()
			router.GET("/patients/:patient_id/plans", handler.ListByPatient)

			req := httptest.NewRequest(http.MethodGet, tt.route, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantError != "" {
				var got map[string]any
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("unmarshal response: %v", err)
				}
				if got["error"] != tt.wantError {
					t.Fatalf("error = %v, want %q", got["error"], tt.wantError)
				}
				return
			}

			if repo.listedID != patientID {
				t.Fatalf("listed patient id = %s, want %s", repo.listedID, patientID)
			}

			var got []map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
			if len(got) != 1 {
				t.Fatalf("len = %d, want 1", len(got))
			}
			if got[0]["id"] != planID.String() {
				t.Fatalf("id = %v, want %s", got[0]["id"], planID.String())
			}
		})
	}
}
