## Context
The user explicitly selected archival rather than deletion. At initial inspection there were 4,984 failed jobs, 1,140 pending/retry jobs, and 1,134 successful jobs. Counters may advance until the worker is stopped. A verified database backup is stored privately under bin/before-job-purge-20260923T062131Z.

## Goals / Non-Goals
Archive the selected backlog without changing successful jobs, receipts, attempts or retained content. Hide archived jobs from operational lists and failed-job counts. Preserve immutable keys to prevent automatic recreation. Do not erase real source/quality failures, resume provider calls, or pause future acquisition permanently.

## Decisions
Use nullable archived_at and archived_by on processing_jobs, constrained to unsuccessful non-running states. A database guard freezes archived rows and rejects accidental revival even by an old binary. A transaction takes a queue table lock, rejects running work in the selected cutoff, and archives queued/retry/failed rows. The command is admin-only and cutoff-bounded. Enqueue continues returning the original identity; claim/relaunch/deferral explicitly exclude archived rows. Default job listing shows all active jobs (after cleanup only succeeded jobs remain), and failed-job views exclude archives. Processing accounting queries retain all attempts.

Notification detection ignores archived failures but preserves incidents as history; archival must not generate a false processing-success assertion or send unsolicited notifications. Deployment stops worker and backup, verifies a fresh private DB backup, migrates and archives, then resumes services. New ordinary jobs may subsequently appear.

## Risks / Trade-offs
Archival is logical, not disk reclamation. An old worker cannot process archived rows but may error on its database guard; rollback must keep workers stopped. Source HTTP 404 and provider holds remain separately visible. Running jobs abort archival rather than being mislabeled.

## Migration Plan
Add migration 057. Validate queue behavior, admin views and historical token totals on isolated PostgreSQL. Deploy a pinned image, stop writers, backup, run the cutoff-bounded command, verify fingerprints and zero pending/failed unarchived rows, and restart services with provider hold unchanged.

## Open Questions
None; the operator selected preservation of history and prevention of old-job replay.

## Pending document reset
Migration 058 adds interpretation_archives keyed by retained version, with actor,
cutoff and archival timestamp. Successful interpretations are excluded. Original
versions and every result/attempt remain unchanged. The command requires all old
unsuccessful queue jobs to be archived first and takes the same queue lock, so
processing cannot race the operation. Deployment stops workers and uses a fresh
backup and cutoff. Default pending-document projections exclude archived versions;
source suspension/quality evidence remains independently visible. Scheduling and
the interpretation handler guard suppress archived versions; a newly acquired
version has a new identity and is eligible normally. Raw retention can eventually
delete an archive via FK cascade, just as it deletes the associated version.
This is a snapshot reset, not a publication-date filter or a claim that archived
documents have valid interpretations. The provider hold remains unchanged.
