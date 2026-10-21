# Public progress snapshot

These checkboxes reflect the private project task list on 30 September 2026. A checked item records project work; it does not by itself establish public source acceptance, public deployment, or that unpublished evidence is available in this repository. Operational notes and private evidence references are maintained separately.

## 1. Territorial identity and migration

- [x] 1.1 Add region catalog, configuration revisions and lifecycle events using existing migration conventions; include an attributed versioned official region reference and verify unique codes, disabled defaults, revision conflicts and immutable events in isolated PostgreSQL tests.
- [x] 1.2 Add versioned municipality-region membership, region-scoped dataset selections and source associations with indexes; verify cross-region mismatch rejection, historical resolution and identical zone labels in two regions using database integration tests.
- [x] 1.3 Implement idempotent Toscana backfill and a read-only migration report for unresolved associations; verify existing source IDs, revisions, collection/public state, selected geography and historical query results survive migration, including a second execution of the migration.

## 2. Configuration and execution semantics

- [x] 2.1 Implement bounded municipality CSV validation, preview and atomic adoption with provenance and completeness evidence; test duplicate identifiers, wrong-region rows, partial imports, failed adoption rollback and retirement of a municipality while preserving history.
- [x] 2.2 Implement regional configuration and enable/disable commands with operator attribution and expected revisions; verify incomplete setup rejection, conflict handling, independent source flags and region-specific selection without modifying global Toscana defaults.
- [x] 2.3 Introduce explicit compatible processing profiles and territorial association for source creation/configuration, including legacy request resolution; test compatible municipal sources in a second region, unsupported regional processing, ambiguous legacy territory and preservation of Toscana source behavior.
- [x] 2.4 Enforce region execution gates in scheduled acquisition, job claims, execution handoff and manual preview/reprocessing; verify disabled regions cannot start work, running jobs can finish, other regions continue and re-enable creates no catch-up batch.
- [x] 2.5 Preserve the existing public perimeter across discovery, facts, documents, copies, view snapshots and API/MCP paths and reject out-of-scope public enablement; test with two-region retained data, overlapping zone labels and legacy Toscana clients to demonstrate no accidental public expansion.

## 3. Scoped administrative reads

- [x] 3.1 Add regional overview and complete municipality list/detail readers with region validation, search, province/coverage filters and deterministic pagination; test municipalities without sources, more than one page, full-scope counts and bounded query counts.
- [x] 3.2 Add territorial result readers for retained documents, processing states and consolidated regional/local facts independent of public enablement; test two-region isolation, today/tomorrow Europe/Rome boundaries, missing mappings, partial zones, uncertain validity and unpublished results.
- [x] 3.3 Add a composed territorial history reader with date/source/kind filters and stable pagination over configuration, source, document and processing events; test equal timestamps, historical membership changes, retired municipalities, disabled territories and explicit retention/unavailable states.
- [x] 3.4 Classify retained listing snapshots from their source configuration, omit them from default territorial history, and provide explicit listing and all-events filters; verify older snapshots and document versions remain independently accessible in synthetic PostgreSQL fixtures.

## 4. Region and municipality interface

- [x] 4.1 Make Regioni the administrative entry and organize the shared shell into Regioni, Operazioni and Sistema while preserving old destinations and overview metrics; verify route compatibility, enabled/all-region catalog states, truthful summaries and partial failures in handler tests.
- [x] 4.2 Build regional detail sections Risultati, Comuni, Configurazione and Storico with breadcrumbs and complete municipality navigation; verify server-side filters, pagination, no-register guidance, unconfigured/disabled states and region-scoped links.
- [x] 4.3 Build municipal Dati, Configurazione and Storico sections showing regional warnings separately from local measures, documents and sources; verify no-source municipalities, multiple sources, historical municipalities, evidence links and not-found responses for mismatched region/ISTAT paths.
- [x] 4.4 Add native regional setup/import/adoption/enablement forms and contextual source controls with safe return destinations; test full second-region onboarding, stale and invalid submissions, operator attribution, existing JSON requests, foreign origins and no effects from GET navigation.

## 5. Integrated acceptance and delivery evidence

- [x] 5.1 Extend the existing dashboard browser harness for region → municipality → configuration → history and second-region setup in isolated fixtures; verify 390px/768px/1440px, keyboard, 200% zoom, JavaScript disabled and error states, and retain screenshots plus observed acceptance results.
- [x] 5.2 Run `go test ./...`, `go vet ./...`, affected PostgreSQL integration suites and strict OpenSpec validation; document actual results plus an isolated migration/rollback rehearsal and an operator guide for regional setup, disablement, processing compatibility and the unchanged public perimeter. Do not use real source activation or paid inference for acceptance.

## 6. Regional entry ordering (2 October 2026)

- [x] 6.1 Count enabled municipalities in the current regional register and display enabled/total counts; order both regional entry routes and state filters by region enablement, enabled municipality count descending, then name/code; preserve unavailable labels.
- [x] 6.2 Verify rendered ordering and filters with synthetic fixtures, current membership and partial failures with isolated PostgreSQL, then refresh the development backoffice and verify the integrated browser.

Validation: `go test ./internal/backoffice ./internal/backend/operations`, `go vet ./internal/backoffice ./internal/backend/operations` and `go test -tags=integration ./internal/backoffice -run '^TestTerritorialMunicipalityReaders$' -count=1` passed. The development admin was rebuilt and refreshed; the integrated browser verified the 20-row ordering and enabled/total column via tailnet. The public service and collector retained their container and image identities.
