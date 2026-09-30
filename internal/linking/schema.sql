CREATE INDEX extracted_measures_text_search ON extracted_measures USING gin (to_tsvector('simple', subject || ' ' || COALESCE(place,'')));

CREATE TABLE linking_results (
 run_id bigint PRIMARY KEY REFERENCES processing_runs(id),
 current_extraction_run_id bigint NOT NULL,
 current_measure_ordinal integer NOT NULL,
 status text NOT NULL CHECK (status IN ('linked','unresolved','no_relation')),
 reason_code text NOT NULL CHECK (reason_code IN ('partial_reopening','cross_document_evidence','explicit_reference','ambiguous_relation','insufficient_evidence','no_candidates','output_invalid','provider_error','attempts_exhausted')),
 relation text CHECK (relation IN ('reopens','cancels','extends','corrects','updates')),
 candidate_extraction_run_id bigint,
 candidate_measure_ordinal integer,
 provider_response_id text,
 returned_model text,
 input_tokens bigint CHECK (input_tokens >= 0),
 output_tokens bigint CHECK (output_tokens >= 0),
 cache_read_tokens bigint CHECK (cache_read_tokens >= 0),
 created_at timestamptz NOT NULL,
 FOREIGN KEY(current_extraction_run_id,current_measure_ordinal) REFERENCES extracted_measures(run_id,ordinal),
 FOREIGN KEY(candidate_extraction_run_id,candidate_measure_ordinal) REFERENCES extracted_measures(run_id,ordinal),
 CHECK ((candidate_extraction_run_id IS NULL) = (candidate_measure_ordinal IS NULL)),
 CHECK ((status = 'linked') = (relation IS NOT NULL AND candidate_extraction_run_id IS NOT NULL)),
 CHECK (status <> 'linked' OR (provider_response_id IS NOT NULL AND returned_model IS NOT NULL))
);

CREATE TABLE linking_candidates (
 run_id bigint NOT NULL REFERENCES linking_results(run_id),
 ordinal integer NOT NULL CHECK (ordinal > 0),
 candidate_extraction_run_id bigint NOT NULL,
 candidate_measure_ordinal integer NOT NULL,
 retrieval_method text NOT NULL CHECK (retrieval_method = 'baseline_text'),
 text_rank real NOT NULL CHECK (text_rank >= 0),
 exact_place boolean NOT NULL,
 PRIMARY KEY(run_id,ordinal),
 FOREIGN KEY(candidate_extraction_run_id,candidate_measure_ordinal) REFERENCES extracted_measures(run_id,ordinal)
);

CREATE TABLE linking_evidence (
 run_id bigint NOT NULL REFERENCES linking_results(run_id),
 side text NOT NULL CHECK (side IN ('current','candidate')),
 ordinal integer NOT NULL CHECK (ordinal > 0),
 field_name text NOT NULL,
 resource_url text NOT NULL,
 page_number integer CHECK (page_number > 0),
 quote text NOT NULL,
 PRIMARY KEY(run_id,side,ordinal)
);

CREATE TRIGGER linking_results_immutable BEFORE UPDATE OR DELETE ON linking_results FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
CREATE TRIGGER linking_candidates_immutable BEFORE UPDATE OR DELETE ON linking_candidates FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
CREATE TRIGGER linking_evidence_immutable BEFORE UPDATE OR DELETE ON linking_evidence FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
