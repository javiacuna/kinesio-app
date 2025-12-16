package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/materials/domain"
)

type ReturnMaterialUseCase struct {
	repo domain.Repository
}

func NewReturnMaterialUseCase(repo domain.Repository) *ReturnMaterialUseCase {
	return &ReturnMaterialUseCase{repo: repo}
}

func (uc *ReturnMaterialUseCase) Execute(ctx context.Context, loanID uuid.UUID) (domain.MaterialLoan, error) {
	loan, found, err := uc.repo.GetLoanByID(ctx, loanID)
	if err != nil {
		return domain.MaterialLoan{}, err
	}
	if !found {
		return domain.MaterialLoan{}, domain.ErrNotFound
	}
	if loan.ReturnedAt != nil {
		return domain.MaterialLoan{}, domain.ErrAlreadyReturned
	}

	now := time.Now().UTC()

	// marcar returned y devolver stock
	if err := uc.repo.MarkReturned(ctx, loanID, now); err != nil {
		return domain.MaterialLoan{}, err
	}
	if err := uc.repo.IncrementAvailable(ctx, loan.MaterialID, loan.Qty); err != nil {
		return domain.MaterialLoan{}, err
	}

	out, _, err := uc.repo.GetLoanByID(ctx, loanID)
	if err != nil {
		return domain.MaterialLoan{}, err
	}
	return out, nil
}
