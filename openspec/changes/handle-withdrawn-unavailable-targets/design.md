# Design

## Context

See proposal.md for motivation and the spec for the required behavior.

How the scheduled municipal check works today (`internal/backend/acquisition/preview.go`, `scheduledAcquire`):

1. It traverses every configured section. Any listing failure, unrecognized page or page-limit overflow aborts the check before discovery is recorded.
2. `TrackingStore.Remember` sets `last_seen_at` to the check's `observedAt` for every discovered URL.
3. `TrackingStore.Plan` selects the targets for this check: the revision window, ongoing or unresolved `measure_state`, explicit references, and targets without an active `exclude_stale_unavailable` disposition.
4. In the planned-document loop, a target still in 404/410 backoff sets `source_http_error` and `contentRecognized=false`. A fetch that returns 404/410 calls `RecordUnavailable` and fails the check at the end. Other crawl errors abort immediately.

Other facts that shape the design:

- `CheckOutcome` feeds `ScheduleStore.Finish`. A complete check resets consecutive failures and advances `last_complete_at`.
- Source-delay notifications derive only from `last_complete_at`.
- Dispositions live in `acquisition_target_dispositions` (`exclude_stale_unavailable` or `resume`). They can be written only with manual SQL.
- **Observed in development on 2026-10-21:** `acquisition_targets.measure_state` and `explicit_reference` are never written by production code. Every inspected target is `none` and `false`, and `SetReviewState` and `AddExplicitReference` have no callers. Guards that rely only on these columns are therefore vacuous today.

## Goals / Non-Goals

**Goals:**
- Compute the withdrawn classification inside the existing check, so it adds no extra crawling and no extra notification paths.
- Make the protection against hiding a real problem rely on data that is actually maintained.
- Reuse one guarded disposition operation for administration and CLI.

**Non-Goals:**
- Detecting renamed-URL pairs or asserting equivalence.
- Changing Cittadino Informato API detail or regional product handling.
- Populating `measure_state` or `explicit_reference` from interpretation. That is a separate gap: this change honours the columns but does not rely on them.
- Changing how source-delay notifications are computed.
- Changing how pagination is discovered on sites whose listings exceed the visited pages.

## Decisions

### 1. Classify in the engine from per-target plan attributes
`Plan` returns two new attributes for each planned document:

- `Listed`: `last_seen_at >= observedAt`, which is true when the URL was rediscovered in this traversal.
- `Protected`: the target is explicitly referenced, its `measure_state` is ongoing or unresolved, or a `domain_local_measures` row exists for any retained version of the same `source_id` and `official_url`.

The loop then decides withdrawn as `!Listed && !Protected`, for a target that is in deferral or whose fetch returned 404/410 in this check. This covers a URL that disappears and fails within the same check, which happens on the first check after a rename (the Calcinaia pattern).

- *Alternative:* a SQL-only classification in `Plan`. Rejected: it cannot classify a target whose first 404 happens during the current check.
- *Alternative:* a time grace period before a target counts as withdrawn. Rejected: absence from a complete traversal plus a 404 is already the reviewed manual criterion, and retries continue anyway.

### 2. Derive measure protection from domain measures, conservatively
Any municipal domain measure derived from any retained version protects the target. The design does not try to decide whether the measure has expired.

- *Alternative:* evaluate temporal validity. Rejected: validity is often undetermined, and a wrong "ended" verdict would silently hide a still-relevant notice.

The cost is that some old measure-bearing notices stay blocking until an operator excludes them with an explicit acknowledgement (decision 5).

### 3. Withdrawn targets do not set the error state
In the planned loop:

- A deferred withdrawn target is skipped without touching `errorCode` or `contentRecognized`.
- A fetch of a withdrawn target that returns 404/410 still calls `RecordUnavailable`, so backoff and history continue, but it does not set `unavailableErr`.
- A blocking target keeps today's behavior.
- A 200 response clears the unavailable state through the existing `ClearUnavailable` path.

`acquisitionState` gains a `withdrawnUnavailable` counter, which is copied to `CheckOutcome.WithdrawnUnavailable`.

### 4. Persist the count on checks with an additive migration
A new embedded schema file adds `acquisition_checks.withdrawn_unavailable_count integer NOT NULL DEFAULT 0 CHECK (>= 0)`. `Finish` writes the count; the lease-expiry insert keeps the default.

- *Alternative:* a new `check_state` value. Rejected: it would break existing consumers of the three-state enumeration and the completeness constraints.

### 5. One guarded disposition operation
A disposition store in `internal/backend/acquisition` provides two operations, used by both the CLI and administration:

- `ListUnavailable(source)`: unavailable targets with their classification, computed with the same predicates as decision 1 against the latest check, plus their disposition history.
- `RecordDisposition(source, configuration, url, decision, observedLastSeen, actor, reason, evidence, acknowledgeMeasures)`.

`RecordDisposition` runs in one transaction. It locks the target row (`FOR UPDATE`), requires the active configuration, an unrecovered unavailable row and `last_seen_at = observedLastSeen`, and inserts exactly one row:

- `exclude_stale_unavailable` requires no explicit reference. If the target is `Protected` by a measure, it also requires `acknowledgeMeasures`.
- `resume` requires that the latest decision is an exclusion.

An additive column `acquisition_target_dispositions.measure_acknowledged boolean NOT NULL DEFAULT false` stores the acknowledgement. Errors are stable codes: `invalid_disposition`, `target_not_found`, `target_changed`, `target_recovered`, `target_referenced`, `target_already_excluded`, `configuration_inactive`, `measure_acknowledgement_required`, `nothing_to_resume`. Evidence must be a non-empty JSON object.

### 6. CLI and administration surfaces
- **CLI** (`cmd/iwa/controls.go`):
  - `source-unavailable-targets <source-id>` prints JSON.
  - `source-unavailable-decide <source-id> <configuration> <decision> <operator>` reads one JSON object `{url, observed_last_seen_at, reason, evidence, acknowledge_measures}` from stdin, because reasons and evidence objects do not fit positional arguments. Run it with `scripts/compose-env.sh <env> run --rm -T --no-deps --pull never admin ... < decision.json`.
- **Administration**: `GET /admin/sources/{id}` gains an "URL non disponibili" section. A no-JavaScript form posts to `POST /admin/sources/{id}/unavailable-targets/decisions` with hidden URL, configuration and observed time, plus operator, reason, a JSON evidence field and the acknowledgement checkbox. It redirects back to the source with a result message, following the existing source-form conventions.

## Risks / Trade-offs

- [A transient listing omission combined with a transient 404 marks a target withdrawn for one check] → It keeps being retried and rediscovery restores it. Only targets without derived measures qualify, and the withdrawn count stays visible on the check.
- [Listings with more pages than discovery visits (Calcinaia Notizie) can omit an in-window target, so absence is weaker evidence] → A 404 is still required. The limit is documented in the territorial guide, and pagination discovery is out of scope.
- [`explicit_reference` is unpopulated, so referenced documents are not protected by references] → Measure-derived protection covers the main risk. The gap is recorded so that reference population can be added later.
- [A notice whose interpretation is still pending has no derived measure yet] → Its original is retained and interpretation continues from that version. Once a measure exists, the target becomes protected on the next check.
- [Rollback to an older image] → The added columns have defaults and are ignored by the old code. Withdrawn targets become blocking again, which is today's behavior. No reverse migration is needed.

## Migration Plan

1. Implement and test with unit and PostgreSQL integration tests.
2. Build a development image and run `admin migrate`.
3. Recreate only `worker` and `admin` with the documented image selections, preserving the rollback settings.
4. Verify on development:
   - The CLI lists Calcinaia's excluded targets.
   - A disposition round-trip (exclude then resume) works on a synthetic fixture source in integration tests, not on live data.
   - Fresh checks of the active municipal sources stay complete.
5. Wait for the next natural rename, or confirm through integration tests that a withdrawn case completes the check.
6. Record the outcome in the operations runbook and the Calcinaia guide.
7. Rollback: restore the previous image selections. The columns stay.
