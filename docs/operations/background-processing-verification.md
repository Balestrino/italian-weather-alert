# Background processing controls: repository verification

Verified on 1 October 2026 for [explicit workers and hierarchical controls](../../openspec/changes/require-explicit-background-workers/proposal.md). These results cover repository behavior and synthetic fixtures. They do not establish source acceptance, a production deployment, host capacity or a live collector handover.

## Resolved service selection

`python3 scripts/test_environment_operations.py` passed all 18 cases. `python3 scripts/check-deployment-config.py` passed template validation without starting operational services.

| Selection | Development / staging | Production |
| --- | --- | --- |
| Default | Five core services, no processing or backup | No services |
| Core production profile | Not applicable | Five core services, no processing or backup |
| `processing-worker` | Core plus processing, no backup | No services |
| `production` + `production-worker` | Not applicable | Core plus processing, no backup |
| `application-backup` (plus production core where needed) | Core plus backup, no processing | Core plus backup, no processing |

The regressions also verify explicit service targeting, production profile replacement rather than inheritance, invalid default-enabled profiles and optional-worker absence in runtime checks. All-service topology, isolated storage/secrets/ports and inactive-production checks remain intact.

## Domain, CLI and admin

The complete domain and backoffice integration suites passed against disposable PostgreSQL databases, including an isolated migration/dump/restore rehearsal. Additional integration suites passed for the CLI, acquisition, classification, embedding, extraction, interpretation, jobs, linking, OCR, operations, registry and territorial guards. Fixtures use synthetic identities and evidence; no official-source acquisition, provider invocation or operational notification is required.

The lifecycle cases verify new disabled municipalities, compatibility backfill and repeat-migration preservation, audited changes, stale revisions, invalid membership, register replacement/retired history, row-lock admission, parent override, preserved child/source choices, admitted work completion and no catch-up enqueue. Regional and other municipal sources retain their own eligibility.

The executable and admin are tested against the same fixture registry with the worker absent. Cases cover source activation/suspension/status in both directions, preview and policy prerequisites, incompatible territorial profiles, obsolete drafts, operator identity, private-detail-free nonzero CLI errors and public flag preservation. Controls leave processing jobs, source checks and retained versions empty. Territorial history includes municipal decisions only in the appropriate scope, and municipality reads retain a constant number of queries regardless of page size.

`TestTerritoryDashboardBrowser` and [its browser check](../../scripts/check-territories.cjs) passed with JavaScript disabled, keyboard submission, a 390px viewport and 200% rendering zoom. Native region/municipality actions retain filters/pagination, display the parent blocking reason and preserve a municipality's saved choice after region disablement. The fixture server binds to loopback and blocks requests to other origins.

To reproduce browser validation, install Playwright with Chromium separately and supply its paths:

```sh
IWA_BROWSER_NODE="<node-executable>" IWA_PLAYWRIGHT_MODULE="<playwright-module-path>" \
  go test -tags=integration ./internal/backoffice -run '^TestTerritoryDashboardBrowser$' -count=1
```

## General checks and operational boundary

`go test -race ./...`, `go vet ./...`, `go build ./cmd/...`, strict OpenSpec validation and local documentation-link/whitespace checks passed. `govulncheck@v1.8.0` found no reachable vulnerabilities; it also reported vulnerabilities in imported packages/modules that the code does not call.

No operational worker was stopped, started or recreated for this change. Profiles do not stop an existing collector. Apply the [adoption and handover procedure](environments.md#adoption-and-collector-handover) only during the separately scheduled rollout, preserving the current collector until the reviewed production collector is ready. An older worker image that ignores municipal disablement cannot safely process after those controls are in use; preserve additive data and keep processing stopped until a compatible rollback image is selected.
