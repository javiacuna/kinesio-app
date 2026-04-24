package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/staff/domain"
	"github.com/javiacuna/kinesio-backend/internal/staff/ports"
)

type SaveStaffMemberInput struct {
	ID          string
	FirstName   string
	LastName    string
	Email       string
	Role        string
	Phone       *string
	Active      bool
	FirebaseUID *string
}

type SaveStaffMemberUseCase struct {
	repo ports.Repository
}

func NewSaveStaffMemberUseCase(repo ports.Repository) *SaveStaffMemberUseCase {
	return &SaveStaffMemberUseCase{repo: repo}
}

func (uc *SaveStaffMemberUseCase) Create(ctx context.Context, in SaveStaffMemberInput) (domain.StaffMember, map[string]string, error) {
	member, errs := buildStaffMember(in)
	if len(errs) > 0 {
		return domain.StaffMember{}, errs, domain.ErrValidation
	}

	created, err := uc.repo.Create(ctx, member)
	return created, nil, err
}

func (uc *SaveStaffMemberUseCase) Update(ctx context.Context, in SaveStaffMemberInput) (domain.StaffMember, map[string]string, error) {
	member, errs := buildStaffMember(in)
	id, err := uuid.Parse(strings.TrimSpace(in.ID))
	if err != nil {
		errs["id"] = "UUID invalido"
	}
	if len(errs) > 0 {
		return domain.StaffMember{}, errs, domain.ErrValidation
	}

	member.ID = id
	updated, err := uc.repo.Update(ctx, member)
	return updated, nil, err
}

func buildStaffMember(in SaveStaffMemberInput) (domain.StaffMember, map[string]string) {
	errs := map[string]string{}
	firstName := strings.TrimSpace(in.FirstName)
	lastName := strings.TrimSpace(in.LastName)
	email := domain.NormalizeEmail(in.Email)
	role := domain.Role(strings.TrimSpace(in.Role))

	if firstName == "" {
		errs["first_name"] = "Campo obligatorio"
	}
	if lastName == "" {
		errs["last_name"] = "Campo obligatorio"
	}
	if email == "" {
		errs["email"] = "Campo obligatorio"
	} else if !strings.Contains(email, "@") {
		errs["email"] = "Formato invalido"
	}
	if !domain.IsValidRole(role) {
		errs["role"] = "Rol invalido"
	}

	return domain.NewStaffMember(firstName, lastName, email, role, in.Phone, in.Active, in.FirebaseUID), errs
}
