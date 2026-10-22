CREATE TABLE interpretation_input_equivalences (
 run_id bigint PRIMARY KEY REFERENCES interpretation_input_manifests(run_id),
 document_id bigint NOT NULL REFERENCES retained_documents(id),
 version_id bigint NOT NULL REFERENCES retained_versions(id),
 fingerprint text NOT NULL CHECK(fingerprint ~ '^[0-9a-f]{64}$'),
 fingerprint_version text NOT NULL CHECK(fingerprint_version<>''),
 created_at timestamptz NOT NULL
);
CREATE INDEX interpretation_input_equivalence_match ON interpretation_input_equivalences(document_id,fingerprint,version_id);
CREATE INDEX interpretation_input_equivalence_version ON interpretation_input_equivalences(version_id);
CREATE TRIGGER interpretation_input_equivalence_immutable BEFORE UPDATE OR DELETE ON interpretation_input_equivalences FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
