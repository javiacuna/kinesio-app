-- +goose Up
CREATE TABLE IF NOT EXISTS appointments (
  id UUID PRIMARY KEY,
  patient_id UUID NOT NULL,
  kinesiologist_id UUID NOT NULL,
  start_at TIMESTAMPTZ NOT NULL,
  end_at TIMESTAMPTZ NOT NULL,
  status TEXT NOT NULL DEFAULT 'scheduled', -- scheduled | cancelled
  notes TEXT NULL,
  cancelled_reason TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Índices útiles para agenda y solapamientos
CREATE INDEX IF NOT EXISTS ix_appointments_kine_start ON appointments (kinesiologist_id, start_at);
CREATE INDEX IF NOT EXISTS ix_appointments_patient_start ON appointments (patient_id, start_at);

-- Validación básica de integridad (end > start)
-- (en Postgres se puede con CHECK; lo agrego)
ALTER TABLE appointments
  ADD CONSTRAINT ck_appointments_time CHECK (end_at > start_at);

-- +goose Down
DROP TABLE IF EXISTS appointments;
