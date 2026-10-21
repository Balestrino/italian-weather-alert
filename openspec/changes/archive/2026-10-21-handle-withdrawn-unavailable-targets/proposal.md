# Proposal

## Why

Municipal sites often correct a notice's slug after publication. The former URL then returns 404, disappears from the listings and keeps every later check incomplete until an operator writes a reviewed disposition by hand in SQL. Calcinaia stayed in `collection_failure` for seven days, with a continuous source-delay incident, while new notices were still being acquired and interpreted ([CLN-032](../../../../docs/territori/09-toscana/comuni/050004-calcinaia.md#cln-032--url-rinominati-e-controllo-fallito-per-una-settimana)). The same pattern produced six stale URLs in one month, each needing a manual disposition. The check outcome therefore reports a delay that does not exist, hides real delays behind a permanent incident, and costs operator time.

## What Changes

- Classify a tracked municipal target as **withdrawn unavailable** when all of these hold: its latest fetch returned 404/410, it is absent from the current complete traversal of the configured listings, it has no explicit reference and no ongoing or unresolved measure.
- A withdrawn unavailable target no longer makes the check incomplete. The check completes, records how many withdrawn targets it saw, and keeps them visible separately from collection failures.
- Targets that are still listed, explicitly referenced or supporting an ongoing or unresolved measure keep the current blocking behavior. So do targets returning other HTTP errors and targets of unsupported acquisition methods.
- Withdrawn targets stay in the revision plan while their window allows, with the existing bounded backoff. A later HTTP 200 recovers them and creates a new version. Rediscovery in a listing makes them blocking again if they are still unavailable.
- No automatic disposition, URL substitution, equivalence or cancellation is inferred. Retained originals, interpretations and failure history are preserved.
- Add an administration view and dedicated CLI commands to list a source's unavailable targets with their classification and dispositions, and to record a reviewed `exclude_stale_unavailable` or `resume` decision. The decision requires operator identity, a reason, evidence and the observed target state, under the same guarded conditions used manually today.

## Capabilities

### New Capabilities

- `unavailable-municipal-targets`: Classification of 404/410 municipal targets into blocking and withdrawn states, their effect on check completeness and visibility, and operator-reviewed dispositions through administration and CLI.

### Modified Capabilities

None in `openspec/specs/`. This change refines the draft `official-source-ingestion` scenario in `define-toscana-alert-service` that keeps every unreviewed 404/410 target blocking. After this change, that rule applies only to targets that are still listed, referenced or supporting a measure. The draft text is reconciled as part of the tasks.

## Impact

- `internal/backend/acquisition`: the revision plan, the scheduled acquisition loop in `preview.go`, tracking queries in `tracking.go`, check persistence in `schedule.go`, and an additive migration for the withdrawn count on checks.
- `internal/backoffice/admin_sources.go`: source detail plus a new disposition form.
- `cmd/iwa/controls.go`: new list and decision commands.
- Notification source-delay detection is unchanged: it follows `last_complete_at`, so it closes naturally once checks complete.
- Cittadino Informato API details and regional products keep their current behavior.
- No public API, publication gate, source acceptance or provider behavior changes.
- Rollout requires a development image update of worker and admin.
