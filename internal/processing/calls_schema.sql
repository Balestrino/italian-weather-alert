CREATE TABLE processing_provider_calls (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 run_id bigint NOT NULL,
 attempt_number integer NOT NULL,
 ordinal integer NOT NULL CHECK (ordinal > 0),
 input_sha256 text NOT NULL CHECK (input_sha256 ~ '^[0-9a-f]{64}$'),
 provider text NOT NULL CHECK (length(provider) BETWEEN 1 AND 200),
 requested_model text NOT NULL CHECK (length(requested_model) BETWEEN 1 AND 200),
 started_at timestamptz NOT NULL,
 finished_at timestamptz,
 state text NOT NULL DEFAULT 'started' CHECK (state IN ('started','received','failed','uncertain')),
 http_status integer CHECK (http_status BETWEEN 100 AND 599),
 error_code text NOT NULL DEFAULT '',
 category text NOT NULL DEFAULT '' CHECK (category IN ('','authentication','permission','quota','rate_limit','request','availability','rejected','transport','invalid_response')),
 provider_code text NOT NULL DEFAULT '',
 request_id text NOT NULL DEFAULT '',
 response_id text NOT NULL DEFAULT '',
 returned_model text NOT NULL DEFAULT '',
 finish_reason text NOT NULL DEFAULT '',
 input_tokens bigint CHECK (input_tokens >= 0),
 output_tokens bigint CHECK (output_tokens >= 0),
 cache_read_tokens bigint CHECK (cache_read_tokens >= 0),
 usage_status text NOT NULL DEFAULT 'unavailable' CHECK (usage_status IN ('reported','partial','unavailable')),
 retry_after_ms bigint NOT NULL DEFAULT 0 CHECK (retry_after_ms >= 0),
 UNIQUE(run_id,attempt_number,ordinal),
 FOREIGN KEY(run_id,attempt_number) REFERENCES processing_run_attempts(run_id,number),
 CHECK ((state='started') = (finished_at IS NULL)),
 CHECK (finished_at IS NULL OR finished_at >= started_at),
 CHECK (usage_status <> 'unavailable' OR (input_tokens IS NULL AND output_tokens IS NULL AND cache_read_tokens IS NULL))
);
CREATE INDEX processing_calls_started ON processing_provider_calls(started_at);
CREATE FUNCTION processing_call_once() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF OLD.state <> 'started' OR NEW.run_id <> OLD.run_id OR NEW.attempt_number <> OLD.attempt_number
 OR NEW.ordinal <> OLD.ordinal OR NEW.input_sha256 <> OLD.input_sha256 OR NEW.provider <> OLD.provider
 OR NEW.requested_model <> OLD.requested_model OR NEW.started_at <> OLD.started_at THEN
 RAISE EXCEPTION 'provider call identity and finished receipt are immutable';
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER processing_call_once BEFORE UPDATE ON processing_provider_calls FOR EACH ROW EXECUTE FUNCTION processing_call_once();
CREATE TRIGGER processing_call_no_delete BEFORE DELETE ON processing_provider_calls FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
