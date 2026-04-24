package gorm

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/javiacuna/kinesio-backend/internal/staff/domain"
	"github.com/javiacuna/kinesio-backend/internal/staff/ports"
	"gorm.io/gorm"
)

var _ ports.Repository = (*Repository)(nil)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, member domain.StaffMember) (domain.StaffMember, error) {
	m := toModel(member)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		if duplicateErr := duplicateError(err); duplicateErr != nil {
			return domain.StaffMember{}, duplicateErr
		}
		return domain.StaffMember{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) Update(ctx context.Context, member domain.StaffMember) (domain.StaffMember, error) {
	updatedAt := time.Now().UTC()
	updates := map[string]any{
		"first_name":   member.FirstName,
		"last_name":    member.LastName,
		"email":        member.Email,
		"role":         string(member.Role),
		"phone":        member.Phone,
		"active":       member.Active,
		"firebase_uid": member.FirebaseUID,
		"updated_at":   updatedAt,
	}

	tx := r.db.WithContext(ctx).Model(&StaffMemberModel{}).Where("id = ?", member.ID).Updates(updates)
	if tx.Error != nil {
		if duplicateErr := duplicateError(tx.Error); duplicateErr != nil {
			return domain.StaffMember{}, duplicateErr
		}
		return domain.StaffMember{}, tx.Error
	}
	if tx.RowsAffected == 0 {
		return domain.StaffMember{}, domain.ErrNotFound
	}
	member.UpdatedAt = updatedAt
	return member, nil
}

func (r *Repository) List(ctx context.Context, includeInactive bool) ([]domain.StaffMember, error) {
	q := r.db.WithContext(ctx).Model(&StaffMemberModel{})
	if !includeInactive {
		q = q.Where("active = true")
	}

	var models []StaffMemberModel
	if err := q.Order("last_name ASC, first_name ASC").Find(&models).Error; err != nil {
		return nil, err
	}

	out := make([]domain.StaffMember, 0, len(models))
	for _, m := range models {
		out = append(out, toDomain(m))
	}
	return out, nil
}

func duplicateError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return nil
	}
	if pgErr.ConstraintName == "ux_staff_members_email" {
		return domain.ErrDuplicateEmail
	}
	return nil
}

func toModel(member domain.StaffMember) StaffMemberModel {
	return StaffMemberModel{
		ID:          member.ID,
		FirstName:   member.FirstName,
		LastName:    member.LastName,
		Email:       member.Email,
		Role:        string(member.Role),
		Phone:       member.Phone,
		Active:      member.Active,
		FirebaseUID: member.FirebaseUID,
	}
}

func toDomain(m StaffMemberModel) domain.StaffMember {
	return domain.StaffMember{
		ID:          m.ID,
		FirstName:   m.FirstName,
		LastName:    m.LastName,
		Email:       m.Email,
		Role:        domain.Role(m.Role),
		Phone:       m.Phone,
		Active:      m.Active,
		FirebaseUID: m.FirebaseUID,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
