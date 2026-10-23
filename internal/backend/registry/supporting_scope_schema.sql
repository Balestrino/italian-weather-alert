-- Sources retained only as supporting evidence (referenced albo acts) are never
-- in the public perimeter: their acts appear only as evidence of clarified measures.
CREATE OR REPLACE FUNCTION registry_source_in_public_scope(source text) RETURNS boolean LANGUAGE plpgsql STABLE AS $$
DECLARE region text; supporting boolean;
BEGIN
 SELECT c.body->>'access_method'='jcitygov-albo' INTO supporting FROM registry_sources s
 JOIN registry_configurations c ON c.source_id=s.id AND c.revision=COALESCE(s.active_revision,s.latest_revision) WHERE s.id=source;
 IF COALESCE(supporting,false) THEN RETURN false; END IF;
 -- Legacy schemas contain only the original public perimeter. Once migrated,
 -- every source must have an explicit association, including historical sources.
 IF to_regclass('territorial_source_associations') IS NULL THEN RETURN true; END IF;
 SELECT region_code INTO region FROM territorial_current_sources WHERE source_id=source;
 RETURN COALESCE(region='09',false);
END $$;
