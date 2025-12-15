package gorm

import (
	"context"
	"errors"
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
