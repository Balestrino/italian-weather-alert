# Design

## Context

See proposal.md for motivation. The existing UI uses Go templates embedded in `internal/server/admin_ui.go`. `admin_overview.go` renders global operational counts, and `ui/shell.html` exposes tool-oriented navigation. Existing forms already enforce source revisions and attribution; reuse these contracts.

`registry.Source.Territory` is free text (Toscana for regional products, ISTAT for municipal sources). `domain.Municipality` has no region field. Geography dataset selections are global by kind in `geography_dataset_selections`; historical reads already depend on immutable datasets. `operations.Alerts` reads regional bulletins globally and uses public queries for interpreted facts. Some public query predicates admit all non-municipal sources, and registry acceptance/release rules explicitly refer to `ToscanaRisks`. These are incompatible with simply adding region tabs over existing queries.

There are no main specs yet. Relevant contracts live in changes `modernize-admin-dashboard` and `define-toscana-alert-service`. Preserve their accessibility, source evidence, immutable geography, uncertainty and public first-release scope. This change adds administrative multi-region management; it does not claim national public alert coverage.

## Goals / Non-Goals

**Goals:** explicit region identity throughout administrative reads and writes, safe onboarding of a second region, source-independent complete municipality navigation, historically correct scope, and reuse of existing operational controls.

**Non-Goals:** replacing the Go UI stack, arbitrary scraping profiles generated from URLs, adapting all regional warning systems, automatic publication, bulk enabling every municipal source, or rewriting historical facts during migration.

## Decisions

### Navigation and screen structure

Keep server-rendered native links/forms. `/admin/` renders the regional overview; `/admin/regions` is its canonical catalog link. Detail routes use stable official region codes and ISTAT codes:

| Page | Sections | Principal content |
|---|---|---|
| `/admin/regions` | Enabled / all regions | State, setup readiness, municipality/source coverage, latest results, issues, configure/enable action |
| `/admin/regions/{region}` | Risultati, Comuni, Configurazione, Storico | Regional products and local-result summary, full municipality register, regional settings, audit/results history |
| `/admin/regions/{region}/municipalities/{istat}` | Dati, Configurazione, Storico | Applicable warnings, local measures/documents, sources/settings, versioned timeline |

Use a validated `tab` query parameter for sections, with search/filter/page parameters retained in links. Breadcrumbs always show Regioni → region → municipality. Sources, jobs, documents, quality and usage move under Operazioni; backups and existing global tools remain under Sistema. Existing URLs still work and expose context links when the association is known. The old overview metrics remain available under Operazioni.

Alternative: keep tool navigation and add a territory dropdown. Rejected because the user would still assemble the same context across unrelated pages.

### Explicit territorial identity and versioned associations

Add an additive territorial schema managed through existing migration conventions: official region catalog; region configuration revisions and lifecycle events; region-scoped dataset selections; versioned municipality membership and source associations. Use official region codes as identifiers, not names or province-prefix guesses. Preserve existing source and dataset IDs and legacy Territory fields for API compatibility. New creation paths must resolve an explicit association, including legacy source-create requests; an ambiguous territory is rejected rather than silently routed.

Region revisions identify selected geography datasets and available processing profiles. Source configurations remain authoritative for URLs, collection frequency, permissions, acceptance and public state. The municipal configuration page composes these existing settings and the applicable regional mapping; it introduces no implicit permission inheritance or duplicated mutable source settings. Regional sources associate with a region; municipal sources additionally associate with an ISTAT municipality. Internal comparison sources retain their role and explicit regional scope.

History resolves immutable membership/configuration versions at the event boundary. Do not rewrite retained geography datasets or associations to follow current names. Adopted snapshots may retire a municipality, but historical detail links stay resolvable and labeled historical.

Alternative: infer region from free-text Territory on each request. Rejected because it cannot isolate zones, validate associations or preserve changes in geography.

### Onboarding and enablement semantics

Provide a catalog of Italian regions from a versioned attributed official reference. New entries are unconfigured. Native setup forms accept a bounded CSV municipality import with explicit region code, ISTAT, name and province plus provenance metadata (official URL, version, verification date, checksum and completeness evidence). Preview validates all rows and counts; explicit adoption writes an immutable snapshot atomically. A partial dataset is retained only as a draft and cannot satisfy enablement. This is a user-supplied dataset workflow, not a new scheduled geographic web crawler.

Setup flow: choose region → import/validate/adopt complete register → save regional configuration → explicitly enable → configure sources. Zone and postal mappings are optional at enablement; unavailable mappings are disclosed and block dependent result claims. Dataset selections are scoped by region and kind, preserving selection time for historical queries.

Region enablement is an additional execution gate. Disabled regions are readable but cannot start new acquisition or inference, including scheduled claims and manual preview/reprocessing. Check the gate transactionally when claiming work and again at execution handoff; document the transition to executing as the boundary for work allowed to finish. Do not overwrite per-source settings or manufacture catch-up jobs when re-enabling. All mutations require actor and expected revision, producing append-only events with the before/after reference.

Alternative: define enabled as “at least one source collecting.” Rejected because it gives the operator no regional lifecycle control and confuses source health with administrative intent.

### Processing compatibility and the public perimeter

Introduce explicit processing-profile compatibility for source activation. Existing generic municipal acquisition can be selected for compatible sources in additional regions. CFR-specific regional extraction and Toscana risk/acceptance rules belong to the Toscana profile; never choose them solely because ProductID is regional. An unavailable profile yields an actionable unsupported-processing status. Adding a new region with a compatible source is supported; supporting an incompatible portal needs a separately implemented profile.

Keep public coverage pinned to the current Toscana perimeter in discovery, dataset resolution, regional facts, document/copy access and API/MCP reads. Existing public view snapshots retain their scope. Reject public enablement outside this perimeter, even for an accepted administrative source. This is deliberate reconciliation with `define-toscana-alert-service`; it avoids accidentally exposing new region data through existing broad SQL predicates. The private regional reader must not rely exclusively on public eligibility filters.

### Dedicated territorial read models

Extend operations with bounded region/municipality queries for overview, source configuration, documents, results and timeline. Apply region and optional municipality predicates before aggregates and pagination; never filter an already truncated global page. Validate region/municipality association before reading details. Use region plus dataset identity when joining zone labels; use the selected version appropriate to the observation/history boundary.

Read retained bulletins, processing statuses and consolidated facts in private scope. Keep the existing temporal/evidence semantics of alerts, using Europe/Rome for today/tomorrow, pending interpretation labels and explicit uncertain validity. A regional warning shown on a municipality page requires an applicable zone mapping; missing mapping produces an explicit limitation and a link to the regional bulletin, not a guessed municipal alert. Local and regional results remain separate.

Count all matches with the same predicates as lists; use deterministic keyset pagination (name/ISTAT for municipalities, event time/kind/stable ID for history). Avoid one query per municipality. Scope source/document/job links and validated local return paths to the same territory. Partial reader failures show unavailable sections rather than zero or empty successful results.

### History composition

Compose typed history entries from territorial events, registry versions/events, retained versions and processing/job outcomes, without copying them into a competing audit database. Expose filters for dates, source and event kind. Show recorded time, actor when present, revision/version, outcome, and evidence links. Keep valid time and observation time distinct; display retention limitations and old associations. Disabled territories remain fully inspectable.

## Risks / Trade-offs

- Global geography assumptions and broad non-municipal SQL predicates → add two-region fixtures with identical zone labels and municipal names, and test both private isolation and public non-expansion before onboarding.
- Existing sources without reliable association → backfill only verified Toscana/ISTAT associations and report unresolved records; do not guess. Block production rollout if currently active sources would become unassociated.
- National onboarding may be mistaken for universal parser support → show processing compatibility separately from enabled state and test an unsupported regional source explicitly.
- Disablement during execution → use the claim/execute boundary above and show already-running work in the UI; no claim of immediate cancellation.
- Register completeness cannot be inferred from a successful CSV parse → require attributed completeness evidence, count validation and explicit operator adoption; do not generate missing municipalities from source lists.
- Wide history and many municipalities → bounded queries, matching indexes, query-count checks and pagination tests beyond one page.
- Existing UI contracts → preserve deep links/JSON, test 390px/768px/1440px, keyboard, 200% zoom and JavaScript disabled using the existing browser harness.

## Migration Plan

1. Add schema and idempotent backfill/reporting; keep old geography history and source identifiers intact. Assign the existing verified Toscana geography and sources without changing collection/public/acceptance state. Seed Toscana enablement to preserve current behavior, not to activate dormant sources.
2. Introduce scoped reads, public-perimeter guards and execution gates together. Existing legacy global selection history remains the Toscana fallback until a verified equivalent regional selection is adopted; new region selections never write global defaults.
3. Add onboarding and territorial pages, retaining old operational routes. Validate in isolated PostgreSQL and browser fixtures with Toscana plus another region, including disabled, empty, partial and unsupported cases.
4. Record migration dry-run evidence, full tests and screenshots. Do not enable real new sources as a UI acceptance test. Deployment and real region onboarding occur through deliberate operational actions after implementation validation.
5. Rollback before any new-region activation can restore the previous binary with additive tables retained. After new regions exist, first disable their execution and verify no pending/executing work can be consumed by an older binary; retain data and associations. Do not downgrade blindly: older public queries lack territorial isolation. Prefer reverting UI alone while keeping the scoped backend, or restore a verified pre-change backup during an explicitly planned rollback.
