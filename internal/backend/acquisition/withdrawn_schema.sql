ALTER TABLE acquisition_checks
 ADD COLUMN withdrawn_unavailable_count integer NOT NULL DEFAULT 0
 CHECK (withdrawn_unavailable_count >= 0);
ALTER TABLE acquisition_target_dispositions
 ADD COLUMN measure_acknowledged boolean NOT NULL DEFAULT false;
