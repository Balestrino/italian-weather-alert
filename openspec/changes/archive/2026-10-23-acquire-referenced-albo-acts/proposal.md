# Proposal

## Why

Municipal notices often announce measures "by ordinance" without attaching the act, and their own dates can be wrong.

The 31.4 inventory of 23 October 2026 found this case in the 9 September Calcinaia criticality notice:
- the headline says the closures start "dalle 0.00 di giovedì 10 settembre";
- the body says "giovedì 10 agosto";
- the expiry is "10 agosto".

The five public measures therefore keep validity `unknown`. The cited act, Ordinanza 38/2026, is on the municipal albo pretorio and states the dates unambiguously: "a partire dalle ore 00.00 del giorno 10 settembre 2026, fino al perdurare dell'emergenza".

Today the service cannot retrieve referenced acts. It also does not use acts attached to notices to clarify dates. This leaves determinable validity unknown. The ingestion requirement "A municipal notice references an albo act" is unmet.

## What Changes

- **Detect explicit act references** in the retained content of relevant primary notices that have projected measures. References can be numbered ("ordinanza n. 38/2026") or unnumbered ("apposita ordinanza", "ordinanza comunale"). Each reference is recorded with its literal evidence.
- **Resolve references automatically**, in this order:
  - an act attached to the notice itself;
  - a numbered act found in the albo by category, number and year;
  - an unnumbered act found in the albo only when exactly one candidate matches category, act date window and subject.

  No candidate, several candidates or an unreachable albo are recorded as `not_found`, `ambiguous` or `unavailable`. They are never guessed.
- **Acquire resolved albo acts** through a new, separately registered albo source per municipality (first: Calcinaia, JCityGov "papca"). The albo record detail and its unsigned PDF are retained as an immutable document version. The albo is never collected exhaustively and is never an alert channel. Public responses expose only the official link, never a copy.
- **Clarify notice measures from the act.** A reviewed literal parser reads the act's dispositive clause: the start, the end or condition, and the listed dispositions. A notice measure whose validity is unknown or conflicting, and which matches a listed disposition, gets a new validity value. That value carries the act as evidence and an interpretation event. The act creates no measures of its own. A notice validity that is determinate and different is kept, and the disagreement is recorded as a conflict.
- **Make it visible.** Measure evidence names the act and the locator of its passage. Administration and the CLI list each notice's references, resolutions and clarifications. Earlier knowledge boundaries are unchanged.

## Capabilities

### New Capabilities

- `referenced-act-evidence`: detection, automatic resolution, bounded acquisition of acts referenced by municipal notices, and clarification of the citing notice's measures, with history, public-link and failure rules.

### Modified Capabilities

None. The ingestion requirement for referenced albo acts belongs to the active `define-toscana-alert-service` change; this change implements it for Calcinaia without changing its text.

## Impact

- **Registry:** a JCityGov albo access method with configuration validation, and a `calcinaia-albo` source registered in development with its reviewed policy.
- **Acquisition:** a JCityGov client for registry-number and paged-list search, detail parsing and attachment download. Requests are bounded per notice and recorded with retries.
- **Processing:** new jobs for reference resolution, act acquisition (followed by the existing OCR) and measure clarification. Albo documents are never classified or extracted.
- **Domain:** an append-only migration for references, resolutions and clarifications. Validity values and interpretation events are only appended.
- **Public query, API and MCP:** an act evidence item on clarified measures, with the official link and passage; same read path and history rules.
- **Administration and CLI:** a read-only list of references per source and notice.
- **Development data:** the Calcinaia 9 September notice (doc 60) and 21 August monitoring (doc 213), and the 20 August notice with the attached Ordinanza 34/2026 (doc 59), are the acceptance cases.
- **Out of scope:** other municipalities' albi; linking acts to measures of other notices; generic albo search for other purposes; changes to relevance or extraction prompts.
