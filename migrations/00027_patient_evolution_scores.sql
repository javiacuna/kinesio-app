-- +goose Up
ALTER TABLE patient_evolutions
  ADD COLUMN IF NOT EXISTS mobility_score INT NULL,
  ADD COLUMN IF NOT EXISTS strength_score INT NULL,
  ADD COLUMN IF NOT EXISTS functional_score INT NULL;

-- +goose StatementBegin
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_patient_evolutions_mobility_score'
  ) THEN
    ALTER TABLE patient_evolutions
      ADD CONSTRAINT chk_patient_evolutions_mobility_score
      CHECK (mobility_score IS NULL OR (mobility_score >= 0 AND mobility_score <= 100));
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_patient_evolutions_strength_score'
  ) THEN
    ALTER TABLE patient_evolutions
      ADD CONSTRAINT chk_patient_evolutions_strength_score
      CHECK (strength_score IS NULL OR (strength_score >= 0 AND strength_score <= 100));
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_patient_evolutions_functional_score'
  ) THEN
    ALTER TABLE patient_evolutions
      ADD CONSTRAINT chk_patient_evolutions_functional_score
      CHECK (functional_score IS NULL OR (functional_score >= 0 AND functional_score <= 100));
  END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
ALTER TABLE patient_evolutions
  DROP CONSTRAINT IF EXISTS chk_patient_evolutions_functional_score,
  DROP CONSTRAINT IF EXISTS chk_patient_evolutions_strength_score,
  DROP CONSTRAINT IF EXISTS chk_patient_evolutions_mobility_score,
  DROP COLUMN IF EXISTS functional_score,
  DROP COLUMN IF EXISTS strength_score,
  DROP COLUMN IF EXISTS mobility_score;
