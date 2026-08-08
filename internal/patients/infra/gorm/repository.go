package gorm

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
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
		ID:                    p.ID,
		DNI:                   p.DNI,
		FirstName:             p.FirstName,
		LastName:              p.LastName,
		Email:                 p.Email,
		Phone:                 p.Phone,
		BirthDate:             p.BirthDate,
		ClinicalNotes:         p.ClinicalNotes,
		FinancierID:           p.FinancierID,
		FinancierMemberNumber: p.FinancierMemberNumber,
		Active:                p.Active,
	}

	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		if duplicateErr := patientDuplicateError(err); duplicateErr != nil {
			return domain.Patient{}, duplicateErr
		}
		return domain.Patient{}, err
	}

	p.CreatedAt = m.CreatedAt
	p.UpdatedAt = m.UpdatedAt
	return p, nil
}

func (r *Repository) Update(ctx context.Context, p domain.Patient) (domain.Patient, error) {
	updatedAt := time.Now().UTC()
	updates := map[string]any{
		"dni":                             p.DNI,
		"first_name":                      p.FirstName,
		"last_name":                       p.LastName,
		"email":                           p.Email,
		"phone":                           p.Phone,
		"birth_date":                      p.BirthDate,
		"clinical_notes":                  p.ClinicalNotes,
		"clinical_notes_updated_by_email": p.ClinicalNotesUpdatedByEmail,
		"clinical_notes_updated_by_role":  p.ClinicalNotesUpdatedByRole,
		"clinical_notes_updated_at":       p.ClinicalNotesUpdatedAt,
		"financier_id":                    p.FinancierID,
		"financier_member_number":         p.FinancierMemberNumber,
		"active":                          p.Active,
		"updated_at":                      updatedAt,
	}

	tx := r.db.WithContext(ctx).Model(&PatientModel{}).Where("id = ?", p.ID).Updates(updates)
	if tx.Error != nil {
		if duplicateErr := patientDuplicateError(tx.Error); duplicateErr != nil {
			return domain.Patient{}, duplicateErr
		}
		return domain.Patient{}, tx.Error
	}
	if tx.RowsAffected == 0 {
		return domain.Patient{}, domain.ErrNotFound
	}

	return r.GetUpdated(ctx, p.ID.String(), updatedAt)
}

func (r *Repository) GetUpdated(ctx context.Context, id string, fallbackUpdatedAt time.Time) (domain.Patient, error) {
	p, found, err := r.GetByID(ctx, id)
	if err != nil {
		return domain.Patient{}, err
	}
	if !found {
		return domain.Patient{}, domain.ErrNotFound
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = fallbackUpdatedAt
	}
	return p, nil
}

func (r *Repository) SetActive(ctx context.Context, id string, active bool) error {
	tx := r.db.WithContext(ctx).
		Model(&PatientModel{}).
		Where("id = ?", strings.TrimSpace(id)).
		Updates(map[string]any{
			"active":     active,
			"updated_at": time.Now().UTC(),
		})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func patientDuplicateError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return nil
	}

	switch pgErr.ConstraintName {
	case "ux_patients_dni":
		return domain.ErrDuplicateDNI
	case "ux_patients_email":
		return domain.ErrDuplicateEmail
	default:
		return nil
	}
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

func (r *Repository) FindByDNIAndEmail(ctx context.Context, dni, email string) (domain.Patient, bool, error) {
	var m PatientModel
	err := r.db.WithContext(ctx).
		Where("dni = ? AND lower(email) = lower(?)", strings.TrimSpace(dni), strings.TrimSpace(email)).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Patient{}, false, nil
		}
		return domain.Patient{}, false, err
	}
	return m.ToDomain(), true, nil
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
		ID:                          m.ID,
		DNI:                         m.DNI,
		FirstName:                   m.FirstName,
		LastName:                    m.LastName,
		Email:                       m.Email,
		Phone:                       m.Phone,
		BirthDate:                   m.BirthDate,
		ClinicalNotes:               m.ClinicalNotes,
		ClinicalNotesUpdatedByEmail: m.ClinicalNotesUpdatedByEmail,
		ClinicalNotesUpdatedByRole:  m.ClinicalNotesUpdatedByRole,
		ClinicalNotesUpdatedAt:      m.ClinicalNotesUpdatedAt,
		FinancierID:                 m.FinancierID,
		FinancierMemberNumber:       m.FinancierMemberNumber,
		Active:                      m.Active,
		CreatedAt:                   m.CreatedAt,
		UpdatedAt:                   m.UpdatedAt,
	}
	return p, true, nil
}

func (r *Repository) List(ctx context.Context, limit int, offset int, includeInactive bool) ([]domain.Patient, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var models []PatientModel
	tx := r.db.WithContext(ctx).Model(&PatientModel{})
	if !includeInactive {
		tx = tx.Where("active = true")
	}
	if err := tx.Order("last_name asc, first_name asc").
		Limit(limit).
		Offset(offset).
		Find(&models).Error; err != nil {
		return nil, err
	}

	out := make([]domain.Patient, 0, len(models))
	for _, m := range models {
		out = append(out, m.ToDomain())
	}
	return out, nil
}

func (r *Repository) Search(ctx context.Context, query string, limit int, includeInactive bool) ([]domain.Patient, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return []domain.Patient{}, nil
	}

	var models []PatientModel
	tx := r.db.WithContext(ctx).Model(&PatientModel{})
	if !includeInactive {
		tx = tx.Where("active = true")
	}

	if _, err := strconv.Atoi(q); err == nil {
		tx = tx.Where("dni LIKE ?", q+"%")
	} else {
		like := "%" + q + "%"
		tx = tx.Where(`(
			lower(email) LIKE ? OR
			lower(first_name) LIKE ? OR
			lower(last_name) LIKE ?
		)`, like, like, like)
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
		ID:                          m.ID,
		DNI:                         m.DNI,
		FirstName:                   m.FirstName,
		LastName:                    m.LastName,
		Email:                       m.Email,
		Phone:                       m.Phone,
		BirthDate:                   m.BirthDate,
		ClinicalNotes:               m.ClinicalNotes,
		ClinicalNotesUpdatedByEmail: m.ClinicalNotesUpdatedByEmail,
		ClinicalNotesUpdatedByRole:  m.ClinicalNotesUpdatedByRole,
		ClinicalNotesUpdatedAt:      m.ClinicalNotesUpdatedAt,
		FinancierID:                 m.FinancierID,
		FinancierMemberNumber:       m.FinancierMemberNumber,
		Active:                      m.Active,
		CreatedAt:                   m.CreatedAt,
		UpdatedAt:                   m.UpdatedAt,
	}
}
