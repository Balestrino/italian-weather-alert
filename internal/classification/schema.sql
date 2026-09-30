CREATE TABLE classification_results (
 run_id bigint PRIMARY KEY REFERENCES processing_runs(id),
 document_version_id bigint NOT NULL REFERENCES retained_versions(id),
 status text NOT NULL CHECK (status IN ('classified','undetermined')),
 relevant boolean,
 reason_code text NOT NULL CHECK (reason_code IN ('regional_warning','local_weather_measure','weather_operational_update','not_relevant','incomplete_required_content','empty_content')),
 evidence_quote text NOT NULL,
 content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
 content_complete boolean NOT NULL,
 provider_response_id text,
 returned_model text,
 input_tokens bigint CHECK (input_tokens >= 0),
 output_tokens bigint CHECK (output_tokens >= 0),
 cache_read_tokens bigint CHECK (cache_read_tokens >= 0),
 created_at timestamptz NOT NULL,
 CHECK ((status = 'classified') = (relevant IS NOT NULL)),
 CHECK ((status = 'classified') = (provider_response_id IS NOT NULL AND returned_model IS NOT NULL)),
 CHECK ((status = 'classified') = (evidence_quote <> '')),
 CHECK (status <> 'classified' OR content_complete),
 CHECK ((relevant IS FALSE) = (reason_code = 'not_relevant')),
 CHECK (status <> 'undetermined' OR reason_code IN ('incomplete_required_content','empty_content'))
);
CREATE INDEX classification_results_document ON classification_results(document_version_id,created_at);

CREATE TRIGGER classification_results_immutable BEFORE UPDATE OR DELETE ON classification_results FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
