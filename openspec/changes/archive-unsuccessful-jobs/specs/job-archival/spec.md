## ADDED Requirements
### Requirement: Audited unsuccessful-job archival
The system SHALL archive queued, retry-wait and failed jobs before an explicit cutoff with actor and timestamp, while preserving successful jobs and all receipt, result, attempt and document history.

#### Scenario: Archive a stopped backlog
- **WHEN** an administrator archives work before a cutoff and no selected job is running
- **THEN** only unsuccessful selected jobs are archived and repeat invocation archives zero additional jobs

#### Scenario: Running work exists
- **WHEN** selected work is running
- **THEN** archival fails atomically without changes

### Requirement: Archived jobs cannot run again
The system SHALL preserve deduplication identities and exclude archived jobs from claim, relaunch and provider deferral.

#### Scenario: Repeated collection of old evidence
- **WHEN** the same job is enqueued after archival
- **THEN** its existing identity is returned without creating or running another job

#### Scenario: New evidence
- **WHEN** a new document version creates a distinct job
- **THEN** the job remains eligible for ordinary processing

### Requirement: Operational views exclude archives
The system SHALL hide archived jobs from operational lists, failed-job counts and active failure detection while preserving historical usage and genuine source/quality errors.

#### Scenario: Cleanup completed
- **WHEN** all unsuccessful jobs are archived
- **THEN** the default jobs page contains only successful jobs until new work arrives and historical token totals are unchanged

### Requirement: Reset existing pending documents
The system SHALL allow audited archival of all pending retained versions before a
cutoff, preserving evidence and historical usage while excluding archived versions
from pending views and automatic interpretation.

#### Scenario: Pending backlog reset
- **WHEN** the administrator archives current pending documents after retiring their queue work
- **THEN** missing or invalid interpretations are archived, successful interpretations are preserved, and repeated reset is idempotent

#### Scenario: Old evidence is revisited
- **WHEN** acquisition or a stale job refers to an archived version
- **THEN** interpretation makes no provider calls for that version

#### Scenario: Evidence changes after reset
- **WHEN** a new retained version is acquired
- **THEN** normal scheduling remains eligible and the version appears as pending until processed
