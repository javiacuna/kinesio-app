package usecase

import (
	"context"

	"github.com/javiacuna/kinesio-backend/internal/staff/domain"
	"github.com/javiacuna/kinesio-backend/internal/staff/ports"
)

type ListStaffMembersUseCase struct {
	repo ports.Repository
}

func NewListStaffMembersUseCase(repo ports.Repository) *ListStaffMembersUseCase {
	return &ListStaffMembersUseCase{repo: repo}
}

func (uc *ListStaffMembersUseCase) Execute(ctx context.Context, includeInactive bool) ([]domain.StaffMember, error) {
	return uc.repo.List(ctx, includeInactive)
}
