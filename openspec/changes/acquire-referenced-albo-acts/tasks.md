# Tasks

## 1. Data model and registry

- [x] 1.1 Add migration `101_referenced_acts` with `act_references`, `act_resolutions`, `act_albo_records` and `act_clarifications`. All four are append-only (`processing_catalog_immutable()` triggers) and take part in retention. Verify idempotent migration and that UPDATE and DELETE are rejected, in a PostgreSQL integration test.
- [x] 1.2 Add the `jcitygov-albo` access method and the `JCityGovAlbo` configuration (origin, list and detail paths, allowed categories, page bound, served primary source) with validation mirroring Cittadino Informato. The policy must permit collection and retention and must not permit copies. Verify with unit tests for valid configurations and for each invalid one.

## 2. Albo client

- [x] 2.1 Run a private, read-only preview of the Calcinaia albo:
  - list paging;
  - exact registry search;
  - detail parsing;
  - unsigned attachment download;
  - whether the OpenFormat export is complete and stable.

  Record redistributable synthetic fixtures of the page shapes. Keep captures under `.local/operations/`. Verify that the fixtures reproduce the observed fields for Ordinanze 34/2026 and 38/2026.
- [x] 2.2 Implement the JCityGov client: bounded binary search of list pages by registry order for a date window, page reads within the bound, exact registry lookup, detail parsing (category, act number and date, registry, object, attachments) and attachment download with hash. Every failure is typed and fails closed. Verify with fixture-based unit tests, including truncated pages, markup changes and the page bound.

## 3. Detection and resolution

- [x] 3.1 Implement deterministic reference detection over interpreted notice content with the reviewed patterns and the navigation fingerprint exclusion. Verify with synthetic fixtures:
  - positive: numbered with year, numbered with date, unnumbered operative forms;
  - negative: footer link, "Visto" boilerplate, negated or hypothetical forms.
- [x] 3.2 Implement resolution in the order `attached`, `numbered`, `unique_candidate`, then `not_found`, `ambiguous` or `unavailable`, with the category, date-window and disposition-term rules, appending outcomes. Verify with unit and integration tests for each outcome:
  - an ordinance and a decree with the same object;
  - two matching ordinances;
  - an unreachable albo, then a later successful retry.

## 4. Acquisition and clarification

- [ ] 4.1 Add the `resolve_act_references` and `acquire_albo_act` jobs. Enqueue them after a projection with measures, at most 3 references per version. Retain the detail and the PDF once per albo record, run only OCR on albo versions, and exclude albo sources from classification, extraction, projection and coverage. Verify with integration tests: idempotent jobs, a shared act for two notices, no interpretation runs for albo versions, unchanged coverage.
- [ ] 4.2 Implement the literal dispositive-clause parser for municipal emergency ordinances (start, end or condition, dispositions with kinds and items). Verify with fixtures from both ordinance texts and negative templates that yield `clarification_unavailable`.
- [ ] 4.3 Implement the `clarify_act_measures` job. Match notice measures to dispositions by kind and subject head term. Append the validity and the interpretation event for measures that are unknown or conflicting; record conflicts for determinate differing validity. Verify each spec scenario in integration tests, including the conditional start of 2026-09-10T00:00 Europe/Rome, the conflict case and an unmatched disposition.

## 5. Interfaces

- [ ] 5.1 Add act evidence items to public measure evidence within the cutoff, and deny copies for albo versions. Verify:
  - integration tests for a boundary before and after a clarification, and for a copy request;
  - a transport test that HTTP and MCP return identical clarified measures.
- [ ] 5.2 Add the read-only admin view and the CLI command `source-act-references <source>`. Verify with handler and CLI tests, including no edit routes.

## 6. Verification and adoption

- [ ] 6.1 Run unit, race, vet and `-tags=integration` for the touched packages and `cmd`, plus `openspec validate --strict`, and record the results.
- [ ] 6.2 Adopt in development: apply the migration through admin, register and preview `calcinaia-albo` with a settings backup, and deploy admin, worker and public from one revision. Process the acceptance notices with bounded reprocessing and verify through API and MCP:
  - doc 60: `unique_candidate` → Ordinanza 38/2026, measures conditional from 2026-09-10T00:00;
  - doc 213: resolution to Ordinanza 34/2026, or `ambiguous`/`not_found` with the reason;
  - doc 59: `attached` Ordinanza 34/2026;
  - earlier boundaries unchanged.

  Record the private evidence.
- [ ] 6.3 Update the Calcinaia guide, the operations documentation and the changelog, and update task 31.4 of `define-toscana-alert-service` with the verified results and the remaining limitations.
