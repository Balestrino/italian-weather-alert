CREATE TABLE interpretation_reuse (
 run_id bigint PRIMARY KEY REFERENCES interpretation_input_manifests(run_id),
 original_run_id bigint NOT NULL REFERENCES interpretation_input_manifests(run_id),
 created_at timestamptz NOT NULL,
 CHECK(original_run_id<run_id)
);
CREATE INDEX interpretation_reuse_origin ON interpretation_reuse(original_run_id);
CREATE TRIGGER interpretation_reuse_immutable BEFORE UPDATE OR DELETE ON interpretation_reuse FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
