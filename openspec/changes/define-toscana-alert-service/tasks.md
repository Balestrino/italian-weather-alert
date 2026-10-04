# Public progress snapshot

The original checkboxes reflect the private project task list on 30 September 2026. Section 12 records new public-checkout release work and must be verified independently. A checked item records project work; it does not by itself establish public source acceptance, public deployment, or that unpublished evidence is available in this repository. Operational notes and private evidence references are maintained separately.

## 1. Verify evidence and external contracts

- [x] 1.1 Consolidate applicable national/Toscana acts and amendment chains in a source-linked evidence table; verify each used provision against retained official text and identify unresolved consolidation without asserting current applicability. Effective dates and continuing verification are tracked in 1.9–1.10.
- [x] 1.2 Finalize configured Calcinaia channels, sections and recognized external referrals; deliver a scope matrix with issuer/publisher distinctions and institutional competence evidence only where needed for acts attributed to an associated entity. Verify each external channel has a scoped official referral; traversal and access/reuse checks are tracked in 1.11–1.12.
- [x] 1.3 Complete the vigilance source contract for `https://cfr.toscana.it/index.php?IDS=2&IDSS=71`, using the CFR homepage only as a secondary emission/adoption signal; verify formats, necessary graphical resources, access, product-specific reuse, documented daily publication window and 429 behavior. Deliver representative retained samples and compare levels, zones and validity against originals. Close only when these checks pass; service acceptance and trial remain in section 9.
- [x] 1.4 Reconcile official versioned municipality-zone mappings with ISTAT; deliver an attributed mapping with territorial precision, applicable-period evidence and multi-zone/fusion cases. Verify reconciliation and explicitly resolve or bound unsupported mappings before use; postal-code data is tracked separately in 1.16.
- [x] 1.5 Verify Regolo.ai support and exact identifiers/contracts for Qwen3.8-27B and DeepSeek-OCR-2, including scanned-document inputs, authentication, limits, usage and pricing fields; deliver reproducible probe results without secrets and report incompatibilities before changing provider/model choices.
- [x] 1.6 Inspect the Calcinaia closure ordinance and assemble human-reviewed expected outcomes for relevant/irrelevant notices, partial reopening, retained restrictions, conflicting dates, unreadable/scanned PDFs and later uninterpreted documents; distinguish observed cases from simulations and retain supporting passages/pages.
- [x] 1.7 Specify the concrete API/MCP schemas and transports for the five agreed operation groups, temporal precision, cursor consistency/expiry, retention calendar boundaries and shared-IP accounting; verify the contract against the specification scenarios, including month-end retention and timezone ambiguity.
- [ ] 1.8 Prepare the dedicated trial VM with an initial 4-vCPU, 8-GiB RAM and 80-GiB disk allocation; configure the existing SMTP relay/operational mailbox privately, confirm that off-host PBS covers the complete VM, and record the chosen PBS schedule and retention before readiness. Verify capacity telemetry and the sustained 70% disk/RAM review threshold; no GPU or separate off-host S3 bucket is required.
- [x] 1.9 Verify publication, entry into force and operational applicability of provisions and geographic versions identified in 1.1/1.4, including unresolved BURT and regional-act dates; deliver linked date evidence, affected scopes and explicit unresolved intervals without substituting acquisition dates.
- [x] 1.10 Define the continuing normative verification procedure: record an operator owner, official sources, review cadence, last/next checks and amendment-review triggers; exercise it on a retained amendment and deliver an impact record for affected rules/mappings, evaluation and controlled adoption. This is an operator procedure, not autonomous legal interpretation or a new source-discovery subsystem.
- [x] 1.11 Verify Calcinaia over the declared bootstrap period using the direct municipal website as primary source, Cittadino Informato as a secondary discovery/coverage sentinel and the albo only for acts referenced by primary notices. Compare all known relevant notices, including the previously missing secondary-platform sample, and retain URLs, date anomalies, traversal boundaries and diagnostic gaps. Close only when known omissions in the declared primary scope are corrected and rechecked; secondary-platform coverage must not establish completeness or municipal facts by itself.
- [x] 1.12 Verify access/reuse and required attachment acquisition separately for the municipal site and each configured external/albo channel from 1.2; deliver per-channel evidence and readable originals or explicit failures. Close only for the declared usable channels, keep exclusions visible, and leave affected integrations pending when a required channel or attachment remains unresolved.
- [x] 1.13 Complete the criticality/alert source contract for `https://cfr.toscana.it/index.php?IDS=2&IDSS=76`, using the CFR homepage only as a secondary emission/adoption signal; verify formats, necessary graphical resources, access, product-specific reuse, documented daily publication window and 429 behavior. Deliver retained green/non-green samples and compare levels, zones and validity against originals. Close only when these checks pass; service acceptance and trial remain in section 9.
- [x] 1.14 Complete the event-driven monitoring source contract for `https://www.cfr.toscana.it/mobile3/avviso-criticita/`; verify access, product-specific reuse, formats/resources, issuance conditions and 429 behavior. Deliver a real operational sample and verify extraction preserves monitoring semantics without raising missing-publication alarms when no event is expected. A historical normative example alone cannot close this task.
- [x] 1.15 Verify DPC access and reuse separately for internal comparison; deliver retained access/license evidence and a comparison matrix of risk, territory, issuance and validity with non-comparable cases. Verify that DPC permissions are not applied to CFR products and no public DPC warning feed is introduced.
- [x] 1.16 Identify and verify an attributable, reusable postal-code candidate dataset with version/update evidence; deliver permitted-use evidence and ambiguity/coverage checks. Leave this task open while no usable dataset exists; verify API/MCP unavailable-mapping behavior independently and do not block name/ISTAT lookup or verified municipality-zone mappings.

## 2. Application and persistence foundation

- [x] 2.1 Scaffold the modular Go application and Docker Compose services for PostgreSQL, RustFS and internal Crawl4AI, with separate local administration/public listeners and secret configuration; verify startup, connectivity, service health and absence of credentials in logs or exposed configuration.
- [x] 2.2 Implement the authority/channel/product registry and versioned source configuration in PostgreSQL; verify scoped referrals, issuer/publisher separation, per-product access/reuse evidence, declared sections/evaluation periods, pending source states and independently recorded collection/public activation.
- [x] 2.3 Implement RustFS object retention with hashes and PostgreSQL document/version references; verify unchanged acquisition deduplication, stable-URL revisions, complete required resource retention and recoverable partial database/object-write failures.
- [x] 2.4 Implement the PostgreSQL job queue and Go workers with persisted attempts, recovery and idempotent effects; verify restart during processing does not lose work or duplicate document versions/public effects.
- [x] 2.5 Implement model/prompt/configuration version records and per-run/per-attempt accounting; verify available input/output/cache tokens, timing, pricing provenance and unknown usage, separating bootstrap, ordinary collection, OCR and embeddings.

## 3. Acquisition and revision tracking

- [x] 3.1 Integrate internal Crawl4AI for operator-configured source sections with originals preserved; verify full listing traversal, discovered documents and acquisition preview output using Calcinaia cases without autonomous channel discovery.
- [x] 3.2 Implement per-source scheduling, default 10-minute checks, independent 30-minute delay thresholds, content recognition and bounded source backoff; verify HTTP 200 with invalid content, first-error visibility, 429 retry instructions and incomplete checks that do not advance complete-check time.
- [x] 3.3 Implement initial 30-day recovery with explicitly referenced older documents and recurring checks of recent notices, previously irrelevant notices, older ongoing/unresolved measures and their attachments; verify a changed attachment on an unchanged old page is detected without relabeling historical publication as prior service knowledge.
- [x] 3.4 Implement expected-publication schedules only from documented cadence, with configurable tolerance; verify missing expected products, successful unchanged checks and occasional municipal silence produce distinct states.
- [x] 3.5 Integrate vigilance, criticality/alert and monitoring independently through the verified contracts and selected CFR routes from 1.3/1.13/1.14, polling every 10 minutes by default and respecting longer 429 backoff. Compare levels, zones, validity and graphical information against retained samples without treating OCR text as proof of map color, turning monitoring into a new alert level or treating event-driven silence as a missing product.
- [x] 3.6 Prevent a municipal document HTTP 404/410 from starving later targets; persist per-target failure/status and bounded retries, keep checks incomplete during deferral, retain recovery evidence, and verify with unit/integration tests and a deployed Calcinaia check.
- [x] 3.7 Support an evidence-backed disposition for a previously discovered 404/410 URL absent from later configured listings; preserve failure/check history, exclude only the reviewed source/configuration/URL, reactivate on rediscovery, and verify the Calcinaia check after deployment without asserting missing-page equivalence.
- [x] 3.8 Ignore generated Drupal Views DOM identifiers in comments and class names when deciding HTML content identity, retain exact originals, and verify equivalent responses reuse a version while substantive changes still create one in isolated storage tests.
- [x] 3.9 Update integration fixtures after Drupal identifier deduplication; verify marker-only responses share a retained version and separately retained equivalent HTML still exercises contiguous preflight grouping and classification reuse.

## 4. OCR, classification and interpretation

- [x] 4.1 Implement configurable external inference adapters for Regolo.ai while keeping local-provider replacement possible; verify provider contract probes, secret handling and a configurable maximum of three total temporary-error attempts with increasing waits.
- [x] 4.2 Integrate DeepSeek-OCR-2 for scanned PDFs/images and acquisition of necessary linked ordinances; verify original/page evidence, extracted text, unreadable pages and missing attachments against reviewed cases.
- [x] 4.3 Implement Qwen3.8-27B relevance classification on full new/changed content, independently versioning its instructions; verify pertinent notices, unrelated closures, misleading keywords and empty source categories against expected outputs.
- [x] 4.4 Implement Qwen structured extraction from relevant notices and their attachments with schema/evidence validation; verify supported measures, indeterminate fields, original passage/page references and acquired-but-uninterpreted visibility without treating syntactic checks as proof of correctness.
- [x] 4.5 Implement baseline candidate selection using municipality, place and PostgreSQL text search, followed by Qwen update linking; verify partial reopenings, missing explicit links, ambiguous relations and candidates older than 30 days with unresolved measures.
- [x] 4.6 Select an embedding model after resource/provider assessment and implement optional semantic candidate retrieval; compare identical cases with the feature enabled/disabled, verify baseline operation without embeddings and record incremental consumption and linking quality.
- [x] 4.7 Implement dependency-aware interpretation scheduling and explicit selected reprocessing; verify unchanged content does not incur new automatic model calls, modified attachments reprocess affected evidence, ambiguous content is not repeatedly retried, and operator-selected source/period/error scopes preserve earlier runs.
- [x] 4.8 Implement deterministic bounded segmentation for long Qwen inputs and a validated evidence-preserving merge; retain source-version, URL and page/position references per segment, reject duplicate/conflicting/unsupported facts, leave unresolved fields explicit and verify that full-content classification is preserved without relying on a longer timeout.
- [x] 4.9 Implement a deterministic normalized OCR view and extraction-specific operational windows with non-overlapping cores plus bounded context; verify reversible retained-page/offset mapping, operative statements at every boundary, no cross-window evidence and single ownership of overlapping facts without increasing the 90-second timeout.
- [x] 4.10 Implement a versioned compact extraction envelope with one evidence table per window and indexed field references, deriving persisted indeterminate fields locally from explicit nulls; verify malformed, duplicate, out-of-range, cross-window and non-literal references are rejected and retained long-case responses stay within configured input/output budgets.
- [x] 4.11 Implement append-only temporal candidates and conflict-aware extraction merge/projection; verify incompatible supported values retain both original expressions, evidence and one conflict identity while only the resolved field becomes undetermined, including the cross-window Calcinaia `10 settembre`/`10 agosto` case and preservation of unrelated measures.
- [x] 4.12 Correct the current extraction parser against retained acceptance failures: preserve supported measures with redundant references for a null optional place, bound local temporal supplementation to the operative clause, preserve unrelated explicit dates and legacy processing semantics, and validate selected replay before internal rollout.
- [x] 4.13 Reject ordinance-heading-only operative assertions and preserve every explicitly retained subject in coordinated closures after a partial reopening; retain literal core-owned evidence, unknown unsupported times and legacy processing semantics, version the correction, and validate the complete stored municipal replay without paid calls or production rollout.

## 5. Domain behavior and retention

- [x] 5.6 Import and select the retained, reviewed Toscana municipality-zone dataset in the running service through an auditable operator command; preserve unresolved applicability dates and multi-zone/partial territory, verify Calcinaia A4 through the API, and keep source publication gates unchanged.
- [x] 5.7 Connect retained CFR originals to versioned regional domain records with evidence-backed risk/zone/level/validity, distinguish current and future periods, preserve unsupported graphical/monitoring semantics as explicit limitations, and verify worker processing plus selected retained-version replay against real fixtures and API/MCP queries.
- [x] 5.8 Deploy the validated mapping/projection changes, replay an explicit bounded retained scope, and record live acceptance blockers and per-source readiness without fabricating trial, human review or recovery evidence.
- [x] 5.1 Implement seven-risk/three-product representation, regional origin, local measures and operational phases; verify municipal republication preserves regional authorship and local notices without bulletin links do not invent a color.
- [x] 5.2 Implement official versioned many-to-many municipality-zone applicability and municipality/postal candidate lookup; verify attributed act/dataset versions, applicable-period evidence and unresolved dates, preservation of earlier mapping versions, partial-territory summaries, name ambiguity, unsupported territory and unavailable postal mappings while name/ISTAT lookup still works. Follow the separate geography and CAP prerequisites in the dependency table.
- [x] 5.3 Implement measure-scoped updates and temporal precision with original expressions, UTC instants and justified recorded Europe/Rome assumptions; verify the Calcinaia sequence, date-only values, source/platform differences and no invented exact reopening time.
- [x] 5.4 Implement separate provenance, interpretation and updating dimensions and preceding-state warnings; verify source failure does not imply expiry or validity, newer uninterpreted notices remain discoverable, and confirmed interpretation defects mark affected results unreliable.
- [x] 5.5 Implement configurable retention initially three months from each version's first acquisition and dependency-safe RustFS/PostgreSQL cleanup; verify unchanged checks, month boundaries, duration changes on existing data, protected evidence and post-cessation eligibility without promising recovery of deleted data.
- [x] 5.9 Correct regional projection against the reviewed REG-NONGREEN transcription, supporting explicit Italian shared-day and two-day intervals without guessing ambiguous dates. Version the parser, prevent duplicate projections across parser versions at a knowledge boundary, and verify/replay only selected current retained bulletins.

## 6. Public API and MCP

- [x] 6.1 Implement shared queries for the five agreed operation groups and applicable history; verify prominent local provisions, distinct current/future products, document evidence/versions, history bounds and separate regional/local coverage.
- [x] 6.2 Expose anonymous read-only JSON API and remote MCP through the reviewed contract; verify equivalent results at the same data view/evaluation time with a real MCP client and no collection/inference triggered by public queries.
- [x] 6.3 Implement shared per-IP API/MCP budgets with configurable allowance/window/page size and published accounting; verify cross-interface enforcement, retry responses, shared-IP effects and trusted-proxy behavior, then set initial numbers from load measurements.
- [x] 6.4 Implement consistent query views and cursors with configurable lifetime initially 30 minutes; verify updates between pages cannot mix versions, expiry requires restart, storage failure returns unavailability and cached freshness remains truthful.
- [x] 6.5 Implement application-mediated retained-copy access disabled by default and overridden by per-source restrictions; verify exact-version links when enabled, denial when disabled/restricted, internal RustFS isolation and persistent official links/metadata on interpretation failure.
- [x] 6.6 Publish API/MCP usage examples for discovery, situation, search, documents and coverage/history; verify examples against the local service and confirm public interfaces expose no administration, internal DPC records or generated safety judgments.

## 7. Internal administration and source lifecycle

- [x] 7.1 Implement local-only administration with SSH-tunnel access documented for a VM; verify public listeners and MCP cannot reach administration, secret values or configuration mutations.
- [x] 7.2 Implement persisted source/section editing, draft previews, explicit collection activation and configuration history; verify URL/section/method changes keep the prior configuration until a new successful preview while interval changes/suspension apply immediately.
- [x] 7.3 Implement separate acceptance/public enablement and interpretation suspension/resumption controls; verify internal trial data is not published, accepted enabled sources publish notices automatically, collection continues during interpretation suspension, and resumption requires correction/validation without manual measure edits.
- [x] 7.4 Implement operational pages for source status, failed jobs, uninterpreted documents, quality findings, consumption and durations, with selected relaunch/reprocessing; verify totals include retries, expose unavailable usage and separate bootstrap from steady-state costs.

## 8. Notifications, backups and diagnostics

- [x] 8.1 Implement private SMTP/recipient configuration, incident grouping, configurable reminders and recovery email; verify source delay, exhausted attempts and backup failure send the expected messages without a message per transient error.
- [ ] 8.2 Configure and verify off-host PBS whole-VM protection for the dedicated service VM, including PostgreSQL, VM-hosted RustFS and recovery configuration; record schedule/retention, verify failure visibility through the operational notification path and leave the existing application-level backup worker disabled unless separately needed.
- [ ] 8.3 Restore a PBS whole-VM backup into an isolated target; verify service startup, originals, versions, interpretations and configuration links, safe recovery of incomplete jobs, and that old checks are not presented as freshly successful.
- [x] 8.4 Implement internal availability, timeliness and interpretation diagnostics plus scoped DPC comparisons; verify both evidence sides are retained, non-comparability is explicit and DPC does not alter public regional warnings.

## 9. Evaluation and observational trial

- [x] 9.1 After 4.9–4.11, rerun the complete human-reviewed regression for discovery, OCR, full-content classification, operational-window extraction/conflict-aware merge and linking; include long and scanned ordinances, the cross-window date conflict, observed irrelevant notices and regional cases, deliver separate omission, unsupported-assertion and indeterminate-field reports, and keep failed source activation blocked until correction and successful rerun.
- [x] 9.2 Compare baseline and semantic linking, model/prompt versions, latency and per-stage consumption on the same cases; verify a proposed processing change passes evaluation before explicit selected reprocessing and does not automatically replay all history.
- [ ] 9.3 Start and track at least seven complete 24-hour intervals of paced observation across vigilance, criticality/alert, monitoring and Calcinaia; persist daily checks and per-scope evidence of checked sections, delays, attributable original/attachment comparisons (human or explicitly operator-delegated assistant review) and failure behavior. Use retained cases for absent events and extend for omissions, delays, missing evidence, unresolved issues or insufficient observations.
- [x] 9.4 Summarize trial tokens, OCR/embedding charges, retry costs, hosting/storage and bootstrap versus steady-state consumption; verify pricing assumptions and missing metrics are explicit before proposing an operating budget.
- [x] 9.5 Import the reviewed vigilance, criticality/alert, monitoring and Calcinaia source configurations on the dedicated VM; run and retain successful previews before separately enabling internal collection. Verify scheduled acquisition and interpretation outcomes, manual capacity/error checks and public-disabled state for all four sources. Record observed failures and deferred infrastructure honestly; do not claim seven-day trial completion or public acceptance.

## 10. MVP release readiness

- [ ] 10.1 Review separate acceptance reports for vigilance, criticality/alert, monitoring and Calcinaia across API/MCP, seven risks, scanned attachments, history and operational failure states; verify each accepted scope can be enabled independently, partial coverage is never described as complete and full MVP readiness requires all four scopes.
- [ ] 10.2 Verify production on the shared VM behind the existing HTTPS reverse proxy, sole trusted `X-Forwarded-For` hop, private administration through SSH tunnel or configured tailnet-only Tailscale Serve, external model endpoints, existing SMTP path, independent environment resources, concurrent-load capacity thresholds, PBS backup/isolated restore after host failure and rollback procedures; deliver a readiness record before any separately authorized deployment.
- [x] 10.3 Support the operator-requested direct private Tailscale Serve dashboard: validate one explicit HTTPS tailnet origin, preserve local access and cross-site/Host protections, keep loopback backend/public isolation, test and verify the running Serve endpoint, and document access/rollback. This does not complete overall deployment readiness in 10.2.

## 11. Subsequent planned local rollout

- [x] 11.1 Complete Livorno's scoped source dossier and preview configuration; verify official referrals, access/reuse, representative publications and unresolved attribution before its trial.
- [x] 11.2 Complete Pisa's scoped source dossier and preview configuration; verify official referrals, access/reuse, representative publications and unresolved attribution before its trial.
- [x] 11.3 Complete Pontedera's scoped source dossier and preview configuration; verify direct channels, associated-service competence where relevant, access/reuse and representative publications before its trial.
- [x] 11.4 Complete Cascina's scoped source dossier and preview configuration; verify official referrals, access/reuse and representative publications before its trial.
- [x] 11.5 Apply the same evaluation, observation and separate public-activation gates to each additional municipality; deliver individual accepted/pending reports before claiming complete five-municipality local coverage.

## 12. Environment separation and release operations

- [x] 12.1 Document existing development, isolated staging and planned dedicated production topology, including data and secret boundaries, release checks, manual approval, image identity and rollback without claiming operational readiness.
- [x] 12.2 Provide and validate a staging image build/publish helper that records the exact public GHCR digest of a clean revision; document production digest-only deployment with local rebuilding disabled.
- [x] 12.3 Create a separate staging checkout and Compose project on the current VM with distinct PostgreSQL/RustFS volumes, secrets, ports and image tag; validate readiness and isolation without changing the existing development project or enabling unreviewed live collection.
- [ ] 12.4 Publish and test-pull an approved public GHCR image and record its digest; keep the operator's production approval as a separate explicit step.
- [ ] 12.5 Record the production six-hour data-loss and one-hour recovery targets in the PBS runbook, with an initial three-hour backup cadence, retention and age-alert procedure. Complete actual backup, alert and timed isolated restore verification through tasks 8.2, 8.3 and 10.2 before asserting those targets are met.
- [x] 12.6 Support one public checkout for development and staging with separate ignored environment files and secret directories, fixed Compose project names, unchanged named volumes, and checks that prevent accidental cross-environment configuration. Document release handling from a clean shared revision.
- [x] 12.7 Migrate the existing development and staging Compose projects to the shared checkout; preserve data and credentials, verify mounts, readiness and isolation, and retire the private checkout from active deployment use without deleting its unique evidence.
- [x] 12.8 Prepare an inactive `iwa-production` project in the shared checkout with separate ignored settings, secrets and ports; require explicit profiles, a digest-only image placeholder and initial CPU/RAM ceilings, and verify resolved configuration plus absence of production containers without starting services.

## Public environment verification — 1 October 2026

The [environment hardening change](../harden-environment-operations/tasks.md) and [disposable rehearsal summary](../../../docs/operations/environment-verification.md) verify the public checkout’s environment tooling, including task 12.6. Backup/restore and VM reboot checks were handed to the operator at their request; that handoff does not certify readiness. Tasks 1.8, 8.2, 8.3, 10.2, 12.4 and 12.5 remain open until their actual operational results exist. Repository configuration changes do not apply restart policies to existing containers automatically.

## 13. Development acquisition and interpretation recovery — 1 October 2026

- [x] 13.1 Review a stale unavailable municipal target against complete configured listings, record a scoped disposition without erasing evidence, and verify a fresh complete development acquisition check.
- [x] 13.2 Verify the completed incident recovery, preserve its policy and ledger privately, validate representative interpretation work in a bounded canary, and release the obsolete development scope restriction while preserving normal retry, archive and provider controls.
- [x] 13.3 Document repeatable operational recovery, run development runtime/readiness checks, and record validation without changing source acceptance or production activation.
- [x] 13.4 Preserve rejected extraction request/response evidence, fail closed on diagnostic storage failure, verify private append-only capture with synthetic tests, and validate the source-scoped output correction on development.
- [x] 13.5 Resolve embedding/linking recovery scope through their extraction run, preserve document quarantine and call limits, and verify allowed child jobs can finish a selected recovery pipeline.
- [x] 13.6 Bound classification run identities containing reprocessing selections and manifests; preserve existing short keys, isolate distinct selections, and verify selected live reprocessing succeeds.
- [x] 13.7 Recognize literal opening of an explicitly named municipal civil-protection centre as activation in the selected output correction; preserve legacy validation, reject unrelated/negated openings and avoid borrowing alert validity.
- [x] 13.8 Restore provider entitlement for a credential-wide quota hold, resume through a controlled probe, relaunch the reviewed failed OCR job and verify ordinary interpretation progress without a renewed rejection. Requires an active provider account or replacement credential; application fixes alone do not complete this step.

## 14. Optional local Qwen fallback — 1 October 2026

- [x] 14.1 Add an opt-in, source-scoped local chat endpoint and choose complete runner configurations before processing; preserve actual model/configuration provenance, provider scopes, prices and compatible cache boundaries.
- [x] 14.2 Test primary preference, quota/circuit failover, local unavailability, source boundaries, strict output failures, preserved attempt limits and pre-claim deferral against synthetic providers and PostgreSQL.
- [x] 14.3 Add opt-in live local-model classification, extraction, linking and image OCR tests with synthetic Italian notices, long-document cases and literal evidence checks; distinguish transport mocks from real model validation and report unsupported capabilities.
- [x] 14.4 Deploy the evaluated fallback on development within the selected source scope, preserve the remote hold and prior image/settings, and verify representative local runs plus service stability.
- [x] 14.5 Document local model setup, runtime configuration, rollback and the explicit OCR/embedding capability boundaries without claiming acceptance of untested stages.
- [x] 14.6 Expand the explicitly configured development fallback and output-contract source lists to all registered sources, preserving rollback settings, the existing image and independent source/territorial/public controls.
- [x] 14.7 Verify resolved worker configuration, local vision/OCR capability and actual local processing for the expanded scope, preserving remote holds and separately blocked embedding work.
- [x] 14.8 Document register-snapshot selection, future-source updates and operational validation without publishing private settings or claiming source acceptance.

The local fallback was verified with repository checks, isolated PostgreSQL tests, real local Qwen text/image tests and a bounded development canary. Reviewed local quote presentation and civil-protection extraction corrections retain strict contiguous evidence and independent immutable versions. Development acquisition completed a fresh municipal check, OCR and representative text interpretation succeeded, and ordinary processing resumed with the credential-wide remote hold preserved. Semantic embedding work remains separately blocked by remote entitlement; task 13.8, source acceptance and production readiness remain open. Private captures, deployment settings, dumps and call ledgers are excluded from the public repository. See [local fallback operations](../../../docs/operations/local-llm-fallback.md).

The subsequent source-list expansion in tasks 14.6–14.8 was verified on
2 October 2026 with fallback configuration/routing tests, a live synthetic
vision/OCR check and complete development OCR results from the local runner.
Resolved and running worker settings matched; the worker retained its existing
image, other service identities and source controls were preserved, and the
development HTTP/runtime smoke passed. Rollback settings, database archive and
detailed receipts remain private. This operational selection does not restore
remote entitlement or supply embedding support; newly registered sources require
an explicit list update.

## 15. Territorial source knowledge guides — 2 October 2026

- [x] 15.1 Add indexed regional/municipal Markdown guides, a shared template and scoped platform notes, with current guidance, open problems and dated discoveries that distinguish knowledge from intervention status.
- [x] 15.2 Seed Toscana, Calcinaia, Cascina, Livorno, Pisa and Municipium with publicly supportable findings and explicit verification limits; retain environment-specific observations privately and preserve the coverage tracker's dated status assessments.
- [x] 15.3 Link guides from the main README, documentation index and coverage tracker, add contributor upkeep instructions and align proposal/design/specification; verify local links and anchors, ISTAT identities, discovery IDs, checklist counts, private-evidence exclusion and whitespace without activating sources or changing acquisition behavior.

Documentation verification checked local paths/anchors, five territorial identities against the adopted ISTAT register, nine unique discovery IDs and required discovery fields, unchanged coverage identity/status columns and private archive exclusion/permissions. The new requirement and its three scenarios were checked structurally; the OpenSpec CLI was unavailable, so no CLI validation is claimed. These checks do not resolve source acquisition problems or establish acceptance. See the [territorial guide index](../../../docs/territori/README.md).

## 16. Municipal acquisition recheck — 2 October 2026

- [x] 16.1 Inspect development territorial/source gates, retain the baseline privately and run a fresh normal collector check for each active municipal source without enabling inactive sources or changing publication.
- [x] 16.2 Compare fresh check receipts with checked targets, retained version completeness and original objects; retain precise outcomes and missing-resource diagnostics in the private territorial archive, separating discovery from acquisition and acquisition from interpretation.
- [x] 16.3 Update municipal guides and the dated discovery register, add the missing Pontedera guide, scope Municipium findings to the observed Cascina/Livorno pages and verify links, discovery fields, private evidence exclusion and changelog.
- [x] 16.4 Review Livorno's official attachment referral, legal notices, HTML link context and current collector boundaries; probe the exact ordinance URL, preserve its unsuccessful HTTP response privately and document a proposed source-scoped external-resource boundary without claiming activation or successful acquisition.
- [x] 16.5 Re-evaluate the Livorno access probe with an ordinary browser click and the checkout's `DirectHTTP` on the node and development Crawl4AI network; validate the downloaded PDF, preserve evidence privately and revise the earlier access conclusion without claiming collector integration or future availability.

These tasks record a completed diagnostic recheck, including unresolved acquisition failures. They do not certify that every enabled municipality collects successfully, resolve the attachment-selection/access problems, verify semantic interpretation, complete trial/acceptance or change the coverage snapshot. Public findings are in [Cascina](../../../docs/territori/09-toscana/comuni/050008-cascina.md), [Livorno](../../../docs/territori/09-toscana/comuni/049009-livorno.md) and the [territorial index](../../../docs/territori/README.md); detailed operational outcomes remain private.

## 17. Scoped Livorno external attachments — 2 October 2026

- [x] 17.1 Add revision-scoped attachment discovery and reviewed external-resource permissions with exact HTTPS origin/path boundaries, preserving legacy defaults and explicit dependencies.
- [x] 17.2 Enforce resource and redirect bounds, validate configured PDFs, preserve missing-reference failures and apply the same checks to preview and scheduled acquisition before verified retention.
- [x] 17.3 Verify allowed/forbidden resource boundaries, footer exclusion, missing containers, invalid PDFs, preview parity and unchanged legacy behavior with redistributable fixtures and the application PDF parser.
- [x] 17.4 Apply the scoped Livorno revision on development after a successful preview, preserve rollback/evidence and verify a fresh complete worker check plus retained attachments without changing other source configurations or public activation.
- [x] 17.5 Update current territorial/platform guidance, dated discovery history and operational verification references; keep captures/configurations private and record the change in the changelog.

Repository tests, scoped vet checks, the real PDF parser in the application image and an isolated PostgreSQL failure/recovery test passed. A scoped development revision was activated after a successful preview; a fresh ordinary worker receipt and retained PDF reads verified the configured acquisition, with other source states and municipal activation flags preserved. Detailed receipts, settings, hashes, captures and rollback snapshots remain private. These checks do not certify future availability, interpretation, source acceptance, complete municipal coverage or production readiness. See [municipal attachment operations](../../../docs/operations/municipal-attachments.md).

## 18. Cascina dates, attachments and interpretation reuse — 2 October 2026

- [x] 18.1 Verify the one-digit official date, document container and paired retained HTML; preserve private development rollback evidence and confirm exact CSRF-only differences before defining normalization.
- [x] 18.2 Add a Cascina-only versioned HTML policy shared by preflight and interpretation manifests; test CSRF equivalence, source boundaries, operative text/date/link changes and incomplete resources, retaining raw evidence.
- [x] 18.3 Validate one/two-digit listing dates and the page-content attachment boundary with synthetic fixtures; verify persistent date-based planning and zero-call warm interpretation reuse on disposable PostgreSQL.
- [x] 18.4 Review official referral/reuse for Cascina-only API PDFs and apply the development revision after preview, enable existing reuse switches, verify fresh complete collection and corrected tracker dates, and retire only reviewed stale unattempted COC work while preserving originals, provider holds and fallback scope.
- [x] 18.5 Update territorial/platform guidance and dated discoveries, document operational verification/rollback, run relevant checks and record the change without claiming source acceptance or production deployment.

Repository tests, scoped vet and isolated PostgreSQL date/preflight/attachment
checks passed, including Cascina result reuse without a second model call. The
source revision passed preview and two ordinary development checks; stored dates
and parser reads verified the configured acquisition. The second check retained
unchanged versions, while the isolated test verified CSRF equivalence and warm
validated reuse. Only reviewed stale unattempted COC work was archived; originals
remain available. Admin and worker use the correction image; HTTP/runtime
boundary smoke passed. Full release image alignment remains partial because
public and backup retain earlier images. Provider quota hold, fallback scope,
source acceptance and production readiness remain separate. The OpenSpec CLI
was unavailable; no CLI validation is claimed. Detailed evidence and rollback
are private. See [municipal attachment operations](../../../docs/operations/municipal-attachments.md).


## 19. Reviewed local processing before models — 2 October 2026

- [x] 19.1 Add validated source-scoped local-processing policies and separate immutable configurations; include policies in retained references, processing identities and equivalence manifests while preserving legacy defaults and histories.
- [x] 19.2 Select complete configured article containers in document order with missing/empty/ambiguous fallback, preserving attachments and evidence positions; verify misleading and late relevant content.
- [x] 19.3 Add bounded Poppler native-text extraction for explicitly reviewed text-only PDF scopes, with whole-resource OCR fallback for incomplete, scanned, mixed, graphical, encrypted or invalid inputs; preserve page provenance and accurate zero-call accounting.
- [x] 19.4 Add positive-only deterministic relevance for strictly recognized regional formats, with literal evidence and existing model fallback; verify unsupported/incomplete cases, compatible reuse and downstream extraction admission.
- [x] 19.5 Verify repository regressions, disposable PostgreSQL persistence and real Poppler parsing with synthetic fixtures; update operations/territorial guidance, checklist counts and changelog without activating sources or claiming live acceptance.

Verification passed the full Go race suite, vet and command builds, focused
regressions for legacy classification/OCR and contiguous preflight reuse,
disposable PostgreSQL tests for native text and deterministic classification,
strict OpenSpec validation, and real Poppler tests in the application image.
The fixtures verify complete-resource fallbacks, page provenance, zero calls and
prices on local paths, immutable configurations, policy-separated identities,
validated local result reuse and extraction scheduling. Local configuration and
territorial guidance are documented in [local processing operations](../../../docs/operations/local-processing.md).
Source policies remain opt-in and no live settings, provider accounts, acceptance
status or production deployment were changed. Native text requires reviewed
text-only scopes; meaningful vector graphics remain outside those scopes.

## 20. Bounded local-processing replay — 2 October 2026

- [x] 20.1 Replay candidate article and regional policies against a bounded sample of retained development documents using read-only database access; check complete article evidence, fallback and historical relevance agreement without provider calls.
- [x] 20.2 Probe unique retained municipal PDFs with the current native extractor and real Poppler; distinguish technically eligible files from scopes approved for activation.
- [x] 20.3 Keep detailed evidence private and update territorial discovery registers, operations guidance, specification and changelog with the observed results and remaining rollout boundaries.

## 21. Bounded development trial — 2 October 2026

- [x] 21.1 Prepare an eight-version, separate-queue trial using the checked-out implementation, reviewed candidate policies and existing local model gates; preserve a database snapshot and cap calls at twenty.
- [x] 21.2 Persist baseline/candidate classifications and four reviewed native PDF resources in development through real queue claims and processing stores; keep results available for dashboard inspection without automatic downstream publication or broad policy activation.
- [x] 21.3 Verify persisted evidence, usage, model-call limits, dashboard access and worker termination; record results and remaining quality/rollout limits in the territorial registers and operations guidance.

The fifty-version read-only replay passed content/fallback checks, with 51.8%
less municipal HTML input, fifteen local regional decisions, and six of fifteen
unique PDFs accepted for native text (eleven pages). The completed development
trial persisted sixteen successful jobs on eight fixed versions: four municipal
baseline/candidate pairs agreed, their input tokens fell 36.8%, three regional
classifications and four native PDF resources (seven pages) made no model calls.
Sixteen local calls stayed below the twenty-call cap; usage/cost provenance and
browser visibility were verified, and the separate worker exited. Focused race,
vet, strict OpenSpec and disposable PostgreSQL checks passed. The empty-zone
vigilance correction has positive and conservative-fallback regressions. Current
source policies and ordinary routing remain intact; continuing rollout and
semantic downstream acceptance remain separate. Detailed captures, configurations,
receipts, trial programs and snapshot are private.

## 22. Development fallback concurrency trials — 2 October 2026

- [x] 22.1 Verify two local server slots and matching resolved/running worker settings; preserve a verified database archive and private rollback evidence, then add a second development worker without recreating the original or changing its image.
- [x] 22.2 Verify overlapping local requests from distinct queue jobs, returned model provenance, validated results and worker stability while preserving provider holds and source controls.
- [x] 22.3 Document replica scope, trial results and return to one worker; update the specification, checklist counts and changelog without changing the default replica count or claiming a throughput improvement from occupancy.

The development trial verified two-slot occupancy in 59 of 61 samples over two
minutes, overlapping HTTP 200 receipts from distinct worker/job identities and
complete persisted OCR pages from both replicas with the requested local model.
Both workers remained running with zero restarts, and all pre-existing container
identities were preserved. The HTTP/runtime-boundary smoke passed. Full release
image alignment remains unverified because the existing admin image differs from
the environment's configured image. The two replicas remained active for the
requested trial; no new image, schema, source policy or acceptance status was
introduced. Slot occupancy establishes concurrency, not a throughput improvement.
Private settings, database archive, samples and receipts remain ignored. See
[local fallback operations](../../../docs/operations/local-llm-fallback.md).

- [x] 22.4 Activate two additional development worker replicas on operator request; verify four running workers with matching image/settings and preserved existing identities, record the current server slot count and document return to two replicas.

The subsequent operator request brought development to four running replicas
with matching images/settings and preserved first/second worker identities.
The server then reported one slot; this verifies worker activation rather than
four simultaneous model executions. Pre-claim database deadlocks later exhausted
recovery and restarted workers, requiring the fix below. The HTTP/runtime-boundary smoke passed;
private runtime evidence is retained and return to two replicas is documented.

- [x] 22.5 Serialize pre-claim inference queue maintenance across replicas with a transaction advisory lock released before job/model execution; verify four competing callbacks, callback failure and cancelled-waiter release on disposable PostgreSQL with race checks, command checks and vet.
- [x] 22.6 Deploy the corrected image to all four development replicas; verify parallel calls, stored results, provider holds, HTTP smoke and no further pre-claim deadlocks/restarts during the observation window, preserving private logs, database archive and rollback settings.

The corrected image passed focused processing integration/race checks, command
and processing tests, vet and the application image build. All four development
workers were recreated with that image and preserved inference settings. Over
an eighty-second slot observation, four slots were active in 34 of 41 samples;
receipts confirmed HTTP 200 responses with matching model provenance. Twenty-nine
validated classifications were stored across all four workers. Two quotation
validation failures remained rejected by the existing strict parser. No new
pre-claim deadlocks or worker restarts occurred after the fix during observation,
and the HTTP/runtime-boundary smoke passed. Provider holds remained separate.
This trial establishes concurrent processing, not semantic source acceptance or
long-term capacity; detailed logs, receipts, settings and rollback archive remain
private. The repository default remains one replica.

- [x] 22.7 Add two corrected development worker replicas on operator request for a total of six; preserve the existing four identities and settings, verify startup and processing/restart observations against current server slots, and document return to four workers with private evidence retained.

The six-worker expansion preserved all four existing container identities and
matched their corrected image/settings. Each of the six replicas performed queue
attempts during verification, with zero restarts and no observed pre-claim
deadlocks. The local server initially returned HTTP 503 and subsequently exposed
six slots; the first sixty-second slot window observed no active model slots.
This verifies six active queue workers, not six simultaneous model executions.
The nondisruptive HTTP/runtime-boundary smoke passed. Provider retry/circuit
controls remained active; settings, logs and observations remain private.

## 23. Private browser access to API documentation — 2 October 2026

- [x] 23.1 Open Scalar in the integrated browser through a dedicated private Tailscale Serve proxy to the development public listener; preserve existing Serve routes, verify all six endpoint references and the same-origin OpenAPI contract, and document targeted proxy removal with private rollback settings retained.

## 24. Source acceptance prerequisite audit — 2 October 2026

- [x] 24.1 Inspect the active CFR/Calcinaia revisions, declared policies, source regressions and existing observational campaigns; distinguish elapsed observation from reviewed original/attachment/failure/event evidence and preserve the source-scoped prerequisite inventory privately.
- [x] 24.2 Persist truthful assessments of both existing campaigns through the administrative API, prepare the current retained-resource manifest and unresolved municipal comparisons, and verify that the audit does not grant acceptance or public enablement.
- [x] 24.3 Verify registry acceptance/regression gates, campaign evidence requirements and shared public query groups with synthetic unit tests and disposable PostgreSQL; update territorial guidance, the coverage precheck note and changelog without claiming real-source acceptance.

Both recorded campaign assessments remain `extended`. The recent campaign has
sufficient elapsed observation and checks within its configured delay threshold,
but reviewed originals/resources and error/event evidence are missing. Calcinaia
also retains a failed service regression requiring a corrected rerun against the
reviewed contract; successful software tests do not close that regression. The
private dossier identifies these prerequisites and retained-resource identities.
Tasks 9.3 and 10.1 remain open; no source was accepted or publicly enabled.
Infrastructure/public readiness gates remain separate and unverified by this audit.

## 25. Manual municipality publication in development — 2 October 2026

- [x] 25.1 Add default-off municipality publication state, immutable actor/time events and revision-checked private admin/CLI controls available only in explicit development; pin all Compose runtime environments and preserve collection/source acceptance flags.
- [x] 25.2 Share development visibility across API/MCP municipal data, applicable CFR facts, document metadata and history; constrain selected municipalities/zones and active policy-permitted acquisitions, disclose development limitations and preserve independent copy restrictions.
- [x] 25.3 Verify strict staging/production behavior, same-zone isolation, historical mapping applicability, revocation and saved-view expiry, native-form protections, audit persistence and default/invalid environment handling with synthetic tests and disposable PostgreSQL; update documentation and coverage boundaries.
- [x] 25.4 Deploy the reversible development public/admin change, enable Calcinaia explicitly and verify selected/unselected API behavior and administrative state with private rollback evidence; report uncertain regional levels and missing municipal facts without claiming source acceptance.

Development public/admin were rebuilt, migrated and deployed; the browser's native
form explicitly enabled Calcinaia. API/MCP agree on its saved view and the original
HTTPS endpoint returns applicable CFR facts with development limitations. A second
municipality sharing A4 remains unselected, and unscoped regional facts stay in A4.
Source collection/public flags and acceptance counts are unchanged; existing workers
and other services were preserved. Unknown regional levels and absent interpreted
municipal measures remain explicit. The nondisruptive boundary/secret/HTTP smoke passed.
This targeted deployment does not certify uniform application-image rollout: the full
runtime image checker reports preserved older backup/worker images. Private database,
listener settings, response comparisons and rollback details are retained. Source
acceptance tasks 9.3 and 10.1 and production release gates remain open.

## 26. Cittadino Informato acquisition with multi-source verification — 3 October 2026

- [x] 26.1 Formalize the requested additional institutional acquisition/discovery channel for selected municipalities in proposal, design and the three specifications; preserve independent municipal/CFR collection, scoped referrals, candidate limits and separate permissions/acceptance/activation. Record platform guidance and dated findings without claiming runtime adoption. The subsequent 3 October revision replaces the original reviewed-grounds gate with technical multi-source verification; this historical planning task does not complete the revised runtime work.
- [x] 26.2 Deliver and verify a working multi-source acquisition-to-verification system for Regione Toscana/CFR, the selected municipality and cittadinoinformato.it. After 26.3 and 26.4, exercise a bounded development run with regional and local cases and inspect persisted receipts identifying candidate, channel roles, source/version evidence, passages, check times, compared fields and outcomes. Verify missing/unavailable/non-comparable/not-applicable/conflicting evidence behavior and diagnostic-only admission without required primary support. Completion requires working tested comparisons, not obtaining licenses, agreements or legal-basis documentation; retain bounded public access, independent primary collection and link-only platform copies.
- [x] 26.3 Implement separate source identities and opt-in bounded acquisition for explicitly selected recognized municipality sections, with versioned originals, date meanings, dependencies, retries and visible failures; do not inherit recognition or permission across municipalities. Verified with synthetic unit/disposable PostgreSQL cases and two bounded isolated Calcinaia HTTP checks on 3 October 2026; pending verification candidates, no live source activation. See [operations](../../../docs/operations/cittadino-informato.md).
- [x] 26.4 Implement persistent, evidence-bound multi-source verification receipts and candidate-to-domain admission: municipal publications/referenced acts for local fields, comparable originating CFR products for regional republications; record all three channel roles and distinguish corroboration, missing primary evidence, unavailable checks, non-comparable evidence, not-applicable checks and actual conflicts; preserve changes/history, avoid duplicate measures and never use majority vote between republications.
- [x] 26.5 Verify synthetic and reviewed real-source cases covering primary-only/platform-only notices, date/version mismatches, unavailable checks, actual conflicts, local measures without regional alerts, unchanged/revised content, same-platform unselected municipalities and equivalent API/MCP attribution; retain link-only copies and explicit coverage limits.
- [x] 26.6 Run a bounded operator-selected development preview/trial for Calcinaia before separately activating scheduled acquisition; verify independent municipal/CFR checks and comparison receipts, scoped visibility and rollback, without completing source acceptance or enabling other municipalities by inference.

**26.2 preliminary progress — 3 October 2026 (superseded by the completed run below):** bounded manual HTTP review identified the
Calcinaia-linked REST index, declared updates/risk routes and successful
municipality/project responses. WordPress page metadata has old container dates
and an empty body, distinct from the current HTML bulletin; it cannot supply
notice issuance or operative validity. ANCI Toscana is identified as the visitor
data controller, not thereby as holder of every content/database right. The
[review dossier](../../../docs/fonti/piattaforme/cittadino-informato-review.md)
records the original review and revised technical closure criteria. The subsequent
user decision supersedes the licensing/documentation completion gate. Task 26.2
remains unchecked: the later 26.3/26.4 work implements acquisition and the
programmatic verifier, but the bounded end-to-end development run is still missing;
there was no collector activation or external communication in that review.
The completed run below supersedes only the outstanding 26.2 execution.

**26.3 implementation — 3 October 2026:** the typed opt-in API adapter, registry
and storage guards are implemented. Tests verify scoped acquisition, byte-exact
originals, repeat/revised versions, date meanings, invalid/missing dependencies,
rate limits and unavailable-detail deferral/recovery. Two bounded isolated HTTP
checks verify Calcinaia access and unchanged-version reuse. This later work completes
26.3 only; persistent multi-source comparisons/admission (26.4) and the complete
26.2 development run remain open; the later 26.4 work below adds the verifier.
Private source captures are excluded from Git.

**Revised sequence:** implement 26.3 acquisition and 26.4 persistent comparisons,
then close 26.2 from the bounded development run and inspected receipts. Complete
26.5's broader synthetic/real-source evaluation before 26.6's bounded development
preview/trial; scheduled adoption remains a separate activation. A local-only measure may mark CFR not applicable; a regional
claim uses the originating CFR product without requiring municipal republication.
No three-channel quorum or majority vote replaces field-level primary evidence.

This records the earlier authorized planning revision. Its document changes
replaced task 26.2's completion criterion without implementing acquisition or
projection. The subsequent 26.3/26.4 work adds that code; source acceptance and
operational activation remain separate.
The historical checked bootstrap comparison in task 1.11 remains dated evidence,
not proof of this new recurring workflow. Document consistency, scenario structure
and local links are verified separately from runtime behavior and source acceptance.

**26.4 implementation — 3 October 2026:** migration 068, the programmatic
verifier and private admin CLI persist immutable evidence selectors, passages,
source/version/configuration identities, comparison directions, all channel
roles, check times and per-field outcomes. Atomic primary-only domain projection
preserves CFR colors and supports local-only/primary-only facts without majority
vote; direct platform store writes cannot bypass verification. Idempotence,
concurrent receipt/measure reuse, revisions and returns to earlier content,
existing-domain reuse with temporal/interpretation assessment, completed PDF OCR
ownership, missing/unavailable/non-comparable/conflict cases and dependency-safe
retention are verified with synthetic disposable PostgreSQL. Query tests verify
current and past knowledge selection without duplicate measures. See
[operations](../../../docs/operations/cittadino-informato.md#verifica-persistente-e-ammissione--task-264).
This completes 26.4; the verifier consumes interpreted selectors and prior check
results, not an automatic counterpart-discovery workflow. No new real-source
run, live migration/activation, generic extraction-worker wiring or API/MCP
verification-output extension is claimed by 26.4. The subsequent 26.2 run below
closes that execution; tasks 26.5 and 26.6 remain open.


**26.2 development verification — 3 October 2026:** independent municipal and
CFR acquisition, scoped platform API acquisition and evidence-selector verification
ran on the development PostgreSQL/RustFS services with dedicated private data. Ten
persisted receipts cover three real-source cases and seven labelled synthetic
controls: supported independent local fields, non-comparable local/regional
republications, missing/unavailable primary checks, not-applicable roles, matching
local/regional facts, edition mismatch and comparable conflicts without majority
voting. Receipt readback, literal selectors/hashes/configurations/check times,
request idempotence, immutable-input conflicts, histories, primary fact reuse and
diagnostic-only unsupported candidates passed. The wider platform window's
external municipal PDF failure remains recorded; no complete traversal of that
window or widened platform permissions is claimed. Two successful passes use the
display window since 1 October plus today/tomorrow risks; the exact local notice
was checked separately with the independently retained municipal publication and
referenced PDF. Real regional map colors remain unverified. Repository unit,
disposable integration/query tests and nondisruptive development smoke passed.
Existing containers/source controls were preserved; trial sources remain disabled
and copies link-only. Private evidence: `CIN-DEV-20261003-01`. See
[procedure and limits](../../../docs/operations/cittadino-informato.md#prova-completa-in-development--task-262).
This completes 26.2 only: broader reviewed evaluation/API-MCP attribution (26.5),
scheduled adoption (26.6), generic worker wiring and source acceptance remain open.

**Review delegation — 3 October 2026:** the operator asks the assistant to
complete the remaining review by analyzing the PDFs. Historical human
confirmations remain scoped to their original cases. New comparisons identify
the assistant and delegation; no new human approval is inferred. This changes
the review method, not the seven-day duration, failed-regression, per-source
acceptance or publication criteria.


**26.5/26.6 verification — 3 October 2026:** the operator delegated document
review to the assistant. Original PDFs and municipal notices were compared,
publication dates removed from the local edition selectors and the platform's
literal validity selected. New immutable receipts preserve the earlier ones;
local operative agreement does not override unknown identity or a missing
platform attachment. Real and synthetic comparisons retain primary-only facts,
non-comparable regional republications, conflicts, unavailable checks, revisions,
unselected municipalities and private-channel redaction. Public document/fact
responses now carry bounded evidence-aware receipts; API/MCP schemas and real
MCP client tests agree. Synthetic disposable PostgreSQL tests cover knowledge
boundaries and primary deduplication. This completes 26.5.

The operator-selected 26.6 trial uses a separate development database, independent
bounded municipal/CFR/platform previews and explicitly labelled re-evaluation of
retained real cases. API and a real MCP client return equivalent attribution;
only Calcinaia has temporary development visibility. Revocation hides the document
again, collection controls return disabled and audit history remains. No scheduled
collector is connected to the trial database. Source acceptance and other
municipalities remain unenabled. Private evidence: `CIN-TRIAL-20261003-01`.
This completes the bounded trial required by 26.6; it does not activate scheduling.

**Source-review outcome — 3 October 2026:** the original thirteen municipal
regression expectations are preserved and pass their current comparisons,
including listing discovery and hash-matched retained OCR. A separate stronger
completeness suite records the current partial-reopening extraction/merge failure;
that failure is not hidden by the thirteen passes. The live generic municipal
projection and regional graphical interpretation limitations remain unresolved.
Delegated original/attachment reviews identify their actor and scope, with failed
or unresolved outcomes where needed. Campaign assessments and independent source
acceptance remain separate; tasks 9.3 and 10.1 are not completed by these reviews.
Unchanged originals can be reviewed under a later campaign configuration only
when a finalized matching acquisition already proves that reuse; regression tests
reject unproven and future reuse.


## 27. Repair the live Calcinaia situation — 3 October 2026

- [x] 27.1 Reproduce the development API response and distinguish real interpretation limits from missing extraction-to-domain integration, missing provenance history and stale deployed images; preserve the response and source controls privately.
- [x] 27.2 Wire validated primary municipal extraction and supported update links to atomic, idempotent domain projection; preserve field evidence, temporal ambiguity and knowledge boundaries, exclude evaluation/platform-only/incomplete/suspended inputs, and verify shared API/MCP schemas, source quality and first notices requiring attention with synthetic PostgreSQL tests; preserve historical fallback catalogs when registering local worker configurations.
- [x] 27.3 Build and deploy the reversible development correction, replay a bounded set of retained ordinary primary results without new model calls, carry registered provenance into quality history, and verify Calcinaia on the actual HTTPS API/MCP with source controls preserved and private rollback evidence.
- [x] 27.4 Complete and validate product-specific CFR graphical interpretation against the retained maps/PDFs, including applicability and all seven risks, before replacing unresolved per-zone levels or claiming full source acceptance.

- [x] 27.5 Resolve the current Calcinaia criticality situation from retained matching PDF vector evidence for all seven risks and both days; preserve unknown unsupported zones, PDF/page attribution, precise explicit table intervals and historical projections, test synthetic non-green/missing/conflicting cases, and verify actual HTTPS API/MCP after development deployment. This bounded completion does not close the full product/zone interpretation and source acceptance in 27.4.

**27.4 verification — 4 October 2026:** `cfr-graphics-v4` interprets the recognized
retained criticality/vigilance products for 26 zones, seven risks and both days.
Two criticality and two vigilance PDF/map comparisons verify 364 entries each,
including excluded river/coastal masks and the historical non-green levels.
The historical PDF has leading adoption pages and a separately identified
reviewed HTML table transcription. Vigilance retains daily/cumulative rainfall,
area-average units, symbol presence and literal HTML evaluation membership without
alert colors or inferred absence of risk. Synthetic fixtures verify positive
symbols for wind/sea/snow/ice; additional real positive cases remain source
acceptance evidence. Raster/ambiguous graphics, changed legends and mismatched
editions/dates preserve uncertainty or supported HTML fallback.

Local vector font normalization fixes the retained historical SVG failure,
with original physical page attribution. Disposable PostgreSQL verifies
append-only history, idempotence, missing resources and supported coverage;
real MCP client tests verify weather-bearing API/MCP pagination equivalence.
Race tests, vet/build, vulnerability scan and strict specification validation
pass. Private hashes, originals and expected/result comparisons are retained as
`CFR-GRAPHICS-20261004-01`. See [procedure](../../../docs/operations/cfr-graphics.md).
This closes the graphical software task; existing live services/source controls,
municipal completeness and source acceptance are unchanged.

## 28. Concise municipality situation — 3 October 2026

- [x] 28.1 Replace the raw API/MCP situation presentation with processed conclusions and separate regional, municipal and Cittadino Informato summaries; preserve validity, applicability, evidence, missing-channel semantics, source identity, source acceptance and detailed operations.
- [x] 28.2 Align shared schemas, documented consumer paths and snapshot pagination; verify rich API/MCP equivalence, uncertainty, missing/platform-only channels, pending backlog, full-scope conclusions, deduplicated evidence and truthful saved-view freshness with synthetic and PostgreSQL checks.
- [x] 28.3 Build and deploy the reversible development presentation update and verify the actual Calcinaia HTTPS API/MCP response and controls, retaining private rollback evidence.

**Task 28 verification — 3 October 2026:** synthetic rich situation tests verify
source attribution, unavailable/platform channels, deduplicated evidence, literal
uncertainty, unknown/non-applicable/elevated levels, pending notices and full-scope
conclusions across pages. Real MCP client, disposable PostgreSQL, race and vet
checks pass. The development public service alone runs the new presentation;
actual HTTPS API/MCP on the same dataset_version agree, the JSON Schema validates,
and compact JSON decreases from approximately 186 KB to 14 KB in the observed
response without losing the fourteen regional facts or seven municipal measures.
Historic unknown levels, unselected municipality isolation and unchanged source
publication/acceptance states are verified. Other persistent development services
are preserved; prior image/environment remain available for rollback. Private
record: `CLN-SUMMARY-20261003-01`. This presentation change does not complete
source acceptance, municipal dates/completeness or continuous platform collection.


## Remaining-task verification — 4 October 2026

Task 9.4 delivers separate cost summaries of both retained campaigns, including
explicit unavailable hosting/storage inputs, unknown usage/cost/retry metrics
and historical pricing uncertainty. Current Regolo rates were checked without
retroactive price changes. Both reports remain budget-ineligible; task 9.3 and
source acceptance are not completed by accounting. See [trial costs](../../../docs/operations/trial-costs.md).

Task 11.5 delivers four individual pending reports and implements explicit
single-municipality observational campaigns with the pilot's duration, original,
attachment and failure gates. Synthetic PostgreSQL tests verify completion,
premature/insufficient evidence, invalid scopes and preserved legacy MVP
campaigns. Actual opening attempts remain rejected for unresolved prerequisites;
source controls are preserved. This completes independent gate/report work,
without declaring evaluation/trial success, acceptance or five-municipality
coverage. See [municipal acceptance](../../../docs/operations/municipal-acceptance.md).

The [PBS runbook](../../../docs/operations/pbs-recovery.md) now specifies RPO/RTO,
three-hour cadence, proposed retention, age escalation and timed isolated
recovery. Tasks 1.8, 8.2, 8.3, 10.2 and 12.5 remain unchecked until actual
infrastructure, backup, alert and restore outcomes are verified. GHCR publication
in 12.4 still requires an approved staged image and working package access.


Task 13.8 is verified after an authenticated Qwen probe and controlled credential
resume. The reviewed historical OCR job had already succeeded through its recorded
local-fallback relaunch; replaying that command is idempotent. A bounded selected
canary completes its previously failed remote OCR run with persisted pages and
receipts. A separate ordinary municipal case completes classification, extraction
and linking without renewed provider rejection. An invalid-quotation attempt is
preserved as a distinct quality failure. Global/Qwen/OCR gates close; embeddings
remain disabled and untested here. See [provider recovery](../../../docs/operations/acquisition-recovery.md#verifica-del-recupero-provider--4-ottobre-2026).

The checklist now has 153/161 completed tasks and eight open tasks, summarized in
[readiness](../../../docs/operations/toscana-readiness.md). Source acceptance,
complete observational evidence and infrastructure readiness remain independent.


The clean application revision was subsequently adopted in development across
public/admin, all six existing worker replicas and the already active application
backup service. Other containers/volumes and source controls were preserved;
whole-VM PBS remains unverified. Runtime image/configuration checks, HTTP readiness,
private municipal campaign rejection/legacy MVP readback and actual SDK API/MCP
saved-view equivalence pass. See [development adoption](../../../docs/operations/toscana-readiness.md#adozione-in-development).
This adoption also makes the retained CFR graphical implementation from 27.4
available to the ordinary development path without granting source acceptance.
