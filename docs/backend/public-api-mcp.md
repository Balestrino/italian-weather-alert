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

In development the public host port binds to `0.0.0.0`, so another machine can
open `http://<host-address>:8080/docs` and use the same origin for JSON API and
MCP. Substitute `IWA_PUBLIC_PORT` if configured. Staging/production public ports
and every administrative host port remain on loopback.

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

## Development publication

Explicit development installations can publish manually selected municipalities,
including their applicable regional products, before production acceptance.
`development_publication` on discovery/coverage records and `meta.limitations`
identify this scope; situation source products retain pending acceptance. Revocation
expires saved views and cursors on API and MCP. See [operator controls and limits](../operations/development-publication.md).

## Situation response

`GET /v1/municipalities/{municipality_istat}/situation` and
`get_municipality_situation` return the same `data` shape:

```json
{
  "municipality": {"istat": "050004", "name": "Calcinaia", "zones": ["A4"]},
  "processed_data": {
    "summary": "Conclusions supported by the available sources.",
    "regional_alerts": [],
    "local_measures": [],
    "operational_phases": [],
    "references": []
  },
  "source_summaries": {
    "regional": {"name": "Regione Toscana / CFR", "status": "not_collected", "summary": "No publishable data in this view.", "products": []},
    "municipal": {"name": "Comune di Calcinaia", "status": "not_collected", "summary": "No publishable data in this view.", "products": []},
    "cittadino_informato": {"name": "Cittadino Informato", "status": "not_collected", "summary": "No publishable data in this view.", "products": []}
  }
}
```

This illustrates structure only, not the live Calcinaia contents. The usual `meta`
envelope retains evaluation/view/history/cursor information. Processed facts carry
`status`, `reliability`, compact `validity`, affected `limitations` and
`reference_ids`; references identify exact documents, versions, official URLs and
PDF pages. `supported` describes interpreted evidence, not source acceptance.
An `undetermined` measure is documented but cannot be confirmed in force. Regional
`not_applicable` remains distinct from green and unknown. Conclusions describe
prepared evidence and never generate conduct advice.

Supported vigilance facts in both regional search and `processed_data.regional_alerts`
can carry optional `weather`: `phenomenon`, `graphical_status` (`depicted`,
`not_depicted`, `unresolved`) and `under_evaluation` from the HTML table. Rainfall
also carries literal `rainfall_band`, `total_rainfall_band`, `total_period`,
`unit: "mm"` and `amount_scope: "area_average"`. The total describes the PDF's
cumulative period rather than the daily validity. Both hydrological risk categories
share the rainfall observation. Vigilance warning levels stay `not_applicable`;
`not_depicted` never establishes absence of risk. Older projections can omit
`weather`. See [graphical interpretation and verification](../operations/cfr-graphics.md).

Each source summary states its contents and data availability; individual
products preserve check timestamps and `coverage_status`. `available` identifies
fresh available acquisitions, not complete interpretation or accepted coverage.
`partial` can indicate mixed availability or delayed checks. No acquired/publishable
platform data is `not_collected`, not proof that the platform has no notices.
No isolated trial or private channel contents are inserted into this view.

This replaces the former raw top-level situation lists. Consumers should use
`processed_data.local_measures`, `processed_data.regional_alerts` and
`processed_data.operational_phases`. Full document metadata, verification receipts
and detailed quality/coverage remain in search, document and coverage operations.
The status frontend consumes coverage and does not depend on the changed fields.
See [snapshot pagination](public-views.md) for cursor transition and repeatable
full-scope summaries.
