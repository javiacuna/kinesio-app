-- +goose Up
CREATE TABLE IF NOT EXISTS staff_members (
  id UUID PRIMARY KEY,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  email TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('admin', 'recepcionista', 'kinesiologo')),
  phone TEXT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  firebase_uid TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_staff_members_email
  ON staff_members (lower(email));

-- +goose Down
DROP TABLE IF EXISTS staff_members;
