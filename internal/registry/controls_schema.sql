ALTER TABLE registry_sources
 ADD COLUMN check_seconds_override integer CHECK (check_seconds_override > 0),
 ADD COLUMN delay_seconds_override integer CHECK (delay_seconds_override > 0),
 ADD COLUMN interval_revision integer NOT NULL DEFAULT 0;
ALTER TABLE registry_events DROP CONSTRAINT registry_events_kind_check;
ALTER TABLE registry_events ADD CHECK (kind IN ('preview','acceptance','collection_enabled','collection_disabled','public_enabled','public_disabled','intervals_changed'));
