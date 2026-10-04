# Retained CFR graphical interpretation

Task 27.4 adds `cfr-graphics-v4` for retained criticality and vigilance products.
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
cannot supply a level. Explicit alert table intervals retain their precise hours.

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
An unrecognized mark or a listed phenomenon without a matching symbol remains
unresolved. Source coverage retains aggregate limitations; supported individual
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
`cfr-graphics-v4` and their actual knowledge time. Repeated projection is
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
