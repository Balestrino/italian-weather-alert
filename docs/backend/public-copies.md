# Application-mediated retained copies

Task 6.5 adds one public content route for an exact retained document version:

```text
GET /v1/documents/{document_id}/versions/{version_id}/content
```

The route is disabled by default. Enabling the global switch is necessary but
not sufficient: the document's source must still be publicly enabled and its
**current active configuration** must record `policy.copies_permitted=true`.
That permission already requires evidenced retention and publication policy in
the registry. A current restriction always wins over a link stored in an older
query view.

When authorized, every `Document.copy_url` is the configured public origin plus
the exact document/version path. It is never a RustFS URL, object key, presigned
URL or mutable official-source URL. The content handler repeats current policy
authorization, looks up the original resource for that exact version, reads it
through the internal document store and verifies byte length and SHA-256 before
delivery. Responses use `Content-Disposition: attachment`, `nosniff`, an exact
SHA-256 ETag and `Cache-Control: no-store`. The normal shared per-IP public
budget applies.

Disabled access returns `copy_access_disabled` (403). A source-level or current
publication restriction returns `copy_access_restricted` (403), an unknown
document/version returns `unknown_identifier` (404), and database/object or
integrity failure returns `service_unavailable` (503). Document metadata and
its official link remain available with `copy_url=null` when copies are denied;
interpretation failure does not erase either metadata or retained bytes.

## Configuration and isolation

| Variable | Default | Meaning |
| --- | --- | --- |
| `IWA_PUBLIC_COPY_ACCESS_ENABLED` | `false` | global application gate |
| `IWA_PUBLIC_BASE_URL` | empty | required HTTP(S) origin when enabled |
| `IWA_RUSTFS_*` secret files | mounted privately | read retained objects only after authorization |

The public process does not load object credentials while the global gate is
off. Compose makes the secret files available for an explicit enablement, but
RustFS remains on the internal network with no host port and no public bucket
policy. Source restrictions are data, not deployment toggles: operators must
activate an evidenced configuration with `copies_permitted=true`; the global
switch cannot override a source prohibition.

## Verification

```sh
go test ./internal/backend/publiccopy ./internal/backend/transport/httpapi ./internal/backoffice ./internal/backend/config
go test -tags=integration ./internal/backend/publicquery -run TestFiveSharedPublicQueryGroups -v
go test -race ./...
go vet ./...
docker compose config --quiet
```

Server tests verify exact IDs/bytes, security and cache headers, shared request
accounting, all contract error mappings, and removal of links from an existing
dataset view after policy revocation without rerunning the domain query. The
isolated PostgreSQL/object fixture enables copies for the municipal source but
not the regional source, reads two stable-URL versions by exact ID, and confirms
that the newer uninterpreted version retains its official URL, metadata and
bytes. Default-off and source-restricted reads are denied. Compose verification
continues to assert that RustFS has no published host port and that secret
values do not appear in rendered configuration, logs or responses.
