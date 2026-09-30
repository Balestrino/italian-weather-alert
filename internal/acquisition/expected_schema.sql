ALTER TABLE acquisition_source_status
 ADD COLUMN expected_anchor timestamptz,
 ADD COLUMN expected_interval_seconds integer,
 ADD COLUMN expected_tolerance_seconds integer,
 ADD COLUMN last_publication_at timestamptz,
 ADD COLUMN last_check_state text NOT NULL DEFAULT 'never' CHECK (last_check_state IN ('never','complete_changed','complete_unchanged','collection_failure')),
 ADD COLUMN publication_state text NOT NULL DEFAULT 'not_expected' CHECK (publication_state IN ('not_expected','awaiting','observed','missing')),
 ADD CHECK ((expected_anchor IS NULL) = (expected_interval_seconds IS NULL)),
 ADD CHECK ((expected_anchor IS NULL) = (expected_tolerance_seconds IS NULL)),
 ADD CHECK (expected_interval_seconds IS NULL OR expected_interval_seconds >= 60),
 ADD CHECK (expected_tolerance_seconds IS NULL OR (expected_tolerance_seconds >= 0 AND expected_tolerance_seconds <= expected_interval_seconds));

ALTER TABLE acquisition_checks
 ADD COLUMN new_document_count integer NOT NULL DEFAULT 0 CHECK (new_document_count >= 0),
 ADD COLUMN publication_observed_at timestamptz,
 ADD COLUMN check_state text NOT NULL DEFAULT 'collection_failure' CHECK (check_state IN ('complete_changed','complete_unchanged','collection_failure')),
 ADD COLUMN publication_state text NOT NULL DEFAULT 'not_expected' CHECK (publication_state IN ('not_expected','awaiting','observed','missing'));
