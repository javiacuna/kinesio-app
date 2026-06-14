-- +goose Up
CREATE TABLE IF NOT EXISTS support_tickets (
  id UUID PRIMARY KEY,
  user_email TEXT NOT NULL,
  user_role TEXT NULL,
  subject TEXT NOT NULL,
  message TEXT NOT NULL,
  attachment_file_name TEXT NULL,
  attachment_content_type TEXT NULL,
  attachment_size_bytes BIGINT NULL,
  attachment_storage_path TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_support_tickets_user_created
  ON support_tickets(user_email, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS ix_support_tickets_user_created;
DROP TABLE IF EXISTS support_tickets;
