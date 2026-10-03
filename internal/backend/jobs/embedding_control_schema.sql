CREATE TABLE processing_embedding_control (
    singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    enabled boolean NOT NULL DEFAULT false,
    revision integer NOT NULL DEFAULT 0 CHECK (revision >= 0),
    actor text NOT NULL DEFAULT 'migration',
    updated_at timestamptz NOT NULL DEFAULT now()
);
INSERT INTO processing_embedding_control(singleton) VALUES (true);
CREATE TABLE processing_embedding_control_events (
    revision integer NOT NULL CHECK (revision > 0),
    enabled boolean NOT NULL,
    actor text NOT NULL CHECK (length(btrim(actor)) > 0),
    created_at timestamptz NOT NULL,
    source_id text NOT NULL DEFAULT '',
    PRIMARY KEY(source_id,revision)
);
CREATE TABLE processing_source_embedding_controls (
    source_id text PRIMARY KEY,
    enabled boolean NOT NULL DEFAULT false,
    revision integer NOT NULL CHECK (revision > 0),
    actor text NOT NULL CHECK (length(btrim(actor)) > 0),
    updated_at timestamptz NOT NULL
);
-- The queue can be migrated/tested before the processing ledger exists.
-- Resolve historical embedding payloads through their retained extraction run.
CREATE FUNCTION processing_embedding_source_allowed(payload jsonb) RETURNS boolean
LANGUAGE plpgsql STABLE AS $$
DECLARE allowed boolean;
BEGIN
    IF to_regclass('processing_runs') IS NULL THEN RETURN false; END IF;
    EXECUTE 'SELECT EXISTS(SELECT 1 FROM processing_runs r
        JOIN processing_source_embedding_controls c ON c.source_id=r.source_id
        WHERE r.id=($1->>''extraction_run_id'')::bigint AND r.stage=''extraction'' AND c.enabled)'
        INTO allowed USING payload;
    RETURN allowed;
END $$;
