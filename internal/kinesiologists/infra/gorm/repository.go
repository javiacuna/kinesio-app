package gorm

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/javiacuna/kinesio-backend/internal/kinesiologists/domain"
	"github.com/javiacuna/kinesio-backend/internal/kinesiologists/ports"
	"gorm.io/gorm"
)

var _ ports.Repository = (*Repository)(nil)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, k domain.Kinesiologist) (domain.Kinesiologist, error) {
	m := KinesiologistModel{
		ID:            k.ID,
		FirstName:     k.FirstName,
		LastName:      k.LastName,
		Email:         k.Email,
		LicenseNumber: k.LicenseNumber,
		WorkStartTime: k.WorkStartTime,
		WorkEndTime:   k.WorkEndTime,
		Active:        k.Active,
	}

	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		if duplicateErr := duplicateError(err); duplicateErr != nil {
			return domain.Kinesiologist{}, duplicateErr
		}
		return domain.Kinesiologist{}, err
	}

	return toDomain(m), nil
}

func (r *Repository) Update(ctx context.Context, k domain.Kinesiologist) (domain.Kinesiologist, error) {
	updatedAt := time.Now().UTC()
	updates := map[string]any{
		"first_name":      k.FirstName,
		"last_name":       k.LastName,
		"email":           k.Email,
		"license_number":  k.LicenseNumber,
		"work_start_time": k.WorkStartTime,
		"work_end_time":   k.WorkEndTime,
		"active":          k.Active,
		"updated_at":      updatedAt,
	}

	tx := r.db.WithContext(ctx).Model(&KinesiologistModel{}).Where("id = ?", k.ID).Updates(updates)
	if tx.Error != nil {
		if duplicateErr := duplicateError(tx.Error); duplicateErr != nil {
			return domain.Kinesiologist{}, duplicateErr
		}
		return domain.Kinesiologist{}, tx.Error
	}
	if tx.RowsAffected == 0 {
		return domain.Kinesiologist{}, domain.ErrNotFound
	}

	k.UpdatedAt = updatedAt
	return k, nil
}

func (r *Repository) List(ctx context.Context, onlyActive bool) ([]domain.Kinesiologist, error) {
	q := r.db.WithContext(ctx).Model(&KinesiologistModel{})
	if onlyActive {
		q = q.Where("active = true")
	}

	var ms []KinesiologistModel
	if err := q.Order("last_name ASC, first_name ASC").Find(&ms).Error; err != nil {
		return nil, err
	}

	out := make([]domain.Kinesiologist, 0, len(ms))
	for _, m := range ms {
		out = append(out, toDomain(m))
	}
	return out, nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (domain.Kinesiologist, bool, error) {
	var m KinesiologistModel
	err := r.db.WithContext(ctx).
		Where("lower(email) = ?", domain.NormalizeEmail(email)).
		First(&m).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Kinesiologist{}, false, nil
		}
		return domain.Kinesiologist{}, false, err
	}

	return toDomain(m), true, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (domain.Kinesiologist, bool, error) {
	var m KinesiologistModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Kinesiologist{}, false, nil
		}
		return domain.Kinesiologist{}, false, err
	}

	return toDomain(m), true, nil
}

func duplicateError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return nil
	}

	if pgErr.ConstraintName == "ux_kinesiologists_email" {
		return domain.ErrDuplicateEmail
	}
	return nil
}

func toDomain(m KinesiologistModel) domain.Kinesiologist {
	return domain.Kinesiologist{
		ID:            m.ID,
		FirstName:     m.FirstName,
		LastName:      m.LastName,
		Email:         m.Email,
		LicenseNumber: m.LicenseNumber,
		WorkStartTime: m.WorkStartTime,
		WorkEndTime:   m.WorkEndTime,
		Active:        m.Active,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}
