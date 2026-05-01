package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/materials/domain"
)

type CreateMaterialInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	TotalQty    int     `json:"total_qty"`
	ActorEmail  string  `json:"-"`
	ActorRole   string  `json:"-"`
}

type CreateMaterialUseCase struct {
	repo domain.Repository
}

func NewCreateMaterialUseCase(repo domain.Repository) *CreateMaterialUseCase {
	return &CreateMaterialUseCase{repo: repo}
}

func (uc *CreateMaterialUseCase) Execute(ctx context.Context, in CreateMaterialInput) (domain.Material, map[string]string, error) {
	validation := map[string]string{}

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

	now := time.Now().UTC()

	m := domain.Material{
		ID:             uuid.New(),
		Name:           name,
		Description:    in.Description,
		TotalQty:       in.TotalQty,
		AvailableQty:   in.TotalQty,
		CreatedAt:      now,
		UpdatedAt:      now,
		CreatedByEmail: trimOptionalString(in.ActorEmail),
		CreatedByRole:  trimOptionalString(in.ActorRole),
		UpdatedByEmail: trimOptionalString(in.ActorEmail),
		UpdatedByRole:  trimOptionalString(in.ActorRole),
	}

	out, err := uc.repo.CreateMaterial(ctx, m)
	if err != nil {
		return domain.Material{}, nil, err
	}
	return out, nil, nil
}

func trimOptionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
