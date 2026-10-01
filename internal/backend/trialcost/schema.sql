CREATE TABLE trial_cost_inputs (
 campaign_id text NOT NULL REFERENCES observation_campaigns(id),
 category text NOT NULL CHECK (category IN ('hosting','storage')),
 workload text NOT NULL CHECK (workload IN ('bootstrap','ordinary')),
 component text NOT NULL CHECK (component <> ''),
 revision integer NOT NULL CHECK (revision > 0),
 status text NOT NULL CHECK (status IN ('reported','unavailable','not_applicable')),
 basis text,
 quantity bigint CHECK (quantity >= 0),
 unit text,
 unit_size bigint CHECK (unit_size > 0),
 price_microunits bigint CHECK (price_microunits >= 0),
 cost_microunits bigint CHECK (cost_microunits >= 0),
 currency text CHECK (currency ~ '^[A-Z]{3}$'),
 effective_from timestamptz,
 effective_through timestamptz,
 evidence jsonb NOT NULL CHECK (jsonb_typeof(evidence) = 'object'),
 notes text NOT NULL,
 actor text NOT NULL CHECK (actor <> ''),
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 PRIMARY KEY(campaign_id,category,workload,component,revision),
 CHECK (length(notes) <= 4000),
 CHECK (effective_through IS NULL OR effective_from IS NULL OR effective_through > effective_from),
 CHECK (
   (status='reported' AND basis IN ('measured','rate') AND quantity IS NOT NULL AND unit IS NOT NULL AND unit<>''
    AND unit_size IS NOT NULL AND price_microunits IS NOT NULL AND cost_microunits IS NOT NULL
    AND currency IS NOT NULL AND effective_from IS NOT NULL AND effective_through IS NOT NULL)
   OR
   (status IN ('unavailable','not_applicable') AND basis IS NULL AND quantity IS NULL AND unit IS NULL
    AND unit_size IS NULL AND price_microunits IS NULL AND cost_microunits IS NULL AND currency IS NULL
    AND effective_from IS NULL AND effective_through IS NULL AND notes<>'')
 )
);

CREATE TABLE trial_budget_proposals (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 campaign_id text NOT NULL REFERENCES observation_campaigns(id),
 actor text NOT NULL CHECK (actor <> ''),
 through_at timestamptz NOT NULL,
 currency text NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
 evidence jsonb NOT NULL CHECK (jsonb_typeof(evidence) = 'object'),
 report jsonb NOT NULL CHECK (jsonb_typeof(report) = 'object'),
 bootstrap_cost_microunits bigint NOT NULL CHECK (bootstrap_cost_microunits >= 0),
 ordinary_trial_cost_microunits bigint NOT NULL CHECK (ordinary_trial_cost_microunits >= 0),
 ordinary_monthly_cost_microunits bigint NOT NULL CHECK (ordinary_monthly_cost_microunits >= 0),
 retry_trial_cost_microunits bigint NOT NULL CHECK (retry_trial_cost_microunits >= 0),
 evaluation_cost_microunits bigint NOT NULL CHECK (evaluation_cost_microunits >= 0),
 reprocessing_cost_microunits bigint NOT NULL CHECK (reprocessing_cost_microunits >= 0),
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 UNIQUE(campaign_id,through_at)
);

CREATE FUNCTION trial_cost_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'trial cost evidence is append-only'; END;
$$;
CREATE TRIGGER trial_cost_inputs_immutable BEFORE UPDATE OR DELETE ON trial_cost_inputs FOR EACH ROW EXECUTE FUNCTION trial_cost_immutable();
CREATE TRIGGER trial_budget_proposals_immutable BEFORE UPDATE OR DELETE ON trial_budget_proposals FOR EACH ROW EXECUTE FUNCTION trial_cost_immutable();
