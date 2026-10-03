# Embedding activation

Use **Sistema → Embedding** in the private admin panel. The global control and
each source choice start disabled, including sources registered before this
change. A source requires both flags before embedding work becomes eligible.
Each action requires an operator and the current control revision; reload after
a revision conflict. Source embedding revisions are independent of collection
configuration revisions.

Disabling the global control preserves source choices. Disabling either level
excludes pending embeddings from new claims without changing their retry budget.
Already admitted work may finish. New extractions use ordinary linking while
embedding is disabled; semantic candidate retrieval also respects these flags.
Enabling does not reprocess completed historical extractions, enable collection,
grant publication or start workers. Existing territorial, provider and recovery
gates still apply. Previously archived jobs stay archived.

Equivalent CLI commands run against an initialized, explicitly selected
environment under the admin role. Replace angle-bracket placeholders with the
source, operator and revision returned by status:

```sh
scripts/compose-env.sh development run --rm --no-deps --pull never admin embedding-status
scripts/compose-env.sh development run --rm --no-deps --pull never admin embedding-enable <revision> <operator>
scripts/compose-env.sh development run --rm --no-deps --pull never admin embedding-disable <revision> <operator>
scripts/compose-env.sh development run --rm --no-deps --pull never admin source-embedding-status <source-id>
scripts/compose-env.sh development run --rm --no-deps --pull never admin source-embedding-enable <source-id> <revision> <operator>
scripts/compose-env.sh development run --rm --no-deps --pull never admin source-embedding-disable <source-id> <revision> <operator>
```

JSON clients can read `GET /admin/embedding` with `Accept: application/json` and
post `{ "enabled": true, "expected_revision": 0, "actor": "operator" }` to
`/admin/embedding` or `/admin/sources/<source-id>/embedding`. Saved enablement
describes policy, not provider or worker health.

The worker's independent provider settings are `IWA_EMBEDDING_ADAPTER`,
`IWA_EMBEDDING_ENDPOINT`, `IWA_EMBEDDING_MODEL`,
`IWA_EMBEDDING_API_KEY_FILE` and `IWA_EMBEDDING_GATE_SCOPE`. Current defaults use
the existing HTTPS Regolo embedding adapter, model and shared key file; the
vector contract remains 1024 dimensions. Settings are loaded lazily after global
and source activation. Changing provider settings requires recreating the chosen
environment's workers. Saved flags apply to running compatible workers without
recreation. The former `IWA_SEMANTIC_LINKING_ENABLED` flag no longer activates
embedding work; remove it from private environment files.

Migration `068_embedding_control` adds separate control/audit tables and keeps
existing jobs and source configuration intact. Replace every processing replica
when adopting this migration: older workers do not enforce its flags. Preserve
database and image/settings evidence privately. Keep workers stopped if rolling
back to an incompatible image; no reverse migration or queue deletion is needed.

Synthetic checks cover default states, both-level admission, preserved retry
budgets, source-choice retention, current scheduling and CLI/admin parity. Live
adoption and any requested backlog archival are separate operations, with details
under ignored `.local/operations/`. These controls do not establish source
acceptance or provider availability.
