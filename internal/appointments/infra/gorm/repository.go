package gorm

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
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
		ID:              a.ID,
		PatientID:       a.PatientID,
		KinesiologistID: a.KinesiologistID,
		StartAt:         a.StartAt,
		EndAt:           a.EndAt,
		Status:          string(a.Status),
		Notes:           a.Notes,
		CancelledReason: a.CancelledReason,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return domain.Appointment{}, err
	}
	a.CreatedAt = m.CreatedAt
	a.UpdatedAt = m.UpdatedAt
	return a, nil
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
		"start_at":         a.StartAt,
		"end_at":           a.EndAt,
		"status":           string(a.Status),
		"notes":            a.Notes,
		"cancelled_reason": a.CancelledReason,
		"updated_at":       time.Now().UTC(),
	}
	if err := r.db.WithContext(ctx).Model(&AppointmentModel{}).Where("id = ?", a.ID).Updates(updates).Error; err != nil {
		return domain.Appointment{}, err
	}
	// Volver a leer para timestamps consistentes
	return r.read(ctx, a.ID)
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
		ID:              m.ID,
		PatientID:       m.PatientID,
		KinesiologistID: m.KinesiologistID,
		StartAt:         m.StartAt.UTC(),
		EndAt:           m.EndAt.UTC(),
		Status:          domain.Status(m.Status),
		Notes:           m.Notes,
		CancelledReason: m.CancelledReason,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
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
