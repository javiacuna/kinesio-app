-- +goose Up
CREATE TABLE IF NOT EXISTS patient_evolutions (
  id UUID PRIMARY KEY,
  patient_id UUID NOT NULL REFERENCES patients(id),
  kinesiologist_id UUID NOT NULL REFERENCES kinesiologists(id),
  appointment_id UUID NULL REFERENCES appointments(id),
  pain_level INT NULL,                  -- 0..10 
  notes TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_patient_evolutions_patient ON patient_evolutions(patient_id);
CREATE INDEX IF NOT EXISTS idx_patient_evolutions_kinesiologist ON patient_evolutions(kinesiologist_id);
CREATE INDEX IF NOT EXISTS idx_patient_evolutions_appointment ON patient_evolutions(appointment_id);

-- +goose Down
DROP TABLE IF EXISTS patient_evolutions;
