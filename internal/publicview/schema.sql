CREATE TABLE public_query_views (
    id text PRIMARY KEY,
    operation text NOT NULL CHECK (operation IN ('discover_municipalities','get_municipality_situation','search_alerts','get_document','get_source_coverage')),
    request_hash text NOT NULL CHECK (length(request_hash) = 64),
    request jsonb NOT NULL CHECK (jsonb_typeof(request) = 'object'),
    data jsonb NOT NULL,
    history jsonb NOT NULL CHECK (jsonb_typeof(history) = 'object'),
    evaluation_time timestamptz NOT NULL,
    known_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL CHECK (expires_at > created_at)
);

CREATE INDEX public_query_views_expiry ON public_query_views(expires_at);
