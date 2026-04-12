-- +goose Up
CREATE TABLE IF NOT EXISTS patient_evolution_photos (
  id UUID PRIMARY KEY,
  evolution_id UUID NOT NULL REFERENCES patient_evolutions(id) ON DELETE CASCADE,
  url TEXT NOT NULL,
  caption TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_patient_evolution_photos_evolution
  ON patient_evolution_photos(evolution_id);

-- +goose Down
DROP TABLE IF EXISTS patient_evolution_photos;
