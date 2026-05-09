-- +goose Up
CREATE TABLE IF NOT EXISTS specialties (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS practices (
  id UUID PRIMARY KEY,
  specialty_id UUID NOT NULL REFERENCES specialties(id),
  name TEXT NOT NULL,
  description TEXT NULL,
  active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT ux_practices_specialty_name UNIQUE (specialty_id, name)
);

CREATE TABLE IF NOT EXISTS kinesiologist_practices (
  kinesiologist_id UUID NOT NULL REFERENCES kinesiologists(id) ON DELETE CASCADE,
  practice_id UUID NOT NULL REFERENCES practices(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (kinesiologist_id, practice_id)
);

INSERT INTO specialties (id, name, active)
VALUES ('d81c0796-03ef-41df-89d8-266c9a34698c', 'Kinesiología', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO practices (id, specialty_id, name, active)
VALUES
  ('e0619b43-d9d8-4bdd-9dc0-d9c4876e3bb4', 'd81c0796-03ef-41df-89d8-266c9a34698c', 'Kinesiología motora', true),
  ('b4c0a111-7f2e-4ee5-bd9d-7c3c2199b9b5', 'd81c0796-03ef-41df-89d8-266c9a34698c', 'Kinesiología respiratoria', true),
  ('9f857a61-59ad-4d6a-821e-529a3449f586', 'd81c0796-03ef-41df-89d8-266c9a34698c', 'Kinesiología pediátrica', true)
ON CONFLICT (id) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_practices_specialty
  ON practices(specialty_id, active, name);

CREATE INDEX IF NOT EXISTS idx_kinesiologist_practices_practice
  ON kinesiologist_practices(practice_id);

-- +goose Down
DROP TABLE IF EXISTS kinesiologist_practices;
DROP TABLE IF EXISTS practices;
DROP TABLE IF EXISTS specialties;
