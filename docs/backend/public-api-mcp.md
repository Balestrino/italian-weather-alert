# Public JSON API and MCP

## Interactive documentation

Open `/docs` on the public listener (locally, `http://127.0.0.1:8080/docs`)
for Scalar's interactive reference. `/openapi.json` serves the embedded OpenAPI
contract and `/contratto.schema.json` serves its referenced response schemas.
The reference includes the five query groups below and the retained-copy
download endpoint. Requests from the reference target the current public server
directly and use its usual access rules and rate limits.

The page loads Scalar 1.72.0 from jsDelivr, so interactive rendering requires
browser access to that CDN. The JSON specifications are served locally. These
documentation routes are available only on the public listener.

For a browser on another machine, the VM's loopback address is not the browser's
localhost. An operator-selected Tailscale Serve HTTPS port can proxy the public
listener privately, preserving existing admin/frontend routes. For example:

```sh
sudo tailscale serve --bg --https=8444 http://127.0.0.1:8080
```

Open `https://<vm-tailnet-name>:8444/docs`; Scalar loads the same-origin contract
and targets that public API origin. Preserve the prior Serve configuration and
verify browser rendering and `/openapi.json`. To remove only this proxy, run
`sudo tailscale serve --https=8444 off`. The 2 October 2026 development browser
check rendered all six documented endpoints; existing Serve routes were preserved.
Environment addresses and rollback captures remain private.

Integration follows the [Scalar HTML/JS documentation](https://scalar.com/products/api-references/integrations/html-js).

Task 6.2 exposes the five read-only query groups from `internal/publicquery`
on the public listener. Access is anonymous: neither HTTP nor MCP requires an
account, API key, or authorization header.

## Public surface

| Query group | JSON API | MCP tool |
| --- | --- | --- |
| Municipality and zone discovery | `GET /v1/municipalities` | `discover_municipalities` |
| Municipality situation | `GET /v1/municipalities/{istat}/situation` | `get_municipality_situation` |
| Document, measure, or regional search | `GET /v1/search` | `search_alerts` |
| Document and retained versions | `GET /v1/documents/{id}` | `get_document` |
| Source coverage and updating | `GET /v1/sources/coverage` | `get_source_coverage` |

Every operation uses the same application dispatcher for parsing temporal
boundaries, calling the read-only query interface, constructing `data`/`meta`
or error envelopes, and mapping failures. Successful metadata includes a
content-derived `dataset_version`, evaluation and delivery times, history
bounds/gaps, limitations, view expiry, and a null next cursor. MCP application
failures carry the same structured error and set `isError=true`; protocol
failures remain protocol errors.

The MCP endpoint is `/mcp`, using stateless Streamable HTTP and protocol
`2026-07-28`. The official Go SDK supplies `server/discover`, standard method
and name header validation, and JSON-RPC handling. The service additionally
rejects other protocol versions, limits request bodies, propagates disconnect
cancellation, and applies Go cross-origin protection. The five descriptors and
their complete input/output JSON Schemas are embedded directly from the
reviewed `api/public/mcp-tools.json` artifact.

Only the public role registers these routes and tools. The administrative role
does not register `/v1` or `/mcp`; the public query interface has no collection,
inference, source mutation, or administration method. Public calls therefore
cannot schedule upstream work through this transport.

## Deliberate current boundaries

Task 6.2 does not claim the later operational controls are complete. Task 6.3
adds the [shared public request budget and response cap](public-limits.md).
Task 6.4 adds [persistent consistent views and authenticated cursors](public-views.md),
so `dataset_version`, `page_size`, and `cursor` now share snapshots across both
interfaces. Task 6.5 adds [application-mediated retained-copy access](public-copies.md),
disabled by default and reauthorized at each delivery. Task 6.6 publishes
[verified API/MCP consumer examples](public-usage.md) for all five groups.

## Verification

Run from the repository root:

```sh
go test ./internal/backend/transport/httpapi ./internal/backoffice -run 'TestPublic' -v
go test -race ./...
go vet ./...
docker build -t iwa-app:task-6.2 .
openspec validate define-toscana-alert-service --strict
```

The server test connects through `httptest.Server` with the official MCP Go
client, negotiates protocol `2026-07-28`, lists exactly the five reviewed tools,
and calls every tool. For identical explicit `evaluation_time` and `known_at`,
each structured MCP result is compared with the corresponding anonymous JSON
API response. Tests also cover structured MCP application errors, exact
read-only annotations, legacy-protocol rejection, cross-origin rejection, and
absence of public routes/tools from the administrative listener. The fake used
at this transport boundary implements only the five read methods and verifies
exactly one application call per API/MCP request; the separate task-6.1
PostgreSQL fixture verifies that those shared queries do not change collection
or inference records.

The task-6.6 example verification additionally executes the published JSON
examples against a local service, rejects administrative paths on the public
listener, checks that no generated-safety field is present in the public MCP
schemas, and uses the PostgreSQL fixture to prove a retained internal DPC
comparison document remains absent from public search and coverage.
