# OpenSpec: project plans and progress

This directory publishes IWA's change proposals, designs, capability specifications, and task checklists so contributors can see the intended behavior and work still open. Start with a change's `proposal.md`, then read its `design.md`, `specs/`, and `tasks.md`.

The original checklists are a **30 September 2026 snapshot** of work tracked in the private project. Later public-checkout tasks and changes are marked in their own files. A checked task does not certify public source acceptance, deployment, or that all supporting evidence is redistributable. The public [coverage tracker](../docs/coverage.md) describes source status. Detailed operational notes and research evidence remain separate while their contents and reuse terms are reviewed.

| Change | Done | Open |
| --- | ---: | ---: |
| [Environment operations hardening](changes/harden-environment-operations/proposal.md) | 22 | 6 |
| [Toscana alert service](changes/define-toscana-alert-service/proposal.md) | 77 | 10 |
| [Admin dashboard](changes/modernize-admin-dashboard/proposal.md) | 16 | 0 |
| [Territorial administration](changes/organize-admin-by-territory/proposal.md) | 18 | 0 |
| [Automatic changelog](changes/automate-changelog/proposal.md) | 4 | 0 |
| [Evening alert email report](changes/add-evening-alert-email-report/proposal.md) | 7 | 1 |
| [Unsuccessful job archival](changes/archive-unsuccessful-jobs/proposal.md) | 5 | 0 |
| [Inference efficiency](changes/reduce-regolo-token-waste/proposal.md) | 29 | 3 |

Environment hardening has 18 verified repository tasks and four checked operator handoffs for backup/restore and VM checks; the handoffs do not certify their results. See the [disposable verification summary](../docs/environment-verification.md).

Open tasks include environment capacity/policy rollout and production release gates, the Toscana observational trial and public readiness gates, GHCR image publication, off-host backup and restore verification, the evening report rollout, and parts of the provider efficiency rollout. See each `tasks.md` for the precise checklist.

For new work, propose a change through an issue or pull request and update the relevant specification and checklist alongside code. Keep publication status and evidence limitations explicit. The repository includes [OpenSpec agent workflows](../docs/agent-tools.md); an agent or CLI is not required to inspect the plans.
