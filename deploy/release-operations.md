# Release images and production recovery

This is the public procedure for preparing and checking an IWA release. It describes the intended production setup; it is not evidence that a production VM, PBS backup or public source is already active. Keep actual hostnames, credentials, backup records and approval evidence in private operational records. The [environment guide](../docs/environments.md) describes the dev/staging split.

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

## Production deployment

On the dedicated VM, copy [production.env.example](production.env.example) to ignored `.env` and replace the placeholder with the **approved digest**. Store production secrets only in its own ignored `.secrets` directory or another private deployment path. Set the public proxy trust and private administration origin for the actual deployment. The base Compose file has `build: .` for local development; the production `.env` selects [compose.production.yaml](compose.production.yaml), which removes those build directives. Also use `--no-build` on application starts and the pinned `IWA_APP_IMAGE` value. Do not run a production build or use a moving `latest` tag.

Before each deployment, record the current digest, confirm a recent recoverable PBS backup and that its alerts are working, verify the proposed migration remains compatible with the previous image, and check `docker compose config --quiet`. For initial startup, bring up PostgreSQL, RustFS and Crawl4AI, pull the approved image, run the admin `migrate` and `storage-init` commands, then start public and admin with `docker compose up -d --no-build --wait`. Start worker only after its provider and source settings have been reviewed. For an update, apply the same migration and health sequence while preserving volumes and source publication controls.

After deployment, check the external HTTPS API/MCP route, local readiness, exact private admin boundary, source-check ages, a known document/evidence reference, worker errors and backup monitoring. A failed application release rolls back to the previous compatible image digest and configuration. Do not reverse additive database migrations or delete PostgreSQL/RustFS volumes as an application rollback. An incompatible or corrupt data state requires stopping writers and a coordinated recovery of database and objects; decide explicitly how to handle data written after the selected backup.

## PBS backup and recovery target

Protect the entire production VM on a physically separate Proxmox Backup Server. Confirm that every VM disk holding PostgreSQL, RustFS, the Compose checkout, `.env`, `.secrets` and required host configuration is included. External reverse-proxy or tailnet configuration outside the VM needs its own recovery record. Keep the application `backup` worker disabled while whole-VM PBS protection is the selected strategy.

The agreed targets are **no more than six hours of lost data** and **one hour from outage detection to verified public API/MCP recovery**. Begin with a PBS backup every three hours, plus a pre-release backup. Alert on a failed job and when the newest successful, recoverable backup approaches four hours old; treat six hours as a breached objective. Measure real backup duration and adjust cadence if it erodes that margin. An initial retention proposal is the latest 16 snapshots, 14 daily, 8 weekly and 6 monthly snapshots; confirm actual PBS capacity and source retention obligations before adopting it. Schedule verification of new backups and periodic reverification of older ones, with failure notifications to the operational mailbox.

Use Proxmox snapshot mode with the running guest agent's filesystem freeze/thaw where supported, then prove recoverability rather than assuming a successful snapshot is application consistent. Before public activation and after material storage changes, restore a PBS backup to an isolated VM without production network identity or outbound collection, provider calls or notification delivery. Time from simulated outage detection until the restored public API/MCP is healthy and representative PostgreSQL records resolve to the correct RustFS originals and versions. Check that historical source-check times remain historical and incomplete jobs recover safely. Record backup age, restore time, data checks and any gap privately. A one-hour target is not achieved until the timed drill proves it.

The public [OpenSpec readiness tasks](../openspec/changes/define-toscana-alert-service/tasks.md) keep PBS configuration, isolated restore and production deployment verification open until performed and evidenced. The committed staging and production Compose examples are checked by `python3 scripts/check-deployment-config.py`; run `docker compose config --format json` on the production VM as well and confirm its application services contain no `build` setting.

Platform references: [GHCR image access and digest pulls](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry), [Proxmox VE backup and restore modes](https://github.com/proxmox/pve-docs/blob/master/vzdump.adoc), and [PBS pruning and verification](https://pbs.proxmox.com/docs/maintenance.html).
