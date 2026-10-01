## Context

See [proposal.md](proposal.md) for scope and motivation. The workspace contains a substantial implementation and tests alongside the research and planning artifacts. This revision captures the agreed target behavior and deployment design; completed code tasks do not by themselves certify live source acceptance or operational readiness.

Research context (the detailed investigation records are retained separately):

- National and Toscana legal responsibilities, effective dates, and later amendments require continuing verification.
- The operational CFR routes selected for the MVP are the vigilance page at `https://cfr.toscana.it/index.php?IDS=2&IDSS=71`, the criticality/alert page at `https://cfr.toscana.it/index.php?IDS=2&IDSS=76`, and monitoring at `https://www.cfr.toscana.it/mobile3/avviso-criticita/`. The CFR homepage is a secondary emission/adoption signal. Graphical resources, observed HTTP 429 behavior, and product-specific extraction remain part of acceptance.
- Cittadino Informato and the Calcinaia municipal site have non-equivalent dates and coverage. An explicit municipal referral does not make the secondary platform a complete primary source.
- Livorno, Pisa, Pontedera, and Cascina are subsequent municipal rollout targets. Calcinaia is the initial local channel; each municipality requires separate source acceptance.

The latest agreed requirements are in this change's three specs. Research notes remain dated evidence; not all later conversational decisions have been copied into docs. Firenze remains research material, not an initial local collection target.

The authenticated task-9.1 runs after the first bounded-segmentation implementation exposed two remaining constraints. Prompt-only and greedy/non-thinking candidates can improve individual cases but do not make the full regression stable, and a verbose per-field evidence envelope can exceed the 90-second worker window on an ordinance page containing many measures. The current extraction result also cannot retain two supported temporal candidates for one field before domain projection, so rejecting the entire merge or selecting one candidate both lose required information.

## Goals / Non-Goals

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

A changed duration applies to existing data at the next cleanup: increases protect remaining data longer, reductions make older unprotected data eligible, and deleted data is not restored. When protection ends, already-expired versions become eligible. Define calendar arithmetic and dependency-safe cleanup in the technical contract, including month-end cases. Retention cannot silently discard evidence required by an exception. Source-specific reuse constraints remain an acceptance condition and any resulting limits must be disclosed.

### 6. Diagnose quality rather than manually curate notices

Internal monitoring assesses source availability, update timeliness and interpretation correctness. Keep reproducible cases and expected outcomes for code/process evaluation, including false positives and missed publications. DPC supplies internal comparisons only where risk, territory and time are actually comparable. Successful HTTP responses and LLM self-assessment alone do not establish correctness.

Build human-reviewed expected outputs for real Calcinaia notices, irrelevant content, partial reopenings, conflicting dates, attachments and scanned PDFs; label simulations separately. A missed relevant notice or unsupported measure in evaluation blocks public activation of the affected source until correction and successful rerun. Also measure unresolved fields so an extractor that returns no useful facts cannot pass on precision alone. Run at least seven complete 24-hour observational intervals across vigilance, criticality/alert, monitoring and Calcinaia, persist daily checks, and compare originals and attachments manually. Retain cases for significant events absent during the trial. Extend the trial for omissions, delays, missing evidence, unresolved defects or insufficient observations. Source acceptance is distinct from any per-notice editorial gate.

### 7. Modular Go service and durable storage

Use one modular Go application with background worker processes for collection and interpretation; API/MCP requests only read prepared data. PostgreSQL stores the domain model, configuration history, interpretation runs and a durable job queue with status, attempts and errors. Workers recover incomplete jobs after restart and use idempotent processing to avoid duplicate versions or public effects. This keeps scheduling and records in one persistence system; SQLite and an additional queue service are not the selected MVP design.

Store original HTML, necessary page resources, PDFs and scan images in RustFS through S3, with hashes and object references in PostgreSQL. RustFS remains internal to the service VM. A service setting, disabled by software default, controls public copy access through the application; the initial deployment enables eligible CFR/Regione Toscana copies and direct Calcinaia copies only where the recorded source policy permits them. Cittadino Informato and albo pretorio remain link-only. Official links remain present. Database/object writes, cleanup and backup must preserve referential consistency.

Docker Compose runs Go processes, PostgreSQL, RustFS and local Crawl4AI on a dedicated VM initially sized at 4 vCPUs, 8 GiB RAM and 80 GiB disk, with no GPU requirement. OCR/inference use external APIs initially; local replacements remain possible. Review capacity when sustained disk or memory consumption exceeds 70%, and resize from measurements rather than treating the initial allocation as a permanent limit.

### 8. Configured discovery, OCR and separate LLM stages

Operators configure sources and sections through administration; the service discovers documents within those boundaries rather than autonomously discovering new channels. Crawl4AI is an internal acquisition component called by Go workers. Preserve original bytes and required resources alongside extracted text. For Calcinaia, the direct municipal site is the primary collection source, Cittadino Informato is a secondary sentinel used to diagnose coverage differences, and the albo pretorio is queried only for acts referenced by primary notices. A secondary-only observation cannot establish a municipal fact or completeness without primary evidence.

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

Protect the entire dedicated VM with Proxmox Backup Server on a physically separate host. This covers PostgreSQL, VM-hosted RustFS and the configuration required to restart the service; the existing application-level object-storage backup worker may remain disabled. PBS schedule and retention are deployment settings to record before readiness. Before public activation, restore a PBS backup into an isolated environment and verify startup, evidence references, retained versions and truthful updating state. Backup failures must remain visible to operators through PBS monitoring or its notification integration.

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

The architecture and collection approach below are decided. These checks remain prerequisites for their dependent tasks, not evidence of completed implementation:

1. Complete and verify the product-specific collection contracts for the selected CFR routes and the Calcinaia primary/sentinel/referenced-act boundaries, including required attachments, reuse conditions, known-publication comparisons and graphical zone/level extraction.
2. Implement conflict-aware operational extraction windows, compact indexed evidence and validated candidate merge, then rerun the full retained-case extraction evaluation, including long/scanned ordinances and cross-window temporal conflicts.
3. Complete live regional and municipal integration and verify the reviewed API/MCP contract, consistent views/cursors, retention behavior and the initial public limits behind the existing proxy.
4. Prepare the dedicated trial VM, configure the existing SMTP path, confirm PBS whole-VM protection, perform an isolated restore, complete at least seven full observational days across all four scopes and then derive the operating budget from measured consumption.

## Migration Plan

There is no production dataset to migrate. Add the conflict-candidate and compact-envelope persistence changes through append-only migrations and new prompt/schema/configuration versions; do not rewrite earlier extraction runs. Keep the currently selected processing configuration until the complete retained regression passes, then use explicit selected reprocessing rather than replaying all history. Rollback selects the previous compatible configuration while retaining candidate runs and acquired originals.

Continue from the existing implementation following the task dependency table. Deploy Docker Compose to the dedicated 4-vCPU, 8-GiB, 80-GiB VM behind the existing HTTPS reverse proxy; keep administration private through an SSH tunnel or the explicitly configured tailnet-only Tailscale Serve endpoint. For each scope, resolve its prerequisites, run retained-case evaluation, then conduct the shared observational trial and issue a separate acceptance result. Activate only verified scopes; others remain explicitly pending, and do not announce full MVP completion until all four are accepted. Before public activation, restore a PBS whole-VM backup into isolation and verify evidence links and honest updating status. Rollback restores compatible application/configuration versions without discarding acquired originals; source publication can be suspended independently. Deployment remains a separate authorized operation.

## Environment separation and release operations

Keep the existing current-VM Compose project and its PostgreSQL/RustFS data as development; do not rename its project in place. Manage development and staging from one public checkout with different Compose project names, ignored environment files, secret directories, image tags and loopback ports. Compose project scoping supplies independent named volumes and networks; parameterized secret paths supply independent credentials. Staging starts empty; release checks must populate it with controlled fixtures or another deliberately prepared, consistent private test dataset. No live source collection or paid inference starts unless deliberately configured. The two projects still share host CPU, memory and disk, so measure capacity and avoid heavy development work during release checks. The historical separate-checkout setup remains recorded in task 12.3; task 12.6 replaces that directory layout without replacing its data.

Build the application image once from a clean, identified revision in the shared checkout. Keep the checkout at that revision through staging validation and image publication. Run CI-equivalent checks, migrations and storage initialization against staging, then check service readiness, API/MCP behavior, private administration isolation, worker outcomes when the change needs them, and logs. The fault-injection Compose smoke test stops RustFS and belongs only on a disposable test project or in an explicitly planned staging interruption. Publish the tested image to public GHCR with a revision tag and record the registry digest. The operator approves that digest and the staging evidence; production on the separate VM pulls it and starts with image building disabled. Do not copy development or staging data to production as an implicit release step. Preserve the previous compatible image and configuration. Release approval is separate from source acceptance and publication gates.

Protect the entire production VM with physically separate PBS storage, including PostgreSQL, RustFS, host configuration and private recovery files. The data-loss objective is six hours; schedule backups more often than that (initially every three hours), alert on missed or aging successful backups, and tune cadence and retention against measured backup duration and capacity. Keep the application backup worker disabled while this whole-VM strategy is active. Before public activation, rehearse a timed restore into an isolated VM with outbound collection, provider calls and notifications disabled. Recovery time runs from outage detection to healthy public API/MCP access with verified database-to-object references and honest source-check ages; the target is one hour, which remains unproven until measured. A normal application rollback selects the previous compatible digest without reversing additive migrations; a corrupt or incompatible data state requires a coordinated PostgreSQL/RustFS recovery and a separate incident decision about post-backup data loss.

## Open Questions

The concrete public hostname, private SMTP connection/recipient values, PBS retention, and final operating budget remain deployment-time selections or trial outputs. The destination class (off-host PBS), six-hour data-loss and one-hour recovery targets, initial three-hour backup cadence, initial request limits and use of semantic retrieval during the trial are decided. Local inference remains a future option; no local inference capacity is assumed.
