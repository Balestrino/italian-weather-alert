# Public API latency

The 8 October 2026 repair reduces repeated database work when constructing a
complete public query. API and MCP still use the same query service and saved
views; historical boundaries, publication controls and evidence remain intact.

Version chronology, acquisition visibility and receipt lookups have additive
indexes in migration `069_public_query_lookups`. Existing migration checksums
remain unchanged. Receipt misses use indexed existence checks before planning
full visibility/evidence queries. Applicable municipality zones are filtered in
SQL before constructing regional facts.

Each normalized operation owns its source quality, document interpretation,
metadata, attachment and evidence reads. Nested operations share that state at
the same knowledge/evaluation boundary; later requests start fresh. Fact-specific
limitations stay independent. PostgreSQL timestamp comparisons retain microsecond
precision. Readers release result sets before nested queries so a small connection
pool can serve concurrent requests without waiting on connections held by those
same requests. Municipality discovery batches source coverage and development
publication selection across candidate ISTAT codes.

## Verification

Use a fixed `evaluation_time` and `known_at` to compare complete results before
and after a change. Measure HTTP status and elapsed time independently of body
size; include discovery, coverage, situation, regional/measure/document search,
a document with version history, and saved-view continuation. Compare API and MCP
against the same saved view and page size. Updating status still reflects delivery
time, so report any genuine freshness change separately from data differences.

Synthetic PostgreSQL tests cover idempotent migration, receipt visibility,
historical interpretation and projection, source suspension and publication
revocation. Regressions check one source-quality read per source per operation,
new-request isolation, independent limitations, timestamp precision and concurrent
coverage/discovery/version-history reads with one database connection. Unit/race
checks and the development listener build complete the implementation checks.

Keep live profiles, query plans, fixed-time responses, container identities,
configuration snapshots and detailed rollout/rollback records under ignored
`.local/operations/performance/`. Development adoption is independent of source
acceptance and production release approval.

## Development adoption and rollback

Build the tested development image, preserve the existing listener image and
configuration, and apply migrations through an admin-role one-off. Replace only
the development public service with `--no-deps --no-build --pull never`; verify
readiness, unchanged other containers, complete response equivalence and timing.
Use `scripts/compose-env.sh development` for the selected environment and keep
private `IWA_PUBLIC_IMAGE` selection in `.local/development.env`. The development
public listener defaults to `IWA_APP_IMAGE` when no override is supplied; staging
and production continue to pin every application surface to `IWA_APP_IMAGE`.

Restore the previously recorded public image/configuration to roll back the
listener. The added indexes can remain: older software ignores the additional
migration record and uses the same data schema. No source, acquisition,
interpretation, saved-view or publication data is rewritten by this repair.
