CREATE TABLE territorial_dataset_regions (
 dataset_id text PRIMARY KEY REFERENCES geography_datasets(id),
 region_code text NOT NULL REFERENCES territorial_regions(code),
 complete boolean NOT NULL DEFAULT false,
 completeness_evidence text NOT NULL DEFAULT '',
 UNIQUE(region_code,dataset_id)
);
CREATE TABLE territorial_municipalities (
 region_code text NOT NULL,
 dataset_id text NOT NULL,
 istat text NOT NULL,
 PRIMARY KEY(region_code,dataset_id,istat),
 FOREIGN KEY(region_code,dataset_id) REFERENCES territorial_dataset_regions(region_code,dataset_id),
 FOREIGN KEY(dataset_id,istat) REFERENCES geography_municipalities(dataset_id,istat)
);
CREATE INDEX territorial_municipality_lookup ON territorial_municipalities(istat,region_code,dataset_id);
CREATE TABLE territorial_dataset_selections (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 region_code text NOT NULL,
 kind text NOT NULL CHECK(kind IN ('municipality_registry','zone_mapping','postal_candidates')),
 dataset_id text NOT NULL,
 actor text NOT NULL CHECK(btrim(actor)<>''),
 selected_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 FOREIGN KEY(region_code,dataset_id) REFERENCES territorial_dataset_regions(region_code,dataset_id)
);
CREATE INDEX territorial_selection_history ON territorial_dataset_selections(region_code,kind,selected_at DESC,id DESC);
CREATE TABLE territorial_source_associations (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 source_id text NOT NULL REFERENCES registry_sources(id),
 region_code text NOT NULL REFERENCES territorial_regions(code),
 municipality_dataset_id text,
 municipality_istat text,
 profile text NOT NULL CHECK(profile<>''),
 actor text NOT NULL CHECK(btrim(actor)<>''),
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 CHECK((municipality_dataset_id IS NULL)=(municipality_istat IS NULL)),
 FOREIGN KEY(region_code,municipality_dataset_id,municipality_istat) REFERENCES territorial_municipalities(region_code,dataset_id,istat)
);
CREATE INDEX territorial_source_association_history ON territorial_source_associations(source_id,recorded_at DESC,id DESC);
CREATE INDEX territorial_source_region ON territorial_source_associations(region_code,municipality_istat,source_id);
CREATE FUNCTION territorial_validate_dataset() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE parent text; dataset_kind text;
BEGIN
 SELECT municipality_dataset_id,kind INTO parent,dataset_kind FROM geography_datasets WHERE id=NEW.dataset_id;
 IF TG_TABLE_NAME='territorial_dataset_regions' AND parent IS NOT NULL AND NOT EXISTS(SELECT 1 FROM territorial_dataset_regions WHERE dataset_id=parent AND region_code=NEW.region_code) THEN
  RAISE EXCEPTION 'dataset parent belongs to another region';
 END IF;
 IF TG_TABLE_NAME='territorial_dataset_selections' THEN
  IF dataset_kind<>NEW.kind THEN RAISE EXCEPTION 'dataset kind mismatch'; END IF;
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER territorial_dataset_region_check BEFORE INSERT ON territorial_dataset_regions FOR EACH ROW EXECUTE FUNCTION territorial_validate_dataset();
CREATE TRIGGER territorial_selection_check BEFORE INSERT ON territorial_dataset_selections FOR EACH ROW EXECUTE FUNCTION territorial_validate_dataset();
CREATE TRIGGER territorial_dataset_regions_immutable BEFORE UPDATE OR DELETE ON territorial_dataset_regions FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE TRIGGER territorial_municipalities_immutable BEFORE UPDATE OR DELETE ON territorial_municipalities FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE TRIGGER territorial_selections_immutable BEFORE UPDATE OR DELETE ON territorial_dataset_selections FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE TRIGGER territorial_source_associations_immutable BEFORE UPDATE OR DELETE ON territorial_source_associations FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE VIEW territorial_current_sources AS
 SELECT DISTINCT ON(source_id) source_id,region_code,municipality_dataset_id,municipality_istat,profile,recorded_at,id
 FROM territorial_source_associations ORDER BY source_id,recorded_at DESC,id DESC;
