package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/materials/domain"
)

type LoanMaterialInput struct {
	MaterialID      string  `json:"material_id"`
	PatientID       string  `json:"patient_id"`
	KinesiologistID string  `json:"kinesiologist_id"`
	Qty             int     `json:"qty"`
	Notes           *string `json:"notes,omitempty"`
}

type LoanMaterialUseCase struct {
	repo domain.Repository
}

func NewLoanMaterialUseCase(repo domain.Repository) *LoanMaterialUseCase {
	return &LoanMaterialUseCase{repo: repo}
}

func (uc *LoanMaterialUseCase) Execute(ctx context.Context, in LoanMaterialInput) (domain.MaterialLoan, map[string]string, error) {
	validation := map[string]string{}

	mID, err := uuid.Parse(strings.TrimSpace(in.MaterialID))
	if err != nil {
		validation["material_id"] = "invalid_uuid"
	}
	pID, err := uuid.Parse(strings.TrimSpace(in.PatientID))
	if err != nil {
		validation["patient_id"] = "invalid_uuid"
	}
	kID, err := uuid.Parse(strings.TrimSpace(in.KinesiologistID))
	if err != nil {
		validation["kinesiologist_id"] = "invalid_uuid"
	}

	if in.Qty <= 0 {
		validation["qty"] = "must_be_>_0"
	}

	if len(validation) > 0 {
		return domain.MaterialLoan{}, validation, domain.ErrValidation
	}

	// chequear stock antes de crear el préstamo
	mat, found, err := uc.repo.GetMaterialByID(ctx, mID)
	if err != nil {
		return domain.MaterialLoan{}, nil, err
	}
	if !found {
		return domain.MaterialLoan{}, map[string]string{"material_id": "not_found"}, domain.ErrNotFound
	}
	if mat.AvailableQty < in.Qty {
		return domain.MaterialLoan{}, map[string]string{"qty": "insufficient_stock"}, domain.ErrInsufficientStock
	}

	now := time.Now().UTC()

	loan := domain.MaterialLoan{
		ID:              uuid.New(),
		MaterialID:      mID,
		PatientID:       pID,
		KinesiologistID: kID,
		Qty:             in.Qty,
		Notes:           in.Notes,
		LoanedAt:        now,
		ReturnedAt:      nil,
	}

	// en repo: transacción stock-- + create loan
	// para mantenerlo simple, : decrement available y luego create loan.
	// ideal: una transacción real en repo.
	if err := uc.repo.DecrementAvailable(ctx, mID, in.Qty); err != nil {
		return domain.MaterialLoan{}, nil, err
	}

	out, err := uc.repo.CreateLoan(ctx, loan)
	if err != nil {
		// compensación: si falló crear el loan, devolvemos stock
		_ = uc.repo.IncrementAvailable(ctx, mID, in.Qty)
		return domain.MaterialLoan{}, nil, err
	}

	return out, nil, nil
}
