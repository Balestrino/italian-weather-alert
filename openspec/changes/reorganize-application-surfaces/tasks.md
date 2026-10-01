# Tasks

## 1. Component boundaries

- [x] 1.1 Extract backoffice and neutral HTTP utilities, group backend services, and update composition; verify existing HTTP isolation and unit tests.
- [x] 1.2 Move embedded contracts to `api/public` and update Docker, fixtures and CI paths; verify contract tests and all command builds.

## 2. Public frontend

- [x] 2.1 Add the independent public status command and embedded UI with bounded reads, error states and cursor navigation; verify synthetic availability, escaping, pagination, limits and route-isolation tests.
- [x] 2.2 Bundle the frontend binary and provide an explicit optional standalone Compose example; verify image build and resolved configuration without starting project services.

## 3. Documentation and final verification

- [x] 3.1 Reorganize current guides, add architecture/component documentation and update contributor links and OpenSpec index; verify local Markdown links and current path references.
- [x] 3.2 Run race, vet, build, synthetic database integration and deployment checks, record actual results and a dated changelog entry; verify final diff and checklist against the public checkout.

## Validation record — 1 October 2026

The public checkout passed `go test -race ./...`, `go vet ./...`, command builds,
the complete synthetic `-tags=integration -p 2` suite, environment/changelog
regression tests and deployment-template validation. The CI coverage command
including `api/public` measured 68.8% statement coverage (66% required).
`govulncheck@v1.8.0` reported no reachable vulnerabilities. The Docker image built
with all three packaged binaries. Standalone frontend Compose configuration was
checked for loopback publication and absence of credentials, mounts and private
data-network access. The authoritative JSON contracts are byte-for-byte unchanged;
local Markdown links and production package boundaries passed their checks.

The public status page was checked with a synthetic HTTP backend in Chromium at
390px, 768px and 1440px, at 200% zoom, with keyboard navigation and JavaScript
disabled. Private administrative browser acceptance remained optional and was
not rerun; its assets were moved unchanged. These results do not certify source
acceptance, public routing, deployment or operational release gates.
