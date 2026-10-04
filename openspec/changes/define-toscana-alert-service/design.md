## Context

### Cascina corrections — 2 October 2026

Use the source-revision layout `2 Jan 2006` for the observed unpadded Italian
listing days; test one/two-digit days and date-based tracker selection. Keep
unknown dates eligible and explicit-reference/ongoing-measure exceptions intact.
Discover PDF attachments inside the observed `page-content` container, preserving
explicit dependencies. Review the official municipal decree/notice referral and
reuse conditions for the exact Cascina API origin and `/s3/1520/allegati/`
attachment directory; grant linked municipal PDFs only, with bounded redirects
and real PDF validation. Generic S3 footer resources and other external
origins remain excluded; acquisition of non-PDF dependencies is separate work.

Paired retained Cascina HTML differs only in generated 40-character CSRF values
in an exact meta tag and hidden input. Normalize only those reviewed tag forms
for `cascina-municipal` with one versioned policy shared by preflight and manifests.
Preserve original bytes, operative text, dates, links and required attachments.
Incomplete inputs remain distinct; source opt-in and compatible validated results
are still required for reuse. Test warm reuse without a provider call.

Preserve private database/image/configuration rollback evidence before development
activation. Apply the source revision after preview, verify fresh complete checks
and corrected tracker dates, and add Cascina to the existing preflight/OCR/result
reuse lists. Retire only the explicitly reviewed stale unattempted COC jobs and
versions through a guarded audited operation, keeping originals and histories.
Provider holds, local fallback scope, acceptance and production remain separate.

See [proposal.md](proposal.md) for scope and motivation. The workspace contains a substantial implementation and tests alongside the research and planning artifacts. This revision captures the agreed target behavior and deployment design; completed code tasks do not by themselves certify live source acceptance or operational readiness.

Research context (the detailed investigation records are retained separately):

- National and Toscana legal responsibilities, effective dates, and later amendments require continuing verification.
- The operational CFR routes selected for the MVP are the vigilance page at `https://cfr.toscana.it/index.php?IDS=2&IDSS=71`, the criticality/alert page at `https://cfr.toscana.it/index.php?IDS=2&IDSS=76`, and monitoring at `https://www.cfr.toscana.it/mobile3/avviso-criticita/`. The CFR homepage is a secondary emission/adoption signal. Graphical resources, observed HTTP 429 behavior, and product-specific extraction remain part of acceptance.
- Cittadino Informato and the Calcinaia municipal site have non-equivalent dates and coverage. The 3 October 2026 decision formally adds the platform as a scoped institutional acquisition/discovery channel with primary verification; recognition does not establish completeness, permissions or activation.
- Livorno, Pisa, Pontedera, and Cascina are subsequent municipal rollout targets. Calcinaia is the initial local channel; each municipality requires separate source acceptance.

The latest agreed requirements are in this change's three specs. Research notes remain dated evidence; not all later conversational decisions have been copied into docs. Firenze remains research material, not an initial local collection target.

The authenticated task-9.1 runs after the first bounded-segmentation implementation exposed two remaining constraints. Prompt-only and greedy/non-thinking candidates can improve individual cases but do not make the full regression stable, and a verbose per-field evidence envelope can exceed the 90-second worker window on an ordinance page containing many measures. The current extraction result also cannot retain two supported temporal candidates for one field before domain projection, so rejecting the entire merge or selecting one candidate both lose required information.

## Goals / Non-Goals

The development recovery procedure is documented in
[acquisition recovery](../../../docs/operations/acquisition-recovery.md). Reviewed
stale-target decisions preserve history and reactivate on rediscovery. Temporary
incident restrictions are explicitly released after a bounded canary; embedding
and linking inherit their extraction run's document scope. Rejected extraction
responses use the existing private append-only diagnostic store, and oversized
classification identities use a deterministic digest while preserving existing
bounded keys. These operations do not establish public source acceptance.

**Goals:** evidence-supported output, independent quality dimensions, revision-aware local measures, reproducible internal diagnosis, equivalent public interfaces and sustainable anonymous access.

**Non-Goals:** autonomous discovery of new source channels, manual editing or approval of individual public notices, mandatory local model inference, complete historical reconstruction or public administration. Initial public limits are an operational guardrail rather than a capacity guarantee; concrete route/tool names are recorded in the completed technical contract (task 1.7), with live verification still required.

## Decisions

### 1. Separate authority, channel, document and interpretation

Use a logical source registry to distinguish responsible authorities, publishing channels, platform operators and supported territories/products. Municipal referrals include their actual scope and evidence; platform documentation is shared while municipal recognition is individual.

Start attribution from the concrete publication: identify its issuer, territory and document kind, distinguishing regional republication, local information and municipal provisions. A clearly identified municipal ordinance on a recognized external channel retains municipal authority. Reconstructing every associated-service convention is not a prerequisite for acquiring that act; the other acceptance conditions still apply. When attributing a decision to an associated entity exercising transferred functions, verify the relevant institutional acts, function, municipality and applicable period. Keep unresolved competence scoped to the affected attribution rather than blocking unrelated verified municipal publications. Channel recognition and evidence of transferred functions remain separate registry relationships.

A document version retains acquired evidence independently of any interpretation. Extracted warning facts and individual measures reference supporting passages and the interpretation version. This separation permits reprocessing after a code correction without rewriting the original or confusing a republication with a new issuing authority. A single merged alert record would lose partial updates and provenance.

### 2. Separate acquisition success from interpretation success

Collection records reachability, recognizable content and completion of the declared checking scope. Interpretation records supported fields and unresolved ones. For a publicly enabled source, a newly acquired document remains discoverable through metadata and its official link even when interpretation fails; retained-copy access follows the separate default-off publication setting; the previously interpreted state is accompanied by a newer-document warning. This follows the agreed visibility requirement instead of silently quarantining every ambiguous publication.

Checks default to 10 minutes, with a separate 30-minute complete-check delay threshold, both configurable per source. Immediate errors, backoff and documented publication schedules have separate states. No fixed municipal publication schedule is inferred. Collection frequency does not guarantee discovery latency or complete real-world awareness.

For CFR, configure expected publication windows only for documented daily vigilance and criticality/alert products. Monitoring remains event-driven, so a complete unchanged check outside an expected event is successful and does not create a missing-publication alarm. All collectors honor upstream retry instructions, including HTTP 429 backoff, even when this exceeds the nominal ten-minute cycle.

### 3. Model updates at measure scope and preserve temporal meanings

Measures identify the supported action and affected place/territory. Updates can change one measure while retaining others, including exceptions within a page. Unsupported links remain unresolved. Operational phase, local measure and regional color are separate facts.

Keep original temporal text and distinguish source publication/modification, event time, operational validity, page expiry, platform dates and service acquisition/interpretation. Do not turn a date-only value into an exact event time. Store determinable instants in UTC and preserve precision and the source expression. Apply Europe/Rome only when justified by the source, recording that assumption. Date-only values remain dates; timezone ambiguity remains unresolved. Concrete field schemas and interval boundary examples are verified in the technical contract.

Extraction retains temporal candidates before producing the resolved measure view. Each candidate carries its temporal meaning, original expression and evidence. Incompatible candidates for the same measure/meaning receive one conflict identity and project to an undetermined resolved field while remaining available for diagnosis and later supported clarification. This uses the existing domain distinction between temporal meanings and conflict groups rather than treating model order, page order or a publication date as a tie-breaker.

The Calcinaia sequence is the first evaluation case: the underpass remains closed despite a general roads-cleared statement; a later reopening does not lift other restrictions. A 09:36 page update does not establish the event's exact time. The ordinance linked from the closure notice must be examined for the duration it supplies.

### 4. One shared public interpretation, two access interfaces

API and MCP read the same published document/measure state and quality metadata. Query-time presentation separates current and future already-issued regional information, local provisions, unresolved states and history. Public queries do not trigger collection. Internal DPC comparisons and diagnostic tools remain outside public warning results.

Expose versioned query views so pagination and equivalent cross-interface requests do not silently mix updates. Use five shared operation groups: municipality/zone discovery; municipality situation; filtered warning/measure search; document/evidence/version details; source coverage/updating status. Applicable operations support retained history. Cursors identify a fixed data view and evaluation time, with a configurable lifetime initially 30 minutes; expiry requires restarting. A shared per-IP API/MCP budget starts at 120 requests per 60-second window with at most 100 items per response/page. The public service sits behind the existing HTTPS reverse proxy, and only that proxy is trusted to supply the client address through `X-Forwarded-For`; document shared-IP effects. Transport routes, MCP tool names and field schemas are documented in the completed task 1.7 technical contract; live deployment verification remains required.

### 5. Retain evidence with the conclusions it supports

Retention is configurable, initially three calendar months from the first acquisition of each version, for structured records and acquired originals/versions. Unchanged checks do not reset that clock. Bootstrap collection covers the previous 30 days and explicitly referenced older documents without claiming complete pre-startup history. Measures still valid or without established cessation protect the documents and versions needed to reconstruct them beyond the horizon. An unresolved measure does not become confirmed active through this exception.

A changed duration applies to existing data at the next cleanup: increases protect remaining data longer, reductions make older unprotected data eligible, and deleted data is not restored. When protection ends, already-expired versions become eligible. Define calendar arithmetic and dependency-safe cleanup in the technical contract, including month-end cases. Retention cannot silently discard evidence required by an exception. Source-specific reuse constraints remain an acceptance condition and any resulting limits must be disclosed. For the scoped Cittadino Informato workflow, task 26.2 uses the multi-source technical criterion below rather than a license or legal-basis documentation prerequisite.

### 6. Diagnose quality rather than manually curate notices

Internal monitoring assesses source availability, update timeliness and interpretation correctness. Keep reproducible cases and expected outcomes for code/process evaluation, including false positives and missed publications. DPC supplies internal comparisons only where risk, territory and time are actually comparable. Successful HTTP responses and LLM self-assessment alone do not establish correctness.

Build attributable reviewed expected outputs for real Calcinaia notices, irrelevant content, partial reopenings, conflicting dates, attachments and scanned PDFs; label simulations separately. A missed relevant notice or unsupported measure in evaluation blocks public activation of the affected source until correction and successful rerun. Also measure unresolved fields so an extractor that returns no useful facts cannot pass on precision alone. Run at least seven complete 24-hour observational intervals across vigilance, criticality/alert, monitoring and Calcinaia, persist daily checks, and compare originals and attachments by human review or an explicitly operator-delegated assistant review. Retain cases for significant events absent during the trial. Extend the trial for omissions, delays, missing evidence, unresolved defects or insufficient observations. Source acceptance is distinct from any per-notice editorial gate.

### 7. Modular Go service and durable storage

For task 11.5, campaign scope is persisted separately: omitted scope and legacy
campaigns remain `mvp`, requiring the four pilot products. Explicit `municipality`
requires one municipal source from the five planned Toscana municipalities and
removes only the regional-product requirement for that individual report. All
active collection prerequisites, observation duration, daily checks, gaps,
original/attachment comparisons and failure/event cases remain unchanged.
An additive migration preserves the original schema checksum and append-only
campaigns/assessments. Individual pending reports preserve unresolved scope
prerequisites; starting a campaign never grants acceptance/public enablement.

Use one modular Go application with background worker processes for collection and interpretation; API/MCP requests only read prepared data. PostgreSQL stores the domain model, configuration history, interpretation runs and a durable job queue with status, attempts and errors. Workers recover incomplete jobs after restart and use idempotent processing to avoid duplicate versions or public effects. This keeps scheduling and records in one persistence system; SQLite and an additional queue service are not the selected MVP design.

Store original HTML, necessary page resources, PDFs and scan images in RustFS through S3, with hashes and object references in PostgreSQL. RustFS remains internal to the service VM. A service setting, disabled by software default, controls public copy access through the application; the initial deployment enables eligible CFR/Regione Toscana copies and direct Calcinaia copies only where the recorded source policy permits them. Cittadino Informato and albo pretorio remain link-only. Official links remain present. Database/object writes, cleanup and backup must preserve referential consistency.

Docker Compose runs Go processes, PostgreSQL, RustFS and local Crawl4AI in separate development, staging and production projects on the current VM. Production container CPU, RAM and process ceilings bound individual services but do not limit named-volume growth or prevent all host contention. OCR/inference use external APIs initially; local replacements remain possible. Measure concurrent load, storage growth and restore time before production activation, expand the VM from evidence, and review capacity when sustained disk or memory consumption exceeds 70%.

### 8. Configured discovery, OCR and separate LLM stages

Operators configure sources and sections through administration; the service discovers documents within those boundaries rather than autonomously discovering new channels. Crawl4AI is an internal acquisition component called by Go workers. Preserve original bytes and required resources alongside extracted text where permitted. For selected municipalities, register Cittadino Informato as an additional institutional acquisition/discovery channel with its own municipality referral, sections and policy. Platform notices generate candidates checked against municipal publications/referenced acts and, for quoted regional warnings, originating CFR products. Preserve independent primary collection; a platform omission must not hide a municipal notice. Secondary-only candidates do not establish confirmed measures or completeness. The albo pretorio remains limited to acts referenced by primary notices. The 3 October decision below is a planned extension, not a live policy activation.

Classify full content of new/changed documents using Qwen3.8-27B. For pertinent notices, fetch necessary attachments and interpret them together, retaining evidence per document and page. Include scanned PDFs using DeepSeek-OCR-2. Use separate Qwen stages/prompts for relevance, structured measure extraction and links to previous measures. OCR remains retained verbatim; before classification or extraction, presentation-only markers introduced by OCR are normalized through a deterministic view that keeps offsets resolvable to the retained OCR page and original document.

Classification continues to cover every deterministic full-content segment. Extraction uses a distinct operational-window builder instead of reusing classification chunks blindly. Each window has a non-overlapping core plus bounded surrounding context selected on stable text boundaries. A fact may be emitted only when its operative predicate begins in the core; subject, place and temporal evidence may come from that window's context. This preserves local qualification across boundaries while preventing overlapping windows from owning the same fact. Window size, context allowance and completion budget are versioned configuration selected through retained-case evaluation; increasing the 90-second worker timeout is not a substitute for bounding input and output.

The extraction response contains a compact evidence table once per window. Candidate measures and their asserted fields reference evidence-table indices instead of repeating URL, page and quotation for every field. The local parser expands those references to the existing persisted field evidence, verifies literal normalized spans against the retained resource/page and rejects unknown, duplicated or cross-window indices. Provider `null` is the explicit unknown value; the persisted `indeterminate_fields` list is derived locally from null optional fields instead of asking the model to repeat the same state consistently.

An ordinance subject heading alone cannot establish an operative measure. It may still support a temporal conflict for a measure independently established by operative evidence. When a partial reopening explicitly repeats coordinated closures that remain in force, bounded local completion preserves each stated subject with literal, core-owned evidence; it does not infer validity times, extend the reopening or supplement from negated statements or context-owned predicates. The correction has a new configuration identity while previous processing semantics remain available behind the existing source gate.

Segment merge collapses only exact duplicate supported candidates. Complementary candidates combine when their identities and asserted values are compatible. Incompatible supported values are retained with evidence under a conflict identity, and only the affected resolved field becomes undetermined; unrelated fields and measures remain available. A failed required window still leaves the document uninterpreted rather than accepting a partial document silently. Code validates schemas, ownership and evidence references; these checks do not alone prove semantic correctness. Unreadable attachments and unsupported fields remain explicit; external document text is untrusted data, never operational instructions.

Prompt-only correction, a longer timeout and automatic selection of one conflicting value were considered and rejected. Four authenticated candidates improved latency or isolated cases but failed the complete reviewed regression. The compact conflict-aware envelope addresses response size and information loss directly, while retained-case evaluation remains the adoption gate for model-generation controls such as non-thinking or temperature.

Go selects same-municipality candidates by structured fields and PostgreSQL text search, including older ongoing/unresolved measures. Semantic retrieval is enabled during the trial and adds candidates without replacing the always-available structured/textual baseline. Compare quality and cost with and without semantic retrieval on the same retained cases. Qwen may link updates without explicit source hyperlinks only when the documents support the relation; ambiguity does not revoke or replace a prior measure.

Initially call both Qwen3.8-27B and DeepSeek-OCR-2 using Regolo.ai API keys, with separate configurable model endpoints and adapters. Credentials remain in secret configuration, excluded from the repository, logs, public responses and browser-visible configuration. Do not assume a shared API format or confirmed model availability until the provider contract check. This interface also permits later local inference without changing the domain behavior.

### 9. Rechecks, bootstrap and controlled reprocessing

At each scheduled cycle check configured listings, notices from the previous 30 days (including those previously classified as irrelevant), older notices supporting ongoing/unresolved measures, and their linked attachments. Detect revisions at stable URLs, including changed attachments on an unchanged page. The initial bootstrap uses the same 30-day lookback and follows explicit references to older documents; dates describe source chronology separately from acquisition. Source backoff may delay a cycle and must be disclosed.

An operator may record an evidence-backed disposition for a specific URL that
returns 404/410 and has disappeared from later fully traversed configured
listings. The plan excludes only that source/configuration/URL while its last
discovery remains unchanged; rediscovery reactivates it automatically. Keep the
original failure, historical incomplete checks and review evidence. This
defines the currently observable collection boundary and does not establish
the unavailable page's content, irrelevance or equivalence to a reachable URL.

Unchanged content does not trigger a new LLM run. Changed content or evidence dependencies can trigger affected stages. Temporary OCR/model errors get a configurable maximum of three total attempts with increasing delays. Ambiguous readable content is not retried merely to seek a different answer. Exhausted jobs keep documents and explicit failure states visible under source publication policy.

Record model, prompt/configuration versions and outputs for every run. Model/prompt/logic changes first pass retained-case evaluation; operators explicitly select source, period or failure status for reprocessing. Do not automatically reprocess all history. Preserve prior runs and cost accounting, with auditable selection of the published interpretation.

### 10. Internal administration and source lifecycle

Bind administration to a local-only interface separate from public API/MCP; access a remote VM through an SSH tunnel or explicitly configured tailnet-only Tailscale Serve. Serve terminates HTTPS and preserves the exact configured Host/Origin; the backend remains loopback-published, forwarded headers do not grant access, and tailnet ACLs control which operators can reach the administrative service. Administrators manage source URLs, sections, intervals, copy restrictions and configuration versions in PostgreSQL. Provide status, failed jobs, uninterpreted documents, quality findings, usage and timing views, and explicit reprocessing controls. This is operational control, not manual editing of extracted measures.

A new source starts as a draft. Run an acquisition preview showing discovered documents/errors, then explicitly enable periodic collection. Collection initially serves internal evaluation; public enablement is a separate action after source acceptance. Once enabled, publish subsequent notices automatically. Changes to URLs, sections or acquisition methods require a new successful preview before applying; keep the previous active configuration in the meantime. Interval changes and suspension apply immediately.

Ordinary source outages preserve the last acquired information with freshness limitations. Confirmed interpretation defects allow suspension of new interpretations for that source while collection continues; mark affected published results as unreliable. After correction, validation and explicit reprocessing, publication can resume without manually rewriting measures. Recognized provenance, collection enabled, technically accepted and publicly enabled are separate states.

### 11. Usage, notification and recovery operations

Measure before setting an operating budget. Track each attempt by document/version, stage, model, input/output/cache tokens where reported, elapsed time, retries and estimated cost with pricing provenance. Missing provider usage is unknown, not zero. Separate bootstrap from steady-state collection, OCR charges, optional embeddings, hosting, storage and future local compute. Present totals and per-stage/provider breakdowns in administration.

Email is the initial notification channel, using an existing SMTP relay and operational mailbox whose details remain private deployment configuration. Notify on source delay thresholds, exhausted processing attempts and backup failures exposed by the selected backup system or notification integration, group repeated errors into an incident, send reminders at a configurable interval and a recovery message after restoration. Transient errors remain visible in the dashboard. Telegram and other channels are future extensions.

Task 9.4 reports each campaign separately at a fixed knowledge boundary; overlapping
campaigns are not summed. Preserve unknown attempts and unavailable hosting/storage
inputs as audited revisions. Current pricing checks do not supply historical
effective dates or account entitlement, and local inference without a provider
charge still needs attributed hosting costs. Completed accounting with explicit
unknowns does not complete observation or satisfy the budget completeness gate.
The operational procedure is [trial costs](../../../docs/operations/trial-costs.md).

The [PBS runbook](../../../docs/operations/pbs-recovery.md) records the six-hour
data-loss and one-hour recovery targets, three-hour initial cadence, proposed
retention and external age monitoring. Actual schedule adoption, backup,
notification delivery and timed isolated restore remain separate readiness proofs.

Protect the entire shared VM with Proxmox Backup Server on a physically separate host. This covers the three projects' PostgreSQL/RustFS volumes and the configuration required to restart them; the production application-level object-storage backup worker remains disabled while PBS is active. A host outage affects all three environments. PBS schedule and retention are deployment settings to record before readiness. Before public activation, restore a PBS backup into an isolated environment and verify startup, evidence references, retained versions and truthful updating state. Backup failures must remain visible to operators through PBS monitoring or its notification integration.

### Internal MVP milestone — authorized 2026-09-18

Import, preview and enable collection separately for vigilance, criticality/alert, monitoring and Calcinaia. Keep every source publicly disabled and use local administration for internal data inspection; do not bypass API/MCP publication gates to expose unaccepted data. Verify worker processing and retained evidence, record actual failures and consumption, and start observation only for activated revisions. Manual capacity checks must accompany bootstrap on the current 30-GiB disk. The seven-day evaluation and public release gates are unchanged; no synthetic elapsed trial or infrastructure acceptance is permitted.

### 12. Advance by verified scope and explicit dependencies

Preserve the Toscana regional three-product plus Calcinaia milestone. Maintain separate source-contract evidence, acceptance and public enablement for vigilance, criticality/alert, monitoring and Calcinaia. Each regional record identifies its selected operational channel, format and required graphics, access/reuse conditions, documented cadence, samples and unresolved conditions. The CFR homepage is only a secondary emission/adoption signal. DPC license evidence applies only to DPC internal comparison. An unavailable scope stays explicit in coverage; independently accepted and enabled scopes remain usable, but the service SHALL NOT claim the whole MVP is complete until all four scopes pass acceptance and are enabled.

For Calcinaia, record configured sections/channels, evaluated period, traversal and pagination, known-notice comparisons and required attachments. A successful complete check covers those configured boundaries only. Finding all retained samples does not prove municipality-wide completeness. Known omissions within a declared scope must be corrected and re-evaluated before its acceptance. Narrower accepted coverage must retain exclusions and limitations; it must not silently redefine the MVP target.

Track normative and geographic evidence by act/dataset, version, relevant provisions, verification date and applicable period, distinguishing publication, entry into force and operational applicability. Keep unresolved dates and amendment checks attached to the affected scope. An operator-maintained procedure records responsibility, official sources, review cadence, last/next checks and change triggers. New evidence prompts an impact review of rules, mappings and interpretations before controlled adoption or selected reprocessing; preserve earlier versions. This procedure does not add autonomous legal interpretation or source discovery.

Separate official municipality-zone mapping from the CAP candidate dataset. A missing usable CAP dataset leaves postal lookup explicitly unavailable while name/ISTAT selection and verified zone applicability remain functional. Do not infer candidates from unsupported sources. The CAP verification task remains open, and this fallback does not count as completion of the planned postal-data integration.

The dependency table in tasks.md governs sequencing: foundation, generic domain states and contract-driven behavior can advance using reviewed fixtures while source acceptance work remains open. Real source integration requires the relevant product contract; current normative/geographic assertions require applicable evidence. Live inference requires authenticated provider verification. The internal MVP and internal observation may run on the existing dedicated 8-vCPU/approximately-8-GiB/30-GiB VM after verified source configuration and successful previews. By the operator decision of 2026-09-18, disk expansion, SMTP, PBS and continuous telemetry are deferred for this internal milestone, with manual error/capacity checks and explicit absence of verified off-host recovery. These infrastructure requirements remain prerequisites for public readiness; internal observation evidence does not close them. Source acceptance still requires evaluation and the observational trial, and public activation remains a separate per-scope operator action.

## Risks / Trade-offs

- Missing upstream notices or changes: declare the checked sections and limitations; compare known real cases against each channel rather than claim total coverage from one platform.
- Ambiguous original documents: preserve the ambiguity and fetch necessary attachments; do not silently repair an authority's text.
- Platform dates and categories differ from original semantics: retain original evidence and validate fields independently for each channel.
- Unknown-duration measures can retain evidence indefinitely: expose unresolved status and measure retention cost; do not delete them solely because they are old.
- Provider outages, token charges and scanned-document quality: measure every attempt, test OCR evidence against originals, and use bounded retries with visible failures.
- Smaller extraction windows can separate relevant context or repeat facts across overlaps: retain deterministic core/context ownership, reject cross-window evidence and exercise boundary cases in the reviewed regression.
- OCR normalization can make a quotation appear detached from retained evidence: preserve the verbatim OCR result and a deterministic offset mapping from normalized text to its page before accepting any evidence reference.
- Compact evidence indices or conflict candidates can be malformed: validate ownership, bounds, literal spans, candidate identity and conflict membership locally; a failed required window leaves the document uninterpreted.
- An append-only extraction-schema migration can leave old and new result shapes side by side: bind every run to its schema/configuration version, preserve old runs and publish only an explicitly evaluated selected interpretation.
- Optional semantic retrieval can miss or overinclude candidates: retain baseline retrieval and compare results on the same evaluation set.
- Anonymous access can increase cost: publish configurable request/page limits and account for collection separately from query load.
- External assistants can omit uncertainty: provide structured evidence alongside each fact and document the integration contract; do not promise control over their generated replies.

## Remaining Verification Prerequisites

The 3 October preliminary Cittadino Informato review found a Calcinaia-linked
REST index and municipality/project responses. The subsequent user decision
replaces the original documentation/licensing prerequisite with an implemented
multi-source verification system. Task 26.3 validates the selected notice routes;
WordPress container dates and its empty REST body cannot substitute
for dynamically rendered notice versions. The [scoped review dossier](../../../docs/fonti/piattaforme/cittadino-informato-review.md)
records the original findings and revised completion criteria. Tasks 26.3/26.4
implement acquisition and verification; the subsequent bounded development run
closes 26.2 through inspectable receipts, rather than legal-basis documents.

The architecture and collection approach below are decided. These checks remain prerequisites for their dependent tasks, not evidence of completed implementation:

1. Complete and verify the product-specific collection contracts for the selected CFR routes and the municipal/additional-platform/referenced-act boundaries, including independent primary collection, required attachments, recorded multi-source comparisons and graphical zone/level extraction. Cittadino Informato readiness follows task 26.2's technical criterion; the other source/product contracts retain their own access/reuse requirements.
2. Implement conflict-aware operational extraction windows, compact indexed evidence and validated candidate merge, then rerun the full retained-case extraction evaluation, including long/scanned ordinances and cross-window temporal conflicts.
3. Complete live regional and municipal integration and verify the reviewed API/MCP contract, consistent views/cursors, retention behavior and the initial public limits behind the existing proxy.
4. Prepare the dedicated trial VM, configure the existing SMTP path, confirm PBS whole-VM protection, perform an isolated restore, complete at least seven full observational days across all four scopes and then derive the operating budget from measured consumption.

## Migration Plan

There is no production dataset to migrate. Add the conflict-candidate and compact-envelope persistence changes through append-only migrations and new prompt/schema/configuration versions; do not rewrite earlier extraction runs. Keep the currently selected processing configuration until the complete retained regression passes, then use explicit selected reprocessing rather than replaying all history. Rollback selects the previous compatible configuration while retaining candidate runs and acquired originals.

Continue from the existing implementation following the task dependency table. Prepare production as a third isolated Compose project on the current VM, then measure and expand host capacity before separately authorized startup behind the existing HTTPS reverse proxy; keep administration private through an SSH tunnel or the explicitly configured tailnet-only Tailscale Serve endpoint. For each scope, resolve its prerequisites, run retained-case evaluation, then conduct the shared observational trial and issue a separate acceptance result. Activate only verified scopes; others remain explicitly pending, and do not announce full MVP completion until all four are accepted. Before public activation, restore a PBS whole-VM backup into isolation and verify evidence links and honest updating status. Rollback restores compatible application/configuration versions without discarding acquired originals; source publication can be suspended independently. Deployment remains a separate authorized operation.

## Environment separation and release operations

Keep the existing Compose project and its PostgreSQL/RustFS data as development; do not rename its project in place. Manage development, staging and prepared production from one public checkout with different Compose project names, ignored environment files, secret directories, image references and loopback ports. Compose project scoping supplies independent named volumes and networks; parameterized secret paths supply independent credentials. Staging starts empty; release checks must populate it with controlled fixtures or another deliberately prepared, consistent private test dataset. Prepared production remains stopped, with all services behind explicit profiles and separate worker activation. No live source collection or paid inference starts unless deliberately configured. All three projects share host CPU, memory and disk, so measure capacity and avoid heavy development work during release checks. The historical separate-checkout setup remains recorded in task 12.3; tasks 12.6 and 12.8 replace that directory layout without replacing its data.

Build the application image once from a clean, identified revision in the shared checkout. Keep the checkout at that revision through staging validation and image publication. Run CI-equivalent checks, migrations and storage initialization against staging, then check service readiness, API/MCP behavior, private administration isolation, worker outcomes when the change needs them, and logs. The fault-injection Compose smoke test stops RustFS and belongs only on a disposable test project or in an explicitly planned staging interruption. Publish the tested image to public GHCR with a revision tag and record the registry digest. The operator approves that digest and the staging evidence; production on the shared VM pulls it and starts with image building disabled, separate named volumes and initial resource ceilings verified under load. Do not copy development or staging data to production as an implicit release step. Preserve the previous compatible image and configuration. Release approval is separate from source acceptance and publication gates.

Protect the entire shared VM with physically separate PBS storage, including all PostgreSQL/RustFS volumes, host configuration and private recovery files. The data-loss objective is six hours; schedule backups more often than that (initially every three hours), alert on missed or aging successful backups, and tune cadence and retention against measured backup duration and capacity. Keep the production application backup worker disabled while this whole-VM strategy is active. Before public activation, rehearse a timed restore into an isolated VM with outbound collection, provider calls and notifications disabled. Recovery time runs from outage detection to healthy public API/MCP access with verified database-to-object references and honest source-check ages; the target is one hour, which remains unproven until measured. A normal application rollback selects the previous compatible digest without reversing additive migrations; a corrupt or incompatible data state requires stopping writers and a coordinated PostgreSQL/RustFS recovery for all affected projects plus a separate incident decision about post-backup data loss.

## Open Questions

The concrete public hostname, private SMTP connection/recipient values, PBS retention, and final operating budget remain deployment-time selections or trial outputs. The destination class (off-host PBS), six-hour data-loss and one-hour recovery targets, initial three-hour backup cadence, initial request limits and use of semantic retrieval during the trial are decided. Local inference is an opt-in development fallback; no production local inference capacity is assumed.

## Optional local chat fallback

The local endpoint is disabled by default and names explicit sources. Selection occurs at the job-handler boundary before the selected runner creates a model-specific run, rather than changing providers inside a recorded call. Separate local catalog entries, gates and response receipts keep provenance, cache keys and unknown local cost distinct from the remote provider. Pre-claim holds permit work only when an eligible alternative is available. A newly established provider hold can schedule the alternative within the existing retry budget; validation/storage errors never become fallback triggers.

The evaluated local fallback covers text classification, extraction and linking, plus separately enabled image OCR. The semantic embedding endpoint remains independent; a chat model does not satisfy its 1024-dimensional embedding contract. Live tests use synthetic Italian notices and existing domain validators, while deployment evidence remains private. Development activation is limited to the selected source and does not enable public acceptance.

The evaluated local policy uses greedy decoding (temperature 0, seed 42) with thinking disabled, recorded in immutable stage settings and configuration identities. Local classification additionally records its reason-code mapping prompt. Source selection requires the evaluated classification/extraction output contracts. Vision is separately opt-in: a passing synthetic OCR test and enabled image input are prerequisites. Equivalent-copy materialization first tries a compatible primary cache, then an eligible local cache on a specific reuse miss, with both transports prohibited regardless of gate state. Semantic embeddings retain their own provider and can still delay linking.

Local classification tolerates typographic double-quote presentation through a separately versioned contiguous matcher that returns the original source substring and preserves Unicode byte offsets. Remote and legacy parsing remain unchanged; invented words and disjoint quotations remain rejected. Local extraction records a concise prompt with explicit civil-protection opening rules, validated on long notices as well as isolated operative sentences. Catalog names/revisions isolate concurrent remote and local configurations, including distinct text/image model capabilities; immutable earlier local versions remain auditable.

### Development fallback concurrency trial — 2 October 2026

Use two replicas of the existing worker to exercise a local server's two slots.
Each replica runs one inference job at a time; PostgreSQL claims coordinate jobs
and scheduled source checks, while notification delivery uses durable locks.
Verify the resolved image/settings against the current worker, retain a database
archive and add the replica with `--no-recreate --no-deps --no-build --pull never`.
This uses the existing runtime rather than introducing another inference policy.
The count also applies to primary-provider work after recovery and scales the
other worker loops. It does not cap unrelated server clients. Observe overlapping
receipts, validated results and worker stability; slot occupancy alone does not
establish a throughput improvement. Keep operational evidence private, leave the
repository default at one replica and document explicit scale-down recovery.

The subsequent operator request expands development to four replicas using the
same image/settings without recreating the first two. Check current server slots
again: worker count and simultaneous server executions are independent. When the
server exposes fewer slots, requests may queue there. Record this capacity limit
and the explicit return to two workers alongside runtime verification.

Four-worker observation reproduced deadlocks in the pre-claim bulk queue updates,
exhausting worker recovery and causing process restarts. Wrap the complete
recovery/equivalence/provider-hold maintenance callback in a PostgreSQL
transaction advisory lock (`730072`). The transaction reserves one pool
connection; callback queries use the remaining pool connections. Hold the lock
only for queue maintenance and release it before claims or provider calls.
Commit on success; bounded independent-context rollback releases ownership after
errors or cancellation. Integration checks exercise four competing callbacks,
parallel work after maintenance and lock release after failure/cancelled waiting.
Deploy the corrected image to all four replicas so they share the lock protocol;
an older worker does not coordinate on this lock.

The next operator request expands the corrected development worker to six
replicas. Add the two replicas without recreating the existing four or changing
their image/settings. Verify current server capacity, running queue loops and
restart/deadlock observations; preserve private rollback evidence and document
`--scale worker=4` as the return to the previous count. The existing lock protocol
and source/provider controls apply to all six replicas.

## Partial municipal reopening recovery — 4 October 2026

The retained local replay fails because a canonicalized evidence quote preserves
`d’Arno` while its place still uses `d'Arno`. Return the contiguous retained place
spelling from contextual validation before list-scope checks; keep the existing
known/null-place merge conflict guard. Complete only recognized standalone,
core-owned sentences asserting that a named road/underpass/bridge/tunnel has been
reopened to circulation. Rebind a matching generic update to that explicit
predicate, or supplement the missing reopening, with literal evidence. The
recognized sentence has no reopening time: a parenthetical earlier closure or
page timestamp does not supply one. Reject uncertain subjects, quoted/conditional
statements and ordinance headings. Preserve the legacy parsing route.

Register immutable extraction configuration v25 with logic v20 for remote and
local models, partitioning manifests/reuse from prior interpretations without
changing the prompts or automatically replaying history. Synthetic parser and
durable PostgreSQL cases plus the retained-response replay verify this correction;
ordinary live reprocessing and campaign/acceptance reassessment remain separate.
See [municipal interpretation](../../../docs/operations/municipal-interpretation.md).

## Territorial source knowledge guides

Maintain a human-readable knowledge base under `docs/territori/`, with region directories keyed by two-digit codes and municipality files keyed by six-digit ISTAT identity. Create files when research begins. Each guide puts current source maps, operational rules and open problems before its dated discovery register. A shared template records stable finding IDs, scope, evidence dates, knowledge status and intervention status independently. Subsequent entries reference superseded findings without deleting them.

Keep shared platform notes under `docs/fonti/piattaforme/` and link them from the municipality that supplied the evidence. Platform observations do not transfer permissions or collection exceptions to other territories. Seed Toscana and its four investigated municipalities with public references, code behavior and explicit verification limits; the Livorno attachment/footer findings describe proposed collector work, not an implemented fix.

The public coverage tracker remains the acceptance/status reference with its existing snapshot preserved. Environment-specific configurations, detailed outcomes and unpublished captures remain under ignored `.local/operations/territori/`; public guides contain only redistributable summaries and links. The local archive requires separate private retention/backup. Contributor guidance requires consulting and updating the relevant guide during source work. Documentation validation checks paths/anchors, territorial identities, finding IDs and evidence boundaries, without launching collection or inference.

## Scoped municipal attachments — 2 October 2026

An optional `attachments` contract belongs to the immutable source revision. `content_class` selects the document content for linked PDF discovery; a missing container fails visibly and footer/navigation links are excluded. Existing configurations preserve their previous discovery, and explicit registered dependencies are not silently removed. Each `external` scope declares an exact HTTPS `origin`, canonical directory `path_prefix`, official `referral` evidence and its own collection/retention `policy`. These permissions cover linked attachments only and do not enable another source or a new channel. Livorno's reviewed case uses its municipal resource host and directory; actual operational settings remain private.

Preview and scheduled collection share document/attachment retention. `DirectHTTP.CrawlBounded` validates the initial URL and each redirect before sending the request, retains the existing prohibition on leaving the file's initial origin, and constrains paths using the source contract. Opt-in `validate_pdf` checks response type, PDF framing and the Poppler parser already shipped in the application image. HTML error pages and truncated/unreadable PDFs become visible missing references; persistence remains content-addressed and verified by the document store. Scanned PDFs do not require extractable text for acquisition. Missing-resource reasons and publication metadata participate in acquisition identities to keep diagnostic changes replayable and prevent collisions between previews and later dated collection. Invalid PDFs use the existing unavailable-reference vocabulary and the distinct scheduled error `invalid_attachment_pdf`, avoiding a storage-schema change.

Use a reviewed candidate revision and successful preview before changing active collection. Preserve original evidence and prior revision/image for rollback, then verify a fresh normal worker check and stored attachment bytes. Existing backoff, source retry instructions, provider controls and public-enablement gates remain separate. A development recovery does not establish acceptance or future remote availability.


## Local processing before model inference — 2 October 2026

A versioned optional `local_processing` source policy requires dated review
evidence. It declares simple article-body selectors, text-only PDF directory
scopes, and/or one recognized regional product format. No existing source is
opted in by this implementation. This is an interpretation policy, not permission
to discover or acquire resources. It follows the active reviewed source revision,
like resource eligibility. Policy fingerprints partition run, preflight and result
reuse identities, and separate immutable local-first processing configurations
preserve the original model configurations as fallbacks. Policy changes do not
replay historical work automatically.

Article selectors are a bounded subset (tag, #id, .class, tag.class, tag#id).
All configured selectors must match exactly one nonempty container; distinct
containers are combined in document order without duplicating nested content.
Missing, ambiguous or empty containers fall back to full-page text. Attachments
and original bytes remain intact. A configured selector asserts the reviewed
full article scope; fixture checks cannot certify an actual website's coverage.

PDF native text is used only in explicitly reviewed text-only path scopes. Local
Poppler commands read private temporary files under bounded contexts and output
limits. Require readable text for every numbered page and reject any raster
images, encrypted content, malformed output or invalid Unicode. Otherwise retain
the existing image/OCR path for the entire resource. Vector graphics cannot be
reliably excluded by text extraction, so mapped/graphical products must never be
declared text-only. Persist original URL, resource hash, page number and local
extractor identity; no provider call, token usage or remote charge is invented.

For a configured regional format, reuse the existing strict HTML projection.
Only a successfully parsed explicit fact/table or explicit monitoring no-event
statement can establish a positive regional relevance result. Complete retained
content and literal evidence remain mandatory. Unsupported layouts, incomplete
resources and all unresolved or ordinary notices use normal full-content model
classification. No keyword-based negative filter is added. Relevant notices
continue through the existing extraction pipeline; this does not replace
semantic interpretation of municipal measures or graphical evidence. Existing
provider admission/holds remain in force, including pre-claim holds.

Validate synthetic selectors, attachment coverage, misleading/late notices,
text/scanned/mixed/graphical PDF fallbacks, literal regional facts, result/run
provenance, policy-separated equivalence and durable persistence on disposable
PostgreSQL. Source-specific review/activation and live quality observation remain
separate from this repository implementation.

Native PDF page inputs require additive migration `066_ocr_native_text`, expanding
only the page-result media constraint. Prior images, records and migration
checksums remain valid. Article selectors apply to original HTML notices; HTML
attachments retain their full text. Native and model page writes preserve their
selected path across retries rather than mixing extraction methods in one run.
See [local processing operations](../../../docs/operations/local-processing.md).


## Bounded local-processing verification — 2 October 2026

A read-only retained-development replay compares candidate selectors with full
page text, checks title/article/metadata and attachment preservation, exercises
missing/ambiguous selector fallback, and probes unique municipal PDFs with real
Poppler. It separates technical native-text eligibility from approval of a
text-only directory. A validated vigilance table remains relevant when its six
phenomena have no listed zones: use the literal regional bulletin title only
after strict table/date recognition, without inferring an all-clear.

For the explicitly requested development trial, use eight fixed versions and a
separate queue with one attempt per job and at most twenty calls to the existing
local model. Trial document readers attach reviewed candidate policies only to
those versions and exact reviewed PDF hashes. Processing stores, manifests,
queue claims and call receipts are real; policy identity separates results from
ordinary reuse. Four municipal baseline/candidate pairs use the same model;
regional classification and reviewed native PDFs exercise zero-call paths.
Preserve a private database snapshot and input/output evidence. Existing source
policies and normal worker routing remain intact; no automatic downstream
extraction, embedding, linking or publication is requested by this diagnostic
worker. The worker exits after the fixed queue, with results retained for admin
inspection. This bounded trial is separate from a continuing source rollout or
semantic extraction acceptance.

## Cittadino Informato acquisition and corroboration — 3 October 2026

The revised user-authorized target is a scoped acquisition/discovery channel with
an implemented verification system spanning Regione Toscana/CFR, municipality and
cittadinoinformato.it. Task 26.2 no longer requires obtaining license, agreement or
legal-basis documentation. Calcinaia has a verified municipal referral; each other
municipality requires its own recognition and explicit selection. Use a separate
channel/source identity, bounded public sections/endpoints, operator-configured
cadence and upstream retry instructions. Prefer technically verified API/feed
access. Validate the notice routes independently of CMS container metadata.

Acquisition creates attributable candidates and bounded verification work,
independent of anonymous API/MCP queries. Persist a receipt identifying the
candidate, all three channel roles, consulted source/version evidence and passages,
attempted-check times, compared fields, outcome and reasons. Distinguish supported
fields, primary evidence not found, unavailable checks, non-comparable evidence,
checks not applicable to the claim and actual conflicts. Preserve earlier receipts
and reevaluate changed versions without rewriting historical knowledge.

Compare municipal actions, places and original validity expressions against
municipal notices or referenced acts; compare regional republications against the
originating CFR product, risk, zone, issuance and validity. CFR may be not applicable
to a local-only measure; municipal republication is not required for a supported
regional claim. Do not require three-channel unanimity or count republications as
independent majority votes. Unsupported candidates remain diagnostic and cannot
create confirmed municipal fields or overwrite the originating CFR risk level.
Deduplicate matching publications and preserve each channel's evidence.

Continue independent municipal and CFR discovery, including primary-only notices.
Equivalent API/MCP output retains channel attribution, primary evidence and scoped
verification outcomes. Public platform copies remain link-only. Task 26.2 closes
only after a bounded development run exercises acquisition and persistent
verification for regional and local cases, and failure/conflict cases pass with
inspectable receipts. Tasks 26.3 and 26.4 supply the acquisition and persistence
implementation; task 26.5 extends evaluation and task 26.6 controls the subsequent
scheduled trial. This planning revision implements no collector, completes no
runtime task and does not accept or publicly activate a source.

Task 26.3 implementation uses the typed `cittadino_informato` opt-in contract,
separate municipality-bound external source and direct HTTP adapter. The register
and retention store enforce municipality/query/dependency boundaries. Retain exact
API JSON under its API URL, with public notice links and date meanings in metadata;
only explicit `data_pubblicazione` supplies the source publication date. The optional
absolute display-start filter limits discovery without establishing publication or
operative coverage. Bounded pagination, existing target/version tracking,
unavailable-detail backoff and validated PDF dependencies preserve visible failures
and recoveries. Operator selection/referral supplies configuration evidence without
a license-document prerequisite. Notices and risks remain pending verification.
Synthetic persistence tests and a bounded isolated Calcinaia HTTP run verify this
acquisition step; its original pending-verification state is preserved. The subsequent
26.4 implementation below adds receipts/admission; scheduled activation remains separate. See [operations](../../../docs/operations/cittadino-informato.md).


Task 26.4 adds an evidence-selector verification API and private admin CLI over
retained originals/completed OCR. Values are derived from owned resources, not
accepted as caller-supplied facts; exact local act/edition or regional
product/risk/zone/absolute issuance/validity establishes comparability. Receipts
are immutable and idempotent per request, record all roles and comparison
source/version directions, and commit atomically with primary-only domain
projection. Unsupported platform fields are never copied. A supported primary
can be projected independently of a discordant republication or a missing
platform notice; monitoring adds no alert color. Native domain store writes
cannot admit a Cittadino Informato-origin fact directly.

Stable primary/version facts are reused. Append-only scope/acquisition mappings
select the current verified revision within the query's knowledge boundary,
including a return to an earlier content hash, without removing historical
facts. This does not infer cancellation between different acts. Local projection
records field-supported interpretation separately from provenance/acceptance.
Retention includes bidirectional proof dependencies while any counterpart is
protected, then purges fully expired graphs. Synthetic PostgreSQL and public
query tests verify these behaviors. The entry point consumes interpreted
selectors/check results. Task 26.2 now verifies that entry point after independent
real-source acquisition on development PostgreSQL/RustFS with dedicated data.
Ten receipts distinguish three real-source cases from seven synthetic controls;
incomplete external platform dependencies and unverified CFR map colors remain
explicit diagnostics, with supported local primary fields admitted separately.
The wider failed platform window is preserved, and the successful display window
and exact local detail have separate boundaries. Request retry/history and primary
fact reuse are checked without enabling scheduled sources or public copies.
See [bounded development verification](../../../docs/operations/cittadino-informato.md#prova-completa-in-development--task-262).
Automatic counterpart discovery, worker orchestration and broader reviewed
real-source/API-MCP evaluation remain separate work in 26.5/26.6; the generic
worker and existing live source controls were preserved.

## Manual municipal development publication — 2 October 2026

Use additive migration `067_development_publication` for default-off municipal choices and immutable events, with no automatic backfill or rewrite of source acceptance/public flags. Private admin forms and explicit CLI commands use expected revisions. The current municipality register determines eligibility, and historic choices can be revoked. `IWA_ENVIRONMENT` defaults to strict production and Compose pins each environment. The development query store alone adds a municipality-scoped visibility path for policy-permitted active configurations and retained acquisitions under that revision. Municipalities sharing a regional zone do not inherit each other's selection; unscoped regional facts are restricted to mappings selected at the query knowledge boundary. Regional bulletin metadata remains attributable to the shared original. Copy authorization stays separate.

Municipality/coverage response fields identify development publication; acceptance remains pending where pending and quality limitations disclose the manual override. No uncertain regional map color or missing municipal measure is invented. A publication scope stored privately with each saved view binds it to environment, municipal revisions, region revisions, source active/public settings and territorial associations. API/MCP reject views after scope changes, including revocation and a copied database served by a strict runtime. Production source gates remain independently testable. Verification uses synthetic facts, native forms, API/MCP views, disposable PostgreSQL and resolved Compose templates; live development activation is recorded separately from acceptance.

## Attributable document review and public receipts — 3 October 2026

The operator explicitly delegates the remaining document review to the assistant.
Preserve historical human approvals and record new assistant analysis against
the original PDF pages, retained HTML/JSON and actual output. Identify the review
method, actor, evidence date and unresolved fields rather than relabeling model
output or the delegation itself as approval of the facts. Software comparisons
against expected outcomes remain a separate step. Failed cases still prevent
acceptance.

Public document, measure, regional and phase records may carry the latest 100
publishable verification receipts at the same knowledge boundary. Apply source
publication and development municipality selection to the candidate and each
evidence channel. Redact unpublished evidence identities, values and selectors
while retaining the channel role/outcome with a limitation. Exclude request keys,
OCR run identities and operator-only notes. Expose comparison direction, literal
passages and source/version/hash attribution only for visible evidence. API and
MCP share these structures and limits; candidate admission is distinct from the
independently supported primary projection and source acceptance.


## Live municipal integration repair — 3 October 2026

The deployed situation can return no local facts despite stored successful extraction. Add a primary municipal projection at the ordinary extraction scheduling boundary, before downstream linking, and an update-link projection after linking. Both use retained results without additional model calls. Preserve atomic measure/evidence binding, literal temporal expressions, knowledge dates and source suspensions. Evaluation runs remain separate. Carry the reviewed official-provenance evidence from the active configuration into an independent append-only assessment without overriding explicit existing assessments. Public coverage summarizes ordinary interpretation of latest visible notices independently of acceptance. First pending notices must remain visible without prior measures; configured listing containers and pagination are excluded. The public subject accompanies action/place. Deployment and bounded retained-data recovery are verified separately from code tests. Graphical CFR interpretation remains an explicit follow-up and cannot be replaced by forced green values.

Local worker startup must tolerate several immutable fallback revisions sharing a catalog name. Preserve legacy local registrations where compatible; when the legacy name/revision key collides, derive a separate identity and name from the exact fallback. Never rewrite an existing configuration to make startup pass. Verify both fallback associations, repeated registration and preservation of prior hashes.


## Retained criticality vector interpretation — 3 October 2026

The current Calcinaia query needs per-risk results rather than fourteen unknown
map entries. Extract PDF text bounding boxes and vector paths with the already
packaged Poppler utilities, without model calls or external downloads. Match
issuance before combining HTML and PDF. Associate each risk/date heading with its
26 labelled zone polygons, applying each path's affine transformation and a
point-in-polygon test at the label. Recognize the official fill palette; white
outlined polygons denote excluded main-river/coastal mask areas. A missing or
ambiguous association stays unknown. Bound PDF size, command time, page count and
output; unsupported vector layouts fall back to supported HTML evidence.

Keep explicit HTML alert intervals when matching maps cover the same days;
deduplicate repeated intervals and reject conflicting colors. New parser records
use `cfr-vector-v3`, preserve older projections and become visible only at the
actual knowledge time. Encode PDF URL/page/locator in the existing evidence
locator and expose it through the shared public query. Limitations concerning
unresolved labels elsewhere in the bulletin remain in source coverage, while an
unambiguously supported zone's fact is assessed separately. This verifies the
current Calcinaia criticality scope, not all label placements, vigilance graphics,
municipal expiry or source acceptance.

## Complete retained CFR graphics — 4 October 2026

Task 27.4 extends the preceding bounded interpreter to every labelled zone in the
recognized retained products. `cfr-graphics-v4` is an independent append-only
projection revision. Identify map headings by their following two-day row, match
issuance anywhere in the original PDF and require issuance-day/next-day maps.
Accept separately rendered irregular fills and only reviewed, tightly bounded A6/I
label offsets with an unambiguous outline; anonymous rectangles and arbitrary
nearest-zone choices are insufficient. Retain excluded river/coastal masks.

Vigilance uses its own five panels, dates, eight rainfall bands and labelled
symbol legend. Carry optional weather metadata through the existing evidence
locator into shared search and situation facts: daily rainfall band, separate
cumulative band/period, millimetres, area-average scope, graphical status and HTML
evaluation membership. Rainfall represents both hydrological risk categories;
vigilance never acquires criticality colors. An absent symbol means not depicted,
never absence of risk. Unknown marks or listed-but-unrecognized phenomena remain
unresolved. Raster maps, changed legends and unsupported geometry fall back to
supported HTML evidence. Aggregate source limitations do not invalidate otherwise
supported zone facts; interpretation support does not confer source acceptance.

Bound Poppler operations and add substitute fonts. If direct SVG rendering fails,
discard its output, render a local single-page vector PDF and retry SVG while
preserving the original physical page. No remote resource or model call is used.
Validate two retained criticality and two vigilance samples, separately identify
the historical non-green PDF's reviewed HTML transcription, and retain private
hashes/results. Synthetic controls cover positive symbols for all weather
categories; real positive wind/sea/snow/ice cases remain acceptance evidence to
collect. Disposable PostgreSQL and API/MCP tests verify history, idempotence,
missing-resource fallback and metadata. Live adoption remains separate.

## Concise situation presentation — 3 October 2026

Keep the shared domain query and immutable complete snapshots. At delivery, refresh
source updating against served time, build a deterministic concise presentation
from the full pinned situation, and paginate only regional facts, local measures
and phases. Full-scope conclusions, all three source summaries and deduplicated
evidence references repeat on each page. Pending unstructured notices affect the
municipal completeness statement but no longer produce hundreds of public rows
or additional situation pages. Detailed document/quality/verification operations
remain available. Source identity uses the configured channel platform, never a
source-name heuristic; unavailable source evidence is not cited.

`processed_data` carries an attributed factual summary, separate regional/local/
phase facts with precise validity, field limitations and compact reference IDs.
`source_summaries` contains `regional`, `municipal` and `cittadino_informato`, each
with content, availability and individual product/check/acceptance summaries.
Missing platform data means not collected in the publication view, never no
notices. A municipal statement lists documented measures and distinguishes known
current validity from undetermined validity. Known regional green does not establish
municipal all-clear. Regional products and zones retain separate applicability.

The response shape intentionally replaces the previous situation data contract;
HTTP route, envelope, search/document/coverage and saved snapshot identity stay
the same. New cursors use processed-data collection paths. Earlier situation
cursors with old collection positions fail explicitly with cursor_mismatch and
require restarting; a retained dataset_version can render the same pinned domain
knowledge in the new presentation. API/MCP share the same rendering and schemas.

## Selected archive recovery — 4 October 2026

The operator authorizes recovery of exact archived versions with an immutable
audit. Keep `interpretation_archives` effective for automatic scheduling and
ordinary work. Add independent recovery requests linked to exact reprocessing
selections, the active source revision, retained acquisition and the latest
passing regression. Only the new classification carrying that selection and
its extraction/embedding/linking descendants may pass the archive guard.
Do not delete archival rows, reset historical attempts or reuse old job payloads.
Store the complete selection atomically before enqueueing; identical request
retries resume enqueue gaps through the existing durable job identities.
Private API writes and readback expose the audit; public interfaces expose no
recovery controls. Source acceptance and remaining campaign reviews are separate.
