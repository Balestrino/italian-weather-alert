CREATE FUNCTION registry_source_in_public_scope(source text) RETURNS boolean LANGUAGE plpgsql STABLE AS $$
DECLARE region text;
BEGIN
 -- Legacy schemas contain only the original public perimeter. Once migrated,
 -- every source must have an explicit association, including historical sources.
 IF to_regclass('territorial_source_associations') IS NULL THEN RETURN true; END IF;
 SELECT region_code INTO region FROM territorial_current_sources WHERE source_id=source;
 RETURN COALESCE(region='09',false);
END $$;
CREATE VIEW registry_public_sources AS SELECT * FROM registry_sources WHERE registry_source_in_public_scope(id);
CREATE FUNCTION registry_dataset_in_public_scope(dataset text) RETURNS boolean LANGUAGE plpgsql STABLE AS $$
DECLARE region text;
BEGIN
 IF to_regclass('territorial_dataset_regions') IS NULL THEN RETURN true; END IF;
 SELECT region_code INTO region FROM territorial_dataset_regions WHERE dataset_id=dataset;
 RETURN COALESCE(region='09',false);
END $$;
