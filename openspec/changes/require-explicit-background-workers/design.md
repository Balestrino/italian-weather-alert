# Design

## Context

See [proposal.md](proposal.md) for motivation and [the specification](specs/background-processing-policy/spec.md) for acceptance behavior. The base Compose file contains unprofiled `worker` and `backup` services. Staging inherits both and only removes application build settings. Production already assigns optional profiles to both services and profiles every core service.

The environment wrapper clears inherited Compose overrides, pins project/configuration identity and already supports explicit profiles. The deployment checker loads all services for topology checks, while production default selection is checked separately. Existing environment tests and Docker configuration resolution provide a validation path that requires no source acquisition or provider calls.

The admin panel already exposes collection activation and suspension, calling `registry.Store.EnableCollection` and `Disable` with a source revision and actor. The existing binary has a `preview` command but no dedicated collection enable/suspend commands. Collection activation validates preview evidence, collection permission and territorial configuration, then records the change in the selected registry. CLI support must reuse these operations.

Region enablement already exists through `domain.Store.SetRegionEnabled` and the regional setup form. The overview currently only displays state and configuration links; Comuni displays source counts without an independent enablement flag. `territorial_municipalities` is immutable geographical membership, so execution choices must not be written into those imported rows. Shared SQL execution gates already protect acquisition and interpretation based on the region and profile.

## Goals / Non-Goals

**Goals:** Make listener startup independent of background work, provide hierarchical region/municipality/source controls with equivalent CLI/admin operations, preserve explicit controlled checks, and document how existing collectors adopt the policy.

**Non-Goals:** Add host-wide locking, share databases or caches between environments, split the acquisition and inference loops into separate services, or activate production. The existing worker also runs configured notifications; stopping it stops those loops too. Manual admin acquisition previews remain deliberate operations.

## Decisions

### 1. Profile both optional services in the base configuration

Use `processing-worker` for the base collection/inference service and `application-backup` for the base backup service. Staging inherits these names. This matches the existing production approach and protects a plain `up` without changing worker behavior or requiring new runtime flags. Keep the explicit-target path documented, because Compose permits naming a profiled service directly.

A worker enablement environment variable would add application branching and could leave other startup loops active. A host-wide lease would prevent deliberate parallel experiments and require additional lifecycle design; neither is needed for default-startup protection.

### 2. Replace production worker profiles rather than merge them

When adding a base profile, verify the actual Compose merge result. The production worker must have only `production-worker`; use the supported `!override` tag if needed to prevent list merging from retaining `processing-worker`. The backup profile remains `application-backup`. The tested Compose version already supports tagged overlays; confirm parsing through template validation.

### 3. Validate resolved selection, not just YAML strings

Extend `scripts/deployment_checks.py` using the existing `Environment` interface. Resolve the default configuration, all services, and selected worker/backup/core profiles, and check selected service names. Preserve full topology validation and inactive production inspection. Add focused regression cases to `scripts/test_environment_operations.py` including production inheritance and explicit service selection with Compose dry-run or equivalent nonmutating resolution. Existing runtime checks must still tolerate inactive optional workers.

### 4. Keep operational responsibility explicit

Update developer setup, environment management and releases with profile activation and explicit `stop worker` after checks. Document fixture-based nonproduction validation, disabled notifications and collection when copying representative data, and production's eventual continuous role. Explain that `restart: unless-stopped` still recovers an already activated worker: profiles only affect Compose selection and do not stop it on a Docker restart.

### 5. Reuse the source registry lifecycle from CLI and admin

Add binary commands `source-enable-collection <source-id> <revision> <actor>`, `source-suspend-collection <source-id> <revision> <actor>` and `source-status <source-id>` under the admin role. Invoke them using `scripts/compose-env.sh <environment> run --rm --no-deps --pull never admin <command> ...` against an already initialized environment. Validate command arity and arguments, use bounded database operations, and return JSON state/action results or nonzero status with stable errors that do not expose private configuration. Call the same registry methods as the existing admin routes rather than write registry flags directly.

Keep the existing admin activation/suspension forms and state view. Add concise guidance there and in CLI output that source enablement only configures collection and an active worker is required. Do not infer worker liveness merely from source state; live worker monitoring is outside this change. Neither path starts Docker services or performs a preview as part of enablement. Preview remains a separate deliberate operation, and public publication keeps its existing acceptance gates.

Verify interface parity against synthetic registry fixtures, including missing preview, invalid/stale revision, collection policy, territorial gates, audit actor and visibility of a CLI change through the admin view. Source control checks can run with no worker process, preventing acquisition or inference side effects. Direct SQL scripts and a curl-only recipe were considered: the former bypasses invariants and the latter does not provide the dedicated CLI requested by the operator.

### 6. Make region and municipality choices independent execution gates

Add explicit region actions to `/admin/regions`, reusing the audited region lifecycle and existing configuration prerequisites. Keep Configura regione as the route into the region workspace, with Comuni providing individual municipality actions and status. Preserve advanced regional setup for datasets and processing profiles. Actions must submit explicit values, actor and expected revision through native POST forms and return to validated local destinations with filters intact.

Introduce additive municipal lifecycle state and events keyed by region code and ISTAT, with a monotonic expected revision and row locking. Validate against the adopted register when enabling a current municipality; retain disabled/retired identities for history and permit safe disablement even when setup becomes incomplete. Preserve choices across register replacement rather than reinitializing them per dataset. New identities default to disabled; backfill previously configured local-source identities as enabled with migration attribution to preserve existing execution eligibility, without changing source flags or region state. Do not modify existing migration checksums or immutable geographical rows.

Extend shared territorial SQL eligibility to require the municipal state for local sources, using a consistent region-then-municipality lock order for execution admission and mutation. Existing acquisition claims, queued document/inference job admission, manual previews and interpretation launch boundaries must honor it. Already admitted operations can finish, matching current region semantics. Disabling a parent is a gate, not a bulk rewrite of source or child flags. This avoids losing explicit subordinate choices and permits normal eligibility restoration without creating catch-up batches.

Expose configured versus effective state and blocking reasons in overview, municipality list/detail, source guidance and territorial history. A disabled municipality only gates its local source processing; it does not stop the region's shared bulletin collection or hide retained regional warnings. Existing source counts must not claim effective acquisition solely from `collection_enabled` when a territorial gate blocks execution.

Add `region-enable|region-disable <region-code> <revision> <actor>`, `region-status <region-code>`, `municipality-enable|municipality-disable <region-code> <istat> <revision> <actor>` and `municipality-status <region-code> <istat>` under the admin role and explicit environment wrapper. Use shared audited domain operations, JSON results and operator-safe errors as for source controls. Test parent override, child-choice preservation, wrong membership, concurrency/revision conflicts, dataset replacement and CLI/admin parity with synthetic fixtures. Regional enablement never enables municipalities or source collection automatically.

## Risks / Trade-offs

- Production profile merging could leave an unintended activation path → assert the resolved production worker profile and service selection.
- A previously activated worker keeps running after a configuration edit → provide explicit stop/status adoption steps and make no claim that repository changes stopped it.
- An intentional parallel test can still duplicate provider work → document bounded checks and source scope; this change is not a cross-environment lock.
- Stopping the current sole collector before production is ready would interrupt collection → preserve it through repository implementation and perform a deliberate handover separately.
- Default commands now omit background services → mark the behavior change and retain documented explicit activation for operator workflows.
- Source enablement might be mistaken for active collection → distinguish the persisted source flag from worker execution in CLI output and admin guidance.
- Dataset replacement or parent disablement could erase municipality choices → store lifecycle state separately and keep historical geographical records immutable.
- A new municipality gate could interrupt established collection → backfill existing configured local-source identities with migration attribution and verify effective eligibility before and after on synthetic data.

## Migration Plan

1. Implement profiles, resolved-configuration checks, source and territorial CLI/admin parity, an additive municipal lifecycle migration, focused regressions and documentation together; record the implementation separately from this proposal in the changelog.
2. Run `python3 scripts/test_environment_operations.py` and `python3 scripts/check-deployment-config.py`, checking all three default and optional service sets without operational startup. Run command argument tests and synthetic source/territorial lifecycle integration tests covering CLI/admin parity and migration compatibility without official-source or provider calls.
3. Inspect existing workers through environment-specific status commands. Repository implementation leaves operational containers unchanged.
4. Keep the current live collector until production is ready under the existing release procedure. At handover, stop the development worker explicitly, verify it is stopped, and activate only the reviewed production collector. End each subsequent nonproduction check with an explicit stop and status verification. Keep private host evidence outside commits.
5. Rollback restores compatible configuration and commands, preserves additive municipality state/history, and does not automatically restart a stopped worker. Once municipal disablement is in use, rolling back to an older binary that ignores its gate must not be presented as safe for processing; keep workers stopped until a compatible gate-enforcing version is selected. No reverse migration is used.
