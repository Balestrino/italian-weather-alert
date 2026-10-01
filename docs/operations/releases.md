# Software releases, images and production recovery

This is the public procedure for preparing and checking an IWA release. The production Compose project is configured on the same VM as development and staging but has not been started; PBS backup and public source readiness remain open. Keep actual hostnames, credentials, backup records and approval evidence in private operational records. The [environment guide](environments.md) describes all three projects.

## Versioned software releases

The first software release is [v0.1.0](../releases/v0.1.0.md), an internal-pilot baseline. A Git tag and GitHub release identify published source code. Image publication and production activation follow the separate checks below.

Prepare software releases on `dev`: move the dated entries being released from `CHANGELOG.md` into a version/date section, keep exactly one `## Unreleased` section for future changes, and add a new dated entry describing the release preparation. Keep the release notes concise, covering the available features, installation references and known limits. Preserve historical entries and keep private deployment evidence out of the notes.

Run the changelog check against `origin/main`, and promote `dev` through a pull request after CI, Coverage and Security pass for its current head. Use a merge commit to retain the per-commit history. Wait for the required workflows on the merged `main` revision, then create an annotated `v<version>` tag on that exact revision and publish the reviewed notes as a GitHub release. Align `dev` with the merged `main` history before continuing development. A version below 1.0 remains subject to interface changes; notes must explicitly state the source and service readiness limits.

## Image identity and manual approval

Use one clean Git revision in the shared checkout. Run CI-equivalent tests and the checks relevant to the change. Do not change the checkout between build, staging validation and publish. Build once with:

```sh
scripts/release-image.sh build
```

The helper prints `ghcr.io/balestrino/italian-weather-alert:<full-commit-sha>`. Set `IWA_APP_IMAGE` in staging's ignored `.local/staging.env` to that exact tag. Follow the exact [staging initialization commands](environments.md#fresh-staging), which reuse that candidate for migrations, storage initialization and listener startup with rebuilding disabled. Keep the worker off unless the change requires a controlled source or inference check. Development/staging default startup excludes processing and backup; a test uses `processing-worker` explicitly and ends with `stop worker` plus `ps --all worker`. See [background selection and activation](environments.md#background-service-selection).

Every release check includes:

1. Passing Go test, race, vet, build, integration coverage and vulnerability workflows for the intended revision.
2. Successful `scripts/compose-env.sh staging config --quiet`, staging migration/storage initialization, healthy dependencies/listeners, `python3 scripts/smoke-compose.py staging`, and `python3 scripts/check-deployment-config.py --environment staging --runtime`.
3. Representative JSON API and MCP queries against known staging fixtures, truthful coverage/updating state, and absence of administration on the public listener. Check the administrative listener only from its restricted path.
4. Review of logs, failed jobs, source status and any change-specific source or inference regression. A code pass does not accept or publicly enable a source.

The ordinary [Compose smoke check](../../scripts/smoke-compose.py) changes no services. Its `staging --allow-interruption` option stops/restarts staging RustFS and belongs only on a throwaway stack or a planned staging interruption; production interruption is rejected.

After the staging checks, authenticate to GHCR using private credentials and run `scripts/release-image.sh publish`. Publication first compares the configured and running staging application image IDs/revision labels with the clean-revision candidate and refuses a mismatch before pushing. It then pushes the tested local image and prints a `ghcr.io/...@sha256:...` reference. The first GHCR package is private by default; an operator must explicitly make the image public before using anonymous production pulls. Record the Git revision, staging results, published digest, prior production digest and operator approval together. A public image contains application code, never deployment secrets. Keep at least the previous compatible digest available for rollback.

## Prepared production project and resource limits

The prepared project uses ignored `.local/production.env` and `.local/production/secrets` in the shared checkout. [The environment wrapper](../../scripts/compose-env.sh) fixes the `iwa-production` project name and adds [compose.production.yaml](../../deploy/compose.production.yaml), which removes local build directives. The example reserves loopback ports `38080` and `38081`; the proxy and private operator route have not been connected to them. The image remains a placeholder until an approved digest replaces it. `config` and `ps --all` are safe preparation checks; neither starts containers.

Every production service has a profile, so an unqualified `up` selects no service after valid image selection. An explicitly named service can activate its own profile; profiles do not constitute deployment approval. The `production` profile contains the listeners, PostgreSQL, RustFS and Crawl4AI. The `production-worker` profile is separate, and the `application-backup` profile is not used while PBS protects the VM. The initial per-container ceilings are:

| Service | RAM | CPU |
| --- | ---: | ---: |
| Public, admin | 256 MiB each | 0.5 each |
| PostgreSQL, RustFS | 768 MiB each | 1.0 each |
| Crawl4AI | 1536 MiB | 1.5 |
| Worker, when separately enabled | 768 MiB | 1.5 |
| Application backup, normally disabled | 512 MiB | 0.5 |

The five core services can reach **3.5 GiB** in aggregate; the worker adds **0.75 GiB**. These are ceilings, not measured requirements. Named volumes have no disk quota, and a container may fail if its limit is too low. Rehearse the expected workload in staging, measure host headroom and disk growth, and adjust the limits before approving production startup. Account for all environments, the OS, image-build peaks and database/object growth. Review sustained RAM or disk use above the agreed 70% threshold. Choose host expansion, adjusted limits/retention or workload scheduling from measurements before activation; a template budget alone cannot show sufficient headroom.

## Controlled fixture procedure

Use only synthetic or authorized fixtures. Keep staging workers off, leave provider keys absent, and keep notifications disabled during fixture validation. The repository already provides a reproducible synthetic transport corpus and an isolated PostgreSQL fixture covering all five query groups:

```sh
go test ./internal/backend/transport/httpapi ./internal/backoffice -run TestPublishedUsageExamplesAgainstLocalService -v
go test -tags=integration ./internal/backend/publicquery -run TestFiveSharedPublicQueryGroups -v
```

These tests create their own fixture service/database; they do not populate or modify the running staging database. They test the examples in [public-usage-examples.json](../backend/public-usage-examples.json), including retained versions, municipal measures, regional facts, coverage and public/admin separation. Record their result alongside the exact candidate's runtime checks. On fresh staging, additionally verify empty coverage/search are truthful and discover the expected MCP tool surface:

```sh
curl --fail http://127.0.0.1:28080/v1/sources/coverage
curl --fail 'http://127.0.0.1:28080/v1/search?kind=document'
curl --fail -H 'Content-Type: application/json' -H 'Accept: application/json, text/event-stream' -H 'MCP-Protocol-Version: 2026-07-28' --data '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' http://127.0.0.1:28080/mcp
```

For release checks requiring populated staging, prepare a consistent private synthetic database/object fixture through the documented administration and retention interfaces. Record its known records, version IDs, object hashes, expected query results and limitations. Transfer database and objects together; keep collection/inference/notifications disabled. Run the [five API/MCP examples](../backend/public-usage.md) against the selected staging ports, replacing automated fixture IDs/times with those in the reviewed fixture. Do not label automated fixture acceptance as acceptance of a real alert source. An empty staging query alone is insufficient evidence for a data-processing release.

## Production deployment

Before deployment, select the **approved digest** in `.local/production.env` and set proxy trust, public origin and private administration origin for the reviewed route. Digest syntax is enforced before image-consuming operations; local diagnostics with the preparation placeholder remain usable. Builds are rejected. Record the prior digest/configuration, a recent recoverable PBS backup and working backup alerts, and migration compatibility with the prior image. Complete concurrent-load, restore and routing readiness checks before approving initial activation.

After approval, initial startup is:

```sh
scripts/compose-env.sh production --profile production config --quiet
scripts/compose-env.sh production --profile production pull public admin postgres rustfs crawl4ai
scripts/compose-env.sh production --profile production up -d --no-build --wait --wait-timeout 180 postgres rustfs crawl4ai
scripts/compose-env.sh production --profile production run --rm --no-deps --pull never admin migrate
scripts/compose-env.sh production --profile production run --rm --no-deps --pull never admin storage-init
scripts/compose-env.sh production --profile production up -d --no-build --pull never --wait --wait-timeout 180 public admin
python3 scripts/check-deployment-config.py --environment production --runtime
python3 scripts/smoke-compose.py production
curl --fail http://127.0.0.1:38080/health/ready
curl --fail http://127.0.0.1:38081/health/ready
```

The private admin listener must remain reachable only through the reviewed SSH/tailnet path. Check external HTTPS API/MCP separately, including trusted client-address forwarding, absence of admin/configuration routes, truthful coverage/source ages and known document/evidence references. Production retains its own state; no data copy or source activation occurs as part of image promotion.

Only after provider/source settings are reviewed and collection is deliberately authorized, enable the optional worker. Use [region → municipality → source activation](environments.md#territorial-and-source-activation) through CLI or admin first, with the production environment selected explicitly. Keep the current collector until the handover is scheduled; at that handover stop its development/staging worker and verify it is stopped before starting production. The continuous collector then runs only in production; this policy does not deduplicate intentionally parallel checks.

```sh
scripts/compose-env.sh production --profile production --profile production-worker up -d --no-build --pull never worker
scripts/compose-env.sh production --profile production --profile production-worker logs --tail=100 worker
```

For example, inspect the production territorial/source choices without starting processing:

```sh
scripts/compose-env.sh production --profile production run --rm --no-deps --pull never admin region-status 09
scripts/compose-env.sh production --profile production run --rm --no-deps --pull never admin municipality-status 09 050004
scripts/compose-env.sh production --profile production run --rm --no-deps --pull never admin source-status <source-id>
```

Replace the source placeholder before execution. Source/publication flags, territorial gates and worker selection are independent. Profiles do not stop a previously running worker or prevent its `unless-stopped` recovery after a host restart. Explicit service targets can activate profiles; default exclusion alone is not a runtime stop.

Its provider key belongs in `.local/production/secrets/regolo_api_key` or the private configured path and must be readable by application UID 65532. Do not select `application-backup` while whole-VM PBS protection is in use.

## Upgrade and rollback

Validate a new candidate in staging, publish it and obtain approval for its exact digest. Record a pre-release backup and the current compatible digest/configuration. Change only the intended private image/configuration settings, then repeat production pull, migration, storage initialization, listener startup and health checks above. Update an already activated worker explicitly with its reviewed settings. Application changes may need scheduled writer interruption; migrations must remain compatible with the prior image.

For application rollback, restore the **previous compatible digest and configuration** in `.local/production.env`, pull it and recreate listeners without rebuilding:

```sh
scripts/compose-env.sh production --profile production pull public admin
scripts/compose-env.sh production --profile production up -d --no-build --pull never --wait --wait-timeout 180 public admin
python3 scripts/smoke-compose.py production
```

Roll back an activated worker separately with both profiles, and verify known evidence references, source ages and backup monitoring. Preserve additive migrations, database/object volumes and acquired evidence. The municipal state migration preserves existing configured-source eligibility without overriding explicit choices; copied data therefore still require disabled territories and stopped nonproduction workers. Once municipal disablement is used, do not run an older worker image that ignores that gate: keep processing stopped until a compatible gate-enforcing image is selected. Rolling back profiles requires inspecting resolved selection again, since old defaults may select workers. Never use `down -v` or reverse migrations as an image rollback. An incompatible/corrupt data state requires stopping writers and restoring PostgreSQL plus RustFS together; explicitly decide how to handle post-backup data.

## Roll out dependency restart policies

All services now declare `unless-stopped`, but an already created container retains its prior policy until updated. After confirming recovery protection, capture selected container image IDs/mounts/policies privately, obtain the dependency IDs using the selected wrapper's `ps -q postgres rustfs crawl4ai`, and verify their project/service labels. Apply `docker update --restart unless-stopped` only to those verified IDs. This changes restart policies without rebuilding/recreating containers or replacing volumes. Verify unchanged image IDs/mounts and run the selected runtime/readiness checks. Record prior policies for rollback.

Reboot recovery needs an actual disposable-host rehearsal, then a separately scheduled current-host restart after the restore check. A nondisruptive runtime check does not prove a host reboot or a one-hour restore target.

## PBS backup and recovery target

Protect the entire shared VM on a physically separate Proxmox Backup Server. Confirm that every VM disk holding all three projects' PostgreSQL/RustFS data, the Compose checkout, `.local` settings and secrets, and required host configuration is included. External reverse-proxy or tailnet configuration outside the VM needs its own recovery record. A host failure now stops development, staging and production together. Keep the application `backup` worker disabled while whole-VM PBS protection is the selected strategy.

The agreed targets are **no more than six hours of lost data** and **one hour from outage detection to verified public API/MCP recovery**. Begin with a PBS backup every three hours, plus a pre-release backup. Alert on a failed job and when the newest successful, recoverable backup approaches four hours old; treat six hours as a breached objective. Measure real backup duration and adjust cadence if it erodes that margin. An initial retention proposal is the latest 16 snapshots, 14 daily, 8 weekly and 6 monthly snapshots; confirm actual PBS capacity and source retention obligations before adopting it. Schedule verification of new backups and periodic reverification of older ones, with failure notifications to the operational mailbox.

Use Proxmox snapshot mode with the running guest agent's filesystem freeze/thaw where supported, then prove recoverability rather than assuming a successful snapshot is application consistent. Before public activation and after material storage changes, restore a PBS backup to an isolated VM without production network identity or outbound collection, provider calls or notification delivery. Time from simulated outage detection until the restored public API/MCP is healthy and representative PostgreSQL records resolve to the correct RustFS originals and versions. Check that historical source-check times remain historical and incomplete jobs recover safely. Record backup age, restore time, data checks and any gap privately. A one-hour target is not achieved until the timed drill proves it.

The public [OpenSpec readiness tasks](../../openspec/changes/define-toscana-alert-service/tasks.md) keep PBS configuration, isolated restore and production deployment verification open until performed and evidenced. The committed examples are checked by `python3 scripts/check-deployment-config.py`; before deployment, inspect `scripts/compose-env.sh production --profile production config --format json` and confirm the project name, limits, ports and absence of application `build` settings.

Platform references: [GHCR image access and digest pulls](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry), [Proxmox VE backup and restore modes](https://github.com/proxmox/pve-docs/blob/master/vzdump.adoc), and [PBS pruning and verification](https://pbs.proxmox.com/docs/maintenance.html).
