CREATE TABLE registry_authorities (
 id text PRIMARY KEY CHECK (id <> ''),
 name text NOT NULL CHECK (name <> ''),
 official_url text NOT NULL CHECK (official_url <> '')
);
CREATE TABLE registry_channels (
 id text PRIMARY KEY CHECK (id <> ''),
 publisher_id text NOT NULL REFERENCES registry_authorities(id),
 platform text NOT NULL CHECK (platform <> ''),
 url text NOT NULL CHECK (url <> ''),
 external boolean NOT NULL
);
CREATE TABLE registry_products (
 id text PRIMARY KEY,
 public_eligible boolean NOT NULL
);
INSERT INTO registry_products VALUES ('vigilance',true),('criticality',true),('monitoring',true),('municipal',true),('dpc_comparison',false);
CREATE TABLE registry_sources (
 id text PRIMARY KEY CHECK (id <> ''),
 authority_id text NOT NULL REFERENCES registry_authorities(id),
 channel_id text NOT NULL REFERENCES registry_channels(id),
 product_id text NOT NULL REFERENCES registry_products(id),
 territory text NOT NULL CHECK (territory <> ''),
 latest_revision integer NOT NULL DEFAULT 0,
 active_revision integer,
 collection_enabled boolean NOT NULL DEFAULT false,
 public_enabled boolean NOT NULL DEFAULT false,
 CHECK (NOT collection_enabled OR active_revision IS NOT NULL),
 CHECK (NOT public_enabled OR active_revision IS NOT NULL)
);
CREATE TABLE registry_configurations (
 source_id text NOT NULL REFERENCES registry_sources(id),
 revision integer NOT NULL CHECK (revision > 0),
 body jsonb NOT NULL CHECK (jsonb_typeof(body) = 'object'),
 actor text NOT NULL CHECK (actor <> ''),
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY (source_id,revision)
);
ALTER TABLE registry_sources ADD FOREIGN KEY (id,active_revision) REFERENCES registry_configurations(source_id,revision);
CREATE TABLE registry_events (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 source_id text NOT NULL,
 revision integer NOT NULL,
 kind text NOT NULL CHECK (kind IN ('preview','acceptance','collection_enabled','collection_disabled','public_enabled','public_disabled')),
 actor text NOT NULL CHECK (actor <> ''),
 evidence jsonb NOT NULL CHECK (jsonb_typeof(evidence) = 'object'),
 created_at timestamptz NOT NULL DEFAULT now(),
 FOREIGN KEY(source_id,revision) REFERENCES registry_configurations(source_id,revision)
);
CREATE INDEX registry_events_lookup ON registry_events(source_id,revision,kind);
CREATE FUNCTION registry_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'registry history is append-only'; END;
$$;
CREATE TRIGGER registry_config_immutable BEFORE UPDATE OR DELETE ON registry_configurations FOR EACH ROW EXECUTE FUNCTION registry_immutable();
CREATE TRIGGER registry_event_immutable BEFORE UPDATE OR DELETE ON registry_events FOR EACH ROW EXECUTE FUNCTION registry_immutable();
