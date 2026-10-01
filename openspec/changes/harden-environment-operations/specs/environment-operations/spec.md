# Spec Delta

## Purpose

Provide reproducible and reliable management of three isolated IWA environments on one host, with explicit targeting, protected release identity and evidence-based operational readiness.

## ADDED Requirements

### Requirement: Fixed environment identity

Environment management SHALL require an explicit development, staging or production selection and SHALL use its fixed project name, configuration files and private environment file. Inherited environment overrides and conflicting command-line options SHALL NOT redirect an operation to another project's resources or replace the selected configuration. Management commands SHALL work from another working directory and from an extracted repository archive; release publication SHALL require a clean Git checkout.

#### Scenario: A caller overrides the staging project
- **WHEN** a staging command supplies a conflicting project, configuration-file, environment-file or project-directory option
- **THEN** the command fails before executing the requested operation and identifies the unsupported option without revealing private values

#### Scenario: A user manages an extracted download
- **WHEN** a user selects an environment from a repository archive without Git metadata
- **THEN** its configuration and management commands use that download's repository root and release publication explains its Git-checkout prerequisite

### Requirement: Recovery of an already activated stack

Application services and their PostgreSQL, object-storage and crawling dependencies SHALL restart automatically after a host restart or unexpected container exit, unless deliberately stopped. An environment that has never been activated SHALL remain inactive. Restart verification SHALL check service readiness and retained database-to-object references without claiming an unperformed host reboot succeeded.

#### Scenario: An isolated host restarts
- **WHEN** an isolated host containing an activated stack restarts
- **THEN** its listeners and dependencies recover without manual container starts and retained records still resolve their stored objects

### Requirement: Explicit and bounded smoke checks

Smoke checks SHALL require an explicit environment and obtain the target project, ports, service membership and secret paths from that environment's resolved configuration. Their default behavior SHALL be nondisruptive. Fault injection SHALL require an explicit interruption option, SHALL be rejected for production, SHALL verify target identity before stopping a dependency and SHALL attempt dependency recovery even when an assertion fails. Optional inactive workers and absent unused provider keys SHALL NOT cause ordinary readiness checks to fail.

#### Scenario: A user checks staging with conflicting shell variables
- **WHEN** a user runs an ordinary staging smoke check while shell variables refer to development
- **THEN** only staging is inspected and no container is stopped, started or recreated

#### Scenario: A staging fault assertion fails
- **WHEN** explicitly requested staging fault injection fails after stopping object storage
- **THEN** recovery is attempted for that same staging service and development resources remain untouched

#### Scenario: A user requests production fault injection
- **WHEN** a production smoke check includes the interruption option
- **THEN** it fails before any disruptive operation

### Requirement: Preservation and visibility of release identity

The supported staging release procedure SHALL reuse the application image built by the release helper for listener starts, migrations and storage initialization without rebuilding it. Publication SHALL verify that the running staging application containers use the intended revision and image identity before publishing the tested image. Management checks SHALL report the configured image, running revision and checkout revision, treating a difference as visible information rather than automatically rebuilding or replacing running services. Production approval SHALL remain a distinct operator decision.

#### Scenario: A release candidate passes through staging
- **WHEN** a clean revision is built and its documented staging initialization and validation steps run
- **THEN** the local candidate image identity and revision label remain unchanged and the same image is selected for publication

#### Scenario: Staging runs a different image
- **WHEN** publication is requested while staging's application containers use a different image identity or revision
- **THEN** publication fails before pushing and identifies the mismatch without disclosing private configuration

### Requirement: Production application image enforcement

Production operations that pull, create, start or restart application containers SHALL require a complete GHCR application image reference containing a SHA-256 digest of 64 hexadecimal characters. Mutable tags, missing digests and placeholder values SHALL be rejected before the operation. Production application builds SHALL be rejected. Preparation, inspection and stop/remove commands SHALL remain usable while an image placeholder is configured; diagnostic output SHALL distinguish prepared configuration from deployment-ready image selection. Default preparation SHALL select no services, and production limits SHALL remain explicit.

#### Scenario: A user starts production with a mutable tag
- **WHEN** a production application operation is requested with a tag or placeholder instead of a valid digest
- **THEN** it fails before pulling, creating or starting containers and explains how to configure the approved digest

#### Scenario: A user inspects inactive production
- **WHEN** production has its documented image placeholder and a user requests configuration or status inspection
- **THEN** inspection remains usable, reports the missing release selection and creates no containers

### Requirement: Repeatable container-readable secret initialization

Newly generated secret and private configuration files SHALL have the intended container-readable permissions inside an owner-private directory regardless of the invoking user's umask. Repeated initialization SHALL preserve existing credential contents and existing file access arrangements. Existing unreadable files SHALL be diagnosable without printing their contents or silently changing custom permissions.

#### Scenario: Secret generation runs under a strict umask
- **WHEN** an empty environment is initialized under umask 077
- **THEN** its new files are readable by the intended container users while their containing directory remains owner-private

#### Scenario: Existing credentials are initialized again
- **WHEN** initialization encounters existing credentials or custom access permissions
- **THEN** their values and file permissions are preserved

### Requirement: Reproducible onboarding and routine operations

The repository SHALL publish consistent instructions for development, staging and inactive production on one host, distinguishing a fresh download from an existing installation. Instructions SHALL include prerequisites, required directories, environment-specific secret locations and ports, dependencies, migration and storage initialization, readiness, status/logs, stop/start, upgrade and rollback. A fresh development startup SHALL require no inference key or live-source activation. Staging instructions SHALL preserve the release image and use controlled fixtures. Production documentation SHALL explain explicit service/profile selection and preserve its release and readiness gates.

#### Scenario: A new user follows the fresh-checkout procedure
- **WHEN** a user follows the documented commands on a disposable host without pre-existing private files
- **THEN** development becomes ready, staging can validate a built candidate, inactive production can be inspected, and routine commands target the stated environment without translating another guide's paths or ports

### Requirement: Verified host readiness

Production readiness SHALL require measured concurrent workload capacity, recoverable off-host whole-VM backup, a timed isolated restore, the tested published image digest, verified public/private routing and an approved deployment/rollback rehearsal. Capacity review SHALL follow the existing sustained 70% disk/RAM threshold and account for image builds, storage growth and all active environments. Recovery targets SHALL remain no more than six hours of lost data and one hour to verified public service. Repository/template validation SHALL NOT certify these operational outcomes, and unfinished external work SHALL remain explicitly open with private evidence kept out of commits.

#### Scenario: Repository checks pass but host readiness is incomplete
- **WHEN** configuration and regression tests pass while capacity, backup, restore, routing or deployment evidence is missing
- **THEN** repository work can be recorded as complete but production readiness remains pending

#### Scenario: A release is approved for a protected host
- **WHEN** concurrent capacity, backup/restore, release identity, routing and rollback checks have passed and the operator approves activation
- **THEN** the approved digest is deployed with separate production state and the measured results support the corresponding operational checklist entries
