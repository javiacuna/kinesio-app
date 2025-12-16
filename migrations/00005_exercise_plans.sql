-- +goose Up
CREATE TABLE IF NOT EXISTS exercise_plans (
  id UUID PRIMARY KEY,
  patient_id UUID NOT NULL REFERENCES patients(id),
  kinesiologist_id UUID NOT NULL REFERENCES kinesiologists(id),
  frequency TEXT NOT NULL,              -- "daily" | "weekly"
  duration_weeks INT NOT NULL DEFAULT 1,
  observations TEXT NULL,
  status TEXT NOT NULL DEFAULT 'active', -- "active" | "closed"
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_exercise_plans_patient ON exercise_plans(patient_id);
CREATE INDEX IF NOT EXISTS idx_exercise_plans_kinesiologist ON exercise_plans(kinesiologist_id);

CREATE TABLE IF NOT EXISTS exercise_plan_items (
  id UUID PRIMARY KEY,
  plan_id UUID NOT NULL REFERENCES exercise_plans(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  description TEXT NULL,
  video_url TEXT NULL,
  guide_url TEXT NULL,
  estimated_minutes INT NOT NULL DEFAULT 10,
  sets INT NULL,
  reps INT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_exercise_plan_items_plan ON exercise_plan_items(plan_id);

-- +goose Down
DROP TABLE IF EXISTS exercise_plan_items;
DROP TABLE IF EXISTS exercise_plans;
