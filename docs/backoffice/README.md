# Backoffice

`internal/backoffice` contains private administrative HTML/JSON handlers,
presenters and the embedded `ui/` templates, CSS and JavaScript. `cmd/iwa`
composes it when `IWA_ROLE=admin`; configuration and data services remain in the
backend.

The default local address is `http://127.0.0.1:8081/admin/`. The interface starts
with regions and municipalities; Operazioni and Sistema provide cross-territory
tools. Existing URLs and JSON/form clients remain compatible. Core navigation,
filters and actions work without JavaScript.

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
