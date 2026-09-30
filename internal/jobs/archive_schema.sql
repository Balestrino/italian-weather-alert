ALTER TABLE processing_jobs ADD COLUMN archived_at timestamptz;
ALTER TABLE processing_jobs ADD COLUMN archived_by text;
ALTER TABLE processing_jobs ADD CONSTRAINT processing_jobs_archive_valid CHECK (
 (archived_at IS NULL AND archived_by IS NULL) OR
 (archived_at IS NOT NULL AND archived_by IS NOT NULL AND length(btrim(archived_by))>0
  AND state IN ('queued','retry_wait','failed'))
);
CREATE FUNCTION processing_archived_job_frozen() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF OLD.archived_at IS NOT NULL AND NEW IS DISTINCT FROM OLD THEN
  RAISE EXCEPTION 'archived jobs cannot be modified';
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER processing_archived_job_frozen BEFORE UPDATE ON processing_jobs
 FOR EACH ROW EXECUTE FUNCTION processing_archived_job_frozen();
