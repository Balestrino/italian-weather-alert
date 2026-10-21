## Purpose

Let anonymous visitors inspect service availability, published source coverage and source updating information without gaining access to private operational data.

## ADDED Requirements

### Requirement: Independent public status interface
The frontend SHALL run without database, storage, crawler or inference credentials and SHALL remain able to render a status page when the configured public backend is unavailable. It SHALL use public read-only HTTP interfaces and SHALL NOT expose administrative paths or act as a general-purpose proxy.

#### Scenario: Backend is offline
- **WHEN** the public backend cannot be reached
- **THEN** the frontend renders a usable page reporting unavailable information without treating failure as no sources or healthy service

#### Scenario: Private path requested
- **WHEN** a visitor requests an administrative path through the frontend
- **THEN** it returns not found without forwarding the request

### Requirement: Distinct availability coverage and freshness
The frontend SHALL distinguish technical availability, source publication state, declared coverage and updating state. Source checks and response observation times SHALL be labeled separately. Pending, suspended, delayed, unknown and unavailable information SHALL remain explicit; missing sources SHALL NOT imply absence of alerts or nationwide coverage. Values SHALL be escaped and the page SHALL work without JavaScript.

#### Scenario: Service is available but collection is delayed
- **WHEN** the backend is ready and a returned source is delayed
- **THEN** the page reports service availability and that source's delayed updating independently

#### Scenario: No publicly eligible sources returned
- **WHEN** the coverage query succeeds with an empty result
- **THEN** the page explains that no sources were returned for the requested scope and makes no all-clear claim

### Requirement: Honest bounded coverage navigation
The frontend SHALL preserve the public API's pagination boundary and expose a link to subsequent pages when a cursor is returned. It SHALL distinguish malformed responses, query errors, expired cursors and rate limiting from successful empty data. Requests SHALL have bounded time and response size, and errors SHALL NOT reveal backend addresses or internal response bodies.

#### Scenario: More sources than fit on one page
- **WHEN** the public response includes a next cursor
- **THEN** the frontend labels its list as one page and provides navigation using that cursor without claiming complete counts

#### Scenario: Backend rejects a query
- **WHEN** the public query is rate limited, expired or unavailable
- **THEN** the coverage section reports the failure and offers a way to restart rather than displaying an empty successful result
