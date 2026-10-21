## Why
The operator wants to clear pending and failed jobs while retaining successful results and historical token accounting. Physical deletion would break receipt provenance and allow old evidence to be scheduled again.

## What Changes
- Add audited, terminal logical archival for unsuccessful jobs before an explicit cutoff.
- Exclude archived jobs from claims, relaunch, provider deferrals and operational job/error views.
- Retain original job identities, attempts, consumption and results.
- Deploy and archive the selected existing backlog after a verified backup.

## Capabilities
### New Capabilities
- `job-archival`: Remove unsuccessful work from the operational queue while retaining history and deduplication.

## Impact
Queue schema and store, administrative job projections, notification detection and an admin-only maintenance command. Existing provider holds stay unchanged; new source versions can still create new jobs.

## Operator extension: pending document reset
The operator additionally selected archival of every currently pending document,
retaining originals/accounting and processing only new documents or new versions
from the reset onward. Add an audited per-version archive, hide it from pending
views, and suppress automatic interpretation of archived evidence.
