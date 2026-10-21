# Design

## Context

See proposal.md for motivation and the specs for required behavior.

What happens today:

- `scheduledAcquire` (`internal/backend/acquisition/preview.go`) runs `TrackingStore.Remember`, then `Plan`, then fetches the planned documents.
- `Plan` selects targets in the window, targets whose `measure_state` is `ongoing` or `unresolved`, and, during bootstrap, explicitly referenced targets.
- `SetReviewState` and `AddExplicitReference` exist but have no callers.
- Withdrawn-target protection and the operator-acknowledgement check share `measureSupportSQL`, which treats any domain measure from any retained version as protecting.

Where the interpretation data come from:

- Municipal measures are projected by `domain.ProjectMunicipalExtraction` into `domain_local_measures`, with one `validity` row in `domain_temporal_values`. The precision is `interval` (start and supported end), `instant` (start only), `conditional` or `unknown`; conflicting candidates stay `unknown`.
- Linked later measures are applied into `domain_measure_updates` with `reopens`, `cancels`, `extends`, `corrects` or `updates`.
- Classification results live in `classification_results`.
- `domain` imports `acquisition`, so acquisition must read these tables with SQL only. The domain tables are migrated after the acquisition tables and are absent from acquisition-only test databases.

In development on 2026-10-21, every notice with measures has an open end, some are older than the window, and the applied updates are `reopens` and `extends` only, with no `cancels`.

## Goals / Non-Goals

**Goals:**
- One deterministic, time-aware derivation, refreshed in the same check that plans, so the plan and the withdrawn classification always use current state.
- No inference, crawling or interpretation changes. Linked updates are reused as they are.

**Non-Goals:**
- Inferring the end of a measure from free text, publication expiry ("Data scadenza") or page disappearance.
- Treating partial reopenings or extensions as closing, or following an extension into the extending document's validity.
- Changing the domain projection, the linking or the public situation.
- Cittadino Informato and regional products. They do not use this tracked plan.

## Decisions

### 1. Refresh inside the scheduled check, before `Plan`
Add `TrackingStore.RefreshInterpretationState(ctx, source, revision, now, unresolvedDays)` and call it after `Remember` and before `Plan`. It is a single `UPDATE acquisition_targets … FROM (derived CTE)` scoped to the source, so the refresh is atomic per source. It sets `measure_state`, `relevance_state`, `explicit_reference = explicit_reference OR derived` and `interpretation_state_at = now`.

- *Alternative:* update state from the projection and linking code paths. Rejected: time-based transitions (`ongoing` to `ceased`, `unresolved` to `aged_unresolved`) need a periodic refresh anyway, and writing from `domain` into acquisition tables would couple the two packages further.
- *Alternative:* compute state inside the `Plan` query on every call. Rejected: the CLI, administration and dispositions also need the state, and a persisted value with a timestamp is auditable.

### 2. Closed and open measures
A measure is closed when its `validity` row is `interval` with `end_instant < now`, or when an applied `domain_measure_updates` row with `relation = 'cancels'` and `applied_at <= now` targets it. Every other measure is open:
- an `interval` that has not ended counts as `ongoing`;
- `instant`, `conditional`, `unknown` and a missing validity row count as open-ended.

Per target, over the measures of every retained version of the same `source_id` and `official_url`:
- `ongoing` if any open measure has an interval end at or after `now`;
- otherwise `unresolved` or `aged_unresolved` if any open-ended measure exists, depending on age;
- otherwise `ceased` if any measure exists;
- otherwise `none`.

The age is `now − coalesce(source_publication_date, first_discovered_at)`.

- *Alternative:* treat `reopens` as closing. Rejected: in the retained cases these are partial reopenings that keep other restrictions.

### 3. Open-ended limit in the source configuration
Add `registry.Discovery.UnresolvedMeasureDays` (`unresolved_measure_days`). A value of 0 means the default of 90; valid values are 1–366. The default is applied where the refresh is called, so existing stored configurations are not rewritten. A different value requires an ordinary configuration revision.

### 4. Relevance and explicit reference
- **Relevance:** take the latest retained version of the document (by `first_acquired_at`, then `id`) and its latest `classified` result. `relevant` gives `relevant`, `false` gives `irrelevant`, and no result gives `unknown`.
- **Explicit reference:** true when any applied `domain_measure_updates` row has a `target_measure_id` derived from the document. It is OR-ed with the stored value, so references are never cleared.

### 5. Missing tables and failures
The refresh checks with `to_regclass` that `domain_local_measures`, `domain_temporal_values`, `domain_measure_updates`, `classification_results` and `retained_documents` exist. If any is missing, it changes nothing. A database error makes the check fail with the stable code `interpretation_state_failed`, before any planning.

### 6. Protection and acknowledgement use different predicates
- Withdrawn protection in `Plan` and in `ListUnavailable` becomes `t.explicit_reference OR t.measure_state IN ('ongoing','unresolved')`.
- The operator-acknowledgement check in `RecordDisposition` keeps the stricter "any derived domain measure" predicate, renamed from `measureSupportSQL` to `derivedMeasureSQL`. An operator therefore still confirms before excluding a notice that ever produced a measure.
- `UnavailableTarget` gains `measure_state` for display.

### 7. Schema
A new embedded migration, `099_target_interpretation_state`:
- replaces `acquisition_targets_measure_state_check` with the same values plus `aged_unresolved`;
- adds `interpretation_state_at timestamptz` (NULL until the first refresh).

The plan index on `(source_id, source_publication_date, measure_state, url)` is still valid.

### 8. Visibility
- `TrackingStore.FollowedTargets(ctx, source, now)` lists targets with `measure_state <> 'none'` or `explicit_reference`. Each row includes the state, relevance, reference, publication date, age, limit, `interpretation_state_at`, and the current recheck reason: `window`, `ongoing`, `unresolved` or `bootstrap_reference`, or `not_rechecked`.
- New CLI command: `source-tracked-targets <source-id>`.
- Administration: a "Documenti seguiti" section on `GET /admin/sources/{id}`. It is read-only, and a read failure must not hide the rest of the page.

## Risks / Trade-offs

- [False-positive measures keep a notice rechecked for up to the limit] → The cost is bounded by the limit (one fetch per notice per check). The extra rechecks are visible in the followed list.
- [An aged notice may still be in force] → It is shown as `aged_unresolved` and no longer rechecked. A revocation published as a new notice is still discovered from the listings, and the limit can be raised per source.
- [The state lags interpretation by up to one check interval] → A measure derived after a check takes effect at the next check (one hour for municipal sources). The projection itself is unchanged.
- [Withdrawn protection is weaker for `ceased` and `aged_unresolved` notices] → This is intended. Operator exclusion still requires the acknowledgement, and originals stay retained.
- [More fetches right after rollout] → The measure-bearing development notices older than the window re-enter the plan, adding a few fetches per check for Calcinaia and Livorno.

## Migration Plan

1. Implement, then run the unit and PostgreSQL integration tests, including the domain-side tests that need real projected measures.
2. Build a development image, run `admin migrate`, and recreate `worker` and `admin` with preserved rollback settings.
3. Verify on development:
   - after the first check, the followed list shows the measure-bearing notices as `unresolved`;
   - the older measure-bearing notices are fetched again;
   - all active checks stay complete.
4. Record the outcome in the Calcinaia and Livorno guides.
5. Rollback: restore the previous image selections. The additive constraint value and column stay. The old code ignores `aged_unresolved` (it is not planned) and treats it like `none`.
