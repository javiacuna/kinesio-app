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

	"github.com/javiacuna/kinesio-backend/internal/evolutions/domain"
	"github.com/javiacuna/kinesio-backend/internal/evolutions/usecase"
	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
)

type evolutionRepo struct {
	created  domain.PatientEvolution
	listed   []domain.PatientEvolution
	listedID uuid.UUID
}

func (r *evolutionRepo) Create(ctx context.Context, e domain.PatientEvolution) (domain.PatientEvolution, error) {
	r.created = e
	return e, nil
}

func (r *evolutionRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.PatientEvolution, bool, error) {
	return domain.PatientEvolution{}, false, nil
}

func (r *evolutionRepo) ListByPatient(ctx context.Context, patientID uuid.UUID, limit int) ([]domain.PatientEvolution, error) {
	r.listedID = patientID
	return r.listed, nil
}

func TestCreateEvolutionEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patientID := uuid.New()
	kinesiologistID := uuid.New()
	painLevel := 4
	mobilityScore := 70
	strengthScore := 65
	functionalScore := 80

	tests := []struct {
		name       string
		route      string
		body       map[string]any
		wantStatus int
		wantError  string
	}{
		{
			name:  "creates evolution for patient",
			route: "/patients/" + patientID.String() + "/evolutions",
			body: map[string]any{
				"kinesiologist_id": kinesiologistID.String(),
				"pain_level":       painLevel,
				"mobility_score":   mobilityScore,
				"strength_score":   strengthScore,
				"functional_score": functionalScore,
				"notes":            " Mejor movilidad ",
				"photos": []map[string]any{
					{
						"url":     "https://example.com/foto.jpg",
						"caption": "Rodilla",
					},
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:  "validates body",
			route: "/patients/bad-patient/evolutions",
			body: map[string]any{
				"kinesiologist_id": "bad-kine",
				"pain_level":       99,
				"mobility_score":   101,
				"notes":            "",
				"photos": []map[string]any{
					{"url": ""},
				},
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "validation_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &evolutionRepo{}
			createUC := usecase.NewCreateEvolutionUseCase(repo)
			handler := NewHandler(createUC, nil, nil)

			router := gin.New()
			router.POST("/patients/:patient_id/evolutions", handler.CreateForPatient)

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
			if repo.created.MobilityScore == nil || *repo.created.MobilityScore != mobilityScore {
				t.Fatalf("mobility score = %v, want %d", repo.created.MobilityScore, mobilityScore)
			}
		})
	}
}

func TestListPatientEvolutionsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patientID := uuid.New()
	kinesiologistID := uuid.New()
	evolutionID := uuid.New()
	now := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	painLevel := 3
	mobilityScore := 75

	tests := []struct {
		name       string
		route      string
		wantStatus int
		wantError  string
	}{
		{
			name:       "lists evolutions",
			route:      "/patients/" + patientID.String() + "/evolutions",
			wantStatus: http.StatusOK,
		},
		{
			name:       "validates patient id",
			route:      "/patients/not-a-uuid/evolutions",
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid_patient_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &evolutionRepo{
				listed: []domain.PatientEvolution{
					{
						ID:              evolutionID,
						PatientID:       patientID,
						KinesiologistID: kinesiologistID,
						PainLevel:       &painLevel,
						MobilityScore:   &mobilityScore,
						Notes:           "Mejor movilidad",
						CreatedAt:       now,
						UpdatedAt:       now,
					},
				},
			}
			listUC := usecase.NewListEvolutionsByPatientUseCase(repo)
			handler := NewHandler(nil, listUC, nil)

			router := gin.New()
			router.GET("/patients/:patient_id/evolutions", handler.ListByPatient)

			req := httptest.NewRequest(http.MethodGet, tt.route, nil)
			req.Header.Set("Authorization", middleware.DemoReceptionistToken)
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
			if got[0]["id"] != evolutionID.String() {
				t.Fatalf("id = %v, want %s", got[0]["id"], evolutionID.String())
			}
			if got[0]["mobility_score"] != float64(mobilityScore) {
				t.Fatalf("mobility_score = %v, want %d", got[0]["mobility_score"], mobilityScore)
			}
		})
	}
}
