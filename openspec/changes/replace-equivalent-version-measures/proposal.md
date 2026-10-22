# Proposal

## Why

Public measure search hides a measure when a newer validated extraction of the **same retained version** has projected measures. This replacement does not reach other versions of the same document. Each template-only version can be reinterpreted after a configuration revision or a bounded reprocessing. Its measures are then added to those of the earlier version, even though the interpreted text is identical.

On 2026-10-22 ([CLN-039](../../../docs/territori/09-toscana/comuni/050004-calcinaia.md#cln-039--misure-duplicate-tra-versioni-dello-stesso-avviso)), two of the seven Calcinaia documents with public measures showed duplicates from equivalent inputs:
- the 9 September criticality notice: 3 measures from version 4291 and 5 from version 28212, with only one linked;
- the criticality update: versions 325 and 28211.

Every reinterpretation of equivalent versions repeats the problem. This includes the regular template churn and the reprocessing of task 31.2.

## What Changes

- Extend the existing same-version replacement to versions with an **equivalent interpretation input** (`interpretation-equivalence-v1` fingerprint, same document and source). The rule: a measure projected from version A is not current when a version B of the same document with an equivalent input has a newer validated, complete ordinary extraction whose measures are projected, all known within the source cutoff at the query boundary. B is either A itself with a later run, or a version acquired after A.
- Genuine revisions with a different fingerprint keep today's behaviour. Their measures accumulate and are related only by linking, because disappearance never establishes cessation.
- Presentation only: no measure, binding, receipt or extraction is rewritten or deleted. Queries before the replacing projection still show the earlier measures. API and MCP keep sharing the same read path.
- A newer equivalent extraction that is not relevant, has no projected measures or failed does not replace anything. The same holds today within one version.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `equivalent-interpretation-carry-forward`: adds a requirement that current public measures follow the latest validated projection among equivalent versions of a document. The existing requirements are unchanged.

## Impact

- `internal/backend/publicquery/facts.go`: the replacement predicate of `measures()` joins `interpretation_input_equivalences` within the source cutoff. Same-version replacement is preserved as the special case where both sides are the same version.
- No migration. It reuses the fingerprints and the `(document_id, fingerprint, version_id)` index added by `carry-forward-equivalent-interpretations`.
- Unchanged:
  - tracked-target measure state, which is a conservative union of all versions under `tracked-target-interpretation-state`;
  - linking candidate retrieval;
  - measure detail reads by explicit identifier;
  - processing, configuration identities, publication gates and acceptance.
