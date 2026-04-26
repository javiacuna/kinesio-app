-- +goose Up
ALTER TABLE kinesiologists
  ADD COLUMN IF NOT EXISTS work_start_time TEXT NOT NULL DEFAULT '08:00',
  ADD COLUMN IF NOT EXISTS work_end_time TEXT NOT NULL DEFAULT '20:00';

ALTER TABLE kinesiologists
  ADD CONSTRAINT chk_kinesiologists_work_start_time
  CHECK (work_start_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$');

ALTER TABLE kinesiologists
  ADD CONSTRAINT chk_kinesiologists_work_end_time
  CHECK (work_end_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$');

ALTER TABLE kinesiologists
  ADD CONSTRAINT chk_kinesiologists_work_time_range
  CHECK (work_start_time < work_end_time);

-- +goose Down
ALTER TABLE kinesiologists
  DROP CONSTRAINT IF EXISTS chk_kinesiologists_work_time_range,
  DROP CONSTRAINT IF EXISTS chk_kinesiologists_work_end_time,
  DROP CONSTRAINT IF EXISTS chk_kinesiologists_work_start_time;

ALTER TABLE kinesiologists
  DROP COLUMN IF EXISTS work_end_time,
  DROP COLUMN IF EXISTS work_start_time;
