# Public progress snapshot

These checkboxes reflect the private project task list on 30 September 2026. A checked item records project work; it does not by itself establish public source acceptance, public deployment, or that unpublished evidence is available in this repository. Operational notes and private evidence references are maintained separately.

## 1. Archival and queue fencing

- [x] 1.1 Add audited archival migration, atomic cutoff-bounded store operation and admin-only command; verify claim/relaunch/deduplication, running-job rejection, repeat safety and preservation of successful jobs.

## 2. Operational visibility

- [x] 2.1 Exclude archives from operational job/error lists, counts, provider deferrals and active notification failure detection while retaining accounting; validate views and historical totals.

## 3. Deploy and clean existing backlog

- [x] 3.1 Deploy the tested image, stop writers and verify a fresh backup, archive the selected backlog, verify preserved evidence/accounting and provider hold, restart services and record actual counts.

## 4. Reset pending documents at the operator's request

- [x] 4.1 Add audited per-version pending archival, admin-only cutoff command, scheduling/handler guards and pending-view exclusion; validate preserved successful results, accounting, old-version suppression and new-version eligibility.
- [x] 4.2 Deploy, stop writers and back up, archive remaining queue work and all currently pending versions at one cutoff, verify zero old pending work and intact history, then restart and record the reset time.
