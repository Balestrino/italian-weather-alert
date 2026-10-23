# Proposal

## Why

The task 31.5 reassessment of 23 October 2026 (TOS-023) compared 14 real CFR criticality editions with the official tables. Daily levels and explicit intervals match, but rows reading «in corso fino alle ore …» are not parsed.

When such a row is present, the whole HTML table is dropped and only the daily maps remain. On 8 October, orange on zone A4 therefore stayed current after 17:00, although the table says yellow from then on. The same reassessment found two more limits:
- outside explicit intervals the rest of the day has no record, instead of an explicit green;
- historical views always report updating as «delayed», because freshness is recomputed at serve time.

The criticality dossier cannot be accepted while a published level can outlast its official end.

## What Changes

**Ongoing rows**
- A criticality row reading «in corso fino alle ore HH.MM di <giorno>, DD <mese> YYYY» becomes an explicit interval:
  - it starts at the edition's literal issuance instant and ends at the stated instant, in Europe/Rome;
  - the original expression is kept, with an assumption stating that the alert was already in effect at issuance.
- A table with such rows is no longer discarded. All of its rows keep their precise validity, and they are merged with the daily maps as other intervals already are.

**Rest of the day**
- When a zone/risk/day has explicit intervals, the parts of that day not covered by any listed interval become explicit green intervals.
  - They are derived from a complete table and a recognized map set.
  - They carry evidence and an assumption naming the derivation.
  - No green is derived when the table or maps are unsupported, or when the map marks the zone not applicable.

**Projection and history**
- A new immutable projection logic, `cfr-graphics-v6`.
- Bounded reprojection of the retained criticality editions affected (13961, 15089, 24693, 27743) and of those with explicit intervals. Earlier knowledge stays readable at its own boundary.

**Freshness in views**
- A view requested for an explicit past evaluation time reports updating as of that evaluation time.
- Views for the current or a future evaluation time keep refreshing freshness at delivery.

## Capabilities

### New Capabilities
- `regional-criticality-validity`: validity of CFR criticality facts, covering:
  - ongoing rows;
  - the rest of a day outside explicit intervals;
  - versioned reprojection that preserves history.
- `public-view-freshness`: how saved public views report updating state for current and historical evaluation times.

### Modified Capabilities
- None. The CFR criticality and API requirements are still in the open change `define-toscana-alert-service` and have no main spec yet. The new capabilities add behavior without contradicting them: the «afternoon warning not extended to midnight» rule is extended to ongoing rows.

## Impact

**Code**
- `internal/backend/acquisition/regional_facts.go`: ongoing row parsing; issuance is available from the HTML observation.
- `internal/backend/acquisition/regional_vector.go`: `MergeCriticalityMaps` remainder intervals.
- `internal/backend/domain/regional_projection.go`: logic version `cfr-graphics-v6`.
- `internal/backend/transport/httpapi/views.go`: the refresh rule.
- Tests with synthetic HTML/PDF fixtures.

**Data**
- New appended regional records only, with no migration expected. Reprojection uses the existing admin command `project-regional-version`.

**APIs**
- For the reprojected editions, API and MCP return bounded validity for ongoing alerts and explicit green outside alert intervals.
- Historical views no longer show «delayed» purely because of serve time.
- Public operation names and parameters are unchanged.

**Operations**
- Image adoption in development, bounded reprojection, and the TOS-023 comparison rerun on the same editions.
