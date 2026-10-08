## Purpose

Provide developers and citizens using assistants with equivalent free API and MCP access to attributable Toscana warnings, local measures and history.

## ADDED Requirements

### Requirement: Current alert summary agrees with detailed facts
Situation summaries SHALL identify supported non-green current criticality levels and their risk/zone scope even when overall interpretation is partial. Reliability limitations SHALL NOT turn a supported orange or yellow value into an unknown level. Future and expired intervals SHALL remain distinct from current warnings in API and MCP.

#### Scenario: Orange is supported by partial HTML projection
- **WHEN** the current detailed criticality fact is orange with a supported explicit interval and other graphical values remain unresolved
- **THEN** the summary states the orange alert and the remaining partial coverage separately, with the same evidence and temporal boundary in API and MCP

### Requirement: Public access without registration
The first release SHALL provide a documented read-only public JSON API and remote MCP access without payment, registration or a required API key. Both interfaces SHALL support municipality/zone discovery, warning and measure search, document details and versions, retained history, and source coverage and updating information. They SHALL expose only regional and recognized local public products; internal DPC comparisons and monitoring administration SHALL NOT appear as public warning records or public administrative operations.

#### Scenario: API documentation is opened from a private remote browser
- **WHEN** an operator requests Scalar access from a browser that cannot reach the VM's loopback public listener
- **THEN** a dedicated private proxy may expose that public listener and its same-origin documentation/contract while preserving existing administrative routes; browser verification confirms the endpoint reference renders and records targeted proxy removal separately from public source acceptance

#### Scenario: An anonymous client requests a municipality
- **WHEN** a client submits a supported public query without credentials
- **THEN** the service returns the permitted data subject to published usage limits without requiring registration

### Requirement: Municipality selection and geographic scope
Both interfaces SHALL support lookup by municipality name or ISTAT code. Postal-code lookup SHALL return attributable municipality candidates and allow municipality selection before querying municipal status; unavailable postal mappings SHALL be explicit. Ambiguous municipality names SHALL NOT silently select a municipality. Regional applicability SHALL cover verified Toscana mappings; the first operational MVP SHALL identify Calcinaia as its local pilot and report its independently verified coverage. Livorno, Pisa, Pontedera and Cascina SHALL be identified as subsequent planned local scopes until individually activated. Other Toscana municipalities SHALL retain regional applicability without implying local collection. Known areas outside Toscana SHALL return explicit unsupported-area information.

#### Scenario: A postal code maps to multiple municipalities
- **WHEN** a client searches that postal code
- **THEN** the service returns candidates and requires a selected municipality for its municipal-status query

#### Scenario: No usable postal-code dataset is available
- **WHEN** a client requests postal-code lookup and no attributable dataset with verified permitted use is configured
- **THEN** both interfaces explicitly report postal mapping unavailable, return no invented candidates, and retain municipality-name and ISTAT lookup plus independently verified municipality-zone applicability

#### Scenario: A municipality without activated local collection is requested
- **WHEN** verified regional mapping exists for that Toscana municipality
- **THEN** applicable regional products are available and local collection is explicitly pending or outside the planned local scope, as applicable

### Requirement: Equivalent evidence and historical queries
Equivalent API and MCP queries against the same data version and evaluation time SHALL return the same facts, document/measure identities, provenance, interpretation limitations, updating status and coverage. Both interfaces SHALL expose current information, future already-issued warnings and retained history distinctly. History SHALL support distinguishing source chronology from service acquisition/interpretation chronology and disclose earliest available data and gaps. Requests beyond retained history SHALL NOT imply there were no warnings.

#### Scenario: A developer and an assistant request the same retained state
- **WHEN** their filters, evaluation time and dataset version match
- **THEN** facts, revisions and evidence limitations agree across API and MCP

#### Scenario: A requested date precedes retained history
- **WHEN** a client requests information outside the available historical bounds
- **THEN** the response reports that limitation rather than an apparently complete empty result

#### Scenario: Retained vigilance maps contain weather observations
- **WHEN** a supported vigilance projection is visible at the requested service-knowledge boundary
- **THEN** regional search and situation facts expose the same optional weather metadata in API/MCP, including graphical status, HTML evaluation membership and distinct daily/cumulative rainfall bands
- **AND** PDF/page provenance and earlier weather-free projections remain available at their original knowledge boundaries

### Requirement: Citizen-facing evidence and official instructions
Municipal-status results SHALL give relevant local provisions priority while keeping regional products distinct and available. Structured MCP results SHALL carry authority, supporting document/link, territory, risk or measure, documented validity, last complete check and uncertainty next to the affected information. The service SHALL expose official instructions with attribution and SHALL NOT generate independent safety judgments, invent emergency advice or claim to issue official warnings. Acquired external text SHALL be treated as data rather than executable instructions. These guarantees SHALL apply to service output and SHALL NOT be represented as control over an external assistant's final wording.

For facts discovered through Cittadino Informato, equivalent API/MCP output SHALL retain channel attribution, supporting primary document identities and field-level multi-source verification results for Regione Toscana/CFR, municipality and platform, including unavailable, missing, non-comparable, not-applicable and conflicting checks. Diagnostic candidates without required primary support SHALL NOT be presented as confirmed local measures or originating regional alert levels. Technical acquisition readiness, source acceptance, corroboration and public publication SHALL remain independent; task 26.2 SHALL follow the multi-source verification completion criterion in official-source-ingestion without a license-documentation gate. The existing link-only restriction on platform copies SHALL still apply. Public records MAY include at most the latest 100 publishable verification receipts at the same service-knowledge boundary, with an explicit truncation limitation. Candidate and evidence visibility SHALL follow the applicable source-publication and development-municipality gates. Unpublished channel identities, passages and values, private request/candidate keys, OCR selectors and operator notes SHALL NOT be disclosed; channel outcomes and their evidence-visibility limits SHALL remain explicit.

#### Scenario: A closure accompanies a regional yellow warning
- **WHEN** the municipality has documented a relevant local closure
- **THEN** the closure is prominent with its official evidence while the regional warning remains separate context

#### Scenario: A newer publication cannot be interpreted
- **WHEN** an MCP client requests the affected local situation
- **THEN** the response includes document metadata, the official link and the interpretation warning, with a retained-copy link only when enabled and permitted and does not present the preceding state as fully updated

### Requirement: Configurable and published usage limits
The service SHALL enforce a shared per-IP request budget across anonymous API and MCP access, with configurable allowance, window and maximum response/page size. The initial deployment allowance SHALL be 120 requests per IP in each 60-second window and the maximum response/page size SHALL be 100 items. It SHALL publish limits and accounting policy, including that callers sharing an IP share the budget. Limit responses SHALL include actionable retry information: HTTP 429 for API requests and an appropriate MCP or transport-level response. User queries SHALL NOT initiate arbitrary upstream URL retrieval. Monitoring administration SHALL remain internal.

#### Scenario: A client exceeds the configured allowance
- **WHEN** a request exceeds the published budget
- **THEN** the service returns an explicit limit response with retry information without increasing collection load

### Requirement: Consistent results and explicit unavailability
Successful paginated queries SHALL identify a consistent data version; pagination SHALL preserve that view and its evaluation time or explicitly report its expiry. Cursor lifetime SHALL be configurable, initially 30 minutes; an expired cursor SHALL require restarting the query. Invalid parameters, unknown identifiers, unsupported territory, incomplete source coverage and unavailable service storage SHALL be distinguishable. Storage failure SHALL produce service unavailability, not a successful empty result. Cached responses SHALL preserve truthful updating status as time passes.

#### Scenario: New data arrives between pages
- **WHEN** a client follows a cursor after a new publication is acquired
- **THEN** the query remains on the original data version or reports cursor expiry rather than silently mixing versions

#### Scenario: The dataset cannot be read
- **WHEN** storage is unavailable
- **THEN** the API reports unavailability, such as HTTP 503, and MCP reports the equivalent service failure rather than an all-clear result

### Requirement: Five equivalent public operation groups
API and MCP SHALL provide equivalent municipality/zone discovery, municipality situation, filtered warning/measure search, document/evidence/version details and source coverage/updating operations. Applicable operations SHALL support available history. Both interfaces SHALL use the same underlying published interpretation and SHALL NOT initiate upstream collection or inference on user queries.

#### Scenario: Two clients request a municipality situation
- **WHEN** API and MCP requests use the same municipality, data view and evaluation time
- **THEN** both return equivalent regional products, local provisions, evidence and limitations

### Requirement: Optional access to retained copies
Public document results SHALL retain official source links. Public access to acquired copies SHALL be controlled by a service setting disabled by software default, with per-source restrictions that override global enablement. The initial MVP deployment MAY enable the global setting for CFR/Regione Toscana products and direct Calcinaia content only where recorded source policy permits redistribution. Cittadino Informato and albo pretorio content SHALL remain link-only. When permitted, a copy link SHALL resolve through the application to the precise acquired version; underlying storage SHALL remain internal. Disabling copy access SHALL NOT hide document metadata, official links or interpretation warnings and SHALL NOT prevent internal retention where permitted.

#### Scenario: Public copies are disabled
- **WHEN** a client requests document details with default service configuration
- **THEN** metadata and official links are returned but acquired-copy access is not offered or served

#### Scenario: A source forbids redistributing copies
- **WHEN** public copy access is globally enabled but the source restricts redistribution
- **THEN** that source's copies remain inaccessible publicly while its permitted metadata and official links remain available

#### Scenario: Selective retained copies are enabled for the MVP
- **WHEN** public copy access is enabled and a client requests documents from sources with different redistribution policies
- **THEN** permitted CFR/Regione Toscana and eligible direct municipal versions are served through the application while Cittadino Informato and albo pretorio records expose only their official links

### Requirement: Administration isolated from public interfaces
Administrative pages and operations SHALL be accessible only through a local interface, separate from public API/MCP; remote-host access SHALL use an administrative SSH tunnel or explicitly configured tailnet-only Tailscale Serve over HTTPS. The latter SHALL accept only the configured exact external origin, preserve cross-site request protections and keep the backend loopback-published; tailnet access policy SHALL restrict administrative access. Forwarded headers SHALL NOT grant administrative access. Source configuration, collection/public enablement, publication suspension, reprocessing, diagnostics and consumption reports SHALL NOT be exposed as public endpoints or MCP tools.

#### Scenario: An anonymous caller attempts a source activation
- **WHEN** the caller uses public API or MCP access
- **THEN** no administrative activation operation is available

### Requirement: Manual municipality publication in development
An explicitly configured development runtime SHALL allow private operators to enable or revoke publication for individual current supported Toscana municipalities, with an expected revision and an immutable actor/time audit. The setting SHALL be disabled by default and independent of collection, source acceptance and global source publication. It SHALL expose retained municipal data and applicable CFR products through equivalent API/MCP queries without requiring production acceptance reviews, observational completion or release readiness. Only active source configurations with recorded provenance, permitted publication and no unresolved policy conditions SHALL qualify. Versions SHALL have an acquisition under that active configuration. Unsupported products, DPC comparisons and local sources of unselected municipalities SHALL remain excluded. Unscoped regional fact searches SHALL include only zones applicable to selected municipalities; selecting one municipality SHALL NOT enable other municipalities sharing its zone. Shared official regional bulletin metadata MAY describe the complete bulletin.

Responses SHALL identify development publication on municipality and source coverage records and disclose the manual override in interpretation/history limitations. Source acceptance and coverage states SHALL remain truthful, including pending acceptance and uncertain map levels. Retained-copy access SHALL continue to require its independent existing authorization. Revoking a selection or changing its relevant publication configuration SHALL invalidate saved development views and cursors. Development views SHALL NOT be served by strict runtimes.

`IWA_ENVIRONMENT` SHALL accept only development, staging or production and default to production. Compose environments SHALL pin their corresponding runtime value. Staging and production SHALL neither expose these controls nor honor saved development choices, including when a development database is deliberately copied. Existing source acceptance, release and explicit production publication gates SHALL continue to apply.

#### Scenario: Only Calcinaia is enabled in development
- **WHEN** an operator enables publication for ISTAT 050004
- **THEN** its retained municipal data and applicable regional facts become available with development limitations; another municipality sharing A4 remains unselected and source acceptance remains pending

#### Scenario: Publication is revoked after the first page
- **WHEN** the operator revokes the municipal setting
- **THEN** new queries follow the revoked selection and API/MCP access to an earlier saved view explicitly expires

#### Scenario: A development database is used in production
- **WHEN** development publication choices exist in the database and the runtime is staging, production or unspecified
- **THEN** those choices grant no public visibility and manual activation commands are unavailable


### Requirement: Municipal progress is independent of source acceptance
API/MCP SHALL distinguish source acceptance from interpretation progress over latest visible municipal notices at the requested knowledge boundary. Typed section containers and pagination SHALL NOT be treated as individual pending notices. Irrelevant classified publications SHALL NOT require a measure extraction to count as interpreted. Validated extraction awaiting domain projection SHALL remain partial. The situation SHALL disclose incomplete interpretation of unresolved first municipal notices in its municipal source summary even when no preceding measure exists; document discovery/details SHALL retain their official links. Public local measures MAY expose their evidence-supported subject separately from action and place.

#### Scenario: A selected development municipality has processed notices but pending acceptance
- **WHEN** ordinary municipal interpretations exist and source acceptance remains pending
- **THEN** coverage retains pending acceptance and separately reports actual interpretation progress rather than setting interpretation to not_processed solely because acceptance is pending

#### Scenario: A first relevant notice has not been interpreted
- **WHEN** a visible municipal notice has no preceding structured measure and no supported interpretation
- **THEN** the situation municipal summary reports incomplete interpretation, while document discovery/details retain the document and official link without inventing a measure or all-clear

### Requirement: Concise deductions and separate source summaries
The municipality situation API/MCP data SHALL expose municipality identity,
`processed_data` and `source_summaries`. Processed data SHALL contain factual
conclusions supported by the prepared sources, distinct regional risks/levels/
zones, local actions/subjects/places, operational phases, validity and affected
limitations, with deduplicated document/version/URL/page references. It SHALL NOT
return the raw backlog of uninterpreted documents or repeated quality/receipt
structures. Those details SHALL remain available through document, search and
coverage operations. Toscana SHALL have separate regional, municipal and
Cittadino Informato summaries with source-specific contents and availability;
regional products SHALL remain individually identified. No publishable channel
data SHALL mean not collected/unavailable, never absence of notices. Acquisition
freshness, interpretation and source acceptance SHALL remain distinguishable.

#### Scenario: Regional green with ambiguous municipal validity
- **WHEN** the CFR supports green for applicable current risks and municipal publications document restrictions with conflicting or missing dates
- **THEN** the processed data reports those regional levels and documented restrictions separately, marks their current validity undetermined and does not infer all-clear or confirmation that restrictions are in force

#### Scenario: The platform has no publishable acquisitions
- **WHEN** a Toscana municipality has regional and municipal data but no publishable Cittadino Informato data in the query view
- **THEN** the platform summary states not collected/unavailable without inventing platform contents, assuming no notices or reusing a separate trial's data

#### Scenario: Many documents await interpretation
- **WHEN** the pinned municipal situation contains many pending notices
- **THEN** the municipal source summary reports incomplete interpretation; situation pages contain only structured facts, and full-scope conclusions/source summaries stay consistent across API/MCP pages while check freshness changes truthfully at delivery

#### Scenario: Complete situation query exceeds ten seconds

- **WHEN** constructing a complete saved situation takes more than ten seconds but less than the bounded 30-second HTTP write budget
- **THEN** the transport returns the complete paginated response rather than closing the connection without a response
- **AND** request-read and header limits remain bounded independently
- **AND** API/MCP saved-view and historical evidence semantics are unchanged

#### Scenario: A partial projection supports an elevated precise interval
- **WHEN** a current regional fact supports an elevated alert despite unrelated unresolved projection entries
- **THEN** the summary names its risk, color and zone with the supported interval in its recorded timezone
- **AND** fact reliability and unsupported entries remain explicit across all pages

### Requirement: Efficient complete public reads

Public API/MCP queries SHALL use indexed version and receipt lookups and reuse repeated source quality and document interpretation reads only within one operation at one evaluation/knowledge boundary. Municipality discovery SHALL batch coverage reads rather than issue queries for every municipality. Reuse SHALL preserve independent fact limitations, source suspension, publication revocation, historical cutoffs and complete saved-view pagination; no cross-request result cache SHALL delay those changes.

#### Scenario: Repeated facts share a source
- **WHEN** a complete query includes many facts or documents from the same source
- **THEN** source quality is loaded once for that source and scope within the operation
- **AND** each fact retains its independent interpretation and limitations

#### Scenario: Publication or interpretation changes between requests
- **WHEN** a subsequent request follows a publication revocation, source suspension or new interpretation
- **THEN** the new request reevaluates current controls at its requested knowledge boundary, without reusing prior-request quality or interpretation results

#### Scenario: A development public-listener repair is adopted independently
- **WHEN** development selects `IWA_PUBLIC_IMAGE` for a verified query repair
- **THEN** the public listener uses that image while other application surfaces retain `IWA_APP_IMAGE`
- **AND** the staging and production overlays retain their shared release image selection
