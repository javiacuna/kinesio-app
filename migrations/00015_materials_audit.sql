-- +goose Up
ALTER TABLE materials
  ADD COLUMN IF NOT EXISTS created_by_email TEXT NULL,
  ADD COLUMN IF NOT EXISTS created_by_role TEXT NULL,
  ADD COLUMN IF NOT EXISTS updated_by_email TEXT NULL,
  ADD COLUMN IF NOT EXISTS updated_by_role TEXT NULL;

ALTER TABLE material_loans
  ADD COLUMN IF NOT EXISTS loaned_by_email TEXT NULL,
  ADD COLUMN IF NOT EXISTS loaned_by_role TEXT NULL,
  ADD COLUMN IF NOT EXISTS returned_by_email TEXT NULL,
  ADD COLUMN IF NOT EXISTS returned_by_role TEXT NULL;

CREATE INDEX IF NOT EXISTS idx_material_loans_active_loaned_at
  ON material_loans(loaned_at DESC)
  WHERE returned_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_material_loans_active_loaned_at;

ALTER TABLE material_loans
  DROP COLUMN IF EXISTS returned_by_role,
  DROP COLUMN IF EXISTS returned_by_email,
  DROP COLUMN IF EXISTS loaned_by_role,
  DROP COLUMN IF EXISTS loaned_by_email;

ALTER TABLE materials
  DROP COLUMN IF EXISTS updated_by_role,
  DROP COLUMN IF EXISTS updated_by_email,
  DROP COLUMN IF EXISTS created_by_role,
  DROP COLUMN IF EXISTS created_by_email;
