# Spec Delta

## Purpose

Retrieve the official acts that municipal notices explicitly cite, and use their literal dispositive text to clarify the citing notice's measures. The service never treats an albo as an alert channel, never guesses an act's identity and never rewrites earlier knowledge.

## ADDED Requirements

### Requirement: Detection of explicit act references
For a primary municipal notice version whose latest validated interpretation is relevant and has projected measures, the service SHALL detect explicit references to acts in its retained interpreted content.

A reference SHALL name an act kind, such as ordinanza, decreto, delibera or determina:
- with a number and year, a **numbered** reference;
- or without them, an **unnumbered** reference, for example "apposita ordinanza" or "ordinanza comunale".

Each reference SHALL be recorded once per notice version with its kind, any number, year and date, its literal passage and its locator. Site navigation, such as a footer link to the albo, SHALL NOT count as a reference. Detection SHALL be deterministic and SHALL NOT call a model.

#### Scenario: Unnumbered ordinance in a criticality notice
- **WHEN** a relevant notice with projected measures states that "il Sindaco ha predisposto mediante apposita ordinanza" closures and prohibitions
- **THEN** one unnumbered ordinance reference is recorded with that passage

#### Scenario: Navigation link only
- **WHEN** a notice contains only the site footer link "Albo pretorio"
- **THEN** no reference is recorded

#### Scenario: Notice not relevant or without measures
- **WHEN** a notice version is not relevant, or has no projected measure
- **THEN** no reference detection or albo request is performed for it

### Requirement: Automatic resolution of references
The service SHALL resolve each reference in this order, recording exactly one outcome with its method and evidence:
1. `attached`: a retained attachment of the same notice version is the cited act. Its retained text states an act of the same kind and, for a numbered reference, the same number and year.
2. `numbered`: the configured albo of the notice's municipality returns exactly one record of the reference's kind with that number and year.
3. `unique_candidate`, for unnumbered references only: exactly one albo record matches all of these:
   - the reference's kind;
   - an act date from seven days before the notice's source publication date up to that date;
   - an object that contains at least one disposition term of the notice's projected measures.

When no record matches, the outcome SHALL be `not_found`. When more than one matches, it SHALL be `ambiguous`, listing the candidates. When the albo cannot be read completely for the bounded search, it SHALL be `unavailable` with the failure reason. A municipality without a configured albo SHALL yield `not_found` for references that are not attached.

The service SHALL NOT choose among several candidates. It SHALL NOT infer an act from publication dates alone. Outcomes SHALL be appended, never rewritten. An `unavailable` outcome MAY be retried with bounded backoff; a later outcome supersedes it from its recording time.

#### Scenario: Unnumbered reference with one matching ordinance
- **WHEN** the 9 September 2026 notice cites "apposita ordinanza" for closures of cemeteries and parks
- **AND** the albo has exactly one ordinance dated 9 September 2026 whose object mentions parks and cemeteries
- **THEN** the reference resolves `unique_candidate` to that ordinance

#### Scenario: Ordinance and decree with the same object
- **WHEN** the albo holds an ordinance and a decree with the same object and date, and the notice cites an ordinance
- **THEN** only the ordinance is a candidate and the reference resolves to it

#### Scenario: Two matching ordinances
- **WHEN** two ordinances in the window both match the notice's disposition terms
- **THEN** the outcome is `ambiguous` with both candidates and no act is acquired

#### Scenario: Act attached to the notice
- **WHEN** a notice that cites an ordinance retains that ordinance as a PDF attachment
- **THEN** the reference resolves `attached` to the retained resource without any albo request

#### Scenario: Albo unreachable
- **WHEN** the albo search fails
- **THEN** the outcome is `unavailable` with its reason, the notice's measures are unchanged, and a later retry may resolve it

### Requirement: Bounded acquisition of referenced albo acts
Albo acts SHALL be acquired only for resolved `numbered` or `unique_candidate` references, through a separately registered albo source of the same municipality. That source's reviewed policy SHALL permit collection and retention. Each acquisition SHALL retain the albo record's detail and the act's main document as an immutable document version with hashes, attributed to the issuing municipality.

The service SHALL NOT traverse, list or retain other albo records except the bounded search pages needed to resolve a reference. Bounded means at most a configured number of pages per reference. The albo SHALL NOT be an alert channel: its records SHALL NOT be classified, extracted, projected as measures or counted in coverage. Public responses SHALL expose the act only through its official albo link and passage, never through a retained copy. Acquisition failures SHALL be recorded and retried within the same bounds, without changing the notice's measures.

#### Scenario: Resolved ordinance is retained once
- **WHEN** two notices resolve to the same albo ordinance
- **THEN** the ordinance is retained once and both references point to the same act version

#### Scenario: Albo records are not interpreted
- **WHEN** an albo act version is retained
- **THEN** no classification, extraction or measure projection runs for it and source coverage is unchanged

#### Scenario: Public copy denied
- **WHEN** a client requests the retained copy of an albo act
- **THEN** the service returns the official link only and refuses the copy

### Requirement: Clarification of notice measures from the cited act
For a resolved act, the service SHALL read its dispositive clause with a reviewed literal parser:
- the start, including date and time;
- the end, or the condition that ends it;
- each listed disposition with its kind.

A measure projected from the citing notice version SHALL be clarified only when its current validity is `unknown`, or is unresolved because of conflicting candidates, and when its kind and its subject's head term match one listed disposition.

A clarification SHALL append a validity value derived only from the act's literal passages, in Europe/Rome:
- a determinate start with a stated end SHALL give an interval;
- a determinate start with a conditional end, such as "fino al perdurare dell'emergenza", SHALL give a conditional validity whose start is known and whose condition is the literal text.

The clarification SHALL also append an interpretation event naming the act. The act SHALL be added as evidence of the measure with its locator and passage.

The service SHALL NOT clarify a measure whose own validity is determinate. When that validity disagrees with the act, the service SHALL record a conflict and SHALL NOT change the measure. When the act's clause cannot be parsed, the service SHALL record `clarification_unavailable` and leave the measure unchanged. The act SHALL NOT create measures, and SHALL NOT end, extend or relate measures of other notices.

#### Scenario: Conflicting notice dates clarified by the ordinance
- **WHEN** a notice's measure has validity `unknown` because the notice states both "10 settembre" and "10 agosto"
- **AND** the resolved ordinance orders the same closure "a partire dalle ore 00.00 del giorno 10 settembre 2026, fino al perdurare dell'emergenza"
- **THEN** the measure's current validity is conditional, starting at 2026-09-10T00:00 Europe/Rome with condition "fino al perdurare dell'emergenza"
- **AND** its evidence includes the ordinance's official link and passage

#### Scenario: Determinate notice validity differs from the act
- **WHEN** a measure has a determinate interval and the act states a different start
- **THEN** the measure is unchanged and a conflict naming both passages is recorded

#### Scenario: Disposition not in the act
- **WHEN** a notice measure's subject matches no disposition listed by the act
- **THEN** that measure is not clarified

#### Scenario: Unparseable act
- **WHEN** the act's dispositive clause does not match the reviewed literal patterns
- **THEN** `clarification_unavailable` is recorded and no validity is appended

### Requirement: History, interfaces and visibility of referenced acts
References, resolutions, acquisitions, clarifications and conflicts SHALL be append-only and timestamped. Public reads at a knowledge boundary SHALL use only those recorded within the source interpretation cutoff, so queries before a clarification return the earlier validity and evidence.

Equivalent API and MCP operations SHALL present the same clarified validity, evidence and limitations. Administration and the command-line interface SHALL list, per source and notice version, each reference with its literal passage, outcome, method, candidates, acquired act and clarified measures, without allowing edits. No retained notice, act, measure or earlier validity SHALL be rewritten or deleted by this capability.

#### Scenario: Query before the clarification
- **WHEN** a measure is read at a knowledge boundary before its clarification was recorded
- **THEN** it shows the earlier validity and evidence without the act

#### Scenario: Interfaces agree
- **WHEN** a clarified measure is read through the HTTP API and MCP at the same saved view
- **THEN** both return the identical validity, evidence and limitations

#### Scenario: Operator inspects references
- **WHEN** an operator lists references for the Calcinaia source
- **THEN** each notice's references show outcome, method, candidates and clarified measures, read-only
