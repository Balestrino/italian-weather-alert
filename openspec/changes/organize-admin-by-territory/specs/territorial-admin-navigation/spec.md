# Spec Delta

## Purpose

Organizzare il pannello amministrativo attorno a regioni e comuni, rendendo immediatamente consultabili configurazione, risultati e storico senza ricostruire il contesto fra strumenti globali.

## ADDED Requirements

### Requirement: Territory first administrative entry
The administrative entry SHALL present regions with enablement state, configuration readiness, municipality coverage, latest result time and operational issues. Enabled regions SHALL be immediately identifiable; unconfigured and disabled regions SHALL remain discoverable for onboarding. Primary navigation SHALL separate Regioni, Operazioni and Sistema. Technical destinations SHALL remain reachable without dominating the territorial overview. Summaries SHALL cover the full stated scope, include observation time and distinguish unavailable data from zero.

#### Scenario: Open the dashboard
- **WHEN** an operator opens the administrative entry
- **THEN** they can identify enabled regions, inspect configuration and results, and start setup of another region from the same page

#### Scenario: Summary dependency unavailable
- **WHEN** one regional summary cannot be read
- **THEN** that information is labeled unavailable while other successfully loaded summaries remain visible

### Requirement: Region detail and complete municipality navigation
Each region SHALL have stable detail links with Risultati, Comuni, Configurazione and Storico sections and breadcrumbs. Comuni SHALL list the entire adopted register using server-side search by name or ISTAT, province and coverage filters, deterministic pagination and full matching counts. Filters SHALL survive pagination and changes SHALL reset pagination. Municipalities without sources SHALL be included. Configuration and results links SHALL preserve the regional context.

#### Scenario: Navigate beyond the first page
- **WHEN** a region has more municipalities than fit on a page and the operator filters by province
- **THEN** counts cover all matching municipalities and every page preserves that filter

#### Scenario: Enter an unconfigured region
- **WHEN** an operator opens a region without an adopted municipality register
- **THEN** the page explains the missing register and provides its setup action rather than claiming that the region has zero municipalities

### Requirement: Municipality workspace
Each municipality SHALL have a stable page identified by region and ISTAT code, with Dati, Configurazione and Storico sections. Dati SHALL show applicable regional warnings separately from local measures, retained documents, processing state, freshness, provenance and known limitations. Configurazione SHALL show local sources, effective settings and regional mappings with contextual actions. A municipality without local sources SHALL still show available applicable regional results and a setup path for local collection. Invalid region/municipality combinations SHALL return a not-found response without data from another territory.

#### Scenario: Municipality with multiple local sources
- **WHEN** an operator opens a municipality with multiple sources
- **THEN** all associated sources are identifiable, results are attributed correctly and each configuration action identifies its target source

#### Scenario: Foreign municipality in the URL
- **WHEN** a request combines a valid region with a municipality belonging to another region
- **THEN** the detail request returns not found and reveals no results under the incorrect regional context

### Requirement: Evidence based territorial results
Regional and municipal results SHALL distinguish retained documents, pending or failed interpretation, consolidated facts and public eligibility. Today and tomorrow SHALL use Europe/Rome. Missing results SHALL NOT imply no alert; uncertain validity SHALL remain explicit. Regional warnings SHALL use the applicable versioned zone mapping, including partial municipality coverage. Result links SHALL expose underlying evidence without widening territorial scope. Administrative retained data SHALL remain inspectable even when it is not publicly enabled.

#### Scenario: Unpublished bulletin awaiting interpretation
- **WHEN** a retained regional bulletin has no consolidated interpretation and is not publicly enabled
- **THEN** the private region page shows its provenance and pending state without inventing warning colors or presenting it as publicly available

#### Scenario: Partial municipality warning
- **WHEN** only one of multiple zones covering a municipality has an elevated warning
- **THEN** the municipality view identifies the relevant zone and partial applicability instead of claiming a uniform municipality-wide level

### Requirement: Inspectable territorial history
Region and municipality history SHALL include separate filters for configuration and enablement events, source operations, acquired document versions and processing outcomes. It SHALL expose timestamp, actor where recorded, source, revision or version, outcome and evidence links. Date and source filters, deterministic ordering and pagination SHALL operate on the whole scope. Past records SHALL retain their original territorial attribution. Missing actors, incomplete retention and unavailable history SHALL be labeled explicitly; current state SHALL NOT replace historical values.

#### Scenario: Inspect a configuration change and later result
- **WHEN** an operator selects a date range in municipal history
- **THEN** they can inspect source configuration revisions and the document versions and processing outcomes recorded in that range with their original provenance

#### Scenario: History after disablement
- **WHEN** a region or source is disabled
- **THEN** its preserved results and history remain navigable and clearly identify the current disabled state

### Requirement: Accessible and compatible contextual actions
Territorial pages SHALL work without JavaScript, at a 390px viewport and 200% zoom, with keyboard navigation, visible focus and text status labels. Existing administrative deep links and JSON requests SHALL remain compatible. Contextual actions SHALL preserve source revision and operator checks and return to the relevant territory using validated local destinations. Host/Origin protections, no-store responses, private local assets, output escaping and public-router isolation SHALL remain effective.

#### Scenario: Configure without JavaScript
- **WHEN** an operator navigates region to municipality and submits a source form with JavaScript disabled
- **THEN** navigation, filtering, validation and the return to the municipality remain functional

#### Scenario: Unsafe contextual return
- **WHEN** a request supplies an external return URL or a mismatched territorial target
- **THEN** it is rejected or replaced with a safe validated local destination without weakening source authorization checks

#### Scenario: Public request or foreign origin
- **WHEN** territorial pages are requested through the public listener or a foreign origin submits a territorial change
- **THEN** the public route is unavailable and the foreign-origin change is rejected without effects
