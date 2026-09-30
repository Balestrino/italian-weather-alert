# Proposal

## Why

The operator currently receives incident emails but must open administration to review the weather warning situation. An evening email should summarize retained warnings and municipal measures for enabled territories, while exposing gaps and stale information instead of presenting missing data as an all-clear.

## What Changes

- Add one daily Italian-language evening report delivered through the existing private SMTP relay and recipients.
- Select enabled regions and municipalities with collection-enabled municipal sources; report collection-only scopes with their publication/acceptance limitations.
- Include supported regional risk/zone/level/validity, distinct municipal measures and operational phases, current and tomorrow information where explicitly dated, original links and acquisition/check freshness.
- Represent unsupported profiles, pending interpretation, missing mappings, unknown validity, partial reopenings and unavailable/stale sources explicitly.
- Add a configurable local wall-clock schedule (21:00 Europe/Rome, selected by the operator), persistent report snapshots and an outbox with restart-safe daily deduplication and delivery retries.
- Provide private preview and delivery status; preserve existing incident emails and public activation gates. No extra model/OCR calls.

## Capabilities

### New Capabilities

- `evening-alert-email-report`: Scheduled private reports of warning and municipal measure status across enabled territories, with evidence, freshness, bounded delivery and observable failures.

### Modified Capabilities

None. Existing draft requirements in `define-toscana-alert-service` remain intact; this adds an operator delivery channel without changing public publication rules.

## Impact

Touches `internal/notifications`, private notification configuration, worker wiring in `cmd/iwa`, new report read/storage/rendering components, PostgreSQL migrations and private administration/CLI preview. Reuses territorial associations and private evidence/temporal readers in `internal/publicquery/administrative.go`. SMTP credentials and recipients stay in the existing secret file. Operational deployment must preserve the current provider guards and output-fix source selection; it must not inadvertently roll out other unselected changes.
