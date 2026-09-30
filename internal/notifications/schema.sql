CREATE TABLE notification_source_watches (
 source_id text PRIMARY KEY REFERENCES registry_sources(id),
 configuration integer NOT NULL,
 first_seen_at timestamptz NOT NULL
);
CREATE TABLE notification_backup_results (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 scope text NOT NULL,
 run_id text NOT NULL,
 succeeded boolean NOT NULL,
 finished_at timestamptz NOT NULL,
 UNIQUE(scope,run_id)
);
CREATE INDEX notification_backup_latest ON notification_backup_results(scope,finished_at DESC,id DESC);
CREATE TABLE notification_incidents (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 category text NOT NULL CHECK (category IN ('source_delay','processing_failed','backup_failed')),
 scope text NOT NULL,
 opened_at timestamptz NOT NULL,
 last_observed_at timestamptz NOT NULL,
 recovered_at timestamptz,
 last_notified_at timestamptz
);
CREATE UNIQUE INDEX notification_open_incident ON notification_incidents(category,scope) WHERE recovered_at IS NULL;
CREATE TABLE notification_messages (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 message_id text NOT NULL UNIQUE,
 incident_id bigint NOT NULL REFERENCES notification_incidents(id),
 kind text NOT NULL CHECK (kind IN ('opened','reminder','recovered')),
 created_at timestamptz NOT NULL,
 sent_at timestamptz,
 attempts integer NOT NULL DEFAULT 0,
 next_attempt_at timestamptz NOT NULL,
 last_error_code text
);
CREATE INDEX notification_delivery ON notification_messages(next_attempt_at,id) WHERE sent_at IS NULL;
