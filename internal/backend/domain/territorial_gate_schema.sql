CREATE FUNCTION territorial_source_allowed(source text) RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE region text; active boolean; profile text; product text; profiles jsonb;
BEGIN
 SELECT a.region_code,COALESCE(NULLIF(c.body->>'processing_profile',''),a.profile),s.product_id INTO region,profile,product
 FROM territorial_current_sources a JOIN registry_sources s ON s.id=a.source_id
 LEFT JOIN registry_configurations c ON c.source_id=s.id AND c.revision=s.active_revision WHERE a.source_id=source;
 IF region IS NULL THEN RETURN false; END IF;
 SELECT r.enabled,v.configuration->'profiles' INTO active,profiles FROM territorial_regions r
 LEFT JOIN territorial_region_versions v ON v.region_code=r.code AND v.revision=r.revision WHERE r.code=region FOR SHARE OF r;
 RETURN COALESCE(active AND profiles ? profile AND
 ((product='municipal' AND profile='municipal-html') OR (region='09' AND
 ((profile='toscana-cfr' AND product IN ('criticality','vigilance','monitoring')) OR (profile='dpc-comparison' AND product='dpc_comparison')))),false);
END $$;
CREATE FUNCTION territorial_job_allowed(payload jsonb) RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE version_id bigint; source text;
BEGIN
 IF payload ? 'source_id' THEN RETURN territorial_source_allowed(payload->>'source_id'); END IF;
 IF payload ? 'document_version_id' THEN
  version_id:=(payload->>'document_version_id')::bigint;
 ELSIF payload ? 'extraction_run_id' THEN
  SELECT document_version_id INTO version_id FROM extraction_results WHERE run_id=(payload->>'extraction_run_id')::bigint;
 ELSE RETURN true;
 END IF;
 SELECT d.source_id INTO source FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id WHERE v.id=version_id;
 IF source IS NULL THEN RETURN false; END IF;
 RETURN territorial_source_allowed(source);
END $$;
