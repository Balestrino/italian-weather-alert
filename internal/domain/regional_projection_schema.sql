CREATE TABLE domain_regional_projections (
 document_version_id bigint NOT NULL REFERENCES retained_versions(id),
 logic_version text NOT NULL,
 source_id text NOT NULL REFERENCES registry_sources(id),
 product text NOT NULL REFERENCES domain_regional_products(id),
 status text NOT NULL CHECK(status IN ('partial','unsupported','no_event')),
 statement text NOT NULL,
 limitations jsonb NOT NULL CHECK(jsonb_typeof(limitations)='array'),
 projected_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 PRIMARY KEY(document_version_id,logic_version)
);
ALTER TABLE domain_regional_records ADD COLUMN projection_logic text;
ALTER TABLE domain_regional_records ADD COLUMN evidence_locator text;
CREATE TRIGGER domain_regional_projections_immutable BEFORE UPDATE OR DELETE ON domain_regional_projections FOR EACH ROW EXECUTE FUNCTION domain_immutable();
