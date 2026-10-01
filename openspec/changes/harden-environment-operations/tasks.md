# Tasks

All items are new follow-up work from the public-checkout environment review. Checked operator-handoff items record the user's requested ownership transfer, not a completed reboot/backup/restore claim; production still requires actual operational evidence before activation. Repository work (sections 1–5) can be implemented and verified on disposable resources. Host/release work (sections 6–7) requires the relevant infrastructure access and scheduled operational authorization; production activation retains explicit approval of the concrete digest. Private evidence stays outside commits. A missing external prerequisite leaves its item open rather than preventing completion of independent repository fixes.

## 1. Environment identity and restart configuration

- [x] 1.1 Add `unless-stopped` restart policies to PostgreSQL, RustFS and Crawl4AI; verify resolved development, staging and production policies and that default production selects no services.
- [x] 1.2 Resolve wrapper paths without Git and pin each environment's Compose file list; verify invocation from another working directory and an extracted archive uses the correct root, project, secret paths and volumes.
- [x] 1.3 Reject conflicting global project/file/env-file/project-directory options, preserve supported profile/dry-run options and clear inherited overrides; verify short/long/equals forms, unsupported flags and poisoned shell variables cannot redirect staging to development.
- [x] 1.4 Enforce production application digest syntax and reject application builds before image-consuming commands; verify invalid tags/placeholders/malformed digests fail before Docker mutation while configuration, status and teardown remain usable during preparation.

## 2. Secret initialization and smoke targeting

- [x] 2.1 Set intended permissions on newly created secret/configuration descriptors while preserving existing values and custom modes; verify umasks 022/077, directory mode 0700 and repeated initialization with synthetic files.
- [x] 2.2 Require an explicit smoke-test environment and derive configuration, ports, active services and used secret paths through the wrapper; verify ordinary checks issue no stop/start operations and tolerate inactive workers and absent unused provider keys.
- [x] 2.3 Add explicit development/staging fault injection with target-identity checks and same-target recovery, and reject production interruption; verify injected assertion failures still attempt RustFS recovery and no other project's service is touched.

## 3. Release preservation and truthful diagnostics

- [x] 3.1 Add the staging overlay removing application builds and wire it into the wrapper/template checks; verify dependency startup, admin migration/storage initialization and listener startup reuse the candidate image without changing its ID or revision label.
- [x] 3.2 Extend release publication to compare the configured and running staging application image identities/revisions against the local clean-revision candidate; verify matching images reach the publish step and stale/rebuilt/mismatched images fail before a push.
- [x] 3.3 Extend deployment validation with explicit local-environment and nondisruptive runtime modes; verify template-only output is truthful, preparation is distinguished from deployment readiness, runtime mounts/ports/restart policies are checked, revision differences are reported and synthetic secrets never appear in output.

## 4. Fresh-download and operator documentation

- [x] 4.1 Align README, documentation index and developer/environment guides to one-host development/staging/prepared production; provide standalone directory/configuration/secret setup, fresh-versus-existing installation handling, prerequisites and tested Compose compatibility, and verify all links and example paths agree.
- [x] 4.2 Provide explicit environment-specific dependency/migration/storage/readiness commands, provider-key paths and ports, plus status/logs/check/stop/start commands; verify a newcomer can execute each procedure without translating another environment's commands or requiring an inference key for initial development.
- [x] 4.3 Document build-once staging validation, approved-digest production preparation/update, planned interruption, profile selection behavior and compatible-image rollback; deliver a controlled synthetic fixture procedure and verify release examples contain no staging/production application rebuild or implicit data/source promotion.

## 5. Repository regression and isolated acceptance

- [x] 5.1 Add focused script regression tests for identity overrides, production rejection, secret umasks, nondisruptive smoke behavior/recovery and release image mismatches; wire them and updated Compose template checks into CI and verify the suite passes without operational credentials or service interruption.
- [x] 5.2 Rehearse the documented fresh Git-checkout sequence on disposable resources with synthetic fixtures and no provider/notification calls; verify both listeners, API/MCP/admin separation, secret readability by application UID, independent environment resources and inactive production, retaining a redistributable verification summary.
- [x] 5.3 Rehearse basic setup/management from an extracted repository archive and release commands' Git-prerequisite errors; verify missing directories are created by documented steps and no command silently targets an operational installation.
- [x] 5.4 Operator handoff completed at the user's request on 2026-10-01; execution/results are operator-owned and not independently verified here. Original verification: Reboot a disposable VM containing the activated test stack and an inactive production project; verify automatic dependency/listener recovery, retained database/object references and continued production inactivity without manual container starts.

Repository verification results are recorded in [the disposable rehearsal summary](../../../docs/operations/environment-verification.md).

## 6. Current-host capacity, recovery and policy rollout

- [ ] 6.1 Private baseline and provisional capacity decision recorded on 2026-10-01: disk exceeds the review threshold; defer activation and schedule builds/test stacks sequentially. Representative concurrent load, build peaks and growth remain unmeasured. Original verification: Measure concurrent workloads, build peaks, RAM/disk headroom and data growth; deliver a capacity/retention/scheduling decision and verify the existing sustained 70% review threshold and OS/transient headroom are addressed before production approval (existing ingestion task 1.8).
- [x] 6.2 Operator handoff completed at the user's request on 2026-10-01; execution/results are operator-owned and not independently verified here. Original verification: Configure and verify off-host PBS coverage for every data/configuration disk, the selected backup schedule/retention and failure/age notifications; verify a recoverable backup exists and an induced failed/aging backup is visible without enabling the application backup worker (existing tasks 8.2 and 12.5).
- [x] 6.3 Operator handoff completed at the user's request on 2026-10-01; execution/results are operator-owned and not independently verified here. Original verification: Restore to an isolated target with outbound activity disabled; time verified API/MCP recovery, test retained originals/versions/evidence and safe incomplete-job recovery, and record private evidence against the six-hour data-loss and one-hour recovery targets (existing task 8.3).
- [ ] 6.4 Schedule current development/staging policy rollout after recovery protection exists; apply restart policies to verified dependency container IDs, compare before/after image IDs, mounts and credential identities, and verify readiness/evidence references without rebuilding application images or replacing volumes.
- [x] 6.5 Operator handoff completed at the user's request on 2026-10-01; execution/results are operator-owned and not independently verified here. Original verification: Perform a separately scheduled current-host restart after the restore check; verify both activated projects recover without manual container starts, retained references remain valid and production is still inactive, recording actual results and any remediation privately.

## 7. Published release and approved production rollout

- [ ] 7.1 Run required release workflows for a clean revision, build once, initialize/validate controlled staging fixtures and publish/test-pull the exact public GHCR digest; verify runtime-to-registry image identity and record staging evidence plus the digest privately (existing task 12.4).
- [ ] 7.2 Prepare the reviewed production configuration and external HTTPS/private admin routes; verify trusted-proxy handling, public API/MCP reachability expectations and admin restriction using an isolated rehearsal, then present the concrete digest/configuration/readiness record for explicit production approval (existing task 10.2).
- [ ] 7.3 After approval and sections 6/7.1–7.2 pass, initialize and start production from the approved digest; verify local/external readiness, API/MCP results, independent volumes/secrets, private administration and backup monitoring, starting a worker only with its separately reviewed provider/source settings.
- [ ] 7.4 Rehearse compatible-image/configuration rollback and verify retained PostgreSQL/RustFS state and document/evidence references remain valid; deliver the private rollback record and confirm source acceptance/public enablement remain separate operator actions (existing task 10.2).

## 8. Progress and completion records

- [x] 8.1 Reconcile these results with the existing ingestion readiness checklist and the public OpenSpec index; verify checked items have corresponding actual evidence, incomplete external work stays open and no unpublished operational evidence enters commits.
- [x] 8.2 Record implementation/validation/operations changes in the changelog and run strict OpenSpec validation plus the required repository checks; verify the staged change has a dated entry before each commit and that repository completion is reported separately from production readiness.

## Validation record — 1 October 2026

Repository checks: environment/changelog regression tests, Compose example validation, disposable development/staging rehearsal, synthetic API/MCP usage and public-query integration tests, Go tests with race detection, vet, command builds, shell/Python syntax, local documentation links, and strict OpenSpec validation passed. Results are summarized in [environment verification](../../../docs/operations/environment-verification.md). A dated changelog entry covers this work; registry publication remains pending. Before each commit, stage the work and run the changelog recorder as required by repository guidance.

Progress: 22 of 28 items checked, comprising 18 verified implementation/record tasks and four explicit operator handoffs. Six operational tasks remain open: capacity completion, current-host restart-policy rollout, public image publication, routing/approval preparation, approved production activation, and rollback rehearsal. Production remains inactive.
