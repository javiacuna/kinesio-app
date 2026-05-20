-- +goose Up
CREATE TABLE IF NOT EXISTS notifications (
  id UUID PRIMARY KEY,
  recipient_email TEXT NOT NULL,
  recipient_role TEXT NULL,
  type TEXT NOT NULL,
  title TEXT NOT NULL,
  message TEXT NOT NULL,
  entity_type TEXT NULL,
  entity_id UUID NULL,
  read_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_notifications_recipient_created
  ON notifications(recipient_email, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_notifications_unread
  ON notifications(recipient_email, read_at)
  WHERE read_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS ix_notifications_unread;
DROP INDEX IF EXISTS ix_notifications_recipient_created;
DROP TABLE IF EXISTS notifications;
