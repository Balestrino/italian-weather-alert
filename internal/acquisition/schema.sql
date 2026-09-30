CREATE TABLE acquisition_source_status (
 source_id text PRIMARY KEY REFERENCES registry_sources(id),
 configuration integer NOT NULL,
 check_seconds integer NOT NULL CHECK (check_seconds > 0),
 delay_seconds integer NOT NULL CHECK (delay_seconds > 0),
 backoff_base_seconds integer NOT NULL CHECK (backoff_base_seconds > 0),
 backoff_max_seconds integer NOT NULL CHECK (backoff_max_seconds >= backoff_base_seconds),
 next_check_at timestamptz NOT NULL,
 consecutive_failures integer NOT NULL DEFAULT 0 CHECK (consecutive_failures >= 0),
 last_started_at timestamptz,
 last_reachable_at timestamptz,
 last_content_at timestamptz,
 last_complete_at timestamptz,
 first_error_at timestamptz,
 last_error_at timestamptz,
 last_error_code text,
 retry_after timestamptz,
 claimed_by text,
 claim_token text,
 lease_expires_at timestamptz,
 FOREIGN KEY(source_id,configuration) REFERENCES registry_configurations(source_id,revision),
 CHECK ((claim_token IS NULL) = (claimed_by IS NULL)),
 CHECK ((claim_token IS NULL) = (lease_expires_at IS NULL)),
 CHECK ((last_error_code IS NULL) = (last_error_at IS NULL))
);
CREATE INDEX acquisition_source_due ON acquisition_source_status(next_check_at,source_id);
CREATE INDEX acquisition_source_lease ON acquisition_source_status(lease_expires_at,source_id) WHERE claim_token IS NOT NULL;

CREATE TABLE acquisition_checks (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 source_id text NOT NULL,
 configuration integer NOT NULL,
 worker_id text NOT NULL CHECK (worker_id <> ''),
 started_at timestamptz NOT NULL,
 finished_at timestamptz NOT NULL CHECK (finished_at >= started_at),
 reachable boolean NOT NULL,
 content_recognized boolean NOT NULL,
 complete boolean NOT NULL,
 error_code text,
 retry_after timestamptz,
 listing_count integer NOT NULL DEFAULT 0 CHECK (listing_count >= 0),
 document_count integer NOT NULL DEFAULT 0 CHECK (document_count >= 0),
 FOREIGN KEY(source_id,configuration) REFERENCES registry_configurations(source_id,revision),
 CHECK (NOT complete OR (reachable AND content_recognized AND error_code IS NULL)),
 CHECK (complete OR error_code IS NOT NULL)
);
CREATE INDEX acquisition_checks_source ON acquisition_checks(source_id,finished_at DESC,id DESC);
