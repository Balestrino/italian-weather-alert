CREATE TABLE retention_policy_versions (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 months integer NOT NULL CHECK (months BETWEEN 1 AND 120),
 actor text NOT NULL CHECK (actor <> ''),
 created_at timestamptz NOT NULL
);
INSERT INTO retention_policy_versions(months,actor,created_at)
VALUES (3,'migration-default',clock_timestamp());

CREATE TABLE retention_version_protections (
 version_id bigint NOT NULL,
 scope text NOT NULL CHECK (scope <> ''),
 reference text NOT NULL CHECK (reference <> ''),
 protect_through timestamptz,
 created_at timestamptz NOT NULL,
 PRIMARY KEY(version_id,scope,reference)
);

CREATE TABLE retention_cleanup_runs (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 policy_version_id bigint NOT NULL REFERENCES retention_policy_versions(id),
 evaluated_at timestamptz NOT NULL,
 deleted_versions integer NOT NULL CHECK (deleted_versions >= 0),
 queued_objects integer NOT NULL CHECK (queued_objects >= 0),
 committed_at timestamptz NOT NULL
);

CREATE TABLE retention_pending_objects (
 hash text PRIMARY KEY CHECK(hash ~ '^[0-9a-f]{64}$'),
 object_key text NOT NULL UNIQUE,
 first_cleanup_run_id bigint NOT NULL REFERENCES retention_cleanup_runs(id),
 queued_at timestamptz NOT NULL
);

CREATE TABLE retention_object_deletion_attempts (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 cleanup_run_id bigint NOT NULL REFERENCES retention_cleanup_runs(id),
 hash text NOT NULL CHECK(hash ~ '^[0-9a-f]{64}$'),
 attempted_at timestamptz NOT NULL,
 outcome text NOT NULL CHECK(outcome IN ('deleted','failed')),
 error_code text,
 CHECK ((outcome='failed') = (error_code IS NOT NULL))
);

CREATE FUNCTION retention_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'retention history is append-only'; END;
$$;
CREATE TRIGGER retention_policies_immutable BEFORE UPDATE OR DELETE ON retention_policy_versions FOR EACH ROW EXECUTE FUNCTION retention_immutable();
CREATE TRIGGER retention_protections_immutable BEFORE UPDATE OR DELETE ON retention_version_protections FOR EACH ROW EXECUTE FUNCTION retention_immutable();
CREATE TRIGGER retention_runs_immutable BEFORE UPDATE OR DELETE ON retention_cleanup_runs FOR EACH ROW EXECUTE FUNCTION retention_immutable();
CREATE TRIGGER retention_attempts_immutable BEFORE UPDATE OR DELETE ON retention_object_deletion_attempts FOR EACH ROW EXECUTE FUNCTION retention_immutable();

-- Append-only evidence may only be removed by the retention transaction, which
-- sets this transaction-local flag while deleting the complete dependency graph.
CREATE OR REPLACE FUNCTION domain_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF current_setting('iwa.retention_cleanup',true)='on' THEN RETURN OLD; END IF;
 RAISE EXCEPTION 'domain facts are append-only';
END;
$$;
CREATE OR REPLACE FUNCTION processing_catalog_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF current_setting('iwa.retention_cleanup',true)='on' THEN RETURN OLD; END IF;
 RAISE EXCEPTION 'processing version history is append-only';
END;
$$;
CREATE OR REPLACE FUNCTION ocr_results_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF current_setting('iwa.retention_cleanup',true)='on' THEN RETURN OLD; END IF;
 RAISE EXCEPTION 'OCR evidence is append-only';
END;
$$;
