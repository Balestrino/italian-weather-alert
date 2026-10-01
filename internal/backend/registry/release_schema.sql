CREATE TABLE registry_acceptance_reviews (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 source_id text NOT NULL,
 revision integer NOT NULL,
 acceptance_event_id bigint NOT NULL UNIQUE REFERENCES registry_events(id),
 product_id text NOT NULL CHECK (product_id IN ('vigilance','criticality','monitoring','municipal')),
 territory text NOT NULL CHECK (territory <> ''),
 status text NOT NULL CHECK (status IN ('pending','accepted')),
 actor text NOT NULL CHECK (actor <> ''),
 evidence jsonb NOT NULL CHECK (jsonb_typeof(evidence) = 'object'),
 report jsonb NOT NULL CHECK (jsonb_typeof(report) = 'object'),
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 FOREIGN KEY(source_id,revision) REFERENCES registry_configurations(source_id,revision)
);
CREATE INDEX registry_acceptance_reviews_scope ON registry_acceptance_reviews(product_id,territory,recorded_at DESC,id DESC);
CREATE INDEX registry_acceptance_reviews_source ON registry_acceptance_reviews(source_id,revision,recorded_at DESC,id DESC);
CREATE TRIGGER registry_acceptance_reviews_immutable BEFORE UPDATE OR DELETE ON registry_acceptance_reviews FOR EACH ROW EXECUTE FUNCTION registry_immutable();
