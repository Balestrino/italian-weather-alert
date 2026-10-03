# Public progress snapshot

These checkboxes reflect the private project task list on 30 September 2026. A checked item records project work; it does not by itself establish public source acceptance, public deployment, or that unpublished evidence is available in this repository. Operational notes and private evidence references are maintained separately.

## 1. Shared shell and Jobs pilot

- [x] 1.1 Introduce bundled shared templates, dark compact CSS tokens, responsive navigation and private asset routes; apply the shell to the landing page and Jobs pilot, narrowly extend CSP, and verify asset MIME types, local/tailnet access, foreign-origin rejection and public-route absence with server tests plus a desktop/mobile browser check.
- [x] 1.2 Replace the Jobs generic table/action list with labeled primary columns, text status badges and expandable evidence/actions; add allowlisted server-side job-state filtering with preserved source/job context and pagination, and verify escaped long values, empty/error states, filter round trips and database filtering through focused server and PostgreSQL tests.
- [x] 1.3 Add the native guided job-relaunch form and readable HTML outcome/error views while retaining JSON and payload-form compatibility; verify operator/attempt validation, stale-attempt rejection, duplicate-field rejection, single-job scope and successful no-JavaScript submission.

## 2. Operational overview

- [x] 2.1 Add snapshot-consistent source-issue, failed-job and pending-document aggregates plus matching source-issue filtering; verify counts and drilldown predicates with isolated PostgreSQL fixtures containing more than 100 entities, repeated attempts, mixed states and empty scopes.
- [x] 2.2 Wire overview summaries and open incidents into the administrative landing page with timestamps, explicit refresh and matching links; verify partial dependency failure, unavailable versus zero, absence of side effects and browser navigation from each summary.

## 3. Operational sections

- [x] 3.1 Migrate Sources status, Documents, Findings and CFR-DPC reports to shared section-specific tables with expandable full evidence, contextual actions and preserved supported URL filters; verify each report family for escaping, empty/error states and navigation/pagination context using focused handler tests and browser inspection.
- [x] 3.2 Migrate Usage reports with readable units/timestamps and compact attempt details; verify partial usage, unknown costs, distinct currencies, workload/stage/model breakdowns and price provenance remain visible and JSON output remains compatible using existing accounting fixtures and focused render tests.

## 4. Source management

- [x] 4.1 Migrate source registry, details and history to the shared shell, group configuration and evidence into readable sections, and label advanced JSON editors; verify all existing source destinations and complete configuration/history evidence remain accessible, escaped and usable on narrow screens.
- [x] 4.2 Add guided source preview, collection activation/pause and interval forms with native decoding and retained attribution/revision checks; verify successful submissions, stale revisions, duplicate/invalid fields and existing JSON/payload-form compatibility through source handler and integration tests.
- [x] 4.3 Add guided acceptance, public enablement, interpretation pause/resume and selective reprocessing forms with explicit target/scope and readable outcomes; verify prerequisites, revisions, separation of collection/publication, absence of unintended broad reprocessing and no-JavaScript submission using lifecycle integration tests and browser checks.

## 5. Secondary pages and enhancements

- [x] 5.1 Apply the shell and readable detail/result conventions to backups, notifications, observation campaigns/trial costs, processing evaluations/regression and release diagnostics; record an HTML route inventory and verify every family, including unavailable dependencies, while preserving advanced diagnostic forms and machine-readable endpoints.
- [x] 5.2 Add current-page search, accessible navigation toggling and visible disableable shortcut help; verify slash/Escape behavior, editable-field and modifier exclusions, focus return, search scope labeling and full core usability with JavaScript disabled in browser tests.

## 6. Integrated acceptance and documentation

- [x] 6.1 Verify the complete console at 390px, 768px and 1440px and 200% zoom with keyboard navigation, contrast checks, long evidence, empty/error/partial-data states and no third-party requests; fix defects and retain screenshots plus the executed browser procedure as reviewable evidence.
- [x] 6.2 Update deploy/administration.md with navigation, guided/advanced actions, freshness semantics, verification and binary rollback; run go test -race ./..., go vet ./..., the affected isolated PostgreSQL integration checks and strict OpenSpec validation, document actual results and keep deployment separate from local implementation completion.

## 7. Today and tomorrow alerts (operator request, September 23)

- [x] 7.1 Add a private read-only today/tomorrow page with separate official regional and municipal sections, Europe/Rome dates, latest retained bulletin provenance, explicit unavailable/unconsolidated/undated states and no inference side effects. Verify date boundaries, escaping, private access, database reads and responsive browser rendering.
- [x] 7.2 Deploy the verified page to the existing admin URL and record health, unchanged recovery limits/provider usage and navigation checks.

## 8. Complete operations overview (operator request, October 2)

- [x] 8.1 Add snapshot-consistent job state totals and queue/kind summaries, recorded reason counts and matching state/kind/queue/error/archive drilldowns; verify full counts, archive separation and pagination with isolated PostgreSQL.
- [x] 8.2 Explain incomplete document totals using exclusive per-source categories with the existing list exclusions; verify multiple OCR jobs count once, suspended sources, missing triggers and incomplete results.
- [x] 8.3 Add UTC historical attempt bars for 24 hourly buckets, 7 days and 30 days, including zero buckets, retries and archived work, with exact accessible values and truthful partial/cutoff semantics; verify SQL boundaries and server rendering.
- [x] 8.4 Verify private access, input validation, safe rendering, independent unavailable sections, responsive/no-JavaScript browser behavior and ordinary/race/vet checks; document scope and validation.
- [x] 8.5 Update only the existing development admin service and verify overview, periods, drilldowns and readiness; preserve rollback identity and private evidence separately.
