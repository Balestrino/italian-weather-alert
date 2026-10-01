# Spec Delta

## Purpose

Prevent accidental activation of expensive background processing while providing equivalent CLI and admin controls for source collection in isolated environments on a shared host.

## ADDED Requirements

### Requirement: Default startup excludes background workers

Development and staging startup without explicit background service selection SHALL select the public/admin listeners and their existing dependencies but SHALL exclude collection/inference and application backup workers. Production preparation SHALL continue to select no services, and production core startup SHALL exclude both workers.

#### Scenario: Ordinary nonproduction startup
- **WHEN** development or staging service selection is resolved without optional profiles or explicit worker targets
- **THEN** the service set includes the listeners and their dependencies and excludes both background workers

#### Scenario: Production core startup
- **WHEN** production service selection enables only its core profile
- **THEN** collection/inference and application backup workers are not selected

### Requirement: Independent explicit background activation

Each environment SHALL support deliberate selection of its collection/inference worker independently of its application backup worker. Production SHALL retain its existing worker and backup profile names and SHALL NOT acquire a nonproduction worker activation path through inherited configuration. Explicit service targets SHALL remain supported and SHALL be documented as deliberate activation, even without selecting a profile.

#### Scenario: A controlled development or staging worker check
- **WHEN** an operator explicitly selects the nonproduction collection/inference profile
- **THEN** that worker is selected and the application backup worker remains excluded

#### Scenario: Production selection uses a nonproduction profile
- **WHEN** production service selection enables only a nonproduction worker profile
- **THEN** the production collection/inference worker is not selected

#### Scenario: A worker is explicitly targeted
- **WHEN** an operator names the collection/inference service in a startup command
- **THEN** it can be selected as a deliberate operation and the documentation explains that profiles are not a prohibition against explicit targets

### Requirement: Equivalent CLI and admin source collection controls

Operators SHALL be able to enable and suspend collection for an individual source and inspect its collection state through both a dedicated CLI and the admin panel. CLI operations SHALL require explicit environment selection through the supported environment wrapper, and the admin panel SHALL operate on its own environment's registry. Both interfaces SHALL apply the same source revision, successful acquisition preview, collection policy and territorial prerequisites, require an operator identity for mutations, and persist equivalent audit events. CLI mutations SHALL return machine-readable success output or a nonzero exit status with an operator-safe error when prerequisites fail; neither interface SHALL bypass the registry lifecycle using direct state updates.

#### Scenario: A source is enabled from either interface
- **WHEN** an operator enables a source with a valid revision, successful preview and permitted collection scope through the CLI or admin panel
- **THEN** the selected environment persists collection enablement and an audit event with the operator identity, and the other interface in that same environment observes the updated state

#### Scenario: A source lacks activation prerequisites
- **WHEN** an operator enables a source without a successful preview or with invalid revision, policy or territorial prerequisites
- **THEN** both interfaces reject the operation with the same lifecycle outcome, and the CLI returns a nonzero exit status

#### Scenario: A source is suspended in staging
- **WHEN** an operator suspends staging source collection through either interface
- **THEN** that source's staging collection state changes without changing the development or production registry or deleting retained history

### Requirement: Hierarchical territorial activation in admin and CLI

The admin region overview at `/admin/regions` SHALL offer explicit region enable/disable actions and the Configura regione navigation path. The selected region's Comuni tab SHALL list the complete adopted municipality register with configured enablement state, effective territorial eligibility and explicit actions for each municipality, including municipalities without sources. Both region and municipality enablement, disablement and state inspection SHALL also be available through dedicated CLI commands scoped to an explicitly selected environment. Mutations SHALL require operator identity and expected revision, validate territorial membership, and reject stale edits atomically. Region enablement SHALL retain its validated-register prerequisites. The UI SHALL work without JavaScript and preserve regional context, filters and pagination after an action.

#### Scenario: Activate a region then one municipality
- **WHEN** an operator enables a configured region from the overview, follows Configura regione and opens Comuni to enable one municipality
- **THEN** only that municipality's configured enablement changes, neither action enables sources or starts a worker, and the same persisted state is visible through the CLI

#### Scenario: Municipality action has incorrect scope or revision
- **WHEN** either interface targets a municipality outside the selected region or supplies a stale territorial revision
- **THEN** the action fails without changing any territory and the admin interface explains how to reload the current state

### Requirement: Parent territorial gates preserve subordinate choices

Automatic municipal acquisition SHALL require an active worker, an enabled region, an enabled municipality and enabled source collection, plus existing policy/profile gates. Regional sources SHALL require the region gate without depending on individual municipality gates. Disabling a region SHALL block admission of new acquisition and interpretation for all associated sources; disabling a municipality SHALL block such admission only for its local sources. Explicit acquisition previews and interpretation launches SHALL respect the same territorial gates. Already admitted work SHALL be allowed to finish. Disabling a parent SHALL preserve configured child enablement, individual source flags and retained data/history; re-enabling SHALL restore only eligibility under those saved settings without creating an explicit catch-up or reprocessing batch. Administrative reads and applicable retained regional results SHALL remain inspectable when municipal local processing is disabled. Public publication SHALL remain separately controlled.

#### Scenario: Region disablement overrides an enabled municipality
- **WHEN** an operator disables a region containing enabled municipalities and collection-enabled sources
- **THEN** no new associated acquisition or interpretation is admitted and child/source choices remain saved and visible as blocked by the region

#### Scenario: Disable one municipality
- **WHEN** an operator disables a municipality while its region and another municipality remain enabled
- **THEN** new work for that municipality's local sources is blocked while regional sources and the other municipality retain their eligibility

#### Scenario: Restore a disabled parent
- **WHEN** an operator re-enables a region or municipality
- **THEN** previously disabled child territories and sources remain disabled, permitted work regains eligibility through normal scheduling, and historical records are preserved

### Requirement: Stable municipal choices across dataset updates and migration

Municipal enablement SHALL be stored independently of immutable geographical datasets using validated region and ISTAT identity. Replacing an adopted municipality register SHALL preserve choices for retained identities and history for retired identities; newly introduced municipalities SHALL start disabled. Introducing municipality gates SHALL preserve existing execution eligibility for municipalities with previously configured local sources, record migration attribution, and leave municipalities without existing sources disabled. Migration SHALL NOT enable sources, start workers or activate regions.

#### Scenario: Import an updated municipality register
- **WHEN** a new register retains an existing municipality and introduces another
- **THEN** the retained municipality keeps its saved choice, the new municipality starts disabled, and older geographical and lifecycle history remains readable

#### Scenario: Adopt municipality gates on an existing database
- **WHEN** the additive migration runs against existing local source associations
- **THEN** those municipalities retain their previous eligibility subject to region/source gates, no new source is enabled, and unrelated municipalities remain disabled

### Requirement: Source enablement is independent of worker startup

Source collection controls SHALL remain available while the processing worker is stopped. Enabling a source SHALL persist the requested collection state without starting a worker, performing acquisition or invoking LLM/OCR as a side effect. Automatic collection SHALL require both an active worker and enabled source collection, subject to existing execution gates. CLI results and admin guidance SHALL distinguish configured source enablement from actual worker execution and SHALL NOT imply that collection has begun solely because enablement succeeded. Collection enablement SHALL NOT grant public publication, which remains a separate lifecycle decision.

#### Scenario: Enablement while the worker is stopped
- **WHEN** an operator enables a source from CLI or admin while the environment's processing worker is stopped
- **THEN** collection enablement is saved, the worker remains stopped, no acquisition or inference occurs, and the interface explains that automatic processing requires an active worker

#### Scenario: The worker starts after explicit source enablement
- **WHEN** an operator deliberately starts the processing worker in an environment containing enabled and disabled sources
- **THEN** automatic collection considers only enabled sources allowed by the existing execution gates and does not grant public publication

### Requirement: Shared-host processing lifecycle

The supported operating procedure SHALL designate production as the continuous live collector once production is deliberately activated. Development and staging SHALL use controlled fixtures or bounded, explicit collection/inference checks, with workers stopped after checks. Instructions for representative data copies SHALL require collection, inference and notifications to remain disabled before the copy is used. Documentation SHALL state that environment isolation does not deduplicate processing across environments.

#### Scenario: A nonproduction check finishes
- **WHEN** an operator follows the documented development or staging background processing procedure
- **THEN** the procedure includes explicit activation and stopping of the worker and warns that overlapping live sources duplicate processing across environments

### Requirement: Preserve active collection during adoption

Adoption instructions SHALL distinguish Compose service selection from the state and restart behavior of existing containers. Repository configuration changes SHALL NOT automatically stop an active collector, activate production, rename volumes or rotate credentials. The procedure SHALL preserve the current collector until a deliberate handover, and SHALL include explicit stopping and verification of nonproduction workers when their live collection responsibility ends.

#### Scenario: Existing development collection is active
- **WHEN** optional profiles are introduced while a development worker is running
- **THEN** repository verification leaves that worker untouched and the adoption procedure explains how to stop it and verify its state when the operator performs the handover

### Requirement: Verify background service selection without external work

Deployment validation SHALL check both default exclusion and deliberate selection of background services across all three environments. Regression validation SHALL resolve configuration or use synthetic fixtures without starting operational workers, contacting official sources, calling inference providers or sending notifications.

#### Scenario: Configuration validation detects default worker activation
- **WHEN** a changed environment configuration selects a background worker by default
- **THEN** deployment validation fails and identifies the affected environment and service without revealing private values
