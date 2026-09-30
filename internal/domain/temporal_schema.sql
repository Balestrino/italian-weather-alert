CREATE TABLE domain_measure_bindings (
 local_measure_id text PRIMARY KEY REFERENCES domain_local_measures(id),
 extraction_run_id bigint NOT NULL,
 extraction_ordinal integer NOT NULL,
 UNIQUE(extraction_run_id,extraction_ordinal),
 FOREIGN KEY(extraction_run_id,extraction_ordinal) REFERENCES extracted_measures(run_id,ordinal)
);

CREATE TABLE domain_measure_updates (
 linking_run_id bigint PRIMARY KEY REFERENCES linking_results(run_id),
 update_measure_id text NOT NULL REFERENCES domain_local_measures(id),
 target_measure_id text NOT NULL REFERENCES domain_local_measures(id),
 relation text NOT NULL CHECK (relation IN ('reopens','cancels','extends','corrects','updates')),
 applied_at timestamptz NOT NULL,
 CHECK (update_measure_id <> target_measure_id)
);
CREATE INDEX domain_measure_updates_target ON domain_measure_updates(target_measure_id,applied_at);

CREATE TABLE domain_temporal_values (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 entity_kind text NOT NULL CHECK (entity_kind IN ('local_measure','operational_phase','regional_record')),
 entity_id text NOT NULL CHECK (entity_id <> ''),
 meaning text NOT NULL CHECK (meaning IN ('publication','source_modification','event','validity','page_expiry','platform','acquisition','attestation')),
 original_expression text,
 precision text NOT NULL CHECK (precision IN ('instant','date','interval','conditional','unknown')),
 instant timestamptz,
 date_value date,
 end_instant timestamptz,
 timezone text,
 assumption text,
 condition text,
 conflict_group text,
 evidence_document_version_id bigint NOT NULL REFERENCES retained_versions(id),
 created_at timestamptz NOT NULL,
 CHECK (original_expression IS NOT NULL OR precision='unknown'),
 CHECK ((timezone IS NULL AND assumption IS NULL) OR (timezone IS NOT NULL AND assumption IS NOT NULL AND assumption <> '')),
 CHECK (timezone IS NULL OR precision IN ('instant','interval','conditional')),
 CHECK (
   (precision='instant' AND instant IS NOT NULL AND date_value IS NULL AND end_instant IS NULL AND condition IS NULL) OR
   (precision='date' AND instant IS NULL AND date_value IS NOT NULL AND end_instant IS NULL AND timezone IS NULL AND condition IS NULL) OR
   (precision='interval' AND instant IS NOT NULL AND date_value IS NULL AND end_instant IS NOT NULL AND end_instant >= instant AND condition IS NULL) OR
   (precision='conditional' AND date_value IS NULL AND end_instant IS NULL AND condition IS NOT NULL AND condition <> '') OR
   (precision='unknown' AND instant IS NULL AND date_value IS NULL AND end_instant IS NULL AND timezone IS NULL)
 )
);
CREATE INDEX domain_temporal_entity ON domain_temporal_values(entity_kind,entity_id,id);

CREATE TRIGGER domain_measure_bindings_immutable BEFORE UPDATE OR DELETE ON domain_measure_bindings FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE TRIGGER domain_measure_updates_immutable BEFORE UPDATE OR DELETE ON domain_measure_updates FOR EACH ROW EXECUTE FUNCTION domain_immutable();
CREATE TRIGGER domain_temporal_values_immutable BEFORE UPDATE OR DELETE ON domain_temporal_values FOR EACH ROW EXECUTE FUNCTION domain_immutable();
