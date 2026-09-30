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
