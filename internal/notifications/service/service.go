package service

import (
	"context"
	"html"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/javiacuna/kinesio-backend/internal/notifications/domain"
	"github.com/javiacuna/kinesio-backend/internal/notifications/email"
)

type Repository interface {
	Create(ctx context.Context, item domain.Notification) error
	ListByRecipient(ctx context.Context, email string, limit int, unreadOnly bool) ([]domain.Notification, error)
	CountUnread(ctx context.Context, email string) (int64, error)
	MarkRead(ctx context.Context, id uuid.UUID, email string) error
	MarkAllRead(ctx context.Context, email string) error
}

type Mailer interface {
	Send(ctx context.Context, msg email.Message) error
}

type Service struct {
	repo   Repository
	mailer Mailer
}

func New(repo Repository, mailer Mailer) *Service {
	return &Service{repo: repo, mailer: mailer}
}

func (s *Service) Create(ctx context.Context, item domain.Notification) error {
	if err := s.repo.Create(ctx, item); err != nil {
		return err
	}
	s.sendEmail(item)
	return nil
}

func (s *Service) ListByRecipient(ctx context.Context, email string, limit int, unreadOnly bool) ([]domain.Notification, error) {
	return s.repo.ListByRecipient(ctx, email, limit, unreadOnly)
}

func (s *Service) CountUnread(ctx context.Context, email string) (int64, error) {
	return s.repo.CountUnread(ctx, email)
}

func (s *Service) MarkRead(ctx context.Context, id uuid.UUID, email string) error {
	return s.repo.MarkRead(ctx, id, email)
}

func (s *Service) MarkAllRead(ctx context.Context, email string) error {
	return s.repo.MarkAllRead(ctx, email)
}

func (s *Service) sendEmail(item domain.Notification) {
	if s.mailer == nil || strings.TrimSpace(item.RecipientEmail) == "" {
		return
	}

	msg := email.Message{
		ToEmail: item.RecipientEmail,
		Subject: item.Title,
		Text:    item.Title + "\n\n" + item.Message,
		HTML:    notificationHTML(item),
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.mailer.Send(ctx, msg); err != nil {
			log.Warn().
				Err(err).
				Str("notification_id", item.ID.String()).
				Str("recipient_email", item.RecipientEmail).
				Msg("notification email could not be sent")
		}
	}()
}

func notificationHTML(item domain.Notification) string {
	title := html.EscapeString(strings.TrimSpace(item.Title))
	message := html.EscapeString(strings.TrimSpace(item.Message))
	return `<!doctype html>
<html>
  <body style="font-family:Arial,sans-serif;color:#111827;line-height:1.5">
    <h2 style="margin:0 0 12px">` + title + `</h2>
    <p style="margin:0 0 16px">` + message + `</p>
    <p style="margin:0;color:#6b7280;font-size:13px">Kinesio App</p>
  </body>
</html>`
}
