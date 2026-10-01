# Public API and MCP usage

Task 6.6 publishes one machine-readable set of synthetic examples in
[`public-usage-examples.json`](public-usage-examples.json). The examples cover
all five public query groups and are executed unchanged against a local HTTP
service by `TestPublishedUsageExamplesAgainstLocalService`. They are transport
examples, not current alert information.

## JSON API

Set the local public origin and call the five read-only routes:

```sh
PUBLIC_ORIGIN=http://127.0.0.1:8080

curl --fail-with-body "$PUBLIC_ORIGIN/v1/municipalities?istat=050004&evaluation_time=2026-09-17T12%3A00%3A00Z&known_at=2026-09-17T11%3A30%3A00Z"
curl --fail-with-body "$PUBLIC_ORIGIN/v1/municipalities/050004/situation?evaluation_time=2026-09-17T12%3A00%3A00Z&known_at=2026-09-17T11%3A30%3A00Z"
curl --fail-with-body "$PUBLIC_ORIGIN/v1/search?kind=regional&product=vigilance&evaluation_time=2026-09-17T12%3A00%3A00Z&known_at=2026-09-17T11%3A30%3A00Z"
curl --fail-with-body "$PUBLIC_ORIGIN/v1/documents/7?include_versions=true&evaluation_time=2026-09-17T12%3A00%3A00Z&known_at=2026-09-17T11%3A30%3A00Z"
curl --fail-with-body "$PUBLIC_ORIGIN/v1/sources/coverage?municipality_istat=050004&evaluation_time=2026-09-17T12%3A00%3A00Z&known_at=2026-09-17T11%3A30%3A00Z"
```

Document ID `7` exists only in the automated fixture. Replace it with an ID
returned by search in a populated local environment. `evaluation_time` chooses
the situation being evaluated; `known_at` prevents facts learned later from
appearing retroactively. They are fixed here so API and MCP results can be
compared. An omitted pair uses the service's current boundary.

Every success has `data` and `meta`. Preserve `meta.dataset_version` when
moving between API and MCP or requesting another page of the same view. Follow
`meta.next_cursor` without changing filters; an expired cursor requires a new
query. `history_start`, `history_gaps`, and `limitations` describe retained
service history and must not be interpreted as a complete pre-service archive.

## MCP

Connect an MCP client to `http://127.0.0.1:8080/mcp` using stateless Streamable
HTTP and protocol `2026-07-28`. Tool discovery returns exactly these calls; the
arguments below are the MCP half of the same published examples:

```json
{"name":"discover_municipalities","arguments":{"istat":"050004","evaluation_time":"2026-09-17T12:00:00Z","known_at":"2026-09-17T11:30:00Z"}}
{"name":"get_municipality_situation","arguments":{"municipality_istat":"050004","evaluation_time":"2026-09-17T12:00:00Z","known_at":"2026-09-17T11:30:00Z"}}
{"name":"search_alerts","arguments":{"kind":"regional","product":"vigilance","evaluation_time":"2026-09-17T12:00:00Z","known_at":"2026-09-17T11:30:00Z"}}
{"name":"get_document","arguments":{"document_id":"7","include_versions":true,"evaluation_time":"2026-09-17T12:00:00Z","known_at":"2026-09-17T11:30:00Z"}}
{"name":"get_source_coverage","arguments":{"municipality_istat":"050004","evaluation_time":"2026-09-17T12:00:00Z","known_at":"2026-09-17T11:30:00Z"}}
```

For an exact cross-interface comparison, first call the API example and add its
`meta.dataset_version` to the corresponding MCP `arguments`. MCP application
failures set `isError=true` and retain the same structured error envelope as
the API. Rate accounting is returned in `_meta["iwa.dev/rateLimit"]`; API and
MCP requests consume the same per-IP budget.

## Consumer boundaries

The public listener exposes observations, evidence, official risk labels,
quality dimensions, coverage limitations, and official document links. It
does not expose administration or configuration mutation, internal DPC
comparison records, secrets, collection/inference controls, or generated
safety advice and recommendations. A regional level or municipal measure is a
source-attributed fact, not a service-generated instruction. Consumers should
follow the linked issuing authority for official conduct guidance.

The default-off retained-copy route is documented separately in
[`public-copies.md`](public-copies.md). It is not an additional query group and
never substitutes the official URL.

## Verification

Run from the repository root:

```sh
go test ./internal/backend/transport/httpapi ./internal/backoffice -run TestPublishedUsageExamplesAgainstLocalService -v
go test -tags=integration ./internal/backend/publicquery -run TestFiveSharedPublicQueryGroups -v
```

The first command starts a local in-process HTTP service, connects with the
official MCP Go client, executes every API path and MCP tool argument object
from the JSON artifact, and compares the resulting shared view. It also checks
the exact five-tool surface, absent administrative routes, and absence of
generated-safety fields. The isolated PostgreSQL test retains an internal DPC
comparison document, proves that its source cannot be publicly enabled, and
checks that neither public search nor coverage returns it.
