CREATE TABLE development_publication_municipalities (
 region_code text NOT NULL REFERENCES territorial_regions(code),
 istat text NOT NULL CHECK(istat ~ '^[0-9]{6}$'),
 enabled boolean NOT NULL DEFAULT false,
 revision integer NOT NULL DEFAULT 0 CHECK(revision >= 0),
 PRIMARY KEY(region_code,istat)
);
CREATE TABLE development_publication_events (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 region_code text NOT NULL,
 istat text NOT NULL,
 revision integer NOT NULL CHECK(revision > 0),
 enabled boolean NOT NULL,
 actor text NOT NULL CHECK(btrim(actor)<>''),
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 UNIQUE(region_code,istat,revision),
 FOREIGN KEY(region_code,istat) REFERENCES development_publication_municipalities(region_code,istat)
);
CREATE TRIGGER development_publication_events_immutable BEFORE UPDATE OR DELETE ON development_publication_events FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE VIEW development_publication_current_municipalities AS
 SELECT st.* FROM development_publication_municipalities st
 JOIN territorial_regions r ON r.code=st.region_code
 JOIN territorial_region_versions rv ON rv.region_code=r.code AND rv.revision=r.revision
 WHERE st.enabled AND EXISTS(SELECT 1 FROM territorial_municipalities m
 JOIN geography_municipalities g ON g.dataset_id=m.dataset_id AND g.istat=m.istat
 WHERE m.region_code=st.region_code AND m.istat=st.istat AND m.dataset_id=rv.configuration->>'municipality_dataset' AND g.supported);
