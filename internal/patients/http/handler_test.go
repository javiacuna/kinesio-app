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

	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	"github.com/javiacuna/kinesio-backend/internal/patients/domain"
	"github.com/javiacuna/kinesio-backend/internal/patients/usecase"
)

type registerPatientRepo struct {
	existsDNI   bool
	existsEmail bool
	patient     domain.Patient
	patients    []domain.Patient
	found       bool
	activeID    string
	activeValue bool
	activeErr   error
}

func (r *registerPatientRepo) Create(ctx context.Context, p domain.Patient) (domain.Patient, error) {
	now := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	p.CreatedAt = now
	p.UpdatedAt = now
	return p, nil
}

func (r *registerPatientRepo) Update(ctx context.Context, p domain.Patient) (domain.Patient, error) {
	p.CreatedAt = r.patient.CreatedAt
	p.UpdatedAt = time.Date(2026, 4, 11, 13, 0, 0, 0, time.UTC)
	r.patient = p
	return p, nil
}

func (r *registerPatientRepo) SetActive(ctx context.Context, id string, active bool) error {
	r.activeID = id
	r.activeValue = active
	if r.activeErr != nil {
		return r.activeErr
	}
	return nil
}

func (r *registerPatientRepo) ExistsByDNI(ctx context.Context, dni string) (bool, error) {
	return r.existsDNI, nil
}

func (r *registerPatientRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.existsEmail, nil
}

func (r *registerPatientRepo) GetByID(ctx context.Context, id string) (domain.Patient, bool, error) {
	return r.patient, r.found, nil
}

func (r *registerPatientRepo) FindByDNIAndEmail(ctx context.Context, dni, email string) (domain.Patient, bool, error) {
	return domain.Patient{}, false, nil
}

func (r *registerPatientRepo) List(ctx context.Context, limit int, offset int, includeInactive bool) ([]domain.Patient, error) {
	if offset > len(r.patients) {
		return []domain.Patient{}, nil
	}
	end := offset + limit
	if end > len(r.patients) {
		end = len(r.patients)
	}
	return r.patients[offset:end], nil
}

func (r *registerPatientRepo) Search(ctx context.Context, query string, limit int, includeInactive bool) ([]domain.Patient, error) {
	return r.patients, nil
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
			handler := NewHandler(registerUC, nil, nil, nil, nil, nil)

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

func TestUpdatePatientEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patientID := uuid.New()
	now := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	existing := domain.Patient{
		ID:        patientID,
		DNI:       "12345678",
		FirstName: "Javier",
		LastName:  "Acuna",
		Email:     "javier@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name       string
		id         string
		repo       *registerPatientRepo
		body       map[string]any
		wantStatus int
		wantError  string
	}{
		{
			name: "updates patient",
			id:   patientID.String(),
			repo: &registerPatientRepo{patient: existing, found: true},
			body: map[string]any{
				"dni":            "12345678",
				"first_name":     "Javi",
				"last_name":      "Acuna",
				"email":          "Javi@example.com",
				"phone":          " 3511234567 ",
				"clinical_notes": " Sin dolor ",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "validates invalid id and body",
			id:   "not-a-uuid",
			repo: &registerPatientRepo{patient: existing, found: true},
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
			name: "returns not found",
			id:   patientID.String(),
			repo: &registerPatientRepo{found: false},
			body: map[string]any{
				"dni":        "12345678",
				"first_name": "Javier",
				"last_name":  "Acuna",
				"email":      "javier@example.com",
			},
			wantStatus: http.StatusNotFound,
			wantError:  "not_found",
		},
		{
			name: "rejects duplicate dni",
			id:   patientID.String(),
			repo: &registerPatientRepo{patient: existing, found: true, existsDNI: true},
			body: map[string]any{
				"dni":        "87654321",
				"first_name": "Javier",
				"last_name":  "Acuna",
				"email":      "javier@example.com",
			},
			wantStatus: http.StatusConflict,
			wantError:  "dni_duplicado",
		},
		{
			name: "rejects duplicate email",
			id:   patientID.String(),
			repo: &registerPatientRepo{patient: existing, found: true, existsEmail: true},
			body: map[string]any{
				"dni":        "12345678",
				"first_name": "Javier",
				"last_name":  "Acuna",
				"email":      "otro@example.com",
			},
			wantStatus: http.StatusConflict,
			wantError:  "email_duplicado",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateUC := usecase.NewUpdatePatientUseCase(tt.repo)
			handler := NewHandler(nil, updateUC, nil, nil, nil, nil)

			router := gin.New()
			router.PUT("/patients/:patient_id", handler.UpdatePatient)

			payload, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("marshal body: %v", err)
			}

			req := httptest.NewRequest(http.MethodPut, "/patients/"+tt.id, bytes.NewReader(payload))
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

			if tt.wantStatus == http.StatusOK {
				if got["id"] != patientID.String() {
					t.Fatalf("id = %v, want %s", got["id"], patientID.String())
				}
				if got["first_name"] != "Javi" {
					t.Fatalf("first_name = %v, want Javi", got["first_name"])
				}
				if got["email"] != "javi@example.com" {
					t.Fatalf("email = %v, want normalized email", got["email"])
				}
				if got["phone"] != "3511234567" {
					t.Fatalf("phone = %v, want trimmed phone", got["phone"])
				}
			}
		})
	}
}

func TestListPatientsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	patients := []domain.Patient{
		{ID: uuid.New(), DNI: "11111111", FirstName: "Ana", LastName: "Lopez", Email: "ana@example.com", CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), DNI: "22222222", FirstName: "Bruno", LastName: "Perez", Email: "bruno@example.com", CreatedAt: now, UpdatedAt: now},
	}
	repo := &registerPatientRepo{patients: patients}
	listUC := usecase.NewListPatientsUseCase(repo)
	searchUC := usecase.NewSearchPatientsUseCase(repo)
	handler := NewHandler(nil, nil, nil, nil, listUC, searchUC)

	tests := []struct {
		name       string
		target     string
		wantStatus int
		wantLen    int
	}{
		{
			name:       "lists patients",
			target:     "/patients",
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name:       "lists patients with pagination",
			target:     "/patients?limit=1&offset=1",
			wantStatus: http.StatusOK,
			wantLen:    1,
		},
		{
			name:       "keeps search behavior when query is present",
			target:     "/patients?query=ana&limit=20",
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/patients", handler.Search)

			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			req.Header.Set("Authorization", "Bearer demo-recepcionista-token")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			var got []map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestSearchPatientsAllowsKinesiologist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	repo := &registerPatientRepo{patients: []domain.Patient{
		{ID: uuid.New(), DNI: "11111111", FirstName: "Ana", LastName: "Lopez", Email: "ana@example.com", CreatedAt: now, UpdatedAt: now},
	}}
	handler := NewHandler(nil, nil, nil, nil, usecase.NewListPatientsUseCase(repo), usecase.NewSearchPatientsUseCase(repo))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.AuthUserKey, middleware.AuthUser{UID: "kin", Email: "kin@example.com", Role: "kinesiologo"})
		c.Next()
	})
	router.GET("/patients", handler.Search)

	req := httptest.NewRequest(http.MethodGet, "/patients?query=ana", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestDeletePatientEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patientID := uuid.New()

	tests := []struct {
		name       string
		id         string
		repo       *registerPatientRepo
		wantStatus int
		wantError  string
	}{
		{
			name:       "archives patient",
			id:         patientID.String(),
			repo:       &registerPatientRepo{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "validates invalid id",
			id:         "not-a-uuid",
			repo:       &registerPatientRepo{},
			wantStatus: http.StatusBadRequest,
			wantError:  "validation_error",
		},
		{
			name:       "returns not found",
			id:         patientID.String(),
			repo:       &registerPatientRepo{activeErr: domain.ErrNotFound},
			wantStatus: http.StatusNotFound,
			wantError:  "not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleteUC := usecase.NewDeletePatientUseCase(tt.repo, nil)
			handler := NewHandler(nil, nil, deleteUC, nil, nil, nil)

			router := gin.New()
			router.DELETE("/patients/:patient_id", handler.DeletePatient)

			req := httptest.NewRequest(http.MethodDelete, "/patients/"+tt.id, nil)
			req.Header.Set("Authorization", "Bearer demo-recepcionista-token")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantError == "" {
				if rec.Body.Len() != 0 {
					t.Fatalf("body = %q, want empty body", rec.Body.String())
				}
				if tt.repo.activeID != tt.id {
					t.Fatalf("activeID = %q, want %q", tt.repo.activeID, tt.id)
				}
				if tt.repo.activeValue {
					t.Fatalf("activeValue = true, want false")
				}
				return
			}

			var got map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
			if got["error"] != tt.wantError {
				t.Fatalf("error = %v, want %q", got["error"], tt.wantError)
			}
		})
	}
}
