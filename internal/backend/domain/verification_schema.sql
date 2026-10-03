CREATE TABLE domain_verification_receipts (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 request_id text NOT NULL UNIQUE CHECK (request_id <> ''),
 request_sha256 text NOT NULL CHECK (request_sha256 ~ '^[0-9a-f]{64}$'),
 candidate_key text NOT NULL CHECK (candidate_key <> ''),
 candidate_source_id text NOT NULL REFERENCES registry_sources(id),
 candidate_version_id bigint NOT NULL REFERENCES retained_versions(id) ON DELETE CASCADE,
 kind text NOT NULL CHECK(kind IN ('local_measure','operational_phase','regional_record')),
 municipality_istat text NOT NULL CHECK(municipality_istat ~ '^[0-9]{6}$'),
 verified_at timestamptz NOT NULL,
 body jsonb NOT NULL
);
CREATE INDEX verification_candidate_history ON domain_verification_receipts(candidate_source_id,candidate_key,id);

CREATE TABLE domain_verification_dependencies (
 receipt_id bigint NOT NULL REFERENCES domain_verification_receipts(id) ON DELETE CASCADE,
 supported_version_id bigint NOT NULL REFERENCES retained_versions(id) ON DELETE CASCADE,
 evidence_version_id bigint NOT NULL REFERENCES retained_versions(id) ON DELETE CASCADE,
 PRIMARY KEY(receipt_id,supported_version_id,evidence_version_id)
);
CREATE TRIGGER verification_receipts_immutable BEFORE UPDATE OR DELETE ON domain_verification_receipts FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE TRIGGER verification_dependencies_immutable BEFORE UPDATE OR DELETE ON domain_verification_dependencies FOR EACH ROW EXECUTE FUNCTION domain_immutable();

CREATE TABLE domain_verification_projections (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 kind text NOT NULL CHECK(kind IN ('local_measure','operational_phase','regional_record')),
 source_id text NOT NULL REFERENCES registry_sources(id),
 logical_key text NOT NULL,
 document_version_id bigint NOT NULL REFERENCES retained_versions(id) ON DELETE CASCADE,
 domain_record_id text NOT NULL,
 recorded_at timestamptz NOT NULL,
 acquisition_at timestamptz NOT NULL,
 UNIQUE(kind,domain_record_id,acquisition_at)
);
CREATE INDEX verification_projection_history ON domain_verification_projections(kind,source_id,logical_key,acquisition_at DESC,recorded_at DESC,id DESC);
CREATE TRIGGER verification_projections_immutable BEFORE UPDATE OR DELETE ON domain_verification_projections FOR EACH ROW EXECUTE FUNCTION domain_immutable();

-- Supersede only a known revision of the same explicitly scoped primary.
-- Unmapped records keep their existing selection semantics.
CREATE FUNCTION domain_verification_record_current(record_kind text, record_id text, known_at timestamptz)
RETURNS boolean LANGUAGE sql STABLE AS $$
 SELECT COALESCE((
  SELECT (
   SELECT latest.domain_record_id=record_id
   FROM domain_verification_projections latest
   WHERE latest.kind=scope.kind AND latest.source_id=scope.source_id AND latest.logical_key=scope.logical_key
    AND latest.recorded_at<=known_at
   ORDER BY latest.acquisition_at DESC,latest.recorded_at DESC,latest.id DESC LIMIT 1
  )
  FROM domain_verification_projections scope
  WHERE scope.kind=record_kind AND scope.domain_record_id=record_id AND scope.recorded_at<=known_at
  LIMIT 1
 ), true)
$$;
