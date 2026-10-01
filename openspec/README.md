# OpenSpec: project plans and progress

This directory publishes IWA's change proposals, designs, capability specifications, and task checklists so contributors can see the intended behavior and work still open. Start with a change's `proposal.md`, then read its `design.md`, `specs/`, and `tasks.md`.

The original checklists are a **30 September 2026 snapshot** of work tracked in the private project. Later public-checkout tasks and changes are marked in their own files. A checked task does not certify public source acceptance, deployment, or that all supporting evidence is redistributable. The public [coverage tracker](../docs/coverage.md) describes source status. Detailed operational notes and research evidence remain separate while their contents and reuse terms are reviewed.

| Change | Done | Open |
| --- | ---: | ---: |
| [Application separation and public status frontend](changes/reorganize-application-surfaces/proposal.md) | 6 | 0 |
| [Environment operations hardening](changes/harden-environment-operations/proposal.md) | 22 | 6 |
| [Explicit workers and hierarchical CLI/admin territorial controls](changes/require-explicit-background-workers/proposal.md) | 6 | 6 |
| [Toscana alert service](changes/define-toscana-alert-service/proposal.md) | 89 | 11 |
| [Admin dashboard](changes/modernize-admin-dashboard/proposal.md) | 16 | 0 |
| [Territorial administration](changes/organize-admin-by-territory/proposal.md) | 18 | 0 |
| [Automatic changelog](changes/automate-changelog/proposal.md) | 6 | 0 |
| [Evening alert email report](changes/add-evening-alert-email-report/proposal.md) | 7 | 1 |
| [Unsuccessful job archival](changes/archive-unsuccessful-jobs/proposal.md) | 5 | 0 |
| [Inference efficiency](changes/reduce-regolo-token-waste/proposal.md) | 29 | 3 |

The Toscana service checklist includes seven subsequent development recovery tasks, verified against acquisition, interpretation canaries, queue progress and runtime checks. See the [recovery procedure](../docs/operations/acquisition-recovery.md). Five subsequent local fallback tasks verify opt-in Qwen chat/image tests and scoped development routing; see [local fallback operations](../docs/operations/local-llm-fallback.md). Provider entitlement restoration and dependent semantic embeddings remain separate open work. These checks do not complete source acceptance or production readiness.

Environment hardening has 18 verified repository tasks and four checked operator handoffs for backup/restore and VM checks; the handoffs do not certify their results. See the [disposable verification summary](../docs/operations/environment-verification.md).

Open tasks include environment capacity/policy rollout and production release gates, the Toscana observational trial and public readiness gates, GHCR image publication, off-host backup and restore verification, the evening report rollout, and parts of the provider efficiency rollout. See each `tasks.md` for the precise checklist.

For new work, propose a change through an issue or pull request and update the relevant specification and checklist alongside code. Keep publication status and evidence limitations explicit. The repository includes [OpenSpec agent workflows](../docs/development/agent-tools.md); an agent or CLI is not required to inspect the plans.

The application-separation change reorganizes implementation and guide locations. Historical snapshot prose can retain earlier source paths; use the [architecture mapping](../docs/architecture/README.md) to locate their current equivalents. Snapshot checks do not certify the current checkout.
