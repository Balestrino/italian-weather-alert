CREATE TABLE interpretation_archives (
 document_version_id bigint PRIMARY KEY REFERENCES retained_versions(id) ON DELETE CASCADE,
 archived_at timestamptz NOT NULL,
 cutoff timestamptz NOT NULL CHECK(cutoff<=archived_at),
 actor text NOT NULL CHECK(length(btrim(actor)) BETWEEN 1 AND 200)
);
CREATE FUNCTION interpretation_archive_frozen() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 RAISE EXCEPTION 'interpretation archive is immutable';
END $$;
CREATE TRIGGER interpretation_archive_frozen BEFORE UPDATE ON interpretation_archives
 FOR EACH ROW EXECUTE FUNCTION interpretation_archive_frozen();
