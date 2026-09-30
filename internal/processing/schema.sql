CREATE TABLE processing_model_versions (
 id text PRIMARY KEY CHECK (id <> ''),
 provider text NOT NULL CHECK (provider <> ''),
 model text NOT NULL CHECK (model <> ''),
 revision text NOT NULL CHECK (revision <> ''),
 capabilities jsonb NOT NULL CHECK (jsonb_typeof(capabilities) = 'object'),
 content_hash text NOT NULL CHECK (content_hash ~ '^[0-9a-f]{64}$'),
 created_at timestamptz NOT NULL,
 UNIQUE(provider,model,revision)
);

CREATE TABLE processing_prompt_versions (
 id text PRIMARY KEY CHECK (id <> ''),
 name text NOT NULL CHECK (name <> ''),
 stage text NOT NULL CHECK (stage IN ('collection','ocr','classification','extraction','linking','embedding')),
 revision text NOT NULL CHECK (revision <> ''),
 body text NOT NULL CHECK (body <> ''),
 content_hash text NOT NULL CHECK (content_hash ~ '^[0-9a-f]{64}$'),
 created_at timestamptz NOT NULL,
 UNIQUE(name,stage,revision)
);

CREATE TABLE processing_price_versions (
 id text PRIMARY KEY CHECK (id <> ''),
 model_version_id text NOT NULL REFERENCES processing_model_versions(id),
 currency text NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
 provenance_url text NOT NULL CHECK (provenance_url <> ''),
 observed_at timestamptz NOT NULL,
 effective_from timestamptz,
 effective_through timestamptz,
 details jsonb NOT NULL CHECK (jsonb_typeof(details) = 'object'),
 content_hash text NOT NULL CHECK (content_hash ~ '^[0-9a-f]{64}$'),
 created_at timestamptz NOT NULL,
 CHECK (effective_through IS NULL OR effective_from IS NULL OR effective_through > effective_from)
);
CREATE TABLE processing_price_rates (
 price_version_id text NOT NULL REFERENCES processing_price_versions(id),
 metric text NOT NULL CHECK (metric <> ''),
 unit_size bigint NOT NULL CHECK (unit_size > 0),
 price_microunits bigint NOT NULL CHECK (price_microunits >= 0),
 PRIMARY KEY(price_version_id,metric)
);

CREATE TABLE processing_configuration_versions (
 id text PRIMARY KEY CHECK (id <> ''),
 name text NOT NULL CHECK (name <> ''),
 stage text NOT NULL CHECK (stage IN ('collection','ocr','classification','extraction','linking','embedding')),
 revision text NOT NULL CHECK (revision <> ''),
 model_version_id text REFERENCES processing_model_versions(id),
 prompt_version_id text REFERENCES processing_prompt_versions(id),
 logic_version text NOT NULL CHECK (logic_version <> ''),
 settings jsonb NOT NULL CHECK (jsonb_typeof(settings) = 'object'),
 content_hash text NOT NULL CHECK (content_hash ~ '^[0-9a-f]{64}$'),
 created_at timestamptz NOT NULL,
 UNIQUE(name,stage,revision)
);

CREATE TABLE processing_runs (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 idempotency_key text NOT NULL UNIQUE CHECK (idempotency_key <> ''),
 request_hash text NOT NULL CHECK (request_hash ~ '^[0-9a-f]{64}$'),
 workload text NOT NULL CHECK (workload IN ('bootstrap','ordinary','evaluation','reprocessing')),
 stage text NOT NULL CHECK (stage IN ('collection','ocr','classification','extraction','linking','embedding')),
 configuration_version_id text NOT NULL REFERENCES processing_configuration_versions(id),
 source_id text REFERENCES registry_sources(id),
 document_version_id bigint REFERENCES retained_versions(id),
 subject jsonb NOT NULL CHECK (jsonb_typeof(subject) = 'object'),
 created_at timestamptz NOT NULL
);
CREATE INDEX processing_runs_document ON processing_runs(document_version_id,stage,created_at);
CREATE INDEX processing_runs_source ON processing_runs(source_id,workload,stage,created_at);

CREATE TABLE processing_run_attempts (
 run_id bigint NOT NULL REFERENCES processing_runs(id),
 number integer NOT NULL CHECK (number > 0),
 queue_job_id bigint,
 queue_attempt_number integer,
 started_at timestamptz NOT NULL,
 finished_at timestamptz,
 duration_ms bigint CHECK (duration_ms >= 0),
 outcome text NOT NULL CHECK (outcome IN ('running','succeeded','failed')),
 usage_status text CHECK (usage_status IN ('reported','partial','unavailable','not_applicable')),
 input_tokens bigint CHECK (input_tokens >= 0),
 output_tokens bigint CHECK (output_tokens >= 0),
 cache_read_tokens bigint CHECK (cache_read_tokens >= 0),
 cache_write_tokens bigint CHECK (cache_write_tokens >= 0),
 other_units jsonb,
 price_version_id text REFERENCES processing_price_versions(id),
 cost_status text CHECK (cost_status IN ('calculated','unknown','not_applicable')),
 estimated_cost_microunits bigint CHECK (estimated_cost_microunits >= 0),
 error_code text,
 PRIMARY KEY(run_id,number),
 FOREIGN KEY(queue_job_id,queue_attempt_number) REFERENCES processing_attempts(job_id,number),
 CHECK ((queue_job_id IS NULL) = (queue_attempt_number IS NULL)),
 CHECK ((outcome = 'running') = (finished_at IS NULL)),
 CHECK ((outcome = 'running') = (duration_ms IS NULL)),
 CHECK ((outcome = 'running') = (usage_status IS NULL)),
 CHECK ((outcome = 'running') = (cost_status IS NULL)),
 CHECK (other_units IS NULL OR jsonb_typeof(other_units) = 'object'),
 CHECK ((cost_status = 'calculated') = (estimated_cost_microunits IS NOT NULL)),
 CHECK (cost_status <> 'not_applicable' OR price_version_id IS NULL),
 CHECK (usage_status NOT IN ('unavailable','not_applicable') OR
   (input_tokens IS NULL AND output_tokens IS NULL AND cache_read_tokens IS NULL AND cache_write_tokens IS NULL AND (other_units IS NULL OR other_units = '{}'::jsonb)))
);

CREATE FUNCTION processing_catalog_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'processing version history is append-only'; END;
$$;
CREATE TRIGGER processing_models_immutable BEFORE UPDATE OR DELETE ON processing_model_versions FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
CREATE TRIGGER processing_prompts_immutable BEFORE UPDATE OR DELETE ON processing_prompt_versions FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
CREATE TRIGGER processing_prices_immutable BEFORE UPDATE OR DELETE ON processing_price_versions FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
CREATE TRIGGER processing_rates_immutable BEFORE UPDATE OR DELETE ON processing_price_rates FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
CREATE TRIGGER processing_configs_immutable BEFORE UPDATE OR DELETE ON processing_configuration_versions FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
CREATE TRIGGER processing_runs_immutable BEFORE UPDATE OR DELETE ON processing_runs FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();

CREATE FUNCTION processing_attempt_once() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF OLD.finished_at IS NOT NULL THEN RAISE EXCEPTION 'processing attempt is already final'; END IF;
 IF NEW.run_id <> OLD.run_id OR NEW.number <> OLD.number OR NEW.started_at <> OLD.started_at
    OR NEW.queue_job_id IS DISTINCT FROM OLD.queue_job_id
    OR NEW.queue_attempt_number IS DISTINCT FROM OLD.queue_attempt_number THEN
   RAISE EXCEPTION 'processing attempt identity is immutable';
 END IF;
 RETURN NEW;
END;
$$;
CREATE TRIGGER processing_attempt_once BEFORE UPDATE ON processing_run_attempts FOR EACH ROW EXECUTE FUNCTION processing_attempt_once();
CREATE TRIGGER processing_attempt_no_delete BEFORE DELETE ON processing_run_attempts FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
