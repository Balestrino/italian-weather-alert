# Contributor agent guidance

Read [OpenSpec plans and progress](openspec/README.md) before changing an existing capability. Keep the relevant specification and task checklist in the same change as implementation and validation. A checked task in the imported snapshot records private project progress; verify the public checkout before claiming it is complete here.

Use synthetic or otherwise redistributable fixtures in this repository. Keep unpublished operational evidence, credentials, source captures, and machine-specific agent configuration out of commits. The [coverage tracker](docs/coverage.md) is the public source-status reference; a completed code task does not establish source acceptance or public availability.

Repository-provided agent workflows and their requirements are described in [Agent tools](docs/agent-tools.md).

Record every plan, fix, code, documentation, or operations change in [CHANGELOG.md](CHANGELOG.md). Before committing, stage the work and run `python3 scripts/changelog.py record --kind Changed --summary "<what changed>"`; choose the appropriate kind. The shared pre-commit hook inserts a fallback entry when this step is missed, and CI requires a dated entry in every change commit. Keep entries concise and free of unpublished operational evidence.
