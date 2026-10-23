# Changelog

This file records repository plans, fixes, implementation, documentation, and operations changes. Entries describe code and plans; they do not certify source acceptance or deployment. New entries go under Unreleased, newest first.

## Unreleased

- 2026-10-23 [Changed] Record the operator's acceptance of cfr-monitoring revision 3 with declared limitations in development; release readiness stays pending until the acceptance review is recorded.
- 2026-10-23 [Added] Plan bound-ongoing-criticality-rows: bounded validity for CFR criticality rows already in progress at issuance, explicit derived green outside listed alert intervals, and evaluation-time freshness for historical public views.
- 2026-10-23 [Changed] Record the task 31.5 reassessment of CFR criticality and monitoring on real cases (TOS-023): criticality tables and intervals match, with one defect for rows reading 'in corso'; monitoring event bulletins are projected without inferred colours; no acceptance recorded.
- 2026-10-23 [Changed] Calcinaia configuration revision 2 reads only the reviewed article body; the same fourteen-check contract now passes in development, and the ordinary reprocessing of the two failing notices matches (CLN-044).
- 2026-10-23 [Fixed] Bound extraction run identities that include a local-processing configuration and input manifest to the processing store's 200-byte limit, as classification already does; existing identities are unchanged.
- 2026-10-23 [Changed] Record the task 31.5 reassessment: vigilance matches the 20–21 October wind and sea tables but lacks snow and ice cases; the Calcinaia fourteen-check contract fails on the v29 local configuration with two extraction omissions; no acceptance recorded.
- 2026-10-23 [Changed] Archive the completed acquire-referenced-albo-acts change and publish the referenced-act-evidence specification.
- 2026-10-23 [Changed] Adopt referenced-act retrieval in development: the Calcinaia 9 September notice resolves to Ordinanza 38/2026 and its measures become conditional from 10 September, the 20 August notice is clarified by its attached ordinance and the 21 August monitoring stays ambiguous; complete task 31.4 and document the procedure.
- 2026-10-23 [Fixed] Identify the service with an explicit User-Agent on albo requests; the reviewed JCityGov portal rejects the default Go client with HTTP 403.
- 2026-10-23 [Added] Expose referenced-act evidence on clarified public measures, add a read-only operator list (admin source page and source-act-references CLI) and a first-page albo preview; deny scheduled checks of albo sources.
- 2026-10-23 [Added] Add referenced-act jobs: resolve notice references after projection, retain resolved albo acts once with OCR only, and clarify unknown measure validity from the act's dispositive clause; exclude albo sources from classification, listing checks and the public perimeter.
- 2026-10-23 [Added] Add deterministic detection of explicit act references in notice text and their resolution as attached, numbered or unique albo candidates, recording not-found, ambiguous and unavailable outcomes append-only.
- 2026-10-23 [Added] Add append-only referenced-act tables with retention participation, a reviewed JCityGov albo source configuration and a bounded, session-aware JCityGov albo client with synthetic fixtures.
- 2026-10-23 [Planned] Record the task 31.4 inventory of acts cited by Calcinaia notices (Ordinanza 38/2026 resolves the 9 September date conflict), the audited recovery of the archived 20 August notice, and plan acquire-referenced-albo-acts for bounded albo acquisition and measure-date clarification.
- 2026-10-23 [Changed] Archive the completed replace-equivalent-version-measures change and add the current-measure requirement to the equivalent-interpretation-carry-forward specification.
- 2026-10-23 [Changed] Adopt current-measure replacement across equivalent versions in development: Calcinaia shows only the latest equivalent projection of two notices (29 to 24 current measures), with API/MCP agreement and unchanged history; document the selection rule and the conservative tracked-target union.
- 2026-10-23 [Changed] Compute equivalent-version measure replacement once per public query instead of per measure, avoiding a fan-out across every equivalent template version (659 ms to 70 ms on development data) with identical selections.
- 2026-10-23 [Changed] Replace current public measures of an earlier document version once a later version with an equivalent interpretation input has a validated projection; genuine revisions keep accumulating and earlier knowledge boundaries are unchanged.
- 2026-10-22 [Planned] Plan replacing current public measures of earlier versions with the latest validated projection of an equivalent interpretation input of the same document.
- 2026-10-22 [Changed] Record the decision not to reprocess or archive the failed Livorno orange-alert versions and the diagnosis of duplicate measures across versions of the same document.
- 2026-10-22 [Changed] Record the saved-view instant matching fix and its development verification in the task 31.2 notes.
- 2026-10-22 [Fixed] Accept any RFC 3339 spelling of the same evaluation_time and known_at when continuing a saved public view through HTTP pagination or MCP, instead of rejecting it with cursor_mismatch.
- 2026-10-22 [Changed] Record that the PostgreSQL deadlocks in the development log all predate the 2 October queue-maintenance serialization fix.
- 2026-10-22 [Changed] Complete task 31.2: bounded reprocessing of seven Calcinaia versions, API/MCP and unchanged-history verification, and record the recovered Livorno orange-alert measures.
- 2026-10-22 [Fixed] Persist fully determined extracted measures with an empty indeterminate-field array instead of JSON null, which the database rejected and caused the whole document extraction to fail.
- 2026-10-22 [Changed] Archive the completed carry-forward-equivalent-interpretations change and publish the equivalent-interpretation-carry-forward specification.
- 2026-10-22 [Changed] Adopt carry-forward of equivalent interpretations and extraction v29 in development; document equivalence, backfill, public limitations and the verified Calcinaia results.
- 2026-10-22 [Added] Carry forward the latest earlier validated interpretation of an equivalent input when a document version's own interpretation failed or is pending, with configuration-independent manifest fingerprints, a backfill command and explicit provenance limitations.
- 2026-10-22 [Planned] Plan carry-forward of validated interpretations across template-only versions with equivalent input, and correct the scope of the Calcinaia document-state regression.
- 2026-10-22 [Fixed] Recognise positive finite forms of 'vietare' as prohibition evidence (extraction configuration v29) and report the actual cause of invalid extraction outputs; record the Calcinaia 31.2 diagnosis and the open template-version regression.
- 2026-10-22 [Changed] Document the current development publication selection: Calcinaia and Cascina.
- 2026-10-21 [Changed] Archive the completed maintain-target-measure-state change, publish tracked-target-interpretation-state and update the unavailable-municipal-targets classification requirement.
- 2026-10-21 [Changed] Maintain tracked target measure, relevance and reference state from validated interpretation, recheck older notices with open measures within a configurable limit, and verify the development adoption.
- 2026-10-21 [Planned] Plan maintaining tracked target measure, relevance and reference state from validated interpretation, with a configurable limit for open-ended measures.
- 2026-10-21 [Changed] Archive the completed handle-withdrawn-unavailable-targets change and publish unavailable-municipal-targets as a main capability specification.
- 2026-10-21 [Changed] Explain missing Cascina weather alerts, record its development publication and the failed wider-scope configuration drafts.
- 2026-10-21 [Changed] Verify the development adoption of withdrawn unavailable targets: every active source completes its first check and Livorno recovers from a two-week delay.
- 2026-10-21 [Changed] Stop renamed or withdrawn 404/410 municipal notices from failing source checks, count them separately, and add guarded administration and CLI dispositions for unavailable targets.
- 2026-10-21 [Planned] Plan non-blocking handling of withdrawn 404/410 municipal targets with guarded operator dispositions through administration and CLI.
- 2026-10-21 [Changed] Record the reviewed Calcinaia exclusion of three renamed 404 notices that kept collection failed since 14 October, with the verified complete check and the open renamed-URL limit.
- 2026-10-21 [Security] Upgrade Go to 1.27.3, golang.org/x/net to v0.61.0 and klauspost/compress to v1.18.7 to resolve reachable standard-library and HTTP/2 vulnerabilities reported by govulncheck.
- 2026-10-21 [Changed] Archive six completed OpenSpec changes and publish their requirements as the first main capability specifications.
- 2026-10-11 [Changed] Support exclusive local primary chat and OCR providers with independent credentials, configurable deadlines and retained catalog coexistence; verify development adoption.
- 2026-10-11 [Changed] Suppress unchanged processing failure reminders, consolidate daily backlog notifications and verify targeted development adoption with audited redundant-job retirement.
- 2026-10-09 [Changed] Clean unused development images, old tool releases and regenerable host caches while preserving application data, backups and existing services.

- 2026-10-09 [Changed] Record a complete local OCR page beyond the former timeout in the bounded Livorno capacity trial
- 2026-10-09 [Changed] Enable reviewed Livorno reuse on one development worker and extend the bounded local timeout trial with retained rollback settings and histories.
- 2026-10-09 [Fixed] Add reviewed Livorno CSRF equivalence and a development worker image selection for scoped OCR reuse rollout.

- 2026-10-09 [Changed] Diagnose recurring Livorno OCR work from CSRF-only versions; distinguish shared provider pauses, new timeouts and historical job failures.

- 2026-10-08 [Changed] Reduce development inference concurrency after a bounded timeout trial; record preserved rollback evidence, verified OCR progress and pending model-host capacity checks.

- 2026-10-08 [Changed] Document local model timeout diagnostics and bounded capacity checks for prompt processing, generation and GPU memory placement.

- 2026-10-08 [Changed] Clarify Livorno failure diagnosis, separating unavailable tracked notices, crawler structural rejection and interpretation retries.

## 0.4.0 - 2026-10-08

- 2026-10-08 [Changed] Prepare 0.4.0 with aligned API/MCP metadata, preserved release history, performance and processing upgrade notes, and the fixed text dependency
- 2026-10-08 [Fixed] Allow isolated development public listener image updates and document the verified latency repair rollout
- 2026-10-08 [Fixed] Reduce public API/MCP latency with indexed batch reads and per-query reuse
- 2026-10-07 [Changed] Preserve parameterized primary comparison evidence and document the source audit, completed repair rollout and API/MCP pagination and historical verification
- 2026-10-07 [Changed] Recover decoded JSON literal classification and oversized PDF OCR with immutable catalogs; correct COC activation validity and verify selected ordinary reprocessing
- 2026-10-07 [Changed] Plan latest-source repairs and correct CFR intraday projections, publication accounting, vigilance diagnostics and bounded API responses
- 2026-10-04 [Changed] Adopt scoped Calcinaia Cittadino Informato collection and automatic comparisons in development; record cadence, diagnostic receipts and interpretation limits
- 2026-10-04 [Fixed] Preserve archived primary evidence in automatic comparisons and make historical-query test boundaries deterministic
- 2026-10-04 [Added] Add ordinary Cittadino Informato comparison workers and reviewed external PDF dependency scopes
- 2026-10-04 [Changed] Update Cittadino Informato findings to distinguish isolated trials from ordinary scheduled collection
- 2026-10-04 [Changed] Document the verified Calcinaia coverage matrix, pending interpretations and scoped operational follow-up
- 2026-10-04 [Changed] Complete Toscana MVP acceptance-report review and verify independent four-scope gates, API/MCP failure parity and explicit pending source limitations.
- 2026-10-04 [Changed] Complete Toscana observation task 9.3 with a persisted four-product campaign assessment, audited corrections, unchanged regression contract and preserved historical failures
- 2026-10-04 [Fixed] Exclude retrospective closure clauses from completed municipal reopenings while preserving independently evidenced restrictions and prior failed regressions
- 2026-10-04 [Changed] Preserve independent municipal prohibition scopes and historical projected interpretations; add immutable corrective observation reviews and bounded CFR header handling
- 2026-10-04 [Changed] Verify selected archived municipal recovery in development, preserve historical results, and record the live completeness omission and extended Toscana campaigns.
- 2026-10-04 [Added] Add audited exact-version recovery of archived interpretations, preserve historical guards and verify the unchanged municipal regression
- 2026-10-04 [Fixed] Preserve partial municipal reopenings and original place spelling, with retained replay and persistence verification
- 2026-10-04 [Changed] Align active development applications to the verified revision and validate runtime images, private municipal gates and API/MCP equivalence
- 2026-10-04 [Changed] Add independent municipal observation gates and pending reports, summarize trial costs, verify provider recovery and prepare the PBS runbook
- 2026-10-04 [Fixed] Complete retained CFR criticality and vigilance graphics with product-specific weather metadata, applicability, PDF evidence and preserved history

## 0.3.0 - 2026-10-03

- 2026-10-03 [Changed] Prepare the 0.3.0 release with aligned API/MCP metadata, situation consumer migration notes, source verification upgrades and exact-revision promotion checks.
- 2026-10-03 [Changed] Present municipality situations as processed conclusions and separate source summaries, preserving evidence, uncertainty and API/MCP pagination.
- 2026-10-03 [Fixed] Determine scoped criticality levels from retained PDF vector maps with page evidence, precise alert validity and historical preservation.
- 2026-10-03 [Fixed] Connect primary municipal extractions to public facts, preserve provenance history and worker catalogs, and verify development API/MCP behavior.
- 2026-10-03 [Changed] Complete tasks 26.5 and 26.6, record delegated document reviews and development rollback, and document remaining source-acceptance failures.
- 2026-10-03 [Fixed] Preserve operative municipal directives and completed road updates, and bind attributable campaign reviews to proven reuse of retained evidence.
- 2026-10-03 [Added] Expose scoped multi-source verification receipts through equivalent API/MCP responses with evidence visibility and knowledge-boundary checks.
- 2026-10-03 [Changed] Complete task 26.2 with bounded development acquisition-to-verification checks, persisted real-source diagnostics and labelled synthetic controls; document scope and dependency limits.
- 2026-10-03 [Added] Implement evidence-bound multi-source verification receipts and primary-only domain admission with immutable history, deduplication and protected comparison evidence; verify synthetic PostgreSQL/query cases for task 26.4, leaving the bounded development run and activation separate.
- 2026-10-03 [Added] Implement municipality-scoped Cittadino Informato API acquisition with versioned originals, bounded discovery, date separation and dependency/retry guards; verify synthetic persistence and bounded isolated Calcinaia checks, leaving multi-source admission pending.
- 2026-10-03 [Planned] Replace the Cittadino Informato licensing prerequisite with tested persistent multi-source verification across Regione Toscana/CFR, municipality and platform
- 2026-10-03 [Changed] Document the preliminary Calcinaia Cittadino Informato access and reuse review, exposed API scope and unresolved acquisition prerequisites
- 2026-10-03 [Changed] Plan Cittadino Informato as a scoped institutional acquisition/discovery channel with municipal/CFR corroboration, independent primary checks and separate access/reuse review; leave implementation and activation pending.

- 2026-10-03 [Changed] Record institutional referrals for Calcinaia's Cittadino Informato channel, distinguishing recognized provenance from technical reliability, collection activation and reuse conditions.

- 2026-10-03 [Changed] Document the missing municipal extraction-to-domain connection, situation attention and quality limits, unverified CFR map levels and the distinction between the planned secondary sentinel and active collection.

## 0.2.0 - 2026-10-03

- 2026-10-03 [Fixed] Reserve callback connection capacity explicitly in the concurrent queue-maintenance test fixture, preventing CPU-dependent Coverage deadlocks while preserving serialization and cancellation checks.
- 2026-10-03 [Changed] Prepare the 0.2.0 minor release with aligned OpenAPI/MCP metadata, versioned history, embedding upgrade notes and gated development-to-main promotion.
- 2026-10-02 [Added] Add disabled-by-default global and per-source embedding controls, audited admin/CLI actions and queue admission; document scoped pending embedding archival.
- 2026-10-02 [Changed] Bind development public API/MCP/docs to all IPv4 interfaces; preserve loopback admin and release ports, verify scoped rollout and HTTP/MCP checks.
- 2026-10-02 [Added] Add manually selected municipality publication in development with private revision-checked controls, scoped API/MCP data, explicit limitations and revocable views; preserve staging/production acceptance gates.

- 2026-10-02 [Changed] Audit source acceptance prerequisites, persist incomplete campaign assessments, verify lifecycle/query gates and record remaining reviewed-evidence and municipal-regression requirements.
- 2026-10-02 [Changed] Document municipality situation diagnostics, separating public source acceptance, internal facts, acquisition freshness and unverified regional map levels.
- 2026-10-02 [Changed] Open development Scalar API documentation through a dedicated private Tailscale proxy; verify browser rendering and preserve existing routes with targeted rollback documented.
- 2026-10-02 [Changed] Expand development to six corrected worker replicas, verify preserved existing containers and queue progress, and document server concurrency limits and return to four workers.
- 2026-10-02 [Fixed] Prevent concurrent inference workers from deadlocking during queue maintenance; test cancellation and lock release, deploy four development replicas and verify parallel processing.
- 2026-10-02 [Changed] Run and verify two concurrent development fallback requests using existing worker replicas; document scope, private rollback evidence and return to one worker.
- 2026-10-02 [Changed] Record retained-source replay and bounded development local-processing results, accounting and rollout limits.
- 2026-10-02 [Changed] Reduce model dependence with reviewed article-body selection, native PDF text and deterministic regional relevance, preserving evidence and conservative fallbacks
- 2026-10-02 [Changed] Clarify operations with complete job states, archive and reason drilldowns, document backlog categories and accessible historical attempt bars; verify database, browser and development admin rollout.
- 2026-10-02 [Changed] Document and verify register-wide development fallback selection with existing model and independent source controls

## 0.1.2 - 2026-10-02

- 2026-10-02 [Fixed] Let Docker allocate loopback ports for interpretation integration databases and clean up failed starts, avoiding published-port collisions in Coverage.
- 2026-10-02 [Changed] Prepare the 0.1.2 municipal corrections patch, align OpenAPI/MCP metadata and release notes, and retain the development-to-main software promotion gates.
- 2026-10-02 [Fixed] Correct Cascina listing dates and scoped PDF selection, add reviewed CSRF interpretation reuse, verify development collection and retire only stale unattempted COC work while preserving originals and provider controls.
- 2026-10-02 [Changed] Prioritize enabled regions and current enabled municipality counts in the backoffice regional table; show enabled/total counts and verify synthetic ordering, isolated database reads and the development browser.
- 2026-10-02 [Changed] Configure tailnet-only HTTPS access to the development status frontend and verify frontend and backoffice access in the integrated browser.
- 2026-10-02 [Changed] Add reviewed source-scoped external PDF attachments, content filtering and preview/collector validation; verify failure persistence and activate the verified Livorno development revision with rollback evidence.

- 2026-10-02 [Changed] Verify Livorno's ordinance download with the application HTTP client and a normal browser; refine the earlier access finding and document validation limits.
- 2026-10-02 [Changed] Review Livorno's external attachment boundary and distinguish collector authorization from file accessibility; retain the access probe privately.
- 2026-10-02 [Changed] Recheck municipal acquisition and retained originals; update territorial guidance for footer PDF failures, add Pontedera and keep operational evidence private.
- 2026-10-02 [Added] Add territorial source guides, a shared discovery template and Municipium notes; link the main README, documentation index and coverage inventory, and require contributor upkeep with private evidence kept separate.
- 2026-10-02 [Changed] Enable development collection for the existing Cascina, Livorno and Pisa candidates through audited source controls.

## 0.1.1 - 2026-10-01

- 2026-10-01 [Changed] Bump the software patch version to 0.1.1, align OpenAPI/MCP metadata, and prepare versioned changelog history and patch release notes.
- 2026-10-01 [Changed] Apply the additive municipal migration and update development admin; verify live controls and readiness while preserving the existing collector and record partial image alignment.
- 2026-10-01 [Changed] Document explicit worker selection, environment-scoped territorial and source controls, compatible rollback and separately scheduled collector handover; record synthetic verification.
- 2026-10-01 [Added] Provide audited CLI and admin activation controls for regions, municipalities and source collection, with native forms and territorial blocking status.
- 2026-10-01 [Added] Add audited municipality enablement and hierarchical acquisition/inference gates while preserving existing collection eligibility
- 2026-10-01 [Changed] Make collection and backup workers opt-in in development and staging; verify isolated production profile selection
- 2026-10-01 [Planned] Extend the processing plan with hierarchical region/municipality/source controls in CLI and admin, independent of worker startup and public publication
- 2026-10-01 [Planned] Plan explicit development and staging background worker activation with production profile isolation and controlled shared-host processing

## 0.1.0 - 2026-10-01

- 2026-10-01 [Added] Prepare the first software release with versioned history, pilot release notes and development-to-main promotion guidance.
- 2026-10-01 [Added] Add source-scoped local Qwen chat/image fallback and real model tests; preserve strict evidence, compatible reuse and provider history
- 2026-10-01 [Fixed] Repair interpretation recovery scope, bounded reprocessing identities and evidence diagnostics; restore development acquisition and document provider entitlement recovery
- 2026-10-01 [Changed] Stop staging and rebuild development applications with the optional public status frontend.
- 2026-10-01 [Changed] Organize architecture, component and operations guides; update documentation links and verification records
- 2026-10-01 [Added] Add an independent public status frontend with coverage, freshness and bounded API reads
- 2026-10-01 [Changed] Separate backend services, private backoffice and public transports; relocate embedded API contracts
- 2026-10-01 [Fixed] Enforce environment targeting, image-preserving releases and container-readable secrets; clarify setup and validation
- 2026-10-01 [Planned] Harden environment operations, release identity, fresh-download setup and host readiness verification
- 2026-10-01 [Changed] Prepare inactive production Compose project on shared host with resource limits
- 2026-10-01 [Fixed] Update integration fixtures for Drupal identifier deduplication and equivalent-version reuse
- 2026-10-01 [Changed] Verify shared checkout migration for development and staging
- 2026-10-01 [Changed] Support isolated development and staging from one public checkout
- 2026-09-30 [Added] Add automatic changelog recording and per-commit CI enforcement for repository changes.
- 2026-09-30 [Fixed] Deduplicate generated Drupal listing identifiers and separate listing snapshots from default territorial history.
