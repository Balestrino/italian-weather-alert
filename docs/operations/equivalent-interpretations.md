# Equivalent interpretation carry-forward

Municipal sites often change page templates without changing a notice. Each
change creates a new retained version. After a processing configuration
revision, that version is reinterpreted instead of reusing the earlier result.
From 22 October 2026, if the new interpretation fails or is still pending,
public document reads present the most recent earlier validated interpretation
of an equivalent input of the same document. Change:
`carry-forward-equivalent-interpretations` ([specification](../../openspec/specs/equivalent-interpretation-carry-forward/spec.md)).

## Equivalence

Each reusable interpretation input manifest gets an
`interpretation-equivalence-v1` fingerprint (migration
`100_interpretation_input_equivalences`). The fingerprint ignores:
- the processing configuration;
- the raw bytes of HTML, text and JSON resources, which count only through their interpreted text.

It covers interpreted text sections, metadata, issuer, completeness, resource
roles, media types, required and missing states, eligibility and policies. OCR
resources also keep their retained identity and page input/output hashes.
Incomplete or non-reusable inputs have no fingerprint and are never equivalent.

New classification and extraction runs write the fingerprint together with the
manifest. Manifests stored before the migration need a one-off backfill, which
reads only PostgreSQL rows:

```sh
scripts/compose-env.sh development run --rm --no-deps --pull never admin interpretation-equivalence-backfill <after-run> <limit>
```

Repeat with the reported `next_after_run` until `complete` is true. A rerun
writes nothing.

## Public presentation

Carry-forward applies only when a version's own interpretation at the knowledge
boundary is `failed` or `not_processed`. Only an earlier `supported` or
`partial` interpretation is carried forward, never an `unreliable` one, and
only within the same document. The response keeps the version's own identity,
hash and acquisition time, and adds:

- `carried_forward_from_equivalent_version:<version>:run:<run>:configuration:<configuration>` as the first interpretation limitation;
- `own_interpretation:failed:<code>` or `own_interpretation:not_available`;
- an evidence item with locator `equivalent_version_interpretation` pointing to the originating version.

A version's own later result, including a not-relevant one, takes precedence
from its recording time. Historical queries use only the versions, manifest runs
and results known within the source interpretation cutoff. No interpretation,
receipt, measure or projection is written. Domain measures were already read
from every retained version and are unchanged. API and MCP share the same read
path. Each batch of document reads adds two round trips, however many versions
it contains.

## Verification

The development adoption on 22 October 2026 shows:
- fingerprints match the reusable manifests;
- the three Calcinaia notices of
  [CLN-037](../territori/09-toscana/comuni/050004-calcinaia.md#cln-037--interpretazioni-equivalenti-portate-avanti-in-development)
  present the carried-forward interpretation;
- API and MCP return the same result for the same saved view;
- boundaries before each new version show the earlier version with its own result.

Rollback restores the previous public, admin and worker images. The additive
table can stay. Detailed evidence is kept private (`carry-forward-20261022`).
