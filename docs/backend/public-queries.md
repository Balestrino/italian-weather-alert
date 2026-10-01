# Shared public queries

Task 6.1 introduces `internal/publicquery`, the read-only application layer that
both the JSON API and MCP transport will call. No method in this package writes
to the database, reads object bytes, schedules collection, or starts inference.

## Operation groups

The package implements the five reviewed contract groups:

- `DiscoverMunicipalities`: name, ISTAT, postal candidates and selected
  municipality-zone mapping;
- `MunicipalitySituation`: local measures first, operational phases, distinct
  regional products, newer uninterpreted documents and regional/local coverage;
- `Search`: documents, measures or regional facts with compatible filters and
  documented temporal intersection behavior;
- `Document`: a selected/latest eligible version and optional retained version
  history;
- `Coverage`: public-eligible regional and municipal sources, including pending
  and suspended states rather than treating missing local coverage as all-clear.
  Every source also exposes `coverage_status` and `coverage_limitations` from
  the final source-scoped acceptance review. `accepted_declared_scope` applies
  only to the configured sections; `accepted_with_limitations` is never promoted
  to a municipality-wide or three-product completeness claim.

Each query accepts an evaluation time and a service-knowledge boundary
(`known_at`). Retained versions and derived facts recorded after that boundary
are excluded, and geographic lookups use the dataset selection recorded by that
boundary rather than a later selection. Migration `019_public_queries` adds immutable `recorded_at`
timestamps to regional records, local measures and operational phases so that a
later interpretation of an older source cannot leak into an earlier knowledge
view. Public facts are limited to public-enabled sources; coverage also reports
eligible pending and suspended sources. DPC comparison sources are excluded by
the registry's `public_eligible` policy.

History bounds are computed from first acquisition inside the applicable query
scope. A request beginning before that bound returns an explicit gap, and every
history result states that retained service knowledge is not a complete
pre-startup archive. Document metadata keeps official links and exact version
hashes. The query layer leaves copy links null; task 6.5's transport delivery
layer now evaluates the current default-off, source-restricted policy on every
response and fills only authorized exact-version links.

The query types use the field names and separate provenance, interpretation and
updating dimensions from the task-1.7 JSON contract. Cursor persistence,
pagination snapshots and `dataset_version` are intentionally left to task 6.4;
HTTP/MCP envelopes and transport equivalence are task 6.2; shared per-IP budgets
are task 6.3.

## Verification

Run from the repository root:

```sh
go test -tags=integration ./internal/backend/publicquery -run TestFiveSharedPublicQueryGroups -v
go test -race ./...
go vet ./...
docker build -t iwa-app:task-6.1 .
openspec validate define-toscana-alert-service --strict
```

The isolated PostgreSQL fixture exercises all five groups. It verifies postal
discovery with historical mapping selection, a prominent current Calcinaia closure, a distinct future regional
criticality fact, a newer uninterpreted document, two stable-URL versions,
scoped history gaps, independent regional/local coverage and a `known_at`
boundary that cannot see the later document revision. The fixture is synthetic
and does not activate a real source or claim source acceptance.
