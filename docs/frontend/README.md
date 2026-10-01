# Public system-status frontend

`internal/frontend` and `cmd/iwa-frontend` provide an independent, server-rendered
status page. It reports public backend availability, source publication state,
declared coverage and updating information separately. It supports public API
cursor pagination and works without JavaScript.

This MVP describes current observations. It is not an uptime history or a browser
for alerts and municipal measures. Empty coverage is not an all-clear result;
pending, suspended, delayed and unavailable information remains visible. API
response time and each source's last complete check have separate labels.

## Run locally

From the repository root, with a public backend on port 8080:

```sh
go run ./cmd/iwa-frontend
```

Open `http://127.0.0.1:8082/`. The page remains usable when that backend is offline.

| Variable | Default | Purpose |
| --- | --- | --- |
| `IWA_PUBLIC_BACKEND_URL` | `http://127.0.0.1:8080` | Public backend HTTP(S) origin, without credentials, path, query or fragment |
| `IWA_FRONTEND_LISTEN` | `127.0.0.1:8082` | Frontend listener |

The frontend requires no IWA database, storage, crawler or provider secrets. It
only reads fixed public endpoints, does not follow redirects and does not proxy
visitor-selected URLs. Page reads have a five-second total deadline, each HTTP
request a four-second timeout and each successful response a 2 MiB size limit.

All coverage reads share the frontend server's upstream IP budget. The backend
keeps enforcing its existing anonymous limits; the page reports rate limiting
explicitly. The initial frontend has no cache, so refreshing or opening a page
performs another coverage read. Limits and caching should be evaluated before a
high-traffic rollout.

The frontend's `/health/live` and `/health/ready` describe its own process.
Backend readiness is reported separately on the page. A backend-process failure
can be displayed; a shared-host outage can still make both services unreachable.

## Optional Compose project

[compose.frontend.yaml](../../deploy/compose.frontend.yaml) is a **standalone**
project. It is not an overlay for the core stack and is not managed by
`scripts/compose-env.sh`. Keeping its own project avoids changing the existing
development/staging/production service inventories and validation scripts.

Prepare an ignored frontend configuration:

```sh
install -m 600 deploy/frontend.env.example .local/frontend.env
docker compose --env-file .local/frontend.env -f deploy/compose.frontend.yaml --profile public-frontend config --quiet
```

The example attaches only to the development listener network `iwa_listeners`.
That external network must already exist. For staging or production select the
corresponding `<project>_listeners` network, an unused loopback frontend port,
a distinct frontend project name, and the tested/approved image. The frontend
mounts no secrets and has no database-network connection or startup dependency
on backend readiness.

After selecting an image that includes `/iwa-frontend`, deliberately start it:

```sh
docker compose --env-file .local/frontend.env -f deploy/compose.frontend.yaml --profile public-frontend up -d --pull never frontend
```

This example publishes the frontend on loopback only. Public routing, TLS and
production activation remain explicit deployment work. See
[release operations](../operations/releases.md). To deploy on a separate host,
use the command with the backend's public HTTPS origin and an appropriate local
listener/proxy instead of attaching to this same-host external network.

## Verification

```sh
go test -race ./internal/frontend
go build -o bin/iwa-frontend ./cmd/iwa-frontend
```

Synthetic tests verify delayed sources alongside healthy service availability,
escaping, separate timestamps, empty versus failed reads, expired cursors,
rate limits, oversized/malformed responses, backend outage, cancellation,
redirect rejection, cursor preservation and private-route isolation.
