CREATE TABLE diagnostic_dpc_comparisons (
 id text PRIMARY KEY CHECK (id <> ''),
 regional_source_id text NOT NULL REFERENCES registry_sources(id),
 regional_version_id bigint NOT NULL REFERENCES retained_versions(id),
 dpc_source_id text NOT NULL REFERENCES registry_sources(id),
 dpc_version_id bigint NOT NULL REFERENCES retained_versions(id),
 scope text NOT NULL CHECK (scope <> ''),
 dimensions jsonb NOT NULL CHECK (jsonb_typeof(dimensions)='array' AND jsonb_array_length(dimensions)=4),
 actor text NOT NULL CHECK (actor <> ''),
 compared_at timestamptz NOT NULL,
 CHECK (regional_version_id <> dpc_version_id)
);
CREATE INDEX diagnostic_dpc_comparisons_scope ON diagnostic_dpc_comparisons(regional_source_id,scope,compared_at DESC,id DESC);
CREATE FUNCTION diagnostics_validate_comparison() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE regional_source text; regional_product text; dpc_source text; dpc_product text;
BEGIN
 SELECT d.source_id,s.product_id INTO regional_source,regional_product
 FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id JOIN registry_sources s ON s.id=d.source_id
 WHERE v.id=NEW.regional_version_id;
 SELECT d.source_id,s.product_id INTO dpc_source,dpc_product
 FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id JOIN registry_sources s ON s.id=d.source_id
 WHERE v.id=NEW.dpc_version_id;
 IF regional_source IS DISTINCT FROM NEW.regional_source_id OR regional_product NOT IN ('vigilance','criticality','monitoring')
    OR dpc_source IS DISTINCT FROM NEW.dpc_source_id OR dpc_product<>'dpc_comparison' THEN
   RAISE EXCEPTION 'invalid diagnostic source/version scope' USING ERRCODE='23514';
 END IF;
 IF NOT (SELECT count(DISTINCT item->>'name')=4
             AND bool_and(item->>'name' IN ('risk','territory','issuance','validity'))
             AND bool_and(item->>'state' IN ('same','different','not_comparable'))
          FROM jsonb_array_elements(NEW.dimensions) item) THEN
   RAISE EXCEPTION 'invalid diagnostic dimensions' USING ERRCODE='23514';
 END IF;
 IF EXISTS (SELECT 1 FROM jsonb_array_elements(NEW.dimensions) item
   WHERE (item->>'state'='not_comparable' AND (btrim(item->>'reason')='' OR item->'regional_value'<>'null'::jsonb OR item->'dpc_value'<>'null'::jsonb))
      OR (item->>'state'<>'not_comparable' AND (btrim(item->>'reason')<>'' OR jsonb_typeof(item->'regional_value')<>'string' OR btrim(item->>'regional_value')='' OR jsonb_typeof(item->'dpc_value')<>'string' OR btrim(item->>'dpc_value')=''))
      OR btrim(item->'regional_evidence'->>'resource_url')='' OR btrim(item->'regional_evidence'->>'locator')=''
      OR btrim(item->'dpc_evidence'->>'resource_url')='' OR btrim(item->'dpc_evidence'->>'locator')=''
      OR NOT EXISTS (SELECT 1 FROM retained_resources r WHERE r.version_id=NEW.regional_version_id AND r.url=item->'regional_evidence'->>'resource_url' AND r.missing='')
      OR NOT EXISTS (SELECT 1 FROM retained_resources r WHERE r.version_id=NEW.dpc_version_id AND r.url=item->'dpc_evidence'->>'resource_url' AND r.missing='')) THEN
   RAISE EXCEPTION 'invalid diagnostic evidence' USING ERRCODE='23514';
 END IF;
 RETURN NEW;
END;
$$;
CREATE TRIGGER diagnostic_dpc_comparisons_validate BEFORE INSERT ON diagnostic_dpc_comparisons FOR EACH ROW EXECUTE FUNCTION diagnostics_validate_comparison();
CREATE FUNCTION diagnostics_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'diagnostic history is append-only'; END;
$$;
CREATE TRIGGER diagnostic_dpc_comparisons_immutable BEFORE UPDATE OR DELETE ON diagnostic_dpc_comparisons FOR EACH ROW EXECUTE FUNCTION diagnostics_immutable();
