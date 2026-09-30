# Design

## Context

See proposal.md for motivation. The read-only audit at approximately 2026-09-22 09:26 UTC found 3,567,194 input and 596,401 output tokens. OCR accounted for 1,523,416; classification 1,386,552; extraction 1,253,620; embedding 7 input tokens. Of 1,293 OCR pages, 1,073 repeated an exact image/model/configuration combination after its first occurrence, consuming 1,282,103 tokens. These are historical measured calls, not a guaranteed future saving. Seventy-four locally invalid results consumed 382,303 tokens. Counts continued changing during the audit.

Code and runtime observations:

- `internal/ocr/runner.go` keys runs by version, resource URL and configuration, checks pages only within a run, and computes image hashes after the provider call.
- `internal/documents/model.go` and `store.go` preserve raw hashes; CFR PDFs changed byte hashes on every acquisition, although many rendered images were identical. PDF metadata churn is a hypothesis requiring a paired-file comparison, not a proven cause.
- `internal/classification/content.go` includes document-version ID in its content digest. A lack of duplicate digests does not demonstrate unique textual content. Classification and extraction accumulate successful segments before final result persistence, leaving retry work to examine.
- `internal/acquisition/regional.go` accepts images by configured path prefixes, marks discovered graphics required, and `internal/interpretation/scheduler.go` sends all retained PDF/PNG/JPEG resources to OCR. A lightning icon appears in rejected jobs, but its semantic role must be reviewed before exclusion: existing regional tests explicitly include a thunderstorm symbol.
- `internal/inference/openai_chat.go` discards non-success bodies. It creates an HTTP status in CallError, but runners preserve only the generic code; Retry-After is not enough unless scheduling consumes it.
- `internal/jobs/worker.go` and `internal/acquisition/worker.go` stop on loop errors; heartbeat loss does not immediately cancel the handler. `cmd/iwa/main.go` cancels all workers after the first exit and logs no component/cause. Fifty-two restarts occurred, frequently around 80 minutes apart. This correlation does not establish a root cause.
- The main spec inventory is empty. The pending `define-toscana-alert-service` ingestion spec already mandates retained originals, full-content coverage, bounded retries, evaluated processing changes and explicit unknown usage. This change adds efficiency/operations contracts without editing that proposal or the dashboard change.

## Goals / Non-Goals

**Goals:** eliminate proven duplicate inference, keep failures diagnosable and bounded, preserve evidence and quality, and measure results against a frozen baseline.

**Non-goals:** change provider/model as a shortcut, weaken validators, reduce collection cadence, discard raw revisions, impose an arbitrary monetary cap, publish sources, or replay all historical failures. Rejected-call billing and provider-account changes cannot be inferred from local data.

## Decisions

### 1. Establish call-level evidence before correcting unknown operational causes

Add an additive provider-call ledger linked to processing attempts and page/segment input identities. Persist call intent before transmission and a durable receipt before domain validation. Record model/configuration, bounded timings, HTTP status, allowlisted provider error code, request ID, finish reason and available token fields. Parse error bodies in memory with a size limit; store only sanitized approved fields, not raw bodies, credentials or arbitrary text that may echo prompts. Preserve partial usage on parse/validation failure. Existing per-attempt reports become reconciled aggregates; do not sum ledger and legacy attempts together.

Mark abandoned intents as uncertain, not free. Durable receipts can close an attempt after a crash; a crash between provider response and durable receipt remains an acknowledged billing ambiguity. Historical totals stay immutable and visibly less complete. This is preferable to adding generic log lines because it supports error diagnosis and accounting across restarts.

Use a bounded diagnostic sample (one request per relevant failing input class, initially at most three total calls, no automatic repeat) after explicit live-test authorization to identify rejection status/category; fix the demonstrated request, endpoint, model-access or operator credential/quota issue. Account remediation stays an operator dependency if required. A mocked rejection test does not establish that production recovered.

### 2. Contain failures at the correct scope

Introduce durable provider gates keyed by endpoint plus an opaque credential-scope identifier, with model-specific gates where appropriate. Never persist credential material as an identifier. Authentication failures hold immediately; confirmed account/quota failures hold until operator recovery; 429/5xx use bounded backoff, Retry-After and one half-open probe after cooldown. Proposed configurable default: open after three consecutive temporary failures, 60-second initial cooldown, increasing to 15 minutes, never earlier than Retry-After. HTTP 400 alone must not imply a global outage: persist permanent rejection per exact request identity and configuration, with explicit invalidation/relaunch.

Check gates before consuming a job attempt; deferred jobs retain their remaining attempt budget. New versions cannot bypass an open gate or repeat a known permanently invalid equivalent input. Show held work and reason in local administration. Keep collection running. Controlled recovery admits bounded work and never automatically relaunches terminal historical jobs. This avoids both per-job retry amplification and disabling every source for one malformed image.

### 3. Repair worker lifecycle from a reproduced failure

Return tagged component outcomes from the supervisor and preserve safe error categories at failure sites. Correlate acquisition leases, queue attempts, database availability and worker exits; replay a retained source cycle and simulate faults beyond the observed 80-minute cycle using injected clocks where possible. Correct the evidenced cause in addition to logging it.

Pass a cancellable ownership context into handlers and acquisition checks. Cancel promptly on lost heartbeat/lease, fence completion by claim token, and persist/resolve interrupted calls. Recover transient database/claim failures with capped backoff; classify fatal configuration/schema errors separately. Do not broadly swallow all errors. Keep bounded shutdown and durable retry budgets. Recovery tests cover lost ownership, process cancellation, database interruption and duplicate-effect prevention.

### 4. Reuse immutable OCR artifacts across versions

Compute page input identity before calling Regolo. Cache key includes exact request image digest, media type, provider/model/configuration and renderer/preprocessing identity. Use a PostgreSQL artifact row with ownership lease and fencing, avoiding a database transaction held open for the network call. Other workers wait/defer and reuse only a validated completed artifact. Cache successful readable OCR initially; do not silently promote failed or empty output to reusable success.

Each version receives its own page association with resource URL/page number, reuse origin and artifact identifier. Keep original provider usage on the originating call only; a cache hit has no new provider call and an explicit reuse count. Resource-level manifests can bypass rendering unchanged PDF bytes; different raw PDFs still require local rendering before page equivalence is known. Backfill compatible successful historical artifacts with stable renderer/configuration provenance in bounded batches; ambiguous provenance or conflicting historical outputs are not blindly selected. Keep retention dependencies explicit, including cleanup, backups and restoration. No global cross-source domain-result sharing: only compatible OCR artifacts may be shared under existing access constraints.

Alternative rejected: hashing raw PDFs alone, which misses the observed regenerated-PDF duplication. Fuzzy image/text matching is also excluded; exact matching is safer and already captures substantial waste.

### 5. Separate interpretation identity from raw evidence identity

Preserve `retained_versions` and original byte hashes. Add a versioned interpretation-input manifest scoped to source and logical document. It includes complete normalized meaningful HTML/text, ordered required page/image identities, relevant metadata (including issuance/validity), completeness/missing states, resource eligibility policy and processing configuration. It excludes database version IDs, transport timestamps and only explicitly verified presentation noise. It does not remove dates or infer equivalence from issuance time alone.

For CFR, compare paired retained PDFs locally (metadata, text and rendered page identities) to characterize byte churn. For municipal HTML, identify the changing regions with retained diffs and permit only fixture-proven normalization of non-content attributes or boilerplate. Maps and legends remain page/image inputs even if OCR text is identical; OCR is not a complete representation of graphical meaning. If safe equivalence cannot be established, process normally or expose a local preparation error.

Keep raw evidence digests for provenance and add separate inference-content keys rather than repurposing every existing hash. Only a successfully validated prior result under an identical manifest is reusable. Materialize version-specific evidence mappings and revalidate all literal references and offsets before linking downstream results; otherwise reprocess. Preserve publication time and deduplicate domain effects by logical interpretation identity. A genuinely changed or incomplete revision still raises existing supersession warnings. Explicit selected reprocessing can use compatible stage caches but must honor configuration changes.

Alternative rejected: suppressing raw retained versions, which would erase acquisition evidence; generic text-only comparison could miss changed graphics or attachments.

### 6. Make image eligibility explicit and consistent

Add versioned source rules that classify resources as interpretive evidence or confirmed decoration, recording the reason. Review actual CFR icon, map and legend contexts before changing configuration. Preserve raw acquired evidence; skip OCR only for positively identified decoration. Unknown images default to eligible. Use one eligibility decision in scheduler, AfterOCR completion checks, content gathering and equivalence manifests so exclusions cannot strand a document or make incomplete evidence appear complete. Required-resource policy changes must follow existing registry preview/application rules.

### 7. Reduce invalid paid outputs without weakening validation

Build a fixed reviewed corpus from the 65 classification-output and nine extraction-evidence failures, plus successful and adversarial controls. Inspect whether raw failed responses are retained; unavailable historical responses require labeled reproduction, not guessed diagnosis. Break failures down by schema, truncation, literal quotation, Unicode normalization, unsupported references and semantic mismatch. Tune request/schema instructions only where evidence supports a fix, retaining non-thinking mode and current models initially. Accept provider stop/length semantics explicitly and keep valid usage even on rejection.

Persist per-segment receipts and validated checkpoints before proceeding to the next segment. Retry only unfinished compatible work; permanently invalid or ambiguous outputs remain terminal unless explicitly reprocessed after an evaluated configuration change. No automatic LLM repair loop. Preserve complete-document classification and operational-window coverage. Evaluation must show fewer invalid outputs on reproduced fixable cases and no lost relevant notice, unsupported assertion or evidence regression on reviewed controls.

### 8. Measure outcomes, not just lower totals

Extend existing local diagnostics/trial-cost reports with tokens and requests by date/source/model/stage, actual calls versus cache hits, semantic skips, rejected/invalid/unknown calls and worker exits. Do not label local cache hits as provider cache tokens. Show estimated avoided tokens separately, with the original-call basis and limitations. Add a reproducible read-only audit/replay command; use retained fixtures and fake adapters by default so evaluation itself cannot accidentally consume tokens.

If a provider export is available, reconcile UTC window, key/model scope and out-of-band probes; otherwise record that billing remains unverified. Replaying the frozen 1,293-page audit set should identify 1,073 exact repeat opportunities, subject to compatible provenance; the measured historical associated tokens are a benchmark, not a promised production reduction.

## Risks / Trade-offs

- [False equivalence hides a new warning] → exact, versioned manifests; graphical inputs and completeness included; changed-risk/date/map/ordinance regression cases; uncertain inputs are not skipped.
- [A lightning symbol is meaningful] → review its source context and test maps/legends; no filename-only exclusion.
- [Reuse breaks evidence or retention] → version-specific associations, literal revalidation, dependency-aware cleanup and restore tests.
- [Restart between provider call and receipt duplicates billing] → durable intent, fenced ownership, bounded recovery and explicit uncertainty; no exactly-once billing claim.
- [Circuit gates delay interpretation] → scoped gates, visible held state, bounded probes and preserved collection/freshness warnings.
- [Historical tests overstate savings] → fixed UTC baseline, cold/warm replay counts and separate live observation; do not add overlapping savings categories.
- [Changes overlap unfinished ingestion/dashboard work] → additive migrations and reports, no edits to other planning changes, check their current interfaces before each implementation task.

## Migration Plan

1. Freeze a read-only baseline and retained-case inventory; record running image/revision and database schema. Prepare additive migrations and independently configurable reuse/equivalence switches, default off until evaluated. Keep raw histories intact.
2. Validate call diagnostics, provider containment and worker fixes in isolated tests. With deployment authorization, roll out observability/containment first and diagnose actual provider recovery before any inference backfill.
3. Backfill compatible OCR artifacts offline in bounded resumable batches. Compare cache/equivalence decisions in shadow mode on retained data; no paid calls and no publication effects.
4. Enable OCR reuse first, then reviewed resource rules, then interpretation equivalence and evaluated output/configuration fixes on one source at a time. Apply registry changes through existing preview controls. No model or processing version change automatically replays history.
5. Validate a bounded operator-selected current/failure scope after provider recovery. Document selection size, expected maximum requests and measured usage; keep historical failed work terminal unless selected.
6. Observe at least 24 hours after the final canary change and at least three formerly problematic worker cycles, extending observation if no bulletin revision occurs. Require no unexplained worker exits, no duplicate calls for warm compatible page keys, no repeated permanently rejected identical inputs, successful required processing, and passing changed-evidence/quality controls. This observation does not replace the existing seven-day source-acceptance trial.
7. Roll back behavior through switches and the previous compatible application image, retaining additive tables and evidence. Do not drop data or reopen all held jobs. Restore ordinary inference only within the provider gate and existing retry policy. Mark deployment/observation tasks incomplete until performed and evidenced.

## Open Questions

- What exact HTTP/error category caused the production rejections? Resolve using the bounded diagnostic sample; credentials or quota may need operator action.
- Which worker component and safe error cause match the 80-minute pattern? Resolve through diagnostics and targeted reproduction before claiming the restart finding fixed.
- Which PDF metadata fields and municipal HTML regions actually change? Paired retained-file comparison determines the permitted normalization fixtures.
- Is a Regolo account export available for this key/window? Its absence limits reconciliation, not the local efficiency fixes.

## September 23: local preflight and equivalent-copy dependency

Raw revisions accumulate even while provider calls are held. Add a local preflight
manifest before inference, scoped to logical document, processing configuration,
renderer identity and active resource policy. Preserve every original. Normalize
only the already reviewed municipal generated comment; compare every PDF page at
the production 144 DPI, and retain all structured metadata and other raw resource
identities. Missing evidence is never equivalent. Use the immediately preceding
raw version, not a global historical hash lookup: A→B→A is a real transition.
Archived pre-reset versions are not representatives for fresh work.

Persist representative dependencies separately from classification/extraction
results. Equivalent copies retain their jobs but wait without spending retry
budget until their representative has validated classification and (if relevant)
extraction. Then materialize current-version provenance using existing stage reuse;
for these copies a context guard forbids any provider fallback. A reuse miss is
visible, never a paid retry or a fabricated success. Explicit evaluated reprocessing
is exempt and remains operator selected. Admin pending summaries group only copies
waiting for their representative; raw-version details and materialization failures
remain visible. Public evidence is not redirected or claimed interpreted until
normal validated result materialization finishes. Source polling and raw retention
remain unchanged. Keep pre-claim provider holds even with OCR reuse enabled.

Live shadow review also verifies the exact opening div whose sole class is
`js-view-dom-id-` plus 64 lowercase hexadecimal digits as a generated identifier.
Municipal HTML policy v2 normalizes this opening tag and the known comment using
one shared function; arbitrary classes, text, URLs and card ordering are retained.
The reviewed pages also contain actual listing changes, so they are not declared
equivalent simply because generated identifiers were removed.

The pending-work projection must distinguish discovery evidence from inference
inputs. For the reviewed municipal source, an exact retained-configuration section
URL (optionally its declared numeric `page` query) with neither a discovered target
nor an interpretation trigger is a discovery listing, not pending LLM work. Such
pages remain visible through explicit version detail. Unknown URLs/queries and any
actual target/trigger remain conservatively visible. This corrects the dashboard;
it does not remove documents or pretend a listing was classified.
