# Spec Delta

## Purpose

Deliver an evening operator email that accurately summarizes retained regional warnings and municipal measures for enabled territories, including evidence, freshness and unavailable information.

## ADDED Requirements

### Requirement: Daily schedule and territorial scope
The service SHALL support an independently configurable daily email report at a local wall-clock time and IANA timezone, using the privately configured SMTP relay and recipients. It SHALL include every enabled region and each municipality with at least one collection-enabled municipal source in that region, deduplicated by territorial identity. Disabled regions and municipalities without enabled collection SHALL be excluded. An enabled region without a supported feed SHALL be listed with unavailable coverage. The report SHALL remain private to configured operator recipients and SHALL NOT change public source acceptance or activation.

#### Scenario: Scheduled evening report
- **WHEN** the configured daily local time is reached and report delivery is enabled
- **THEN** one report for that local date is queued for the configured recipients, covering the enabled territories at generation time

#### Scenario: Territory is disabled
- **WHEN** a region is disabled before report generation or all municipal collection sources for a municipality are disabled
- **THEN** that region or municipality is excluded from the generated report

#### Scenario: Region has no supported feed
- **WHEN** an enabled region has no compatible collecting source
- **THEN** the report lists its warning state as unavailable, rather than omitting it or implying no alert

### Requirement: Evidence-backed warning and measure summary
The report SHALL be in Italian and identify its observation time and timezone. It SHALL group regional warnings by region, product, risk and alert zone with supported level and validity, and list municipal measures and operational phases separately from regional colors. It SHALL distinguish currently applicable information, explicitly supported next-day information and uncertain validity. It SHALL include original source links and source check/acquisition freshness. Unsupported interpretation, missing territorial mapping, stale data, pending newer documents, publication limitations and unknown validity SHALL be visible. Absence of supported facts SHALL NOT imply an all-clear, completeness or cancellation. Explicitly retained restrictions SHALL survive partial reopenings. The report SHALL use already retained data without triggering model, OCR or crawling work.

#### Scenario: Partial reopening and unknown validity
- **WHEN** a notice reopens one underpass while retaining cemetery and dog-area closures with no supported validity end
- **THEN** the report distinguishes the reopening from retained closures and labels uncertain temporal applicability without inventing a duration

#### Scenario: Newer document is uninterpreted
- **WHEN** older supported facts exist and a newer retained document has not been successfully interpreted
- **THEN** the report exposes the updating limitation and does not describe the older facts as a fully updated situation

#### Scenario: Explicit next-day warning
- **WHEN** a retained regional warning explicitly applies to the next local calendar day
- **THEN** it appears in the next-day section with its supported validity, distinct from the current situation

#### Scenario: Missing facts or mapping
- **WHEN** no supported facts or reliable municipal-zone mapping exist for an included territory
- **THEN** the report states that information or mapping is unavailable and does not infer a green level or municipal maximum

### Requirement: Durable bounded delivery and failure visibility
The service SHALL persist the generated report and daily identity before attempting delivery, deduplicate generation across concurrent workers and restarts, retry SMTP failures with bounded backoff and retain a stable Message-ID for retries. It SHALL report sent status only after SMTP DATA acceptance and SHALL expose generation and delivery failures in private administration. A restart after the configured time SHALL generate the current local date's missing report without backfilling earlier dates. A queued report SHALL retain its original snapshot on delivery retries. Preview SHALL NOT send email or consume the scheduled daily identity. Report generation failures SHALL NOT send a normal-looking empty report or interrupt existing incident notifications. Large reports SHALL declare truncation and preserve aggregate scope/counts rather than silently dropping territories.

#### Scenario: Restart and concurrent scheduling
- **WHEN** multiple workers evaluate the same local date or restart after that date's send time
- **THEN** at most one report snapshot is queued for that date and retries preserve that snapshot

#### Scenario: SMTP is unavailable
- **WHEN** SMTP rejects delivery of a queued report
- **THEN** the report remains pending with a visible failure and bounded next-attempt time, and incident delivery continues independently

#### Scenario: Daylight saving transition
- **WHEN** the timezone's offset changes
- **THEN** the report remains scheduled at the configured local evening time with one daily identity per local date

#### Scenario: Read-only preview
- **WHEN** an operator previews the report
- **THEN** the same rendering and territorial scope are shown without sending mail or changing scheduled delivery state

#### Scenario: Database read fails
- **WHEN** report generation cannot obtain the required retained state
- **THEN** generation failure is visible and retried without sending an empty all-clear report
