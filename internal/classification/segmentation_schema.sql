CREATE TABLE classification_segments (
 run_id bigint NOT NULL REFERENCES classification_results(run_id),
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
 relevant boolean NOT NULL,
 reason_code text NOT NULL CHECK (reason_code IN ('regional_warning','local_weather_measure','weather_operational_update','not_relevant')),
 evidence_quote text NOT NULL CHECK (evidence_quote <> ''),
 provider_response_id text NOT NULL CHECK (provider_response_id <> ''),
 returned_model text NOT NULL CHECK (returned_model <> ''),
 input_tokens bigint CHECK (input_tokens >= 0),
 output_tokens bigint CHECK (output_tokens >= 0),
 cache_read_tokens bigint CHECK (cache_read_tokens >= 0),
 PRIMARY KEY(run_id,ordinal),
 CHECK ((relevant IS FALSE) = (reason_code = 'not_relevant'))
);
CREATE INDEX classification_segments_document ON classification_segments(document_version_id,run_id,ordinal);
CREATE TRIGGER classification_segments_immutable BEFORE UPDATE OR DELETE ON classification_segments FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
