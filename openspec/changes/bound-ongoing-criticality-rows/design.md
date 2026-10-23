# Design

## Context

How CFR criticality is projected today:
- `ProjectRegionalHTML` reads the criticality table. `rowInterval` accepts only «dalle ore … alle ore …» with two clock/date pairs. Any other row returns `ErrUnrecognizedContent` for the whole table.
- `ProjectCFR` then uses the vector maps alone. The table is discarded, so `MergeCriticalityMaps` never runs and every fact keeps `date` precision.
  - Editions 13961, 15089, 24693 and 27743 projected this way; they have no interval facts.
  - On 8 October orange stayed current on A4 after 17:00.
- When intervals exist, `MergeCriticalityMaps` replaces the day fact with those intervals only, so the rest of the day has no record.
- Validity rows are written in `domain_temporal_values`. The domain store sets one fixed Europe/Rome assumption for every interval or instant.
- Projection records are immutable and keyed by logic version (`cfr-graphics-v5`). `project-regional-version <id>` reprojects one retained version idempotently under the current logic.
- Public views store `EvaluationTime` and `CreatedAt`. `refreshData` recomputes updating at `servedAt` for every view.

Evidence and counts: TOS-023 and the private register `criticality-monitoring-31-5-20261023`.

## Goals / Non-Goals

**Goals:**
- Bounded validity for ongoing rows, with no table discarded because of them.
- Explicit, labelled green for the rest of a day that has explicit alert intervals.
- Truthful updating state for historical views.
- Reprojection of the affected retained editions without rewriting history.

**Non-Goals:**
- Changing vigilance or monitoring interpretation.
- Inferring colours from narrative text.
- Repairing the 7 October interim knowledge, before the v5 reprojection. It stays as it was known.
- Source acceptance; criticality acceptance remains an operator decision.
- New public parameters or schema migrations.

## Decisions

1. **Ongoing row start = literal issuance instant.**
   - The row says the alert is already in effect, and the issuance instant is the earliest moment this edition can assert it. Its earlier start belongs to the previous edition, which the service already serves at its own knowledge time.
   - The issuance comes from `ObserveRegionalHTML` (`IssuanceExpression` → `regionalIssuanceInstant`). Without a parseable issuance the row is unsupported.
   - Alternatives rejected:
     - start of the day: invents validity before issuance;
     - start copied from the previous edition: couples editions and fails when that edition is missing;
     - `conditional` precision: loses the stated end and cannot be compared with map peaks.
   - The fact keeps the literal expression and a per-fact assumption: «already in effect at issuance; earlier start in preceding editions».

2. **Per-fact assumption.**
   - Add an optional assumption to `RegionalFact`; the domain store uses it when present and the current fixed text otherwise.
   - Existing records and v5 behavior are unchanged.
   - No migration, because the `assumption` column already exists.

3. **Remainder green derived inside `MergeCriticalityMaps`.**
   - For each map fact whose zone/risk/day has explicit intervals, compute the complement of their union within the local day: [00:00, next 00:00) in Europe/Rome, which handles 23 and 25 hour days.
   - Emit `green` intervals for the gaps, but only when the map fact level is a colour (not `not_applicable` or `unknown`) and the table was recognized.
   - Each derived fact carries:
     - the map locator plus «derived: no listed alert for this period»;
     - the original expression of the map day heading;
     - the derivation assumption.
   - Gaps shorter than one minute, such as the 23.59 boundary, are skipped. The CFR uses 23.59 as end of day, so a one-minute green at 23:59 would be noise.
   - Alternative rejected: keep the day fact alongside the intervals. That reintroduces the midnight-extension problem the existing requirement forbids.

4. **Logic version `cfr-graphics-v6`.**
   - Bump `RegionalProjectionLogic`. New acquisitions use v6 automatically.
   - Retained editions are reprojected explicitly with `project-regional-version`, bounded to:
     - the four editions with ongoing rows;
     - the criticality editions since 4 October that have explicit intervals.
   - Vigilance and monitoring reprojected under v6 must produce facts identical to v5 apart from identifiers. This is verified on the same retained editions and not reprojected unless needed.

5. **Freshness refresh only for current and future views.**
   - In `refreshData`, refresh at `servedAt` only when `view.EvaluationTime` is not before `view.CreatedAt`.
   - Otherwise keep the state computed by `publicquery` at the evaluation time.
   - The decision uses stored view fields, so cursor pages and MCP calls behave the same.
   - Alternative rejected: a request flag. It adds a public parameter and lets API and MCP diverge.

## Risks / Trade-offs

- [Derived green states something the bulletin does not draw.] → It is limited to editions whose table and full map set are recognized, labelled as derived, and never produced for not-applicable or unresolved zones. The TOS-023 comparison rerun checks every derived interval against the official table.
- [Ongoing start at issuance understates the true start.] → The assumption text says so. Earlier editions keep serving their own intervals at their knowledge times.
- [Merge conflicts may increase once ongoing rows enter the merge (table/map peak mismatch).] → The existing failure path keeps the supported fallback and reports the limitation. The reprojection report is checked for `partial` or `unsupported` statuses before adoption.
- [Historical freshness differs from what clients saw before.] → This is intended. Current views keep their behavior, and the change is covered by API/MCP equivalence tests.

## Migration Plan

1. Implement with synthetic fixtures.
   - Ongoing row cases: valid, ended before issuance, DST.
   - Mixed tables.
   - Remainder green, including not-applicable masks.
   - Freshness for historical, current and future views.
2. Build the image and adopt the worker, admin and public services in development. No migration.
3. Reproject the bounded edition set with `project-regional-version`. Record the receipts and the knowledge time.
4. Rerun the TOS-023 comparisons:
   - day-level and interval checks;
   - A4 API/MCP instants at the same plan, with known_at after the reprojection;
   - the historical view before the reprojection.
5. Rollback.
   - Restoring the previous images stops new v6 projections.
   - v6 records already appended stay current, because queries select each version's latest projection whatever the running code.
   - A defect in them is corrected by a later logic version and a bounded reprojection, never by deletion.
   - For this reason the reprojection set is first run on two editions (13961 and 26641) and compared before the rest.
