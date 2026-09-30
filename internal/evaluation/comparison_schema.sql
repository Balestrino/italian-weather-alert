CREATE TABLE processing_change_evaluations (
 id text PRIMARY KEY CHECK (id <> ''),
 corpus_sha256 text NOT NULL CHECK (corpus_sha256 ~ '^[0-9a-f]{64}$'),
 baseline_process_version text NOT NULL CHECK (baseline_process_version ~ '^[0-9a-f]{64}$'),
 candidate_process_version text NOT NULL CHECK (candidate_process_version ~ '^[0-9a-f]{64}$'),
 passed boolean NOT NULL,
 actor text NOT NULL CHECK (actor <> ''),
 report jsonb NOT NULL CHECK (jsonb_typeof(report) = 'object'),
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER processing_change_evaluations_immutable BEFORE UPDATE OR DELETE ON processing_change_evaluations FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
