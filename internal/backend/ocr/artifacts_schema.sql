CREATE TABLE ocr_artifacts (
 artifact_key text PRIMARY KEY CHECK (artifact_key ~ '^[0-9a-f]{64}$'),
 identity jsonb NOT NULL,
 state text NOT NULL CHECK (state IN ('pending','ready')),
 owner_token text NOT NULL,
 generation bigint NOT NULL DEFAULT 1,
 lease_expires_at timestamptz NOT NULL,
 result jsonb,
 created_at timestamptz NOT NULL,
 completed_at timestamptz,
 CHECK ((state='ready') = (result IS NOT NULL AND completed_at IS NOT NULL))
);
CREATE FUNCTION ocr_artifact_ready_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF OLD.state='ready' THEN RAISE EXCEPTION 'completed OCR artifact is immutable'; END IF;
 RETURN NEW;
END;
$$;
CREATE TRIGGER ocr_artifact_ready_immutable BEFORE UPDATE ON ocr_artifacts FOR EACH ROW EXECUTE FUNCTION ocr_artifact_ready_immutable();

CREATE TABLE ocr_artifact_associations (
 run_id bigint NOT NULL,
 page_number integer NOT NULL,
 artifact_key text NOT NULL REFERENCES ocr_artifacts(artifact_key),
 reused boolean NOT NULL,
 created_at timestamptz NOT NULL,
 PRIMARY KEY(run_id,page_number),
 FOREIGN KEY(run_id,page_number) REFERENCES ocr_page_results(run_id,page_number)
);
CREATE INDEX ocr_artifact_associations_key ON ocr_artifact_associations(artifact_key);
CREATE TRIGGER ocr_artifact_associations_immutable BEFORE UPDATE OR DELETE ON ocr_artifact_associations FOR EACH ROW EXECUTE FUNCTION ocr_results_immutable();
