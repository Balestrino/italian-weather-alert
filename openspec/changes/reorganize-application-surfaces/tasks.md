# Tasks

## 1. Component boundaries

- [x] 1.1 Extract backoffice and neutral HTTP utilities, group backend services, and update composition; verify existing HTTP isolation and unit tests.
- [x] 1.2 Move embedded contracts to `api/public` and update Docker, fixtures and CI paths; verify contract tests and all command builds.

## 2. Public frontend

- [x] 2.1 Add the independent public status command and embedded UI with bounded reads, error states and cursor navigation; verify synthetic availability, escaping, pagination, limits and route-isolation tests.
- [x] 2.2 Bundle the frontend binary and provide an explicit optional standalone Compose example; verify image build and resolved configuration without starting project services.

## 3. Documentation and final verification

- [ ] 3.1 Reorganize current guides, add architecture/component documentation and update contributor links and OpenSpec index; verify local Markdown links and current path references.
- [ ] 3.2 Run race, vet, build, synthetic database integration and deployment checks, record actual results and a dated changelog entry; verify final diff and checklist against the public checkout.
