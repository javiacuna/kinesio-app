-- +goose Up
ALTER TABLE patients
  ADD COLUMN IF NOT EXISTS clinical_notes_updated_by_email TEXT NULL,
  ADD COLUMN IF NOT EXISTS clinical_notes_updated_by_role TEXT NULL,
  ADD COLUMN IF NOT EXISTS clinical_notes_updated_at TIMESTAMPTZ NULL;

ALTER TABLE exercise_plans
  ADD COLUMN IF NOT EXISTS created_by_email TEXT NULL,
  ADD COLUMN IF NOT EXISTS created_by_role TEXT NULL,
  ADD COLUMN IF NOT EXISTS updated_by_email TEXT NULL,
  ADD COLUMN IF NOT EXISTS updated_by_role TEXT NULL;

CREATE TABLE IF NOT EXISTS patient_attachments (
  id UUID PRIMARY KEY,
  patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
  file_name TEXT NOT NULL,
  content_type TEXT NOT NULL,
  size_bytes BIGINT NOT NULL,
  storage_path TEXT NOT NULL,
  kind TEXT NOT NULL,
  notes TEXT NULL,
  uploaded_by_email TEXT NULL,
  uploaded_by_role TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_patient_attachments_patient
  ON patient_attachments(patient_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS patient_attachments;

ALTER TABLE exercise_plans
  DROP COLUMN IF EXISTS updated_by_role,
  DROP COLUMN IF EXISTS updated_by_email,
  DROP COLUMN IF EXISTS created_by_role,
  DROP COLUMN IF EXISTS created_by_email;

ALTER TABLE patients
  DROP COLUMN IF EXISTS clinical_notes_updated_at,
  DROP COLUMN IF EXISTS clinical_notes_updated_by_role,
  DROP COLUMN IF EXISTS clinical_notes_updated_by_email;
