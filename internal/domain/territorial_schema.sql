CREATE TABLE territorial_regions (
 code text PRIMARY KEY CHECK(code ~ '^[0-9]{2}$'),
 name text NOT NULL UNIQUE CHECK(name <> ''),
 reference jsonb NOT NULL,
 revision integer NOT NULL DEFAULT 0 CHECK(revision >= 0),
 enabled boolean NOT NULL DEFAULT false
);
CREATE TABLE territorial_region_versions (
 region_code text NOT NULL REFERENCES territorial_regions(code),
 revision integer NOT NULL CHECK(revision > 0),
 configuration jsonb NOT NULL CHECK(jsonb_typeof(configuration)='object'),
 enabled boolean NOT NULL,
 actor text NOT NULL CHECK(btrim(actor) <> ''),
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 PRIMARY KEY(region_code,revision)
);
CREATE TABLE territorial_region_events (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 region_code text NOT NULL,
 revision integer NOT NULL,
 kind text NOT NULL CHECK(kind IN ('configuration','enabled','disabled','dataset_adopted','migration')),
 actor text NOT NULL CHECK(btrim(actor) <> ''),
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 FOREIGN KEY(region_code,revision) REFERENCES territorial_region_versions(region_code,revision)
);
CREATE INDEX territorial_region_event_history ON territorial_region_events(region_code,recorded_at,id);
CREATE TRIGGER territorial_versions_immutable BEFORE UPDATE OR DELETE ON territorial_region_versions FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE TRIGGER territorial_events_immutable BEFORE UPDATE OR DELETE ON territorial_region_events FOR EACH ROW EXECUTE FUNCTION domain_immutable();
