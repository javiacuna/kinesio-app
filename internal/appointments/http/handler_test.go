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
	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	kineDomain "github.com/javiacuna/kinesio-backend/internal/kinesiologists/domain"
	patientDomain "github.com/javiacuna/kinesio-backend/internal/patients/domain"
	"github.com/javiacuna/kinesio-backend/internal/videocalls"
)

type appointmentRepo struct {
	overlap       bool
	createErr     error
	appointment   domain.Appointment
	found         bool
	appointments  []domain.Appointment
	patientActive bool
	listPatientID uuid.UUID
}

func (r *appointmentRepo) Create(ctx context.Context, a domain.Appointment) (domain.Appointment, error) {
	if r.createErr != nil {
		return domain.Appointment{}, r.createErr
	}
	now := time.Date(2027, 4, 11, 12, 0, 0, 0, time.UTC)
	a.CreatedAt = now
	a.UpdatedAt = now
	return a, nil
}

func (r *appointmentRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Appointment, bool, error) {
	return r.appointment, r.found, nil
}

func (r *appointmentRepo) Update(ctx context.Context, a domain.Appointment) (domain.Appointment, error) {
	a.CreatedAt = r.appointment.CreatedAt
	a.UpdatedAt = time.Date(2027, 4, 11, 13, 0, 0, 0, time.UTC)
	r.appointment = a
	return a, nil
}

func (r *appointmentRepo) IsPatientActive(ctx context.Context, patientID uuid.UUID) (bool, error) {
	if !r.patientActive {
		return false, nil
	}
	return true, nil
}

func (r *appointmentRepo) HasOverlap(ctx context.Context, kinesiologistID uuid.UUID, startAt, endAt time.Time, excludeID *uuid.UUID) (bool, error) {
	return r.overlap, nil
}

func (r *appointmentRepo) ListByKinesiologistAndRange(ctx context.Context, kinesiologistID uuid.UUID, startDay, endDay time.Time) ([]domain.Appointment, error) {
	return r.appointments, nil
}

func (r *appointmentRepo) ListByPatientAndRange(ctx context.Context, patientID uuid.UUID, from time.Time, to time.Time) ([]domain.Appointment, error) {
	r.listPatientID = patientID
	return r.appointments, nil
}

type fakeVideoProvider struct{}

func (p fakeVideoProvider) Name() string { return "fake" }

func (p fakeVideoProvider) CreateRoom(ctx context.Context, input videocalls.CreateRoomInput) (videocalls.CreatedRoom, error) {
	meetingID := "fake-room-" + input.AppointmentID
	return videocalls.CreatedRoom{
		Provider:  p.Name(),
		MeetingID: &meetingID,
		URL:       "https://video.example.com/" + input.AppointmentID,
	}, nil
}

type appointmentPatientRepo struct {
	patients []patientDomain.Patient
}

func (r *appointmentPatientRepo) Search(ctx context.Context, query string, limit int, includeInactive bool) ([]patientDomain.Patient, error) {
	return r.patients, nil
}

type appointmentKinesiologistRepo struct {
	kinesiologist kineDomain.Kinesiologist
	found         bool
}

func (r *appointmentKinesiologistRepo) Create(ctx context.Context, k kineDomain.Kinesiologist) (kineDomain.Kinesiologist, error) {
	return k, nil
}

func (r *appointmentKinesiologistRepo) Update(ctx context.Context, k kineDomain.Kinesiologist) (kineDomain.Kinesiologist, error) {
	return k, nil
}

func (r *appointmentKinesiologistRepo) List(ctx context.Context, onlyActive bool) ([]kineDomain.Kinesiologist, error) {
	return []kineDomain.Kinesiologist{r.kinesiologist}, nil
}

func (r *appointmentKinesiologistRepo) FindByEmail(ctx context.Context, email string) (kineDomain.Kinesiologist, bool, error) {
	return r.kinesiologist, r.found, nil
}

func (r *appointmentKinesiologistRepo) GetByID(ctx context.Context, id string) (kineDomain.Kinesiologist, bool, error) {
	return r.kinesiologist, r.found, nil
}

func (r *appointmentKinesiologistRepo) ListSpecialties(ctx context.Context, includeInactive bool) ([]kineDomain.Specialty, error) {
	return nil, nil
}

func (r *appointmentKinesiologistRepo) SaveSpecialty(ctx context.Context, specialty kineDomain.Specialty) (kineDomain.Specialty, error) {
	return specialty, nil
}

func (r *appointmentKinesiologistRepo) ListPractices(ctx context.Context, includeInactive bool) ([]kineDomain.Practice, error) {
	return nil, nil
}

func (r *appointmentKinesiologistRepo) SavePractice(ctx context.Context, practice kineDomain.Practice) (kineDomain.Practice, error) {
	return practice, nil
}

func testKinesiologist(id uuid.UUID) *appointmentKinesiologistRepo {
	return &appointmentKinesiologistRepo{
		kinesiologist: kineDomain.Kinesiologist{
			ID:            id,
			Email:         "kine@example.com",
			WorkStartTime: "08:00",
			WorkEndTime:   "20:00",
			Active:        true,
		},
		found: true,
	}
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
			repo: &appointmentRepo{patientActive: true},
			body: map[string]any{
				"patient_id":       patientID.String(),
				"kinesiologist_id": kinesiologistID.String(),
				"start_at":         "2027-04-11T14:00:00Z",
				"end_at":           "2027-04-11T14:30:00Z",
				"modality":         "virtual",
				"video_call_url":   "https://meet.google.com/abc-defg-hij",
				"notes":            " Primera visita ",
			},
			auth:       "Bearer demo-recepcionista-token",
			wantStatus: http.StatusCreated,
		},
		{
			name: "requires receptionist token",
			repo: &appointmentRepo{patientActive: true},
			body: map[string]any{
				"patient_id":       patientID.String(),
				"kinesiologist_id": kinesiologistID.String(),
				"start_at":         "2027-04-11T14:00:00Z",
				"end_at":           "2027-04-11T14:30:00Z",
			},
			wantStatus: http.StatusUnauthorized,
			wantError:  "unauthorized",
		},
		{
			name: "validates ids and time range",
			repo: &appointmentRepo{patientActive: true},
			body: map[string]any{
				"patient_id":       "bad-patient",
				"kinesiologist_id": "bad-kine",
				"start_at":         "2027-04-11T14:30:00Z",
				"end_at":           "2027-04-11T14:00:00Z",
			},
			auth:       "Bearer demo-recepcionista-token",
			wantStatus: http.StatusBadRequest,
			wantError:  "validation_error",
		},
		{
			name: "validates video call url",
			repo: &appointmentRepo{patientActive: true},
			body: map[string]any{
				"patient_id":       patientID.String(),
				"kinesiologist_id": kinesiologistID.String(),
				"start_at":         "2027-04-11T14:00:00Z",
				"end_at":           "2027-04-11T14:30:00Z",
				"modality":         "virtual",
				"video_call_url":   "meet.google.com/abc-defg-hij",
			},
			auth:       "Bearer demo-recepcionista-token",
			wantStatus: http.StatusBadRequest,
			wantError:  "validation_error",
		},
		{
			name: "rejects overlap",
			repo: &appointmentRepo{overlap: true, patientActive: true},
			body: map[string]any{
				"patient_id":       patientID.String(),
				"kinesiologist_id": kinesiologistID.String(),
				"start_at":         "2027-04-11T14:00:00Z",
				"end_at":           "2027-04-11T14:30:00Z",
			},
			auth:       "Bearer demo-recepcionista-token",
			wantStatus: http.StatusConflict,
			wantError:  "overlap",
		},
		{
			name: "returns patient not found",
			repo: &appointmentRepo{createErr: domain.ErrPatientNotFound, patientActive: true},
			body: map[string]any{
				"patient_id":       patientID.String(),
				"kinesiologist_id": kinesiologistID.String(),
				"start_at":         "2027-04-11T14:00:00Z",
				"end_at":           "2027-04-11T14:30:00Z",
			},
			auth:       "Bearer demo-recepcionista-token",
			wantStatus: http.StatusNotFound,
			wantError:  "patient_not_found",
		},
		{
			name: "returns kinesiologist not found",
			repo: &appointmentRepo{createErr: domain.ErrKinesiologistNotFound, patientActive: true},
			body: map[string]any{
				"patient_id":       patientID.String(),
				"kinesiologist_id": kinesiologistID.String(),
				"start_at":         "2027-04-11T14:00:00Z",
				"end_at":           "2027-04-11T14:30:00Z",
			},
			auth:       "Bearer demo-recepcionista-token",
			wantStatus: http.StatusNotFound,
			wantError:  "kinesiologist_not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createUC := usecase.NewCreateAppointmentUseCase(tt.repo)
			handler := NewHandler(createUC, nil, nil, nil, nil, nil, testKinesiologist(kinesiologistID))

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
			if got["modality"] != string(domain.ModalityVirtual) {
				t.Fatalf("modality = %v, want virtual", got["modality"])
			}
			if got["video_call_url"] != "https://meet.google.com/abc-defg-hij" {
				t.Fatalf("video_call_url = %v, want video call URL", got["video_call_url"])
			}
		})
	}
}

func TestCreateAppointmentEndpointForPatientUsesOwnProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patientID := uuid.New()
	otherPatientID := uuid.New()
	kinesiologistID := uuid.New()
	repo := &appointmentRepo{patientActive: true}
	patients := &appointmentPatientRepo{
		patients: []patientDomain.Patient{
			{
				ID:     patientID,
				Email:  "javi.emiliano@gmail.com",
				Active: true,
			},
		},
	}

	tests := []struct {
		name       string
		body       map[string]any
		wantStatus int
		wantError  string
	}{
		{
			name: "creates appointment without patient id",
			body: map[string]any{
				"kinesiologist_id": kinesiologistID.String(),
				"start_at":         "2027-04-11T14:00:00Z",
				"end_at":           "2027-04-11T14:30:00Z",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "rejects another patient id",
			body: map[string]any{
				"patient_id":       otherPatientID.String(),
				"kinesiologist_id": kinesiologistID.String(),
				"start_at":         "2027-04-11T14:00:00Z",
				"end_at":           "2027-04-11T14:30:00Z",
			},
			wantStatus: http.StatusForbidden,
			wantError:  "forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createUC := usecase.NewCreateAppointmentUseCase(repo)
			handler := NewHandler(createUC, nil, nil, nil, nil, nil, testKinesiologist(kinesiologistID), patients)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set(middleware.AuthUserKey, middleware.AuthUser{
					Email: "javi.emiliano@gmail.com",
					Role:  "paciente",
				})
				c.Next()
			})
			router.POST("/appointments", handler.Create)

			payload, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("marshal body: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/appointments", bytes.NewReader(payload))
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
				t.Fatalf("patient_id = %v, want %s", got["patient_id"], patientID)
			}
		})
	}
}

func TestUpdateAppointmentEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appointmentID := uuid.New()
	patientID := uuid.New()
	kinesiologistID := uuid.New()
	now := time.Date(2027, 4, 11, 12, 0, 0, 0, time.UTC)
	existing := domain.Appointment{
		ID:              appointmentID,
		PatientID:       patientID,
		KinesiologistID: kinesiologistID,
		StartAt:         time.Date(2027, 4, 11, 14, 0, 0, 0, time.UTC),
		EndAt:           time.Date(2027, 4, 11, 14, 30, 0, 0, time.UTC),
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
			repo:   &appointmentRepo{appointment: existing, found: true, patientActive: true},
			method: http.MethodPut,
			body: map[string]any{
				"start_at": "2027-04-11T15:00:00Z",
				"end_at":   "2027-04-11T15:45:00Z",
				"notes":    " Reprogramado ",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "keeps patch compatibility",
			id:     appointmentID.String(),
			repo:   &appointmentRepo{appointment: existing, found: true, patientActive: true},
			method: http.MethodPatch,
			body: map[string]any{
				"notes": "Actualizado",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "validates invalid id",
			id:     "not-a-uuid",
			repo:   &appointmentRepo{appointment: existing, found: true, patientActive: true},
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
			repo:   &appointmentRepo{appointment: existing, found: true, overlap: true, patientActive: true},
			method: http.MethodPut,
			body: map[string]any{
				"start_at": "2027-04-11T15:00:00Z",
				"end_at":   "2027-04-11T15:45:00Z",
			},
			wantStatus: http.StatusConflict,
			wantError:  "overlap",
		},
		{
			name:   "rejects overlap when reactivating cancelled appointment",
			id:     appointmentID.String(),
			repo:   &appointmentRepo{appointment: cancelled, found: true, overlap: true, patientActive: true},
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
			getUC := usecase.NewGetAppointmentByIDUseCase(tt.repo)
			handler := NewHandler(nil, nil, updateUC, nil, getUC, nil, testKinesiologist(kinesiologistID))

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
	now := time.Date(2027, 4, 11, 12, 0, 0, 0, time.UTC)
	existing := domain.Appointment{
		ID:              appointmentID,
		PatientID:       uuid.New(),
		KinesiologistID: uuid.New(),
		StartAt:         time.Date(2027, 4, 11, 14, 0, 0, 0, time.UTC),
		EndAt:           time.Date(2027, 4, 11, 14, 30, 0, 0, time.UTC),
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
			repo: &appointmentRepo{appointment: existing, found: true, patientActive: true},
			body: map[string]any{
				"reason": "Paciente aviso",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "validates invalid id",
			id:         "not-a-uuid",
			repo:       &appointmentRepo{appointment: existing, found: true, patientActive: true},
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

func TestCancelAppointmentEndpointForPatientRequiresOwnAppointment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appointmentID := uuid.New()
	patientID := uuid.New()
	now := time.Date(2027, 4, 11, 12, 0, 0, 0, time.UTC)
	patients := &appointmentPatientRepo{
		patients: []patientDomain.Patient{
			{
				ID:     patientID,
				Email:  "javi.emiliano@gmail.com",
				Active: true,
			},
		},
	}

	tests := []struct {
		name        string
		appointment domain.Appointment
		wantStatus  int
		wantError   string
	}{
		{
			name: "cancels own appointment",
			appointment: domain.Appointment{
				ID:              appointmentID,
				PatientID:       patientID,
				KinesiologistID: uuid.New(),
				StartAt:         time.Date(2027, 4, 11, 14, 0, 0, 0, time.UTC),
				EndAt:           time.Date(2027, 4, 11, 14, 30, 0, 0, time.UTC),
				Status:          domain.StatusScheduled,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "rejects another patient appointment",
			appointment: domain.Appointment{
				ID:              appointmentID,
				PatientID:       uuid.New(),
				KinesiologistID: uuid.New(),
				StartAt:         time.Date(2027, 4, 11, 14, 0, 0, 0, time.UTC),
				EndAt:           time.Date(2027, 4, 11, 14, 30, 0, 0, time.UTC),
				Status:          domain.StatusScheduled,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
			wantStatus: http.StatusForbidden,
			wantError:  "forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &appointmentRepo{appointment: tt.appointment, found: true, patientActive: true}
			getUC := usecase.NewGetAppointmentByIDUseCase(repo)
			cancelUC := usecase.NewCancelAppointmentUseCase(repo)
			handler := NewHandler(nil, nil, nil, cancelUC, getUC, nil, patients)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set(middleware.AuthUserKey, middleware.AuthUser{
					Email: "javi.emiliano@gmail.com",
					Role:  "paciente",
				})
				c.Next()
			})
			router.DELETE("/appointments/:id", handler.Cancel)

			req := httptest.NewRequest(http.MethodDelete, "/appointments/"+appointmentID.String(), bytes.NewReader([]byte(`{}`)))
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
			if got["status"] != string(domain.StatusCancelled) {
				t.Fatalf("status = %v, want cancelled", got["status"])
			}
		})
	}
}

func TestListAppointmentsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appointmentID := uuid.New()
	kinesiologistID := uuid.New()
	now := time.Date(2027, 4, 11, 12, 0, 0, 0, time.UTC)
	repo := &appointmentRepo{
		appointments: []domain.Appointment{
			{
				ID:              appointmentID,
				PatientID:       uuid.New(),
				KinesiologistID: kinesiologistID,
				StartAt:         time.Date(2027, 4, 11, 14, 0, 0, 0, time.UTC),
				EndAt:           time.Date(2027, 4, 11, 14, 30, 0, 0, time.UTC),
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
			target:     "/appointments?kinesiologist_id=" + kinesiologistID.String() + "&date=2027-04-11",
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

func TestListAppointmentsByPatientEndpointUsesLoggedPatientEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appointmentID := uuid.New()
	patientID := uuid.New()
	otherPatientID := uuid.New()
	kinesiologistID := uuid.New()
	now := time.Date(2027, 4, 11, 12, 0, 0, 0, time.UTC)
	repo := &appointmentRepo{
		appointments: []domain.Appointment{
			{
				ID:              appointmentID,
				PatientID:       patientID,
				KinesiologistID: kinesiologistID,
				StartAt:         time.Date(2027, 4, 11, 14, 0, 0, 0, time.UTC),
				EndAt:           time.Date(2027, 4, 11, 14, 30, 0, 0, time.UTC),
				Status:          domain.StatusScheduled,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
		},
	}
	patients := &appointmentPatientRepo{
		patients: []patientDomain.Patient{
			{
				ID:     patientID,
				Email:  "javi.emiliano@gmail.com",
				Active: true,
			},
		},
	}

	listUC := usecase.NewListAppointmentsByPatientUseCase(repo)
	handler := NewHandler(nil, nil, nil, nil, nil, listUC, patients)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.AuthUserKey, middleware.AuthUser{
			Email: "javi.emiliano@gmail.com",
			Role:  "paciente",
		})
		c.Next()
	})
	router.GET("/appointments/patient", handler.ListByPatient)

	req := httptest.NewRequest(
		http.MethodGet,
		"/appointments/patient?patient_id="+otherPatientID.String()+"&from=2027-04-01T00:00:00Z&to=2027-04-30T23:59:59Z",
		nil,
	)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}

	req = httptest.NewRequest(
		http.MethodGet,
		"/appointments/patient?from=2027-04-01T00:00:00Z&to=2027-04-30T23:59:59Z",
		nil,
	)
	rec = httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if repo.listPatientID != patientID {
		t.Fatalf("patient id = %s, want %s", repo.listPatientID, patientID)
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
}

func TestGenerateVideoCallEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appointmentID := uuid.New()
	patientID := uuid.New()
	kinesiologistID := uuid.New()
	now := time.Date(2027, 4, 11, 12, 0, 0, 0, time.UTC)
	repo := &appointmentRepo{
		appointment: domain.Appointment{
			ID:              appointmentID,
			PatientID:       patientID,
			KinesiologistID: kinesiologistID,
			StartAt:         time.Date(2027, 4, 11, 14, 0, 0, 0, time.UTC),
			EndAt:           time.Date(2027, 4, 11, 14, 30, 0, 0, time.UTC),
			Status:          domain.StatusScheduled,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		found: true,
	}
	generateUC := usecase.NewGenerateAppointmentVideoCallUseCase(repo, fakeVideoProvider{})
	handler := NewHandler(nil, nil, nil, nil, nil, nil, generateUC)

	router := gin.New()
	router.POST("/appointments/:id/video-call", handler.GenerateVideoCall)

	req := httptest.NewRequest(http.MethodPost, "/appointments/"+appointmentID.String()+"/video-call", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer demo-recepcionista-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got["modality"] != string(domain.ModalityVirtual) {
		t.Fatalf("modality = %v, want virtual", got["modality"])
	}
	if got["video_call_url"] != "https://video.example.com/"+appointmentID.String() {
		t.Fatalf("video_call_url = %v", got["video_call_url"])
	}
	if got["video_provider"] != "fake" {
		t.Fatalf("video_provider = %v, want fake", got["video_provider"])
	}
}
