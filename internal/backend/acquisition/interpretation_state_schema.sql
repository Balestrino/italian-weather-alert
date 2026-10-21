ALTER TABLE acquisition_targets DROP CONSTRAINT acquisition_targets_measure_state_check;
ALTER TABLE acquisition_targets ADD CONSTRAINT acquisition_targets_measure_state_check
 CHECK (measure_state IN ('none','ongoing','unresolved','aged_unresolved','ceased'));
ALTER TABLE acquisition_targets ADD COLUMN interpretation_state_at timestamptz;
