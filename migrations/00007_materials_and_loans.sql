-- +goose Up
CREATE TABLE IF NOT EXISTS materials (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NULL,
  total_qty INT NOT NULL DEFAULT 0,
  available_qty INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_materials_name ON materials (lower(name));

CREATE TABLE IF NOT EXISTS material_loans (
  id UUID PRIMARY KEY,
  material_id UUID NOT NULL REFERENCES materials(id),
  patient_id UUID NOT NULL REFERENCES patients(id),
  kinesiologist_id UUID NOT NULL REFERENCES kinesiologists(id),
  qty INT NOT NULL DEFAULT 1,
  notes TEXT NULL,
  loaned_at TIMESTAMPTZ NOT NULL,
  returned_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_material_loans_patient ON material_loans(patient_id);
CREATE INDEX IF NOT EXISTS idx_material_loans_material ON material_loans(material_id);
CREATE INDEX IF NOT EXISTS idx_material_loans_returned ON material_loans(returned_at);

-- +goose Down
DROP TABLE IF EXISTS material_loans;
DROP TABLE IF EXISTS materials;
