# Public progress snapshot

These checkboxes reflect the private project task list on 30 September 2026. A checked item records project work; it does not by itself establish public source acceptance, public deployment, or that unpublished evidence is available in this repository. Operational notes and private evidence references are maintained separately.

## 1. Private schedule and SMTP transport

- [x] 1.1 Add backward-compatible optional daily_report configuration, validated evening local time and IANA timezone; use the operator-selected 21:00 Europe/Rome schedule, and verify invalid schedules, missing configuration and timezone/DST behavior with unit tests.
- [x] 1.2 Extend the shared SMTP transport for report subject/body and stable identity while preserving incident delivery; verify header-injection rejection, private recipients, TLS/no-downgrade behavior and unchanged incident messages with loopback SMTP tests.

## 2. Territorial snapshot and report rendering

- [x] 2.1 Add a consistent report reader selecting enabled regions, current territorial datasets and deduplicated municipalities with collection-enabled sources; reuse scoped private temporal/evidence reads and verify disabled regions, suspended sources, unaccepted collection-only scopes and unsupported enabled profiles with isolated PostgreSQL integration tests.
- [x] 2.2 Render deterministic bounded Italian reports containing separate regional risks/zones/levels, municipal measures/phases, current/next-day/uncertain validity, original links, freshness and pending/partial coverage; verify partial reopening, conflicting dates, missing maps, newer failed interpretation, empty scope and explicit truncation with retained-case/unit tests, without inference calls.

## 3. Durable scheduling and private visibility

- [x] 3.1 Add an additive migration and durable local-date report snapshot/outbox with locking, generation failure tracking, stable Message-ID and bounded delivery backoff; verify concurrent generation, restart after send time, no historical backfill, snapshot-preserving retries and SMTP acceptance with isolated PostgreSQL/SMTP integration tests.
- [x] 3.2 Wire bounded report scheduling/delivery into the worker independently of incident reconciliation/delivery, and add private read-only preview plus generation/delivery status; verify defaults leave scheduling disabled, preview does not enqueue, report failures do not block incident email, and public endpoints cannot access report data.

## 4. Acceptance and operational activation

- [x] 4.1 Run Go unit/race/vet and relevant isolated integrations, strict OpenSpec validation and a read-only preview against current retained production data; retain the rendered review artifact and evidence that generation incurs zero model/OCR/crawl calls and accurately reports current enabled scope and limitations.
- [ ] 4.2 Prepare and review an isolated report-only deployment image based on the running application revision, preserving existing provider guards, configuration/catalog/source selections and secret readability; enable the selected schedule, restart the worker and record status plus the first scheduled report's SMTP acceptance without claiming recipient inbox receipt until confirmed. Verify rollback instructions and do not bundle unrelated pending changes.
