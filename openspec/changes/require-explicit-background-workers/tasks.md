# Tasks

Checked tasks record repository implementation verified with synthetic fixtures; unchecked tasks remain open. Repository verification uses synthetic fixtures and configuration resolution; it does not activate or stop operational workers. Production rollout and the live collector handover remain separate operations under the existing release procedure.

## 1. Explicit service selection

- [x] 1.1 Add `processing-worker` and `application-backup` profiles to the base worker and backup services; verify resolved default development/staging select only public, admin, postgres, rustfs and crawl4ai, and each optional profile selects its own worker without the other.
- [x] 1.2 Preserve production's separate profile membership, replacing inherited worker profiles when needed; verify default production selects no services, the core profile excludes both workers, `production-worker` selects the processing worker, and `processing-worker` does not select it.

## 2. CLI and admin source controls

- [x] 2.1 Add admin-role CLI commands for source collection enablement, suspension and state inspection, reusing existing registry operations with validated source/revision/actor arguments and JSON results; verify command parsing, role enforcement, operator-safe nonzero failure and successful state changes in synthetic fixtures without starting a worker or making acquisition/provider calls.
- [x] 2.2 Preserve the admin panel's collection controls and add CLI/admin guidance separating source enablement from worker execution and public publication; verify synthetic integration cases give equivalent CLI/admin outcomes for valid activation, missing preview, stale revision, policy/territorial restrictions, audit identity, suspension and cross-interface state visibility while the worker is absent.

## 3. Territorial activation hierarchy

- [x] 3.1 Add an additive, audited municipality lifecycle with validated region/ISTAT identity, expected revision, disabled defaults for new identities and compatibility backfill for municipalities with existing configured local sources; verify synthetic migration, stale revision, wrong membership, dataset replacement and retired-identity history cases without modifying immutable geography or existing migration checksums.
- [x] 3.2 Extend shared territorial execution admission for acquisition, queued document/inference jobs and manual previews/interpretation to honor municipality gates for local sources with region-first locking; verify parent disablement blocks new work, admitted work may finish, regional/other municipal sources remain unaffected by one municipality's disablement, and re-enable preserves child/source choices without enqueuing a catch-up batch.
- [x] 3.3 Add explicit region enable/disable actions to `/admin/regions` and municipal enable/disable actions to Configura regione → Comuni, with configured/effective status, blocking reasons and lifecycle history; verify native forms, actor/revision checks, territorial context, filter/pagination retention, keyboard use, 390px viewport and 200% zoom using synthetic fixtures.
- [x] 3.4 Add dedicated CLI region/municipality enable, disable and status commands using the shared domain lifecycle; verify CLI/admin parity, explicit environment scope, prerequisite/conflict errors, JSON state results, no automatic source/public activation and no worker or provider side effects.

## 4. Deployment validation

- [x] 4.1 Extend deployment configuration checks to validate default and explicit background service sets for every environment; verify missing/default-enabled profiles and production profile inheritance fail with an environment/service diagnosis, while full topology and inactive-production checks still pass.
- [x] 4.2 Add focused environment regression coverage for default exclusion, independent profile selection, production inheritance, explicit service targets and inactive-worker runtime compatibility; verify `python3 scripts/test_environment_operations.py` and `python3 scripts/check-deployment-config.py` pass without operational startup, source acquisition, provider calls or notifications.

## 5. Operator procedure and completion

- [x] 5.1 Update developer setup, environment management and release guides with opt-in commands, region → municipality → source activation steps, explicit environment-scoped territorial/source CLI examples and equivalent admin actions, post-test stop/status commands, copied-data safeguards and the eventual production-only continuous collector policy; verify every documented CLI command names an environment and explain territorial/source/worker/publication separation, explicit-target activation, lack of cross-environment deduplication and existing-container restart behavior.
- [x] 5.2 Document adoption preserving the current collector until an explicit handover, additive migration/compatible rollback, and update the OpenSpec index/checklist/changelog alongside implementation; verify local documentation links, strict validation of this change and `git diff --check`, and report repository validation separately from any unperformed runtime handover.

## 6. Embedding activation controls — added 2 October 2026

- [x] 6.1 Add disabled global/source flags and audited revisions; exclude disabled embedding claims without consuming attempts, preserve admitted completion/source choices, and verify repeated migration plus newly registered source defaults with synthetic PostgreSQL.
- [x] 6.2 Add native admin global/source forms and equivalent admin-role CLI controls; verify shared state, stale/invalid arguments, private access boundaries and no worker/provider side effects.
- [x] 6.3 Wire live per-extraction admission, ordinary linking while disabled, current-policy semantic retrieval and lazy embedding configuration; remove activation by the legacy environment flag and verify scheduling/provider-initialization behavior.
- [x] 6.4 Update configuration/operations guidance, run affected repository/integration/deployment checks, adopt the development admin/workers with all flags disabled, and archive only the explicitly requested pending embedding backlog with private verification evidence.
