# Proposal

## Why

Municipal sites change their page templates without changing a notice. Each change creates a new retained version. Interpretation reuse is limited to the same processing configuration, so after a configuration revision the new version is classified and extracted again. If that run fails, or has not run yet, the public document view reports the latest version as `failed` or `not_processed`. It does not mention that an earlier version with the same interpretation input has a validated interpretation.

On 2026-10-22 ([CLN-036](../../../../docs/territori/09-toscana/comuni/050004-calcinaia.md#cln-036--interpretazioni-pendenti-diagnosticate-dalle-risposte-conservate)) three Calcinaia weather notices were in this state: the 9 September criticality notice, the 21 August monitoring update and the 7–8 October alert. Their earlier measures are still published with evidence from the earlier version. The document, however, appears failed, and its interpretation no longer matches the measures it supports. Every future configuration revision repeats the pattern for every relevant notice whose template changes. That revision includes the pending development adoption of extraction configuration v29 (task 31.2 of `define-toscana-alert-service`).

## What Changes

- Define an **equivalent interpretation input** that ignores configuration. Two versions of the same document qualify when they have the same interpreted text sections, metadata, completeness, issuer and resource roles, eligibility and policies. Raw bytes of HTML/text resources do not have to match when their interpreted text does. OCR resources must have identical page outputs. Incomplete or non-reusable manifests never qualify.
- In public document reads, when a version's own interpretation at the knowledge boundary is `failed` or `not_processed`, present the most recent validated interpretation of an earlier equivalent version of the same document. The response labels it as carried forward: the original version, run and configuration, plus the version's own outcome and failure code, as an interpretation limitation.
- The version's own result wins as soon as one is recorded, whatever it concludes: supported, not relevant, partial or unreliable. Carry-forward never hides a newer result and never crosses documents or sources.
- Presentation only: no retained result, receipt, measure or projection is written, copied or rewritten. Historical queries carry forward only what both versions had at that knowledge boundary.
- API and MCP use the same read path, so their responses stay identical. The search, document detail, versions list and saved views agree.
- Development adoption of extraction v29 and the bounded reprocessing for task 31.2 follow this change.

## Capabilities

### New Capabilities

- `equivalent-interpretation-carry-forward`: configuration-independent equivalence of interpretation inputs between versions of one document, and presentation of the latest earlier validated interpretation, with provenance and limitations, when a version's own interpretation failed or is pending.

### Modified Capabilities

None in `openspec/specs/`. The public document interpretation contract still lives in the unarchived `define-toscana-alert-service` change. This capability refines how it is presented without changing those requirements.

## Impact

- `internal/backend/publicquery`: document interpretation reads (`loadDocumentInterpretations`, `documentVersion`, document search and saved views) resolve carry-forward in batches. Query counts stay bounded, with no extra reads per version.
- `internal/backend/classification`: a configuration-independent equivalence fingerprint derived from the input manifest. It is persisted additively for new runs and backfilled from retained manifests, so public reads do not parse manifest JSON.
- Additive migration only. No change to processing schedules, configuration identities, model calls, publication gates or acceptance.
- Public response shape: new interpretation limitation strings and an evidence locator such as `equivalent_version_interpretation`. Existing fields keep their meaning.
