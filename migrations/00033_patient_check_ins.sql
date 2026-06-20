-- +goose Up
CREATE TABLE IF NOT EXISTS patient_check_ins (
  id UUID PRIMARY KEY,
  patient_id UUID NOT NULL REFERENCES patients(id),
  pain_level INT NULL,
  mobility_score INT NULL,
  strength_score INT NULL,
  functional_score INT NULL,
  notes TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_patient_check_ins_pain_level
    CHECK (pain_level IS NULL OR (pain_level >= 0 AND pain_level <= 10)),
  CONSTRAINT chk_patient_check_ins_mobility_score
    CHECK (mobility_score IS NULL OR (mobility_score >= 0 AND mobility_score <= 100)),
  CONSTRAINT chk_patient_check_ins_strength_score
    CHECK (strength_score IS NULL OR (strength_score >= 0 AND strength_score <= 100)),
  CONSTRAINT chk_patient_check_ins_functional_score
    CHECK (functional_score IS NULL OR (functional_score >= 0 AND functional_score <= 100))
);

CREATE INDEX IF NOT EXISTS idx_patient_check_ins_patient_created_at
  ON patient_check_ins(patient_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS patient_check_ins;
