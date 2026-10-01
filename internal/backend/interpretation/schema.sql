CREATE TABLE interpretation_triggers (
 document_version_id bigint PRIMARY KEY REFERENCES retained_versions(id), evidence_hash text NOT NULL CHECK(evidence_hash~'^[0-9a-f]{64}$'), reason text NOT NULL CHECK(reason IN('content_changed','dependency_changed')),
 workload text NOT NULL CHECK(workload IN('ordinary','bootstrap','evaluation')), classification_job_id bigint REFERENCES processing_jobs(id), created_at timestamptz NOT NULL);
CREATE TABLE interpretation_reprocessing_requests (
 id text PRIMARY KEY, source_id text REFERENCES registry_sources(id), acquired_from timestamptz, acquired_through timestamptz, error_code text, actor text NOT NULL, created_at timestamptz NOT NULL,
 CHECK(acquired_through IS NULL OR acquired_from IS NULL OR acquired_through>=acquired_from));
CREATE TABLE interpretation_reprocessing_selections (
 request_id text NOT NULL REFERENCES interpretation_reprocessing_requests(id), document_version_id bigint NOT NULL REFERENCES retained_versions(id), classification_job_id bigint REFERENCES processing_jobs(id), PRIMARY KEY(request_id,document_version_id));
