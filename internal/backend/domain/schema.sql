CREATE TABLE domain_risks (
 id text PRIMARY KEY CHECK (id IN ('minor_network_hydro','main_network_hydraulic','thunderstorms','wind','coastal_waves','snow','ice')),
 official_label text NOT NULL CHECK (official_label <> '')
);
INSERT INTO domain_risks(id,official_label) VALUES
 ('minor_network_hydro','Rischio idrogeologico-idraulico del reticolo minore'),
 ('main_network_hydraulic','Rischio idraulico del reticolo principale'),
 ('thunderstorms','Temporali forti'),
 ('wind','Vento'),
 ('coastal_waves','Mareggiate'),
 ('snow','Neve'),
 ('ice','Ghiaccio');

CREATE TABLE domain_regional_products (
 id text PRIMARY KEY CHECK (id IN ('vigilance','criticality','monitoring'))
);
INSERT INTO domain_regional_products(id) VALUES ('vigilance'),('criticality'),('monitoring');

CREATE TABLE domain_regional_records (
 id text PRIMARY KEY CHECK (id <> ''),
 document_version_id bigint NOT NULL REFERENCES retained_versions(id),
 source_id text NOT NULL REFERENCES registry_sources(id),
 product text NOT NULL REFERENCES domain_regional_products(id),
 originating_authority_id text NOT NULL REFERENCES registry_authorities(id),
 publisher_id text NOT NULL REFERENCES registry_authorities(id),
 platform text NOT NULL CHECK (platform <> ''),
 municipal_republication boolean NOT NULL,
 UNIQUE(id,product)
);

CREATE TABLE domain_regional_facts (
 regional_record_id text NOT NULL,
 product text NOT NULL,
 ordinal integer NOT NULL CHECK (ordinal > 0),
 risk text NOT NULL REFERENCES domain_risks(id),
 official_risk_label text NOT NULL CHECK (official_risk_label <> ''),
 zone text NOT NULL DEFAULT '',
 level text CHECK (level IN ('green','yellow','orange','red','unknown','not_applicable')),
 PRIMARY KEY(regional_record_id,ordinal),
 FOREIGN KEY(regional_record_id,product) REFERENCES domain_regional_records(id,product),
 CHECK (product <> 'monitoring' OR level IS NULL)
);

CREATE TABLE domain_local_measures (
 id text PRIMARY KEY CHECK (id <> ''),
 document_version_id bigint NOT NULL REFERENCES retained_versions(id),
 source_id text NOT NULL REFERENCES registry_sources(id),
 municipality_istat text NOT NULL CHECK (municipality_istat ~ '^[0-9]{6}$'),
 issuing_authority_id text NOT NULL REFERENCES registry_authorities(id),
 publisher_id text NOT NULL REFERENCES registry_authorities(id),
 platform text NOT NULL CHECK (platform <> ''),
 kind text NOT NULL CHECK (kind IN ('closure','reopening','restriction','prohibition','suspension','activation','deactivation','operational_update','observation')),
 subject text NOT NULL CHECK (subject <> ''),
 place text,
 regional_record_id text REFERENCES domain_regional_records(id)
);

CREATE TABLE domain_operational_phases (
 id text PRIMARY KEY CHECK (id <> ''),
 document_version_id bigint NOT NULL REFERENCES retained_versions(id),
 source_id text NOT NULL REFERENCES registry_sources(id),
 municipality_istat text NOT NULL CHECK (municipality_istat ~ '^[0-9]{6}$'),
 authority_id text NOT NULL REFERENCES registry_authorities(id),
 publisher_id text NOT NULL REFERENCES registry_authorities(id),
 platform text NOT NULL CHECK (platform <> ''),
 phase text NOT NULL CHECK (phase <> ''),
 regional_record_id text REFERENCES domain_regional_records(id)
);

CREATE FUNCTION domain_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'domain facts are append-only'; END;
$$;
CREATE TRIGGER domain_regional_records_immutable BEFORE UPDATE OR DELETE ON domain_regional_records FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE TRIGGER domain_regional_facts_immutable BEFORE UPDATE OR DELETE ON domain_regional_facts FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE TRIGGER domain_local_measures_immutable BEFORE UPDATE OR DELETE ON domain_local_measures FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE TRIGGER domain_operational_phases_immutable BEFORE UPDATE OR DELETE ON domain_operational_phases FOR EACH ROW EXECUTE FUNCTION domain_immutable();
