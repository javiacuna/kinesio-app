-- +goose Up
CREATE TABLE IF NOT EXISTS kinesiologists (
  id UUID PRIMARY KEY,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  email TEXT NOT NULL,
  license_number TEXT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_kinesiologists_email
  ON kinesiologists (lower(email));

-- FK hacia appointments (NOT VALID para no romper datos previos)
ALTER TABLE appointments
  ADD CONSTRAINT fk_appointments_kinesiologist
  FOREIGN KEY (kinesiologist_id)
  REFERENCES kinesiologists(id)
  NOT VALID;

-- +goose Down
ALTER TABLE appointments
  DROP CONSTRAINT IF EXISTS fk_appointments_kinesiologist;

DROP TABLE IF EXISTS kinesiologists;
