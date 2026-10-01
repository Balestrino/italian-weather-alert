CREATE TABLE processing_provider_rejections (
 scope text NOT NULL,model text NOT NULL,configuration_id text NOT NULL,
 input_sha256 text NOT NULL CHECK (input_sha256 ~ '^[0-9a-f]{64}$'),
 error_code text NOT NULL,http_status integer NOT NULL CHECK (http_status BETWEEN 100 AND 599),
 category text NOT NULL,created_at timestamptz NOT NULL,
 PRIMARY KEY(scope,model,configuration_id,input_sha256)
);
