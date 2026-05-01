package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/materials/domain"
)

type UpdateMaterialInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	TotalQty    int     `json:"total_qty"`
	ActorEmail  string  `json:"-"`
	ActorRole   string  `json:"-"`
}

type UpdateMaterialUseCase struct {
	repo domain.Repository
}

func NewUpdateMaterialUseCase(repo domain.Repository) *UpdateMaterialUseCase {
	return &UpdateMaterialUseCase{repo: repo}
}

func (uc *UpdateMaterialUseCase) Execute(ctx context.Context, id string, in UpdateMaterialInput) (domain.Material, map[string]string, error) {
	validation := map[string]string{}

	materialID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		validation["id"] = "invalid_uuid"
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		validation["name"] = "required"
	}
	if in.TotalQty < 0 {
		validation["total_qty"] = "must_be_>=_0"
	}
	if len(validation) > 0 {
		return domain.Material{}, validation, domain.ErrValidation
	}

	current, found, err := uc.repo.GetMaterialByID(ctx, materialID)
	if err != nil {
		return domain.Material{}, nil, err
	}
	if !found {
		return domain.Material{}, nil, domain.ErrNotFound
	}

	loanedQty := current.TotalQty - current.AvailableQty
	if in.TotalQty < loanedQty {
		return domain.Material{}, map[string]string{"total_qty": "below_currently_loaned"}, domain.ErrValidation
	}

	current.Name = name
	current.Description = in.Description
	current.TotalQty = in.TotalQty
	current.AvailableQty = in.TotalQty - loanedQty
	current.UpdatedAt = time.Now().UTC()
	current.UpdatedByEmail = trimOptionalString(in.ActorEmail)
	current.UpdatedByRole = trimOptionalString(in.ActorRole)

	out, err := uc.repo.UpdateMaterial(ctx, current)
	if err != nil {
		return domain.Material{}, nil, err
	}
	return out, nil, nil
}
