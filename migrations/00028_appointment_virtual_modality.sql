-- +goose Up
ALTER TABLE appointments
  ADD COLUMN IF NOT EXISTS modality TEXT NOT NULL DEFAULT 'in_person',
  ADD COLUMN IF NOT EXISTS video_call_url TEXT NULL,
  ADD COLUMN IF NOT EXISTS video_provider TEXT NULL,
  ADD COLUMN IF NOT EXISTS video_meeting_id TEXT NULL;

ALTER TABLE appointments
  DROP CONSTRAINT IF EXISTS ck_appointments_modality;

ALTER TABLE appointments
  ADD CONSTRAINT ck_appointments_modality CHECK (modality IN ('in_person', 'virtual'));

-- +goose Down
ALTER TABLE appointments
  DROP CONSTRAINT IF EXISTS ck_appointments_modality;

ALTER TABLE appointments
  DROP COLUMN IF EXISTS video_meeting_id,
  DROP COLUMN IF EXISTS video_provider,
  DROP COLUMN IF EXISTS video_call_url,
  DROP COLUMN IF EXISTS modality;
