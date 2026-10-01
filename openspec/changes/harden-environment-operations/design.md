# Design

## Context

See [proposal.md](proposal.md) for the review scope and [the environment operations specification](specs/environment-operations/spec.md) for acceptance behavior.

The repository already has one base Compose file, a production overlay, three environment examples, an explicit environment wrapper and release/configuration/smoke helpers. `x-app` supplies `restart: unless-stopped` only to application services; infrastructure services have no restart policy. The wrapper clears inherited `IWA_` and `COMPOSE_` variables but forwards caller arguments after its own fixed options. Its root discovery uses Git, so extracted downloads cannot use it.

Staging uses the base file, including `build: .`. The developer instructions referenced by release operations include an application build that omits the release helper's `VCS_REF`. Publication currently validates a local image's revision label, not the identity of the running staging containers. The smoke helper uses plain Compose, shell/default ports and `.secrets/`, and always stops RustFS. The deployment checker validates templates with an injected synthetic digest, not local deployment readiness. Secret generation specifies mode 0644 at creation but does not counteract the process umask.

There are no durable main specs yet. The new capability supplements the imported ingestion plan, preserving its project identities, readiness and source-acceptance gates. Existing installation credentials and named volumes must survive the changes.

## Goals / Non-Goals

**Goals:** Retain the existing tools as the operator entry points; make environment identity enforceable; preserve release image identity; provide a repeatable fresh-download path; and distinguish repository verification from host operational evidence.

**Non-Goals:** Change application APIs or database schema, activate sources as a side effect of setup, rotate existing secrets, rename projects/volumes, replace Compose/PBS, or automatically approve production. This proposal contains host readiness tasks but authorizes no host rollout, reboot or public activation by itself.

## Decisions

### 1. Recover existing stacks with Docker restart policies

Add `restart: unless-stopped` to PostgreSQL, RustFS and Crawl4AI in the base file. Production inherits this after activation; profiles continue to leave a never-started production project inactive. Use Docker's existing policy rather than introduce a second systemd supervisor. This preserves the operator's intentional stops and keeps management uniform across environments.

Validate resolved policies in CI and actual boot recovery on a disposable VM. For the current host, schedule a maintenance rollout and apply the new restart policies to verified dependency container IDs, preserving their running images and mounts rather than implicitly recreating services. Record prior policies for rollback. Readiness and database-to-object checks follow the update; a live-host reboot is a separately scheduled check after recovery protection exists.

### 2. Give the wrapper ownership of environment identity

Resolve the root from the wrapper's own physical location, removing Git as a dependency of basic environment management. Keep Git required by release creation/publication. Explicitly select the base Compose file for development and the base-plus-environment-overlay files for staging and production, avoiding implicit discovery of alternate configuration files.

Validate leading Compose global options before execution. Permit documented options such as explicit profiles and dry-run behavior; reject project name, project directory, environment file and Compose file replacements, including short, long and equals forms. Fail on unknown leading options rather than silently forwarding an unreviewed override. Subcommand options remain available, with production build restrictions handled separately. Continue clearing inherited `IWA_`/`COMPOSE_` variables; verify the resolved project and file selection in adversarial tests.

Alternative: moving `-p` later only addresses one duplicate flag and does not prevent configuration-file replacement. The wrapper is a guard against accidental misuse; direct Docker access remains an operator capability.

### 3. Validate templates, local configuration and production operations separately

Keep `scripts/check-deployment-config.py` as the validation entry point. Its default mode validates committed examples and reports template validation explicitly. Add `--environment development|staging|production` for local configuration and `--runtime` for nondisruptive checks of running services, readiness, actual bindings, mounted volumes, restart policies and image revision differences. Use the wrapper's resolved JSON configuration as the input; do not shell-source private `.env` files. Output selected operational fields without dumping full environment values, secret contents or logs.

The production wrapper must reject `build` and `--build`, and validate the effective application image before `pull`, `up`, `run`, `create`, `start` or `restart`. Require a GHCR path followed by `@sha256:` and exactly 64 hexadecimal characters. Keep `config`, `ps`, logs, stop and teardown usable with the preparation placeholder, while local validation marks the image as unselected. A syntactically valid digest is not proof of operator approval or registry availability. Actual publication, pull and approval evidence remain required in the release runbook.

Alternative: a Compose required-variable expression rejects only missing values and cannot enforce digest syntax. A synthetic digest in a template test cannot validate an actual operator setting.

### 4. Preserve the exact candidate through staging and publication

Add a staging overlay that removes application build directives, and select it through the wrapper. Development retains explicit builds. The release helper remains the one place that builds a candidate with its Git revision label. Document staging dependencies, migrations, storage initialization and listener starts directly, using the existing candidate. Listener `up` uses `--no-build --pull never`; migration/storage `run` uses `--pull never` and the image-only overlay because Compose does not support `run --no-build`. These commands must never direct users back to a development build step.

Before pushing, publication checks the configured staging image and the actual public/admin image IDs and revision labels against the local candidate; check optional application workers when running. A wrong or older image fails publication before network mutation. Expose configured/running/checkout revisions through local runtime checks. An older running development revision is informational and never triggers automatic rebuilding. If a tag is rebuilt or moved, the runtime identity comparison exposes the mismatch even if the revision label was retained.

Alternative: fixing only the docs leaves an accidental `staging build` available; the overlay makes release-image reuse part of resolved configuration. No application changes or release-approval datastore are needed.

### 5. Make smoke checks nondisruptive unless explicitly requested

Use `python3 scripts/smoke-compose.py <environment>` for ordinary checks and add `--allow-interruption` for deliberately planned development/staging fault injection. Reject the interruption option for production before any operation. Derive ports, active services and secret files from the wrapper's resolved configuration and runtime inspection. Verify project labels, actual listener bindings and expected volumes before fault injection. Invoke stop/start through the same wrapper, and retain recovery in `finally` on the same target.

Include HTTP readiness and public/admin boundary checks by default. Handle optional stopped workers and unused missing provider keys according to active services. Check secret leakage against resolved secret files used by the checked services; do not print secrets, response bodies or unrestricted container inspection. This also replaces the old `.secrets/` assumption in documentation.

### 6. Set access permissions only on newly created files

Keep exclusive creation and value preservation. Apply mode 0644 with `os.fchmod` on each newly created descriptor, including disabled notification/backup configuration files, inside the existing mode-0700 secret directory. Descriptor-based permission setting avoids changing an unrelated path after creation. Existing files keep their contents and custom ACL/access arrangements. Local checks diagnose unreadable files and document a targeted repair instead of silently chmodding every credential.

Test umasks 022 and 077, repeated initialization and custom existing modes with synthetic credentials in temporary directories. Confirm readability by the actual application UID during the disposable container rehearsal.

### 7. Publish one environment workflow with explicit examples

Use `docs/environments.md` as the canonical managed-environment guide. Make `docs/developer-setup.md` a direct fresh-development procedure using the wrapper, and put full staging release and production prepare/deploy/rollback commands in the relevant guides. Align the README and documentation index to the same-host topology. Create required `.local/` directories before copying examples. Identify Bash, Python, Docker/Compose, curl and optional Go; identify Git as required for release creation and state a tested Compose version supporting the overlays.

Keep fresh setup separate from adoption of existing volumes. Use environment-specific provider-key paths, ports and secret directories in every example. Show status, logs, nondisruptive checks, planned fault tests, stop/start, explicit development rebuilds, image-only staging upgrades, production approved-digest upgrades and compatible-image rollback. Explain that Compose profiles are selection controls: explicitly targeting a service can activate it, so profiles do not replace release approval.

Rehearse the documented sequence from a fresh checkout and an extracted archive on disposable resources, with synthetic fixtures and no provider calls. The opt-in `scripts/rehearse-environments.py --run` creates a temporary Git snapshot and unique test namespaces, initializes under a strict umask, exercises image-preserving startup/fault recovery/stop-start, checks inactive production, and cleans up only its own resources. Archive management is covered by regression tests. Record the tested command sequence and redistributable verification results. This Docker rehearsal does not replace the separate VM reboot check. Avoid a new configuration framework or mandatory agent tooling for users.

### 8. Complete host readiness as measured operations

The existing ingestion tasks remain the source of truth for capacity (1.8), PBS protection (8.2), isolated recovery (8.3), publication (12.4) and production verification (10.2). This change references them and adds precise execution/checkpoints; do not check both lists merely because code is complete.

Measure simultaneous development/staging and representative production workload on an isolated host, including build peaks, object/database growth and disk retention. Review sustained disk/RAM use against the existing 70% threshold. Choose and record host expansion, limit changes or scheduling restrictions from evidence, preserving a margin for the OS and transient work; compare against current production ceilings without treating ceilings as measured usage.

Before current-host operational verification, establish off-host PBS coverage for all environment data and recovery settings. Configure the already selected backup/age-alert policy, prove a failed/aging backup is visible, and complete the isolated restore drill with outbound activity disabled. Preserve the existing six-hour data-loss and one-hour recovery targets. Publish/test-pull the staged release, review public proxy trust and private admin routing, and then perform an approved production deployment and rollback rehearsal. Operational evidence and secrets remain private; publish only reviewed completion/status summaries.

## Risks / Trade-offs

- Stricter CLI behavior breaks conflicting wrapper options and implicit smoke targets → provide clear errors, migration examples and negative regression tests.
- Existing services require policy rollout to receive new restart settings → schedule application-preserving maintenance and verify runtime configuration rather than claiming a file edit fixed the host.
- A preparation placeholder must remain inspectable → test preparation commands and invalid-image rejection as distinct behaviors.
- Staging may be intentionally stopped when publication is attempted → require running staging validation before publishing and explain the missing evidence.
- Mode 0644 is broader inside containers → retain the host's owner-private secret directory, mount only required files and preserve custom existing ACLs.
- Capacity and recovery cannot be proven in ordinary CI → keep external tasks open until measured evidence and required infrastructure access exist.
- A shared-host outage affects every environment → retain off-host whole-VM protection and a tested independent restore target.

## Migration Plan

1. Implement repository changes and focused regression tests; update docs/spec progress/changelog in the same change.
2. Run template checks and the fresh-download rehearsal on disposable resources, leaving operational projects untouched.
3. Establish recoverable host protection, then schedule current-host policy rollout and environment-targeting checks while preserving volumes, credentials and selected application images. Verify source/evidence references after the update.
4. Build a clean identified candidate, run staging migrations/fixtures/queries without rebuilding, and publish/test-pull its digest with private release evidence.
5. Complete capacity, backup alerts, timed restore and routing checks. Obtain the existing explicit production approval for the concrete digest and deployment configuration.
6. Activate production through the reviewed runbook and verify HTTPS API/MCP, private administration, optional workers and rollback. Update operational task status only after the corresponding evidence exists.

Rollback of repository tooling restores the previous scripts/docs and compatible configuration without changing data. Application rollback selects the previous compatible digest/configuration and preserves additive migrations. Corrupt data recovery restores PostgreSQL and objects together from the chosen protected state, following the existing incident procedure.

## Open Questions

- What capacity expansion or scheduling restriction will the measured workload require? Resolve during the capacity task before approving production.
- Which private PBS/proxy endpoints and retention settings will be used? Resolve from the operator's infrastructure records during the existing readiness tasks; none are needed to implement repository fixes.
