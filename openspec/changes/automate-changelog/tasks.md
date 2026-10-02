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

## 4. Patch release preparation

- [x] 4.1 Bump OpenAPI and MCP implementation metadata to 0.1.1, group the post-0.1.0 dated changes under the new release heading, retain Unreleased, and add patch notes and current documentation links without modifying historical release notes.
- [x] 4.2 Verify version alignment, preserved dated history, changelog regressions/enforcement, public transport tests, local release-note links and strict OpenSpec validation; reconcile the progress index without creating a publication tag or claiming production activation.

Patch validation on 1 October 2026: OpenAPI/MCP metadata both identify 0.1.1, historical 0.1.0 entries and notes were preserved, all six changelog regressions and public transport tests passed, 45 local documentation links resolved, and strict change validation plus whitespace checks passed. Go race tests, vet and command builds also passed. Tagging and publication remain separate release operations.

## 5. Municipal corrections patch preparation — 2 October 2026

- [x] 5.1 Bump OpenAPI/MCP metadata to 0.1.2, group the dated municipal corrections under the patch heading, preserve historical 0.1.0/0.1.1 entries and notes, and update current release references with concise upgrade and pilot limits.
- [x] 5.2 Verify metadata alignment, unchanged historical entries/notes, changelog regressions/enforcement, public transport tests and local links; update the progress index and prepare promotion through the existing CI/Coverage/Security software-release workflow without implying image publication or production activation.
- [x] 5.3 Repair the interpretation integration fixture's observed published-port collision by letting Docker allocate its loopback port, registering cleanup before startup and verifying the affected PostgreSQL suite without weakening coverage gates.

Preparation validation on 2 October 2026: OpenAPI/MCP metadata identify 0.1.2; historical entries and release notes are unchanged. All six changelog regressions, public transport tests, the full isolated interpretation integration suite, tagged vet, local-link and whitespace checks passed. Docker now allocates the integration fixture loopback port after an observed Coverage startup collision. The OpenSpec CLI was unavailable; no CLI validation is claimed. Promotion, tagging and GitHub publication remain gated on the exact revision workflows; container publication and production activation remain separate.
