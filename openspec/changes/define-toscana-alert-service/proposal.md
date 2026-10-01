## Why

Official Toscana warnings and municipal measures are distributed across regional products, municipal publications and explicitly linked external platforms. Developers and citizens using MCP need a free service that preserves those authorities and clearly states what is known, uncertain or no longer verifiable.

## What Changes

- Cover Toscana's seven categories: minor-network hydrogeological/hydraulic risk, main-network hydraulic risk, strong thunderstorms, wind, coastal waves, snow and ice. Include regional vigilance, criticality/alert and monitoring as distinct products.
- Provide regional applicability by municipality and versioned official alert-zone mappings, including municipalities spanning multiple zones.
- Deliver the first operational MVP for Toscana regional products and Calcinaia local communications. For Calcinaia, use the municipal website as the primary publication source, Cittadino Informato as a secondary discovery/coverage diagnostic and the albo only to retrieve acts referenced by municipal notices; secondary-channel gaps do not become primary-source facts or completeness claims. Retain Livorno, Pisa, Pontedera and Cascina as subsequent local rollout targets. Use operator-configured sources and sections; external channels require scoped official referrals. Recognized provenance does not substitute for technical source acceptance.
- Distinguish regional warnings, including municipal republications, from local observations, operational activations and individual measures. Present relevant local provisions prominently, with supporting documents and explicit uncertainty.
- Include necessary attachments, ordinances and scanned PDFs. Classify full content, use bounded evidence-preserving segmentation for long-document extraction, extract evidence-supported measures and link updates automatically, leaving ambiguity explicit. Keep public document metadata and official links visible when interpretation fails. Enable application-mediated retained copies for verified CFR/Regione and eligible direct municipal content, while keeping Cittadino Informato, albo and any restricted content link-only.
- Track provenance, interpretation and updating separately. Default to configurable per-source checks every 10 minutes and an updating-delay threshold of 30 minutes without a complete successful check. These are collection settings, not delivery guarantees.
- Offer equivalent public API and MCP access without registration, initially limited to 120 shared API/MCP requests per IP per 60 seconds and 100 items per page, with configurable limits and consistent paginated views. Bootstrap from the preceding 30 days and explicitly referenced older documents, distinguishing source dates from service knowledge. Retention defaults to three months per acquired version and is configurable, with ongoing/unresolved-measure exceptions.
- Provide private administration through a local listener, SSH tunnel or explicitly configured tailnet-only Tailscale Serve for source configuration, separate collection/public activation, diagnostics, explicit reprocessing and token/cost reporting. Use an existing SMTP relay and operational mailbox for email notifications. Protect the dedicated MVP VM through off-host Proxmox Backup Server backups and require an isolated restore rehearsal; concrete PBS schedule and retention remain deployment configuration. Validate human-reviewed cases and conduct a trial of at least seven days before source acceptance, enabling semantic candidate retrieval while retaining structured/textual fallback and measuring its incremental cost. Provide internal quality monitoring to improve code and processes. Use DPC products for internal comparison only in the first release; no editorial approval or manual correction queue for individual notices.
- Keep development, staging and the prepared production Compose project on the current VM with separate data, secrets and listener ports. Production stays stopped until capacity, off-host backup, isolated recovery and release checks pass. Build and test a release image in staging, publish it to public GHCR, record its immutable digest, and require operator approval of that digest before production deployment. Target at most six hours of lost data and one hour to restore verified public service after a host failure; prove the target with a timed isolated PBS restore rather than assuming it from configuration.

## Internal MVP milestone — authorized 2026-09-18

The operator authorizes an internal MVP on the current dedicated VM (8 vCPU, approximately 8 GiB RAM, 30 GiB disk): import the three CFR products and Calcinaia, require successful previews, then explicitly enable internal collection and interpretation. Disk expansion to 80 GiB, private SMTP, PBS protection/restore and continuous capacity telemetry are deferred for this internal milestone only. Operators inspect errors and capacity manually; retained data has no verified off-host recovery. Internal observation may proceed and retain evidence, but does not certify infrastructure readiness or public acceptance. Tasks 1.8, 8.2, 8.3 and 10.2 remain open and public deployment/enablement still require separate authorization.

## Capabilities

### New Capabilities

- `official-source-ingestion`: Configured official channels, collection, attachments/OCR, source lifecycle, evaluation, internal administration and operational monitoring.
- `toscana-alert-data`: Geographic and temporal applicability, document and measure relationships, uncertainty and retained history.
- `public-alert-access`: Equivalent free API/MCP access for developers and citizens, with attribution, coverage and operational limits.

### Modified Capabilities

None. These are proposed new capabilities; there are no main specifications or implemented application capabilities yet.

## Impact

Use a modular Go application with PostgreSQL for structured data and durable jobs, RustFS S3 for originals, and three isolated Docker Compose projects on the shared VM. Measure and expand host capacity before production activation, apply production container resource ceilings, and protect the whole VM through off-host PBS. Publish API/MCP through an existing HTTPS reverse proxy on a dedicated hostname while administration remains private. Run Crawl4AI internally; initially call Qwen3.8-27B, DeepSeek-OCR-2 and the selected embedding model through Regolo.ai, keeping providers configurable and baseline retrieval operational without embeddings. Requires retained evidence, public API/MCP queries and local administration. The source acceptance criteria apply independently to each declared product and channel; accepted scopes may be enabled progressively, but pending sources remain visible and prevent claiming the complete Toscana-Calcinaia MVP.

The initial MVP planning revision records the agreed scope and executable work breakdown. Research in [docs](../../../docs/README.md) supplies dated evidence, not proof of production readiness. Provider contracts, source access/reuse, regional formats, mappings and infrastructure capacity require verification. Measure trial costs before setting an operating budget. Later release-operation work remains subject to its own validation and operator approval.

## Non-goals

Public DPC warning feeds; local collection for all Toscana municipalities in this release; Firenze as an initial local pilot; fire, heat-health, avalanche and IT-alert products; complete pre-startup historical reconstruction; autonomous safety advice; autonomous discovery of new source channels; publicly accessible administration; per-notice editorial approval. Telegram and other notification channels are deferred. Production activation and source acceptance remain separate operator actions.
