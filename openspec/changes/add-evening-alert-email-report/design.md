# Design

## Context

See proposal.md for motivation. Existing `internal/notifications` detects durable source-delay, processing-failure and backup incidents and sends plain-text email from a PostgreSQL outbox. SMTP authentication and actual recipient delivery have been verified. The current SMTP sender accepts incident-specific `Message` values and builds an incident subject/body. It therefore needs a transport extension for report subject/body without fabricating an operational incident.

Territorial state is stored in `territorial_regions`, current region configuration and source associations. A municipality has no separate enabled boolean: active collection on a municipal source supplies that boundary. `publicquery.ReadTerritorialFacts` is a private regional reader sharing evidence and temporal semantics without public acceptance gating. In contrast, `operations.Alerts` currently mixes collecting bulletin metadata and publicly visible facts for today/tomorrow; copying that view would hide private pending facts and fail to scope every enabled region correctly. Private reads must be filtered to collection-enabled sources and the selected current territorial datasets before report assembly.

The worker runs as UID 65532. The SMTP file has a read-only ACL for that UID under an owner-private directory. The production image is older than the current checkout, which includes independently selected extraction corrections; rolling out a full checkout image would have unrelated effects.

## Goals / Non-Goals

**Goals:** One auditable daily snapshot from retained state, independent report delivery and incident delivery, reuse of established temporal/evidence semantics, bounded resource consumption and observable missing coverage.

**Non-Goals:** Weather inference, new crawling/OCR/LLM calls, public email subscription, new geography imports, automatic activation of sources or unselected processing changes, and reconstructing a whole day's event chronology from a single end-of-day snapshot.

## Decisions

1. Add a private optional `daily_report` configuration object to the notification file, absent by default for compatibility. It carries enabled, local time and timezone; relay, from and recipients remain shared. The operator selected 21:00 Europe/Rome; this is the initial enabled schedule. Restrict this evening schedule to 18:00–23:59 so DST changes cannot create ambiguous/nonexistent send hours; validate IANA timezone, including embedded tzdata. Unlike a UTC cron or in-memory timer, a local-date identity preserves wall-clock behavior and restart recovery.

2. Add a dedicated report service and durable table/outbox keyed uniquely by report type and local calendar date. Transactional admission/locking prevents duplicate snapshots. Persist observed time, schedule timezone, subject/body, counts and limitations, attempts, next attempt, sanitized errors and stable Message-ID. Generate only today's due report after restart; prior undelivered queued reports can still retry. Read territorial state consistently; never convert a required query failure into an empty report. Record generation failures without consuming a successful generation slot. Keep report generation and delivery bounded, and run them independently of the incident loop so either path's failure cannot starve the other. A process crash after SMTP acceptance and before DB commit can duplicate delivery with the same Message-ID, as already documented for incidents.

3. Use private scoped territorial/evidence reads, filtering source IDs to collection-enabled sources under enabled regions. Deduplicate municipalities by ISTAT and selected current geography dataset. Preserve private source limitations, supported projection status, unknown/conditional/conflicting validity, newer interpretation failures and seven separate regional risks. Add scoped reader support only where existing private helpers cannot represent the required evidence; leave public query behavior unchanged. Count current/next-day/uncertain facts separately and include source freshness. Never infer a regional color from municipal measures, alert validity from page metadata, or a missing municipal-zone map.

4. Render deterministic Italian plain text with distinct region and municipality sections, current and explicitly dated next-day information, unknown applicability, original links, last complete checks, pending interpretations and coverage limitations. Bound the body to 256 KiB using UTF-8-safe truncation while retaining territorial counts and an explicit truncation notice. An absence of data produces an unavailable statement for that included territory, not a green/all-clear label. An empty enabled scope produces an explicit no-enabled-territories report.

5. Extend SMTP transport to accept a validated report subject/body and stable message identity while retaining the existing incident API and secure envelope behavior. Keep credentials and recipients out of persisted report contents, logs and administration. Header values must reject injection; source text remains data. Add a private preview/status route or equivalent admin-only CLI using the same renderer. Preview never queues or sends; report status distinguishes queued, delivery failed and SMTP accepted.

Alternatives considered: a host cron script lacks integration with durable delivery and deployment; reusing incident categories for reports would distort incident/recovery semantics; invoking a model for summaries creates unnecessary cost and unsupported interpretations; using the public API hides collection-only scopes and ties private operator coverage to publication gates.

## Risks / Trade-offs

- [Sparse or pending interpretations] → Display unavailable/partial/uncertain state alongside original bulletin links and freshness; a report cannot repair the blocked source acceptance.
- [Repeated SMTP delivery after a crash] → Persist snapshot and stable Message-ID, document SMTP acceptance boundary, retain the existing at-least-once practical behavior.
- [Huge territorial scope] → Deterministic bounded output, explicit truncation and complete scope/count summary; private full preview remains available.
- [Unrelated checkout changes enter production] → Prepare an isolated report-only deployment image based on the running revision; preserve provider guards, catalog selection and collection/public flags, and compare the delta before rollout.
- [Secret replacement loses ACL] → Validate UID 65532 readability before restart and preserve the current private access model.

## Migration Plan

Add a new numbered additive migration through the established migration registry. Missing daily_report configuration leaves existing notification behavior unchanged. Validate unit tests, isolated PostgreSQL/loopback SMTP integration, preview against current retained data and delivery rendering before enabling scheduling. Prepare a reviewed report-only image and deploy with existing environment/guards; enable the selected evening schedule in the private configuration and restart the worker. Rollback disables daily_report and restores the prior image without deleting report history. No live test email beyond explicitly authorized report delivery is required during local validation.
