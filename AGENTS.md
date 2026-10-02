# Contributor agent guidance

Read [OpenSpec plans and progress](openspec/README.md) before changing an existing capability. Keep the relevant specification and task checklist in the same change as implementation and validation. A checked task in the imported snapshot records private project progress; verify the public checkout before claiming it is complete here.

Use synthetic or otherwise redistributable fixtures in this repository. Keep unpublished operational evidence, credentials, source captures, and machine-specific agent configuration out of commits. The [coverage tracker](docs/coverage.md) is the public source-status reference; a completed code task does not establish source acceptance or public availability.

Repository-provided agent workflows and their requirements are described in [Agent tools](docs/development/agent-tools.md).

Before investigating or changing a regional or municipal source, read its [territorial guide](docs/territori/README.md) and referenced platform notes. Update the current guidance and dated discovery register as part of the source work; create a missing guide from the shared template when research begins. Distinguish confirmed observations, hypotheses, proposed fixes and verified behavior, with precise source scope and evidence dates. Keep private operational details under ignored `.local/operations/territori/` and source acceptance in the coverage tracker; a guide does not enable collection or publication.

Record every plan, fix, code, documentation, or operations change in [CHANGELOG.md](CHANGELOG.md). Before committing, stage the work and run `python3 scripts/changelog.py record --kind Changed --summary "<what changed>"`; choose the appropriate kind. The shared pre-commit hook inserts a fallback entry when this step is missed, and CI requires a dated entry in every change commit. Keep entries concise and free of unpublished operational evidence.
