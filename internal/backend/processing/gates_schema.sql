CREATE TABLE processing_provider_gates (
 scope text NOT NULL,
 model text NOT NULL,
 state text NOT NULL DEFAULT 'closed' CHECK (state IN ('closed','open','held')),
 reason text NOT NULL DEFAULT '',
 failures integer NOT NULL DEFAULT 0 CHECK (failures >= 0),
 generation bigint NOT NULL DEFAULT 1,
 cooldown_ms bigint NOT NULL DEFAULT 0 CHECK (cooldown_ms >= 0),
 available_at timestamptz NOT NULL,
 probe_token text NOT NULL DEFAULT '',
 probe_expires_at timestamptz,
 updated_at timestamptz NOT NULL,
 PRIMARY KEY(scope,model)
);
CREATE TABLE processing_provider_gate_events (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 scope text NOT NULL,model text NOT NULL,action text NOT NULL,reason text NOT NULL,
 actor text NOT NULL,created_at timestamptz NOT NULL
);
