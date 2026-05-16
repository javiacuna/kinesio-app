-- +goose Up
ALTER TABLE financial_movements
  ADD COLUMN IF NOT EXISTS cancellation_reason TEXT NULL;

-- +goose Down
ALTER TABLE financial_movements
  DROP COLUMN IF EXISTS cancellation_reason;
