-- +goose Up
ALTER TABLE material_loans ADD COLUMN IF NOT EXISTS due_date TIMESTAMPTZ NULL;
ALTER TABLE material_loans ADD COLUMN IF NOT EXISTS due_reminder_sent_at TIMESTAMPTZ NULL;

ALTER TABLE patients ADD COLUMN IF NOT EXISTS financier_id UUID NULL REFERENCES financiers(id);

-- +goose Down
ALTER TABLE patients DROP COLUMN IF EXISTS financier_id;

ALTER TABLE material_loans DROP COLUMN IF EXISTS due_reminder_sent_at;
ALTER TABLE material_loans DROP COLUMN IF EXISTS due_date;
