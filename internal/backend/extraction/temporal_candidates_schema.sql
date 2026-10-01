CREATE TABLE extracted_temporal_candidates (
 run_id bigint NOT NULL,
 measure_ordinal integer NOT NULL,
 ordinal integer NOT NULL CHECK (ordinal > 0),
 field_name text NOT NULL CHECK (field_name IN ('valid_from','valid_until')),
 original_expression text NOT NULL CHECK (original_expression <> ''),
 conflict_identity text CHECK (conflict_identity <> ''),
 PRIMARY KEY(run_id,measure_ordinal,ordinal),
 FOREIGN KEY(run_id,measure_ordinal) REFERENCES extracted_measures(run_id,ordinal)
);

CREATE TABLE extraction_temporal_candidate_evidence (
 run_id bigint NOT NULL,
 measure_ordinal integer NOT NULL,
 candidate_ordinal integer NOT NULL,
 ordinal integer NOT NULL CHECK (ordinal > 0),
 evidence_ordinal integer NOT NULL CHECK (evidence_ordinal > 0),
 PRIMARY KEY(run_id,measure_ordinal,candidate_ordinal,ordinal),
 UNIQUE(run_id,measure_ordinal,candidate_ordinal,evidence_ordinal),
 FOREIGN KEY(run_id,measure_ordinal,candidate_ordinal)
  REFERENCES extracted_temporal_candidates(run_id,measure_ordinal,ordinal),
 FOREIGN KEY(run_id,measure_ordinal,evidence_ordinal)
  REFERENCES extraction_evidence(run_id,measure_ordinal,ordinal)
);

CREATE INDEX extracted_temporal_candidates_conflict
 ON extracted_temporal_candidates(run_id,measure_ordinal,field_name,conflict_identity);

CREATE TRIGGER extracted_temporal_candidates_immutable
 BEFORE UPDATE OR DELETE ON extracted_temporal_candidates
 FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
CREATE TRIGGER extraction_temporal_candidate_evidence_immutable
 BEFORE UPDATE OR DELETE ON extraction_temporal_candidate_evidence
 FOR EACH ROW EXECUTE FUNCTION processing_catalog_immutable();
