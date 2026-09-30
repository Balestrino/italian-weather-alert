CREATE TABLE backup_schedule (
 singleton boolean PRIMARY KEY DEFAULT true CHECK(singleton),
 enabled boolean NOT NULL,
 interval_seconds integer NOT NULL,
 retention_days integer NOT NULL,
 observed_at timestamptz NOT NULL
);
CREATE TABLE backup_runs (
 id text PRIMARY KEY,
 started_at timestamptz NOT NULL,
 finished_at timestamptz,
 next_due_at timestamptz NOT NULL,
 state text NOT NULL CHECK(state IN ('running','succeeded','failed')),
 error_code text NOT NULL DEFAULT '',
 archive_sha256 text NOT NULL DEFAULT '',
 archive_bytes bigint NOT NULL DEFAULT 0,
 objects integer NOT NULL DEFAULT 0,
 notified boolean NOT NULL DEFAULT false,
 CHECK((state='running')=(finished_at IS NULL))
);
