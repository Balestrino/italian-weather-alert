# Tasks

## 1. Schema and planning attributes

- [x] 1.1 Add an embedded additive migration for `acquisition_checks.withdrawn_unavailable_count` (default 0, non-negative) and `acquisition_target_dispositions.measure_acknowledged` (default false), and register it in `migrate.go`. Verify that repeated migration is idempotent and that existing rows keep their defaults, with the acquisition PostgreSQL integration tests.
- [x] 1.2 Extend `TrackingStore.Plan` to return `Listed` (rediscovered in this traversal) and `Protected` (explicit reference, ongoing or unresolved `measure_state`, or a domain local measure from any retained version of the same source and URL) for each planned document. Verify listed, unlisted, measure-bearing and referenced fixtures in `unavailable_integration_test.go`.

## 2. Check outcome

- [x] 2.1 Change the scheduled planned-document loop in `preview.go`:
  - Withdrawn targets in deferral are skipped without an error state.
  - Withdrawn targets whose fetch returns 404/410 record unavailability without failing the check, including a first 404 in the same check where the URL disappears.
  - Blocking targets, other HTTP errors and rate limits keep their current effect.

  Count withdrawn targets in the outcome. Verify with unit tests for each branch, using a fake crawler and tracking.
- [x] 2.2 Persist `WithdrawnUnavailable` in `ScheduleStore.Finish` and keep the default for lease-expired checks. Verify with integration tests that a check with only withdrawn targets is complete, resets consecutive failures and advances `last_complete_at`; that a mixed blocking-and-withdrawn check stays incomplete with the count; and that a later HTTP 200 recovers the target and retains a new version.
- [x] 2.3 Verify with a notification integration test that source-delay incidents do not open, and recover, when only withdrawn targets remain unavailable.

## 3. Reviewed dispositions

- [x] 3.1 Implement the disposition store: `ListUnavailable` with blocking, withdrawn or excluded classification and history; and a guarded `RecordDisposition` for `exclude_stale_unavailable` and `resume`, with row locking, observed `last_seen_at`, active configuration, the measure acknowledgement and stable error codes. Verify accepted, stale, recovered, inactive-configuration, missing-acknowledgement, resume-without-exclusion and concurrent-submission cases with PostgreSQL integration tests.
- [x] 3.2 Add the CLI commands `source-unavailable-targets <source-id>` and `source-unavailable-decide <source-id> <configuration> <decision> <operator>`, with a JSON stdin payload, to `cmd/iwa/controls.go`. Verify argument parsing, admin-role enforcement, JSON output and nonzero private-detail-free errors in `controls_test.go`.
- [x] 3.3 Add the "URL non disponibili" section to the source detail page and the no-JavaScript decision form posting to `/admin/sources/{id}/unavailable-targets/decisions`, with operator, reason, evidence, acknowledgement and preserved context. Verify rendering, successful decisions, stale rejection and escaping in the admin source tests.

## 4. Documentation and reconciliation

- [x] 4.1 Update `docs/operations/acquisition-recovery.md` with the withdrawn classification, the CLI and administration disposition flow, and the rule that withdrawn targets no longer need manual exclusion. Reconcile the draft `official-source-ingestion` scenarios in `define-toscana-alert-service` that keep every unreviewed 404/410 target blocking or require "no retained original". Verify with `openspec validate --strict` for both changes.
- [x] 4.2 Record the unpopulated `measure_state` and `explicit_reference` columns as an open problem in the Calcinaia guide, with the scope of the observation. Add a CHANGELOG entry with `scripts/changelog.py`.

## 5. Verification and development rollout

- [x] 5.1 Run Go vet, the unit tests with race detection, and the acquisition, notification, backoffice and CLI integration tests. Record the results.
- [ ] 5.2 Build a development image, run `admin migrate`, and recreate only `worker` and `admin` with preserved rollback settings. Verify:
  - The CLI lists Calcinaia's existing exclusions.
  - Fresh checks of active municipal sources are complete.
  - `withdrawn_unavailable_count` is persisted.
  - `check-deployment-config.py --runtime` and `smoke-compose.py development` pass.

  Record the evidence privately and the outcome in the Calcinaia guide, without claiming source acceptance.
