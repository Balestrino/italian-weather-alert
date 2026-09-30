CREATE TABLE geography_datasets (
 id text PRIMARY KEY CHECK (id <> ''),
 kind text NOT NULL CHECK (kind IN ('municipality_registry','zone_mapping','postal_candidates')),
 municipality_dataset_id text REFERENCES geography_datasets(id),
 authority text NOT NULL CHECK (authority <> ''),
 publisher text NOT NULL CHECK (publisher <> ''),
 official_url text NOT NULL CHECK (official_url <> ''),
 version_label text NOT NULL CHECK (version_label <> ''),
 source_sha256 text NOT NULL CHECK (source_sha256 ~ '^[0-9a-f]{64}$'),
 verified_at timestamptz NOT NULL,
 applicable_from date,
 applicable_to date,
 applicability text NOT NULL CHECK (applicability IN ('verified','unresolved','not_applicable')),
 limitations jsonb NOT NULL CHECK (jsonb_typeof(limitations)='array'),
 usable boolean NOT NULL,
 CHECK (applicable_to IS NULL OR applicable_from IS NULL OR applicable_to >= applicable_from),
 CHECK ((kind='municipality_registry') = (municipality_dataset_id IS NULL)),
 CHECK (applicability <> 'unresolved' OR (applicable_from IS NULL AND applicable_to IS NULL))
);

CREATE TABLE geography_municipalities (
 dataset_id text NOT NULL REFERENCES geography_datasets(id),
 istat text NOT NULL CHECK (istat ~ '^[0-9]{6}$'),
 name text NOT NULL CHECK (name <> ''),
 province text NOT NULL CHECK (province <> ''),
 supported boolean NOT NULL,
 PRIMARY KEY(dataset_id,istat)
);
CREATE INDEX geography_municipalities_name ON geography_municipalities(dataset_id,lower(name));

CREATE TABLE geography_zone_mappings (
 dataset_id text NOT NULL REFERENCES geography_datasets(id),
 municipality_dataset_id text NOT NULL,
 municipality_istat text NOT NULL,
 ordinal integer NOT NULL CHECK (ordinal > 0),
 zone text NOT NULL CHECK (zone <> ''),
 source_name text NOT NULL CHECK (source_name <> ''),
 territorial_scope text NOT NULL CHECK (territorial_scope IN ('whole_municipality','partial_municipality')),
 evidence_locator text NOT NULL CHECK (evidence_locator <> ''),
 PRIMARY KEY(dataset_id,ordinal),
 FOREIGN KEY(municipality_dataset_id,municipality_istat) REFERENCES geography_municipalities(dataset_id,istat)
);
CREATE INDEX geography_zone_municipality ON geography_zone_mappings(dataset_id,municipality_istat,zone);

CREATE TABLE geography_postal_mappings (
 dataset_id text NOT NULL REFERENCES geography_datasets(id),
 municipality_dataset_id text NOT NULL,
 postal_code text NOT NULL CHECK (postal_code ~ '^[0-9]{5}$'),
 municipality_istat text NOT NULL,
 PRIMARY KEY(dataset_id,postal_code,municipality_istat),
 FOREIGN KEY(municipality_dataset_id,municipality_istat) REFERENCES geography_municipalities(dataset_id,istat)
);
CREATE INDEX geography_postal_lookup ON geography_postal_mappings(dataset_id,postal_code);

CREATE TABLE geography_dataset_selections (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 kind text NOT NULL CHECK (kind IN ('municipality_registry','zone_mapping','postal_candidates')),
 dataset_id text NOT NULL REFERENCES geography_datasets(id),
 actor text NOT NULL CHECK (actor <> ''),
 selected_at timestamptz NOT NULL
);
CREATE INDEX geography_dataset_selection_latest ON geography_dataset_selections(kind,id DESC);

CREATE TRIGGER geography_datasets_immutable BEFORE UPDATE OR DELETE ON geography_datasets FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE TRIGGER geography_municipalities_immutable BEFORE UPDATE OR DELETE ON geography_municipalities FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE TRIGGER geography_zone_mappings_immutable BEFORE UPDATE OR DELETE ON geography_zone_mappings FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE TRIGGER geography_postal_mappings_immutable BEFORE UPDATE OR DELETE ON geography_postal_mappings FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE TRIGGER geography_dataset_selections_immutable BEFORE UPDATE OR DELETE ON geography_dataset_selections FOR EACH ROW EXECUTE FUNCTION domain_immutable();
