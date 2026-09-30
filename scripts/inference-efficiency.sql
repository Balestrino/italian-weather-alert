-- Parameters: $1 since inclusive, $2 cutoff exclusive, $3 source, $4 model, $5 stage.
-- Execute in a repeatable-read, read-only transaction with UTC timezone.
WITH scope AS (
 SELECT r.*,m.model FROM processing_runs r
 JOIN processing_configuration_versions cfg ON cfg.id=r.configuration_version_id
 LEFT JOIN processing_model_versions m ON m.id=cfg.model_version_id
 WHERE ($3::text='' OR r.source_id=$3) AND ($5::text='' OR r.stage=$5)
), calls AS (
 SELECT c.*,r.source_id,r.stage,
 c.finished_at<$2::timestamptz AS known_at_cutoff
 FROM processing_provider_calls c JOIN scope r ON r.id=c.run_id
 WHERE c.started_at >= $1::timestamptz AND c.started_at<$2
 AND ($4::text='' OR c.requested_model=$4)
), actual AS (
 SELECT source_id,stage,requested_model AS model,count(*) AS calls,
 count(*) FILTER(WHERE NOT coalesce(known_at_cutoff,false) OR usage_status<>'reported') AS unknown_or_partial_calls,
 count(*) FILTER(WHERE known_at_cutoff AND state IN('failed','uncertain')) AS failed_calls,
 sum(input_tokens) FILTER(WHERE known_at_cutoff) AS known_input_tokens,
 sum(output_tokens) FILTER(WHERE known_at_cutoff) AS known_output_tokens,
 sum(cache_read_tokens) FILTER(WHERE known_at_cutoff) AS known_cache_read_tokens
 FROM calls GROUP BY 1,2,3 ORDER BY 1,2,3
), legacy AS (
 SELECT r.source_id,r.stage,r.model,count(*) AS attempts,
 sum(a.input_tokens) FILTER(WHERE a.finished_at<$2) AS known_input_tokens,
 sum(a.output_tokens) FILTER(WHERE a.finished_at<$2) AS known_output_tokens,
 count(*) FILTER(WHERE a.finished_at IS NULL OR a.finished_at>=$2 OR a.usage_status IN('partial','unavailable')) AS unknown_or_partial_attempts
 FROM processing_run_attempts a JOIN scope r ON r.id=a.run_id
 WHERE a.started_at>=$1 AND a.started_at<$2 AND ($4='' OR r.model=$4)
 AND NOT EXISTS(SELECT 1 FROM processing_provider_calls c WHERE c.run_id=a.run_id AND c.attempt_number=a.number)
 AND coalesce(a.usage_status,'unavailable')<>'not_applicable'
 GROUP BY 1,2,3 ORDER BY 1,2,3
), avoided AS (
 SELECT r.source_id,r.stage,r.model,'ocr_page' AS kind,x.created_at,
 (a.result->>'InputTokens')::bigint AS input_tokens,(a.result->>'OutputTokens')::bigint AS output_tokens
 FROM ocr_artifact_associations x JOIN ocr_artifacts a USING(artifact_key) JOIN scope r ON r.id=x.run_id WHERE x.reused
 UNION ALL
 SELECT r.source_id,r.stage,r.model,'segment',u.created_at,
 (c.response->'Usage'->>'InputTokens')::bigint,(c.response->'Usage'->>'OutputTokens')::bigint
 FROM processing_segment_checkpoint_uses u JOIN processing_segment_checkpoints c USING(stage,configuration,input_sha256)
 JOIN scope r ON r.id=u.run_id WHERE u.reused
 UNION ALL
 SELECT r.source_id,r.stage,r.model,'equivalent_interpretation',u.created_at,
 (SELECT sum(a.input_tokens)::bigint FROM processing_run_attempts a WHERE a.run_id=u.original_run_id),
 (SELECT sum(a.output_tokens)::bigint FROM processing_run_attempts a WHERE a.run_id=u.original_run_id)
 FROM interpretation_reuse u JOIN scope r ON r.id=u.run_id
), reuse AS (
 SELECT source_id,stage,model,kind,count(*) AS reused_units,
 sum(input_tokens) AS estimated_avoided_input_tokens,sum(output_tokens) AS estimated_avoided_output_tokens,
 count(*) FILTER(WHERE input_tokens IS NULL OR output_tokens IS NULL) AS estimates_unavailable
 FROM avoided WHERE created_at>=$1 AND created_at<$2 AND ($4='' OR model=$4)
 GROUP BY 1,2,3,4 ORDER BY 1,2,3,4
), skips AS (
 SELECT r.source_id,r.stage,r.model,count(*) AS skipped_resources
 FROM ocr_resource_results o JOIN scope r ON r.id=o.run_id
 WHERE o.status='skipped' AND o.created_at>=$1 AND o.created_at<$2 AND ($4='' OR r.model=$4)
 GROUP BY 1,2,3 ORDER BY 1,2,3
), invalid AS (
 SELECT r.source_id,r.stage,r.model,a.error_code,count(*) AS attempts
 FROM processing_run_attempts a JOIN scope r ON r.id=a.run_id
 WHERE a.finished_at>=$1 AND a.finished_at<$2 AND ($4='' OR r.model=$4)
 AND (a.error_code LIKE '%output_%' OR a.error_code LIKE '%evidence_invalid')
 GROUP BY 1,2,3,4 ORDER BY 1,2,3,4
), rejected AS (
 SELECT source_id,stage,requested_model AS model,http_status,category,provider_code,count(*) AS calls
 FROM calls WHERE known_at_cutoff AND state<>'received' GROUP BY 1,2,3,4,5,6 ORDER BY 1,2,3,4,5,6
), holds AS (
 SELECT scope,model,state,reason,available_at,updated_at FROM processing_provider_gates
 WHERE state<>'closed' AND ($4='' OR model=$4 OR model='*') ORDER BY scope,model
)
SELECT jsonb_build_object('format_version',1,'since',$1::timestamptz,'cutoff',$2::timestamptz,
 'filters',jsonb_build_object('source',$3::text,'model',$4::text,'stage',$5::text),
 'actual_calls',coalesce((SELECT jsonb_agg(to_jsonb(x)) FROM actual x),'[]'::jsonb),
 'legacy_attempt_usage',coalesce((SELECT jsonb_agg(to_jsonb(x)) FROM legacy x),'[]'::jsonb),
 'reuse_estimates',coalesce((SELECT jsonb_agg(to_jsonb(x)) FROM reuse x),'[]'::jsonb),
 'skipped_resources',coalesce((SELECT jsonb_agg(to_jsonb(x)) FROM skips x),'[]'::jsonb),
 'invalid_outputs',coalesce((SELECT jsonb_agg(to_jsonb(x)) FROM invalid x),'[]'::jsonb),
 'provider_rejections',coalesce((SELECT jsonb_agg(to_jsonb(x)) FROM rejected x),'[]'::jsonb),
 'provider_holds_current_global',coalesce((SELECT jsonb_agg(to_jsonb(x)) FROM holds x),'[]'::jsonb),
 'limitations',jsonb_build_array('Known tokens are lower bounds when consumption is unknown. Legacy attempts have no verified call count.',
 'Avoided tokens are estimates from original observations, not verified billing savings. Categories count different units; do not sum their unit counts.',
 'Whole-interpretation estimates include original attempts; OCR and segment reuse are reported separately. No historical duplicate-opportunity estimate is added.',
 'Legacy checkpoint associations without an observed use timestamp are excluded from time-scoped estimates.',
 'Provider holds are current global state, filtered only by model; source, stage and historical time filters cannot reconstruct account-wide gate state.',
 'Time filters use call/attempt start for consumption, reuse time for avoided work, and completion time for invalid outputs. Receipts completed after cutoff are unknown.'));
