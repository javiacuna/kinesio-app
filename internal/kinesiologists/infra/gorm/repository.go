package gorm

import (
	"context"

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
		out = append(out, domain.Kinesiologist{
			ID:            m.ID,
			FirstName:     m.FirstName,
			LastName:      m.LastName,
			Email:         m.Email,
			LicenseNumber: m.LicenseNumber,
			Active:        m.Active,
			CreatedAt:     m.CreatedAt,
			UpdatedAt:     m.UpdatedAt,
		})
	}
	return out, nil
}
