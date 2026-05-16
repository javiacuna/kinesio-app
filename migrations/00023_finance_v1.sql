-- +goose Up
CREATE TABLE IF NOT EXISTS financiers (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  kind TEXT NOT NULL DEFAULT 'particular',
  active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO financiers (id, name, kind, active)
VALUES ('9ebea4c0-8944-4ef5-9647-6ad4357f3d14', 'Particular', 'particular', true)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS practice_tariffs (
  id UUID PRIMARY KEY,
  practice_id UUID NOT NULL REFERENCES practices(id),
  financier_id UUID NOT NULL REFERENCES financiers(id),
  billing_value_cents BIGINT NOT NULL,
  copay_cents BIGINT NOT NULL DEFAULT 0,
  currency TEXT NOT NULL DEFAULT 'ARS',
  valid_from DATE NOT NULL DEFAULT CURRENT_DATE,
  valid_to DATE NULL,
  active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT ck_practice_tariffs_amounts CHECK (billing_value_cents >= 0 AND copay_cents >= 0),
  CONSTRAINT ck_practice_tariffs_dates CHECK (valid_to IS NULL OR valid_to >= valid_from)
);

CREATE INDEX IF NOT EXISTS ix_practice_tariffs_lookup
  ON practice_tariffs(practice_id, financier_id, active, valid_from DESC);

CREATE TABLE IF NOT EXISTS professional_fee_rules (
  id UUID PRIMARY KEY,
  kinesiologist_id UUID NOT NULL REFERENCES kinesiologists(id),
  practice_id UUID NOT NULL REFERENCES practices(id),
  rule_type TEXT NOT NULL,
  fixed_value_cents BIGINT NULL,
  percentage NUMERIC(5,2) NULL,
  active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT ck_professional_fee_rules_type CHECK (rule_type IN ('fixed', 'percentage')),
  CONSTRAINT ck_professional_fee_rules_values CHECK (
    (rule_type = 'fixed' AND fixed_value_cents IS NOT NULL AND fixed_value_cents >= 0 AND percentage IS NULL)
    OR
    (rule_type = 'percentage' AND percentage IS NOT NULL AND percentage >= 0 AND percentage <= 100 AND fixed_value_cents IS NULL)
  )
);

CREATE INDEX IF NOT EXISTS ix_professional_fee_rules_lookup
  ON professional_fee_rules(kinesiologist_id, practice_id, active);

CREATE TABLE IF NOT EXISTS financial_movements (
  id UUID PRIMARY KEY,
  appointment_id UUID NOT NULL UNIQUE REFERENCES appointments(id),
  patient_id UUID NOT NULL REFERENCES patients(id),
  kinesiologist_id UUID NOT NULL REFERENCES kinesiologists(id),
  practice_id UUID NOT NULL REFERENCES practices(id),
  financier_id UUID NOT NULL REFERENCES financiers(id),
  tariff_id UUID NOT NULL REFERENCES practice_tariffs(id),
  fee_rule_id UUID NULL REFERENCES professional_fee_rules(id),
  billing_value_cents BIGINT NOT NULL,
  copay_cents BIGINT NOT NULL,
  payer_value_cents BIGINT NOT NULL,
  professional_fee_cents BIGINT NOT NULL,
  center_amount_cents BIGINT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT ck_financial_movements_amounts CHECK (
    billing_value_cents >= 0
    AND copay_cents >= 0
    AND payer_value_cents >= 0
    AND professional_fee_cents >= 0
  )
);

CREATE INDEX IF NOT EXISTS ix_financial_movements_period
  ON financial_movements(created_at DESC, status);

ALTER TABLE appointments
  ADD COLUMN IF NOT EXISTS practice_id UUID NULL REFERENCES practices(id),
  ADD COLUMN IF NOT EXISTS financier_id UUID NULL REFERENCES financiers(id);

-- +goose Down
ALTER TABLE appointments
  DROP COLUMN IF EXISTS financier_id,
  DROP COLUMN IF EXISTS practice_id;

DROP INDEX IF EXISTS ix_financial_movements_period;
DROP TABLE IF EXISTS financial_movements;
DROP INDEX IF EXISTS ix_professional_fee_rules_lookup;
DROP TABLE IF EXISTS professional_fee_rules;
DROP INDEX IF EXISTS ix_practice_tariffs_lookup;
DROP TABLE IF EXISTS practice_tariffs;
DROP TABLE IF EXISTS financiers;
