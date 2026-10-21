# Tasks

## 1. Schema and configuration

- [ ] 1.1 Add the embedded migration `099_target_interpretation_state`, register it in `migrate.go` and verify with the acquisition PostgreSQL integration tests:
  - it replaces `acquisition_targets_measure_state_check` with the same values plus `aged_unresolved`;
  - it adds `interpretation_state_at timestamptz`;
  - repeated migration is idempotent;
  - existing rows keep their values;
  - `aged_unresolved` is accepted and unknown values are rejected.
- [ ] 1.2 Add `Discovery.UnresolvedMeasureDays` (`unresolved_measure_days`; 0 means the default of 90, valid values 1–366) with validation in `internal/backend/registry`. Verify with registry unit tests that out-of-range values are rejected and existing configurations stay valid.

## 2. State derivation

- [ ] 2.1 Implement `TrackingStore.RefreshInterpretationState` as one atomic per-source update: measure state from domain measures, validity and applied `cancels` updates; relevance from the latest classification of the latest version; explicit reference OR-ed from applied updates; `interpretation_state_at`. It must change nothing when domain or classification tables are absent. Verify with an acquisition integration test without domain tables, so tracked states stay unchanged.
- [ ] 2.2 Verify the derivation with domain-side PostgreSQL integration tests using real projected measures and applied linked updates. Cover:
  - `none`;
  - `ongoing` turning into `ceased` after a supported end;
  - `unresolved` turning into `aged_unresolved` at the limit, including a custom per-source limit;
  - a partial reopening that does not cease the measure;
  - a cancellation that ceases it;
  - conflicting validity treated as open-ended;
  - relevance from the latest version;
  - an explicit reference set by an extension and never cleared.
- [ ] 2.3 Call the refresh in `scheduledAcquire` after `Remember` and before `Plan`, with the source's limit. A failure ends the check with `interpretation_state_failed`, before planning. Verify with unit tests using a fake tracker (call order, error path) and with an end-to-end integration test where an old `unresolved` notice is planned and the same notice is no longer planned once `aged_unresolved`.

## 3. Protection and visibility

- [ ] 3.1 Switch withdrawn protection in `Plan` and `ListUnavailable` to the maintained state and reference, keep "any derived measure" for the disposition acknowledgement (`derivedMeasureSQL`), and add `measure_state` to `UnavailableTarget`. Verify with updated integration tests:
  - `ceased` and `aged_unresolved` targets can be withdrawn;
  - `ongoing` and `unresolved` targets stay blocking;
  - excluding a target with a derived measure still requires the acknowledgement.
- [ ] 3.2 Implement `TrackingStore.FollowedTargets` with recheck reasons (`window`, `ongoing`, `unresolved`, `bootstrap_reference`, `not_rechecked`), age and limit. Add the `source-tracked-targets <source-id>` CLI command. Verify with integration tests of the reasons and with `controls_test.go` for parsing, admin role, JSON output and private-detail-free errors.
- [ ] 3.3 Add the read-only "Documenti seguiti" section to the administration source page, with partial-failure handling. Verify rendering, escaping and the unavailable-list fallback in the admin tests.

## 4. Documentation

- [ ] 4.1 Update `docs/operations/acquisition-recovery.md` with the maintained states, the open-ended limit, the new command and section, and the protection change. Update CLN-033 with the implemented intervention in the Calcinaia guide. Add a CHANGELOG entry with `scripts/changelog.py`. Verify with `openspec validate --strict`.

## 5. Verification and development rollout

- [ ] 5.1 Run Go vet, the unit tests with race detection, and the acquisition, domain, backoffice and CLI integration tests. Record the results.
- [ ] 5.2 Build a development image, run `admin migrate`, recreate only `worker` and `admin` with preserved rollback settings, then verify:
  - after the first check of each municipal source, the followed list shows the measure-bearing notices with derived states and a timestamp;
  - the older measure-bearing Calcinaia and Livorno notices are planned and fetched again;
  - all active checks stay complete;
  - `check-deployment-config.py --runtime` and `smoke-compose.py development` pass.

  Record the evidence privately and the outcome in the Calcinaia and Livorno guides.
