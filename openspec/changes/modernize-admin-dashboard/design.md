# Design

## Context

See proposal.md for motivation. `admin.go` serves a static index; `admin_operations.go` renders raw report columns and a separate list of row actions. Other administration pages use the templates in `admin_sources.go`. `adminDecode` accepts JSON or a form containing JSON in `payload`. Existing tests cover escaping, private route isolation, foreign origins, pagination and mutations.

`operations.Report` contains at most 100 rows, a `More` flag and an observation time; only usage reports currently have full-scope totals. Counting rendered rows cannot produce honest dashboard totals. The existing CSP starts with `default-src 'none'`, so local assets need explicit directives.

## Goals / Non-Goals

**Goals:** A coherent server-rendered administration surface, quick triage and usable routine actions with bounded read costs and progressive enhancement.

**Non-Goals:** Public-facing UI, processing changes, terminal emulation, live log streaming, charts without an operational purpose, global command execution, bulk mutations, SPA migration or automatic deployment. Full source configuration and specialized evaluation payloads remain available as labeled advanced JSON editors in this MVP; routine controls get dedicated fields.

## Decisions

### Shared server-rendered shell

Extract shared layout and components from embedded template strings into templates bundled with the Go binary. Use the same shell for home, all six operation sections, source management, backups, notifications, observations/trial costs, evaluations/regression and release pages, including HTML action results. Keep existing URLs and JSON representations. Compared with adding a SPA, this preserves direct links, form submissions and the current deployment model.

Sidebar: Panoramica, Fonti, Job, Documenti, Qualita, Consumi, Sistema. Fonti includes operational status and registry management; Qualita includes findings, comparisons and evaluations; Sistema groups backups, incidents, observation campaigns and release diagnostics. Nested pages highlight their parent and provide breadcrumbs. Health and status JSON endpoints remain linked diagnostics, not HTML pages to restyle.

### Visual defaults

Use CSS variables with proposed defaults: background `#0d1117`, surfaces `#161b22`, borders `#30363d`, primary text `#e6edf3`, secondary text `#9da7b3`, cyan accent `#58d5e8`. Status colors supplement text labels. Verify actual contrast during implementation rather than assuming palette compliance.

Use local system monospace stacks for navigation, IDs and tables, with a system sans-serif stack for longer explanations. Base text 14px, spacing scale 4/8/12/16/24px, sidebar about 200px, top bar about 48px, desktop row height about 36px and modest 4px radii. Avoid fixed row heights that clip wrapped content. Use tabular numerals, visible focus, underlined inline links and restrained selected/hover backgrounds. Mobile controls expand to comfortable touch targets; navigation collapses through an accessible button and wide tables scroll inside their own containers.

### Truthful overview through dedicated reads

Add an overview read model in `internal/operations` with snapshot-consistent aggregate queries and a timestamp. Count sources in delay/error, failed jobs and retained document versions needing interpretation using the same predicates as their destination lists; count entities, not joined attempts. Link each summary to the corresponding filtered view. Read open incidents through the notification report with its own observation time. Mark a failed subsection unavailable while keeping successful subsections useful. Do not traverse paginated reports or infer service readiness from `/admin/status`.

MVP refresh is explicit navigation/reload, with a visible observation time; no automatic refresh or live claim. Costs remain on the usage page with workload, currency, partial usage and pricing provenance intact.

### Section presenters and filters

Introduce presentation mappings for meaningful Italian labels, primary columns, timestamp/unit formatting and status labels. Preserve all remaining fields in expandable row details, including raw error codes and evidence links. Place contextual links/actions in the same row/detail region, eliminating the duplicate action list. Keep empty lists, unavailable data and actual zero values distinct.

Use server-side GET filters with URLs: retain source, version, job and run identifiers where supported; add explicit allowlisted job-state/source-issue filters needed for overview drilldowns. Changing a filter resets pagination, while pagination preserves all applicable filters. A separate optional client-side search explicitly says it searches only the current page.

### Guided routine forms

Implement native form decoding for job relaunch and source preview, collection activation/pause, intervals, acceptance, public enablement, interpretation pause/resume and selective reprocessing. Use named controls, clear labels, operator attribution and visible target/scope. Carry observed revision/attempt tokens and retain server validation and conflict rejection. Never default into accepting, publishing or relaunching multiple records. Preserve JSON requests and existing payload-form compatibility for clients.

Provide HTML success/error views with an actionable return link, safe field errors and a reload instruction for conflicts. Keep JSON statuses and safe machine error codes stable. Retain advanced source configuration, creation and diagnostic JSON editors with clear descriptions; do not silently discard nested fields.

### Assets and optional JavaScript

Serve embedded CSS and deferred JavaScript at private `/admin/assets/` paths. Add only `style-src 'self'` and `script-src 'self'` to the existing CSP; no CDN, remote fonts, inline script/style exceptions or eval. Preserve Host/Origin, no-store and public-router isolation.

Use JavaScript for navigation toggling, explicitly current-page search and optional keyboard shortcuts. `/` focuses search when available; Escape clears search or closes navigation and returns focus. Ignore shortcuts in inputs, textareas, selects, editable regions and modifier-key combinations. Provide visible help and a way to disable single-character shortcuts. No action executes through a shortcut. Native links, details and forms work without JavaScript.

## Risks / Trade-offs

- Dense tables can hinder reading on narrow screens -> wrap long evidence, use expandable details and local table scrolling; test at 390px, 768px and 1440px plus 200% zoom.
- Styling many template families can omit secondary routes -> maintain an HTML route inventory and verify each family, including result/error pages.
- Aggregate/list predicates can diverge -> shared query semantics and PostgreSQL tests with more than 100 records, repeated attempts and mixed states.
- Form refactors can weaken mutation semantics -> preserve JSON compatibility and test stale revisions, duplicate fields, invalid values and attribution.
- Browser automation may not be installed -> provision an isolated test tool if needed and record actual browser evidence; do not substitute string checks for visual acceptance.

## Migration Plan

Implement and validate the shared shell with Jobs first, then overview, remaining operation sections, source controls and secondary pages. Bundle assets into the binary; no database migration is expected for presentation and aggregate reads. Run focused tests per task and final repository checks, document screenshots and findings, then leave deployment as a separately authorized action. Rollback restores the previous binary/assets together; persisted operational records are unaffected.

### Today/tomorrow extension
Add `/admin/alerts` with two responsive day columns, each separating regional
warnings and municipal measures. Read existing public-query validated facts with
the current knowledge boundary, preserving publication/evidence/status safeguards.
Supplement them with latest retained regional bulletin metadata from an explicit
read-only administrative projection, without promoting aggregated HTML labels to
zone-specific warnings. Parse only explicit Italian calendar dates from retained
validity expressions; never infer validity from acquisition time. Show undated
items separately, preserve source links and collection timestamps, and clearly
label missing/unconsolidated information. No new provider calls or database schema.
