ALTER TABLE interpretation_reprocessing_requests
 ADD COLUMN processing_evaluation_id text,
 ADD COLUMN candidate_process_version text CHECK (candidate_process_version IS NULL OR candidate_process_version ~ '^[0-9a-f]{64}$'),
 ADD CONSTRAINT interpretation_reprocessing_evaluation_pair CHECK ((processing_evaluation_id IS NULL) = (candidate_process_version IS NULL));
