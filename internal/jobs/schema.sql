CREATE TABLE processing_jobs (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 queue text NOT NULL CHECK (queue <> ''),
 kind text NOT NULL CHECK (kind <> ''),
 idempotency_key text NOT NULL CHECK (idempotency_key <> ''),
 payload jsonb NOT NULL CHECK (jsonb_typeof(payload) = 'object'),
 payload_hash text NOT NULL CHECK (payload_hash ~ '^[0-9a-f]{64}$'),
 state text NOT NULL CHECK (state IN ('queued','running','retry_wait','succeeded','failed')),
 max_attempts integer NOT NULL CHECK (max_attempts BETWEEN 1 AND 100),
 retry_base_ms bigint NOT NULL CHECK (retry_base_ms BETWEEN 1 AND 86400000),
 attempt_count integer NOT NULL DEFAULT 0 CHECK (attempt_count BETWEEN 0 AND max_attempts),
 available_at timestamptz NOT NULL,
 claimed_by text,
 claim_token text,
 lease_expires_at timestamptz,
 result jsonb,
 last_error_code text,
 last_error_detail text,
 created_at timestamptz NOT NULL,
 updated_at timestamptz NOT NULL,
 completed_at timestamptz,
 UNIQUE(queue,kind,idempotency_key),
 CHECK (
   (state = 'running' AND claimed_by IS NOT NULL AND claim_token IS NOT NULL AND lease_expires_at IS NOT NULL AND completed_at IS NULL)
   OR
   (state <> 'running' AND claimed_by IS NULL AND claim_token IS NULL AND lease_expires_at IS NULL)
 ),
 CHECK ((state IN ('succeeded','failed')) = (completed_at IS NOT NULL)),
 CHECK (result IS NULL OR jsonb_typeof(result) = 'object')
);
CREATE INDEX processing_jobs_claim
 ON processing_jobs(queue,available_at,id)
 WHERE state IN ('queued','retry_wait','running');
CREATE INDEX processing_jobs_expired_lease
 ON processing_jobs(queue,lease_expires_at,id)
 WHERE state = 'running';

CREATE TABLE processing_attempts (
 job_id bigint NOT NULL REFERENCES processing_jobs(id),
 number integer NOT NULL CHECK (number > 0),
 worker_id text NOT NULL CHECK (worker_id <> ''),
 claim_token text NOT NULL UNIQUE CHECK (claim_token <> ''),
 started_at timestamptz NOT NULL,
 lease_expires_at timestamptz NOT NULL,
 finished_at timestamptz,
 outcome text NOT NULL CHECK (outcome IN ('running','succeeded','retry','failed','abandoned')),
 error_code text,
 error_detail text,
 PRIMARY KEY(job_id,number),
 CHECK ((outcome = 'running') = (finished_at IS NULL))
);

-- This transactional outbox is the idempotency boundary for later public
-- projections. Consumers must use effect_key as their own idempotency key.
CREATE TABLE processing_effects (
 effect_key text PRIMARY KEY CHECK (effect_key <> ''),
 job_id bigint NOT NULL REFERENCES processing_jobs(id),
 kind text NOT NULL CHECK (kind <> ''),
 payload jsonb NOT NULL CHECK (jsonb_typeof(payload) = 'object'),
 payload_hash text NOT NULL CHECK (payload_hash ~ '^[0-9a-f]{64}$'),
 created_at timestamptz NOT NULL
);
CREATE INDEX processing_effects_job ON processing_effects(job_id,effect_key);
