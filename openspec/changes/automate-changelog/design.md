## Context

Git hooks are local and are not installed by cloning a repository. Git's pre-commit hook can change the index before a commit is written; later message hooks cannot reliably add files to that commit. CI therefore checks committed history independently.

## Decisions

Keep one human-readable `CHANGELOG.md` with dated Unreleased bullets. `scripts/changelog.py record` reads staged paths, inserts one entry under Unreleased, and stages the file. An explicit kind and summary produce a useful entry; the hook uses a path-based fallback if none was supplied. Existing unstaged changelog edits are never staged implicitly.

`scripts/changelog.py check --base` inspects every commit since the merge base. A commit that changes any other repository file must add a dated changelog bullet in that same commit. CI runs this check for pushes and pull requests with full Git history. A required status check on protected branches makes it a merge gate; local hook installation improves convenience but is not trusted as enforcement.

The initial changelog records the current fix and this mechanism. Entries describe repository changes only. They do not state that a source was accepted, the service was deployed, or private evidence was published.

## Validation

Use temporary synthetic Git repositories to verify automatic insertion, idempotence, blocked unstaged changelog edits, and CI detection of a bypassed hook. Run the repository's Go and static checks after integration.
