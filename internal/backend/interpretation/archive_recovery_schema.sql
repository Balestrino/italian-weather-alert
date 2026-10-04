CREATE TABLE interpretation_archive_recoveries (
 request_id text PRIMARY KEY REFERENCES interpretation_reprocessing_requests(id),
 revision integer NOT NULL CHECK(revision>0),
 regression_id text NOT NULL REFERENCES registry_regressions(id),
 evidence jsonb NOT NULL CHECK(jsonb_typeof(evidence)='object'),
 selection_hash text NOT NULL CHECK(selection_hash~'^[0-9a-f]{64}$')
);
CREATE TABLE interpretation_archive_recovery_versions (
 request_id text NOT NULL REFERENCES interpretation_archive_recoveries(request_id),
 document_version_id bigint NOT NULL REFERENCES interpretation_archives(document_version_id),
 PRIMARY KEY(request_id,document_version_id),
 FOREIGN KEY(request_id,document_version_id) REFERENCES interpretation_reprocessing_selections(request_id,document_version_id)
);
CREATE TRIGGER interpretation_archive_recoveries_frozen BEFORE UPDATE OR DELETE ON interpretation_archive_recoveries
 FOR EACH ROW EXECUTE FUNCTION interpretation_archive_frozen();
CREATE TRIGGER interpretation_archive_recovery_versions_frozen BEFORE UPDATE OR DELETE ON interpretation_archive_recovery_versions
 FOR EACH ROW EXECUTE FUNCTION interpretation_archive_frozen();
CREATE FUNCTION interpretation_archive_recovery_request_frozen() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF EXISTS(SELECT 1 FROM interpretation_archive_recoveries WHERE request_id=OLD.id) THEN
  RAISE EXCEPTION 'archive recovery request is immutable';
 END IF;
 IF TG_OP='DELETE' THEN RETURN OLD; END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER interpretation_archive_recovery_request_frozen BEFORE UPDATE OR DELETE ON interpretation_reprocessing_requests
 FOR EACH ROW EXECUTE FUNCTION interpretation_archive_recovery_request_frozen();
