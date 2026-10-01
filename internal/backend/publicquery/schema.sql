ALTER TABLE domain_regional_records ADD COLUMN recorded_at timestamptz NOT NULL DEFAULT clock_timestamp();
ALTER TABLE domain_local_measures ADD COLUMN recorded_at timestamptz NOT NULL DEFAULT clock_timestamp();
ALTER TABLE domain_operational_phases ADD COLUMN recorded_at timestamptz NOT NULL DEFAULT clock_timestamp();

CREATE INDEX domain_regional_public_time ON domain_regional_records(recorded_at,id);
CREATE INDEX domain_local_public_time ON domain_local_measures(recorded_at,id);
CREATE INDEX domain_phase_public_time ON domain_operational_phases(recorded_at,id);
