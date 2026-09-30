CREATE TABLE extraction_results (
 run_id bigint PRIMARY KEY REFERENCES processing_runs(id),
 document_version_id bigint NOT NULL REFERENCES retained_versions(id),
 classification_run_id bigint NOT NULL REFERENCES classification_results(run_id),
 status text NOT NULL CHECK (status IN ('extracted','not_applicable','uninterpreted')),
 reason_code text NOT NULL CHECK (reason_code IN ('measures_extracted','not_relevant','classification_undetermined','incomplete_required_content','empty_content','extraction_output_schema_invalid','extraction_evidence_invalid','provider_error','attempts_exhausted')),
 content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
 content_complete boolean NOT NULL,
 provider_response_id text,
 returned_model text,
 input_tokens bigint CHECK (input_tokens >= 0),
 output_tokens bigint CHECK (output_tokens >= 0),
 cache_read_tokens bigint CHECK (cache_read_tokens >= 0),
 created_at timestamptz NOT NULL,
 CHECK ((status = 'extracted') = (reason_code = 'measures_extracted')),
 CHECK ((status = 'extracted') = (provider_response_id IS NOT NULL AND returned_model IS NOT NULL)),
 CHECK (status <> 'extracted' OR content_complete)
);
CREATE INDEX extraction_results_document ON extraction_results(document_version_id,created_at DESC,run_id DESC);

CREATE TABLE extracted_measures (
 run_id bigint NOT NULL REFERENCES extraction_results(run_id),
 ordinal integer NOT NULL CHECK (ordinal > 0),
 kind text NOT NULL CHECK (kind IN ('closure','reopening','restriction','prohibition','suspension','activation','deactivation','operational_update','observation')),
 subject text NOT NULL CHECK (subject <> ''),
 place text,
 valid_from_expression text,
 valid_until_expression text,
 indeterminate_fields jsonb NOT NULL CHECK (jsonb_typeof(indeterminate_fields) = 'array'),
 PRIMARY KEY(run_id,ordinal),
 CHECK ((place IS NULL) = (indeterminate_fields ? 'place')),
 CHECK ((valid_from_expression IS NULL) = (indeterminate_fields ? 'valid_from')),
 CHECK ((valid_until_expression IS NULL) = (indeterminate_fields ? 'valid_until'))
);

CREATE TABLE extraction_evidence (
 run_id bigint NOT NULL,
 measure_ordinal integer NOT NULL,
 ordinal integer NOT NULL CHECK (ordinal > 0),
 field_name text NOT NULL CHECK (field_name IN ('kind','subject','place','valid_from','valid_until')),
 resource_url text NOT NULL CHECK (resource_url <> ''),
 page_number integer CHECK (page_number > 0),
 quote text NOT NULL CHECK (quote <> ''),
 PRIMARY KEY(run_id,measure_ordinal,ordinal),
 FOREIGN KEY(run_id,measure_ordinal) REFERENCES extracted_measures(run_id,ordinal)
);

CREATE TRIGGER extraction_results_immutable BEFORE UPDATE OR DELETE ON extraction_results FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
CREATE TRIGGER extracted_measures_immutable BEFORE UPDATE OR DELETE ON extracted_measures FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
CREATE TRIGGER extraction_evidence_immutable BEFORE UPDATE OR DELETE ON extraction_evidence FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
