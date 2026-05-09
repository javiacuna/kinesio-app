package gorm

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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
		WorkDays:      encodeWorkDays(k.WorkDays),
		Active:        k.Active,
	}

	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&m).Error; err != nil {
			return err
		}
		return replaceKinesiologistPractices(tx, k.ID, k.PracticeIDs)
	}); err != nil {
		if duplicateErr := duplicateError(err); duplicateErr != nil {
			return domain.Kinesiologist{}, duplicateErr
		}
		return domain.Kinesiologist{}, err
	}

	out := toDomain(m)
	out.Practices = r.practicesForKinesiologist(ctx, out.ID)
	out.PracticeIDs = practiceIDs(out.Practices)
	return out, nil
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
		"work_days":       encodeWorkDays(k.WorkDays),
		"active":          k.Active,
		"updated_at":      updatedAt,
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&KinesiologistModel{}).Where("id = ?", k.ID).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		return replaceKinesiologistPractices(tx, k.ID, k.PracticeIDs)
	})
	if err != nil {
		if duplicateErr := duplicateError(err); duplicateErr != nil {
			return domain.Kinesiologist{}, duplicateErr
		}
		return domain.Kinesiologist{}, err
	}

	k.UpdatedAt = updatedAt
	k.Practices = r.practicesForKinesiologist(ctx, k.ID)
	k.PracticeIDs = practiceIDs(k.Practices)
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
	r.attachPractices(ctx, out)
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

	out := toDomain(m)
	out.Practices = r.practicesForKinesiologist(ctx, out.ID)
	out.PracticeIDs = practiceIDs(out.Practices)
	return out, true, nil
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

	out := toDomain(m)
	out.Practices = r.practicesForKinesiologist(ctx, out.ID)
	out.PracticeIDs = practiceIDs(out.Practices)
	return out, true, nil
}

func (r *Repository) ListSpecialties(ctx context.Context, includeInactive bool) ([]domain.Specialty, error) {
	q := r.db.WithContext(ctx).Model(&SpecialtyModel{})
	if !includeInactive {
		q = q.Where("active = true")
	}
	var models []SpecialtyModel
	if err := q.Order("name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Specialty, 0, len(models))
	for _, model := range models {
		out = append(out, toSpecialtyDomain(model))
	}
	return out, nil
}

func (r *Repository) SaveSpecialty(ctx context.Context, specialty domain.Specialty) (domain.Specialty, error) {
	model := SpecialtyModel{
		ID:     specialty.ID,
		Name:   specialty.Name,
		Active: specialty.Active,
	}
	tx := r.db.WithContext(ctx).Model(&SpecialtyModel{}).Where("id = ?", specialty.ID).Updates(map[string]any{
		"name":       specialty.Name,
		"active":     specialty.Active,
		"updated_at": time.Now().UTC(),
	})
	if tx.Error != nil {
		if duplicateErr := duplicateError(tx.Error); duplicateErr != nil {
			return domain.Specialty{}, duplicateErr
		}
		return domain.Specialty{}, tx.Error
	}
	if tx.RowsAffected == 0 {
		if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
			if duplicateErr := duplicateError(err); duplicateErr != nil {
				return domain.Specialty{}, duplicateErr
			}
			return domain.Specialty{}, err
		}
	} else if err := r.db.WithContext(ctx).First(&model, "id = ?", specialty.ID).Error; err != nil {
		return domain.Specialty{}, err
	}
	return toSpecialtyDomain(model), nil
}

func (r *Repository) ListPractices(ctx context.Context, includeInactive bool) ([]domain.Practice, error) {
	q := r.db.WithContext(ctx).Model(&PracticeModel{})
	if !includeInactive {
		q = q.Where("active = true")
	}
	var models []PracticeModel
	if err := q.Order("name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Practice, 0, len(models))
	for _, model := range models {
		out = append(out, toPracticeDomain(model))
	}
	return out, nil
}

func (r *Repository) SavePractice(ctx context.Context, practice domain.Practice) (domain.Practice, error) {
	model := PracticeModel{
		ID:          practice.ID,
		SpecialtyID: practice.SpecialtyID,
		Name:        practice.Name,
		Description: practice.Description,
		Active:      practice.Active,
	}
	tx := r.db.WithContext(ctx).Model(&PracticeModel{}).Where("id = ?", practice.ID).Updates(map[string]any{
		"specialty_id": practice.SpecialtyID,
		"name":         practice.Name,
		"description":  practice.Description,
		"active":       practice.Active,
		"updated_at":   time.Now().UTC(),
	})
	if tx.Error != nil {
		if duplicateErr := duplicateError(tx.Error); duplicateErr != nil {
			return domain.Practice{}, duplicateErr
		}
		return domain.Practice{}, tx.Error
	}
	if tx.RowsAffected == 0 {
		if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
			if duplicateErr := duplicateError(err); duplicateErr != nil {
				return domain.Practice{}, duplicateErr
			}
			return domain.Practice{}, err
		}
	} else if err := r.db.WithContext(ctx).First(&model, "id = ?", practice.ID).Error; err != nil {
		return domain.Practice{}, err
	}
	return toPracticeDomain(model), nil
}

func duplicateError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return nil
	}

	switch pgErr.ConstraintName {
	case "ux_kinesiologists_email":
		return domain.ErrDuplicateEmail
	case "specialties_name_key", "ux_practices_specialty_name":
		return domain.ErrDuplicateName
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
		WorkDays:      decodeWorkDays(m.WorkDays),
		Active:        m.Active,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func toSpecialtyDomain(m SpecialtyModel) domain.Specialty {
	return domain.Specialty{
		ID:        m.ID,
		Name:      m.Name,
		Active:    m.Active,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toPracticeDomain(m PracticeModel) domain.Practice {
	return domain.Practice{
		ID:          m.ID,
		SpecialtyID: m.SpecialtyID,
		Name:        m.Name,
		Description: m.Description,
		Active:      m.Active,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func replaceKinesiologistPractices(tx *gorm.DB, kinesiologistID uuid.UUID, ids []uuid.UUID) error {
	if err := tx.Where("kinesiologist_id = ?", kinesiologistID).Delete(&KinesiologistPracticeModel{}).Error; err != nil {
		return err
	}
	ids = domain.NormalizePracticeIDs(ids)
	if len(ids) == 0 {
		return nil
	}
	models := make([]KinesiologistPracticeModel, 0, len(ids))
	for _, id := range ids {
		models = append(models, KinesiologistPracticeModel{
			KinesiologistID: kinesiologistID,
			PracticeID:      id,
		})
	}
	return tx.Create(&models).Error
}

func (r *Repository) attachPractices(ctx context.Context, items []domain.Kinesiologist) {
	for index := range items {
		items[index].Practices = r.practicesForKinesiologist(ctx, items[index].ID)
		items[index].PracticeIDs = practiceIDs(items[index].Practices)
	}
}

func (r *Repository) practicesForKinesiologist(ctx context.Context, kinesiologistID uuid.UUID) []domain.Practice {
	var models []PracticeModel
	err := r.db.WithContext(ctx).
		Table("practices").
		Select("practices.*").
		Joins("JOIN kinesiologist_practices kp ON kp.practice_id = practices.id").
		Where("kp.kinesiologist_id = ?", kinesiologistID).
		Order("practices.name ASC").
		Find(&models).Error
	if err != nil {
		return nil
	}
	out := make([]domain.Practice, 0, len(models))
	for _, model := range models {
		out = append(out, toPracticeDomain(model))
	}
	return out
}

func practiceIDs(practices []domain.Practice) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(practices))
	for _, practice := range practices {
		out = append(out, practice.ID)
	}
	return out
}

func encodeWorkDays(days []int) string {
	normalized := domain.NormalizeWorkDays(days)
	parts := make([]string, 0, len(normalized))
	for _, day := range normalized {
		parts = append(parts, strconv.Itoa(day))
	}
	return strings.Join(parts, ",")
}

func decodeWorkDays(value string) []int {
	parts := strings.Split(strings.TrimSpace(value), ",")
	days := make([]int, 0, len(parts))
	for _, part := range parts {
		day, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			continue
		}
		days = append(days, day)
	}
	return domain.NormalizeWorkDays(days)
}
