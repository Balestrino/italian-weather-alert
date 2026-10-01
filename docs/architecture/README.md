# Application architecture

IWA keeps one repository and one root Go module, with three explicit component areas:

```text
cmd/                         process composition and operational commands
  iwa/                       public API, private backoffice and workers
  iwa-frontend/              independently runnable public status page
internal/
  backend/                   application services and persistence
    transport/httpapi/       public JSON API and MCP
  backoffice/                private HTML/JSON handlers and presenters
    ui/                      embedded administrative templates and assets
  frontend/                  public status presentation and HTTP client
    ui/                      embedded public templates and assets
  platform/httpserver/       listener, readiness and JSON response utilities
api/public/                  authoritative embedded API/MCP contracts
docs/                        human-readable guides and public coverage tracker
deploy/                      Compose definitions and environment examples
scripts/                     development, validation and release utilities
openspec/                    specifications, plans and progress
```

## Responsibilities and dependencies

The backend owns collection, interpretation, publication policy, durable jobs,
territorial data and public query semantics. Existing packages remain organized
by capability inside `internal/backend`; migrations and tests stay beside the
services they verify. Public HTTP/MCP handlers receive read-only interfaces.

The backoffice owns private route registration, operator forms, HTML/JSON
presentation and local assets. Its handlers call backend interfaces; it does not
duplicate lifecycle rules or publication decisions. Existing `/admin/...` URLs,
JSON clients, Host/Origin checks and loopback/tailnet restrictions remain intact.

The public frontend owns status presentation. It calls only the configured public
backend's `/health/ready` and `/v1/sources/coverage` HTTP endpoints. It has no
database, object-store, crawler or inference credentials and no general-purpose
proxy route. Its process can keep reporting unknown/unavailable backend status
when the backend stops. The [frontend guide](../frontend/README.md) describes the
separate command and optional Compose project.

Composition belongs in `cmd/`: backend production packages must not import
backoffice or frontend packages. Cross-component transport tests can import both
surfaces to verify their isolation. Shared HTTP utilities provide listener
lifecycle, readiness and JSON formatting; they do not register UI routes.

## Contracts and source status

`api/public` contains one authoritative copy of OpenAPI, the shared JSON schema
and MCP tool descriptors, embedded into the public service. `/docs`,
`/openapi.json` and `/contratto.schema.json` retain their existing URLs. The
[backend guide](../backend/README.md) links to usage and compatibility details.

The [coverage tracker](../coverage.md) records documented source acceptance and
public enablement. The frontend is a runtime observation of eligible sources;
it is not a replacement for that reviewed record. Neither a ready HTTP service
nor a municipality in the registry establishes complete source coverage.

## Migration from the earlier layout

| Earlier location | Current location |
| --- | --- |
| `internal/<capability>` | `internal/backend/<capability>` |
| `internal/server/admin*.go`, `internal/server/ui` | `internal/backoffice` |
| Public files in `internal/server` | `internal/backend/transport/httpapi` |
| Listener/readiness utilities in `internal/server/server.go` | `internal/platform/httpserver` |
| Executable contracts in `docs/prerequisiti-mvp` | `api/public` |
| `docs/developer-setup.md`, `docs/agent-tools.md` | `docs/development/setup.md`, `docs/development/agent-tools.md` |
| `docs/environments.md`, `docs/environment-verification.md` | `docs/operations` |
| `deploy/release-operations.md` | `docs/operations/releases.md` |
| `deploy/public-*` guides and usage examples | `docs/backend` |

Historical OpenSpec snapshot prose and dated changelog entries can still name the
earlier implementation layout. This mapping translates those paths; checked
snapshot tasks still require verification against the current public checkout.

## Validation

Run from the repository root:

```sh
go test -race ./...
go vet ./...
go build ./cmd/...
go test -tags=integration -p 2 ./internal/... ./cmd/...
python3 -m unittest scripts/test_environment_operations.py scripts/test_changelog.py
python3 scripts/check-deployment-config.py
docker build -t iwa-app:local .
```

Integration tests use disposable synthetic databases and object storage. Optional
administrative browser tests retain their existing explicit tool configuration;
moving a template does not establish a new visual acceptance result. No command
above enables an alert source or activates production.
