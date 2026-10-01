## 1. Changelog and recorder

- [x] 1.1 Add an initial public `CHANGELOG.md` and a standard-library command that records and stages one dated entry for staged changes, with explicit summary and kind and a safe fallback.
- [x] 1.2 Add synthetic Git tests for automatic insertion, idempotence, unstaged edit protection and per-commit validation.

## 2. Contributor enforcement

- [x] 2.1 Add an installable shared pre-commit hook and a CI check of each changed commit, including a bypassed-hook case.
- [x] 2.2 Document the workflow in contributor and agent guidance, and link the changelog from the repository index.

## 3. First versioned software release

- [x] 3.1 Group existing dated history under 0.1.0, retain Unreleased, add a release-preparation entry and write pilot notes plus the development-to-main publication procedure.
- [x] 3.2 Verify recorder behavior with a versioned history, run changelog regression/enforcement checks and reconcile the public progress index without asserting image publication, source acceptance or production readiness.

Validation on 1 October 2026: all six changelog regression tests passed. A disposable synthetic Git repository verified release-entry idempotence, recording the next change under Unreleased without modifying the versioned section, and per-commit enforcement across release preparation and subsequent development.
