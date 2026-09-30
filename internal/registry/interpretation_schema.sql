ALTER TABLE registry_sources ADD COLUMN interpretation_suspended_at timestamptz;
CREATE TABLE registry_interpretation_suspensions (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 source_id text NOT NULL REFERENCES registry_sources(id),
 revision integer NOT NULL,
 suspended_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 actor text NOT NULL,
 defect jsonb NOT NULL,
 resumed_at timestamptz,
 resumed_by text,
 recovery jsonb,
 FOREIGN KEY(source_id,revision) REFERENCES registry_configurations(source_id,revision),
 CHECK ((resumed_at IS NULL) = (recovery IS NULL))
);
CREATE UNIQUE INDEX registry_one_interpretation_suspension ON registry_interpretation_suspensions(source_id) WHERE resumed_at IS NULL;
-- Freeze interpreted facts, never acquisition metadata. Completed intervals also
-- preserve the interpretation boundary for historical knowledge queries.
CREATE FUNCTION registry_interpretation_cutoff(source text, known_at timestamptz) RETURNS timestamptz LANGUAGE sql STABLE AS $$
 SELECT LEAST(known_at, (SELECT min(suspended_at) FROM registry_interpretation_suspensions
 WHERE source_id=source AND (resumed_at IS NULL OR resumed_at>known_at)))
$$;
-- Query cursors explicitly expire on operator publication/interpretation changes
-- rather than serving a cached result under a superseded source policy. The
-- public-view module is optional in registry-only tools/tests.
CREATE FUNCTION registry_invalidate_public_views() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF OLD.public_enabled IS DISTINCT FROM NEW.public_enabled OR
    OLD.interpretation_suspended_at IS DISTINCT FROM NEW.interpretation_suspended_at THEN
  IF to_regclass('public_query_views') IS NOT NULL THEN
   EXECUTE 'DELETE FROM public_query_views';
  END IF;
 END IF;
 RETURN NEW;
END;
$$;
CREATE TRIGGER registry_invalidate_public_views AFTER UPDATE OF public_enabled,interpretation_suspended_at ON registry_sources
 FOR EACH ROW EXECUTE FUNCTION registry_invalidate_public_views();
