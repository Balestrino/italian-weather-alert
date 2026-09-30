CREATE TABLE processing_segment_checkpoints (
 stage text NOT NULL CHECK(stage IN('classification','extraction')),
 configuration text NOT NULL REFERENCES processing_configuration_versions(id),
 input_sha256 text NOT NULL CHECK(input_sha256 ~ '^[0-9a-f]{64}$'),
 call_id bigint NOT NULL REFERENCES processing_provider_calls(id),
 response jsonb NOT NULL,
 created_at timestamptz NOT NULL,
 PRIMARY KEY(stage,configuration,input_sha256)
);
CREATE TABLE processing_segment_checkpoint_uses (
 run_id bigint NOT NULL,
 attempt_number integer NOT NULL,
 segment_ordinal integer NOT NULL CHECK(segment_ordinal>0),
 stage text NOT NULL,
 configuration text NOT NULL,
 input_sha256 text NOT NULL,
 reused boolean NOT NULL,
 PRIMARY KEY(run_id,attempt_number,segment_ordinal),
 FOREIGN KEY(run_id,attempt_number) REFERENCES processing_run_attempts(run_id,number),
 FOREIGN KEY(stage,configuration,input_sha256) REFERENCES processing_segment_checkpoints(stage,configuration,input_sha256)
);
CREATE TRIGGER processing_checkpoint_immutable BEFORE UPDATE OR DELETE ON processing_segment_checkpoints FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
CREATE TRIGGER processing_checkpoint_use_immutable BEFORE UPDATE OR DELETE ON processing_segment_checkpoint_uses FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
