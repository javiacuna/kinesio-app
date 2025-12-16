package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	// Materials
	CreateMaterial(ctx context.Context, m Material) (Material, error)
	ListMaterials(ctx context.Context, limit int) ([]Material, error)
	GetMaterialByID(ctx context.Context, id uuid.UUID) (Material, bool, error)

	// Loans
	CreateLoan(ctx context.Context, l MaterialLoan) (MaterialLoan, error)
	GetLoanByID(ctx context.Context, id uuid.UUID) (MaterialLoan, bool, error)
	ListLoansByPatient(ctx context.Context, patientID uuid.UUID, onlyActive bool, limit int) ([]MaterialLoan, error)

	// Stock ops (transactional)
	DecrementAvailable(ctx context.Context, materialID uuid.UUID, qty int) error
	IncrementAvailable(ctx context.Context, materialID uuid.UUID, qty int) error
	MarkReturned(ctx context.Context, loanID uuid.UUID, returnedAt time.Time) error
}
