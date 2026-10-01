CREATE TABLE registry_regressions (
 id text PRIMARY KEY CHECK (id <> ''),
 source_id text NOT NULL,
 revision integer NOT NULL,
 suite text NOT NULL CHECK (suite <> ''),
 contract_hash text NOT NULL CHECK (length(contract_hash) = 64),
 previous_id text REFERENCES registry_regressions(id),
 passed boolean NOT NULL,
 actor text NOT NULL CHECK (actor <> ''),
 report jsonb NOT NULL CHECK (jsonb_typeof(report) = 'object'),
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 FOREIGN KEY(source_id,revision) REFERENCES registry_configurations(source_id,revision)
);
CREATE INDEX registry_regressions_scope ON registry_regressions(source_id,revision,suite,recorded_at DESC);
CREATE TRIGGER registry_regressions_immutable BEFORE UPDATE OR DELETE ON registry_regressions FOR EACH ROW EXECUTE FUNCTION registry_immutable();
