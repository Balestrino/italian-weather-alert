CREATE TABLE measure_embeddings (
 run_id bigint PRIMARY KEY REFERENCES processing_runs(id), extraction_run_id bigint NOT NULL, measure_ordinal integer NOT NULL,
 configuration_version_id text NOT NULL REFERENCES processing_configuration_versions(id), dimensions integer NOT NULL CHECK(dimensions>0), vector real[] NOT NULL,
 returned_model text NOT NULL, input_tokens bigint CHECK(input_tokens>=0), created_at timestamptz NOT NULL,
 FOREIGN KEY(extraction_run_id,measure_ordinal) REFERENCES extracted_measures(run_id,ordinal), UNIQUE(extraction_run_id,measure_ordinal,configuration_version_id),
 CHECK(array_length(vector,1)=dimensions));
CREATE TRIGGER measure_embeddings_immutable BEFORE UPDATE OR DELETE ON measure_embeddings FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
ALTER TABLE linking_candidates DROP CONSTRAINT linking_candidates_retrieval_method_check;
ALTER TABLE linking_candidates ADD CONSTRAINT linking_candidates_retrieval_method_check CHECK(retrieval_method IN ('baseline_text','semantic'));
ALTER TABLE linking_candidates ADD COLUMN semantic_score real CHECK(semantic_score IS NULL OR semantic_score BETWEEN -1 AND 1);
