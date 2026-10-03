# Manual municipality publication in development

In the development backoffice, open **Regioni → Toscana → comune → Configurazione → Pubblicazione in development**. Enter an operator name and select **Abilita pubblicazione in development**. The same panel reports its revision and provides **Revoca pubblicazione in development**. Only selected municipalities gain access to their retained municipal data and applicable regional products through API/MCP. A municipality without municipal facts still has no invented local measures; uncertain regional levels remain uncertain.

This is separate from the controls enabling collection. Source acceptance and global source-publication flags are not changed. The [coverage tracker](../coverage.md) continues to describe acceptance; development publication does not close the observational trial, human reviews, failed source regressions or production release gates. Publication still requires source provenance, permitted policy without unresolved conditions and retained acquisitions under the active configuration. DPC comparisons remain internal. Regional document metadata can refer to a shared bulletin covering the whole region; regional fact searches are limited to zones of selected municipalities using the mapping available at the query knowledge boundary. Retained copies keep their independent strict authorization.

For scripts, use the development admin image after migrating:

```sh
scripts/compose-env.sh development run --rm --no-deps admin municipality-development-publication-status 09 050004
scripts/compose-env.sh development run --rm --no-deps admin municipality-development-publication-enable 09 050004 0 operator
scripts/compose-env.sh development run --rm --no-deps admin municipality-development-publication-disable 09 050004 1 operator
```

Use the revision returned by the status command rather than assuming the example values. A stale revision fails without overwriting the current choice. Operator and time are preserved in the municipality's **Storico**. Revoking or changing relevant publication settings expires saved API/MCP views; clients restart the query when `cursor_expired` is returned.

Municipality and source coverage responses include `development_publication`. Interpretation and response limitations identify the manual development override. `public_state`, `coverage_status` and `local_coverage` continue to describe source acceptance/publication; `pending` can therefore accompany available development data.

`IWA_ENVIRONMENT` accepts `development`, `staging`, or `production`; the software default is `production`. Compose pins the selected value for public, admin, worker and backup. Staging and production honor only the ordinary source acceptance/publication path and expose no manual development form/route. CLI commands reject those environments. Copying development database rows does not enable production visibility, and saved development views expire under a strict runtime. Migration `067_development_publication` is additive and enables no municipality automatically.

Verification covers synthetic municipal/regional facts, same-zone unselected municipalities, document access, independent copy restrictions, truthful pending states, unchanged source acceptance, revision conflicts, immutable audit, CSRF rejection, API/MCP cursor revocation and strict environment defaults. Real source acceptance requires the separately specified reviews and observational evidence.

On 2026-10-02 the targeted development public/admin rollout and native browser form
were verified with Calcinaia selected and another A4 municipality unselected. Live
API/MCP agree on the same saved view; source acceptance/public flags are unchanged.
The nondisruptive HTTP/boundary/secret smoke passed. Existing worker and backup
containers were preserved; the full runtime checker therefore still reports older
application images and this check does not establish a uniform release rollout.
Detailed operational settings, database snapshot and rollback evidence remain private.

For rollback to software predating publication-scope checks, first stop the updated
public listener and invalidate saved development views before starting the old image;
older software cannot enforce the added view scope. Preserve source evidence and the
additive migration. Restore the individually recorded listener images/configuration,
without rebuilding or restarting workers. Ordinary revocation on current software
requires only the revision-checked municipal control above.


On 2026-10-03 the primary municipal extraction path was connected to the domain
and verified in development. Completed ordinary extractions retain action,
subject, field evidence, literal validity and projection knowledge time; a COC
activation does not imply a named phase. Supported linking results are applied
when both measure bindings exist. Evaluation runs, platform-only candidates,
incomplete resources and suspended sources are excluded. Repeat projection is
idempotent. A bounded retained-result recovery makes no model calls, excludes
reviewed non-weather false positives and preserves source controls.

Coverage reports interpretation progress separately from source acceptance.
Pending first notices appear in the situation even without previous measures;
configured section listings and pagination are excluded. Existing official
provenance evidence is recorded at the actual assessment time without replacing
explicit assessments. Source acceptance stays pending. Date conflicts and
unverified CFR graphical levels remain unresolved.

The corrected public/admin listeners and six development workers were verified,
including actual HTTPS API/MCP equivalence, pagination, past knowledge and an
unselected municipality. Worker startup preserves prior processing catalogs when
several fallback revisions share a name; new local registrations are bound to the
exact fallback. A failed initial worker rollout was reverted before this fix and
the final workers started successfully. Backup and infrastructure were preserved.
Exact images, controls, excluded results and rollback steps remain private.


The subsequent 3 October criticality correction reads retained PDF vector fills,
labelled zones and risk/date panels using packaged Poppler tools. It matches the
PDF/HTML issuance, cites the PDF URL/page and preserves precise explicit alert
intervals. Unknown colors, missing or ambiguous polygons and unsupported layouts
retain uncertainty or supported HTML fallback. White outlined polygons in the
main-river/coastal masks are distinct from green; page background is never a fact.
The new parser keeps older projections at their original knowledge boundaries.

A bounded development projection and the actual Calcinaia HTTPS API/MCP verify all
seven A4 risks for both days as supported, without regional unknown levels in
that response. This replaces the unresolved A4 map-level result recorded above.
Source-wide limits for other unsupported labels and vigilance graphics remain
separate. The current development public/admin and six workers run the correction;
source controls, backup and infrastructure are preserved. Municipal date conflicts,
unestablished expiry and formal source acceptance are not changed by the regional
parser. The original dated observations above remain historical evidence.
