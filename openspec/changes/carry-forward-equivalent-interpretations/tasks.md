# Tasks

## 1. Equivalence fingerprint

- [x] 1.1 Add the embedded migration `100_interpretation_input_equivalences` and register it in `internal/backend/classification/migrate.go`. It creates an immutable table keyed by manifest `run_id` with document, version, fingerprint, fingerprint version and creation time, plus an index on `(document_id, fingerprint)`. Verify with classification PostgreSQL integration tests:
  - repeated migration is idempotent;
  - updates and deletes are rejected;
  - rows require an existing manifest.
- [x] 1.2 Implement the `interpretation-equivalence-v1` fingerprint from an `InputManifest` as specified in design decision 1: configuration and HTML/text/JSON raw identity excluded, OCR identity and page input/output included, reusable manifests only. Verify with unit tests:
  - markup-only differences under two configurations match;
  - text, metadata, issuer, attachment required/missing state, eligibility policy and OCR page output each break equivalence;
  - non-reusable or incomplete manifests get no fingerprint.
- [x] 1.3 Persist the fingerprint whenever `PutManifest` stores a reusable manifest, and add the bounded, idempotent CLI command `interpretation-equivalence-backfill` (page size, resume after run ID, JSON report). Verify with integration tests:
  - new classification and extraction runs write one row;
  - the backfill fills existing manifests exactly once;
  - a rerun writes nothing;
  - non-reusable manifests are skipped and counted.

## 2. Carry-forward in public document reads

- [x] 2.1 Extend `loadDocumentInterpretations` with the two batched carry-forward queries of design decision 3: candidates for `failed`/`not_processed` IDs by fingerprint, earlier first acquisition and manifest run within the cutoff; then assessment of the candidates, choosing the most recent `supported`/`partial` one and skipping `unreliable`. Verify with publicquery PostgreSQL integration tests:
  - a failed latest version and a supported earlier equivalent;
  - a classification-rejected version with a manifest but no result;
  - no equivalent;
  - an unreliable earlier version;
  - a different document with the same text;
  - the own later result winning (including not relevant).
- [x] 2.2 Present carried-forward assessments with the evidence locator `equivalent_version_interpretation` and the stable limitation strings `carried_forward_from_equivalent_version:…` and `own_interpretation:failed:<code>` or `own_interpretation:not_available`, keeping the version's identity, hash and acquisition time. Verify with unit tests on the limitation and evidence format, and with an integration test asserting no interpretation, receipt, measure or projection row changes during reads.
- [x] 2.3 Verify the knowledge boundaries with integration tests: a boundary before the earlier result carries nothing; a boundary after the own result shows the own result; a boundary between them shows the carried-forward one; source interpretation cutoffs are respected.
- [x] 2.4 Verify consistency and bounds:
  - HTTP and MCP return identical document detail, versions list, document search and saved-view pages;
  - the read cache returns the same assessment within one query;
  - a document with at least 40 template-only versions resolves carry-forward within a fixed number of round trips (existing query-count harness).

## 3. Adoption and documentation

**3.1 verified, 22 October 2026:**
- `go vet ./...`, with and without the `integration` tag, and the full unit suite pass.
- Race runs pass for classification, publicquery, extraction, transport and cmd.
- `-tags=integration` passes for classification, interpretation, publicquery, transport, cmd, extraction and domain (including retention).
- `openspec validate --strict` passes.

- [x] 3.1 Run the unit, race, vet and `-tags=integration` suites for the classification, interpretation, publicquery, transport and cmd packages, plus `openspec validate --strict` for this change, and record the results.
- [ ] 3.2 Adopt in development with private rollback evidence:
  - archive the database and settings;
  - apply the migration;
  - replace admin, worker and public;
  - run the backfill and verify its counts against the reusable manifests;
  - check the runtime readiness.

  Then verify through the public API and MCP that the three Calcinaia notices of CLN-036 show the carried-forward interpretation with the expected limitations, and that boundaries before the earlier results do not.
- [ ] 3.3 Update the Calcinaia guide (new dated entry), the coverage matrix, the operations documentation on public interpretation states and the changelog with the verified behaviour and residual limits. Then hand back to task 31.2 of `define-toscana-alert-service`, which covers v29 adoption, bounded reprocessing and API/MCP comparison.
