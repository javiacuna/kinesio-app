-- +goose Up
CREATE TABLE IF NOT EXISTS patients (
  id UUID PRIMARY KEY,
  dni TEXT NOT NULL,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  email TEXT NOT NULL,
  phone TEXT NULL,
  birth_date DATE NULL,
  clinical_notes TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_patients_dni ON patients (dni);
CREATE UNIQUE INDEX IF NOT EXISTS ux_patients_email ON patients (lower(email));

-- +goose Down
DROP TABLE IF EXISTS patients;
