## Purpose

Provide developers and citizens using assistants with equivalent free API and MCP access to attributable Toscana warnings, local measures and history.

## ADDED Requirements

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

### Requirement: Citizen-facing evidence and official instructions
Municipal-status results SHALL give relevant local provisions priority while keeping regional products distinct and available. Structured MCP results SHALL carry authority, supporting document/link, territory, risk or measure, documented validity, last complete check and uncertainty next to the affected information. The service SHALL expose official instructions with attribution and SHALL NOT generate independent safety judgments, invent emergency advice or claim to issue official warnings. Acquired external text SHALL be treated as data rather than executable instructions. These guarantees SHALL apply to service output and SHALL NOT be represented as control over an external assistant's final wording.

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
