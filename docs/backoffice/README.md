# Backoffice

`internal/backoffice` contains private administrative HTML/JSON handlers,
presenters and the embedded `ui/` templates, CSS and JavaScript. `cmd/iwa`
composes it when `IWA_ROLE=admin`; configuration and data services remain in the
backend.

The default local address is `http://127.0.0.1:8081/admin/`. The interface starts
with regions and municipalities; Operazioni and Sistema provide cross-territory
tools. The regional table places enabled regions first, then sorts each group by
enabled municipalities in the current register (descending), with alphabetical
ties. It shows enabled/total municipality counts; saved municipality flags remain
visible when a region is disabled, and unavailable counts sort after known ones.
Existing URLs and JSON/form clients remain compatible. Core navigation,
filters and actions work without JavaScript.

`/admin/operations/` shows job counts by state and queue/type, separately counted
archived jobs, recorded last reasons, current provider gates and incomplete
document categories by source. Due time is not proof that admission controls
permit a job; persisted activity is not a worker health check. Links open matching
job filters (state, queue, kind and last error) or source document lists.

The historical bars offer 24 hourly UTC buckets, 7 days or 30 days using the
`period=24h|7d|30d` GET parameter. They count finished queue **attempts**, including
retry/relaunch history and attempts of archived jobs, rather than unique jobs,
documents or reconstructed queue backlog. The current bucket is partial. An
expandable table supplies exact values and the chart works without JavaScript or
external assets. Refresh rereads persisted data; reading the page does not
schedule processing or recover failures.

The October 2 dashboard extension has synthetic PostgreSQL verification for full
counts, repeated attempts, archived history, cutoff boundaries and document
categories, plus browser checks at 390px, 768px and 1440px and the CSS viewport
equivalent to 200% zoom, with JavaScript enabled and disabled. Run its optional
browser check with `IWA_BROWSER_NODE` and `IWA_PLAYWRIGHT_MODULE` pointing to your
local Node/Playwright installation:

```sh
go test -tags=integration ./internal/backoffice -run TestDetailedDashboard -count=1
go test ./internal/backoffice -run TestOperationsOverviewBrowser -count=1
```

`IWA_DASHBOARD_SCREENSHOTS` can select a private screenshot destination. Live
queue observations and deployment/rollback identities stay outside Git; the
checks do not establish source acceptance or public availability.

Use the [developer setup](../development/setup.md) for initialization and the
[environment guide](../operations/environments.md) for selected environment
ports. Remote administration uses the existing SSH tunnel or explicitly
configured tailnet-only origin. Assets and administrative operations remain
private; the public frontend cannot reuse retained administrative views, which
can include data that is not publicly enabled.

Routine actions retain observed revision/attempt checks, explicit operator
attribution and separate collection, acceptance and publication controls.
Presentation does not establish source acceptance.

```sh
go test ./internal/backoffice
go test -tags=integration ./internal/backoffice
```

The public transport's tests additionally verify API/MCP isolation from this
interface. See [architecture](../architecture/README.md) and
[OpenSpec plans](../../openspec/README.md) for component boundaries and behavior.
