## Purpose

Make provider consumption and failures explainable, contain repeated rejected work and recover processing without hiding uncertainty or weakening evidence validation.

## ADDED Requirements

### Requirement: Safe provider call diagnostics and accounting
The service SHALL record each provider call's stage, input identity, configuration, timing, available request identifier, HTTP status, bounded sanitized error category and available usage. Secrets, authorization headers and raw sensitive payloads SHALL NOT be logged. Unknown or interrupted consumption SHALL remain separate from measured totals. Calls whose outputs fail local validation SHALL still contribute their known usage exactly once.

#### Scenario: A provider rejects a request
- **WHEN** the response is non-successful
- **THEN** an operator can distinguish authentication, quota, malformed-request and temporary-service failures where the response supports that distinction, with unknown usage explicitly labeled

#### Scenario: A process exits after a call starts
- **WHEN** no durable response or usage receipt is available
- **THEN** recovery exposes an interrupted call with unknown consumption instead of zero usage or an invented successful result

### Requirement: Provider failure containment across jobs
The service SHALL contain confirmed provider-wide authentication, account/quota and availability failures across new jobs in the affected scope, preserve queued work and continue collection. Temporary failures SHALL obey bounded backoff and Retry-After. Request-specific permanent rejection SHALL stop retrying the same input/configuration without unnecessarily disabling healthy work. Recovery SHALL be explicit or use a bounded probe as appropriate to the error category and SHALL NOT trigger an unlimited backlog replay.

#### Scenario: Invalid credentials reject every new document
- **WHEN** a confirmed authentication failure occurs
- **THEN** further calls in the affected credential scope are held, operators see the reason and collection continues until controlled recovery

#### Scenario: One image violates the provider contract
- **WHEN** its input is permanently rejected
- **THEN** unchanged equivalent inputs reuse the failure state until correction or explicit relaunch while unrelated valid inputs remain processable

### Requirement: Recoverable worker lifecycle
Worker exits SHALL identify the failing component and safe cause. Recoverable lease loss and temporary infrastructure faults SHALL preserve bounded recovery and SHALL NOT cause an unbounded coordinated restart loop. Lease loss SHALL stop further owned work and prevent stale completion. Fatal configuration errors SHALL remain visible and fail clearly. Recovery SHALL preserve attempt budgets, evidence and idempotent public effects.

#### Scenario: A heartbeat loses job ownership
- **WHEN** another worker can recover the expired claim
- **THEN** the stale handler is canceled, cannot publish effects and records any uncertain in-flight provider consumption

#### Scenario: A temporary database fault interrupts a worker
- **WHEN** infrastructure becomes available again
- **THEN** bounded recovery resumes durable work with diagnostic history and without duplicated public effects

### Requirement: Validated outputs and checkpointed progress
The service SHALL retain strict structured-output and literal-evidence validation while using evaluated provider requests. Completed valid segments SHALL survive later segment failures and compatible retries. Invalid outputs SHALL have stable diagnostic reasons and SHALL NOT trigger unlimited repair requests, silently become negative classifications or weaken unsupported-field checks.

#### Scenario: The final segment fails after earlier segments succeed
- **WHEN** the job is retried with identical evidence and configuration
- **THEN** validated earlier segments require no additional calls and the failed segment follows the bounded retry policy

#### Scenario: A response asserts unsupported evidence
- **WHEN** local validation detects the unsupported claim
- **THEN** it remains rejected or explicitly undetermined, paid usage is retained and the failure can be included in quality evaluation

### Requirement: Verifiable efficiency and operational reports
Local administration SHALL report measured tokens, unknown calls, validation failures, provider rejections, reuse counts and worker failures by time, source, model and stage. Reuse savings SHALL be labeled estimates separately from measured billing. Provider-account reconciliation SHALL state its time window, account/key scope and missing evidence. Efficiency changes SHALL pass retained-case quality evaluation and staged operational observation before completion is claimed.

#### Scenario: Application totals differ from provider totals
- **WHEN** an operator compares a provider export with a local usage report
- **THEN** the report identifies known scope and timing differences and unknown calls without claiming unsupported exact reconciliation
