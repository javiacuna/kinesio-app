-- +goose Up
ALTER TABLE patients ADD COLUMN IF NOT EXISTS financier_member_number TEXT NULL;

-- +goose Down
ALTER TABLE patients DROP COLUMN IF EXISTS financier_member_number;
