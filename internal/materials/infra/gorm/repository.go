package gorm

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/javiacuna/kinesio-backend/internal/materials/domain"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// -------- Materials

func (r *Repository) CreateMaterial(ctx context.Context, m domain.Material) (domain.Material, error) {
	model := toMaterialModel(m)
	err := r.db.WithContext(ctx).Create(&model).Error
	if err != nil {
		// unique name -> duplicate
		if strings.Contains(strings.ToLower(err.Error()), "ux_materials_name") ||
			strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return domain.Material{}, domain.ErrDuplicateName
		}
		return domain.Material{}, err
	}
	out, _, err := r.GetMaterialByID(ctx, uuid.MustParse(model.ID))
	if err != nil {
		return domain.Material{}, err
	}
	return out, nil
}

func (r *Repository) ListMaterials(ctx context.Context, limit int) ([]domain.Material, error) {
	var ms []MaterialModel
	if err := r.db.WithContext(ctx).Order("name asc").Limit(limit).Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Material, 0, len(ms))
	for _, m := range ms {
		out = append(out, toMaterialDomain(m))
	}
	return out, nil
}

func (r *Repository) GetMaterialByID(ctx context.Context, id uuid.UUID) (domain.Material, bool, error) {
	var m MaterialModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id.String()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Material{}, false, nil
		}
		return domain.Material{}, false, err
	}
	return toMaterialDomain(m), true, nil
}

func (r *Repository) DecrementAvailable(ctx context.Context, materialID uuid.UUID, qty int) error {
	// UPDATE materials SET available_qty = available_qty - qty WHERE id=? AND available_qty >= qty
	res := r.db.WithContext(ctx).
		Model(&MaterialModel{}).
		Where("id = ? AND available_qty >= ?", materialID.String(), qty).
		Update("available_qty", gorm.Expr("available_qty - ?", qty))

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrInsufficientStock
	}
	return nil
}

func (r *Repository) IncrementAvailable(ctx context.Context, materialID uuid.UUID, qty int) error {
	res := r.db.WithContext(ctx).
		Model(&MaterialModel{}).
		Where("id = ?", materialID.String()).
		Update("available_qty", gorm.Expr("available_qty + ?", qty))
	return res.Error
}

// -------- Loans

func (r *Repository) CreateLoan(ctx context.Context, l domain.MaterialLoan) (domain.MaterialLoan, error) {
	model := toLoanModel(l)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.MaterialLoan{}, err
	}
	out, _, err := r.GetLoanByID(ctx, uuid.MustParse(model.ID))
	if err != nil {
		return domain.MaterialLoan{}, err
	}
	return out, nil
}

func (r *Repository) GetLoanByID(ctx context.Context, id uuid.UUID) (domain.MaterialLoan, bool, error) {
	var m MaterialLoanModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id.String()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.MaterialLoan{}, false, nil
		}
		return domain.MaterialLoan{}, false, err
	}
	return toLoanDomain(m), true, nil
}

func (r *Repository) ListLoansByPatient(ctx context.Context, patientID uuid.UUID, onlyActive bool, limit int) ([]domain.MaterialLoan, error) {
	q := r.db.WithContext(ctx).Order("loaned_at desc").Limit(limit).Where("patient_id = ?", patientID.String())
	if onlyActive {
		q = q.Where("returned_at IS NULL")
	}

	var ms []MaterialLoanModel
	if err := q.Find(&ms).Error; err != nil {
		return nil, err
	}

	out := make([]domain.MaterialLoan, 0, len(ms))
	for _, m := range ms {
		out = append(out, toLoanDomain(m))
	}
	return out, nil
}

func (r *Repository) MarkReturned(ctx context.Context, loanID uuid.UUID, returnedAt time.Time) error {
	// solo si no estaba returned
	res := r.db.WithContext(ctx).
		Model(&MaterialLoanModel{}).
		Where("id = ? AND returned_at IS NULL", loanID.String()).
		Update("returned_at", returnedAt)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrAlreadyReturned
	}
	return nil
}

// -------- mappers

func toMaterialModel(m domain.Material) MaterialModel {
	return MaterialModel{
		ID:           m.ID.String(),
		Name:         m.Name,
		Description:  m.Description,
		TotalQty:     m.TotalQty,
		AvailableQty: m.AvailableQty,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func toMaterialDomain(m MaterialModel) domain.Material {
	return domain.Material{
		ID:           uuid.MustParse(m.ID),
		Name:         m.Name,
		Description:  m.Description,
		TotalQty:     m.TotalQty,
		AvailableQty: m.AvailableQty,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func toLoanModel(l domain.MaterialLoan) MaterialLoanModel {
	var returnedAt *time.Time = nil
	if l.ReturnedAt != nil {
		returnedAt = l.ReturnedAt
	}
	return MaterialLoanModel{
		ID:              l.ID.String(),
		MaterialID:      l.MaterialID.String(),
		PatientID:       l.PatientID.String(),
		KinesiologistID: l.KinesiologistID.String(),
		Qty:             l.Qty,
		Notes:           l.Notes,
		LoanedAt:        l.LoanedAt,
		ReturnedAt:      returnedAt,
	}
}

func toLoanDomain(m MaterialLoanModel) domain.MaterialLoan {
	return domain.MaterialLoan{
		ID:              uuid.MustParse(m.ID),
		MaterialID:      uuid.MustParse(m.MaterialID),
		PatientID:       uuid.MustParse(m.PatientID),
		KinesiologistID: uuid.MustParse(m.KinesiologistID),
		Qty:             m.Qty,
		Notes:           m.Notes,
		LoanedAt:        m.LoanedAt,
		ReturnedAt:      m.ReturnedAt,
	}
}
