CREATE TABLE retained_documents (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 source_id text NOT NULL REFERENCES registry_sources(id),
 official_url text NOT NULL,
 UNIQUE(source_id,official_url)
);
CREATE TABLE retained_versions (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 document_id bigint NOT NULL REFERENCES retained_documents(id),
 content_hash text NOT NULL CHECK(content_hash ~ '^[0-9a-f]{64}$'),
 issuer_id text REFERENCES registry_authorities(id),
 first_acquired_at timestamptz NOT NULL,
 complete boolean NOT NULL,
 metadata jsonb NOT NULL,
 UNIQUE(document_id,content_hash)
);
CREATE TABLE retained_objects (
 hash text PRIMARY KEY CHECK(hash ~ '^[0-9a-f]{64}$'),
 object_key text NOT NULL UNIQUE,
 byte_size bigint NOT NULL CHECK(byte_size >= 0)
);
CREATE TABLE retained_resources (
 version_id bigint NOT NULL REFERENCES retained_versions(id),
 url text NOT NULL,
 role text NOT NULL CHECK(role IN ('original','attachment','resource')),
 required boolean NOT NULL,
 source_id text NOT NULL,
 configuration integer NOT NULL,
 media_type text NOT NULL,
 object_hash text REFERENCES retained_objects(hash),
 missing text NOT NULL,
 PRIMARY KEY(version_id,url),
 FOREIGN KEY(source_id,configuration) REFERENCES registry_configurations(source_id,revision),
 CHECK ((object_hash IS NOT NULL AND missing='') OR (object_hash IS NULL AND missing IN ('unavailable','forbidden','not_acquired')))
);
-- Durable staging bridges PostgreSQL and S3 without pretending a distributed
-- transaction exists. Payload is cleared only in the version commit.
CREATE TABLE retained_acquisitions (
 id text PRIMARY KEY,
 document_id bigint NOT NULL REFERENCES retained_documents(id),
 source_id text NOT NULL,
 configuration integer NOT NULL,
 request_hash text NOT NULL,
 content_hash text NOT NULL,
 acquired_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 payload jsonb,
 version_id bigint REFERENCES retained_versions(id),
 FOREIGN KEY(source_id,configuration) REFERENCES registry_configurations(source_id,revision),
 CHECK ((payload IS NOT NULL AND version_id IS NULL) OR (payload IS NULL AND version_id IS NOT NULL))
);
CREATE INDEX retained_pending ON retained_acquisitions(acquired_at,id) WHERE version_id IS NULL;
CREATE INDEX retained_acquisitions_document ON retained_acquisitions(document_id,acquired_at);
