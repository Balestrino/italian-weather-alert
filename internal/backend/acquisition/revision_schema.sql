CREATE TABLE acquisition_bootstrap (
 source_id text PRIMARY KEY REFERENCES registry_sources(id),
 configuration integer NOT NULL,
 completed_at timestamptz,
 FOREIGN KEY(source_id,configuration) REFERENCES registry_configurations(source_id,revision)
);

ALTER TABLE acquisition_checks ADD COLUMN revision_count integer NOT NULL DEFAULT 0 CHECK (revision_count >= 0);

CREATE TABLE acquisition_targets (
 source_id text NOT NULL REFERENCES registry_sources(id),
 url text NOT NULL,
 configuration integer NOT NULL,
 source_publication_date date,
 first_discovered_at timestamptz NOT NULL,
 last_seen_at timestamptz NOT NULL,
 last_checked_at timestamptz,
 last_version_id bigint,
 explicit_reference boolean NOT NULL DEFAULT false,
 relevance_state text NOT NULL DEFAULT 'unknown' CHECK (relevance_state IN ('unknown','irrelevant','relevant')),
 measure_state text NOT NULL DEFAULT 'none' CHECK (measure_state IN ('none','ongoing','unresolved','ceased')),
 PRIMARY KEY(source_id,url),
 FOREIGN KEY(source_id,configuration) REFERENCES registry_configurations(source_id,revision)
);
CREATE INDEX acquisition_targets_plan ON acquisition_targets(source_id,source_publication_date,measure_state,url);

CREATE TABLE acquisition_dependencies (
 source_id text NOT NULL,
 parent_url text NOT NULL,
 resource_url text NOT NULL,
 required boolean NOT NULL,
 first_recorded_at timestamptz NOT NULL,
 PRIMARY KEY(source_id,parent_url,resource_url),
 FOREIGN KEY(source_id,parent_url) REFERENCES acquisition_targets(source_id,url) ON DELETE CASCADE
);
