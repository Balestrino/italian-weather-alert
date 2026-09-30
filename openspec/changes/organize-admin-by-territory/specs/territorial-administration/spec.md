# Spec Delta

## Purpose

Consentire agli operatori di configurare e abilitare regioni e comuni con identità stabili, dati geografici verificabili, isolamento territoriale e storico delle operazioni.

## ADDED Requirements

### Requirement: Regional catalog and onboarding
The administration SHALL provide an Italian regional catalog identified by official region codes, including regions not yet configured. An operator SHALL be able to configure and enable a new region through the dashboard. Regions SHALL expose distinct unconfigured, configured-disabled and enabled states, separately from source health and public availability. New regions SHALL start disabled. Enabling SHALL require a selected, validated municipality dataset belonging to the region and a valid configuration revision; missing prerequisites SHALL be explained.

#### Scenario: Configure a second region
- **WHEN** an operator selects an unconfigured region, imports its validated municipality register, saves its configuration and explicitly enables it
- **THEN** it appears enabled, its municipalities become navigable, and no source is automatically collected, accepted or published

#### Scenario: Incomplete setup
- **WHEN** an operator attempts to enable a region without a validated municipality register
- **THEN** the change is rejected and the missing prerequisite is shown without altering the existing state

### Requirement: Complete versioned municipality registers
The system SHALL import attributed and versioned official municipality records with region code, ISTAT municipality code, name and province. Import SHALL offer validation and a preview before adoption, reject duplicates and inconsistent territorial membership atomically, and record dataset provenance, verification time, checksum and declared completeness. Regional lists SHALL cover all municipalities in the adopted regional register, including those without configured sources. A partial import SHALL NOT be presented as a complete register or satisfy the enablement prerequisite. Adoption SHALL preserve earlier datasets and associations for historical reads.

#### Scenario: Municipality without collection
- **WHEN** the adopted register includes a municipality with no source
- **THEN** its page and list entry exist and display no configured local source, independently of applicable regional results

#### Scenario: Invalid or updated register
- **WHEN** an import contains duplicate ISTAT codes or a municipality assigned to the wrong region
- **THEN** adoption fails atomically and the previous register remains selected

#### Scenario: Municipality removed in a later register
- **WHEN** a newer adopted register retires a municipality
- **THEN** the current list identifies its absence while historical links, names, results and configuration events remain attributable to the earlier version

### Requirement: Region scoped configuration and source association
Regional configuration SHALL identify selected municipality, zone and postal datasets where available, source associations and supported processing profiles. Municipal configuration SHALL identify its local sources, applicable regional mappings and effective source settings. Operators SHALL be able to create and configure sources from the relevant territory page using existing source controls. Source membership SHALL use explicit validated identifiers, preserve association history and reject cross-region mismatches. Regional permissions SHALL NOT be inherited as source permissions. Unsupported processing profiles SHALL be disclosed and SHALL block activation of the affected processing, rather than reuse a Toscana profile implicitly.

#### Scenario: Configure municipal collection
- **WHEN** an operator creates a local source from a municipality page
- **THEN** the source is associated with that municipality and region and existing evidence, configuration and revision validations apply

#### Scenario: Unsupported regional portal
- **WHEN** an enabled region has a source requiring an unavailable regional processing profile
- **THEN** the region remains configurable, the source explains its unsupported processing state and no incompatible parser is executed

### Requirement: Audited territorial lifecycle
Region configuration changes, dataset adoption and enablement changes SHALL require explicit submission, operator attribution and an expected revision. Stale changes SHALL fail without overwriting newer state. Disabling a region SHALL prevent new acquisition or interpretation work from starting for its sources, including manual launches, while preserving source settings and historical data. Work already executing SHALL be allowed to finish and remain visible. Re-enabling SHALL restore eligibility under existing source gates without creating an automatic catch-up or reprocessing batch. Source acceptance and public enablement SHALL remain separate explicit actions.

#### Scenario: Disable with pending work
- **WHEN** an operator disables a region with pending and executing jobs
- **THEN** pending jobs cannot start, executing jobs may finish, and the page explains this scope while preserving results and source settings

#### Scenario: Conflicting edits
- **WHEN** a configuration form references an outdated revision
- **THEN** submission fails with a reload instruction and creates no configuration change

### Requirement: Territorial isolation and compatibility
Current and historical reads SHALL resolve region, municipality, source and zone membership within the requested scope and selected dataset versions. Identical zone labels in different regions SHALL NOT mix results. Existing Toscana data, source revisions and public contracts SHALL remain compatible. Administrative onboarding SHALL NOT automatically extend public discovery, API or MCP coverage to another region; publication outside the existing public perimeter SHALL be rejected with an explicit unsupported-scope explanation. Read-only administrative navigation SHALL NOT schedule collection, inference or publication.

#### Scenario: Same zone label in two regions
- **WHEN** two configured regions both use zone A1 and have retained results
- **THEN** each region and municipality page shows only its applicable results, including historical views

#### Scenario: New administrative region and existing public client
- **WHEN** a new region is enabled administratively
- **THEN** existing public discovery and Toscana results retain their scope and publishing a source outside that scope is rejected explicitly

#### Scenario: Migrate the existing deployment
- **WHEN** territorial administration is introduced on the current database
- **THEN** known Toscana records retain their identifiers, source states and histories, and unresolved legacy associations are reported rather than guessed
