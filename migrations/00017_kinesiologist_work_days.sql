-- +goose Up
ALTER TABLE kinesiologists
  ADD COLUMN IF NOT EXISTS work_days TEXT NOT NULL DEFAULT '1,2,3,4,5';

-- +goose Down
ALTER TABLE kinesiologists
  DROP COLUMN IF EXISTS work_days;
