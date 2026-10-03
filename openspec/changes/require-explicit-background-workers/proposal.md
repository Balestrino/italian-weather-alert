# Proposal

## Why

Development and staging inherit unprofiled collection/inference and backup workers, so an unqualified Compose startup can activate expensive background work alongside the listeners. Separate environment databases do not share acquisition, interpretation or OCR deduplication, making simultaneous live processing unnecessarily costly on a shared host.

## What Changes

- Add environment-scoped global and per-source embedding controls in the admin panel and CLI, both disabled by default even in existing environments. Require both flags for scheduling, queue admission and semantic retrieval; retain paused jobs and source choices, and retire only operator-selected pending embedding jobs through archival.

- **BREAKING**: Exclude the collection/inference worker and application backup worker from default development and staging service selection; give each an explicit opt-in profile.
- Preserve production's separate `production-worker` and `application-backup` selection, its inactive preparation, and existing release gates.
- Verify default and explicitly selected service sets in deployment checks and focused regression tests without contacting sources or inference providers.
- Document production as the eventual continuous live collector, with development/staging workers used for bounded checks and stopped afterwards. Keep copied nonproduction data inactive for collection, inference and notifications.
- Document adoption for existing worker containers: configuration profiles do not stop running processes or change Docker reboot behavior.
- Support per-source collection enablement, suspension and state inspection through a dedicated CLI and the existing admin panel, using the same persisted lifecycle rules and audit events in the explicitly selected environment. Source activation must remain independent of worker startup and public publication.
- Make territorial activation the primary admin workflow: enable/disable regions from `/admin/regions`, then enable/disable individual municipalities from the region's Comuni tab reached through Configura regione. Persist a separate municipality execution gate, expose equivalent CLI controls, and preserve individual source settings when a parent territory is disabled.

## Capabilities

### New Capabilities

- `background-processing-policy`: Explicit background service selection, hierarchical region/municipality/source controls with equivalent CLI/admin operations, and a documented shared-host operating policy across the three environments. This supplements the in-flight `environment-operations` capability in `harden-environment-operations`, territorial administration in `organize-admin-by-territory`, and the source lifecycle in `define-toscana-alert-service`; there are no main specs to modify yet.

### Modified Capabilities

None.

## Impact

Affected files: `compose.yaml`, `deploy/compose.production.yaml` if profile replacement is necessary, `scripts/deployment_checks.py`, `scripts/test_environment_operations.py`, `cmd/iwa/` and its tests, source and territorial admin forms/routes, domain migrations and territorial execution gates, operations read models/history, focused lifecycle integration tests, developer/environment/release guides, the OpenSpec index and changelog. New CLI commands and an additive municipality lifecycle migration are required; existing region and source lifecycle operations remain reusable. No provider changes are needed. Existing named volumes, credentials and selected images remain attached to their environments.

This change prevents accidental default activation; it does not impose a host-wide mutex or share production state with nonproduction. Explicitly starting multiple workers remains possible and duplicates live work when their sources overlap. Repository implementation and runtime adoption are separate: preserve the current development collector until its collection responsibility is deliberately transferred, and do not activate production as part of this change.
