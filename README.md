# Italian Weather Alert (IWA)

[![CI](https://github.com/Balestrino/italian-weather-alert/actions/workflows/ci.yml/badge.svg)](https://github.com/Balestrino/italian-weather-alert/actions/workflows/ci.yml)
[![Coverage](https://github.com/Balestrino/italian-weather-alert/actions/workflows/coverage.yml/badge.svg)](https://github.com/Balestrino/italian-weather-alert/actions/workflows/coverage.yml)
[![License: GPL-3.0-only](https://img.shields.io/badge/License-GPL--3.0--only-blue.svg)](LICENSE)

![Go 1.27](https://img.shields.io/badge/Go-1.27-00ADD8?logo=go&logoColor=white)
![PostgreSQL 16](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white)
![RustFS](https://img.shields.io/badge/RustFS-S3%20storage-5C4EE5)
![Docker Compose](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)
![Crawl4AI](https://img.shields.io/badge/Crawl4AI-source%20collection-2E8B57)
![MCP](https://img.shields.io/badge/MCP-Streamable%20HTTP-6B46C1)

**Our goal is to make all official regional weather and hydrogeological alerts in Italy, together with related municipal notices and measures, publicly accessible through a single JSON API and Model Context Protocol (MCP) interface.**

IWA connects each result to its issuing authority, affected area, validity period, source documents, and known limitations. It keeps a regional warning distinct from a municipality's own decisions, such as a closure or an emergency operations activation.

> **Current status:** The API and MCP interface are implemented and tested locally. Source collection is in an internal Tuscany pilot: regional vigilance, criticality/alert, and monitoring products, plus a Calcinaia municipal source. Public source access is not enabled, and there is no public MCP endpoint yet. Nationwide coverage is the goal, not the current state. See the [coverage tracker](docs/coverage.md) for the documented status of every region and municipality.

> **Unofficial, best-effort project:** IWA is an independent tool, not an official source of alerts or instructions and not endorsed by any public authority. Its information may be incomplete, delayed, unavailable, or incorrect. Consult the issuing regional and municipal authorities directly for current warnings and instructions; their original publications remain authoritative.

## Why this project exists

Official information is spread across regional bulletins, municipal websites, ordinances, and channels that municipalities explicitly recognize. A municipality can repeat a regional warning and also issue its own measures; those are different facts from different authorities. IWA aims to make both findable without losing their provenance.

A municipality's presence in the national registry does **not** mean its alert sources have been checked or enabled. An empty result does **not** prove that there are no warnings or measures.

## What IWA provides

- **Source-aware collection:** configured official channels, versioned original documents and attachments, and a record of when each source was checked.
- **Evidence-linked interpretation:** regional products, municipal notices, and local measures remain separately attributed. Extracted facts point back to supporting documents and expose uncertainty.
- **Geographic and time context:** municipality lookup by name or ISTAT code, CAP candidates where a verified mapping is configured, alert-zone applicability, validity periods, and retained history.
- **Two read-only interfaces:** a JSON API and an MCP endpoint expose the same public query groups: municipality discovery, situation, alert search, document detail, and source coverage.

The implementation uses Go, PostgreSQL for structured data and durable jobs, RustFS for retained originals, Crawl4AI for source collection, and Docker Compose for the local stack. Inference providers are configurable; their output does not replace source evidence.

## Application components

The repository keeps one Go module with separate areas for the [backend](docs/backend/README.md), [private backoffice](docs/backoffice/README.md) and [public status frontend](docs/frontend/README.md). Backend services live in `internal/backend`, administration in `internal/backoffice`, and public presentation in `internal/frontend`. Public API/MCP contracts live in `api/public`.

The status frontend runs independently through `go run ./cmd/iwa-frontend` and reports backend availability, source publication coverage and updating information separately. It uses public HTTP reads without database or operator credentials; it can display unavailable backend information. It does not activate a public deployment or enable alert sources. See the [architecture guide](docs/architecture/README.md) for dependencies and the layout migration.

## Coverage today

The first source rollout is in **Tuscany**. The internal pilot covers the three distinct regional products and the Calcinaia municipal channel. Livorno, Pisa, Pontedera, and Cascina have candidate municipal configurations at different preview stages; each still needs its own observation and acceptance before public enablement.

The [national coverage tracker](docs/coverage.md) lists all 20 regions and their municipalities, with ISTAT codes and CAPs. It is a planning and verification inventory, not a live alert map. Tuscany is the first source pilot; the tracker records the status of its regional and municipal channels.

## For citizens

When public access opens, you will be able to connect an MCP-compatible assistant to IWA and ask, for example:

> Which official weather alerts apply to Pisa today? Has the municipality published related measures? Show me the original sources and when they were last checked.

The assistant should show the regional warning and any municipal measures separately, with links to the issuing authorities and any coverage gaps. IWA does not issue alerts or provide emergency instructions. For decisions about safety, follow the official authorities and their current channels.

## For developers

From a checkout of this repository, with Go 1.27.1:

```sh
go test ./...
go vet ./...
go build ./cmd/...
go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...
```

The [developer setup guide](docs/development/setup.md) walks through the local Compose stack, secrets, migrations, and optional worker configuration. [Verified JSON API and MCP examples](docs/backend/public-usage.md) show the five query groups against a local service. The [API/MCP contract](docs/backend/public-api-mcp.md) describes the interface in more detail.

The [environment guide](docs/operations/environments.md) explains how the development, isolated staging and prepared production projects on one shared host relate. [Release operations](docs/operations/releases.md) describe staging checks, manually approved GHCR image digests and the production backup/restore targets. Publishing the code does not by itself activate a public service or an alert source.

The **CI** badge reports the Go test, vet, and build workflow. The **Coverage** badge reports whether the Go unit and isolated database integration tests passed a 66% statement coverage floor; the measured percentage appears in that workflow's run summary. A separate security workflow checks for known reachable Go vulnerabilities on pushes, pull requests, and weekly. None of these checks measures the completeness of alert-source coverage.

The public test suite keeps synthetic cases. Raw source captures and internal regression evidence are retained separately until their redistribution terms and contents have been reviewed.

## Help us build and test nationwide coverage

Reliable coverage needs both software work and people who know the local sources. Developers can improve integrations, parsing, API/MCP behavior, and tests. Citizens and other testers can compare IWA results with official bulletins and municipal publications, then report missing information, incorrect interpretations, or unclear limits.

See [CONTRIBUTING.md](CONTRIBUTING.md) for development checks and a reproducible report format. Source acceptance and public enablement remain separate for every regional and municipal channel.

## Warranty and liability

The code and the information it produces are provided on a best-effort, **“as is” and “as available”** basis, without any warranty of accuracy, completeness, timeliness, fitness for a particular purpose, or uninterrupted operation. To the extent permitted by applicable law, IWA's maintainers and contributors are not liable for loss or damage arising from use of the code or its output. This notice does not exclude liability that applicable law does not allow to be excluded.

## Documentation and reuse

- [Documentation index](docs/README.md) and [national coverage tracker](docs/coverage.md)
- [OpenSpec plans and progress](openspec/README.md)
- [Agent workflows and optional tools](docs/development/agent-tools.md)
- [Developer setup](docs/development/setup.md), [public API/MCP examples](docs/backend/public-usage.md), and [API/MCP contract](docs/backend/public-api-mcp.md)
- [Development, staging and production](docs/operations/environments.md) and [release operations](docs/operations/releases.md)
- [Changelog](CHANGELOG.md) and [v0.1.0 release notes](docs/releases/v0.1.0.md)
- [Third-party notices](THIRD_PARTY_NOTICES.md)

Original IWA code is licensed under [GPL-3.0-only](LICENSE). Selected third-party data and retained source files have separate terms recorded in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md). The public repository contains only reviewed source snapshots; the broader internal research evidence is kept separately.
