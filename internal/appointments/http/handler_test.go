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

	"github.com/javiacuna/kinesio-backend/internal/appointments/domain"
	"github.com/javiacuna/kinesio-backend/internal/appointments/usecase"
)

type appointmentRepo struct {
	overlap      bool
	createErr    error
	appointment  domain.Appointment
	found        bool
	appointments []domain.Appointment
}

func (r *appointmentRepo) Create(ctx context.Context, a domain.Appointment) (domain.Appointment, error) {
	if r.createErr != nil {
		return domain.Appointment{}, r.createErr
	}
	now := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	a.CreatedAt = now
	a.UpdatedAt = now
	return a, nil
}

func (r *appointmentRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Appointment, bool, error) {
	return r.appointment, r.found, nil
}

func (r *appointmentRepo) Update(ctx context.Context, a domain.Appointment) (domain.Appointment, error) {
	a.CreatedAt = r.appointment.CreatedAt
	a.UpdatedAt = time.Date(2026, 4, 11, 13, 0, 0, 0, time.UTC)
	r.appointment = a
	return a, nil
}

func (r *appointmentRepo) HasOverlap(ctx context.Context, kinesiologistID uuid.UUID, startAt, endAt time.Time, excludeID *uuid.UUID) (bool, error) {
	return r.overlap, nil
}

func (r *appointmentRepo) ListByKinesiologistAndRange(ctx context.Context, kinesiologistID uuid.UUID, startDay, endDay time.Time) ([]domain.Appointment, error) {
	return r.appointments, nil
}

func (r *appointmentRepo) ListByPatientAndRange(ctx context.Context, patientID uuid.UUID, from time.Time, to time.Time) ([]domain.Appointment, error) {
	return nil, nil
}

func TestCreateAppointmentEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patientID := uuid.New()
	kinesiologistID := uuid.New()

	tests := []struct {
		name       string
		repo       *appointmentRepo
		body       map[string]any
		auth       string
		wantStatus int
		wantError  string
	}{
		{
			name: "creates appointment",
			repo: &appointmentRepo{},
			body: map[string]any{
				"patient_id":       patientID.String(),
				"kinesiologist_id": kinesiologistID.String(),
				"start_at":         "2026-04-11T14:00:00Z",
				"end_at":           "2026-04-11T14:30:00Z",
				"notes":            " Primera visita ",
			},
			auth:       "Bearer demo-recepcionista-token",
			wantStatus: http.StatusCreated,
		},
		{
			name: "requires receptionist token",
			repo: &appointmentRepo{},
			body: map[string]any{
				"patient_id":       patientID.String(),
				"kinesiologist_id": kinesiologistID.String(),
				"start_at":         "2026-04-11T14:00:00Z",
				"end_at":           "2026-04-11T14:30:00Z",
			},
			wantStatus: http.StatusUnauthorized,
			wantError:  "unauthorized",
		},
		{
			name: "validates ids and time range",
			repo: &appointmentRepo{},
			body: map[string]any{
				"patient_id":       "bad-patient",
				"kinesiologist_id": "bad-kine",
				"start_at":         "2026-04-11T14:30:00Z",
				"end_at":           "2026-04-11T14:00:00Z",
			},
			auth:       "Bearer demo-recepcionista-token",
			wantStatus: http.StatusBadRequest,
			wantError:  "validation_error",
		},
		{
			name: "rejects overlap",
			repo: &appointmentRepo{overlap: true},
			body: map[string]any{
				"patient_id":       patientID.String(),
				"kinesiologist_id": kinesiologistID.String(),
				"start_at":         "2026-04-11T14:00:00Z",
				"end_at":           "2026-04-11T14:30:00Z",
			},
			auth:       "Bearer demo-recepcionista-token",
			wantStatus: http.StatusConflict,
			wantError:  "overlap",
		},
		{
			name: "returns patient not found",
			repo: &appointmentRepo{createErr: domain.ErrPatientNotFound},
			body: map[string]any{
				"patient_id":       patientID.String(),
				"kinesiologist_id": kinesiologistID.String(),
				"start_at":         "2026-04-11T14:00:00Z",
				"end_at":           "2026-04-11T14:30:00Z",
			},
			auth:       "Bearer demo-recepcionista-token",
			wantStatus: http.StatusNotFound,
			wantError:  "patient_not_found",
		},
		{
			name: "returns kinesiologist not found",
			repo: &appointmentRepo{createErr: domain.ErrKinesiologistNotFound},
			body: map[string]any{
				"patient_id":       patientID.String(),
				"kinesiologist_id": kinesiologistID.String(),
				"start_at":         "2026-04-11T14:00:00Z",
				"end_at":           "2026-04-11T14:30:00Z",
			},
			auth:       "Bearer demo-recepcionista-token",
			wantStatus: http.StatusNotFound,
			wantError:  "kinesiologist_not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createUC := usecase.NewCreateAppointmentUseCase(tt.repo)
			handler := NewHandler(createUC, nil, nil, nil, nil, nil)

			router := gin.New()
			router.POST("/appointments", handler.Create)

			payload, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("marshal body: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/appointments", bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			if tt.auth != "" {
				req.Header.Set("Authorization", tt.auth)
			}
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

			if _, err := uuid.Parse(got["id"].(string)); err != nil {
				t.Fatalf("id is not a uuid: %v", got["id"])
			}
			if got["patient_id"] != patientID.String() {
				t.Fatalf("patient_id = %v, want %s", got["patient_id"], patientID.String())
			}
			if got["kinesiologist_id"] != kinesiologistID.String() {
				t.Fatalf("kinesiologist_id = %v, want %s", got["kinesiologist_id"], kinesiologistID.String())
			}
			if got["status"] != string(domain.StatusScheduled) {
				t.Fatalf("status = %v, want scheduled", got["status"])
			}
			if got["notes"] != "Primera visita" {
				t.Fatalf("notes = %v, want trimmed notes", got["notes"])
			}
		})
	}
}

func TestUpdateAppointmentEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appointmentID := uuid.New()
	patientID := uuid.New()
	kinesiologistID := uuid.New()
	now := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	existing := domain.Appointment{
		ID:              appointmentID,
		PatientID:       patientID,
		KinesiologistID: kinesiologistID,
		StartAt:         time.Date(2026, 4, 11, 14, 0, 0, 0, time.UTC),
		EndAt:           time.Date(2026, 4, 11, 14, 30, 0, 0, time.UTC),
		Status:          domain.StatusScheduled,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	cancelled := existing
	cancelled.Status = domain.StatusCancelled

	tests := []struct {
		name       string
		id         string
		repo       *appointmentRepo
		body       map[string]any
		method     string
		wantStatus int
		wantError  string
	}{
		{
			name:   "updates appointment with put",
			id:     appointmentID.String(),
			repo:   &appointmentRepo{appointment: existing, found: true},
			method: http.MethodPut,
			body: map[string]any{
				"start_at": "2026-04-11T15:00:00Z",
				"end_at":   "2026-04-11T15:45:00Z",
				"notes":    " Reprogramado ",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "keeps patch compatibility",
			id:     appointmentID.String(),
			repo:   &appointmentRepo{appointment: existing, found: true},
			method: http.MethodPatch,
			body: map[string]any{
				"notes": "Actualizado",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "validates invalid id",
			id:     "not-a-uuid",
			repo:   &appointmentRepo{appointment: existing, found: true},
			method: http.MethodPut,
			body: map[string]any{
				"notes": "Nada",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "validation_error",
		},
		{
			name:   "returns not found",
			id:     appointmentID.String(),
			repo:   &appointmentRepo{found: false},
			method: http.MethodPut,
			body: map[string]any{
				"notes": "Nada",
			},
			wantStatus: http.StatusNotFound,
			wantError:  "not_found",
		},
		{
			name:   "rejects overlap when rescheduling",
			id:     appointmentID.String(),
			repo:   &appointmentRepo{appointment: existing, found: true, overlap: true},
			method: http.MethodPut,
			body: map[string]any{
				"start_at": "2026-04-11T15:00:00Z",
				"end_at":   "2026-04-11T15:45:00Z",
			},
			wantStatus: http.StatusConflict,
			wantError:  "overlap",
		},
		{
			name:   "rejects overlap when reactivating cancelled appointment",
			id:     appointmentID.String(),
			repo:   &appointmentRepo{appointment: cancelled, found: true, overlap: true},
			method: http.MethodPut,
			body: map[string]any{
				"status": "scheduled",
			},
			wantStatus: http.StatusConflict,
			wantError:  "overlap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateUC := usecase.NewUpdateAppointmentUseCase(tt.repo)
			handler := NewHandler(nil, nil, updateUC, nil, nil, nil)

			router := gin.New()
			router.PUT("/appointments/:id", handler.Update)
			router.PATCH("/appointments/:id", handler.Update)

			payload, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("marshal body: %v", err)
			}

			req := httptest.NewRequest(tt.method, "/appointments/"+tt.id, bytes.NewReader(payload))
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

			if tt.wantError != "" {
				if got["error"] != tt.wantError {
					t.Fatalf("error = %v, want %q", got["error"], tt.wantError)
				}
				return
			}

			if got["id"] != appointmentID.String() {
				t.Fatalf("id = %v, want %s", got["id"], appointmentID.String())
			}
		})
	}
}

func TestCancelAppointmentEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appointmentID := uuid.New()
	now := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	existing := domain.Appointment{
		ID:              appointmentID,
		PatientID:       uuid.New(),
		KinesiologistID: uuid.New(),
		StartAt:         time.Date(2026, 4, 11, 14, 0, 0, 0, time.UTC),
		EndAt:           time.Date(2026, 4, 11, 14, 30, 0, 0, time.UTC),
		Status:          domain.StatusScheduled,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	tests := []struct {
		name       string
		id         string
		repo       *appointmentRepo
		body       map[string]any
		wantStatus int
		wantError  string
	}{
		{
			name: "cancels appointment",
			id:   appointmentID.String(),
			repo: &appointmentRepo{appointment: existing, found: true},
			body: map[string]any{
				"reason": "Paciente aviso",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "validates invalid id",
			id:         "not-a-uuid",
			repo:       &appointmentRepo{appointment: existing, found: true},
			body:       map[string]any{},
			wantStatus: http.StatusBadRequest,
			wantError:  "validation_error",
		},
		{
			name:       "returns not found",
			id:         appointmentID.String(),
			repo:       &appointmentRepo{found: false},
			body:       map[string]any{},
			wantStatus: http.StatusNotFound,
			wantError:  "not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cancelUC := usecase.NewCancelAppointmentUseCase(tt.repo)
			handler := NewHandler(nil, nil, nil, cancelUC, nil, nil)

			router := gin.New()
			router.DELETE("/appointments/:id", handler.Cancel)

			payload, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("marshal body: %v", err)
			}

			req := httptest.NewRequest(http.MethodDelete, "/appointments/"+tt.id, bytes.NewReader(payload))
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

			if tt.wantError != "" {
				if got["error"] != tt.wantError {
					t.Fatalf("error = %v, want %q", got["error"], tt.wantError)
				}
				return
			}

			if got["status"] != string(domain.StatusCancelled) {
				t.Fatalf("status = %v, want cancelled", got["status"])
			}
			if got["cancelled_reason"] != "Paciente aviso" {
				t.Fatalf("cancelled_reason = %v, want reason", got["cancelled_reason"])
			}
		})
	}
}

func TestListAppointmentsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appointmentID := uuid.New()
	kinesiologistID := uuid.New()
	now := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	repo := &appointmentRepo{
		appointments: []domain.Appointment{
			{
				ID:              appointmentID,
				PatientID:       uuid.New(),
				KinesiologistID: kinesiologistID,
				StartAt:         time.Date(2026, 4, 11, 14, 0, 0, 0, time.UTC),
				EndAt:           time.Date(2026, 4, 11, 14, 30, 0, 0, time.UTC),
				Status:          domain.StatusScheduled,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
		},
	}

	tests := []struct {
		name       string
		target     string
		wantStatus int
		wantError  string
	}{
		{
			name:       "lists agenda",
			target:     "/appointments?kinesiologist_id=" + kinesiologistID.String() + "&date=2026-04-11",
			wantStatus: http.StatusOK,
		},
		{
			name:       "validates filters",
			target:     "/appointments?kinesiologist_id=bad&date=2026/04/11",
			wantStatus: http.StatusBadRequest,
			wantError:  "validation_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listUC := usecase.NewListAppointmentsDayUseCase(repo)
			handler := NewHandler(nil, listUC, nil, nil, nil, nil)

			router := gin.New()
			router.GET("/appointments", handler.ListDay)

			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
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

			var got []map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
			if len(got) != 1 {
				t.Fatalf("len = %d, want 1", len(got))
			}
			if got[0]["id"] != appointmentID.String() {
				t.Fatalf("id = %v, want %s", got[0]["id"], appointmentID.String())
			}
		})
	}
}
