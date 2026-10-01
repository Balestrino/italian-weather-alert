# Proposal

## Why

The environment review found working data isolation but gaps in reboot recovery, environment selection, release image preservation, production image validation, secret permissions and fresh-checkout instructions. These gaps can interrupt the wrong installation or leave a new operator unable to reproduce and manage the intended three-project deployment.

## What Changes

- Give PostgreSQL, RustFS and Crawl4AI the same deliberate restart policy as the application services, and verify recovery on an isolated host.
- Enforce the environment wrapper's project, Compose files and environment-file identity; reject conflicting command-line overrides while retaining explicit profile selection.
- **BREAKING**: Require an explicit environment for the Compose smoke test; make ordinary checks nondisruptive, require an explicit interruption option for fault injection, and reject production fault injection.
- Preserve the release helper's exact image through staging, migration, publication and production promotion; make rebuild behavior explicit and report configured/running/checkout revision differences.
- Reject production application tags, malformed digests and placeholders before image-consuming operations, while keeping inactive preparation and diagnostic commands usable.
- Make newly generated secret files readable by container users regardless of the invoking user's umask, without rotating existing credentials.
- Publish consistent fresh-checkout and existing-installation instructions, explicit commands for all environments, prerequisites, troubleshooting, upgrades and rollback.
- Add regression checks and a disposable onboarding rehearsal, then complete a separately scheduled host-readiness track covering capacity, off-host backup/restore, public image publication, routing and approved deployment.

## Capabilities

### New Capabilities

- `environment-operations`: Reliable, explicit management of development, staging and prepared production; safe validation and smoke checks; reproducible onboarding and release promotion; evidence-based host readiness.

### Modified Capabilities

None. There are no durable capabilities under `openspec/specs/` yet. This focused capability supplements the environment and recovery requirements in `define-toscana-alert-service/specs/official-source-ingestion/spec.md`; it does not duplicate or close that change's source-acceptance or deployment gates.

## Impact

Affected implementation: `compose.yaml`, a new `deploy/compose.staging.yaml`, `deploy/compose.production.yaml`, `scripts/compose-env.sh`, `scripts/smoke-compose.py`, `scripts/check-deployment-config.py`, `scripts/init-secrets.py`, `scripts/release-image.sh`, relevant script tests and CI. Affected documentation: `README.md`, `docs/developer-setup.md`, `docs/environments.md`, `docs/README.md`, `deploy/release-operations.md`, OpenSpec planning/progress and `CHANGELOG.md`.

Existing project names, volumes and credential values must remain attached to their environments. Host changes and production activation require their existing operator authorization and external infrastructure access; planning and repository checks cannot establish operational readiness. Private settings, source captures, capacity measurements and recovery/approval records remain outside commits.
