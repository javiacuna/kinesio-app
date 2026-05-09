-- +goose Up
CREATE TABLE IF NOT EXISTS appointment_packages (
  id UUID PRIMARY KEY,
  patient_id UUID NOT NULL REFERENCES patients(id),
  kinesiologist_id UUID NOT NULL REFERENCES kinesiologists(id),
  sessions_count INT NOT NULL,
  duration_min INT NOT NULL,
  start_date DATE NOT NULL,
  start_time TEXT NOT NULL,
  weekdays_only BOOLEAN NOT NULL DEFAULT true,
  notes TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE appointments
  ADD COLUMN IF NOT EXISTS package_id UUID NULL REFERENCES appointment_packages(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS package_session_number INT NULL;

CREATE INDEX IF NOT EXISTS ix_appointments_package
  ON appointments(package_id, package_session_number);

CREATE INDEX IF NOT EXISTS ix_appointment_packages_patient
  ON appointment_packages(patient_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS ix_appointment_packages_patient;
DROP INDEX IF EXISTS ix_appointments_package;

ALTER TABLE appointments
  DROP COLUMN IF EXISTS package_session_number,
  DROP COLUMN IF EXISTS package_id;

DROP TABLE IF EXISTS appointment_packages;
