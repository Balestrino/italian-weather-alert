CREATE TABLE observation_campaigns (
 id text PRIMARY KEY CHECK (id <> ''),
 actor text NOT NULL CHECK (actor <> ''),
 started_at timestamptz NOT NULL,
 created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE observation_campaign_sources (
 campaign_id text NOT NULL REFERENCES observation_campaigns(id),
 source_id text NOT NULL,
 configuration integer NOT NULL,
 product_id text NOT NULL CHECK (product_id IN ('vigilance','criticality','monitoring','municipal')),
 territory text NOT NULL CHECK (territory <> ''),
 sections jsonb NOT NULL CHECK (jsonb_typeof(sections) = 'array'),
 check_seconds integer NOT NULL CHECK (check_seconds > 0),
 delay_seconds integer NOT NULL CHECK (delay_seconds > 0),
 PRIMARY KEY(campaign_id,source_id),
 FOREIGN KEY(source_id,configuration) REFERENCES registry_configurations(source_id,revision)
);

CREATE TABLE observation_reviews (
 campaign_id text NOT NULL REFERENCES observation_campaigns(id),
 id text NOT NULL CHECK (id <> ''),
 source_id text NOT NULL,
 kind text NOT NULL CHECK (kind IN ('original_comparison','attachment_comparison','attachment_not_applicable','failure_actual','failure_retained','absent_event_retained')),
 status text NOT NULL CHECK (status IN ('pass','fail','unresolved')),
 check_id bigint REFERENCES acquisition_checks(id),
 version_id bigint REFERENCES retained_versions(id),
 case_id text,
 evidence jsonb NOT NULL CHECK (jsonb_typeof(evidence) = 'object'),
 notes text NOT NULL DEFAULT '',
 actor text NOT NULL CHECK (actor <> ''),
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 PRIMARY KEY(campaign_id,id),
 FOREIGN KEY(campaign_id,source_id) REFERENCES observation_campaign_sources(campaign_id,source_id),
 CHECK (length(notes) <= 4000),
 CHECK (
   (kind IN ('original_comparison','attachment_comparison') AND version_id IS NOT NULL AND check_id IS NULL AND case_id IS NULL) OR
   (kind = 'failure_actual' AND check_id IS NOT NULL AND version_id IS NULL AND case_id IS NULL) OR
   (kind = 'attachment_not_applicable' AND check_id IS NULL AND version_id IS NULL AND case_id IS NULL) OR
   (kind IN ('failure_retained','absent_event_retained') AND check_id IS NULL AND version_id IS NULL AND case_id IS NOT NULL)
 )
);

CREATE TABLE observation_assessments (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 campaign_id text NOT NULL REFERENCES observation_campaigns(id),
 actor text NOT NULL CHECK (actor <> ''),
 through_at timestamptz NOT NULL,
 status text NOT NULL CHECK (status IN ('running','extended','complete')),
 evidence jsonb NOT NULL CHECK (jsonb_typeof(evidence) = 'object'),
 report jsonb NOT NULL CHECK (jsonb_typeof(report) = 'object'),
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX observation_assessments_campaign ON observation_assessments(campaign_id,id DESC);
CREATE UNIQUE INDEX observation_one_completion ON observation_assessments(campaign_id) WHERE status='complete';

CREATE FUNCTION observation_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'observation evidence is append-only'; END;
$$;
CREATE TRIGGER observation_campaign_immutable BEFORE UPDATE OR DELETE ON observation_campaigns FOR EACH ROW EXECUTE FUNCTION observation_immutable();
CREATE TRIGGER observation_source_immutable BEFORE UPDATE OR DELETE ON observation_campaign_sources FOR EACH ROW EXECUTE FUNCTION observation_immutable();
CREATE TRIGGER observation_review_immutable BEFORE UPDATE OR DELETE ON observation_reviews FOR EACH ROW EXECUTE FUNCTION observation_immutable();
CREATE TRIGGER observation_assessment_immutable BEFORE UPDATE OR DELETE ON observation_assessments FOR EACH ROW EXECUTE FUNCTION observation_immutable();
