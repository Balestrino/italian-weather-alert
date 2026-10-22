# Design

## Context

- Public document interpretation is computed per version in `publicquery.loadDocumentInterpretations`. The query takes the version's latest non-evaluation extraction result within the source's interpretation cutoff. Without one, it falls back to the latest classification, where not relevant counts as supported. It then derives `supported`, `partial`, `failed`, `unreliable` or `not_processed`. A batch path preloads many versions for search, document versions and saved views.
- Every classification and extraction run stores an immutable input manifest (`interpretation_input_manifests`, `interpretation-input-v2`). The manifest holds the interpreted sections, metadata, completeness, issuer and per-resource identity, and is keyed by the run's configuration. Its hash includes `Configuration` and each raw resource `Identity`, so it cannot match across configuration revisions or template-only HTML changes. `classification_results.content_sha256` includes the version ID and is per-version by construction.
- Reuse (`interpretation_reuse`) is deliberately same-configuration and creates new runs. This change does not touch it.
- Domain measures are read from any version, so earlier measures stay published. The gap is the document-level interpretation.
- On development (2026-10-22), the three affected notices have manifests whose `Sections`, `Metadata`, `Complete`, `IssuerID`, `SourceID` and `DocumentID` are identical across versions. Only `Configuration` and the HTML resource `Identity` differ.

## Goals / Non-Goals

**Goals:**
- A deterministic, configuration-independent equivalence fingerprint, persisted per manifest.
- Batched carry-forward in public document reads, respecting knowledge boundaries and the API/MCP equivalence.

**Non-Goals:**
- Reusing results across configurations in processing, or skipping reinterpretation.
- Changing domain measures, linking, projections or the admin views.
- Classifying pending versions that never recorded a manifest.

## Decisions

1. **Fingerprint derived from the manifest, not from raw retained bytes.** The manifest already records exactly what the model received: the interpreted text and the OCR page hashes. Fingerprint version `interpretation-equivalence-v1` is a SHA-256 of canonical JSON with these fields:
   - `SourceID`, `DocumentID`, `IssuerID`, `Metadata`, `Complete`, and `Sections` (URL, role, page, text, JSON scalars);
   - per resource: URL, role, media type, missing, required, eligible and policy;
   - for OCR or other non-text resources, `Identity` and pages (number, input, output, status, without the OCR model name);
   - HTML, text and JSON resources without `Identity`, because their contribution is fully represented by `Sections`.

   Only manifests marked `reusable` get a fingerprint.
   *Alternative:* rebuild the content from retained objects at query time. This was rejected as too slow and as a duplication of the OCR state.
   *Alternative:* reuse the `interpretation_preflight` fingerprint. It was rejected because it includes the configuration and is source-gated.
2. **Additive table `interpretation_input_equivalences(run_id PK → manifest, document_id, version_id, fingerprint, fingerprint_version, created_at)`**, immutable like manifests. New rows are written in the same transaction path as `PutManifest`. A bounded, idempotent `interpretation-equivalence-backfill` command fills existing manifests in pages. Manifests stay immutable, so a new column would break the immutability trigger.
3. **Carry-forward as a second batched step in `loadDocumentInterpretations`.** After the existing assessment query:
   - collect the IDs whose state is `failed` or `not_processed`;
   - run one query that joins their fingerprints to fingerprints of earlier versions of the same document (`first_acquired_at <= target`, manifest run within the cutoff at `$knownAt`), returning candidate version IDs ordered by recency;
   - assess the candidates with the existing assessment SQL in one more batch, and pick the most recent `supported` or `partial` one.

   This adds two round trips per batch, regardless of the number of versions. Unreliable candidates are skipped.
4. **Presentation.** The dimension takes the candidate's state, with `interpretedAt` and `runID` from the candidate. It gets the evidence `{document_id, version_id: candidate, locator: "equivalent_version_interpretation"}`. Limitations, stable machine-readable strings:
   - `carried_forward_from_equivalent_version:<version>:run:<run>:configuration:<configuration>`;
   - `own_interpretation:failed:<code>` or `own_interpretation:not_available`;
   - the candidate's own limitations, appended.

   The document's own `version_id`, `sha256` and `acquired_at` are unchanged. The HTTP and MCP serializers already share `Document`, so both return identical output.
5. **The read cache stores the final, carried-forward assessment per version.** The `queryReads.interpretations` map stores the final dimension, so nested operations in one query see the same value.

## Risks / Trade-offs

- [The earlier interpretation came from a weaker or older configuration] → Its state and limitations are shown as they are, with the originating configuration. The version's own result replaces it as soon as one is recorded.
- [Equivalence too broad and hides a real change] → Only interpreted text, metadata and resource roles count, and every one of them must match. Text-only scenario tests cover markup-only, text-changed, metadata-changed, attachment-state and OCR-output cases.
- [Two extra queries on document-heavy pages] → Batched and indexed on `(document_id, fingerprint)`. The latency checks of the existing public-latency suite are extended to a page with many versions.
- [Backfill runtime on large development history] → Paged and idempotent, reading only manifest rows. No object storage access.

## Migration Plan

1. Additive migration: the table, indexes and immutability trigger.
2. Deploy admin and worker, then the public service. New manifests get fingerprints. Run the backfill command and verify that counts match the reusable manifests.
3. Verify on development: the three Calcinaia notices report carried-forward `supported` or `partial` with the expected limitations, API and MCP return identical responses, and boundaries before the earlier results show no carry-forward.
4. Then adopt extraction v29 and the bounded 31.2 reprocessing.

Rollback: restore the previous public image. The table is additive and can stay.
