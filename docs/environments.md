# Development, staging and production

IWA uses three Docker Compose projects from one repository directory on one host. Code publication, deployment and alert-source acceptance are separate steps; see the [coverage tracker](coverage.md).

| Environment | Compose project | Public / admin ports | Settings / secrets | Application image |
| --- | --- | --- | --- | --- |
| Development | `iwa` | `8080` / `8081` | `.local/development.env` / `.local/development/secrets/` | Explicit local builds |
| Staging | `iwa-staging` | `28080` / `28081` | `.local/staging.env` / `.local/staging/secrets/` | Candidate built once by the release helper |
| Production | `iwa-production` | `38080` / `38081` | `.local/production.env` / `.local/production/secrets/` | Operator-approved GHCR digest |

These are example ports; select free loopback ports in each private environment file. Only public/admin HTTP listeners publish host ports, bound to `127.0.0.1`. Databases, object storage, crawlers and workers have no host ports. A new installation has no accepted sources or live collection. Production preparation selects no services and creates no containers.

All projects share CPU, RAM and disk. Their named volumes, networks and credentials are separate, but a host outage affects every environment. Review capacity and off-host recovery before production activation.

## Prerequisites

Use Bash, Python 3.9 or newer, Docker Engine with its Compose plugin, and curl. The environment commands and overlays were tested with Docker Compose **v5.5.1**; verify an older plugin can parse the `!reset` overlays using the template check below before proceeding. Git is required for release images; basic setup and management also work from an extracted ZIP. Go 1.27.1 is needed only for native builds and Go tests. Docker pulls/builds need network access and several gigabytes of disk, especially for Crawl4AI.

Run commands from the repository root unless a command uses an absolute script path. The wrapper itself also works from another working directory.

```sh
docker compose version
python3 scripts/check-deployment-config.py
```

The default checker validates committed **templates**, without starting services or certifying host readiness.

## Existing installation: preserve its state

If `iwa` or `iwa-staging` already exists on this Docker host, adopt its settings and credentials before following fresh setup. Check `docker compose ls` and `docker volume ls`. Never initialize new credentials against existing database volumes or change the `iwa` project name to relabel development.

Keep the existing PostgreSQL password, RustFS keys, crawl token, cursor key and any provider/notification configuration in the appropriate ignored `.local/<environment>/secrets/` directory. Copy current private settings into `.local/<environment>.env`; inspect paths and selected images before startup. Do not blindly replace existing environment files with examples. Existing custom file permissions/ACLs are preserved by secret initialization.

The fixed volume names are `iwa_postgres_data` / `iwa_rustfs_data`, `iwa-staging_postgres_data` / `iwa-staging_rustfs_data`, and `iwa-production_postgres_data` / `iwa-production_rustfs_data`. Routine stop, upgrade and rollback keep them. Do not use `down -v` as routine maintenance.

## Fresh development

For an empty Docker host, follow the complete [developer setup](developer-setup.md). It initializes development, migrates the database, prepares storage and starts both listeners without an inference-provider key.

## Fresh staging

Use a Git checkout for a release candidate. Staging uses the [staging overlay](../deploy/compose.staging.yaml), which removes application build directives. Start from a clean identified revision and keep the checkout unchanged through validation and publication.

```sh
mkdir -p .local/staging
install -m 600 deploy/staging.env.example .local/staging.env
python3 scripts/init-secrets.py --directory .local/staging/secrets
scripts/release-image.sh build
```

Set `IWA_APP_IMAGE` in `.local/staging.env` to the exact tag printed by the build helper. Then run:

```sh
scripts/compose-env.sh staging config --quiet
scripts/compose-env.sh staging up -d --wait --wait-timeout 180 postgres rustfs crawl4ai
scripts/compose-env.sh staging run --rm --no-deps --pull never admin migrate
scripts/compose-env.sh staging run --rm --no-deps --pull never admin storage-init
scripts/compose-env.sh staging up -d --no-build --pull never --wait --wait-timeout 180 public admin
python3 scripts/smoke-compose.py staging
python3 scripts/check-deployment-config.py --environment staging --runtime
curl --fail http://127.0.0.1:28080/health/ready
curl --fail http://127.0.0.1:28081/admin/status
```

No application build belongs in these staging steps. Test controlled fixtures using [release operations](../deploy/release-operations.md), then publish that candidate. Workers and inference remain off unless explicitly configured for the intended check.

## Prepare production without activating it

For a fresh empty production project:

```sh
mkdir -p .local/production
install -m 600 deploy/production.env.example .local/production.env
python3 scripts/init-secrets.py --directory .local/production/secrets
scripts/compose-env.sh production config --quiet
scripts/compose-env.sh production ps --all
python3 scripts/check-deployment-config.py --environment production --runtime
```

The example image remains a placeholder until an approved release is published. Configuration, status, logs and stop/teardown commands remain available with that placeholder. Production image operations require `ghcr.io/...@sha256:` followed by 64 hexadecimal characters, and application builds are rejected. Digest syntax alone does not prove publication or operator approval.

The `production` profile selects the five core services; `production-worker` and `application-backup` are separate. Profiles control selection: explicitly targeting a service can activate its profile. They do not replace deployment approval. Use the complete [production deployment and rollback procedure](../deploy/release-operations.md) after host readiness passes. Keep the application backup worker disabled while off-host PBS protects the whole VM.

## Routine management

Always select the environment explicitly. For example, these commands target staging:

```sh
scripts/compose-env.sh staging ps --all
scripts/compose-env.sh staging logs --tail=100 public admin postgres rustfs crawl4ai
scripts/compose-env.sh staging logs --follow --tail=100 admin
python3 scripts/check-deployment-config.py --environment staging
python3 scripts/check-deployment-config.py --environment staging --runtime
python3 scripts/smoke-compose.py staging
scripts/compose-env.sh staging stop
scripts/compose-env.sh staging start --wait --wait-timeout 180 postgres rustfs crawl4ai public admin
```

Use `development` for development. For activated production use `scripts/compose-env.sh production --profile production <command>` for core-service stop/start, and explicitly include `--profile production-worker` when managing an activated worker. `start` restarts existing containers; fresh startup and changed configuration use the initialization/deployment sequences, not a bare `up`. `down` removes containers/networks but retains named data volumes.

The wrapper clears inherited `IWA_` / `COMPOSE_` variables and rejects project, configuration-file, environment-file and project-directory overrides. Put settings in the chosen environment file. Use `--profile`, `--dry-run`, `--ansi`, `--progress` and `--parallel` as supported leading options. Staging/production builds and watch/rebuild operations are rejected. Using Docker directly remains an operator capability, so avoid bypassing these guardrails in routine work.

Runtime checks show configured images, running revision labels and checkout revisions. Differences are visible; checks never automatically rebuild or redeploy. Core services must be healthy and mounted to the selected project's data and secrets. Optional inactive workers and absent unused provider keys are allowed. Runtime restart-policy differences require a planned rollout; editing Compose alone does not change existing containers. `unless-stopped` restarts activated services after an unexpected exit or host restart, while intentionally stopped services stay stopped.

## Planned failure checks

Ordinary smoke checks are nondisruptive. Fault injection stops and restarts only the selected project's RustFS and needs a planned interruption on a throwaway or explicitly scheduled development/staging stack:

```sh
python3 scripts/smoke-compose.py staging --allow-interruption
```

The check verifies runtime project/port/mount identity before stopping anything, attempts recovery even if an assertion fails, and refuses production fault injection. Schedule it explicitly when staging is operational.

## Data and credential boundaries

Generate separate credentials for fresh projects. New files use mode `0644` inside a host directory with mode `0700`, independent of umask, so the container users can read their bind-mounted secrets. Existing credentials and custom permissions are preserved. A private provider key must be readable by its container UID (65532 for the application); repair its mode/ACL deliberately rather than regenerating it. Runtime checks test mounted-file readability without printing values.

Never point development and staging to the same database or RustFS storage. If staging needs a private representative copy, transfer PostgreSQL and RustFS together from a consistent point, and disable collection, inference and notifications before using it. Code promotion does not copy data or activate sources. Keep private files, operational evidence and unpublished captures outside commits.
