CREATE TABLE act_references (
 id bigserial PRIMARY KEY,
 document_version_id bigint NOT NULL REFERENCES retained_versions(id),
 ordinal integer NOT NULL CHECK (ordinal BETWEEN 1 AND 3),
 kind text NOT NULL CHECK (kind IN ('ordinanza','decreto','delibera','determina')),
 act_number integer CHECK (act_number > 0),
 act_year integer CHECK (act_year BETWEEN 1900 AND 2999),
 act_date date,
 passage text NOT NULL CHECK (passage <> '' AND length(passage) <= 2000),
 resource_url text NOT NULL CHECK (resource_url <> ''),
 page_number integer CHECK (page_number > 0),
 detection_version text NOT NULL CHECK (detection_version <> ''),
 detected_at timestamptz NOT NULL,
 UNIQUE (document_version_id, detection_version, ordinal)
);
CREATE INDEX act_references_version ON act_references(document_version_id);
CREATE TABLE act_albo_records (
 id bigserial PRIMARY KEY,
 source_id text NOT NULL REFERENCES registry_sources(id),
 record_id text NOT NULL CHECK (record_id ~ '^[0-9]{1,18}$'),
 category text NOT NULL CHECK (category <> ''),
 subcategory text NOT NULL DEFAULT '',
 act_number integer CHECK (act_number > 0),
 act_year integer CHECK (act_year BETWEEN 1900 AND 2999),
 act_date date,
 registry_year integer NOT NULL CHECK (registry_year BETWEEN 1900 AND 2999),
 registry_number integer NOT NULL CHECK (registry_number > 0),
 object text NOT NULL CHECK (object <> ''),
 detail_url text NOT NULL CHECK (detail_url LIKE 'https://%'),
 document_version_id bigint NOT NULL REFERENCES retained_versions(id),
 recorded_at timestamptz NOT NULL,
 UNIQUE (source_id, record_id)
);
CREATE INDEX act_albo_records_version ON act_albo_records(document_version_id);
CREATE TABLE act_resolutions (
 id bigserial PRIMARY KEY,
 reference_id bigint NOT NULL REFERENCES act_references(id),
 outcome text NOT NULL CHECK (outcome IN ('attached','numbered','unique_candidate','not_found','ambiguous','unavailable')),
 albo_source_id text REFERENCES registry_sources(id),
 albo_record_id text,
 attached_resource_url text,
 candidates jsonb NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(candidates)='array'),
 reason text NOT NULL DEFAULT '',
 recorded_at timestamptz NOT NULL,
 CHECK ((outcome='attached') = (attached_resource_url IS NOT NULL)),
 CHECK ((outcome IN ('numbered','unique_candidate')) = (albo_record_id IS NOT NULL AND albo_source_id IS NOT NULL)),
 CHECK (outcome NOT IN ('ambiguous','unavailable') OR reason <> '')
);
CREATE INDEX act_resolutions_reference ON act_resolutions(reference_id, recorded_at DESC, id DESC);
CREATE TABLE act_clarifications (
 id bigserial PRIMARY KEY,
 resolution_id bigint NOT NULL REFERENCES act_resolutions(id),
 act_document_version_id bigint NOT NULL REFERENCES retained_versions(id),
 local_measure_id text REFERENCES domain_local_measures(id),
 outcome text NOT NULL CHECK (outcome IN ('clarified','conflict','not_matched','clarification_unavailable')),
 disposition_ordinal integer CHECK (disposition_ordinal > 0),
 act_resource_url text,
 act_page_number integer CHECK (act_page_number > 0),
 act_passage text NOT NULL DEFAULT '',
 notice_passage text NOT NULL DEFAULT '',
 parser_version text NOT NULL CHECK (parser_version <> ''),
 recorded_at timestamptz NOT NULL,
 CHECK ((outcome='clarification_unavailable') = (local_measure_id IS NULL)),
 CHECK (outcome NOT IN ('clarified','conflict') OR (disposition_ordinal IS NOT NULL AND act_passage <> '' AND act_resource_url IS NOT NULL))
);
CREATE UNIQUE INDEX act_clarifications_once ON act_clarifications(resolution_id, COALESCE(local_measure_id,''), parser_version);
CREATE INDEX act_clarifications_measure ON act_clarifications(local_measure_id, recorded_at);
CREATE INDEX act_clarifications_act ON act_clarifications(act_document_version_id);
CREATE TRIGGER act_references_immutable BEFORE UPDATE OR DELETE ON act_references FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
CREATE TRIGGER act_albo_records_immutable BEFORE UPDATE OR DELETE ON act_albo_records FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
CREATE TRIGGER act_resolutions_immutable BEFORE UPDATE OR DELETE ON act_resolutions FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
CREATE TRIGGER act_clarifications_immutable BEFORE UPDATE OR DELETE ON act_clarifications FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
