# Proposal

## Why

The September 22, 2026 production audit found 4,163,595 recorded tokens since September 18, including 1,282,103 tokens spent OCRing identical page images again. Frequent evidence-version churn, thousands of provider rejections, invalid structured results and 52 worker restarts require both cost controls and reliability fixes without weakening source coverage or evidence validation.

## What Changes

- Persist useful, sanitized provider failure diagnostics and per-call usage, including interrupted or unknown outcomes; diagnose and correct the actual rejection cause.
- Contain provider-wide failures across newly created jobs, respect retry limits and Retry-After, and expose recovery controls.
- Diagnose worker exits, preserve queue identity and safe error codes, correct the reproduced cause, cancel work on lost leases and recover durable work without duplicated effects.
- Reuse OCR by exact page content and processing configuration across document versions, with concurrency control and explicit provenance.
- Distinguish retained raw-byte revisions from meaningful interpretation changes; retain originals while avoiding inference for verified equivalent evidence.
- Select document and bulletin graphics explicitly, excluding confirmed decorative resources without suppressing maps, legends or image-only notices.
- Reduce paid invalid outputs through evaluated request/schema fixes and segment checkpoints, retaining strict literal-evidence validation.
- Add usage/reuse/failure visibility, historical replay evaluation and staged rollout with bounded recovery rather than automatic history replay.

## Capabilities

### New Capabilities

- `inference-efficiency`: Evidence-preserving OCR reuse, interpretation equivalence and explicit inference eligibility.
- `inference-operations`: Provider diagnostics, failure containment, worker recovery, validated output handling and truthful usage measurement.

### Modified Capabilities

None in the main spec store, which is currently empty. These additive capabilities complement the pending `official-source-ingestion` capability in `define-toscana-alert-service`; its raw evidence, complete-content, retry, quality and publication safeguards remain applicable. Do not duplicate or rewrite that change during implementation.

## Impact

Affected areas: `internal/inference`, `internal/processing`, `internal/ocr`, `internal/acquisition`, `internal/interpretation`, classification/extraction runners, `internal/jobs`, local administration, `cmd/iwa/main.go`, deployment documentation and PostgreSQL migrations. Public result provenance and retention must remain correct when outputs are reused. No public API removals, model replacement, reduced polling cadence, automatic historical reprocessing or raw evidence deletion are proposed. Deployment and provider-account changes are separate from this planning request.
