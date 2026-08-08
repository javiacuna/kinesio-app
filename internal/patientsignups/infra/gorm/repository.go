package gorm

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/javiacuna/kinesio-backend/internal/patientsignups/domain"
	"github.com/javiacuna/kinesio-backend/internal/patientsignups/ports"
)

var _ ports.Repository = (*Repository)(nil)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, req domain.SignupRequest) (domain.SignupRequest, error) {
	m := SignupRequestModel{
		ID:               req.ID,
		FirebaseUID:      req.FirebaseUID,
		DNI:              req.DNI,
		Email:            req.Email,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Status:           string(req.Status),
		MatchedPatientID: req.MatchedPatientID,
		ReviewedByEmail:  req.ReviewedByEmail,
		ReviewedAt:       req.ReviewedAt,
		RejectionReason:  req.RejectionReason,
	}

	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "ux_patient_signup_requests_firebase_uid" {
			return domain.SignupRequest{}, domain.ErrAlreadyPending
		}
		return domain.SignupRequest{}, err
	}

	req.CreatedAt = m.CreatedAt
	req.UpdatedAt = m.UpdatedAt
	return req, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (domain.SignupRequest, bool, error) {
	var m SignupRequestModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", strings.TrimSpace(id)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.SignupRequest{}, false, nil
		}
		return domain.SignupRequest{}, false, err
	}
	return m.ToDomain(), true, nil
}

func (r *Repository) List(ctx context.Context, status string) ([]domain.SignupRequest, error) {
	var models []SignupRequestModel
	tx := r.db.WithContext(ctx).Model(&SignupRequestModel{})
	if status = strings.TrimSpace(status); status != "" {
		tx = tx.Where("status = ?", status)
	}
	if err := tx.Order("created_at desc").Find(&models).Error; err != nil {
		return nil, err
	}

	out := make([]domain.SignupRequest, 0, len(models))
	for _, m := range models {
		out = append(out, m.ToDomain())
	}
	return out, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status domain.Status, matchedPatientID *uuid.UUID, reviewedByEmail *string, reviewedAt time.Time, rejectionReason *string) (domain.SignupRequest, error) {
	updates := map[string]any{
		"status":             string(status),
		"matched_patient_id": matchedPatientID,
		"reviewed_by_email":  reviewedByEmail,
		"reviewed_at":        reviewedAt,
		"rejection_reason":   rejectionReason,
		"updated_at":         time.Now().UTC(),
	}

	tx := r.db.WithContext(ctx).Model(&SignupRequestModel{}).Where("id = ?", strings.TrimSpace(id)).Updates(updates)
	if tx.Error != nil {
		return domain.SignupRequest{}, tx.Error
	}
	if tx.RowsAffected == 0 {
		return domain.SignupRequest{}, domain.ErrRequestNotFound
	}

	updated, found, err := r.GetByID(ctx, id)
	if err != nil {
		return domain.SignupRequest{}, err
	}
	if !found {
		return domain.SignupRequest{}, domain.ErrRequestNotFound
	}
	return updated, nil
}

func (m SignupRequestModel) ToDomain() domain.SignupRequest {
	return domain.SignupRequest{
		ID:               m.ID,
		FirebaseUID:      m.FirebaseUID,
		DNI:              m.DNI,
		Email:            m.Email,
		FirstName:        m.FirstName,
		LastName:         m.LastName,
		Status:           domain.Status(m.Status),
		MatchedPatientID: m.MatchedPatientID,
		ReviewedByEmail:  m.ReviewedByEmail,
		ReviewedAt:       m.ReviewedAt,
		RejectionReason:  m.RejectionReason,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}
