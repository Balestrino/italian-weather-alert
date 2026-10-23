# Tasks

## 1. Ongoing criticality rows

- [x] 1.1 Parse «in corso fino alle ore HH.MM di <giorno>, DD <mese> YYYY» rows in the criticality table as intervals.
  - Start at the edition's literal issuance instant (`IssuanceExpression`) and end at the stated Europe/Rome instant.
  - Keep the original expression and set a per-fact assumption that the alert was already in effect at issuance.
  - Reject a row as unsupported when its end is missing or ambiguous, not later than issuance, falls in a nonexistent or repeated local time, or when the issuance cannot be parsed.

  Verify with synthetic HTML unit tests: a valid ongoing row, an end before issuance, a DST-ambiguous end and a missing issuance.
- [x] 1.2 Add an optional per-fact assumption to `RegionalFact`. The domain store writes it to `domain_temporal_values.assumption` and keeps the existing fixed text when it is empty. Verify that v5-shaped facts keep their assumption, in an existing or new domain integration test.
- [x] 1.3 Verify that a table mixing ongoing and scheduled rows reaches `MergeCriticalityMaps` with every row interval, using a synthetic table and map set reproducing the 8 October shape: orange ongoing until 17.00, then yellow. Peaks must match, no false conflict, and a contradictory overlap must still fail.

## 2. Remainder green

- [x] 2.1 In `MergeCriticalityMaps`, for each zone/risk/day with explicit intervals, emit green intervals for the uncovered parts of the local day.
  - Compute the day in Europe/Rome, handling 23 and 25 hour days.
  - Skip gaps of one minute or less.
  - Do not emit them when the map level is `not_applicable` or `unknown`, or when the table is not recognized.
  - Mark each derived fact by locator suffix and assumption.

  Verify with unit tests for:
  - an evening-only alert;
  - an alert spanning midnight across both map days;
  - an ongoing morning alert;
  - a not-applicable coastal mask;
  - a DST day.
- [x] 2.2 Bump `RegionalProjectionLogic` to `cfr-graphics-v6`. Verify that re-running projection on a v5-projected synthetic version appends v6 records without changing v5 rows, and that vigilance and monitoring facts are unchanged apart from identifiers, in a domain integration test.

- [x] 2.3 Fail closed on PDF tooling failures and give the admin service a `/tmp` tmpfs. Bump the logic to `cfr-graphics-v7`, superseding the degraded v6 pilot records of 13961 and 26641.
  - `PopplerVectorReader` returns `ErrVectorToolingUnavailable` when no temporary directory can be created or a tool cannot start. `ProjectCFR` then returns the error without persisting.
  - Verify with:
    - a unit test with an unwritable `TMPDIR`, a missing executable and a non-PDF;
    - an integration test where a failing reader persists nothing and a later run projects the maps;
    - the compose configuration check.

## 3. Historical freshness

- [x] 3.1 Refresh updating at `servedAt` only when the view's evaluation time is no earlier than its creation time minus the view lifetime. Otherwise keep the state computed at the evaluation time. Verify with HTTP and MCP tests:
  - a historical view (days earlier) stays `ok` within the threshold although served days later;
  - the existing fixed view evaluated 30 seconds before creation still turns `delayed` on a later page;
  - a future evaluation time keeps refreshing;
  - API and MCP stay equal.

## 4. Documentation and repository checks

- [x] 4.1 Update `docs/operations/cfr-graphics.md` with ongoing rows, derived green, `cfr-graphics-v6` and the freshness rule. Record the changelog entry. Verify with `openspec validate bound-ongoing-criticality-rows --strict`, `go vet ./...`, `go test ./...` and the relevant integration tests passing.

## 5. Development adoption and verification

- [ ] 5.1 Build the image from HEAD and adopt admin, worker and public in development with private backups of settings. Verify that `check-deployment-config.py --runtime` and `smoke-compose.py development` pass.
- [ ] 5.2 Reproject editions 13961 and 26641 with `project-regional-version`, then compare them with the official tables using the TOS-023 tooling. The first attempt under v6 ran in the read-only admin container and stored table-only fallback records (see design decision 4); the attempt that counts is under v7. Verify:
  - no mismatch;
  - 13961 orange on A4 ends at 17.00;
  - every derived green lies outside the listed rows.

  Only then reproject the rest of the bounded set: 15089, 24693, 27743 and the criticality editions since 4 October with explicit intervals. Record the receipts privately.
- [ ] 5.3 Rerun the A4 API/MCP comparison after the reprojection, plus one view before it. Public queries require `known_at` ≤ `evaluation_time`, so the corrected validity of past instants is checked from the current view's intervals rather than from a past view. Verify:
  - all non-green levels are exact;
  - green is explicit where the table lists no alert;
  - API equals MCP;
  - the earlier view is unchanged;
  - historical updating states are no longer `delayed` solely because of serve time.
- [ ] 5.4 Record the outcome in TOS-023 or a new TOS entry, `docs/operations/mvp-acceptance-review.md` and the task 31.5 note, without recording criticality acceptance. Verify with `openspec validate define-toscana-alert-service --strict`.
