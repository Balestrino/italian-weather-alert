# Proposal

## Why

The revision plan rechecks notices older than the bootstrap window only when their tracked `measure_state` is `ongoing` or `unresolved`. Bootstrap also includes explicitly referenced older documents. However, no production code writes `measure_state`, `relevance_state` or `explicit_reference`: every target in development is `none`, `unknown` and `false` ([CLN-033](../../../docs/territori/09-toscana/comuni/050004-calcinaia.md#cln-033--stato-misura-e-riferimenti-dei-target-non-alimentati)).

As a result, an ordinance whose closure or COC activation is still in force stops being rechecked as soon as it leaves the 30–60 day window. A later revocation, extension or correction published on the same page would then be missed. On 2026-10-21 every development notice with municipal measures has an unknown or conditional end, and several from Calcinaia and Livorno are already outside the window and no longer rechecked. The withdrawn-target protection added by `handle-withdrawn-unavailable-targets` had to fall back on "any derived measure", because the tracked state cannot be trusted.

## What Changes

- Maintain each tracked target's interpretation state from data already validated and projected: municipal domain measures, their validity, linked updates and the latest classification. The state is recomputed for the source's targets at the start of every scheduled check, so it follows the passage of time without new model calls.
- Measure states:
  - `ongoing`: some measure has a supported validity end that has not passed.
  - `unresolved`: some measure has an unknown, start-only or conditional end, and its notice is younger than a configurable per-source limit (default 90 days).
  - `aged_unresolved`: as `unresolved`, but past that limit. The notice leaves the recurring plan and is reported as no longer rechecked.
  - `ceased`: every measure has a supported end in the past, or was cancelled by a linked later measure.
  - `none`: no derived measure.
  Partial reopenings and extensions never cease a measure by themselves.
- `relevance_state` follows the latest successful classification of the document. `explicit_reference` becomes true when a linked later measure applied to the domain targets a measure of this document. It is never cleared automatically.
- The revision plan keeps its existing selection: older `ongoing` and `unresolved` notices are rechecked, and explicitly referenced ones are included during bootstrap. Withdrawn-target protection uses the maintained state and reference instead of "any derived measure". As a result, `ceased` and `aged_unresolved` notices can be withdrawn when they disappear and return 404/410.
- Show maintained states and the reason a notice is or is no longer rechecked through a CLI command and on the administration source page.
- **BREAKING (internal schema):** the `measure_state` check constraint gains `aged_unresolved`. The migration is additive and needs no data rewrite.

## Capabilities

### New Capabilities

- `tracked-target-interpretation-state`: Derivation and maintenance of tracked targets' measure, relevance and reference state from validated interpretation, the configurable limit for open-ended measures, and their visibility.

### Modified Capabilities

- `unavailable-municipal-targets`: withdrawn-target protection uses the maintained measure state and explicit reference instead of treating any derived measure as protecting.

## Impact

- `internal/backend/acquisition`: state refresh in the tracking store, called from the scheduled check before planning; withdrawn protection in the plan and in dispositions; an additive migration for the constraint and a state timestamp.
- `internal/backend/registry`: optional `discovery.unresolved_measure_days` (1–366, default 90). A configuration revision is needed only to change the default.
- `cmd/iwa/controls.go`: a new `source-tracked-targets` command.
- `internal/backoffice`: a "Documenti seguiti" section on the source page.
- Reads domain measures, temporal values, `domain_measure_updates` and classification results. When those tables are absent, the refresh keeps the tracked state unchanged.
- No new crawling beyond rechecks of the notices that the specification already requires. No inference calls, no publication or acceptance change.
- The draft `official-source-ingestion` requirement on recurring checks of ongoing or unresolved measures becomes effective; its text needs no change.
