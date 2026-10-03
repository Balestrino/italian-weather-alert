## Context

Git hooks are local and are not installed by cloning a repository. Git's pre-commit hook can change the index before a commit is written; later message hooks cannot reliably add files to that commit. CI therefore checks committed history independently.

## Decisions

Keep one human-readable `CHANGELOG.md` with dated Unreleased bullets. `scripts/changelog.py record` reads staged paths, inserts one entry under Unreleased, and stages the file. An explicit kind and summary produce a useful entry; the hook uses a path-based fallback if none was supplied. Existing unstaged changelog edits are never staged implicitly.

`scripts/changelog.py check --base` inspects every commit since the merge base. A commit that changes any other repository file must add a dated changelog bullet in that same commit. CI runs this check for pushes and pull requests with full Git history. A required status check on protected branches makes it a merge gate; local hook installation improves convenience but is not trusted as enforcement.

The initial changelog records the current fix and this mechanism. Entries describe repository changes only. They do not state that a source was accepted, the service was deployed, or private evidence was published.

Software releases move the existing dated bullets into a version/date section and retain exactly one empty Unreleased section for the next changes. Release preparation adds its own dated bullet, so reorganizing history still satisfies per-commit enforcement. The recorder continues to write only under Unreleased; versioned sections are preserved. Reviewed pilot notes are committed with the release preparation. Promotion retains history through a `dev` to `main` merge; tag and GitHub publication target the merged revision after its workflows pass. Container publication and production approval remain separate operational work.

The 0.1.2 preparation groups the municipal corrections after the earlier 0.1.1
preparation, keeping its dated history and notes unchanged. OpenAPI and MCP
metadata identify 0.1.2; release notes describe the source-specific opt-in,
upgrade requirements and pilot limits. Promotion can include multiple prepared
versions since the last published baseline without claiming that each was tagged.

## Validation

The 0.2.0 minor preparation groups subsequent development history, aligns OpenAPI
and MCP implementation metadata and documents the independent embedding controls,
local processing and development-only publication. Every processing replica must
be replaced after the additive migration to enforce disabled embedding defaults.
Promotion uses the existing pull-request, merge-commit and exact-revision workflow
gates; the annotated tag and GitHub release identify the verified main revision.

The first 0.2.0 Coverage run exposed a fixture-capacity deadlock on two-CPU runners:
four competing maintenance transactions exhausted the default four-connection
pool while the owning callback needed another connection. Reproduction with
two-CPU affinity confirms the failure. Give only this concurrent fixture an
explicit five-connection pool; keep serialization, cancellation and lock-release
assertions, the runtime locking protocol and the coverage floor intact.

Use temporary synthetic Git repositories to verify automatic insertion, idempotence, blocked unstaged changelog edits, and CI detection of a bypassed hook. Run the repository's Go and static checks after integration.
