# Proposal

## Why

The MVP administration currently exposes plain HTML, generic database tables and JSON mutation forms. Operators need a compact, readable console that makes failures, pending work and the next action apparent while preserving the existing operational safeguards.

## What Changes

- Introduce a shared dark terminal-inspired layout, compact navigation, semantic status labels and responsive tables across administrative HTML pages.
- Replace the link-only landing page with an operational overview of source issues, failed jobs, documents awaiting interpretation and incidents, with honest scope and freshness information.
- Present section-specific columns, contextual actions, expandable evidence and URL-preserved filters; use Jobs as the first complete implementation slice.
- Provide guided forms for job relaunch and routine source controls, retaining concurrency checks, explicit operator attribution and JSON client compatibility.
- Add optional in-page search and discoverable keyboard navigation; core navigation, filtering and submissions remain functional without JavaScript.
- Serve bundled local assets under the existing private administrative boundary and narrowly extend its content security policy.

## Capabilities

### New Capabilities

- `admin-dashboard`: Compact operational administration, accessible navigation, truthful summaries, readable detail views and guided routine actions.

### Modified Capabilities

None. The main spec inventory is currently empty. Existing isolation, source lifecycle, immutable evidence and accounting requirements in the active `define-toscana-alert-service` change remain constraints; this change adds presentation behavior without changing those contracts.

## Impact

- `internal/server/admin*.go`, shared HTML templates and new bundled CSS/JavaScript assets.
- `internal/operations` and runtime wiring for read-only overview aggregates and state filtering.
- Administrative HTTP and PostgreSQL integration tests, browser verification and `deploy/administration.md`.
- Existing public API/MCP, processing policy and source activation prerequisites remain unchanged. No SPA framework, external asset service, new login or deployment is required by this change.
