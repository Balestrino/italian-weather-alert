CREATE TABLE ocr_render_manifests (
 manifest_key text PRIMARY KEY CHECK (manifest_key ~ '^[0-9a-f]{64}$'),
 pages jsonb NOT NULL CHECK (jsonb_typeof(pages)='array'),
 created_at timestamptz NOT NULL
);
CREATE TRIGGER ocr_render_manifests_immutable BEFORE UPDATE OR DELETE ON ocr_render_manifests FOR EACH ROW EXECUTE FUNCTION ocr_results_immutable();
