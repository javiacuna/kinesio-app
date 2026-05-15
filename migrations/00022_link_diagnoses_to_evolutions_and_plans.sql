-- +goose Up
ALTER TABLE patient_evolutions
  ADD COLUMN IF NOT EXISTS patient_diagnosis_id UUID NULL REFERENCES patient_diagnoses(id) ON DELETE SET NULL;

ALTER TABLE exercise_plans
  ADD COLUMN IF NOT EXISTS patient_diagnosis_id UUID NULL REFERENCES patient_diagnoses(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_patient_evolutions_diagnosis
  ON patient_evolutions(patient_diagnosis_id);

CREATE INDEX IF NOT EXISTS idx_exercise_plans_diagnosis
  ON exercise_plans(patient_diagnosis_id);

-- +goose Down
DROP INDEX IF EXISTS idx_exercise_plans_diagnosis;
DROP INDEX IF EXISTS idx_patient_evolutions_diagnosis;

ALTER TABLE exercise_plans
  DROP COLUMN IF EXISTS patient_diagnosis_id;

ALTER TABLE patient_evolutions
  DROP COLUMN IF EXISTS patient_diagnosis_id;
