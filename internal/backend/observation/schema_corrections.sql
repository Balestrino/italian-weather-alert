CREATE TABLE observation_review_corrections (
 campaign_id text NOT NULL,
 review_id text NOT NULL,
 previous_id text NOT NULL,
 reason text NOT NULL CHECK (length(btrim(reason)) BETWEEN 1 AND 4000),
 PRIMARY KEY(campaign_id,review_id),
 UNIQUE(campaign_id,previous_id),
 FOREIGN KEY(campaign_id,review_id) REFERENCES observation_reviews(campaign_id,id),
 FOREIGN KEY(campaign_id,previous_id) REFERENCES observation_reviews(campaign_id,id),
 CHECK (review_id <> previous_id)
);
CREATE TRIGGER observation_correction_immutable BEFORE UPDATE OR DELETE ON observation_review_corrections FOR EACH ROW EXECUTE FUNCTION observation_immutable();
