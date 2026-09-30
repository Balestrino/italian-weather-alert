-- Operator-selected live recovery. Empty table preserves normal operation.
-- Limits count durable call intents, including unknown/interrupted receipts.
CREATE TABLE processing_recovery_limits (
 singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
 version_ids bigint[] NOT NULL CHECK (cardinality(version_ids) BETWEEN 1 AND 100),
 max_calls integer NOT NULL CHECK (max_calls BETWEEN 1 AND 1000),
 started_at timestamptz NOT NULL,
 actor text NOT NULL CHECK (actor <> '')
);
