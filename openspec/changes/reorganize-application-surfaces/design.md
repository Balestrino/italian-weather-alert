# Design

## Context

See proposal.md for motivation. There is one Go module and one image for the existing public, admin and worker roles. `internal/server` combines private HTML/JSON handlers with public API/MCP. The root specs inventory is empty; existing active-change isolation constraints still apply.

## Goals / Non-Goals

**Goals:** clear source ownership, compatible existing listeners, portable documentation and an independently runnable public status page.

**Non-Goals:** a SPA migration, new login, schema changes, source acceptance, production activation, historical uptime monitoring or a citizen alert browser.

## Decisions

### Keep a single Go module with component packages

Move existing application packages to `internal/backend/<package>` and public transports to `internal/backend/transport/httpapi`. Extract administration into `internal/backoffice`; keep its handlers/presenters in one package and its embedded assets in `ui`, avoiding artificial subpackages with circular dependencies. Place neutral HTTP listener/readiness/JSON utilities in `internal/platform/httpserver`. Composition stays in `cmd/iwa`; backend services do not import UI packages. Separate modules would add versioning and Go internal-import constraints without current benefit.

### Give executable contracts their own home

Move the four embedded-contract files together to `api/public`. Preserve JSON bytes, embedded schema reference names and served `/docs`, `/openapi.json` and `/contratto.schema.json` paths. Evidence and public coverage documentation remain documentation. Update fixture-relative paths after nesting packages.

### Run status separately using existing HTTP reads

Add `cmd/iwa-frontend` and `internal/frontend` with embedded HTML/CSS. It uses only fixed `/health/ready` and `/v1/sources/coverage` paths on an operator-configured HTTP(S) origin, with bounded requests, no redirects and no arbitrary client-selected backend. Pagination forwards only the cursor. No JS or new framework is needed. The existing public API continues to enforce its limits; all server-side frontend queries share the frontend's upstream IP budget, documented explicitly. No new data contract or public source enablement is implied.

An optional standalone Compose project starts the frontend independently, attached only to the selected public listener network. This avoids introducing unknown services into the existing environment management and validation scripts. It grants no private-data-network access and mounts no credentials. An explicit profile selects the service; it has no dependency on backend readiness. Separate deployment can report backend failure; a shared-host outage still affects both.

### Organize guides without rewriting project history

Move development and environment/release guides into their audience directories and API usage guides into `docs/backend`. Keep `docs/README.md` as the navigation entry and add architecture and component guides. Update current links and validation commands. Original OpenSpec snapshot artifacts and dated changelog entries retain historical paths; the new architecture guide provides the old-to-new mapping. Keep `docs/coverage.md`, source evidence and reuse notices authoritative.

## Risks / Trade-offs

- Package splitting can break test-only helpers or isolation checks -> keep tests beside their subjects and retain cross-component HTTP/MCP checks.
- Nesting changes fixture paths, Docker inputs and documented checks -> verify Go, integration, build and deployment checks, and scan current documentation links.
- Frontend health can differ from backend health -> local liveness describes only the frontend; the page reports backend readiness separately.
- Upstream status calls share a request budget -> document the initial limitation and handle 429 explicitly; do not trust arbitrary forwarded headers.
- Status is a present observation, not continuous monitoring -> label timestamps and scope without uptime or national-completeness claims.

## Migration Plan

Extract HTTP utilities and backoffice, group backend packages and contracts, update runtime wiring, then add the frontend and reorganize current documentation. Build all commands and the image; run existing isolation and database suites plus synthetic frontend failures/pagination tests. Record actual validation and update tasks/changelog. No running service is changed; rollback uses the previous image and source layout without a database rollback.
