## Why

Plans, fixes, and operational changes can reach the repository without a concise public record. A shared changelog makes changes discoverable, while an automatic local recorder and CI check prevent omissions when contributors forget to edit it.

## What Changes

- Add `CHANGELOG.md` with an Unreleased section and an initial entry for the listing-version fix.
- Add a standard-library recorder that creates and stages a dated entry from staged changes, accepting a meaningful summary and kind when supplied.
- Add a shared pre-commit hook and per-commit CI check, plus contributor and agent instructions.
- Group released entries under a version/date heading while keeping Unreleased available; document software promotion and publish pilot release notes.

## Impact

All repository change commits need a changelog entry. The local hook supplies a fallback; CI catches commits made without the hook. Entries remain public summaries and must not contain unpublished operational evidence or imply that a source or deployment was accepted.
