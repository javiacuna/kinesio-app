package gorm

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/javiacuna/kinesio-backend/internal/patients/domain"
	"github.com/javiacuna/kinesio-backend/internal/patients/ports"
	"gorm.io/gorm"
)

var _ ports.Repository = (*Repository)(nil)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, p domain.Patient) (domain.Patient, error) {
	m := PatientModel{
		ID:            p.ID,
		DNI:           p.DNI,
		FirstName:     p.FirstName,
		LastName:      p.LastName,
		Email:         p.Email,
		Phone:         p.Phone,
		BirthDate:     p.BirthDate,
		ClinicalNotes: p.ClinicalNotes,
	}

	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return domain.Patient{}, err
	}

	p.CreatedAt = m.CreatedAt
	p.UpdatedAt = m.UpdatedAt
	return p, nil
}

func (r *Repository) ExistsByDNI(ctx context.Context, dni string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&PatientModel{}).
		Where("dni = ?", strings.TrimSpace(dni)).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&PatientModel{}).
		Where("lower(email) = lower(?)", strings.TrimSpace(email)).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (domain.Patient, bool, error) {
	var m PatientModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Patient{}, false, nil
		}
		return domain.Patient{}, false, err
	}

	p := domain.Patient{
		ID:            m.ID,
		DNI:           m.DNI,
		FirstName:     m.FirstName,
		LastName:      m.LastName,
		Email:         m.Email,
		Phone:         m.Phone,
		BirthDate:     m.BirthDate,
		ClinicalNotes: m.ClinicalNotes,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
	return p, true, nil
}

func (r *Repository) Search(ctx context.Context, query string, limit int) ([]domain.Patient, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return []domain.Patient{}, nil
	}

	var models []PatientModel
	tx := r.db.WithContext(ctx).Model(&PatientModel{})

	if _, err := strconv.Atoi(q); err == nil {
		tx = tx.Where("dni LIKE ?", q+"%")
	} else {
		like := "%" + q + "%"
		tx = tx.Where(`
			lower(email) LIKE ? OR
			lower(first_name) LIKE ? OR
			lower(last_name) LIKE ?
		`, like, like, like)
	}

	if limit <= 0 || limit > 50 {
		limit = 20
	}

	if err := tx.Order("last_name asc, first_name asc").Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}

	out := make([]domain.Patient, 0, len(models))
	for _, m := range models {
		out = append(out, m.ToDomain())
	}
	return out, nil
}

func (m PatientModel) ToDomain() domain.Patient {
	return domain.Patient{
		ID:            m.ID,
		DNI:           m.DNI,
		FirstName:     m.FirstName,
		LastName:      m.LastName,
		Email:         m.Email,
		Phone:         m.Phone,
		BirthDate:     m.BirthDate,
		ClinicalNotes: m.ClinicalNotes,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}
