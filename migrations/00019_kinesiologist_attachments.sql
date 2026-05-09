-- +goose Up
CREATE TABLE IF NOT EXISTS kinesiologist_attachments (
  id UUID PRIMARY KEY,
  kinesiologist_id UUID NOT NULL REFERENCES kinesiologists(id) ON DELETE CASCADE,
  file_name TEXT NOT NULL,
  content_type TEXT NOT NULL,
  size_bytes BIGINT NOT NULL,
  storage_path TEXT NOT NULL,
  kind TEXT NOT NULL,
  category TEXT NOT NULL DEFAULT 'otro',
  notes TEXT NULL,
  uploaded_by_email TEXT NULL,
  uploaded_by_role TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_kinesiologist_attachments_kinesiologist
  ON kinesiologist_attachments(kinesiologist_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS kinesiologist_attachments;
