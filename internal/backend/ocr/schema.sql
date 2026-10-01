CREATE TABLE ocr_resource_results (
 run_id bigint PRIMARY KEY REFERENCES processing_runs(id),
 document_version_id bigint NOT NULL,
 resource_url text NOT NULL,
 status text NOT NULL CHECK (status IN ('complete','partial_unreadable','unreadable','missing')),
 page_count integer NOT NULL CHECK (page_count >= 0),
 error_code text,
 created_at timestamptz NOT NULL,
 FOREIGN KEY(document_version_id,resource_url) REFERENCES retained_resources(version_id,url),
 CHECK ((status IN ('unreadable','missing')) = (error_code IS NOT NULL)),
 CHECK (status <> 'missing' OR page_count = 0)
);

CREATE TABLE ocr_page_results (
 run_id bigint NOT NULL REFERENCES processing_runs(id),
 page_number integer NOT NULL CHECK (page_number > 0),
 document_version_id bigint NOT NULL,
 resource_url text NOT NULL,
 status text NOT NULL CHECK (status IN ('complete','unreadable')),
 media_type text NOT NULL CHECK (media_type IN ('image/png','image/jpeg')),
 input_sha256 text NOT NULL CHECK (input_sha256 ~ '^[0-9a-f]{64}$'),
 output_sha256 text NOT NULL CHECK (output_sha256 ~ '^[0-9a-f]{64}$'),
 extracted_text text NOT NULL,
 provider_response_id text NOT NULL,
 returned_model text NOT NULL,
 input_tokens bigint CHECK (input_tokens >= 0),
 output_tokens bigint CHECK (output_tokens >= 0),
 cache_read_tokens bigint CHECK (cache_read_tokens >= 0),
 created_at timestamptz NOT NULL,
 PRIMARY KEY(run_id,page_number),
 FOREIGN KEY(document_version_id,resource_url) REFERENCES retained_resources(version_id,url),
 CHECK ((status = 'complete') = (extracted_text <> ''))
);
CREATE INDEX ocr_pages_evidence ON ocr_page_results(document_version_id,resource_url,page_number,created_at);

CREATE FUNCTION ocr_results_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'OCR evidence is append-only'; END;
$$;
CREATE TRIGGER ocr_resources_immutable BEFORE UPDATE OR DELETE ON ocr_resource_results FOR EACH ROW EXECUTE FUNCTION ocr_results_immutable();
CREATE TRIGGER ocr_pages_immutable BEFORE UPDATE OR DELETE ON ocr_page_results FOR EACH ROW EXECUTE FUNCTION ocr_results_immutable();
