# Referenced acts

Municipal notices often announce measures "by ordinance" without attaching the
act, and their own dates can be wrong. From 23 October 2026 the service
retrieves acts that relevant notices cite explicitly and uses their
dispositive clause to clarify the citing notice's measures. Change:
`acquire-referenced-albo-acts` ([specification](../../openspec/specs/referenced-act-evidence/spec.md)).

## References and resolution

After a notice version projects measures, the service detects explicit act
references in the notice's own text, never in its attachments. A reference is
either numbered ("ordinanza n. 38/2026") or unnumbered ("apposita ordinanza",
"ordinanza comunale"). Detection is deterministic and calls no model; site
navigation such as the footer link to the albo does not count. At most three
references are recorded per version.

Each reference gets one appended outcome:

| Outcome | Meaning |
|---|---|
| `attached` | an attachment of the same notice states an act of the same kind (and number) |
| `numbered` | the albo has exactly one act of that kind, number and year |
| `unique_candidate` | for an unnumbered reference, exactly one albo act matches kind, an act date from seven days before the notice up to its publication date, and a term of the notice's measures in its object |
| `not_found` | no candidate, or no albo configured for the municipality |
| `ambiguous` | several candidates; they are listed and nothing is chosen |
| `unavailable` | the albo could not be read completely within the page bound; retried automatically |

## The albo source

An albo is a separately registered source (access method `jcitygov-albo`) that
serves one primary source. Its policy permits collection and retention and
forbids copies. The service never lists it exhaustively: it reads only the
bounded search pages a reference needs, then the record detail and the act's
unsigned main document. These are retained once per albo record as an
immutable version. Albo versions run OCR only. They are never classified,
extracted, projected, scheduled for listing checks, counted in coverage or
readable as public documents; a preview reads only the first list page.

## Clarification

A reviewed literal parser reads the act's clause after `ORDINA`: start date and
time, an end instant or an ending condition such as "fino al perdurare
dell'emergenza", and the listed dispositions. Recitals that repeat the
dispositions are ignored. A notice measure matches a disposition when kinds
agree and the head term of its subject occurs in that disposition.

- An unknown validity becomes an interval or a conditional validity with a
  known start, in Europe/Rome. The new validity and interpretation event are
  appended; the act is added to the measure's evidence by official link,
  locator `referenced_act:<kind> <number>/<year>` and passage.
- A determinate validity is never changed. If it differs from the act, a
  conflict naming both passages is recorded.
- A measure with no matching disposition, or an act whose clause is not
  recognized, is recorded as such and left unchanged.

Queries before a clarification return the earlier validity and evidence. API
and MCP share the read path. Operators can list references, outcomes,
candidates and clarifications read-only on the source page ("Atti citati") or
with `iwa source-act-references <source>`.

## Verification

The development adoption on 23 October 2026 registered `calcinaia-albo` for
`calcinaia-municipal` and reprocessed three Calcinaia notices:
- the 9 September notice resolved to Ordinanza 38/2026, and its five measures
  became conditional from 10 September 00:00;
- the 20 August notice resolved to its attached Ordinanza 34/2026, clarifying
  three measures;
- the 21 August monitoring stayed `ambiguous` between three ordinances, with
  its validity unknown.

API and MCP agree, and earlier boundaries are unchanged
([CLN-042](../territori/09-toscana/comuni/050004-calcinaia.md#cln-042--atti-citati-recuperati-e-validità-chiarita)).
The JCityGov portal rejects the default Go client, so albo requests identify
the service in their User-Agent. Notices projected before an albo is enabled
are resolved only when they are projected again. Detailed evidence is kept
private (`referenced-acts-20261023`).

## Rollback

Restore the previous images and disable collection for the albo source. This
stops new resolutions and clarifications. Clarifications already recorded are
facts recorded at their time: older readers still return the clarified
validity, without the act evidence item. Removing them would rewrite history,
which the specification forbids.
