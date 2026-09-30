-- A relaunch grants a fresh bounded retry budget without renumbering history.
ALTER TABLE processing_jobs ADD COLUMN attempt_budget integer;
UPDATE processing_jobs SET attempt_budget=max_attempts;
ALTER TABLE processing_jobs ALTER COLUMN attempt_budget SET NOT NULL;
ALTER TABLE processing_jobs ADD CHECK (attempt_budget BETWEEN 1 AND 100);
ALTER TABLE processing_jobs DROP CONSTRAINT processing_jobs_max_attempts_check;
ALTER TABLE processing_jobs ADD CHECK (max_attempts > 0);
CREATE TABLE processing_job_relaunches (
 job_id bigint NOT NULL REFERENCES processing_jobs(id),
 after_attempt integer NOT NULL CHECK (after_attempt > 0),
 actor text NOT NULL CHECK (btrim(actor) <> ''),
 created_at timestamptz NOT NULL,
 PRIMARY KEY(job_id,after_attempt)
);
