## Purpose

Represent regional warnings and individual municipal measures with traceable geography, time, evidence and uncertainty, including public history.

## ADDED Requirements

### Requirement: Seven risks and three regional products
The service SHALL cover minor-network hydrogeological/hydraulic risk, main-network hydraulic risk, strong thunderstorms, wind, coastal waves, snow and ice, preserving official labels. It SHALL distinguish regional meteorological vigilance, criticality/alert and monitoring products. Monitoring SHALL NOT automatically assign a new warning color. Green, yellow, orange, red, unknown and not applicable SHALL remain distinct. Fire, heat-health and avalanche products SHALL remain outside first-release coverage.

#### Scenario: A monitoring document reports changed conditions
- **WHEN** a monitoring product describes an evolving event without explicitly issuing a new alert level
- **THEN** it is exposed as monitoring and does not automatically overwrite the criticality bulletin's color

### Requirement: Origin and municipal authority
The service SHALL distinguish originating authority, publisher, platform and document kind. A regional warning republished by a municipality SHALL remain regional. Added local measures SHALL remain separately identifiable and linked to supporting documents. Documented local operational phases SHALL be separate from regional color; a stronger local response SHALL NOT itself be marked as a discrepancy. Local meteo-hydrological observations or measures SHALL be included without requiring a regional bulletin and without inventing a regional color.

#### Scenario: A municipal page republishes a warning and closes parks
- **WHEN** one page contains a regional alert and a municipal closure
- **THEN** the warning retains its regional origin and the closure is a distinct local measure attributable to the municipality

#### Scenario: A local closure has no regional bulletin link
- **WHEN** an official municipal notice reports a flooded road closure without a linked regional warning
- **THEN** the local notice is included without an invented color or a claim that no regional warning exists

#### Scenario: A municipal ordinance is published by an associated service
- **WHEN** a recognized associated-service channel republishes an identifiable municipal ordinance
- **THEN** the municipality remains the issuing authority, the associated service is recorded as publisher, and the platform does not acquire authority over the measure by hosting it

### Requirement: Versioned municipality and zone applicability
The service SHALL identify municipalities by ISTAT code and use official, attributed and versioned municipality-zone mappings. It SHALL support multiple zones per municipality with separate risks, levels and validity. Any optional maximum-level summary SHALL disclose when it applies only to part of the municipality. Unresolved mapping SHALL be explicit rather than resolved by unsupported geographic assumptions.

#### Scenario: A municipality spans two zones
- **WHEN** only one of its applicable zones has an elevated level
- **THEN** both zones are listed separately and any maximum summary identifies its partial territorial applicability

### Requirement: Traceable normative and geographic applicability
Normative rules and municipality-zone mappings used by the service SHALL identify supporting official acts or datasets, versions, relevant provisions, verification dates and applicable periods where established. Publication, entry into force and operational applicability SHALL remain distinct. Unresolved effective dates or later-amendment checks SHALL be scoped to affected rules or mappings rather than silently asserting current applicability. Updated evidence SHALL preserve provenance of earlier versions and identify affected rules, mappings and interpretations for evaluation before controlled adoption.

#### Scenario: Effective date is not established
- **WHEN** an acquired act supports a mapping but its applicable start date remains unresolved
- **THEN** the service preserves the act and mapping version, discloses the temporal limitation and does not use acquisition or publication time as an invented effective date

#### Scenario: A later act changes a territorial rule
- **WHEN** verified evidence establishes a new mapping or applicability rule
- **THEN** the earlier version remains attributable and the affected scope is evaluated before adopting the new version without silently rewriting historical applicability

### Requirement: Individual measures and scoped updates
Each structured measure SHALL identify its supported action, affected place or scope, evidence and temporal uncertainty. The service SHALL preserve original documents and versions separately from extracted measures. Corrections, extensions, cancellations and reopenings SHALL apply only to the measures supported by the evidence. Automatic links without an explicit source hyperlink SHALL require supporting evidence from the new and preceding documents; candidate similarity alone SHALL NOT establish a relationship. Ambiguous relationships SHALL remain unverified. In-page updates and exceptions SHALL be interpreted in context rather than treated as simultaneous unqualified current statements. Disappearance alone SHALL NOT imply cancellation.

#### Scenario: Reopening leaves other restrictions in force
- **WHEN** a notice announces the reopening of the via Maremmana underpass and confirms other restrictions
- **THEN** the reopening affects the underpass only, the other restrictions remain distinct, and each conclusion identifies its evidence

#### Scenario: A general update includes an exception
- **WHEN** an update reports roads cleared but explicitly states that one underpass remains closed
- **THEN** the service does not infer that the underpass has reopened from the general statement

### Requirement: Separate temporal meanings
The service SHALL retain original time expressions and distinguish publication, source modification, event time, operational validity, page expiry, platform dates and acquisition time. Each temporal candidate SHALL retain its meaning and evidence independently of the resolved field. Determinable instants SHALL be represented in UTC; date-only values SHALL remain dates. Europe/Rome SHALL be assumed only when justified by the source, with the assumption recorded. Unknown precision, timezone ambiguity and conditional validity SHALL remain explicit. It SHALL distinguish current, future already-issued, expired, cancelled, superseded and undetermined information. Publication time SHALL NOT be substituted for an unspecified event time; page or platform dates SHALL NOT automatically determine operational validity.

When incompatible supported expressions claim the same temporal meaning for one fact, the service SHALL retain every expression and its evidence under a shared conflict identity. The resolved temporal field SHALL remain undetermined until supported clarifying evidence is processed; ordering, page position or model preference SHALL NOT silently select one candidate.

#### Scenario: A notice reports an earlier reopening
- **WHEN** a page updated at 09:36 says the underpass has reopened without specifying when
- **THEN** the service reports the reopening as attested by that communication and does not assert 09:36 as the reopening time

#### Scenario: Platform dates conflict with operational text
- **WHEN** an API lists the same start/end date but the source text covers a later day
- **THEN** the distinct dates and their source meanings are preserved without an invented measure expiry

#### Scenario: Summary and dispositive starts conflict
- **WHEN** a municipal summary says a measure starts on 10 September while the dispositive text says 10 August
- **THEN** both original expressions and supporting passages share a recorded conflict, and operational start remains undetermined rather than selecting either date

### Requirement: Undetermined validity and inaccessible updates
A measure without explicit expiry SHALL retain its publication date and disclose that no end is specified and no revocation has been acquired when that is true. It SHALL NOT be presented as certainly still in force. When updates cannot be checked, the last acquired provision SHALL remain available with that limitation and its source and acquisition information. Neither source failure nor record age SHALL establish expiry or continued validity.

#### Scenario: A closure lasts until further notice and the channel fails
- **WHEN** no revocation has been acquired and the channel becomes inaccessible
- **THEN** the last acquired provision remains visible with updates unverified and its present validity undetermined

### Requirement: Three independent evidence dimensions
Responses SHALL expose provenance, interpretation and updating as separate dimensions with supporting evidence, limitations and relevant dates. Updating SHALL distinguish reachability, successful content reading and complete checking of the declared scope. The service SHALL NOT substitute a single confidence score for these dimensions. Statements about present conditions SHALL be bounded by the latest documented information and checks, not a guarantee of real-world completeness.

#### Scenario: A known official source has ambiguous content
- **WHEN** provenance is verified but a measure's end time cannot be interpreted
- **THEN** provenance remains verified, interpretation identifies the unresolved field, and updating reports its independently observed status

### Requirement: Coverage and comparable discrepancies
The service SHALL separately report regional and local coverage. Regional products SHALL remain available when local channels are unverified or unavailable, with local limitations and the last successful acquisition if known. Empty results SHALL NOT mean no local measures or safety. Discrepancies SHALL be evaluated only after comparing risk, territory, issuance and validity; both conflicting publications SHALL remain attributable, with the regional bulletin defining its own technical level.

#### Scenario: Local coverage is unavailable
- **WHEN** regional applicability is known but the municipal channel cannot be verified
- **THEN** regional information is returned alongside an explicit local-coverage limitation rather than a claim of no municipal provisions

### Requirement: Configurable history and ongoing-measure exceptions
Public history SHALL disclose its available bounds and gaps, including the initial 30-day recovery and explicitly referenced older documents. Service-knowledge history SHALL begin at actual acquisition/interpretation, not at the source publication date. Structured records, originals and versions SHALL have a configurable retention duration initially three months from each version's first acquisition. Unchanged checks SHALL NOT reset that clock. No complete pre-startup archive SHALL be promised. Measures still valid or without established cessation, together with the documents and versions necessary to reconstruct them, SHALL be retained beyond that period. Age alone SHALL NOT establish validity. Retention SHALL respect source conditions and disclose resulting evidence limitations. History SHALL distinguish what an authority stated from when the service acquired or interpreted that information.

#### Scenario: An unresolved closure exceeds three months
- **WHEN** a retained closure has no established cessation and reaches the normal retention limit
- **THEN** the measure and necessary supporting evidence are preserved while its unresolved present validity remains explicit

#### Scenario: An old document is acquired late
- **WHEN** a past publication is first collected today
- **THEN** history preserves its original date without implying the service knew its contents before today's acquisition

#### Scenario: The operator changes the retention duration
- **WHEN** the duration is increased or reduced
- **THEN** the next cleanup applies it to already retained data, preserves protected evidence, and does not imply that previously deleted data can be recovered

#### Scenario: Protection ends after ordinary retention has elapsed
- **WHEN** a measure's cessation is established and its evidence no longer supports any other ongoing or unresolved measure
- **THEN** versions older than the configured duration become eligible for cleanup

### Requirement: Optional semantic candidate retrieval
Update linking SHALL work with structured and textual candidate retrieval when semantic retrieval is disabled. Optional semantic retrieval SHALL add candidates without replacing evidence validation or overriding geographic/measure applicability. Candidate selection SHALL include older ongoing/unresolved measures where relevant rather than being limited to the 30-day acquisition window. Evaluation SHALL compare linking results and consumption with and without semantic retrieval on the same cases.

#### Scenario: Semantic retrieval is disabled
- **WHEN** a reopening notice has no explicit link to its earlier closure
- **THEN** structured/textual retrieval still supplies relevant prior candidates and any relationship requires document evidence

#### Scenario: A semantically similar notice concerns another place
- **WHEN** semantic retrieval returns that candidate
- **THEN** similarity alone does not cancel or update the unrelated measure
