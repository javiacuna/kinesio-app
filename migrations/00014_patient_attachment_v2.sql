-- +goose Up
ALTER TABLE patient_attachments
  ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT 'otro',
  ADD COLUMN IF NOT EXISTS patient_visible BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS updated_by_email TEXT NULL,
  ADD COLUMN IF NOT EXISTS updated_by_role TEXT NULL,
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NULL;

-- +goose Down
ALTER TABLE patient_attachments
  DROP COLUMN IF EXISTS updated_at,
  DROP COLUMN IF EXISTS updated_by_role,
  DROP COLUMN IF EXISTS updated_by_email,
  DROP COLUMN IF EXISTS patient_visible,
  DROP COLUMN IF EXISTS category;
