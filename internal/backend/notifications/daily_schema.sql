CREATE TABLE notification_daily_reports (
 report_date date PRIMARY KEY,
 timezone text NOT NULL,
 local_time text NOT NULL,
 message_id text NOT NULL UNIQUE,
 observed_at timestamptz NOT NULL,
 subject text,
 body text CHECK(body IS NULL OR octet_length(body)<=262144),
 generation_attempts integer NOT NULL DEFAULT 0,
 delivery_attempts integer NOT NULL DEFAULT 0,
 next_attempt_at timestamptz NOT NULL,
 last_error_code text,
 sent_at timestamptz,
 CHECK ((subject IS NULL) = (body IS NULL)),
 CHECK (sent_at IS NULL OR body IS NOT NULL)
);
CREATE INDEX notification_daily_reports_due ON notification_daily_reports(next_attempt_at) WHERE sent_at IS NULL;
