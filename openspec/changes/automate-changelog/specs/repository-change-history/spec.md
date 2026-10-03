## ADDED Requirements

### Requirement: Public repository changelog
The repository SHALL keep a `CHANGELOG.md` with dated Unreleased entries for plans, fixes, implementation, documentation and operations changes. Entries SHALL describe repository changes without implying source acceptance or deployment and SHALL omit unpublished operational evidence.

#### Scenario: A fix and its plan are committed
- **WHEN** a contributor commits a fix with an OpenSpec update
- **THEN** the same commit contains a dated changelog entry summarizing the change

### Requirement: Automatic recording and enforcement
A shared pre-commit hook SHALL automatically insert and stage a fallback entry for staged changes that lack one. A contributor SHALL be able to supply a more meaningful kind and summary with a command. The recorder SHALL be idempotent when an entry is already staged and SHALL not stage unrelated unstaged changelog edits. CI SHALL reject each plan, fix, code, documentation or operations commit that lacks a dated changelog entry, including commits made without the hook.

#### Scenario: Contributor forgets an entry
- **WHEN** staged changes are committed with the shared hook active and no entry is staged
- **THEN** the hook creates and stages a dated Unreleased entry before the commit is written

#### Scenario: Hook is bypassed
- **WHEN** a change commit is pushed without a dated changelog entry
- **THEN** the CI changelog check fails and identifies the offending commit

### Requirement: Versioned software history
A software release SHALL preserve dated history under a version/date section and retain exactly one Unreleased section for future recording. Release preparation SHALL add a dated entry and publish notes that describe pilot limitations without implying source acceptance or production activation. The version tag and GitHub release SHALL identify the merged main revision after CI, Coverage and Security pass.

#### Scenario: First software release is prepared
- **WHEN** the development changes are grouped for version 0.1.0
- **THEN** their dated entries remain under the versioned heading, release preparation has its own dated entry and Unreleased remains available

#### Scenario: Development continues after a release
- **WHEN** a contributor records a new staged change
- **THEN** the recorder adds its dated entry under Unreleased without modifying versioned history

#### Scenario: A patch release is prepared
- **WHEN** development changes are grouped for the next patch version
- **THEN** the changelog preserves those dated entries under the patch version/date heading, OpenAPI and MCP implementation metadata identify the same software version, release notes summarize the changes and upgrade considerations, and historical releases remain unchanged

#### Scenario: A minor release is prepared and promoted
- **WHEN** development changes are grouped for version 0.2.0 and the operator requests promotion and push
- **THEN** OpenAPI and MCP metadata agree with the new minor version, dated history and historical notes remain intact, upgrade notes explain disabled embedding defaults and worker replacement, and promotion retains commit history with CI, Coverage and Security passing before merge and publication
