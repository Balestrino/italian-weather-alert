CREATE TABLE domain_provenance_events (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 source_id text NOT NULL,
 configuration integer NOT NULL,
 state text NOT NULL CHECK (state IN ('verified','unresolved')),
 evidence_url text NOT NULL CHECK (evidence_url <> ''),
 evidence_locator text NOT NULL CHECK (evidence_locator <> ''),
 limitations jsonb NOT NULL CHECK (jsonb_typeof(limitations)='array'),
 assessed_at timestamptz NOT NULL,
 FOREIGN KEY(source_id,configuration) REFERENCES registry_configurations(source_id,revision)
);
CREATE INDEX domain_provenance_latest ON domain_provenance_events(source_id,id DESC);

CREATE TABLE domain_interpretation_events (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 local_measure_id text NOT NULL REFERENCES domain_local_measures(id),
 state text NOT NULL CHECK (state IN ('supported','partial','failed','unreliable','not_processed')),
 evidence_document_version_id bigint NOT NULL REFERENCES retained_versions(id),
 reason text NOT NULL CHECK (reason <> ''),
 limitations jsonb NOT NULL CHECK (jsonb_typeof(limitations)='array'),
 actor text NOT NULL CHECK (actor <> ''),
 recorded_at timestamptz NOT NULL
);
CREATE INDEX domain_interpretation_latest ON domain_interpretation_events(local_measure_id,id DESC);

CREATE TABLE domain_preceding_state_warnings (
 local_measure_id text NOT NULL REFERENCES domain_local_measures(id),
 newer_document_version_id bigint NOT NULL REFERENCES retained_versions(id),
 reason text NOT NULL CHECK (reason <> ''),
 recorded_at timestamptz NOT NULL,
 PRIMARY KEY(local_measure_id,newer_document_version_id)
);

CREATE TRIGGER domain_provenance_events_immutable BEFORE UPDATE OR DELETE ON domain_provenance_events FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE TRIGGER domain_interpretation_events_immutable BEFORE UPDATE OR DELETE ON domain_interpretation_events FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE TRIGGER domain_preceding_state_warnings_immutable BEFORE UPDATE OR DELETE ON domain_preceding_state_warnings FOR EACH ROW EXECUTE FUNCTION domain_immutable();
