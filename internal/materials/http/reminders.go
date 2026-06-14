package http

import (
	"context"
	"strconv"
	"strings"

	"github.com/javiacuna/kinesio-backend/internal/materials/domain"
	notificationDomain "github.com/javiacuna/kinesio-backend/internal/notifications/domain"
)

type ReminderNotificationCreator interface {
	Create(ctx context.Context, item notificationDomain.Notification) error
}

// DueReminderNotifier adapts the notifications service to materials' ReminderNotifier port.
type DueReminderNotifier struct {
	notifier ReminderNotificationCreator
}

func NewDueReminderNotifier(notifier ReminderNotificationCreator) *DueReminderNotifier {
	return &DueReminderNotifier{notifier: notifier}
}

func (n *DueReminderNotifier) Create(ctx context.Context, recipientEmail string, recipientRole *string, loan domain.MaterialLoan, material domain.Material) error {
	entityType := "material_loan"
	notificationType := "material_loan_due"
	title := "Recordatorio de devolución de material"

	var b strings.Builder
	b.WriteString("El préstamo de " + strconv.Itoa(loan.Qty) + " unidad(es) de \"" + strings.TrimSpace(material.Name) + "\" venció")
	if loan.DueDate != nil {
		b.WriteString(" el " + loan.DueDate.Format("02/01/2006"))
	}
	b.WriteString(" y todavía no fue devuelto.")

	return n.notifier.Create(ctx, notificationDomain.NewNotification(notificationDomain.NewNotificationInput{
		RecipientEmail: recipientEmail,
		RecipientRole:  recipientRole,
		Type:           notificationType,
		Title:          title,
		Message:        b.String(),
		EntityType:     &entityType,
		EntityID:       &loan.ID,
	}))
}
