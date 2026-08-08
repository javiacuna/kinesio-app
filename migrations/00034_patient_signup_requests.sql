-- +goose Up
CREATE TABLE IF NOT EXISTS patient_signup_requests (
  id UUID PRIMARY KEY,
  firebase_uid TEXT NOT NULL,
  dni TEXT NOT NULL,
  email TEXT NOT NULL,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
  matched_patient_id UUID NULL REFERENCES patients(id),
  reviewed_by_email TEXT NULL,
  reviewed_at TIMESTAMPTZ NULL,
  rejection_reason TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_patient_signup_requests_firebase_uid
  ON patient_signup_requests (firebase_uid);

CREATE INDEX IF NOT EXISTS idx_patient_signup_requests_status_created
  ON patient_signup_requests (status, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_patient_signup_requests_status_created;
DROP INDEX IF EXISTS ux_patient_signup_requests_firebase_uid;
DROP TABLE IF EXISTS patient_signup_requests;
