# Proposal

## Why

Public transports and the private administrative UI currently share `internal/server`, while executable API contracts live under documentation. Explicit component boundaries will make ownership clearer and provide a place for a public system-status frontend.

## What Changes

- Group application services under `internal/backend`, extract private handlers and assets into `internal/backoffice`, and retain one root Go module.
- Move embedded public contracts to `api/public`, preserving their served URLs and contents.
- Introduce an independently runnable, server-rendered public frontend for service availability, source publication coverage and source updating information, using existing public HTTP reads only.
- Organize documentation by architecture, development, component and operations; keep deployment configuration under `deploy` and coverage status in `docs/coverage.md`.
- Update build, fixtures, CI paths and contributor instructions while preserving existing API/MCP and administrative behavior.

## Capabilities

### New Capabilities

- `public-system-status`: a public, read-only status UI with separate availability, publication coverage and freshness information, including explicit failures and pagination.

### Modified Capabilities

None. Main specifications are empty. Existing administrative isolation and public-access requirements in active changes remain constraints; file movement does not change those requirements.

## Impact

Go imports, the shared HTTP package, command runtime wiring, embedded contracts/assets, Dockerfile, CI and fixture paths, documentation links, and a new standalone frontend command. No schema migration, source enablement, production deployment or new frontend framework is required.
