## Purpose

Represent regional warnings and individual municipal measures with traceable geography, time, evidence and uncertainty, including public history.

## ADDED Requirements

### Requirement: Precise intervals and daily criticality peaks
The service SHALL preserve every supported explicit criticality interval when a daily map reports the day's highest level. It SHALL compare that peak with all applicable intervals, reject contradictory overlapping intervals and genuine map/table inconsistencies, and SHALL NOT extend an afternoon warning to midnight. A corrected projection SHALL append a new logic version and preserve earlier knowledge.

#### Scenario: Yellow orange and yellow occur in succession
- **WHEN** the same zone/risk has successive yellow, orange and yellow intervals and its daily map reports orange
- **THEN** all three precise intervals remain available and the daily peak does not cause a false conflict or discard unrelated map facts

### Requirement: Explicit municipal activation time
The service SHALL distinguish start linkages from end expressions for a civil-protection activation. A yearless operative date MAY use attributable matching publication-year context only when calendar and weekday checks agree; original phrases and the assumption SHALL remain inspectable. An unstated end SHALL remain unknown.

#### Scenario: COC opens at the start of the alert
- **WHEN** the operative statement explicitly opens the COC at a named day/month and local time in conjunction with the beginning of an alert
- **THEN** that phrase supports the activation start, never its end, and the regional alert's closing time is not transferred to the COC

### Requirement: Independent places in retained prohibitions
An explicitly enumerated positive operative prohibition SHALL retain each independently stated place with literal field evidence, including when a provider groups the list into one candidate. Conjunctions within a place SHALL NOT create unsupported separate scopes. Local completion SHALL require a recognized retained operative clause owned by the extraction core and SHALL preserve temporal uncertainty and prior interpretation versions. Negative, conditional, quoted, heading-only, context-only and ambiguous exception lists SHALL NOT supply inferred prohibitions.

#### Scenario: A partial reopening retains four prohibition places
- **WHEN** a retained operative clause separately lists four places under one prohibition while reopening only one underpass
- **THEN** the validated interpretation preserves four independently evidenced prohibitions alongside the reopening, closures and activation, without transferring reopening or page-update times

#### Scenario: Reprocessing corrects a grouped measure
- **WHEN** a newer complete ordinary extraction of the same retained version is projected with corrected independent subjects or places
- **THEN** current queries use its measure set at the new knowledge boundary while previous interpretations remain visible in earlier historical views; unsuccessful, unprojected and evaluation results do not hide supported facts

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

#### Scenario: Partial reopening and typographic place spelling
- **WHEN** a completed circulation reopening names one subject while other closures and place-specific prohibitions remain, and the model uses a straight apostrophe for a retained typographic apostrophe
- **THEN** contiguous canonical matching retains the original place spelling before scope validation and merging, preserving each supported restriction independently
- **AND** a recognized standalone, core-owned positive reopening sentence supplies only that subject's reopening, without transferring its earlier closure time or the page update time; negated, conditional, prospective, quoted and heading-only statements do not supply a reopening

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

Cittadino Informato candidates SHALL retain channel-specific evidence and persistent multi-source verification receipts for Regione Toscana/CFR, municipality and platform as defined in official-source-ingestion. Receipts SHALL distinguish corroboration, missing primary evidence, unavailable checks, non-comparable versions/validities, checks not applicable to the claim and actual conflicts. Corroboration SHALL be field-scoped and SHALL NOT transfer dates, validity, authority or completeness by inference. Local-only claims MAY record the CFR role as not applicable; regional claims SHALL retain the originating CFR level without requiring a municipal republication or using majority vote. A primary check failure or an unmatched candidate SHALL remain distinct from an actual comparable disagreement; uncorroborated candidates SHALL NOT create confirmed measures or replace originating regional levels.

Verified primary projections SHALL preserve prior facts and immutable receipts, reuse identical facts for repeat verification, and select revised records only within the service-knowledge boundary for the same primary source and explicit act/scope. A later acquisition returning an earlier retained content hash SHALL reuse that fact while preserving the intervening history. Disappearance, unrelated acts or secondary votes SHALL NOT establish cancellation. Protected measures SHALL retain their necessary comparison evidence; fully expired evidence graphs MAY be removed through dependency-safe retention.

#### Scenario: A primary publication returns to its earlier content
- **WHEN** a verified primary revision is followed by a new acquisition of previously retained content for the same act and scope
- **THEN** current selection follows the latest verified acquisition, reuses the earlier fact and retains the intervening facts and receipts for past knowledge queries

#### Scenario: A local activation is verified without a regional alert
- **WHEN** a platform candidate and municipal act support a local activation but no applicable regional warning is acquired
- **THEN** the supported local measure remains separate and no regional color or absence-of-risk conclusion is inferred

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


### Requirement: Validated municipal extraction reaches the public domain
Validated, complete primary municipal extraction SHALL feed an atomic and idempotent domain projection before ordinary downstream linking. The projection SHALL retain action, subject, place, original evidence and explicit temporal uncertainty. Evaluation-only runs, platform-only candidates, incomplete inputs and suspended interpretations SHALL NOT enter this path. Supported update links SHALL apply only after both primary measures are bound, with evidence on both sides and service-knowledge history preserved. An extracted activation SHALL NOT establish a named operational phase by inference.

#### Scenario: An ordinary extraction contains a closure without a definite end
- **WHEN** a retained primary publication produces a validated closure and a conditional end
- **THEN** API/MCP return the supported action, subject, place and passages, disclose unresolved current validity, and repeated projection creates no duplicate measure

#### Scenario: A newer extraction is used only for evaluation
- **WHEN** a evaluation run interprets an existing municipal document
- **THEN** it changes neither public facts nor the ordinary interpretation status of that document

### Requirement: Municipal processing startup preserves historical catalogs
The worker SHALL support multiple immutable fallback configuration revisions with the same catalog name without overwriting historical configurations or failing local configuration registration.

#### Scenario: A newer fallback encounters an existing local catalog name
- **WHEN** a newer fallback revision shares the original catalog name and its local registration encounters the legacy name/stage/revision constraint
- **THEN** the worker obtains an independent local configuration bound to that exact fallback
- **AND** repeated startup returns the same configuration and preserves the previous configuration and its hash


### Requirement: Evidence-bound criticality map interpretation
The service SHALL derive daily zone/risk levels from the retained criticality PDF's labelled filled vector polygons when its issuance matches the retained HTML and the complete seven-risk, two-day, 26-zone map set is recognized. It SHALL attribute each fact to the exact PDF URL and physical page, including editions with preceding adoption/scenario pages. Green, yellow, orange and red SHALL follow recognized polygon fills. A white zone polygon SHALL indicate non-applicability only for the main-river/coastal product masks; background whitespace SHALL NOT establish a level. Separately rendered irregular zone fills MAY be recognized; standalone anonymous rectangles SHALL NOT substitute for zone outlines. Reviewed A6/island label offsets MAY use bounded typographic distance and an unambiguous separated outline; arbitrary nearest-zone inference SHALL NOT be used. Missing labels/polygons, overlapping shapes, unsupported colors, raster maps and unsupported layouts SHALL remain unresolved. Map dates SHALL match the issuance day and next day. Explicit HTML alert rows SHALL retain their precise validity intervals and SHALL NOT be replaced by whole-day effective intervals.

#### Scenario: Applicable and excluded risks in a complete daily map set
- **WHEN** the retained matching PDF contains an unambiguous green polygon around the municipality's zone label for six risks and an outlined white polygon in the coastal-risk mask
- **THEN** the service returns six green risks and coastal waves as not applicable for each represented day
- **AND** the facts cite the retained PDF pages and their actual projection knowledge time

#### Scenario: Unsupported graphics or different edition
- **WHEN** map evidence is missing, a polygon cannot be associated with its label, a color is unsupported, or PDF and HTML issuance differ
- **THEN** unsupported fields remain unknown or the parser retains the preceding supported explicit HTML evidence
- **AND** a no-criticality statement alone never supplies green colors

#### Scenario: Precise alert table validity and daily maps
- **WHEN** an explicit table defines an afternoon alert interval and matching daily maps show its color
- **THEN** the interval remains the supported validity and repeated daily copies do not duplicate it
- **AND** a conflicting table/map color is rejected

#### Scenario: A historical edition includes leading non-map pages
- **WHEN** the matching issuance and complete labelled maps follow adoption or scenario pages
- **THEN** those preceding pages are not treated as risk maps and evidence retains the original physical map page
- **AND** a failed SVG font rendering may be retried through bounded local vector-PDF font normalization, without accepting partial failed output

### Requirement: Product-specific vigilance graphics
The service SHALL interpret a recognized retained vigilance PDF separately from criticality colors, matching the HTML issuance and both daily map dates. It SHALL require all 26 zones in each of the two rainfall panels, the distinct cumulative panel and the two other-phenomena panels. All seven risk categories SHALL remain represented, with vigilance warning levels `not_applicable` and optional weather metadata. Rainfall SHALL use literal bands from that PDF's eight-band legend, in millimetres averaged over the area; the cumulative band and literal cumulative period SHALL remain separate from daily validity. The two hydrological risk categories MAY share the rainfall observation without inventing hydrological alert levels. Thunderstorms, wind, coastal waves, snow and ice SHALL use that document's labelled graphical symbols. `depicted`, `not_depicted` and `unresolved` SHALL remain distinct; lack of a symbol SHALL NOT establish absence of risk. Explicit HTML phenomenon/zone evaluation SHALL remain distinct from a graphical weather value. Unrecognized marks, missing required resources, edition/date conflicts and unsupported graphics SHALL disclose uncertainty or retain the supported HTML fallback. New projections SHALL preserve older records and become visible only at their actual service-knowledge time.

#### Scenario: Daily rainfall differs from the cumulative panel
- **WHEN** daily panels show different rainfall bands and the third panel shows a cumulative band
- **THEN** API/MCP preserve both daily bands, the separate cumulative band, its literal period, unit and area-average scope
- **AND** neither rainfall white nor a weather symbol supplies a criticality color

#### Scenario: A phenomenon is listed but its graphical symbol is unsupported
- **WHEN** HTML lists a zone under evaluation or the zone contains an unrecognized graphical mark
- **THEN** the affected weather metadata is unresolved rather than asserting the phenomenon is absent
- **AND** supported facts for other zones retain their own evidence assessment while source coverage reports aggregate limitations

#### Scenario: Irrelevant PDF header curves preserve strict map interpretation
- **WHEN** a retained criticality PDF has a cubic header symbol whose entire transformed control-point hull lies strictly above every dated risk panel
- **THEN** the interpreter may exclude that symbol while reading the original maps and physical pages
- **AND** unsupported, transformed or crossing curves inside map panels remain rejected; no inferred colors or source acceptance are introduced.

#### Scenario: Retrospective closure in a completed reopening
- **WHEN** a completed circulation-reopening sentence mentions a previous closure in a subordinate clause
- **THEN** that clause alone MUST NOT create an additional current closure or an end time from the previous-closure expression
- **AND** closures independently evidenced by other operative clauses remain separate.

#### Scenario: Daily vigilance band is supported but cumulative fill is ambiguous

- **WHEN** a daily fill matches this edition's labelled rainfall legend but the cumulative fill does not
- **THEN** the daily graphical status remains depicted and its literal band is preserved
- **AND** total_graphical_status is unresolved and no cumulative band is guessed
- **AND** aggregate limitations identify daily, cumulative and symbol uncertainties separately
