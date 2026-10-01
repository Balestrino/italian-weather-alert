-- Private diagnostic evidence, never a successful classification or checkpoint.
CREATE TABLE processing_invalid_outputs (
 run_id bigint NOT NULL,
 attempt_number integer NOT NULL,
 segment_ordinal integer NOT NULL CHECK (segment_ordinal > 0),
 input_sha256 text NOT NULL CHECK (input_sha256 ~ '^[0-9a-f]{64}$'),
 response_sha256 text NOT NULL CHECK (response_sha256 ~ '^[0-9a-f]{64}$'),
 request_bytes bytea NOT NULL CHECK (octet_length(request_bytes) BETWEEN 1 AND 16777216),
 response_bytes bytea NOT NULL CHECK (octet_length(response_bytes) <= 8388608),
 finish_reason text NOT NULL,
 error_code text NOT NULL,
 cached boolean NOT NULL,
 created_at timestamptz NOT NULL,
 PRIMARY KEY(run_id,attempt_number,segment_ordinal),
 FOREIGN KEY(run_id,attempt_number) REFERENCES processing_run_attempts(run_id,number) ON DELETE CASCADE
);
CREATE TRIGGER processing_invalid_outputs_no_update BEFORE UPDATE ON processing_invalid_outputs
 FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
