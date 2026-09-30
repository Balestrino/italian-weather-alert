CREATE TABLE interpretation_input_manifests (
 run_id bigint PRIMARY KEY REFERENCES processing_runs(id),
 document_id bigint NOT NULL REFERENCES retained_documents(id),
 source_id text NOT NULL REFERENCES registry_sources(id),
 configuration text NOT NULL REFERENCES processing_configuration_versions(id),
 manifest_sha256 text NOT NULL CHECK(manifest_sha256 ~ '^[0-9a-f]{64}$'),
 body jsonb NOT NULL,
 reusable boolean NOT NULL
);
CREATE INDEX interpretation_manifest_match ON interpretation_input_manifests(document_id,source_id,configuration,manifest_sha256) WHERE reusable;
CREATE TRIGGER interpretation_manifest_immutable BEFORE UPDATE OR DELETE ON interpretation_input_manifests FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
