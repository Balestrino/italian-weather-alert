# OpenSpec: project plans and progress

This directory publishes IWA's change proposals, designs, capability specifications, and task checklists so contributors can see the intended behavior and work still open. Start with a change's `proposal.md`, then read its `design.md`, `specs/`, and `tasks.md`.

The original checklists are a **30 September 2026 snapshot** of work tracked in the private project. Later public-checkout tasks and changes are marked in their own files. A checked task does not certify public source acceptance, deployment, or that all supporting evidence is redistributable. The public [coverage tracker](../docs/coverage.md) describes source status. Detailed operational notes and research evidence remain separate while their contents and reuse terms are reviewed.

| Change | Done | Open |
| --- | ---: | ---: |
| [Application separation and public status frontend](changes/reorganize-application-surfaces/proposal.md) | 6 | 0 |
| [Environment operations hardening](changes/harden-environment-operations/proposal.md) | 22 | 6 |
| [Explicit workers and hierarchical CLI/admin territorial controls](changes/require-explicit-background-workers/proposal.md) | 12 | 0 |
| [Toscana alert service](changes/define-toscana-alert-service/proposal.md) | 127 | 11 |
| [Admin dashboard](changes/modernize-admin-dashboard/proposal.md) | 21 | 0 |
| [Territorial administration](changes/organize-admin-by-territory/proposal.md) | 20 | 0 |
| [Automatic changelog](changes/automate-changelog/proposal.md) | 11 | 0 |
| [Evening alert email report](changes/add-evening-alert-email-report/proposal.md) | 7 | 1 |
| [Unsuccessful job archival](changes/archive-unsuccessful-jobs/proposal.md) | 5 | 0 |
| [Inference efficiency](changes/reduce-regolo-token-waste/proposal.md) | 29 | 3 |

The Toscana service checklist includes seven subsequent development recovery tasks, verified against acquisition, interpretation canaries, queue progress and runtime checks. See the [recovery procedure](../docs/operations/acquisition-recovery.md). Five subsequent local fallback tasks verify opt-in Qwen chat/image tests and scoped development routing; three further tasks verify explicit register-wide fallback selection, worker configuration and local OCR progress. Three further two-request trial tasks verify server slots, concurrent development worker calls and stored OCR results, with documented scale-down and private rollback evidence. Three subsequent tasks verify expansion to four worker replicas, fix reproduced pre-claim deadlocks and validate concurrent processing on the corrected development image while distinguishing worker count from available server slots. See [local fallback operations](../docs/operations/local-llm-fallback.md). Provider entitlement restoration and dependent semantic embeddings remain separate open work. These checks do not complete source acceptance or production readiness.

Three subsequent documentation tasks add [territorial source guides](../docs/territori/README.md), a shared template, scoped platform notes and contributor upkeep instructions. Public findings and private operational evidence remain separate; document verification does not resolve the recorded acquisition problems or change the coverage snapshot.

Five subsequent municipal recheck tasks preserve diagnostic evidence and scoped download observations. Five further attachment tasks add revision-scoped external PDF permissions, content filtering and preview/collector validation, with isolated persistence tests and a verified development Livorno rollout. See [municipal attachment operations](../docs/operations/municipal-attachments.md). Five further Cascina tasks verify listing dates, its separately reviewed PDF scope, CSRF interpretation equivalence and retirement of only reviewed stale unattempted work in development. Isolated tests verify validated reuse without a second call; ordinary checks verify unchanged-version retention. These checks do not complete source acceptance or establish future availability.

Five subsequent local-processing tasks add reviewed article-body selectors,
text-only PDF native extraction and positive-only regional-format classification.
Synthetic tests, disposable PostgreSQL and real Poppler checks verify fallbacks,
page provenance, zero-call accounting and policy-separated reuse. See
[local processing operations](../docs/operations/local-processing.md). No live
source policy was activated by those implementation tasks. Six subsequent
verification tasks add a fifty-version read-only replay and an eight-version
development trial. All sixteen trial jobs succeeded; four municipal comparisons
agreed with 36.8% fewer reported input tokens, while regional classifications
and reviewed native PDFs made no model calls. Trial results are retained for
backoffice inspection, and the separate worker exited. Continuing policies,
semantic downstream quality and acceptance remain separate.

Explicit background selection and hierarchical controls have synthetic repository verification for CLI/admin parity, municipal migration and processing admission, plus native-form browser checks. See [background processing verification](../docs/operations/background-processing-verification.md). The current live collector was preserved; production handover remains a separate release operation.

Environment hardening has 18 verified repository tasks and four checked operator handoffs for backup/restore and VM checks; the handoffs do not certify their results. See the [disposable verification summary](../docs/operations/environment-verification.md).

The October 2 operations overview extension has synthetic PostgreSQL and browser verification for complete job states, archived scopes, exclusive document categories and historical attempt bars. The development admin rollout is recorded separately from source acceptance; provider state and last job reasons remain distinct. See [backoffice semantics and verification](../docs/backoffice/README.md).

Open tasks include environment capacity/policy rollout and production release gates, the Toscana observational trial and public readiness gates, GHCR image publication, off-host backup and restore verification, the evening report rollout, and parts of the provider efficiency rollout. See each `tasks.md` for the precise checklist.

For new work, propose a change through an issue or pull request and update the relevant specification and checklist alongside code. Keep publication status and evidence limitations explicit. The repository includes [OpenSpec agent workflows](../docs/development/agent-tools.md); an agent or CLI is not required to inspect the plans.

The application-separation change reorganizes implementation and guide locations. Historical snapshot prose can retain earlier source paths; use the [architecture mapping](../docs/architecture/README.md) to locate their current equivalents. Snapshot checks do not certify the current checkout.
