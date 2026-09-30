CREATE TABLE acquisition_unavailable_targets (
 source_id text NOT NULL,
 configuration integer NOT NULL,
 url text NOT NULL,
 http_status integer NOT NULL CHECK (http_status IN (404,410)),
 consecutive_failures integer NOT NULL CHECK (consecutive_failures >= 0),
 first_failed_at timestamptz NOT NULL,
 last_failed_at timestamptz NOT NULL,
 next_attempt_at timestamptz NOT NULL,
 recovered_at timestamptz,
 PRIMARY KEY(source_id,configuration,url),
 FOREIGN KEY(source_id,configuration) REFERENCES registry_configurations(source_id,revision),
 FOREIGN KEY(source_id,url) REFERENCES acquisition_targets(source_id,url)
);
