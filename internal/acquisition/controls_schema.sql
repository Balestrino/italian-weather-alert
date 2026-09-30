-- Runtime interval edits take effect in the same transaction as their audit
-- event, without replacing the active configuration, leases or check evidence.
CREATE FUNCTION acquisition_apply_intervals() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 UPDATE acquisition_source_status SET
   check_seconds=NEW.check_seconds_override,
   delay_seconds=NEW.delay_seconds_override,
   next_check_at=CASE
     WHEN check_seconds<>NEW.check_seconds_override AND consecutive_failures=0
       AND claim_token IS NULL AND last_complete_at IS NOT NULL
     THEN last_complete_at + NEW.check_seconds_override * interval '1 second'
     ELSE next_check_at END
 WHERE source_id=NEW.id;
 RETURN NEW;
END;
$$;
CREATE TRIGGER acquisition_interval_edit AFTER UPDATE OF check_seconds_override,delay_seconds_override
 ON registry_sources FOR EACH ROW EXECUTE FUNCTION acquisition_apply_intervals();
