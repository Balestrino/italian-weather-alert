# OpenSpec: project plans and progress

The 3 October 2026 planning revision formally includes Cittadino Informato as an
additional institutional acquisition/discovery channel for selected municipalities,
with primary municipal/CFR verification and independent primary collection. Task
26.1 records the documentation update. The subsequent user revision replaces the
license/legal-basis documentation gate with a working multi-source verification
system: implement acquisition and persistent comparisons in 26.3/26.4, then close
26.2 from a bounded development run, now verified with ten persisted receipts.
The later 26.5 evaluation and bounded 26.6 development trial are verified; scheduled adoption remains separate. The decision activates
no platform collector and does not complete source acceptance. See
[platform guidance](../docs/fonti/piattaforme/cittadino-informato.md).

A subsequent bounded manual review identifies Calcinaia's exposed REST index,
declared notice routes and the limits of WordPress page-container metadata.
The [review dossier](../docs/fonti/piattaforme/cittadino-informato-review.md)
preserves the initial review and revised technical closure criteria. Task 26.2
is now completed by a bounded acquisition-to-verification run on development
PostgreSQL/RustFS with dedicated private data, real-source diagnostics and
separately labelled synthetic controls; existing services/source controls remain
preserved. Task 26.3 implements scoped API acquisition, verified
with synthetic persistence cases and two bounded isolated Calcinaia HTTP checks;
no live collector was activated. Task 26.4 now implements immutable evidence-bound
receipts, primary-only projection and historical/current deduplication, verified
with synthetic disposable PostgreSQL and public-query cases. The 26.2 run verifies
receipt readback, history/idempotence, primary-only local admission and diagnostic regional/non-comparable candidates. A wider platform
window with an external municipal PDF remains explicitly incomplete. Counterpart discovery and scheduled adoption remain separate. The subsequent 26.5/26.6 work verifies reviewed real-source/API-MCP attribution and bounded scoped visibility with rollback. See [operations](../docs/operations/cittadino-informato.md).

This directory publishes IWA's change proposals, designs, capability specifications, and task checklists so contributors can see the intended behavior and work still open. Start with a change's `proposal.md`, then read its `design.md`, `specs/`, and `tasks.md`.

The original checklists are a **30 September 2026 snapshot** of work tracked in the private project. Later public-checkout tasks and changes are marked in their own files. A checked task does not certify public source acceptance, deployment, or that all supporting evidence is redistributable. The public [coverage tracker](../docs/coverage.md) describes source status. Detailed operational notes and research evidence remain separate while their contents and reuse terms are reviewed.

| Change | Done | Open |
| --- | ---: | ---: |
| [Application separation and public status frontend](changes/reorganize-application-surfaces/proposal.md) | 6 | 0 |
| [Environment operations hardening](changes/harden-environment-operations/proposal.md) | 24 | 6 |
| [Explicit workers and hierarchical CLI/admin territorial controls](changes/require-explicit-background-workers/proposal.md) | 16 | 0 |
| [Toscana alert service](changes/define-toscana-alert-service/proposal.md) | 159 | 7 |
| [Admin dashboard](changes/modernize-admin-dashboard/proposal.md) | 21 | 0 |
| [Territorial administration](changes/organize-admin-by-territory/proposal.md) | 20 | 0 |
| [Automatic changelog](changes/automate-changelog/proposal.md) | 16 | 0 |
| [Evening alert email report](changes/add-evening-alert-email-report/proposal.md) | 7 | 1 |
| [Unsuccessful job archival](changes/archive-unsuccessful-jobs/proposal.md) | 5 | 0 |
| [Inference efficiency](changes/reduce-regolo-token-waste/proposal.md) | 29 | 3 |

The Toscana service checklist includes seven subsequent development recovery tasks, verified against acquisition, interpretation canaries, queue progress and runtime checks. See the [recovery procedure](../docs/operations/acquisition-recovery.md). Five subsequent local fallback tasks verify opt-in Qwen chat/image tests and scoped development routing; three further tasks verify explicit register-wide fallback selection, worker configuration and local OCR progress. Three further two-request trial tasks verify server slots, concurrent development worker calls and stored OCR results, with documented scale-down and private rollback evidence. Three subsequent tasks verify expansion to four worker replicas, fix reproduced pre-claim deadlocks and validate concurrent processing on the corrected development image while distinguishing worker count from available server slots. One further task verifies an operator-requested six-worker expansion, preserved existing replicas and queue progress, with server/model concurrency reported separately. See [local fallback operations](../docs/operations/local-llm-fallback.md). Provider entitlement restoration is subsequently verified below; dependent semantic embeddings remain separate unverified work. These checks do not complete source acceptance or production readiness.

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

Four subsequent embedding-control tasks, verified on 2 October 2026, add disabled
global and per-source flags, audited CLI/admin actions and live worker admission.
Synthetic PostgreSQL checks verify scheduling, paused attempts, source-choice
retention and lazy provider setup; native form integration and a 390px development
page check verify the admin surface. Development adoption replaced the admin and
all processing replicas with every flag disabled, retired only the requested
pending embedding work and passed nondisruptive environment smoke checks.
See [embedding operations](../docs/operations/embedding-controls.md); private
adoption evidence does not establish provider availability or source acceptance.

Environment hardening has 18 verified repository tasks and four checked operator handoffs for backup/restore and VM checks; the handoffs do not certify their results. See the [disposable verification summary](../docs/operations/environment-verification.md).

Two subsequent development API-binding tasks publish the public listener on
`0.0.0.0`, preserve loopback administration and release-environment ports, and
verify the targeted existing-image rollout with unchanged other containers.
Focused configuration/environment tests, template checks, local-interface HTTP
and API/MCP comparisons passed; remote-machine firewall reachability and
production readiness remain separate from these checks.

The October 2 operations overview extension has synthetic PostgreSQL and browser verification for complete job states, archived scopes, exclusive document categories and historical attempt bars. The development admin rollout is recorded separately from source acceptance; provider state and last job reasons remain distinct. See [backoffice semantics and verification](../docs/backoffice/README.md).

Open tasks include environment capacity/policy rollout and production release gates, the Toscana observational trial and public readiness gates, GHCR image publication, off-host backup and restore verification, the evening report rollout, and parts of the provider efficiency rollout. See each `tasks.md` for the precise checklist.

Three subsequent acceptance-precheck tasks inventory source-scoped prerequisites,
persist both existing campaign assessments and verify lifecycle/query behavior
with synthetic fixtures and disposable PostgreSQL. Both campaigns remain extended;
missing reviewed evidence and the municipal regression keep source acceptance and
public enablement pending. The [coverage tracker](../docs/coverage.md) retains its
original snapshot with a dated precheck note; detailed evidence remains private.

For new work, propose a change through an issue or pull request and update the relevant specification and checklist alongside code. Keep publication status and evidence limitations explicit. The repository includes [OpenSpec agent workflows](../docs/development/agent-tools.md); an agent or CLI is not required to inspect the plans.

The application-separation change reorganizes implementation and guide locations. Historical snapshot prose can retain earlier source paths; use the [architecture mapping](../docs/architecture/README.md) to locate their current equivalents. Snapshot checks do not certify the current checkout.

Two subsequent release-preparation tasks align OpenAPI/MCP metadata to 0.2.0,
preserve dated history and historical releases, and verify changelog regressions,
public transport tests and local links. The [minor release notes](../docs/releases/v0.2.0.md)
describe embedding defaults, worker replacement and pilot limits. Promotion and
software publication retain the exact-revision workflow gates independently of
container publication and production activation.
One further task fixes the reproduced CPU-dependent concurrent maintenance test
fixture, verified with two-CPU affinity and full processing integration/race checks.

Three subsequent implementation tasks add manually selected municipality publication
in explicit development, with immutable private controls, scoped API/MCP visibility
and revocable saved views. Synthetic PostgreSQL, native-form, API/MCP and resolved
Compose checks verify same-zone isolation and strict staging/production behavior.
See [development publication](../docs/operations/development-publication.md).
Source acceptance remains independently pending. Task 25.4 verifies the targeted
development public/admin rollout, native-form activation and live API/MCP equivalence;
workers and other existing services were preserved.

The 3 October delegated document review completes the remaining review method
requested by the operator; it is attributed to the assistant rather than recorded
as new human approval. Tasks 26.5 and 26.6 are verified with public receipt schemas,
synthetic PostgreSQL tests and a bounded isolated development trial using a real
MCP client, independent acquisition and visibility revocation. Source acceptance
remains pending: a current municipal completeness failure, live generic projection
limits and regional graphical interpretation are explicitly retained. See
[the current procedure](../docs/operations/cittadino-informato.md#valutazione-e-trial-delimitato--task-265266).


The later 3 October live repair connects ordinary primary municipal extraction
and supported update links to the domain, with field evidence, literal temporal
ambiguity and immutable knowledge boundaries. A bounded retained-data replay and
actual HTTPS API/MCP checks verify development behavior, provenance history,
pagination and municipality isolation. Public/admin and six workers run the fix;
worker catalog collisions preserve historical versions. Tasks 27.1–27.3 are
verified; graphical CFR interpretation in 27.4, municipal completeness and source
acceptance remain open. See [development operations](../docs/operations/development-publication.md).


Task 27.5 verifies the current Calcinaia criticality response from matching retained
PDF vector maps, with seven risks for both days, precise table validity, PDF/page
attribution, historical preservation and actual HTTPS API/MCP equivalence. No
regional level remains unknown for A4 in that verified response. Other unsupported
label positions and vigilance graphics remain part of 27.4; municipal temporal
conflicts and source acceptance are not completed by this correction.

Task 27.4 is subsequently verified on 4 October: product-specific criticality
and vigilance graphics cover all 26 zones and seven risks in four retained
samples. Applicability, daily/cumulative rainfall, symbols, PDF physical pages,
knowledge history and API/MCP weather metadata have retained/synthetic/disposable
verification. Positive wind/sea/snow/ice cases are synthetic; further real cases
remain acceptance evidence. This supersedes the graphical software gap above in
the recognized formats, without replacing live services or accepting sources.
See [CFR graphics](../docs/operations/cfr-graphics.md).

Tasks 28.1–28.3 replace the raw municipality situation presentation with processed
conclusions and separate regional, municipal and Cittadino Informato summaries.
Synthetic/PostgreSQL, schema, real API/MCP, pagination, historical and development
rollout checks preserve evidence and affected uncertainty while removing the
document backlog from public situation rows. See [response fields](../docs/backend/public-api-mcp.md#situation-response).
The change does not complete source acceptance or continuous platform collection.

Two subsequent 0.3.0 release-preparation tasks align OpenAPI/MCP metadata, group
the preserved dated history and document situation consumer migration, verification
upgrades and source limits. Changelog, public transports, release-note links and
strict specification checks pass. See [v0.3.0 notes](../docs/releases/v0.3.0.md);
software publication uses the exact-revision workflow gates independently of
image publication, continuous source collection or production activation.


The 4 October follow-up completes trial accounting (9.4) and independent municipal
gates/reports (11.5): costs keep missing metrics and historical prices explicit;
Livorno, Pisa, Pontedera and Cascina remain pending. Individual campaigns retain
the pilot's observation requirements and preserve historical MVP campaigns.
See [cost summaries](../docs/operations/trial-costs.md) and
[municipal acceptance](../docs/operations/municipal-acceptance.md).
The [PBS runbook](../docs/operations/pbs-recovery.md) is prepared; actual protection,
alerts, restore and production readiness remain unverified.

Task 13.8 is subsequently verified with restored authenticated provider access,
controlled model probes, retained remote OCR and an ordinary municipal
classification/extraction/linking pipeline. An invalid-quotation attempt remains
failed; embeddings remain disabled and untested by this recovery. See
[provider recovery](../docs/operations/acquisition-recovery.md#verifica-del-recupero-provider--4-ottobre-2026)
and the [eight remaining gates](../docs/operations/toscana-readiness.md).

The subsequent development adoption aligns public/admin, six workers and the
existing application backup service to the clean verified revision. Runtime image
checks, HTTP isolation, private campaign gates and actual API/MCP saved-view
comparisons pass. Source controls and other containers/volumes are preserved;
production stays stopped. See [adoption and remaining gates](../docs/operations/toscana-readiness.md#adozione-in-development).

The subsequent 4 October retained-response verification corrects the partial
municipal reopening merge: original typographic place spelling and a bounded
explicit reopening preserve independent restrictions and undetermined times.
Synthetic/race and disposable PostgreSQL checks pass with immutable extraction
v25 settings. See [municipal interpretation](../docs/operations/municipal-interpretation.md).
This advances 9.3/10.1 without closing them: ordinary live adoption/reprocessing,
same-contract evaluation and remaining campaign evidence are still required.
The checklist remains 153/161 with eight open gates.

The 4 October continuation persists a successful rerun of the same fourteen-check
municipal completeness contract, preserving the previous failure and explicitly
identifying retained provider responses. Tasks 29.1/29.2 add and verify exact-version
archive recovery with immutable audit and bounded reprocessing/descendant admission;
ordinary archive guards remain effective. Task 29.3 subsequently verifies development
adoption, selected ordinary workers, preserved archives/results and knowledge
boundaries, actual API/MCP, pagination and municipality isolation. The fresh
extraction preserves reopening/exception but omits three independent prohibition
scopes: a failed same-contract regression preserves and supersedes the replay
success. Both reassessed campaigns remain extended; eight original gates stay open.
See [municipal interpretation](../docs/operations/municipal-interpretation.md).

Task 9.3 is subsequently verified by an immutable `complete` MVP assessment:
239.94 paced hours, all four products, daily section-scoped checks, attributable
original/attachment corrections and explicitly labelled real/retained failure and
absent-event evidence. The earlier internal campaign keeps its historical delay
breaches. Extraction v27 and selected ordinary reprocessing pass the same reviewed
contract and actual API/MCP/history checks; original failures and archives remain.
The Toscana checklist is now 159/166, with seven open gates. Source acceptance,
10.1 and production readiness remain separate. See
[observational trial](../docs/operations/observational-trial.md).
