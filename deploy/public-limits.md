# Shared public request limits

Task 6.3 places one exact sliding-window budget in front of every registered
public JSON API and MCP request. The initial policy is **120 requests per 60
seconds per canonical client IP**, with a **maximum page size of 100 items**.
Requests do not have token-dependent cost: each admitted public HTTP request is
one unit, including malformed operation parameters, MCP discovery/list/call,
and protocol requests that later fail validation. Health and the separate
administrative listener are outside this public budget.

API and MCP traffic from the same address uses the same in-memory ledger.
People behind one NAT therefore share its allowance. A rejected request does
not consume another unit, returns HTTP 429, includes `Retry-After`, and carries
the contract's `rate_limited` error with the same retry seconds. The response
also publishes the current policy and accounting using the May 2026 IETF
working-draft fields:

```text
RateLimit-Policy: "iwa-public";q=120;w=60
RateLimit: "iwa-public";r=119;t=60
X-IWA-Max-Page-Size: 100
```

`RateLimit-Policy` and `RateLimit` follow
[draft-ietf-httpapi-ratelimit-headers-11](https://datatracker.ietf.org/doc/draft-ietf-httpapi-ratelimit-headers/),
which remains a work in progress rather than a published RFC. MCP discovery
instructions publish the allowance, window, shared-IP behavior and maximum page
size. Successful tool calls repeat limit/remaining/reset/page-size accounting
in `_meta["iwa.dev/rateLimit"]`.

The maximum is enforced even when no `page_size` is supplied. Task 6.4 now
provides [fixed-view pagination](public-views.md): omitted `page_size` uses the
configured maximum, and explicit values from 1 through that maximum are
accepted. Values above it remain `invalid_parameters`.

## Configuration and proxy trust

The public container exposes these settings:

| Variable | Default | Valid range / meaning |
| --- | ---: | --- |
| `IWA_PUBLIC_RATE_ALLOWANCE` | `120` | 1–1,000,000 admitted requests |
| `IWA_PUBLIC_RATE_WINDOW_SECONDS` | `60` | 1–86,400 seconds |
| `IWA_PUBLIC_MAX_PAGE_SIZE` | `100` | 1–1,000 items |
| `IWA_PUBLIC_TRUSTED_PROXIES` | empty | comma-separated IPs or CIDR prefixes |

The socket peer address is authoritative by default. `X-Forwarded-For` is
ignored from every untrusted peer. When the immediate peer is explicitly
trusted, the service walks the forwarded chain from right to left through
trusted hops and uses the first untrusted address. A malformed chain safely
falls back to the socket peer. Configure only proxy networks under the same
operator's control; never enter a general client range merely to obtain an
original address.

The ledger is process-local, matching the single public process in the MVP
Compose topology. Deploying multiple public replicas requires a shared limiter
or deterministic IP affinity before claiming one cross-replica allowance.

## Initial load measurement

The initial numbers were set after the reproducible transport-bound benchmark:

```sh
go test ./internal/server -run '^$' \
  -bench '^BenchmarkPublicCoverageHTTP$' -benchtime=3s -benchmem -count=3
```

On the current six-thread Intel Xeon Gold 6130 environment, a 100-item coverage
response was 42,350–42,352 bytes. Three runs measured 233,845–245,703 ns/op
(about 4,070–4,276 responses/second at benchmark concurrency), 122,472–122,928
bytes allocated and 354 allocations per request. This measures HTTP envelope,
shared limiter and JSON serialization with a prepared in-memory query result;
it does **not** claim PostgreSQL production capacity or source freshness.

The 100-item cap is the largest measured response shape. The 120-per-60-second
allowance permits a client IP an average of two admitted requests per second,
over 2,000 times below this transport-only throughput, leaving deliberate room
for database work, shared NAT users and operational variance. These are initial
MVP controls, not an SLA; remeasure with the deployed database and representative
data before increasing them.

## Verification

```sh
go test ./internal/config ./internal/server
go test -race ./...
go vet ./...
docker compose config --quiet
```

Tests verify exact sliding expiry and retry values, API-to-MCP enforcement,
invalid-request accounting, same-IP behavior across source ports, independent
budgets for distinct forwarded clients behind a trusted proxy, spoof resistance
from untrusted peers, malformed-chain fallback, published HTTP/MCP accounting,
configuration bounds, and rejection of responses or requested pages above the
configured maximum.
