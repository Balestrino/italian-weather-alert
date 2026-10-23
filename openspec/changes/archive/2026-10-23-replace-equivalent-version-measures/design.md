# Design

## Context

- `publicquery.measures()` selects `domain_local_measures` visible at the boundary. It excludes a measure `m` when, for the same `document_version_id`, a later `extraction_results` row (`run_id` greater than the run bound to `m`, status `extracted`, complete) has bindings to measures of the same source recorded within the source cutoff. The selection feeds measure search and the municipality situation. Measure detail reads by identifier do not use it.
- `interpretation_input_equivalences` holds a configuration-independent fingerprint per reusable manifest, indexed on `(document_id, fingerprint, version_id)`. Every classification and extraction run writes one, and existing manifests are backfilled.
- On development, documents 60 and 58 of Calcinaia have equivalent versions whose measures are all current. Documents 43, 15668 and 34667 have genuine text revisions.

## Goals / Non-Goals

**Goals:** extend the replacement predicate to equivalent versions, with bounded cost, the same knowledge-boundary semantics and no writes.

**Non-Goals:**
- deduplicating genuine revisions;
- changing linking or `tracked-target-interpretation-state`, which keeps its conservative union of all versions;
- statuses of measures read by explicit identifier;
- merging the measure identities of the replaced and replacing projections.

## Decisions

1. **One predicate with the same shape.** The current `NOT EXISTS` becomes:

   ```text
   m is replaced if there exist
     prior binding of m (prior.extraction_run_id) and its version A,
     replacement extraction R on version B,
     newer binding of R to a measure recorded within the cutoff,
   such that R is extracted and complete, R's run is not evaluation, and either:
     B = A and R.run_id > prior.extraction_run_id                       (unchanged rule), or
     B is a different version of the same document, acquired after A,
     and B and A share a fingerprint whose rows come from runs created within the cutoff.
   ```

   Ordering uses `(first_acquired_at, id)` of the versions, as in carry-forward. The existing exclusion of evaluation runs and the newer-binding requirement are kept for both branches.

   *Alternative:* collapse duplicates by kind, subject and place across versions. It was rejected because model wording varies ("cimiteri comunali (Calcinaia e Fornacette)" against "cimiteri comunali"), and it would compare facts instead of inputs.

   *Alternative:* select only the latest version's measures for every document. It was rejected because it would make disappearance in genuine revisions end measures, which violates the cessation rule.
2. **Fingerprint knowledge boundary.** Both fingerprints count only if their manifest runs were created within the cutoff, as in carry-forward. The replacing binding already has to be recorded within the cutoff.
3. **No change to the newer-uninterpreted warnings or to statuses** (`superseded`, `relationship_status`). Replaced measures simply leave the current selection, as same-version replacement does today.
4. **Query cost.** The equivalent branch is computed once per query as a set of replaced `(version, source)` pairs (materialized CTEs), not as a per-measure `NOT EXISTS`. The CTEs cover only versions that have measures, in the sources of the requested municipality. The publicquery latency and query-count checks are repeated on a municipality with many template-only versions.

   *Measured during implementation (23 October 2026, development data):* the per-measure form fanned out to every equivalent template version of a document, about 40 for Livorno. It cost 659 ms for Livorno's 606 measures and 23 ms for Calcinaia. A variant starting from later versions cost 1,015 ms. The set-based form computes the replaced versions of every municipality in 70 ms, with identical counts per municipality. Only 224 versions have measures.

## Risks / Trade-offs

- [The replacing interpretation is worse than the earlier one] → This is the same trade-off already accepted within one version: the latest validated interpretation of the same text wins. Earlier boundaries still show the earlier measures. A failed or not-relevant newer result does not replace anything.
- [A replaced measure is the target of a linking relation] → Relations stay retained. A later measure can still reference the replaced one; only its presence in the current selection changes.
- [Tracked-target state still counts replaced measures] → This is intentional and conservative: a target may stay `unresolved` longer. The behaviour is documented in the operations guide.

## Migration Plan

No migration. Deploy the public service, then verify the Calcinaia documents 60 and 58 and the selection at earlier boundaries on development. Roll back by restoring the previous public image.
