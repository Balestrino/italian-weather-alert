# Spec Delta

## Purpose

Give private service operators a compact, accessible dashboard for identifying operational problems, inspecting evidence and performing deliberate administrative actions.

## ADDED Requirements

### Requirement: Consistent compact administrative navigation
Administrative HTML pages SHALL share a dark terminal-inspired shell with compact navigation, a clear current section, readable typography and text-labeled statuses. Existing administrative destinations and diagnostic links SHALL remain reachable. Layout SHALL support narrow screens, keyboard navigation and 200% zoom without hiding actions or evidence; wide tables SHALL scroll within their containers.

#### Scenario: Navigate across administrative page families
- **WHEN** an operator opens operations, source management, backups, incidents, evaluations, campaigns or release pages
- **THEN** navigation and current-section indication remain consistent and the operator can return to the overview

#### Scenario: Narrow viewport and keyboard use
- **WHEN** the operator uses a 390px viewport or navigates with the keyboard at 200% zoom
- **THEN** controls remain reachable, focus is visible, navigation can be opened and closed, and status meaning is available without color alone

### Requirement: Truthful operational overview
The overview SHALL display source issues, failed jobs, documents awaiting interpretation and open incidents with observation times and links to matching views. Counts SHALL cover the stated scope rather than the current page and SHALL distinguish unavailable information from zero. Overview reads SHALL NOT initiate collection, inference or mutations. A failed subsection SHALL NOT suppress other available summaries.

#### Scenario: More results than one page
- **WHEN** the matching scope contains more than 100 records or multiple attempts per entity
- **THEN** its summary reports the complete count of the stated entities and its link opens a matching filtered list

#### Scenario: One dependency is unavailable
- **WHEN** incident data cannot be read but operational aggregates are available
- **THEN** incidents are labeled unavailable and the other summaries remain visible with their timestamps

### Requirement: Readable lists and preserved query context
Operational lists SHALL use section-specific labels and primary columns, contextual actions and expandable remaining evidence. Filters SHALL be encoded in URLs and preserved across pagination. Changing filters SHALL reset pagination. An empty result SHALL be distinguishable from a read failure. Any client-side search SHALL explicitly describe its current-page scope.

#### Scenario: Filter and paginate
- **WHEN** an operator filters jobs by source and state and opens the next page
- **THEN** both filters remain active, are reflected in the URL and constrain the returned records

#### Scenario: Inspect lengthy evidence
- **WHEN** a record contains a long URL, error code or nested evidence
- **THEN** the operator can reveal the complete escaped value and follow applicable evidence links without losing the record context

### Requirement: Accounting meaning survives presentation changes
Usage pages SHALL preserve workload/stage/model distinctions, unit labels, currency separation, price provenance and known-subtotal indicators. Unknown costs or usage SHALL NOT be rendered as zero or complete totals.

#### Scenario: Partial usage and multiple currencies
- **WHEN** selected attempts have incomplete usage or costs in different currencies
- **THEN** the page labels incompleteness and shows separate currency totals with their provenance

### Requirement: Guided routine actions retain operational safeguards
Job relaunch and routine source controls SHALL provide labeled form fields without requiring JSON editing. Routine source controls SHALL include preview, collection activation/pause, intervals, acceptance, public enablement, interpretation pause/resume and selective reprocessing. Forms SHALL identify target, scope and operator, preserve observed revision/attempt checks and require explicit submission. Existing JSON and payload-form requests SHALL remain compatible. Advanced configuration and diagnostic editors SHALL remain available with clear labels.

#### Scenario: Relaunch a failed job
- **WHEN** an operator submits the guided form for a displayed failed job
- **THEN** only that job is selected, the observed attempt and operator are validated, and a readable outcome links back to the relevant view

#### Scenario: Submit stale source controls
- **WHEN** a source revision changes before a form is submitted
- **THEN** the operation rejects the stale change and explains that the operator must reload without overwriting newer state

#### Scenario: Submit invalid input
- **WHEN** a guided form contains missing, duplicate or invalid required values
- **THEN** the operation is rejected with safe, understandable feedback and no mutation occurs

### Requirement: Progressive enhancement and discoverable shortcuts
Core navigation, server filtering, evidence disclosure and routine form submissions SHALL work without JavaScript. Optional shortcuts SHALL be discoverable and disableable, SHALL ignore editable controls and modified key combinations, and SHALL NOT execute administrative actions. Escape SHALL dismiss enhanced navigation/search state with appropriate focus handling.

#### Scenario: JavaScript unavailable
- **WHEN** scripts are disabled or fail to load
- **THEN** an operator can navigate, filter, inspect details and submit a routine action using native controls

#### Scenario: Typing in a form
- **WHEN** an operator types a shortcut character inside a form field
- **THEN** normal text input proceeds without triggering navigation or search

### Requirement: Private assets and compatible service boundaries
Dashboard resources SHALL be served locally through the private administrative boundary without third-party requests. Existing Host/Origin protections, non-cacheable responses, output escaping and public API/MCP isolation SHALL remain effective. Existing JSON response contracts SHALL remain compatible, and browser errors SHALL NOT disclose internal dependency details or secrets.

#### Scenario: Public or foreign-origin request
- **WHEN** a dashboard asset is requested through the public service or an administrative mutation is submitted from a foreign origin
- **THEN** the public asset is unavailable and the mutation is rejected before any operational effect

#### Scenario: Existing JSON client
- **WHEN** an existing administrative client submits supported JSON or requests a JSON report
- **THEN** it receives the compatible machine-readable response and status behavior without HTML presentation markup

### Requirement: Today and tomorrow alert view
The private dashboard SHALL offer today and tomorrow in Europe/Rome, with separate
regional-warning and municipal-measure sections. It SHALL use retained data only,
show provenance and observation times, and distinguish consolidated facts from
retained bulletin metadata awaiting interpretation. Unknown validity SHALL remain
unassigned; missing results SHALL NOT imply no alert. Viewing or refreshing SHALL
NOT schedule collection, inference, reprocessing or publication.

#### Scenario: Bulletin exists but interpretation is pending
- **WHEN** a retained official bulletin explicitly lists today's or tomorrow's date but structured regional facts are unavailable
- **THEN** the matching day shows the bulletin's provenance and an unconfirmed-detail label, without inventing zone-level colors

#### Scenario: Local day boundary or uncertain validity
- **WHEN** UTC and Italian calendar dates differ, or a measure has uncertain validity
- **THEN** day headings use Europe/Rome and uncertain facts are shown separately instead of being assigned to both days
