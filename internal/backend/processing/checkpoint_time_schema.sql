-- Older associations have no observed use time. Keep it unknown, not migration time.
ALTER TABLE processing_segment_checkpoint_uses ADD COLUMN created_at timestamptz;
ALTER TABLE processing_segment_checkpoint_uses ALTER COLUMN created_at SET DEFAULT clock_timestamp();
