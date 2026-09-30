# Consistent public views and cursors

Task 6.4 gives every successful initial public query an immutable PostgreSQL
snapshot. Its random `dataset_version` identifies the snapshot across JSON API
and MCP. The default lifetime is 30 minutes; `view_expires_at` is fixed when the
view is created, while `served_at` changes on every delivery.

The first request resolves default `evaluation_time` and `known_at` before it
queries prepared public data, normalizes the filters and effective page size,
and stores the complete result plus history metadata in migration
`020_public_views`. Pages contain only the configured/requested number of items.
Municipality situations retain an independent position for local measures,
operational phases, regional products, documents requiring attention and
coverage. Documents retain a version position; the other collection operations
retain one data position.

`next_cursor` is a base64url payload authenticated with HMAC-SHA-256. It binds
the view ID, operation, normalized-request hash, collection positions and
expiry. It is an opaque client token, not an MCP session. Changing an explicit
filter or operation, combining it with another dataset version, corrupting the
token or modifying positions yields `cursor_mismatch` (HTTP 400). An expired or
removed view yields `cursor_expired` (HTTP 410) and instructs the caller to
restart. Replaying a valid cursor is idempotent.

A caller can supply `dataset_version` without a cursor to retrieve the first
page of the same snapshot through another interface. Omitted filters inherit
the stored normalized request; explicitly supplied filters must match. This is
how an API result and an MCP result can be compared without rerunning the domain
query or observing intervening acquisitions/interpretations.

Snapshots store public serialized data, not object-store credentials or direct
RustFS locations. PostgreSQL failure while creating or loading a view returns
`service_unavailable`; it never becomes an empty result or all-clear response.
The full snapshot also lets retention remove an underlying version after the
page was created without mixing later data into remaining pages. Task 6.5's
[copy-access policy](public-copies.md) is evaluated separately on every delivery
rather than being granted by an old cursor. Creating a view opportunistically deletes
expired snapshots, so the persistent table does not retain every historical
anonymous request indefinitely.

## Truthful freshness

Facts, evidence, source-check identity, evaluation time and knowledge boundary
stay pinned. On every page delivery, the service recalculates each `ok` or
`delayed` updating state from the pinned `last_complete_check_at`, its recorded
delay threshold and the new `served_at`. Thus an older view can cross its delay
threshold and become `delayed`; it cannot keep describing its check as fresh
merely because the payload was cached. Failed, suspended and unverified states
remain pinned.

## Configuration

| Variable | Default | Meaning |
| --- | ---: | --- |
| `IWA_PUBLIC_VIEW_TTL_SECONDS` | `1800` | view lifetime, 60–86,400 seconds |
| `IWA_PUBLIC_CURSOR_KEY_FILE` | required for public role | private key file, at least 32 bytes |

`scripts/init-secrets.py` creates `public_cursor_key` once and preserves an
existing value. Compose mounts it only into the public container. Rotating the
key intentionally invalidates outstanding cursors; clients restart their
queries. The key is never returned, logged, placed in Compose environment
values or stored with a view.

## Verification

```sh
python3 scripts/init-secrets.py
go test ./internal/server ./internal/publicview ./internal/config
go test -tags=integration ./internal/publicquery -run TestFiveSharedPublicQueryGroups -v
go test -race ./...
go vet ./...
docker compose config --quiet
```

The transport test creates a one-item page, changes the live query fixture and
proves the next page still contains the old snapshot. It also proves API/MCP
equivalence through one `dataset_version` for all five operations, independent
positions, HMAC tamper rejection, changed-filter rejection, explicit expiry and
fresh-query restart, `ok` to `delayed` transition at delivery, and 503 behavior
for create/load failures. The isolated PostgreSQL fixture applies migration 020
twice and round-trips a stored view in addition to the existing five domain
queries.
