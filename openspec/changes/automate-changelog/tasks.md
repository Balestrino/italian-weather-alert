## 1. Changelog and recorder

- [x] 1.1 Add an initial public `CHANGELOG.md` and a standard-library command that records and stages one dated entry for staged changes, with explicit summary and kind and a safe fallback.
- [x] 1.2 Add synthetic Git tests for automatic insertion, idempotence, unstaged edit protection and per-commit validation.

## 2. Contributor enforcement

- [x] 2.1 Add an installable shared pre-commit hook and a CI check of each changed commit, including a bypassed-hook case.
- [x] 2.2 Document the workflow in contributor and agent guidance, and link the changelog from the repository index.
