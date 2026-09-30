# Public progress snapshot

These checkboxes reflect the private project task list on 30 September 2026. A checked item records project work; it does not by itself establish public source acceptance, public deployment, or that unpublished evidence is available in this repository. Operational notes and private evidence references are maintained separately.

## 1. Freeze evidence and regression inputs

- [x] 1.1 Add a reproducible read-only audit command/report with a fixed UTC cutoff, running image/revision, schema, per-stage/source/day usage, duplicate image/configuration calls, invalid outputs, unknown attempts and restart timeline; verify the report against independent SQL aggregates and clearly distinguish the earlier September 22 snapshot from a newly collected baseline.
- [x] 1.2 Assemble retained paired CFR PDFs, municipal HTML changes, CFR graphics contexts and a manifest of the 74 observed validation failures plus successful/negative controls; verify raw hashes, classify missing historical response evidence, and document which PDF/HTML differences are proven presentation-only and which graphics are meaningful.

## 2. Provider diagnostics and truthful consumption

- [x] 2.1 Add additive provider-call intent/receipt storage and safe diagnostic fields in processing/inference, with legacy accounting compatibility; verify migration on a populated database copy, redaction tests, HTTP status/error parsing, malformed responses with available usage and unknown interrupted calls.
- [x] 2.2 Wire OCR, classification, extraction, linking and embeddings to durable call receipts and reconciled attempt aggregates; verify failed local validation retains consumption, mixed known/unknown calls stay explicit, and crash recovery neither loses durable receipts nor double-counts legacy/current totals.
- [x] 2.3 Add a bounded diagnostic mode and production-rejection investigation runbook (one call per input class, initially at most three calls, no retries); verify mock authentication/quota/request errors identify corrective actions without exposing credentials, and leave actual production recovery for task 9.2.

## 3. Contain rejected requests across jobs

- [x] 3.1 Implement durable scoped provider gates, configurable temporary-failure thresholds/cooldowns, Retry-After and single-probe recovery; verify authentication/account holds, 429/5xx backoff, restart persistence and healthy-model isolation using deterministic tests.
- [x] 3.2 Integrate pre-call/pre-attempt deferral and permanent request-identity rejection reuse with queue scheduling and operator recovery; verify new document versions cannot bypass a hold or repeat the same permanently invalid input, waiting does not consume retry budget, and explicit recovery does not relaunch historical terminal jobs.

## 4. Diagnose and repair worker restarts

- [x] 4.1 Add component-tagged supervisor outcomes and safe categorized errors at queue/acquisition failure sites; reproduce the observed exit pattern using retained source cycles, lease timing and injected database failures, and record the confirmed cause or precise remaining production evidence needed in task 9.2.
- [x] 4.2 Implement ownership-context cancellation, bounded transient recovery and the reproduced lifecycle correction in jobs/acquisition/supervision; verify immediate handler cancellation on lease loss, fatal-error visibility, bounded shutdown, preserved budgets and no stale or duplicate effects in integration tests. If the production cause remains unknown, do not claim the restart finding resolved.

## 5. Reuse OCR safely

- [x] 5.1 Add OCR artifact keys, durable fenced ownership and version-specific evidence associations with additive migrations; verify identical-input concurrency, expired ownership, model/prompt/renderer invalidation, and current-version page provenance in PostgreSQL integration tests.
- [x] 5.2 Move image hashing before provider calls and integrate artifact reuse plus unchanged-resource render manifests into the OCR runner; verify warm identical pages make zero calls, changed pages call once, partial resource retries reuse completed pages, and local reuse contributes no new provider tokens.
- [x] 5.3 Add bounded resumable historical artifact seeding and retention/backup compatibility; verify ambiguous/conflicting historical provenance is skipped, reruns are idempotent, dependent evidence survives cleanup/restore, and no network inference or public effects occur during seeding.

## 6. Meaningful changes and resource selection

- [x] 6.1 Add versioned inference-eligibility rules and reviewed source fixtures distinguishing decoration from maps, legends and image-only notices; verify unknown/meaningful graphics remain eligible and the actual lightning symbol is excluded only if task 1.2 establishes it is decorative.
- [x] 6.2 Apply eligibility consistently to scheduling, OCR completion, content gathering and source configuration previews; verify exclusions do not strand classification or conceal missing required resources, and changed collection/required-resource policies still follow registry lifecycle controls.
- [x] 6.3 Add versioned interpretation manifests separate from raw evidence/version hashes, covering full meaningful text, graphical/page identities, issuance/validity, completeness and processing configuration; verify metadata-only PDF and proven HTML-noise variants match while changed dates, risk maps, ordinances and missing resources invalidate equivalence.
- [x] 6.4 Reuse validated equivalent interpretation through current-version evidence mappings and idempotent domain effects; verify literal evidence/offset revalidation, preserved publication time, no duplicate public effects, correct genuine-change supersession warnings and fallback processing when mapping is uncertain.

## 7. Reduce invalid outputs and repeated segments

- [x] 7.1 Classify the retained/reproduced invalid-output corpus by schema, finish reason, truncation, normalization, quotation and evidence-reference cause, and add granular stable diagnostics; verify every available failing case has a reason and unavailable historical responses remain explicitly unknown.
- [x] 7.2 Implement evidence-supported request/schema/normalization corrections with new processing versions; verify fewer invalid outputs on the reproduced fixable cases, complete-content coverage and no lost relevant cases or unsupported accepted assertions on the reviewed controls. Keep unfixable ambiguity explicit and prohibit unbounded repair calls.
- [x] 7.3 Persist validated classification/extraction segment checkpoints linked to call receipts; verify a later-segment failure does not repeat completed compatible segments, changed inputs/configuration invalidate checkpoints, and aggregation retains all actual paid attempts exactly once.

## 8. Reporting and release validation

- [x] 8.1 Extend existing local usage/diagnostic reports with actual calls/tokens, reuse/skip counts, estimated avoided usage, provider holds, invalid/unknown calls and worker failures; verify source/model/stage/time filters against seeded SQL and preserve local-only access and existing dashboard compatibility.
- [x] 8.2 Add a provider-export reconciliation procedure with UTC/key/model scope and out-of-band-call handling; verify it using a fixture export, label missing real account evidence explicitly and never present estimated savings as verified billing.
- [x] 8.3 Run offline cold/warm retained-corpus replay and cross-module integration validation, including new issuance/map/attachment changes, full evidence references, cleanup/restore, crashes and circuit recovery; verify zero new calls for warm compatible OCR, quantify observed duplicate opportunities without double-counting savings, run relevant Go tests and repository-required checks, and save results.
- [x] 8.4 Prepare deployment switches, bounded backfill/recovery commands, migration/rollback instructions and a canary checklist; verify the migration and previous-image compatibility on an isolated database copy and demonstrate rollback without deleting evidence or replaying the backlog.

## 9. Authorized rollout and operational acceptance

- [x] 9.1 Deploy validated observability/provider containment and worker changes within the authorized deployment scope; verify health, continued acquisition, migration state, safe diagnostics and rollback availability, recording image/revision and rollout time.
- [ ] 9.2 Use the bounded live diagnostic sample to identify and correct the actual Regolo rejection cause, and confirm the production worker-exit cause against new diagnostics; verify successful representative required requests and cause-specific regression evidence. If provider account action or an unimplemented runtime fix is needed, record the blocker and keep this task incomplete until corrected and validated.
- [ ] 9.3 Seed compatible OCR artifacts, run shadow equivalence checks and enable reuse, reviewed resource policy, equivalence and evaluated output fixes in order, one source at a time; verify meaningful new evidence still processes, original retention remains intact and the canary makes no duplicate warm-cache calls. Recover only an explicitly selected bounded current/failure scope with recorded request limits.
- [ ] 9.4 Observe at least 24 hours after the final canary change and three formerly failing worker cycles, extending for insufficient publication changes; verify no unexplained worker exits, no repeated permanent rejection of unchanged input, successful required processing, preserved quality/provenance and measured tokens per unique meaningful update. Record billing reconciliation or its evidence limitation and leave this task incomplete until the observation actually finishes.

## 10. Coalesce technical revisions before inference (September 23 extension)

- [x] 10.1 Add versioned pre-inference fingerprints over complete retained metadata/resources, exact PDF page renders and fixture-proven municipal HTML normalization; persist contiguous equivalence groups without conflating A→B→A or incomplete evidence. Validate meaningful-change controls and database concurrency/retention behavior.
- [x] 10.2 Integrate fingerprints with automatic scheduling and pre-claim deferral of equivalent copies until the representative has valid results; enforce provider-free reuse for equivalent copies, expose grouped pending counts and raw-version provenance, and add bounded offline preparation. Verify zero-call fallback, queue budgets, archival boundaries and meaningful-change scheduling.
- [x] 10.3 Deploy with a verified database backup and unchanged provider hold, prepare only unarchived post-reset versions, enable source-scoped reuse, and record measured groups/copies, unchanged provider calls and health. Document consumption monitoring and rollback; do not claim the unfinished live quality canary or 24-hour observation complete.
- [x] 10.4 Extend the reviewed municipal HTML normalization to the exact generated div class observed during live shadow preparation, sharing the rule between preflight and interpretation manifests; preserve visible content, links and real listing changes. Validate and rebuild only derived municipal decisions before finishing deployment.
- [x] 10.5 Correct the pending-work projection to exclude verified municipal discovery listing pages that are neither interpretation triggers nor discovered document targets; preserve explicit raw-version detail and uncertain/actual documents. Validate configured-section/pagination recognition, target/trigger exceptions and the live dashboard before closing task 10.3.
