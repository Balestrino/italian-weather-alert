# Documentation

Italian Weather Alert is an independent, best-effort project to make Italian regional weather and hydrogeological alerts and related municipal notices easier to find. It is not an official alert source. Check the issuing authorities for current warnings and instructions.

- [Coverage tracker](coverage.md): all regions and municipalities, with documented implementation and source-check status. It is not a live alert map.
- [Territorial source guides](territori/README.md): current collection guidance, scoped exceptions, open problems and dated discoveries for investigated regions and municipalities, with shared platform notes.
- [Architecture](architecture/README.md): component responsibilities, dependency boundaries and the old-to-new layout mapping.
- [Backend](backend/README.md): data processing, public API/MCP and contracts.
- [Backoffice](backoffice/README.md): private territorial administration and operator actions.
- [Public frontend](frontend/README.md): independently runnable system-status page and optional Compose project.
- [Developer setup](development/setup.md): fresh development installation, optional worker setup and checks.
- [Environments](operations/environments.md): one-host projects, fresh/existing setup, routine management and data separation.
- [Environment verification](operations/environment-verification.md): disposable rehearsal results and operational checks still outstanding.
- [Release operations](operations/releases.md): image approval, deployment checks and PBS recovery targets.
- [Reviewed municipal PDF attachments](operations/municipal-attachments.md): source/revision boundaries, content filtering, PDF validation, preview and verified retention.
- [Acquisition recovery](operations/acquisition-recovery.md): stale municipal URLs, temporary interpretation limits and bounded development validation.
- [Optional local Qwen fallback](operations/local-llm-fallback.md): real model tests, scoped routing, OCR capability checks and rollback.
- [OpenSpec plans and progress](../openspec/README.md): change proposals, capability specifications, and dated task checklists.
- [Agent tools](development/agent-tools.md): included OpenSpec skills and optional local integrations.
- [Public API/MCP contract](backend/public-api-mcp.md) and [usage examples](backend/public-usage.md).
- [Third-party notices](../THIRD_PARTY_NOTICES.md).

The public code repository includes only source material selected for redistribution. Additional research evidence and internal operational records are kept separately.
