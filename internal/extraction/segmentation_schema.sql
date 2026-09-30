CREATE TABLE extraction_segments (
 run_id bigint NOT NULL REFERENCES extraction_results(run_id),
 ordinal integer NOT NULL CHECK (ordinal > 0),
 total integer NOT NULL CHECK (total >= ordinal),
 document_version_id bigint NOT NULL REFERENCES retained_versions(id),
 resource_url text NOT NULL CHECK (resource_url <> ''),
 role text NOT NULL CHECK (role <> ''),
 page_number integer CHECK (page_number > 0),
 start_byte integer NOT NULL CHECK (start_byte >= 0),
 end_byte integer NOT NULL CHECK (end_byte > start_byte),
 content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
 response_sha256 text NOT NULL CHECK (response_sha256 ~ '^[0-9a-f]{64}$'),
 provider_response_id text NOT NULL CHECK (provider_response_id <> ''),
 returned_model text NOT NULL CHECK (returned_model <> ''),
 measure_count integer NOT NULL CHECK (measure_count >= 0),
 input_tokens bigint CHECK (input_tokens >= 0),
 output_tokens bigint CHECK (output_tokens >= 0),
 cache_read_tokens bigint CHECK (cache_read_tokens >= 0),
 PRIMARY KEY(run_id,ordinal)
);
CREATE INDEX extraction_segments_document ON extraction_segments(document_version_id,run_id,ordinal);
ALTER TABLE extraction_results DROP CONSTRAINT extraction_results_reason_code_check;
ALTER TABLE extraction_results ADD CONSTRAINT extraction_results_reason_code_check CHECK (reason_code IN ('measures_extracted','not_relevant','classification_undetermined','incomplete_required_content','empty_content','extraction_output_schema_invalid','extraction_evidence_invalid','extraction_merge_duplicate','extraction_merge_conflict','provider_error','attempts_exhausted'));
ALTER TABLE extraction_evidence ADD COLUMN segment_ordinal integer;
ALTER TABLE extraction_evidence ADD CONSTRAINT extraction_evidence_segment
 FOREIGN KEY(run_id,segment_ordinal) REFERENCES extraction_segments(run_id,ordinal);
CREATE TRIGGER extraction_segments_immutable BEFORE UPDATE OR DELETE ON extraction_segments FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
