CREATE TABLE territorial_municipality_state (
 region_code text NOT NULL REFERENCES territorial_regions(code),
 istat text NOT NULL CHECK(istat ~ '^[0-9]{6}$'),
 enabled boolean NOT NULL DEFAULT false,
 revision integer NOT NULL DEFAULT 0 CHECK(revision >= 0),
 PRIMARY KEY(region_code,istat)
);
CREATE TABLE territorial_municipality_events (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 region_code text NOT NULL,
 istat text NOT NULL,
 revision integer NOT NULL CHECK(revision > 0),
 enabled boolean NOT NULL,
 kind text NOT NULL CHECK(kind IN ('enabled','disabled','migration')),
 actor text NOT NULL CHECK(btrim(actor)<>''),
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 UNIQUE(region_code,istat,revision),
 FOREIGN KEY(region_code,istat) REFERENCES territorial_municipality_state(region_code,istat)
);
CREATE INDEX territorial_municipality_event_history ON territorial_municipality_events(region_code,istat,recorded_at,id);
CREATE TRIGGER territorial_municipality_events_immutable BEFORE UPDATE OR DELETE ON territorial_municipality_events FOR EACH ROW EXECUTE FUNCTION domain_immutable();

CREATE OR REPLACE FUNCTION territorial_source_allowed(source text) RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE region text; municipality text; active boolean; municipal_active boolean; profile text; product text; profiles jsonb; regional_kind text; municipality_dataset text;
BEGIN
 SELECT a.region_code,a.municipality_istat,COALESCE(NULLIF(c.body->>'processing_profile',''),a.profile),s.product_id,c.body->'regional_product'->>'kind'
 INTO region,municipality,profile,product,regional_kind
 FROM territorial_current_sources a JOIN registry_sources s ON s.id=a.source_id
 LEFT JOIN registry_configurations c ON c.source_id=s.id AND c.revision=s.active_revision WHERE a.source_id=source;
 IF region IS NULL THEN RETURN false; END IF;
 -- Lock parents before children, matching lifecycle mutations. Admitted work may finish.
 SELECT r.enabled,v.configuration->'profiles',v.configuration->>'municipality_dataset' INTO active,profiles,municipality_dataset FROM territorial_regions r
 LEFT JOIN territorial_region_versions v ON v.region_code=r.code AND v.revision=r.revision WHERE r.code=region FOR SHARE OF r;
 IF product='municipal' THEN
  SELECT m.enabled INTO municipal_active FROM territorial_municipality_state m
  WHERE m.region_code=region AND m.istat=municipality FOR SHARE OF m;
  IF NOT COALESCE(municipal_active,false) THEN RETURN false; END IF;
  IF NOT EXISTS(SELECT 1 FROM territorial_municipalities m WHERE m.region_code=region AND m.istat=municipality AND m.dataset_id=municipality_dataset) THEN RETURN false; END IF;
 END IF;
 RETURN COALESCE(active AND profiles ? profile AND
 (regional_kind IS NULL OR (region='09' AND profile='toscana-cfr' AND product=regional_kind)) AND
 ((product='municipal' AND profile='municipal-html') OR (region='09' AND
 ((profile='toscana-cfr' AND product IN ('criticality','vigilance','monitoring')) OR (profile='dpc-comparison' AND product='dpc_comparison')))),false);
END $$;
