package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/javiacuna/kinesio-backend/internal/materials/domain"
)

type ReminderNotifier interface {
	Create(ctx context.Context, recipientEmail string, recipientRole *string, loan domain.MaterialLoan, material domain.Material) error
}

type SendDueRemindersUseCase struct {
	repo     domain.Repository
	notifier ReminderNotifier
}

func NewSendDueRemindersUseCase(repo domain.Repository, notifier ReminderNotifier) *SendDueRemindersUseCase {
	return &SendDueRemindersUseCase{repo: repo, notifier: notifier}
}

func (uc *SendDueRemindersUseCase) Execute(ctx context.Context) (int, error) {
	now := time.Now().UTC()

	loans, err := uc.repo.ListOverdueUnnotified(ctx, now)
	if err != nil {
		return 0, err
	}

	sent := 0
	for _, loan := range loans {
		material, found, err := uc.repo.GetMaterialByID(ctx, loan.MaterialID)
		if err != nil {
			continue
		}
		if !found {
			material = domain.Material{ID: loan.MaterialID, Name: "material"}
		}

		if loan.LoanedByEmail != nil && strings.TrimSpace(*loan.LoanedByEmail) != "" {
			_ = uc.notifier.Create(ctx, *loan.LoanedByEmail, loan.LoanedByRole, loan, material)
		}

		if err := uc.repo.MarkDueReminderSent(ctx, loan.ID, now); err != nil {
			continue
		}
		sent++
	}

	return sent, nil
}
