# Backend

The backend lives in `internal/backend`, with process composition in `cmd/iwa`.
It owns acquisition, retained documents, interpretation, territory, source
publication policy, jobs and public read models. Tests and migrations remain
with their capability packages.

The public transport is `internal/backend/transport/httpapi`. It exposes the
existing JSON API and MCP operations through shared queries while private
administration is composed separately from `internal/backoffice`.

- [Developer setup](../development/setup.md): local stack, migration and optional workers.
- [Public API and MCP contract](public-api-mcp.md): operations and transport compatibility.
- [Usage examples](public-usage.md) and [synthetic executable examples](public-usage-examples.json).
- [Public query semantics](public-queries.md), [consistent views](public-views.md), [usage limits](public-limits.md) and [retained-copy policy](public-copies.md).
- [Authoritative contract artifacts](../../api/public/README.md).
- [Architecture and dependency boundaries](../architecture/README.md).

Run `go test ./internal/backend/...` from the repository root. Add
`-tags=integration` to exercise synthetic PostgreSQL/object-storage fixtures.
Service availability, source acceptance and public source enablement are separate.
