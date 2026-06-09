package gorm

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/javiacuna/kinesio-backend/internal/appointments/domain"
	"github.com/javiacuna/kinesio-backend/internal/appointments/ports"
	"gorm.io/gorm"
)

var _ ports.Repository = (*Repository)(nil)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, a domain.Appointment) (domain.Appointment, error) {
	m := AppointmentModel{
		ID:                   a.ID,
		PatientID:            a.PatientID,
		KinesiologistID:      a.KinesiologistID,
		PracticeID:           a.PracticeID,
		FinancierID:          a.FinancierID,
		PackageID:            a.PackageID,
		PackageSessionNumber: a.PackageSessionNumber,
		StartAt:              a.StartAt,
		EndAt:                a.EndAt,
		Status:               string(a.Status),
		Modality:             appointmentModalityValue(a.Modality),
		VideoCallURL:         a.VideoCallURL,
		VideoProvider:        a.VideoProvider,
		VideoMeetingID:       a.VideoMeetingID,
		Notes:                a.Notes,
		CancelledReason:      a.CancelledReason,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		if relationErr := appointmentRelationError(err); relationErr != nil {
			return domain.Appointment{}, relationErr
		}
		return domain.Appointment{}, err
	}
	a.CreatedAt = m.CreatedAt
	a.UpdatedAt = m.UpdatedAt
	return a, nil
}

func appointmentRelationError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		return nil
	}

	switch pgErr.ConstraintName {
	case "fk_appointments_patient", "appointments_patient_id_fkey", "appointment_packages_patient_id_fkey":
		return domain.ErrPatientNotFound
	case "fk_appointments_kinesiologist", "appointments_kinesiologist_id_fkey", "appointment_packages_kinesiologist_id_fkey":
		return domain.ErrKinesiologistNotFound
	default:
		return nil
	}
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (domain.Appointment, bool, error) {
	var m AppointmentModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Appointment{}, false, nil
		}
		return domain.Appointment{}, false, err
	}
	return toDomain(m), true, nil
}

func (r *Repository) Update(ctx context.Context, a domain.Appointment) (domain.Appointment, error) {
	// Actualizamos por ID
	updates := map[string]any{
		"package_id":             a.PackageID,
		"package_session_number": a.PackageSessionNumber,
		"practice_id":            a.PracticeID,
		"financier_id":           a.FinancierID,
		"start_at":               a.StartAt,
		"end_at":                 a.EndAt,
		"status":                 string(a.Status),
		"modality":               appointmentModalityValue(a.Modality),
		"video_call_url":         a.VideoCallURL,
		"video_provider":         a.VideoProvider,
		"video_meeting_id":       a.VideoMeetingID,
		"notes":                  a.Notes,
		"cancelled_reason":       a.CancelledReason,
		"updated_at":             time.Now().UTC(),
	}
	if err := r.db.WithContext(ctx).Model(&AppointmentModel{}).Where("id = ?", a.ID).Updates(updates).Error; err != nil {
		return domain.Appointment{}, err
	}
	// Volver a leer para timestamps consistentes
	return r.read(ctx, a.ID)
}

func (r *Repository) IsPatientActive(ctx context.Context, patientID uuid.UUID) (bool, error) {
	var row struct {
		Active bool
	}
	err := r.db.WithContext(ctx).
		Table("patients").
		Select("active").
		Where("id = ?", patientID).
		First(&row).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, domain.ErrPatientNotFound
		}
		return false, err
	}

	return row.Active, nil
}

func (r *Repository) read(ctx context.Context, id uuid.UUID) (domain.Appointment, error) {
	var m AppointmentModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return domain.Appointment{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) HasOverlap(ctx context.Context, kinesiologistID uuid.UUID, startAt, endAt time.Time, excludeID *uuid.UUID) (bool, error) {
	// Solapamiento: start < existing_end AND end > existing_start (solo scheduled)
	q := r.db.WithContext(ctx).Model(&AppointmentModel{}).
		Where("kinesiologist_id = ?", kinesiologistID).
		Where("status = ?", string(domain.StatusScheduled)).
		Where("? < end_at AND ? > start_at", startAt, endAt)

	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) HasOverlapIgnoringAppointments(
	ctx context.Context,
	kinesiologistID uuid.UUID,
	startAt time.Time,
	endAt time.Time,
	excludeIDs []uuid.UUID,
) (bool, error) {
	q := r.db.WithContext(ctx).Model(&AppointmentModel{}).
		Where("kinesiologist_id = ?", kinesiologistID).
		Where("status = ?", string(domain.StatusScheduled)).
		Where("? < end_at AND ? > start_at", startAt, endAt)

	if len(excludeIDs) > 0 {
		q = q.Where("id NOT IN ?", excludeIDs)
	}

	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) ListByKinesiologistAndRange(ctx context.Context, kinesiologistID uuid.UUID, startDay, endDay time.Time) ([]domain.Appointment, error) {
	var ms []AppointmentModel
	err := r.db.WithContext(ctx).
		Where("kinesiologist_id = ?", kinesiologistID).
		Where("start_at >= ? AND start_at < ?", startDay, endDay).
		Order("start_at ASC").
		Find(&ms).Error
	if err != nil {
		return nil, err
	}

	out := make([]domain.Appointment, 0, len(ms))
	for _, m := range ms {
		out = append(out, toDomain(m))
	}
	return out, nil
}

func toDomain(m AppointmentModel) domain.Appointment {
	return domain.Appointment{
		ID:                   m.ID,
		PatientID:            m.PatientID,
		KinesiologistID:      m.KinesiologistID,
		PracticeID:           m.PracticeID,
		FinancierID:          m.FinancierID,
		PackageID:            m.PackageID,
		PackageSessionNumber: m.PackageSessionNumber,
		StartAt:              m.StartAt.UTC(),
		EndAt:                m.EndAt.UTC(),
		Status:               domain.Status(m.Status),
		Modality:             appointmentModalityFromDB(m.Modality),
		VideoCallURL:         m.VideoCallURL,
		VideoProvider:        m.VideoProvider,
		VideoMeetingID:       m.VideoMeetingID,
		Notes:                m.Notes,
		CancelledReason:      m.CancelledReason,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
	}
}

func (r *Repository) CreatePackageWithAppointments(
	ctx context.Context,
	pkg domain.AppointmentPackage,
	appointments []domain.Appointment,
) (domain.AppointmentPackage, []domain.Appointment, error) {
	pkgModel := toPackageModel(pkg)
	appointmentModels := make([]AppointmentModel, 0, len(appointments))
	for _, appointment := range appointments {
		appointmentModels = append(appointmentModels, AppointmentModel{
			ID:                   appointment.ID,
			PatientID:            appointment.PatientID,
			KinesiologistID:      appointment.KinesiologistID,
			PracticeID:           appointment.PracticeID,
			FinancierID:          appointment.FinancierID,
			PackageID:            appointment.PackageID,
			PackageSessionNumber: appointment.PackageSessionNumber,
			StartAt:              appointment.StartAt,
			EndAt:                appointment.EndAt,
			Status:               string(appointment.Status),
			Modality:             appointmentModalityValue(appointment.Modality),
			VideoCallURL:         appointment.VideoCallURL,
			VideoProvider:        appointment.VideoProvider,
			VideoMeetingID:       appointment.VideoMeetingID,
			Notes:                appointment.Notes,
			CancelledReason:      appointment.CancelledReason,
		})
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&pkgModel).Error; err != nil {
			if relationErr := appointmentRelationError(err); relationErr != nil {
				return relationErr
			}
			return err
		}
		if len(appointmentModels) > 0 {
			if err := tx.Create(&appointmentModels).Error; err != nil {
				if relationErr := appointmentRelationError(err); relationErr != nil {
					return relationErr
				}
				return err
			}
		}
		return nil
	})
	if err != nil {
		return domain.AppointmentPackage{}, nil, err
	}

	created := toPackageDomain(pkgModel)
	out := make([]domain.Appointment, 0, len(appointmentModels))
	for _, model := range appointmentModels {
		out = append(out, toDomain(model))
	}
	return created, out, nil
}

func appointmentModalityValue(modality domain.Modality) string {
	if modality == "" {
		return string(domain.ModalityInPerson)
	}
	return string(modality)
}

func appointmentModalityFromDB(modality string) domain.Modality {
	if strings.TrimSpace(modality) == "" {
		return domain.ModalityInPerson
	}
	return domain.Modality(modality)
}

func (r *Repository) GetPackageByID(ctx context.Context, id uuid.UUID) (domain.AppointmentPackage, bool, error) {
	var m AppointmentPackageModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.AppointmentPackage{}, false, nil
		}
		return domain.AppointmentPackage{}, false, err
	}
	return toPackageDomain(m), true, nil
}

func (r *Repository) ListByPackage(ctx context.Context, packageID uuid.UUID) ([]domain.Appointment, error) {
	var ms []AppointmentModel
	err := r.db.WithContext(ctx).
		Where("package_id = ?", packageID).
		Order("package_session_number ASC, start_at ASC").
		Find(&ms).Error
	if err != nil {
		return nil, err
	}

	out := make([]domain.Appointment, 0, len(ms))
	for _, m := range ms {
		out = append(out, toDomain(m))
	}
	return out, nil
}

func (r *Repository) UpdatePackage(ctx context.Context, pkg domain.AppointmentPackage) (domain.AppointmentPackage, error) {
	updates := map[string]any{
		"duration_min": pkg.DurationMin,
		"practice_id":  pkg.PracticeID,
		"financier_id": pkg.FinancierID,
		"start_date":   pkg.StartDate,
		"start_time":   pkg.StartTime,
		"work_days":    encodeWorkDays(pkg.WorkDays),
		"notes":        pkg.Notes,
		"updated_at":   time.Now().UTC(),
	}
	if err := r.db.WithContext(ctx).Model(&AppointmentPackageModel{}).Where("id = ?", pkg.ID).Updates(updates).Error; err != nil {
		return domain.AppointmentPackage{}, err
	}

	var model AppointmentPackageModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", pkg.ID).Error; err != nil {
		return domain.AppointmentPackage{}, err
	}
	return toPackageDomain(model), nil
}

func toPackageModel(pkg domain.AppointmentPackage) AppointmentPackageModel {
	return AppointmentPackageModel{
		ID:              pkg.ID,
		PatientID:       pkg.PatientID,
		KinesiologistID: pkg.KinesiologistID,
		PracticeID:      pkg.PracticeID,
		FinancierID:     pkg.FinancierID,
		SessionsCount:   pkg.SessionsCount,
		DurationMin:     pkg.DurationMin,
		StartDate:       pkg.StartDate,
		StartTime:       pkg.StartTime,
		WeekdaysOnly:    pkg.WeekdaysOnly,
		WorkDays:        encodeWorkDays(pkg.WorkDays),
		Notes:           pkg.Notes,
		CreatedAt:       pkg.CreatedAt,
		UpdatedAt:       pkg.UpdatedAt,
	}
}

func toPackageDomain(m AppointmentPackageModel) domain.AppointmentPackage {
	return domain.AppointmentPackage{
		ID:              m.ID,
		PatientID:       m.PatientID,
		KinesiologistID: m.KinesiologistID,
		PracticeID:      m.PracticeID,
		FinancierID:     m.FinancierID,
		SessionsCount:   m.SessionsCount,
		DurationMin:     m.DurationMin,
		StartDate:       m.StartDate,
		StartTime:       m.StartTime,
		WeekdaysOnly:    m.WeekdaysOnly,
		WorkDays:        decodeWorkDays(m.WorkDays),
		Notes:           m.Notes,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func encodeWorkDays(days []int) *string {
	if len(days) == 0 {
		return nil
	}
	seen := map[int]struct{}{}
	parts := make([]string, 0, len(days))
	for _, day := range days {
		if day < 1 || day > 7 {
			continue
		}
		if _, ok := seen[day]; ok {
			continue
		}
		seen[day] = struct{}{}
		parts = append(parts, strconv.Itoa(day))
	}
	if len(parts) == 0 {
		return nil
	}
	value := strings.Join(parts, ",")
	return &value
}

func decodeWorkDays(value *string) []int {
	if value == nil {
		return nil
	}
	parts := strings.Split(strings.TrimSpace(*value), ",")
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		day, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || day < 1 || day > 7 {
			continue
		}
		out = append(out, day)
	}
	return out
}

func (r *Repository) ListByPatientAndRange(
	ctx context.Context,
	patientID uuid.UUID,
	from time.Time,
	to time.Time,
) ([]domain.Appointment, error) {

	var ms []AppointmentModel
	err := r.db.WithContext(ctx).
		Where("patient_id = ?", patientID).
		Where("start_at >= ? AND start_at <= ?", from, to).
		Order("start_at ASC").
		Find(&ms).Error
	if err != nil {
		return nil, err
	}

	out := make([]domain.Appointment, 0, len(ms))
	for _, m := range ms {
		out = append(out, toDomain(m))
	}
	return out, nil
}
