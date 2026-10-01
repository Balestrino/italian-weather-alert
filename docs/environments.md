# Development, staging and production

IWA publishes its source code publicly. A public code release does not by itself mean that a public API/MCP service or a particular alert source is active. See the [coverage tracker](coverage.md) for documented source status and the [release operations guide](../deploy/release-operations.md) for the deployment gate.

| Environment | Location and purpose | Data and access |
| --- | --- | --- |
| Development | Existing `iwa` Compose project on the current VM, managed from the shared public checkout; day-to-day code and source work. | Keeps its current PostgreSQL/RustFS volumes and private secrets. HTTP listeners bind to localhost; operator access may use a restricted tailnet. |
| Staging | Separate `iwa-staging` Compose project from the same checkout on the current VM; build and test a release candidate. | Independent PostgreSQL/RustFS volumes, secrets, image reference and loopback ports. Initially empty; use controlled fixtures for release checks. No live collection or paid inference is enabled by setup. |
| Production | Planned dedicated VM; public API/MCP behind the HTTPS reverse proxy and private administration through the configured operator path. | Independent data and secrets, off-host PBS protection, and only an operator-approved image digest. Public source activation remains a separate decision. |

Development and staging share host CPU, memory and disk even after their data is separated. Check capacity before starting both and avoid heavy development jobs during a release check. Never change the existing development Compose project name merely to relabel it: named volumes are scoped by the project name.

## Configure development and staging from one checkout

Use one checkout on the current VM and keep each environment's ignored settings and secrets in its own directory. The [Compose file](../compose.yaml) selects the secret directory with `IWA_SECRETS_DIR`; [secret initialization](../scripts/init-secrets.py) accepts `--directory`. Never share the PostgreSQL password or object-store keys between the projects. The wrapper requires an explicit environment, removes inherited `IWA_` and `COMPOSE_` overrides, and fixes the Compose project name.

```sh
mkdir -p .local/development .local/staging
cp deploy/development.env.example .local/development.env
cp deploy/staging.env.example .local/staging.env
scripts/compose-env.sh development config --quiet
scripts/compose-env.sh staging config --quiet
```

For an existing installation, copy its current private secrets and environment settings into the appropriate ignored `.local/` paths **before running the configuration checks above**; do not regenerate credentials for existing volumes. For a fresh empty project, run `python3 scripts/init-secrets.py --directory .local/development/secrets` or the corresponding staging path before its first start. Preserve the development project name `iwa` so its named volumes remain attached. Choose free loopback ports if the examples are occupied. `iwa-staging` keeps separate networks and named volumes. Set staging's `IWA_APP_IMAGE` to the exact tag printed by `scripts/release-image.sh build` before starting application services. The [developer setup guide](developer-setup.md) supplies the dependency, migration, storage initialization and readiness commands; run them through `scripts/compose-env.sh staging`. Keep `worker` stopped unless a specific release check requires it and an operator has configured a controlled source and provider key. A new staging database has no accepted sources, so empty results are expected.

Verify the two installations independently with `scripts/compose-env.sh development ps` and `scripts/compose-env.sh staging ps`, their project names, host ports and volume names. Staging should have its own `iwa-staging_postgres_data` and `iwa-staging_rustfs_data` volumes; the existing development volume names and contents must remain untouched. Do not run `docker compose down -v` on either installation as part of routine release work.

## Data boundaries

Use synthetic or otherwise authorized fixtures in the public repository. Development and staging must not point to the same live database or RustFS bucket/volume. If staging needs a representative data copy, transfer PostgreSQL and RustFS together from one consistent point in time, keep the copy private, and disable collection, outbound inference and notifications until its configuration has been reviewed. A database-only copy can leave evidence references broken.

Production starts with its own state. Code promotion never copies development or staging data into production. Source configuration, source acceptance, collection activation and public enablement require their own operator review.
