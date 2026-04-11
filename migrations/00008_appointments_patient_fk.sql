-- +goose Up
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_appointments_patient'
  ) THEN
    ALTER TABLE appointments
      ADD CONSTRAINT fk_appointments_patient
      FOREIGN KEY (patient_id)
      REFERENCES patients(id)
      NOT VALID;
  END IF;
END $$;

-- +goose Down
ALTER TABLE appointments
  DROP CONSTRAINT IF EXISTS fk_appointments_patient;
