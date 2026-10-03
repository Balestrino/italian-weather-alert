# Development, staging and production

IWA uses three Docker Compose projects from one repository directory on one host. Code publication, deployment and alert-source acceptance are separate steps; see the [coverage tracker](../coverage.md).

| Environment | Compose project | Public / admin ports | Settings / secrets | Application image |
| --- | --- | --- | --- | --- |
| Development | `iwa` | `8080` / `8081` | `.local/development.env` / `.local/development/secrets/` | Explicit local builds |
| Staging | `iwa-staging` | `28080` / `28081` | `.local/staging.env` / `.local/staging/secrets/` | Candidate built once by the release helper |
| Production | `iwa-production` | `38080` / `38081` | `.local/production.env` / `.local/production/secrets/` | Operator-approved GHCR digest |

These are example ports; select free ports in each private environment file. The development public listener binds to `0.0.0.0`, exposing JSON API, MCP and documentation on all host IPv4 interfaces at `IWA_PUBLIC_PORT` (default `8080`). Use `http://<host-address>:8080` from another machine or `http://127.0.0.1:8080` locally. Staging and production explicitly publish the public listener on `127.0.0.1`; administration stays on `127.0.0.1` in every environment. Only public/admin HTTP listeners publish host ports. Databases, object storage, crawlers and workers have no host ports. A new installation has no accepted sources or live collection. Production preparation selects no services and creates no containers.

For an existing development stack, apply the API binding without rebuilding its image or restarting dependencies with `scripts/compose-env.sh development up -d --no-deps --no-build --pull never --wait public`. The runtime boundary checks verify the selected environment's host binding and use loopback for local readiness requests. Native public startup with `IWA_ENVIRONMENT=development` also defaults to `0.0.0.0:8080`; an explicit `IWA_LISTEN` takes precedence.

The runtime `IWA_ENVIRONMENT` is pinned to the selected Compose environment; its software default is strict production. Development alone supports [manual municipality publication](development-publication.md) before source acceptance. Staging/production ignore these selections and retain the specified gates.

All projects share CPU, RAM and disk. Their named volumes, networks and credentials are separate, but a host outage affects every environment. Review capacity and off-host recovery before production activation.

## Prerequisites

Use Bash, Python 3.9 or newer, Docker Engine with its Compose plugin, and curl. The environment commands and overlays were tested with Docker Compose **v5.5.1**; verify an older plugin can parse the `!reset` and `!override` overlays using the template check below before proceeding. Git is required for release images; basic setup and management also work from an extracted ZIP. Go 1.27.1 is needed only for native builds and Go tests. Docker pulls/builds need network access and several gigabytes of disk, especially for Crawl4AI.

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

For an empty Docker host, follow the complete [developer setup](../development/setup.md). It initializes development, migrates the database, prepares storage and starts both listeners without an inference-provider key.

## Fresh staging

Use a Git checkout for a release candidate. Staging uses the [staging overlay](../../deploy/compose.staging.yaml), which removes application build directives. Start from a clean identified revision and keep the checkout unchanged through validation and publication.

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

No application build belongs in these staging steps. Test controlled fixtures using [release operations](releases.md), then publish that candidate. Workers and inference remain off unless explicitly configured for the intended check.

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

The `production` profile selects the five core services; `production-worker` and `application-backup` are separate. Profiles control selection: explicitly targeting a service can activate its profile. They do not replace deployment approval. Use the complete [production deployment and rollback procedure](releases.md) after host readiness passes. Keep the application backup worker disabled while off-host PBS protects the whole VM.

## Background service selection

A default development/staging `up` selects only public, admin, postgres, rustfs and crawl4ai. Processing is opt-in with `processing-worker`; backup is separately opt-in with `application-backup`. Production's default selects nothing, `production` selects those five core services, and its worker uses only `production-worker`. The development/staging profile cannot select the production worker.

After initializing the intended environment, use a bounded, source-scoped check when needed:

```sh
scripts/compose-env.sh staging --profile processing-worker up -d --no-build --pull never worker
scripts/compose-env.sh staging logs --tail=100 worker
scripts/compose-env.sh staging stop worker
scripts/compose-env.sh staging ps --all worker
```

Backup has its own selection, without selecting processing:

```sh
scripts/compose-env.sh staging --profile application-backup config --services
```

An explicit service target such as `scripts/compose-env.sh staging up -d --no-build --pull never worker` also activates a profiled worker; omitting the profile is not protection when naming that target. Profiles govern Compose selection, not runtime admission or cross-environment deduplication. They do not stop existing workers: `restart: unless-stopped` can recover a previously activated worker after a Docker/host restart. End a test with `stop worker` and confirm its state using `ps --all worker`. The processing worker also handles configured notifications; stopping it stops those loops too.

## Embedding activation

Use **Sistema → Embedding** or the [equivalent CLI controls](embedding-controls.md).
Both global and per-source flags default to disabled, including existing sources.
Enabling one level alone does not permit embedding work. Flags apply to all
compatible processing replicas in the selected environment; provider settings
remain separate.

## Territorial and source activation

The selected environment owns its registry and flags. Automatic local-source processing requires **region enabled → municipality enabled → source collection enabled → worker active**. A regional source requires its region and supported profile, not a municipality flag. None of these choices grants public publication. A saved source flag does not demonstrate worker liveness.

In the admin panel, use `/admin/regions` to enable/disable a region. Open **Configura regione → Comuni** to enable/disable individual municipalities, including those without sources. Regional setup supplies the complete adopted register and supported processing profiles. Configure a municipality's sources in its configuration tab; acquire a deliberate preview, then enable collection using the existing source form. Source suspension and publication remain separate controls. Native forms require an operator identity and expected revision; after a conflict reload before retrying. The municipality list shows saved state and territorial blocking reason, retaining filters/pagination after actions. History records the operator and revision.

The equivalent CLI commands run on the admin service against an already migrated environment. The following examples explicitly target development; substitute actual region/ISTAT/source identifiers and the revisions returned by status, rather than assuming the example revision is current. Source enablement uses the intended configuration revision; suspension uses its active revision. Territorial mutations increment their own revision. Each status/action returns JSON; rejected arguments, missing prerequisites or stale revisions return nonzero with a stable, private-detail-free error code.

```sh
scripts/compose-env.sh development run --rm --no-deps --pull never admin region-status 09
scripts/compose-env.sh development run --rm --no-deps --pull never admin region-enable 09 <region-revision> <operator>
scripts/compose-env.sh development run --rm --no-deps --pull never admin region-disable 09 <region-revision> <operator>
scripts/compose-env.sh development run --rm --no-deps --pull never admin municipality-status 09 050004
scripts/compose-env.sh development run --rm --no-deps --pull never admin municipality-enable 09 050004 <municipality-revision> <operator>
scripts/compose-env.sh development run --rm --no-deps --pull never admin municipality-disable 09 050004 <municipality-revision> <operator>
scripts/compose-env.sh development run --rm --no-deps --pull never admin source-status <source-id>
scripts/compose-env.sh development run --rm --no-deps --pull never admin source-enable-collection <source-id> <configuration-revision> <operator>
scripts/compose-env.sh development run --rm --no-deps --pull never admin source-suspend-collection <source-id> <active-revision> <operator>
```

Angle-bracket values are documentation placeholders; replace them before executing. These commands save or inspect configuration without starting the worker, previewing sources, calling LLM/OCR or activating publication. They reuse the same audited domain/registry operations as the panel. Source activation still requires preview evidence, collection permission and compatible territorial configuration. A saved child/source choice can be changed while the parent is disabled; execution remains blocked until all gates permit it. To acquire a manual preview, enable its territory first and use the separate preview action.

Use the explicit environment on every command. For initialized staging or approved production, status examples are:

```sh
scripts/compose-env.sh staging run --rm --no-deps --pull never admin municipality-status 09 050004
scripts/compose-env.sh production --profile production run --rm --no-deps --pull never admin source-status <source-id>
```

Disabling a parent gates new acquisition and processing, including queued and manual work. Already admitted work may finish. It preserves child/source flags and historical results; re-enabling restores normal eligibility without enqueueing a catch-up batch. A municipality flag gates its local sources, not shared regional bulletin collection or retained regional warnings. New register identities start disabled; unchanged identities preserve choices across register replacement, and retired identities remain readable in history.

## Adoption and collector handover

Repository changes do not modify live containers. Inspect each existing collector before any rollout:

```sh
scripts/compose-env.sh development ps --all worker backup
scripts/compose-env.sh staging ps --all worker backup
scripts/compose-env.sh production ps --all worker backup
```

Keep the current collector running until a production handover is explicitly scheduled and the [release prerequisites](releases.md) pass. The eventual continuous collector belongs only in production; development/staging processing is for controlled checks that end with an explicit stop. At the scheduled handover, stop the former collector and confirm it is stopped before activating the reviewed production worker under the release procedure. This change provides no host-wide lease: intentionally activating multiple environments still duplicates resource use.

The additive municipal migration preserves prior eligibility for existing municipalities with configured local sources, attributing that backfill to migration. It does not change source/public flags, regional choices, immutable geography or existing migration checksums. Existing explicit municipality choices are never overwritten on a repeated migration. For a new database/register, municipalities default to disabled.

Keep additive state/events on rollback. A prior image that lacks the municipality execution gate is incompatible with active processing after operators have disabled municipalities: keep workers stopped until a compatible image is selected. Restoring old Compose profiles can also default-enable optional services; inspect resolved selection before startup. An image/configuration rollback does not automatically restart a deliberately stopped worker and never requires a reverse migration or deleting volumes.

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

For source-specific collection failures and persistent interpretation limits,
follow [development acquisition recovery](acquisition-recovery.md).

Ordinary smoke checks are nondisruptive. Fault injection stops and restarts only the selected project's RustFS and needs a planned interruption on a throwaway or explicitly scheduled development/staging stack:

```sh
python3 scripts/smoke-compose.py staging --allow-interruption
```

The check verifies runtime project/port/mount identity before stopping anything, attempts recovery even if an assertion fails, and refuses production fault injection. Schedule it explicitly when staging is operational.

## Data and credential boundaries

Generate separate credentials for fresh projects. New files use mode `0644` inside a host directory with mode `0700`, independent of umask, so the container users can read their bind-mounted secrets. Existing credentials and custom permissions are preserved. A private provider key must be readable by its container UID (65532 for the application); repair its mode/ACL deliberately rather than regenerating it. Runtime checks test mounted-file readability without printing values.

Never point development and staging to the same database or RustFS storage. If staging needs a private representative copy, transfer PostgreSQL and RustFS together from a consistent point, and keep both background services stopped, provider keys absent and notifications disabled before using it. The compatibility migration deliberately preserves configured local-source eligibility, so it is not a copied-data safeguard. Disable copied regions and review municipality/source choices through the panel or environment-scoped CLI before any controlled worker activation. Code promotion does not copy data or activate sources. Keep private files, operational evidence and unpublished captures outside commits.
