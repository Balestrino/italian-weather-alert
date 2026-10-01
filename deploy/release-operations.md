# Release images and production recovery

This is the public procedure for preparing and checking an IWA release. The production Compose project is configured on the same VM as development and staging but has not been started; PBS backup and public source readiness remain open. Keep actual hostnames, credentials, backup records and approval evidence in private operational records. The [environment guide](../docs/environments.md) describes all three projects.

## Image identity and manual approval

Use one clean Git revision in the shared checkout. Run CI-equivalent tests and the checks relevant to the change. Do not change the checkout between build, staging validation and publish. Build once with:

```sh
scripts/release-image.sh build
```

The helper prints `ghcr.io/balestrino/italian-weather-alert:<full-commit-sha>`. Set `IWA_APP_IMAGE` in staging's ignored `.local/staging.env` to that exact tag. Use `scripts/compose-env.sh staging` to start dependencies, migrate and initialize storage as described in the [developer setup guide](../docs/developer-setup.md), then bring up the public and admin listeners with the tested image. Keep the worker off unless the change requires a controlled source or inference check.

Every release check includes:

1. Passing Go test, race, vet, build, integration coverage and vulnerability workflows for the intended revision.
2. Successful `docker compose config --quiet`, staging migration and storage initialization, healthy PostgreSQL/RustFS/Crawl4AI and both HTTP listeners.
3. Representative JSON API and MCP queries against known staging fixtures, truthful coverage/updating state, and absence of administration on the public listener. Check the administrative listener only from its restricted path.
4. Review of logs, failed jobs, source status and any change-specific source or inference regression. A code pass does not accept or publicly enable a source.

The [Compose fault-injection test](../scripts/smoke-compose.py) stops RustFS; run it only on a disposable project or during an explicitly planned staging interruption. Do not use it as a routine production smoke check.

After the staging checks, authenticate to GHCR using private credentials and run `scripts/release-image.sh publish`. This pushes the tested local image and prints a `ghcr.io/...@sha256:...` reference. The first GHCR package is private by default; an operator must explicitly make the image public before using anonymous production pulls. Record the Git revision, staging results, published digest, prior production digest and operator approval together. A public image contains application code, never deployment secrets. Keep at least the previous compatible digest available for rollback.

## Prepared production project and resource limits

The prepared project uses ignored `.local/production.env` and `.local/production/secrets` in the shared checkout. [The environment wrapper](../scripts/compose-env.sh) fixes the `iwa-production` project name and adds [compose.production.yaml](compose.production.yaml), which removes local build directives. The example reserves loopback ports `38080` and `38081`; the proxy and private operator route have not been connected to them. The image remains a placeholder until an approved digest replaces it. `config` and `ps --all` are safe preparation checks; neither starts containers.

Every production service has a profile, so an unqualified `up` selects no service. The `production` profile contains the listeners, PostgreSQL, RustFS and Crawl4AI. The `production-worker` profile is separate, and the `application-backup` profile is not used while PBS protects the VM. The initial per-container ceilings are:

| Service | RAM | CPU |
| --- | ---: | ---: |
| Public, admin | 256 MiB each | 0.5 each |
| PostgreSQL, RustFS | 768 MiB each | 1.0 each |
| Crawl4AI | 1536 MiB | 1.5 |
| Worker, when separately enabled | 768 MiB | 1.5 |
| Application backup, normally disabled | 512 MiB | 0.5 |

The five core services can reach **3.5 GiB** in aggregate; the worker adds **0.75 GiB**. These are ceilings, not measured requirements. Named volumes have no disk quota, and a container may fail if its limit is too low. Rehearse the expected workload in staging, measure host headroom and disk growth, and adjust the limits before approving production startup. With 8 GiB RAM and a 50 GiB disk on the shared VM, concurrent development/staging activity and the current free disk require a capacity decision before activation.

## Production deployment

Before deployment, replace `IWA_APP_IMAGE` in `.local/production.env` with the **approved digest** and set public proxy trust, public origin and private administration origin for the actual route. Use `--no-build` on application starts. Do not run a production build or use a moving `latest` tag. Production keeps its own empty data and secrets; code promotion does not copy development or staging evidence.

Before each deployment, record the current digest, confirm a recent recoverable PBS backup and that its alerts are working, verify the proposed migration remains compatible with the previous image, and check `scripts/compose-env.sh production --profile production config --quiet`. For initial startup, bring up PostgreSQL, RustFS and Crawl4AI under the `production` profile, pull the approved image, run the admin `migrate` and `storage-init` commands, then start public and admin with `--no-build --wait`. Start the worker only after its provider and source settings have been reviewed, by explicitly selecting `production-worker` as well as `production`. For an update, apply the same migration and health sequence while preserving volumes and source publication controls. Do not select `application-backup` while whole-VM PBS protection is in use.

After deployment, check the external HTTPS API/MCP route, local readiness, exact private admin boundary, source-check ages, a known document/evidence reference, worker errors and backup monitoring. A failed application release rolls back to the previous compatible image digest and configuration. Do not reverse additive database migrations or delete PostgreSQL/RustFS volumes as an application rollback. An incompatible or corrupt data state requires stopping writers and a coordinated recovery of database and objects; decide explicitly how to handle data written after the selected backup.

## PBS backup and recovery target

Protect the entire shared VM on a physically separate Proxmox Backup Server. Confirm that every VM disk holding all three projects' PostgreSQL/RustFS data, the Compose checkout, `.local` settings and secrets, and required host configuration is included. External reverse-proxy or tailnet configuration outside the VM needs its own recovery record. A host failure now stops development, staging and production together. Keep the application `backup` worker disabled while whole-VM PBS protection is the selected strategy.

The agreed targets are **no more than six hours of lost data** and **one hour from outage detection to verified public API/MCP recovery**. Begin with a PBS backup every three hours, plus a pre-release backup. Alert on a failed job and when the newest successful, recoverable backup approaches four hours old; treat six hours as a breached objective. Measure real backup duration and adjust cadence if it erodes that margin. An initial retention proposal is the latest 16 snapshots, 14 daily, 8 weekly and 6 monthly snapshots; confirm actual PBS capacity and source retention obligations before adopting it. Schedule verification of new backups and periodic reverification of older ones, with failure notifications to the operational mailbox.

Use Proxmox snapshot mode with the running guest agent's filesystem freeze/thaw where supported, then prove recoverability rather than assuming a successful snapshot is application consistent. Before public activation and after material storage changes, restore a PBS backup to an isolated VM without production network identity or outbound collection, provider calls or notification delivery. Time from simulated outage detection until the restored public API/MCP is healthy and representative PostgreSQL records resolve to the correct RustFS originals and versions. Check that historical source-check times remain historical and incomplete jobs recover safely. Record backup age, restore time, data checks and any gap privately. A one-hour target is not achieved until the timed drill proves it.

The public [OpenSpec readiness tasks](../openspec/changes/define-toscana-alert-service/tasks.md) keep PBS configuration, isolated restore and production deployment verification open until performed and evidenced. The committed examples are checked by `python3 scripts/check-deployment-config.py`; before deployment, inspect `scripts/compose-env.sh production --profile production config --format json` and confirm the project name, limits, ports and absence of application `build` settings.

Platform references: [GHCR image access and digest pulls](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry), [Proxmox VE backup and restore modes](https://github.com/proxmox/pve-docs/blob/master/vzdump.adoc), and [PBS pruning and verification](https://pbs.proxmox.com/docs/maintenance.html).
