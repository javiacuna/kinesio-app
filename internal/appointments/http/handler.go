package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/appointments/domain"
	"github.com/javiacuna/kinesio-backend/internal/appointments/usecase"
	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	kinePorts "github.com/javiacuna/kinesio-backend/internal/kinesiologists/ports"
	patientDomain "github.com/javiacuna/kinesio-backend/internal/patients/domain"
)

type Handler struct {
	create        *usecase.CreateAppointmentUseCase
	listDay       *usecase.ListAppointmentsDayUseCase
	update        *usecase.UpdateAppointmentUseCase
	cancel        *usecase.CancelAppointmentUseCase
	getByID       *usecase.GetAppointmentByIDUseCase
	listByPatient *usecase.ListAppointmentsByPatientUseCase
	createPackage *usecase.CreateAppointmentPackageUseCase
	updatePackage *usecase.UpdateAppointmentPackageUseCase
	kinesios      kinePorts.Repository
	patients      patientSearcher
}

type patientSearcher interface {
	Search(ctx context.Context, query string, limit int, includeInactive bool) ([]patientDomain.Patient, error)
}

func NewHandler(
	create *usecase.CreateAppointmentUseCase,
	listDay *usecase.ListAppointmentsDayUseCase,
	update *usecase.UpdateAppointmentUseCase,
	cancel *usecase.CancelAppointmentUseCase,
	getByID *usecase.GetAppointmentByIDUseCase,
	listByPatient *usecase.ListAppointmentsByPatientUseCase,
	lookups ...any,
) *Handler {
	var kineRepo kinePorts.Repository
	var patientRepo patientSearcher
	var createPackage *usecase.CreateAppointmentPackageUseCase
	var updatePackage *usecase.UpdateAppointmentPackageUseCase
	for _, lookup := range lookups {
		switch repo := lookup.(type) {
		case kinePorts.Repository:
			kineRepo = repo
		case patientSearcher:
			patientRepo = repo
		case *usecase.CreateAppointmentPackageUseCase:
			createPackage = repo
		case *usecase.UpdateAppointmentPackageUseCase:
			updatePackage = repo
		}
	}

	return &Handler{
		create:        create,
		listDay:       listDay,
		update:        update,
		cancel:        cancel,
		getByID:       getByID,
		listByPatient: listByPatient,
		createPackage: createPackage,
		updatePackage: updatePackage,
		kinesios:      kineRepo,
		patients:      patientRepo,
	}
}

type createReq struct {
	PatientID       string  `json:"patient_id"`
	KinesiologistID string  `json:"kinesiologist_id"`
	StartAt         string  `json:"start_at"` // RFC3339
	EndAt           string  `json:"end_at"`   // RFC3339
	Notes           *string `json:"notes,omitempty"`
}

type updateReq struct {
	StartAt         *string `json:"start_at,omitempty"`
	EndAt           *string `json:"end_at,omitempty"`
	Status          *string `json:"status,omitempty"` // scheduled|cancelled
	CancelledReason *string `json:"cancelled_reason,omitempty"`
	Notes           *string `json:"notes,omitempty"`
}

type cancelReq struct {
	Reason *string `json:"reason,omitempty"`
}

type createPackageReq struct {
	PatientID       string  `json:"patient_id"`
	KinesiologistID string  `json:"kinesiologist_id"`
	StartDate       string  `json:"start_date"`
	StartTime       string  `json:"start_time"`
	DurationMin     int     `json:"duration_min"`
	SessionsCount   int     `json:"sessions_count"`
	WeekdaysOnly    bool    `json:"weekdays_only"`
	Notes           *string `json:"notes,omitempty"`
}

type updatePackageReq struct {
	StartDate   *string `json:"start_date,omitempty"`
	StartTime   *string `json:"start_time,omitempty"`
	DurationMin *int    `json:"duration_min,omitempty"`
	WorkDays    []int   `json:"work_days,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

type resp struct {
	ID                   string  `json:"id"`
	PatientID            string  `json:"patient_id"`
	KinesiologistID      string  `json:"kinesiologist_id"`
	PackageID            *string `json:"package_id,omitempty"`
	PackageSessionNumber *int    `json:"package_session_number,omitempty"`
	StartAt              string  `json:"start_at"`
	EndAt                string  `json:"end_at"`
	Status               string  `json:"status"`
	Notes                *string `json:"notes,omitempty"`
	CancelledReason      *string `json:"cancelled_reason,omitempty"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
}

type appointmentPackageResp struct {
	ID              string  `json:"id"`
	PatientID       string  `json:"patient_id"`
	KinesiologistID string  `json:"kinesiologist_id"`
	SessionsCount   int     `json:"sessions_count"`
	DurationMin     int     `json:"duration_min"`
	StartDate       string  `json:"start_date"`
	StartTime       string  `json:"start_time"`
	WeekdaysOnly    bool    `json:"weekdays_only"`
	WorkDays        []int   `json:"work_days"`
	Notes           *string `json:"notes,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type packageWriteResp struct {
	Package      appointmentPackageResp `json:"package"`
	Appointments []resp                 `json:"appointments"`
}

func (h *Handler) Create(c *gin.Context) {
	isPatient := h.isCurrentPatient(c)
	if !middleware.HasRole(c, "recepcionista") && !isPatient {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	if isPatient {
		patientID, ok := h.patientIDForCurrentPatient(c, req.PatientID)
		if !ok {
			return
		}
		req.PatientID = patientID
	}
	if !h.validateKinesiologistWorkingHours(c, req.KinesiologistID, req.StartAt, req.EndAt) {
		return
	}

	out, details, err := h.create.Execute(c.Request.Context(), usecase.CreateAppointmentInput{
		PatientID:       req.PatientID,
		KinesiologistID: req.KinesiologistID,
		StartAt:         req.StartAt,
		EndAt:           req.EndAt,
		Notes:           req.Notes,
	})

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": details})
		case errors.Is(err, domain.ErrOverlap):
			c.JSON(http.StatusConflict, gin.H{"error": "overlap"})
		case errors.Is(err, domain.ErrPatientNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "patient_not_found"})
		case errors.Is(err, domain.ErrPatientInactive):
			c.JSON(http.StatusConflict, gin.H{"error": "patient_inactive"})
		case errors.Is(err, domain.ErrKinesiologistNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "kinesiologist_not_found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusCreated, toResp(out))
}

func (h *Handler) CreatePackage(c *gin.Context) {
	if !middleware.HasRole(c, "recepcionista") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if h.createPackage == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	var req createPackageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	_, _, workDays, ok := h.packageKinesiologistWorkingHours(c, req.KinesiologistID)
	if !ok {
		return
	}

	slots, _, err := usecase.BuildPackageSlots(
		req.StartDate,
		req.StartTime,
		req.DurationMin,
		req.SessionsCount,
		req.WeekdaysOnly,
		workDays,
	)
	if err == nil && !h.validatePackageWorkingHours(c, req.KinesiologistID, slots) {
		return
	}

	pkg, appointments, details, err := h.createPackage.Execute(c.Request.Context(), usecase.CreateAppointmentPackageInput{
		PatientID:       req.PatientID,
		KinesiologistID: req.KinesiologistID,
		StartDate:       req.StartDate,
		StartTime:       req.StartTime,
		DurationMin:     req.DurationMin,
		SessionsCount:   req.SessionsCount,
		WeekdaysOnly:    req.WeekdaysOnly,
		WorkDays:        workDays,
		Notes:           req.Notes,
	})
	if err != nil {
		h.writeAppointmentPackageError(c, details, err)
		return
	}

	c.JSON(http.StatusCreated, toPackageWriteResp(pkg, appointments))
}

func (h *Handler) UpdatePackage(c *gin.Context) {
	if !middleware.HasRole(c, "recepcionista") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if h.updatePackage == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	var req updatePackageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	pkg, found, details, err := h.updatePackage.GetPackage(c.Request.Context(), c.Param("package_id"))
	if err != nil {
		h.writeAppointmentPackageError(c, details, err)
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}

	workStartTime, workEndTime, workDays, ok := h.packageKinesiologistWorkingHours(c, pkg.KinesiologistID.String())
	if !ok {
		return
	}

	updatedPackage, appointments, details, err := h.updatePackage.Execute(c.Request.Context(), c.Param("package_id"), usecase.UpdateAppointmentPackageInput{
		StartTime:     req.StartTime,
		StartDate:     req.StartDate,
		DurationMin:   req.DurationMin,
		WorkDays:      req.WorkDays,
		Notes:         req.Notes,
		WorkStartTime: workStartTime,
		WorkEndTime:   workEndTime,
		AllowedDays:   workDays,
	})
	if err != nil {
		h.writeAppointmentPackageError(c, details, err)
		return
	}

	c.JSON(http.StatusOK, toPackageWriteResp(updatedPackage, appointments))
}

func (h *Handler) ListDay(c *gin.Context) {
	// Para agenda normalmente también debería estar autenticado; lo dejamos abierto si querés.
	kid := c.Query("kinesiologist_id")
	date := c.Query("date")

	if user, ok := middleware.CurrentUser(c); ok && strings.EqualFold(user.Role, "kinesiologo") {
		if h.kinesios == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}

		kinesiologist, found, err := h.kinesios.FindByEmail(c.Request.Context(), user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
		if !found || !kinesiologist.Active {
			c.JSON(http.StatusForbidden, gin.H{"error": "kinesiologist_profile_not_found"})
			return
		}

		ownID := kinesiologist.ID.String()
		if strings.TrimSpace(kid) != "" && strings.TrimSpace(kid) != ownID {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		kid = ownID
	}

	items, details, err := h.listDay.Execute(c.Request.Context(), kid, date)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": details})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	resp := make([]resp, 0, len(items))
	for _, it := range items {
		resp = append(resp, toResp(it))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Update(c *gin.Context) {
	if !middleware.HasRole(c, "recepcionista") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req updateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	id := c.Param("id")
	if (req.StartAt != nil || req.EndAt != nil) && !h.validateUpdatedAppointmentWorkingHours(c, id, req.StartAt, req.EndAt) {
		return
	}
	out, details, err := h.update.Execute(c.Request.Context(), id, usecase.UpdateAppointmentInput{
		StartAt:         req.StartAt,
		EndAt:           req.EndAt,
		Status:          req.Status,
		CancelledReason: req.CancelledReason,
		Notes:           req.Notes,
	})

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": details})
		case errors.Is(err, domain.ErrOverlap):
			c.JSON(http.StatusConflict, gin.H{"error": "overlap"})
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		case errors.Is(err, domain.ErrPatientNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "patient_not_found"})
		case errors.Is(err, domain.ErrPatientInactive):
			c.JSON(http.StatusConflict, gin.H{"error": "patient_inactive"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusOK, toResp(out))
}

func (h *Handler) Cancel(c *gin.Context) {
	isPatient := h.isCurrentPatient(c)
	if !middleware.HasRole(c, "recepcionista") && !isPatient {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req cancelReq
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
			return
		}
	}

	if isPatient {
		if !h.canCurrentPatientAccessAppointment(c, c.Param("id")) {
			return
		}
	}

	out, details, err := h.cancel.Execute(c.Request.Context(), c.Param("id"), usecase.CancelAppointmentInput{
		Reason: req.Reason,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": details})
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusOK, toResp(out))
}

func toResp(a domain.Appointment) resp {
	var packageID *string
	if a.PackageID != nil {
		value := a.PackageID.String()
		packageID = &value
	}

	return resp{
		ID:                   a.ID.String(),
		PatientID:            a.PatientID.String(),
		KinesiologistID:      a.KinesiologistID.String(),
		PackageID:            packageID,
		PackageSessionNumber: a.PackageSessionNumber,
		StartAt:              a.StartAt.UTC().Format(timeRFC3339()),
		EndAt:                a.EndAt.UTC().Format(timeRFC3339()),
		Status:               string(a.Status),
		Notes:                a.Notes,
		CancelledReason:      a.CancelledReason,
		CreatedAt:            a.CreatedAt.UTC().Format(timeRFC3339()),
		UpdatedAt:            a.UpdatedAt.UTC().Format(timeRFC3339()),
	}
}

func timeRFC3339() string { return "2006-01-02T15:04:05Z07:00" }

func toPackageWriteResp(pkg domain.AppointmentPackage, appointments []domain.Appointment) packageWriteResp {
	out := make([]resp, 0, len(appointments))
	for _, appointment := range appointments {
		out = append(out, toResp(appointment))
	}
	return packageWriteResp{
		Package:      toPackageResp(pkg),
		Appointments: out,
	}
}

func toPackageResp(pkg domain.AppointmentPackage) appointmentPackageResp {
	return appointmentPackageResp{
		ID:              pkg.ID.String(),
		PatientID:       pkg.PatientID.String(),
		KinesiologistID: pkg.KinesiologistID.String(),
		SessionsCount:   pkg.SessionsCount,
		DurationMin:     pkg.DurationMin,
		StartDate:       pkg.StartDate.In(clinicLocation()).Format("2006-01-02"),
		StartTime:       pkg.StartTime,
		WeekdaysOnly:    pkg.WeekdaysOnly,
		WorkDays:        pkg.WorkDays,
		Notes:           pkg.Notes,
		CreatedAt:       pkg.CreatedAt.UTC().Format(timeRFC3339()),
		UpdatedAt:       pkg.UpdatedAt.UTC().Format(timeRFC3339()),
	}
}

func (h *Handler) writeAppointmentPackageError(c *gin.Context, details map[string]string, err error) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": details})
	case errors.Is(err, domain.ErrOverlap):
		if details != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "overlap", "details": details})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": "overlap"})
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
	case errors.Is(err, domain.ErrPatientNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "patient_not_found"})
	case errors.Is(err, domain.ErrPatientInactive):
		c.JSON(http.StatusConflict, gin.H{"error": "patient_inactive"})
	case errors.Is(err, domain.ErrKinesiologistNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "kinesiologist_not_found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")

	out, found, err := h.getByID.Execute(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}

	c.JSON(http.StatusOK, toResp(out))
}

func (h *Handler) ListByPatient(c *gin.Context) {
	patientID, ok := h.patientIDForList(c)
	if !ok {
		return
	}
	from := c.Query("from")
	to := c.Query("to")

	items, details, err := h.listByPatient.Execute(
		c.Request.Context(),
		patientID,
		from,
		to,
	)

	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_error",
				"details": details,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	resp := make([]resp, 0, len(items))
	for _, it := range items {
		resp = append(resp, toResp(it))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) patientIDForList(c *gin.Context) (string, bool) {
	patientID := strings.TrimSpace(c.Query("patient_id"))
	if !h.isCurrentPatient(c) {
		return patientID, true
	}

	return h.patientIDForCurrentPatient(c, patientID)
}

func (h *Handler) patientIDForCurrentPatient(c *gin.Context, requestedPatientID string) (string, bool) {
	user, _ := middleware.CurrentUser(c)
	if h.patients == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return "", false
	}

	email := strings.TrimSpace(user.Email)
	if email == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "patient_profile_not_found"})
		return "", false
	}

	patients, err := h.patients.Search(c.Request.Context(), email, 10, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return "", false
	}

	for _, patient := range patients {
		if strings.EqualFold(patient.Email, email) && patient.Active {
			ownID := patient.ID.String()
			if strings.TrimSpace(requestedPatientID) != "" && strings.TrimSpace(requestedPatientID) != ownID {
				c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
				return "", false
			}
			return ownID, true
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "patient_profile_not_found"})
	return "", false
}

func (h *Handler) canCurrentPatientAccessAppointment(c *gin.Context, appointmentID string) bool {
	if h.getByID == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return false
	}

	appointment, found, err := h.getByID.Execute(c.Request.Context(), appointmentID)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
			return false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return false
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return false
	}

	patientID, ok := h.patientIDForCurrentPatient(c, "")
	if !ok {
		return false
	}
	if appointment.PatientID.String() != patientID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return false
	}

	return true
}

func (h *Handler) isCurrentPatient(c *gin.Context) bool {
	user, ok := middleware.CurrentUser(c)
	return ok && strings.EqualFold(user.Role, "paciente")
}

func (h *Handler) validateUpdatedAppointmentWorkingHours(c *gin.Context, appointmentID string, startAt *string, endAt *string) bool {
	if h.getByID == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return false
	}

	current, found, err := h.getByID.Execute(c.Request.Context(), appointmentID)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
			return false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return false
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return false
	}

	start := current.StartAt.Format(time.RFC3339)
	end := current.EndAt.Format(time.RFC3339)
	if startAt != nil {
		start = *startAt
	}
	if endAt != nil {
		end = *endAt
	}
	return h.validateKinesiologistWorkingHours(c, current.KinesiologistID.String(), start, end)
}

func (h *Handler) validateKinesiologistWorkingHours(c *gin.Context, kinesiologistID string, startAt string, endAt string) bool {
	if h.kinesios == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return false
	}

	kinesiologistID = strings.TrimSpace(kinesiologistID)
	if _, err := uuid.Parse(kinesiologistID); err != nil {
		return true
	}

	kinesiologist, found, err := h.kinesios.GetByID(c.Request.Context(), kinesiologistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return false
	}
	if !found || !kinesiologist.Active {
		c.JSON(http.StatusNotFound, gin.H{"error": "kinesiologist_not_found"})
		return false
	}

	start, err := time.Parse(time.RFC3339, strings.TrimSpace(startAt))
	if err != nil {
		return true
	}
	end, err := time.Parse(time.RFC3339, strings.TrimSpace(endAt))
	if err != nil {
		return true
	}

	if !appointmentInsideWorkingHours(start, end, kinesiologist.WorkStartTime, kinesiologist.WorkEndTime, kinesiologist.WorkDays) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "validation_error",
			"details": gin.H{
				"start_at": "Fuera del horario laboral o días laborales del kinesiólogo",
			},
		})
		return false
	}

	return true
}

func (h *Handler) validatePackageWorkingHours(c *gin.Context, kinesiologistID string, slots []usecase.PackageSlot) bool {
	if _, err := uuid.Parse(strings.TrimSpace(kinesiologistID)); err != nil {
		return true
	}

	workStartTime, workEndTime, workDays, ok := h.packageKinesiologistWorkingHours(c, kinesiologistID)
	if !ok {
		return false
	}

	for _, slot := range slots {
		if !appointmentInsideWorkingHours(slot.StartAt, slot.EndAt, workStartTime, workEndTime, workDays) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "validation_error",
				"details": gin.H{
					"start_time": "Fuera del horario laboral o días laborales del kinesiólogo",
					"session":    slot.SessionNumber,
				},
			})
			return false
		}
	}
	return true
}

func (h *Handler) packageKinesiologistWorkingHours(c *gin.Context, kinesiologistID string) (string, string, []int, bool) {
	if h.kinesios == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return "", "", nil, false
	}

	kinesiologistID = strings.TrimSpace(kinesiologistID)
	if _, err := uuid.Parse(kinesiologistID); err != nil {
		return "", "", nil, true
	}

	kinesiologist, found, err := h.kinesios.GetByID(c.Request.Context(), kinesiologistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return "", "", nil, false
	}
	if !found || !kinesiologist.Active {
		c.JSON(http.StatusNotFound, gin.H{"error": "kinesiologist_not_found"})
		return "", "", nil, false
	}

	return kinesiologist.WorkStartTime, kinesiologist.WorkEndTime, kinesiologist.WorkDays, true
}

func appointmentInsideWorkingHours(startAt, endAt time.Time, workStartTime, workEndTime string, workDays []int) bool {
	loc := clinicLocation()
	localStart := startAt.In(loc)
	localEnd := endAt.In(loc)
	if localStart.Year() != localEnd.Year() || localStart.YearDay() != localEnd.YearDay() {
		return false
	}
	if len(workDays) > 0 && !workDayAllowed(localStart.Weekday(), workDays) {
		return false
	}

	startMin := localStart.Hour()*60 + localStart.Minute()
	endMin := localEnd.Hour()*60 + localEnd.Minute()
	workStartMin, ok := hhmmMinutes(workStartTime)
	if !ok {
		workStartMin = 8 * 60
	}
	workEndMin, ok := hhmmMinutes(workEndTime)
	if !ok {
		workEndMin = 20 * 60
	}

	return startMin >= workStartMin && endMin <= workEndMin
}

func workDayAllowed(weekday time.Weekday, workDays []int) bool {
	day := int(weekday)
	if day == 0 {
		day = 7
	}
	for _, allowed := range workDays {
		if allowed == day {
			return true
		}
	}
	return false
}

func clinicLocation() *time.Location {
	loc, err := time.LoadLocation("America/Argentina/Cordoba")
	if err == nil {
		return loc
	}
	return time.FixedZone("ART", -3*60*60)
}

func hhmmMinutes(value string) (int, bool) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) != 2 {
		return 0, false
	}
	parsed, err := time.Parse("15:04", strings.TrimSpace(value))
	if err != nil {
		return 0, false
	}
	return parsed.Hour()*60 + parsed.Minute(), true
}
