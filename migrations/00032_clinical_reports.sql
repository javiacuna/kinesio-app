-- +goose Up
CREATE TABLE IF NOT EXISTS clinical_reports (
  id UUID PRIMARY KEY,
  patient_id UUID NOT NULL REFERENCES patients(id),
  kinesiologist_id UUID NOT NULL REFERENCES kinesiologists(id),
  period_from DATE NOT NULL,
  period_to DATE NOT NULL,
  evolution_count INT NOT NULL DEFAULT 0,
  avg_pain_level NUMERIC(5,2) NULL,
  avg_mobility_score NUMERIC(5,2) NULL,
  avg_strength_score NUMERIC(5,2) NULL,
  avg_functional_score NUMERIC(5,2) NULL,
  active_plan_count INT NOT NULL DEFAULT 0,
  active_plan_item_count INT NOT NULL DEFAULT 0,
  summary TEXT NOT NULL,
  recommendations TEXT NULL,
  created_by_email TEXT NULL,
  created_by_role TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_clinical_reports_period CHECK (period_to >= period_from),
  CONSTRAINT chk_clinical_reports_evolution_count CHECK (evolution_count >= 0),
  CONSTRAINT chk_clinical_reports_active_plan_count CHECK (active_plan_count >= 0),
  CONSTRAINT chk_clinical_reports_active_plan_item_count CHECK (active_plan_item_count >= 0)
);

CREATE INDEX IF NOT EXISTS idx_clinical_reports_patient_created_at
  ON clinical_reports(patient_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS clinical_reports;
