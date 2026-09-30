CREATE TABLE interpretation_preflight (
 document_version_id bigint PRIMARY KEY REFERENCES retained_versions(id) ON DELETE CASCADE,
 representative_version_id bigint NOT NULL REFERENCES retained_versions(id) ON DELETE CASCADE,
 configuration text NOT NULL,
 fingerprint text NOT NULL CHECK(length(fingerprint)=64),
 body jsonb NOT NULL,
 complete boolean NOT NULL,
 created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 CHECK(representative_version_id<=document_version_id)
);
CREATE INDEX interpretation_preflight_representative ON interpretation_preflight(representative_version_id);
