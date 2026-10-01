# Environment tooling verification

On 1 October 2026, the repository tooling passed a disposable Docker rehearsal
with Docker Compose v5.5.1. The rehearsal used a fresh Git snapshot, generated
synthetic credentials, temporary ports and unique test project/volume names.
It ran development and staging sequentially on the shared host. It made no
provider or notification calls and removed its own resources afterward.

Verified behavior:

- Dependency startup, database migration, storage initialization and both
  listeners work from the documented environment configuration.
- New secrets generated under umask 077 are readable by their container users.
- Staging migrations and startup preserve the built image ID and revision.
- HTTP readiness and public/admin separation pass; project labels, listener
  bindings, named volumes and secret mounts match the selected environment.
- Development and staging recover from a selected RustFS outage and from an
  explicit container stop/start; their runtime restart policies match Compose.
- All three configurations use independent volumes and generated credentials.
  Prepared production has no selected services or containers.

The regression suite also checks archive-based configuration from another
working directory, conflicting shell/CLI overrides, production digest/build
rejection, preserved existing secret permissions, secret-leak redaction,
nondisruptive smoke checks, fault recovery after assertion failure, and rejection
of staging image mismatches before publication. Matching synthetic candidates
reach a mocked push; this does not certify a registry publication.

The synthetic JSON API/MCP usage tests and five shared public-query integration
groups passed. These use test-owned fixtures and do not establish acceptance or
public availability of any alert source.

To repeat the isolated rehearsal, follow the prerequisites in
[developer setup](developer-setup.md), then run:

```sh
python3 -m unittest scripts/test_environment_operations.py
python3 scripts/check-deployment-config.py
python3 scripts/rehearse-environments.py --run
```

The last command builds an image and starts temporary Docker services; allow
spare host capacity. Its image and namespaces belong only to the rehearsal.

This verification does not prove VM reboot recovery, off-host backup/restore,
concurrent production capacity, public GHCR availability, external routing or
production deployment/rollback. The user took ownership of backup/restore and
VM checks; their results were not independently verified here. The current-host
restart-policy rollout and remaining release gates stay open in the
[implementation checklist](../openspec/changes/harden-environment-operations/tasks.md).
Private operational measurements and evidence remain outside the repository.
