# Retained CFR graphical interpretation

Tasks 27.4 and 32 add `cfr-graphics-v5` for retained criticality and vigilance products.
The ordinary projection boundary uses the matching retained HTML and required
PDF from a complete source version. It performs no upstream fetch or model call.
Source publication, development overrides and acceptance keep their existing gates.

## Recognized evidence

Criticality requires seven risk panels, two dates and 26 distinct labelled zones
per risk/day. Issuance must match HTML, including its time, and map dates must be
the issuance day and next day. Preceding adoption/scenario pages are allowed;
evidence cites the original physical page. Recognized fills yield green, yellow,
orange or red. White zone fills yield non-applicability only in main-river/coastal
masks. Whitespace, anonymous background rectangles and a no-criticality statement
cannot supply a level. Daily map colors are compared with the highest supported table level overlapping
the Europe/Rome calendar day. Every sequential table interval retains its precise
hours; conflicting overlapping intervals or inconsistent daily peaks are rejected.

Vigilance requires five complete labelled panels: two daily rainfall maps, the
separate cumulative map, and two other-phenomena maps. Its own eight-band legend
supplies literal rainfall bands in millimetres averaged over the area. The total
has its own literal period and does not replace daily validity. Document-labelled
symbols supply thunderstorms, wind, sea, snow and ice observations. Both
hydrological risk categories share the rainfall observation; all seven categories
retain vigilance warning level `not_applicable`.

Optional `weather` accompanies regional search and situation facts. `depicted`
means the band or symbol was interpreted, `not_depicted` means no recognized
symbol was present, and `unresolved` discloses unsupported evidence. None means
absence of risk. `under_evaluation` independently records HTML table membership.
The optional `total_graphical_status` assesses cumulative rainfall independently
from the daily `graphical_status`. An unrecognized cumulative fill does not erase
a supported daily band. An unrecognized mark or a listed phenomenon without a
matching labelled symbol remains unresolved. Source coverage retains aggregate limitations; supported individual
facts keep their own evidence assessment.

The interpreter supports reviewed linear vector geometry, compound subpaths,
affine path transformations and separately rendered irregular zone fills.
Only reviewed A6/island offsets permit a tightly bounded typographic association
with a clearly separated outline. Ambiguous labels, unsupported geometry,
changed legends, raster content inside maps and conflicting dates remain
unresolved or fall back to supported HTML evidence.

## Local rendering and history

The image packages Poppler and Liberation fonts. Each read has a temporary font
cache, a 45-second overall timeout, a 32-MiB PDF input limit, at most 12 pages and
a 16-MiB limit per command output. A failed SVG rendering is discarded. A bounded
local single-page vector-PDF render may normalize fonts before a second SVG
render; the locator records that method and the original physical page.

No database migration is needed: PDF URL, page, locator and optional weather
metadata use the retained evidence locator. New immutable projection records use
`cfr-graphics-v5` and their actual knowledge time. Repeated projection is
idempotent. Historical `cfr-vector-v3` and HTML records remain unchanged. A newer
unsupported version does not erase the preceding supported facts, and the query
reports that limitation.

## Verification — 4 October 2026

Assistant review compared retained maps/PDFs with the interpreter output:

| Sample | Scope | Result |
| --- | --- | --- |
| Criticality current | Seven risks, 26 zones, two days | 310 green and 54 excluded values; no unresolved levels |
| Criticality historical non-green | Seven risks, 26 zones, two days, leading adoption pages | 290 green, 20 yellow and 54 excluded values; original physical pages preserved |
| Vigilance with thunderstorms | Five panels and seven risk categories | 364 entries; rainfall bands and mapped thunderstorm zone/day membership agree |
| Vigilance without symbols | Five panels and seven risk categories | 364 entries; rainfall retained and other phenomena not depicted |

The historical non-green PDF uses a separately identified reviewed HTML
transcription of its table, not a newly fetched HTML counterpart. The retained
vigilance samples do not contain positive wind/sea/snow/ice observations;
redistributable synthetic fixtures exercise those symbols and their attribution.
Additional real positive cases belong to source acceptance. Private originals,
hashes, page renders and per-zone expected/result checks are retained under
`.local/operations/territori/09-toscana/graphical-274-20261004/`, record
`CFR-GRAPHICS-20261004-01`.

Repository tests use synthetic rectangular maps and invented symbols. Relevant
checks, from the repository root:

```sh
go test ./internal/backend/acquisition
go test -tags=integration ./internal/backend/publicquery -run TestFiveSharedPublicQueryGroups -count=1
go test -race ./...
go vet ./...
go build ./cmd/...
openspec validate define-toscana-alert-service --strict
```

The disposable PostgreSQL case verifies weather persistence, exact PDF/page
attribution, idempotence, historical knowledge, source interpretation and
missing-resource fallback. The transport case uses the official MCP client and
compares weather-bearing API/MCP output through pinned pagination. Shared schemas
describe the same optional metadata.

This completes the graphical software task in the recognized formats. Existing
development services were not replaced, and source controls were not changed.
Live image adoption/reprojection, campaign observation, municipal completeness
and source acceptance remain separate. The [coverage tracker](../coverage.md)
is the public source-status reference.


## Adozione successiva in development — 4 ottobre 2026

Il successivo [allineamento applicativo](toscana-readiness.md#adozione-in-development)
adotta la revisione pulita che include questo software nel percorso ordinario
development. Passano i controlli di immagine/configurazione, readiness e confronto
API/MCP della situazione sulla stessa vista. La verifica del formato resta quella
dei campioni conservati descritti sopra; l'adozione non certifica nuovi casi reali,
confronti mancanti del trial o accettazione dei prodotti CFR.

## Corrected processing policy — 7 October 2026

The intraday merge regression is covered by synthetic yellow–orange–yellow rows,
true overlapping conflicts and midnight boundaries. Summary prose preserves
supported elevated alert levels even when the projection is partial. Successful
ordinary acquisition records the verified issuance instant in Europe/Rome for
publication accounting, including unchanged versions.

An issuance within the configured tolerance before the expected instant satisfies
the current publication cycle; the early allowance is capped at half the interval
to prevent previous-cycle carryover. Actual issuance and cadence are preserved.
The current summary names the supported risk, color, zone and precise interval.
Primary comparison evidence accepts valid MIME parameters and selects the literal
product/issuance independently of the platform's missing identifiers.

The inspected vigilance edition contains fills absent from its own labelled
legend and an unlabelled wind mark. These are source-evidence limitations, rather
than a basis for a guessed palette or wind observation. Daily and cumulative
limitations are reported separately; no absence of risk is inferred. Revised
projections append knowledge under `cfr-graphics-v5`; prior records remain available
at their historical knowledge boundary. The bounded development adoption passes
eight actual SDK API/MCP comparisons, three pinned pagination pages and four
before/after projection and COC history boundaries. The exact yellow–orange–yellow
intervals and independent daily/cumulative weather metadata are verified. Source
acceptance remains separate, as documented in TOS-022.

## Ongoing alerts and derived green — `cfr-graphics-v7`

The task 31.5 reassessment (TOS-023) found that rows reading «in corso fino alle
ore …» were not parsed, so the whole criticality table was dropped. Only the
daily maps remained: on 8 October 2026 orange stayed current on zone A4 after
17:00, when the table had already switched to yellow. `cfr-graphics-v7` changes
three things.

- **Ongoing rows.** Such a row is an alert already in effect at issuance. Its
  interval starts at the edition's literal issuance and ends at the stated
  Europe/Rome instant. The validity keeps the original expression and an
  assumption stating that the earlier start belongs to preceding editions. A row
  is unsupported, and the table falls back as before, if its end is missing or
  ambiguous, not later than issuance, a repeated or nonexistent local time, or if
  the issuance is unreadable. Tables with ongoing rows now keep every row
  interval and are merged with the daily maps.
- **Derived green.** When a zone, risk and day has explicit alert intervals, the
  uncovered parts of the local day become green intervals. The assumption and
  evidence locator mark them as derived from the absence of a listed alert.
  There is no derivation when the map is not applicable or unresolved, when the
  table or maps are unsupported, or for gaps of one minute or less (the 23.59 day
  end).
- **Freshness in historical views.** Some public views are evaluated more than
  one view lifetime (30 minutes) before they are created. They report updating
  as of their evaluation time, so a past view no longer reads «delayed» just
  because it is served later. Current, recent and future views still refresh
  at every delivery. API and MCP stay identical.

**Environment failures do not persist a fallback.** The first adoption used
`cfr-graphics-v6` and reprojected editions 13961 and 26641 from the admin
container. Its read-only filesystem had no `/tmp`, so the PDF reader failed and
table-only fallback records were stored under v6. Since then:
- a PDF tooling failure (no temporary directory, a tool that cannot start)
  stores nothing and returns an error to retry;
- the admin service has a 64 MB `/tmp` tmpfs;
- the logic is `cfr-graphics-v7`, which supersedes those degraded records.

The 7 October `cfr-graphics-v4` projection of edition 12845, which held only
table intervals, fits the same failure. That cause is an inference and has not
been verified.

New records are appended under `cfr-graphics-v7`. v5 and earlier records stay
readable at their own knowledge time. Queries always serve each version's
latest projection, so restoring an older image does not hide v6 records; a
defect is corrected by a later logic version. Reprojection of retained editions
is explicit and bounded to selected versions with `project-regional-version`.
Change: `bound-ongoing-criticality-rows`.
