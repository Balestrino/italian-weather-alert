CREATE TABLE acquisition_target_dispositions (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 source_id text NOT NULL,
 configuration integer NOT NULL,
 url text NOT NULL,
 decision text NOT NULL CHECK (decision IN ('exclude_stale_unavailable','resume')),
 actor text NOT NULL CHECK (length(btrim(actor)) > 0),
 reason text NOT NULL CHECK (length(btrim(reason)) > 0),
 evidence jsonb NOT NULL CHECK (jsonb_typeof(evidence) = 'object'),
 observed_last_seen_at timestamptz NOT NULL,
 recorded_at timestamptz NOT NULL DEFAULT now(),
 FOREIGN KEY(source_id,configuration,url)
   REFERENCES acquisition_unavailable_targets(source_id,configuration,url)
);
CREATE INDEX acquisition_target_dispositions_latest
 ON acquisition_target_dispositions(source_id,configuration,url,id DESC);
