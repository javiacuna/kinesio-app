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

	"github.com/javiacuna/kinesio-backend/internal/patients/domain"
	"github.com/javiacuna/kinesio-backend/internal/patients/usecase"
)

type registerPatientRepo struct {
	existsDNI   bool
	existsEmail bool
}

func (r *registerPatientRepo) Create(ctx context.Context, p domain.Patient) (domain.Patient, error) {
	now := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	p.CreatedAt = now
	p.UpdatedAt = now
	return p, nil
}

func (r *registerPatientRepo) ExistsByDNI(ctx context.Context, dni string) (bool, error) {
	return r.existsDNI, nil
}

func (r *registerPatientRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.existsEmail, nil
}

func (r *registerPatientRepo) GetByID(ctx context.Context, id string) (domain.Patient, bool, error) {
	return domain.Patient{}, false, nil
}

func (r *registerPatientRepo) Search(ctx context.Context, query string, limit int) ([]domain.Patient, error) {
	return nil, nil
}

func TestRegisterPatientEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		repo       *registerPatientRepo
		body       map[string]any
		wantStatus int
		wantError  string
	}{
		{
			name: "creates patient",
			repo: &registerPatientRepo{},
			body: map[string]any{
				"dni":        "12345678",
				"first_name": "Javier",
				"last_name":  "Acuna",
				"email":      "Javier@example.com",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "validates required fields and email format",
			repo: &registerPatientRepo{},
			body: map[string]any{
				"dni":        "12.345",
				"first_name": "",
				"last_name":  "Acuna",
				"email":      "malformado",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "validation_error",
		},
		{
			name: "rejects duplicate dni",
			repo: &registerPatientRepo{existsDNI: true},
			body: map[string]any{
				"dni":        "12345678",
				"first_name": "Javier",
				"last_name":  "Acuna",
				"email":      "javier@example.com",
			},
			wantStatus: http.StatusConflict,
			wantError:  "dni_duplicado",
		},
		{
			name: "rejects duplicate email",
			repo: &registerPatientRepo{existsEmail: true},
			body: map[string]any{
				"dni":        "12345678",
				"first_name": "Javier",
				"last_name":  "Acuna",
				"email":      "javier@example.com",
			},
			wantStatus: http.StatusConflict,
			wantError:  "email_duplicado",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registerUC := usecase.NewRegisterPatientUseCase(tt.repo)
			handler := NewHandler(registerUC, nil, nil)

			router := gin.New()
			router.POST("/patients", handler.RegisterPatient)

			payload, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("marshal body: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/patients", bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer demo-recepcionista-token")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			var got map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}

			if tt.wantError != "" && got["error"] != tt.wantError {
				t.Fatalf("error = %v, want %q", got["error"], tt.wantError)
			}

			if tt.wantStatus == http.StatusCreated {
				if _, err := uuid.Parse(got["id"].(string)); err != nil {
					t.Fatalf("id is not a uuid: %v", got["id"])
				}
				if got["email"] != "javier@example.com" {
					t.Fatalf("email = %v, want normalized email", got["email"])
				}
			}
		})
	}
}
