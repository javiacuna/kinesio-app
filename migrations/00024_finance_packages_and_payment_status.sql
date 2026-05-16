-- +goose Up
ALTER TABLE appointment_packages
  ADD COLUMN IF NOT EXISTS practice_id UUID NULL REFERENCES practices(id),
  ADD COLUMN IF NOT EXISTS financier_id UUID NULL REFERENCES financiers(id);

ALTER TABLE financial_movements
  ADD COLUMN IF NOT EXISTS collection_status TEXT NOT NULL DEFAULT 'pending',
  ADD COLUMN IF NOT EXISTS professional_payment_status TEXT NOT NULL DEFAULT 'pending',
  ADD COLUMN IF NOT EXISTS collected_at TIMESTAMPTZ NULL,
  ADD COLUMN IF NOT EXISTS professional_paid_at TIMESTAMPTZ NULL;

UPDATE financial_movements
SET
  collection_status = CASE WHEN status = 'collected' OR status = 'paid' THEN 'collected' ELSE collection_status END,
  professional_payment_status = CASE WHEN status = 'paid' THEN 'paid' ELSE professional_payment_status END
WHERE status IN ('collected', 'paid');

CREATE INDEX IF NOT EXISTS ix_financial_movements_collection_status
  ON financial_movements(collection_status, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_financial_movements_professional_payment_status
  ON financial_movements(professional_payment_status, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS ix_financial_movements_professional_payment_status;
DROP INDEX IF EXISTS ix_financial_movements_collection_status;

ALTER TABLE financial_movements
  DROP COLUMN IF EXISTS professional_paid_at,
  DROP COLUMN IF EXISTS collected_at,
  DROP COLUMN IF EXISTS professional_payment_status,
  DROP COLUMN IF EXISTS collection_status;

ALTER TABLE appointment_packages
  DROP COLUMN IF EXISTS financier_id,
  DROP COLUMN IF EXISTS practice_id;
