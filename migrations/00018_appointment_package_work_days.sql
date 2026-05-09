-- +goose Up
ALTER TABLE appointment_packages
  ADD COLUMN IF NOT EXISTS work_days TEXT NULL;

-- +goose Down
ALTER TABLE appointment_packages
  DROP COLUMN IF EXISTS work_days;
