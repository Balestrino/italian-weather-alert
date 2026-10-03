## Purpose

Collect Toscana regional and municipal publications with verifiable authority, retained evidence and explicit collection and interpretation limitations.

## ADDED Requirements

The manually selected development publication exception is defined in [public alert access](../public-alert-access/spec.md#requirement-manual-municipality-publication-in-development). References below to acceptance before public activation apply to verified source publication in staging/production; the development exception grants consultation only and does not establish acceptance, infrastructure readiness or production authorization. Provenance, publication permissions, retained evidence and truthful interpretation remain mandatory.


### Requirement: Official source recognition and scope
The service SHALL register authority, publisher, technical platform, territory, product scope, access method, attribution and reuse evidence separately. The first operational MVP SHALL target Calcinaia local coverage alongside Toscana regional products. Livorno, Pisa, Pontedera and Cascina SHALL remain identified as subsequent planned local scopes, without implying active coverage. Discovery of municipal channels SHALL start from the official municipal website. An external channel SHALL require an explicit official referral from that website, with referring page, destination, context and observation evidence. A referral SHALL NOT imply broader coverage or transferred authority. Attribution of a decision to an associated entity acting within transferred functions SHALL require institutional evidence covering the relevant function, municipality and applicable period. Missing transfer evidence SHALL NOT by itself block acquisition or municipal attribution of a clearly identified municipal act published through a recognized channel; all other source acceptance and interpretation conditions SHALL still apply. Unresolved competence SHALL remain explicit only for the affected attribution or scope, without implying that unrelated verified municipal publications are unverified.

#### Scenario: A platform lists a municipality without a municipal referral
- **WHEN** an external platform exposes a municipality page but no explicit official referral has been verified
- **THEN** the platform remains a candidate and is not treated as a recognized municipal publication channel

#### Scenario: A referral covers plan maps
- **WHEN** the municipality links an external site specifically for plan maps
- **THEN** that evidence does not establish the site as a channel for all municipal alerts or measures

#### Scenario: A recognized external channel republishes a municipal ordinance
- **WHEN** a channel recognized for municipal notices publishes an identifiable mayoral ordinance and the associated-service transfer acts have not yet been fully reconstructed
- **THEN** the service may acquire and attribute the ordinance to the municipality subject to the other acceptance conditions, records the external publisher separately, and does not require unrelated transfer evidence

#### Scenario: An associated entity appears to issue its own decision
- **WHEN** a publication presents a decision as issued by the union of municipalities but the relevant function, territory or period of competence cannot be verified
- **THEN** the service records the stated issuer and missing institutional evidence without asserting verified competence or assigning the decision to a municipality by inference

### Requirement: Source acceptance before verified activation
The service SHALL report sources as not yet verified until official provenance, access/reuse conditions, declared sections and product coverage, representative extraction and update handling have been verified. For the scoped Cittadino Informato workflow, the technical multi-source verification requirement below SHALL define acquisition readiness without requiring acquisition of a license, agreement or separately verified legal basis as a task completion gate; public copies SHALL remain link-only. Acceptance SHALL include real documents and applicable updates, cancellations and attachments, comparisons with originals, delay measurements, failure and ambiguity behavior, and API/MCP equivalence. Untested cases and unresolved conditions SHALL be explicit. Partial coverage SHALL NOT be described as complete, and an accepted sample SHALL NOT imply coverage of every publication or other municipalities.

#### Scenario: One initial municipality is not ready
- **WHEN** regional products are accepted but a planned municipal channel has unresolved acceptance evidence
- **THEN** regional data remains available, that local scope is not yet verified, and complete initial local coverage is not claimed

### Requirement: Product-specific regional acceptance
The service SHALL record acceptance independently for vigilance, criticality/alert and monitoring, including operational channel, format and necessary graphical resources, access and product-specific reuse evidence, documented cadence and extraction evaluation. A historical example or normative template SHALL NOT establish operational availability. An unverified product SHALL remain explicitly unavailable in coverage results without preventing independently accepted and publicly enabled products from being returned. Partial availability SHALL NOT be described as completion of the three-product Toscana MVP.

#### Scenario: Monitoring has no verified operational channel
- **WHEN** vigilance and criticality are accepted and publicly enabled but monitoring has only a retained historical example
- **THEN** the accepted products remain available, monitoring is reported as unverified and unavailable, and the historical example is not served as current monitoring

#### Scenario: Reuse evidence covers only one product
- **WHEN** recorded reuse conditions establish permitted use for vigilance but not monitoring
- **THEN** monitoring acceptance remains pending and does not inherit vigilance permissions or the separate DPC license

### Requirement: Bounded municipal coverage evidence
Municipal acceptance SHALL identify configured channels and sections, evaluated period, listing traversal and pagination, required attachments, known-publication comparisons and unresolved gaps. A complete successful check SHALL certify only the declared checking scope. Neither a nonempty listing nor finding every sampled notice SHALL establish completeness of all municipal publications. A known relevant omission within the declared scope SHALL block that scope's acceptance until corrected and successfully re-evaluated; publishing narrower accepted scopes SHALL retain explicit coverage limitations. For Calcinaia, the direct municipal website SHALL remain the primary reference and an independently collected channel. Cittadino Informato SHALL be registered as an additional institutional acquisition/discovery channel under the scoped verification requirement below. Its role SHALL NOT establish completeness, operational dates or confirmed measures without supporting primary evidence. The albo pretorio SHALL be queried only for acts explicitly referenced by a primary municipal notice. Content found only through the additional channel SHALL remain an explicit diagnostic candidate until supported by a recognized primary publication or referenced act.

#### Scenario: General and typed listings differ
- **WHEN** a relevant known notice is absent from the general listing but appears in a configured typed section
- **THEN** evaluation verifies discovery through the configured sections and records date discrepancies without treating the general listing as exhaustive

#### Scenario: Every retained sample is found
- **WHEN** traversal finds all known sample notices for the evaluated period
- **THEN** coverage reports the evaluated channels, sections, period and remaining limitations without claiming that every municipal publication is collected

#### Scenario: The secondary platform omits a municipal notice
- **WHEN** the Calcinaia website publishes a pertinent notice that is absent from Cittadino Informato
- **THEN** the service collects the primary notice, records the secondary-platform omission as diagnostic evidence and does not reduce the authority of the municipal publication

#### Scenario: A municipal notice references an albo act
- **WHEN** a primary Calcinaia notice cites an ordinance or other act available through the albo pretorio
- **THEN** the service retrieves that referenced act as supporting evidence without treating the whole albo as an exhaustively collected alert channel

### Requirement: Scoped Cittadino Informato acquisition and primary verification
The service SHALL support Cittadino Informato as an additional acquisition/discovery channel for explicitly selected municipalities. Recognition SHALL identify the municipal referral, exact destination, covered sections, municipality ISTAT and observation date. Calcinaia's recognition SHALL NOT recognize or activate any other municipality or the entire platform. Platform operator, publisher and originating municipal or regional authority SHALL remain separate.

Recurring acquisition readiness and task 26.2 completion SHALL be based on an implemented, tested multi-source verification system spanning Regione Toscana/CFR, the selected municipality and cittadinoinformato.it. Obtaining licenses, agreements or documentation of a separately verified legal basis SHALL NOT be a prerequisite or completion criterion for this workflow. Acquisition SHALL use explicitly selected recognized channels, bounded publicly accessible sections/endpoints, operator-configured cadence and upstream retry instructions. Available official API/feed access SHOULD be preferred where technically verified; endpoints and upstream polling limits SHALL NOT be invented. Public platform copies SHALL remain link-only.

Acquired platform publications SHALL generate evidence-bound candidates and bounded verification work. Local measures and operational dates SHALL be verified against the municipal publication or its referenced official act. Regional republications SHALL be compared with the originating CFR product for the same risk, zone, issuance/version and validity. A local measure SHALL NOT require a concurrent regional alert. Comparisons SHALL retain both source/version identities, relevant passages, comparison time and supported fields; textual similarity alone SHALL NOT establish equivalence. Independent municipal and CFR collection SHALL continue within their declared scopes.

The service SHALL distinguish corroborated fields, primary evidence not found, primary check unavailable, non-comparable evidence, checks not applicable to the claim and actual comparable conflicts. Each verification receipt SHALL identify the candidate, the three channel roles, consulted source/version identities and relevant passages where available, attempted-check times, field-level results and reasons for missing or inapplicable comparisons. Regional claims SHALL use the originating CFR product; local claims SHALL use municipal publications or referenced acts. A local-only measure MAY have the CFR check marked not applicable; a missing municipal republication SHALL NOT invalidate a supported originating CFR claim. The service SHALL NOT require unanimous three-channel agreement or resolve disagreements by majority vote between republications of the same original. Missing or failed primary retrieval SHALL NOT be reported as contradiction, cancellation, absence of risk or established validity. Candidates without required primary support SHALL remain diagnostic and SHALL NOT become confirmed local measures or substitute regional levels in API/MCP. Agreement SHALL NOT complete source acceptance or establish real-world completeness. New or changed evidence SHALL trigger versioned reevaluation without rewriting earlier knowledge.

Task 26.2 SHALL remain open until a bounded development run exercises the acquisition-to-verification path and persists inspectable receipts across the three channel roles, including a regional republication and a local measure, with tested missing, unavailable, non-comparable and conflicting evidence behavior. The run SHALL verify independent primary collection and diagnostic-only admission for unsupported candidates. A planning edit or legal/technical documentation inventory alone SHALL NOT complete the task. The subsequent scheduled trial, source acceptance and public activation SHALL remain separate tasks.

#### Scenario: Multi-source verification closes acquisition readiness
- **WHEN** recognized selected channels have a working bounded acquisition and persistent verification path, and the required development run and failure cases pass
- **THEN** task 26.2 can be completed from inspectable multi-source results without obtaining license or legal-basis documentation, while source acceptance and scheduled activation remain separate

#### Scenario: A platform communication is corroborated by a municipal ordinance
- **WHEN** a selected recognized channel reports a closure and the municipal publication or referenced ordinance supports the same action, place and period
- **THEN** the candidate retains both evidence identities and only supported fields enter the validated municipal projection, with independent interpretation and publication gates

#### Scenario: Container metadata does not represent the dynamic notice
- **WHEN** a recognized platform page's CMS response has an empty body or creation/modification dates belonging to the page container rather than its displayed notices
- **THEN** the service does not treat those fields as notice content, issuance or operational validity and verifies a content-specific technical acquisition contract before collecting the notices

#### Scenario: Regional colors disagree for different validities

- **WHEN** a platform republication and CFR product refer to different issuance or validity periods
- **THEN** the comparison records non-comparability and does not report an actual conflict or replace the CFR level

#### Scenario: Primary verification is unavailable
- **WHEN** a platform candidate is acquired but the municipal source cannot be checked
- **THEN** the candidate and failed verification remain diagnostic with explicit limits, without asserting a confirmed measure, revocation or absence

#### Scenario: An independent municipal check finds a platform omission
- **WHEN** the municipal channel publishes a relevant notice absent from the platform
- **THEN** primary collection and interpretation proceed independently and the platform omission is recorded within the compared scope

#### Scenario: A local measure has no applicable regional warning
- **WHEN** a platform candidate is corroborated by a municipal act and no regional warning is applicable to that local claim
- **THEN** the receipt records the CFR role as not applicable, retains municipal/platform evidence and admits only the supported local fields without requiring three matching publications

#### Scenario: Two republications disagree with the originating CFR product
- **WHEN** platform and municipal republications agree with each other but conflict with the comparable originating CFR product
- **THEN** the system preserves the conflict and originating level rather than selecting the republications by majority vote

### Requirement: Retained acquisition and versions
The service SHALL retain original document bytes where permitted, source URL, acquisition time, hash, available publication/update metadata and interpretation version. It SHALL detect changes at stable URLs, preserve prior versions within retention, and avoid duplicate versions for unchanged content. An unchanged successful check SHALL NOT change publication time. Required linked resources SHALL be included or explicitly reported missing.

Generated transport-only Drupal Views DOM identifiers in HTML comments or class names SHALL NOT create a new content version. The retained original SHALL remain byte-exact, and changes to listing content or document provisions SHALL still create versions.

#### Scenario: A page is revised without a new URL
- **WHEN** a known municipal page changes its provisions
- **THEN** the changed version is retained alongside its prior version and interpretation provenance

#### Scenario: A listing response changes only its generated DOM identifier
- **WHEN** repeated HTML responses differ only in a generated Drupal Views DOM identifier
- **THEN** they share one content version while the first retained original remains byte-exact

#### Scenario: Distinct retained versions have equivalent interpretation inputs
- **WHEN** separate retained versions contain the same substantive HTML evidence under a reviewed interpretation policy
- **THEN** contiguous preflight grouping and classification reuse avoid duplicate provider calls while retaining each version's provenance

#### Scenario: A reviewed Cascina page changes only its CSRF values
- **WHEN** complete consecutive retained Cascina inputs differ only in the exact reviewed CSRF meta and hidden-input values under the same processing policy
- **THEN** source-scoped preflight and validated result reuse avoid new provider calls while retaining originals; changed dates, text, links, attachments or incomplete evidence remain distinct

#### Scenario: A municipal listing uses a single-digit day
- **WHEN** a source publishes `8 ottobre 2024` with the evaluated `2 Jan 2006` layout and Italian month normalization
- **THEN** discovery persists the official date and excludes old targets from the recent window except for existing explicit-reference or ongoing/unresolved-measure exceptions

### Requirement: Necessary attachments and ordinances
The service SHALL acquire linked attachments and ordinances when needed to establish a measure, territory or validity. If an attachment cannot be accessed or interpreted, it SHALL expose the available official document or link and identify the fields that cannot be determined. It SHALL NOT substitute bulletin validity or page expiry for a duration specified only in the ordinance.

An operator MAY configure municipal attachment discovery within an explicitly named document-content class. A missing configured content container SHALL fail the check visibly rather than silently omit attachments; generic links outside that container and navigation/footer links SHALL NOT become required attachments. Explicitly registered dependencies SHALL remain required according to their declarations. Configurations without this option SHALL retain their existing discovery behavior.

External attachments MAY be acquired under an explicit scope in the source revision carrying exact HTTPS origin, canonical directory prefix, official referral and separate collection/retention policy evidence. The scope SHALL NOT authorize broader channel discovery or other sources. Initial resource URLs and every redirect SHALL remain within the permitted resource scope and the resource's initial origin; unconfigured external resources SHALL retain visible forbidden references without download. An enabled PDF-validation option SHALL reject non-PDF, truncated or structurally unreadable responses rather than retain them as valid attachments. Preview and programmed collection SHALL apply the same configured attachment discovery, permission and validation checks. A successful programmed check SHALL follow verified object persistence; failures SHALL preserve originals and missing references, existing bounded backoff and source retry instructions, without asserting future availability.

#### Scenario: An authorized external ordinance and a generic footer PDF
- **WHEN** a municipal revision selects its document-content container and declares a reviewed external attachment scope
- **THEN** the necessary linked ordinance is downloaded and validated within that scope while the generic footer PDF is excluded; the declaration does not grant another source the same access

#### Scenario: An external download leaves its declared scope
- **WHEN** the linked URL or a redirect leaves the exact authorized origin or directory prefix
- **THEN** the collector refuses that request, preserves an explicit missing-resource failure and keeps the check incomplete

#### Scenario: A purported PDF contains an error page or is truncated
- **WHEN** PDF validation is configured and an attachment response is HTML or a structurally invalid PDF
- **THEN** both preview and scheduled collection fail visibly, preserve the parent original and mark the attachment invalid without claiming a complete acquisition

#### Scenario: A notice refers to an unreadable ordinance for duration
- **WHEN** the notice announces closures but their duration is only in an unreadable attachment
- **THEN** the notice and attachment link remain available and the closure duration is explicitly undetermined

### Requirement: Evidence-supported interpretation and visible failures
Structured conclusions SHALL be supported by identifiable passages or explicit values in acquired sources. Unsupported or ambiguous fields SHALL remain undetermined. For a publicly enabled recognized source, a newly acquired document SHALL remain publicly discoverable through metadata and its official link even if interpretation fails. Access to a retained copy SHALL follow the separate service setting and per-source restrictions defined in public-alert-access. The service SHALL identify the failure and any affected preceding interpretation as potentially superseded by unprocessed information; it SHALL NOT silently replace supported values with guesses or present previous values as fully current. Source access/reuse constraints SHALL continue to apply.

#### Scenario: A newer document cannot be interpreted
- **WHEN** a new official document is acquired but its measures cannot be established
- **THEN** document metadata and the official link are available with an interpretation warning, no unsupported measures are asserted, and preceding affected states disclose the newer uninterpreted publication

#### Scenario: An extraction response fails validation
- **WHEN** a structured extraction response fails envelope, evidence or model-consistency validation
- **THEN** the service privately preserves the request and verbatim response, attempt/window identity, finish reason and diagnostic code without credentials, retries or successful checkpoints; a capture failure stops the job with an explicit diagnostic-storage error

### Requirement: Configurable checks and complete-check status
The service SHALL support per-source configurable check intervals defaulting to 10 minutes and independent delay thresholds defaulting to 30 minutes since the last complete successful check. A complete check SHALL cover the declared sections, relevant listings and documents requiring revision checks, with successful content recognition. Reachability alone SHALL NOT reset the complete-check time. The first failed check SHALL be visible immediately; a source without a successful baseline SHALL be reported as not yet verified. These settings SHALL NOT be advertised as end-to-end delivery guarantees.

#### Scenario: The source returns HTTP 200 with unusable contents
- **WHEN** transport succeeds but the declared collection scope cannot be read correctly
- **THEN** reachability is recorded separately and the last complete successful check is not advanced

#### Scenario: Delay threshold is reached
- **WHEN** 30 minutes elapse without a complete successful check under default settings
- **THEN** the service reports updating delay while keeping publication validity separate

### Requirement: Bounded access and expected publications
Collection SHALL respect source constraints and upstream retry instructions, use bounded retries, and disclose resulting delays. Expected publication schedules SHALL be configured only from documented product cadence, with a configurable tolerance. The service SHALL distinguish collection failures from an expected publication not observed and from no publication being expected. Silence on an occasional municipal channel SHALL NOT itself be an error.

#### Scenario: A publisher requests a longer retry delay
- **WHEN** a source returns HTTP 429 with a retry time exceeding the configured interval
- **THEN** the next attempt respects that instruction and the resulting collection delay remains visible

#### Scenario: An occasional channel has no new notice
- **WHEN** a complete check finds no changes and no publication is expected
- **THEN** the check succeeds without a missing-publication alarm

#### Scenario: A reviewed stale document remains unavailable
- **WHEN** a previously discovered document still returns HTTP 404/410 and complete later traversals of the configured listings no longer contain it
- **THEN** an operator may record an evidence-backed exclusion scoped to its source, configuration and URL, preserving acquisitions, failed checks and unavailable-target history; rediscovery automatically restores the target to checking

### Requirement: Internal quality monitoring and DPC comparison
The internal monitoring system SHALL retain diagnostic evidence for availability, timeliness and interpretation correctness, including ambiguous relationships, omitted publications, source disagreements and parsing failures. It SHALL support assessment of code and process improvements against retained cases. It SHALL NOT introduce editorial approval, manual content correction or a publication gate for individual notices. DPC products SHALL be used only for internal comparison in the first release, with comparison scope and non-comparability explicit; they SHALL NOT overwrite regional values or appear as public warning products.

#### Scenario: A discrepancy is found during internal comparison
- **WHEN** comparable DPC and regional products differ
- **THEN** the diagnostic preserves both sources, times and comparison scope without changing the public regional warning

#### Scenario: A parser improvement is assessed
- **WHEN** a revised interpretation process is evaluated against a retained ambiguous case
- **THEN** the internal result records whether the issue is resolved without a manual edit of the public notice

### Requirement: Territorial source knowledge guides
The repository SHALL maintain a Markdown guide for each investigated region and municipality, indexed under `docs/territori/` and identified by region code and municipal ISTAT code where applicable. Guides SHALL contain identity and scope, source maps, current operational guidance, scoped rules and exceptions, open problems and a dated discovery register. Each discovery SHALL carry a stable identifier, source/product scope, observation and evidence, knowledge status, intervention status, last verification and a next check or closure criterion. Hypotheses, confirmed observations and superseded conclusions SHALL remain distinguishable from proposed, applied and verified interventions. Subsequent discoveries SHALL preserve and reference earlier entries while updating the current guidance. Shared platform characteristics SHALL be referenced separately and SHALL NOT establish permissions or exceptions for other territories.

Public guides SHALL contain only redistributable summaries and references. Unpublished operational evidence, credentials, captures and machine-specific settings SHALL NOT enter repository changes; private observations SHALL remain in the ignored operational archive with scoped evidence references. Guides SHALL link the coverage tracker without changing its dated acceptance assessments or activating collection/publication. Contributors SHALL consult and update relevant guides during source investigations and changes. New guides SHALL be created when research begins rather than implying investigated coverage for every registered territory.

#### Scenario: An external attachment and a footer PDF are discovered
- **WHEN** a municipal page links an ordinance on an external platform and an unrelated PDF in its footer
- **THEN** its guide records the observed links and current collector behavior separately from proposed corrections, links scoped platform notes and identifies the evidence needed to close each problem without granting cross-domain collection

#### Scenario: A later discovery contradicts current guidance
- **WHEN** a source change or new evidence invalidates a documented rule
- **THEN** a new dated entry references the original discovery, preserves its history and revises current guidance and open problems without treating the old conclusion as current

#### Scenario: Only private operational evidence establishes a failure
- **WHEN** source diagnosis depends on logs or retained responses not cleared for redistribution
- **THEN** details remain in the private archive, public guidance identifies its verification limits and neither guide nor successful document validation establishes source acceptance or public availability

#### Scenario: An operator rechecks enabled municipal acquisition
- **WHEN** the operator requests a new acquisition check for enabled municipalities
- **THEN** verification distinguishes territorial enablement from source collection, checks fresh programmed outcomes and retained originals for active sources, and records inactive sources separately without activating them
- **AND** scoped findings update the territorial guides while environment-specific configurations, check receipts and captures remain private; a successful check does not establish interpretation correctness or municipality-wide coverage

### Requirement: Operator-configured source lifecycle
Operators SHALL manage sources and sections through internal administration with persisted configuration history. A new source SHALL begin in draft, support an acquisition preview, and require explicit collection activation after that preview. New collection SHALL initially be internal; public enablement SHALL require separate operator action after source acceptance. Routine publication after enablement SHALL be automatic without individual-notice approval. Autonomous discovery of new source channels SHALL NOT be required.

#### Scenario: A new source passes a preview
- **WHEN** an operator reviews the preview and enables periodic collection
- **THEN** collection begins for internal evaluation and public access remains disabled until acceptance and explicit public enablement

#### Scenario: An operator changes collection boundaries
- **WHEN** a source URL, section or acquisition method is changed
- **THEN** the new configuration requires another successful preview before application while the previous configuration remains active; interval changes and suspension can apply immediately

### Requirement: Bounded bootstrap and revision checks
Initial collection SHALL cover notices from the preceding 30 days in configured sections and explicitly referenced older documents, disclosing gaps rather than claiming a complete historical archive. Recurring checks SHALL cover configured listings, notices within the preceding 30 days including previously irrelevant ones, older notices supporting ongoing or unresolved measures, and their linked attachments. Unchanged content SHALL NOT trigger automatic reinterpretation; changed supporting attachments SHALL be treated as evidence revisions even if the parent page is unchanged.

#### Scenario: A previously irrelevant page changes
- **WHEN** a page within the 30-day window gains a relevant provision
- **THEN** the next complete check detects its change and submits the full updated content for classification

#### Scenario: An old unresolved closure has a replaced attachment
- **WHEN** the closure notice is older than 30 days and its linked ordinance changes at the same URL
- **THEN** the check retains the new attachment version and schedules interpretation of the affected evidence

#### Scenario: A reviewed unavailable URL disappears from current listings
- **WHEN** an operator documents that a previously discovered URL returns 404/410, has no retained original and no longer appears in later fully traversed configured listings
- **THEN** an exact source/configuration/URL disposition may exclude that stale URL from subsequent revision plans while preserving its failure and check history, without asserting what the missing original contained or equivalence to another URL
- **AND** a later rediscovery of that URL makes it eligible for checking again; unreviewed 404/410 targets continue to make checks incomplete during retry deferral

### Requirement: Scanned documents and supported extraction
The service SHALL classify full content of new or changed notices, extract pertinent measures in structured form, and retain supporting passages for asserted fields. It SHALL include reading scanned PDFs and document images, with page references tied to retained originals. OCR formatting artifacts SHALL be normalized before inference without losing the retained page and position needed to verify a quotation against the OCR output and original document.

Long documents SHALL be divided into deterministic, bounded operational segments that retain source version, URL and page or positional references. Extraction segments SHALL carry enough deterministic local context to keep an operative statement with the subject, place and temporal expressions it qualifies; they SHALL NOT join separated text into unsupported evidence. Repeated evidence SHALL be represented compactly through validated references so that a document with many measures remains within the configured response budget without weakening field-level evidence.

Segment results SHALL be merged through validation that rejects unsupported assertions and exact duplicates. When supported candidate values conflict, the merge SHALL retain each candidate with its evidence and a shared conflict identity, leave only the affected resolved field undetermined and preserve unrelated supported measures. It SHALL NOT select one conflicting value, discard the complete document solely because of that conflict or treat a page/publication date as operational validity. Segmentation SHALL NOT relax full-content classification or field-level evidence requirements. Schema or evidence-reference validation SHALL NOT substitute for semantic evaluation. Ambiguous, conflicting, incomplete or unreadable content SHALL leave affected fields undetermined and documents discoverable according to source publication policy.

#### Scenario: An ordinance is a scanned PDF
- **WHEN** a pertinent notice links a scanned ordinance containing the duration of a closure
- **THEN** the service attempts to read the scan and records the extracted duration with original-document/page evidence, or explicitly reports it undetermined when reading fails

#### Scenario: An ordinance exceeds one extraction request
- **WHEN** a long ordinance must be processed in multiple bounded segments
- **THEN** operational segments and compact evidence references keep each request and response bounded, every asserted field retains segment and original-page evidence, and validated merge preserves supported relationships

#### Scenario: Two segments support incompatible temporal values
- **WHEN** a summary segment supports one operational start and a dispositive segment supports an incompatible start for the same measure
- **THEN** both candidates and their evidence are retained under one conflict, the resolved start remains undetermined, and other supported measure fields remain available

#### Scenario: OCR contains presentation markup
- **WHEN** OCR represents visible emphasis or headings with generated formatting markers
- **THEN** inference receives deterministically normalized text while evidence still resolves to the retained OCR page and original document without accepting a paraphrase

#### Scenario: An ordinance heading lists more objects than its operative body
- **WHEN** an ordinance subject heading lists closures but no operative passage supports one of the asserted measures
- **THEN** that heading alone does not establish the measure; independently supported measures and their conflicting temporal evidence remain available

#### Scenario: A partial reopening retains coordinated closures
- **WHEN** a notice reopens an underpass and explicitly retains closure of cemeteries and a dog exercise area
- **THEN** both retained closures and the bounded reopening remain represented with literal operative evidence, without deriving times from page metadata or revoking other restrictions

#### Scenario: A notice opens the municipal civil-protection centre
- **WHEN** an operative passage explicitly states that the named municipal civil-protection centre will be open
- **THEN** the evaluated extraction configuration can represent its activation with literal evidence, without treating an unrelated office opening as activation or borrowing the regional alert's validity

### Requirement: Durable processing and bounded retries
Processing SHALL preserve job states, errors and attempt history across restarts, prevent duplicate public effects from resumed work, and support a configurable maximum initially three total attempts for temporary processing errors with increasing waits. Ambiguity in readable content SHALL NOT itself trigger repeated attempts. Exhausted processing SHALL preserve acquired evidence and an explicit failure state.

#### Scenario: Processing fails repeatedly
- **WHEN** a temporary interpretation error exhausts the configured attempt count
- **THEN** automatic attempts stop, failure remains visible, acquired evidence is preserved, and an operator can explicitly relaunch the job

#### Scenario: A completed incident leaves a temporary recovery restriction
- **WHEN** the selected incident jobs have succeeded but their persistent recovery scope still defers ordinary work
- **THEN** an operator preserves the prior policy and call ledger, validates a bounded representative canary, and explicitly releases the temporary restriction without resetting job attempts, removing archives, bypassing provider holds or enabling public sources

#### Scenario: A selected recovery schedules measure dependencies
- **WHEN** embedding or linking work identifies its document through an extraction run
- **THEN** the recovery guard applies the parent document's scope and quarantine, permits allowed dependencies, and retains the shared provider-call cap

### Requirement: Measured quality before source publication
Source acceptance SHALL include human-reviewed expected outputs for observed cases and labeled simulations, covering irrelevant notices, omissions, partial reopenings, exceptions, conflicting dates, newer uninterpreted documents and scanned attachments. A missed relevant notice or unsupported asserted measure in evaluation SHALL block public activation of the affected source until correction and successful re-evaluation. Evaluation SHALL also report indeterminate fields and extraction usefulness. An observational collection trial SHALL last at least seven days and SHALL be extended for unresolved problems or insufficient observations; retained cases SHALL exercise significant events absent from the trial.

#### Scenario: A test misses a relevant municipal notice
- **WHEN** expected-source comparison detects the omission during acceptance
- **THEN** the affected source cannot be publicly activated until the omission is corrected and the relevant evaluation passes

### Requirement: Explicit reprocessing and interpretation suspension
Each interpretation SHALL identify its model, instruction/configuration versions and evidence inputs. Model or processing changes SHALL be evaluated on retained cases before operator-selected reprocessing by source, period or failure status; such changes SHALL NOT automatically reprocess the entire history. Prior runs SHALL remain traceable within retention. Operators SHALL be able to suspend new interpretations for a source with a confirmed interpretation defect while collection continues, mark affected published results unreliable, and resume after correction, validation and explicit reprocessing. These controls SHALL NOT allow manual rewriting of extracted measures.

#### Scenario: A model configuration changes
- **WHEN** an operator selects failed documents for reprocessing after evaluation
- **THEN** only the selected scope is reprocessed, each new run is versioned, and previous results remain traceable

#### Scenario: A reprocessing identity includes a selection and input manifest
- **WHEN** the classification run identity exceeds the store's identifier bound
- **THEN** a deterministic digest preserves all identity components and isolates distinct selections while existing bounded identities remain unchanged

#### Scenario: A published interpretation is found defective
- **WHEN** the operator suspends new interpretations for that source
- **THEN** collection continues, affected published results carry a reliability warning, and new interpretations do not enter public results until validated resumption

### Requirement: Internal operational visibility and usage accounting
Local-only administration SHALL expose source/configuration status, failed jobs, uninterpreted documents, evaluation findings, durations and consumption by model, stage and document version. Available input/output/cache token counts, attempts and estimated costs with pricing provenance SHALL be recorded per run. Unavailable usage SHALL be explicit rather than counted as zero. Bootstrap and ordinary collection SHALL be distinguishable, with OCR, embeddings, hosting and storage costs accounted separately. Secrets SHALL NOT appear in public output, logs or browser-visible configuration.

#### Scenario: An external provider omits token usage
- **WHEN** a processing response lacks usage fields
- **THEN** the run records usage as unavailable, retains other measurable charges/timing, and totals disclose the missing accounting

### Requirement: Email incident notifications
The service SHALL support private SMTP/recipient configuration and send administrative email for source delay thresholds, exhausted processing attempts and backup failures reported by the selected backup system or its notification integration. It SHALL group repeated errors, support configurable reminder intervals and send recovery notifications. Transient errors SHALL remain visible internally without requiring immediate email.

#### Scenario: A source remains delayed then recovers
- **WHEN** a source crosses its delay threshold, remains delayed and later completes a successful check
- **THEN** email reports the incident, sends reminders according to configuration and reports recovery without emailing every failed check

### Requirement: Configurable backups and verified recovery
The MVP deployment SHALL protect the shared service VM through off-host Proxmox Backup Server whole-VM backups covering all three environments' structured data, acquired originals in VM-hosted S3-compatible storage and configuration required for recovery. Backup schedule and retention SHALL be configurable deployment values and documented before operational readiness. The production backup plan SHALL target no more than six hours of lost data, schedule backups with margin for a delayed or failed job, and alert operators before the latest successful recoverable backup exceeds that age. A separate production application-level object-storage backup SHALL NOT be required while whole-VM protection is active. Backup failures SHALL be visible to operators and routed through the selected operational notification path. An isolated restoration rehearsal SHALL precede public activation and verify service startup, document/evidence references, retained versions and truthful updating status. The rehearsal SHALL time recovery from outage detection to verified public API/MCP service against a one-hour target; configuration alone SHALL NOT be reported as meeting that target.

#### Scenario: A backup is restored on a replacement host
- **WHEN** a whole-VM PBS backup is restored into an isolated replacement environment
- **THEN** the service starts with its database and original-document storage, retained records resolve their evidence, and pre-backup checks are not described as newly successful

### Requirement: Isolated development and staging with approved image promotion
The existing current-VM installation and its data SHALL remain the development environment. Development, staging and production SHALL use separate Compose project names, PostgreSQL and RustFS volumes, secret files, image references and loopback ports from one shared checkout on that VM. Staging SHALL use its own controlled test data; a copy from another environment requires a deliberate, consistent transfer of both database and objects. Production SHALL remain stopped while only its configuration is prepared, and its services SHALL have explicit CPU and memory ceilings with no default active profile. The production worker SHALL require separate selection. Before production activation, operators SHALL validate host capacity under concurrent workloads and rehearse off-host backup recovery from a host failure. A release image SHALL be built from an identified revision, tested in staging, published to the public GitHub Container Registry and identified by immutable digest. Production SHALL use that exact digest without rebuilding and SHALL require explicit operator approval. Passing software checks SHALL NOT by itself activate a source or establish public source acceptance.

#### Scenario: Development and staging run on the same VM
- **WHEN** staging starts while the existing development installation remains in place
- **THEN** its Compose resources, data, secrets and listener ports do not overlap with development, and development's existing project identity and volumes are preserved

#### Scenario: Production is configured but not started
- **WHEN** the production project is prepared in the shared checkout
- **THEN** its ports, secrets and named volumes are distinct, no production services are selected by default or created, and its resource ceilings are visible in the resolved Compose configuration

#### Scenario: An operator approves a staged release
- **WHEN** the recorded staging checks pass for a published image digest and the operator approves that digest
- **THEN** production pulls and runs that digest without a local rebuild, retaining the previous compatible digest and configuration for application rollback

### Requirement: Internal MVP with deferred infrastructure readiness
The service MAY collect and interpret the four configured MVP sources internally on the current dedicated VM before disk expansion, SMTP, PBS and continuous capacity telemetry are ready, as explicitly authorized by the operator on 2026-09-18. Each source SHALL still require its verified collection contract and a successful acquisition preview before collection activation. Errors and capacity SHALL be checked manually during this phase, and the absence of verified off-host recovery SHALL remain documented. Internal observations MAY contribute retained evaluation evidence but SHALL NOT establish infrastructure readiness, replace seven complete observational days or authorize public access. Public acceptance, release readiness and explicit deployment/public-enablement authorization SHALL remain separate gates.

#### Scenario: Internal collection begins before infrastructure hardening
- **WHEN** the operator enables a successfully previewed source on the current 30-GiB VM while SMTP, PBS and continuous telemetry are deferred
- **THEN** collection and interpretation run internally, the source remains publicly disabled, operational limitations are recorded and infrastructure tasks remain incomplete

### Requirement: Optional source-scoped local inference fallback
An explicitly configured local chat endpoint MAY replace an unavailable remote chat provider for selected sources. The service SHALL select a complete model-specific runner before creating an interpretation run, preserve separate provider gates and immutable configurations, and record the actual requested/returned model without assigning remote prices to local calls. Remote recovery controls, retries, archives, suspension and equivalent-copy restrictions SHALL remain effective.

#### Scenario: The remote provider is held or cooling down
- **WHEN** a quota/authentication hold or temporary availability circuit prevents remote processing for a selected source
- **THEN** a configured local chat runner can process the job without reopening the remote gate; recovery probes and a restored remote provider retain preference

#### Scenario: Both providers are unavailable
- **WHEN** the remote and configured local providers are blocked
- **THEN** queued jobs wait before claiming, preserving attempt budgets and failure evidence

#### Scenario: Fallback is selected for every registered source
- **WHEN** the operator explicitly lists all currently registered source IDs in the fallback and evaluated output-contract selections
- **THEN** admitted jobs from those sources may use the configured fallback while the primary is blocked, disabled sources remain disabled, source publication remains independent, and newly registered sources require an explicit configuration update

#### Scenario: A provider failure first establishes a hold
- **WHEN** an eligible remote job establishes an account hold or availability circuit
- **THEN** a subsequent attempt within the existing budget may select a separate local run; successful results and schema/evidence failures do not trigger another model call

#### Scenario: Two local server slots are used in development
- **WHEN** the operator requests a two-request trial, the local server exposes two slots and two development worker replicas share the existing queue
- **THEN** distinct admitted inference jobs may make concurrent requests with exclusive durable claims, separate call receipts and unchanged model, retry and source controls
- **AND** operators verify actual overlapping requests and validated results, preserve rollback evidence and document returning to one replica; scaling also applies to primary-provider jobs and other worker loops, without changing the default replica count or limiting unrelated clients

#### Scenario: Six corrected worker replicas are requested
- **WHEN** the operator expands the verified development worker from four to six replicas
- **THEN** the two additional replicas use the corrected queue-maintenance protocol and matching image/settings while the existing four remain running; verification distinguishes queue processing from simultaneous model calls and documents return to four workers

#### Scenario: Additional workers exceed the local server slot count
- **WHEN** the operator requests additional development worker replicas and the server exposes fewer slots than the worker count
- **THEN** replicas retain the same image/settings and exclusive queue claims, existing replicas remain running, and verification distinguishes running workers from concurrent model executions and documents returning to the previous count

#### Scenario: Concurrent workers maintain the inference queue
- **WHEN** several workers apply recovery, equivalent-copy and provider-hold deferrals before claiming inference jobs
- **THEN** maintenance is serialized within the database and errors or cancellation release ownership; job claiming and model execution remain parallel after maintenance returns

#### Scenario: A local model has limited capabilities
- **WHEN** a local model supports text but does not have evaluated vision or compatible embedding support
- **THEN** only evaluated text stages use that fallback, OCR/embedding keep their own controls and unresolved dependencies remain visible

#### Scenario: Local-model validation is requested
- **WHEN** opt-in live tests are run against a selected local endpoint
- **THEN** synthetic Italian cases exercise relevance, negative controls, measure extraction and update linking through production request/parsing contracts; mock success alone is not presented as model acceptance

#### Scenario: An equivalent copy has a compatible result from either model
- **WHEN** provider-free materialization finds a primary cache hit, or a specific primary reuse miss with an eligible local cache hit
- **THEN** it uses the compatible result without contacting either transport, keeps model/configuration boundaries and does not extend the queue attempt budget

#### Scenario: Image OCR has been separately evaluated
- **WHEN** image input is enabled on the local server and its synthetic OCR test passes
- **THEN** explicitly enabled OCR may select its own local image configuration, renderer/reuse identity and call receipts; semantic embeddings retain independent capability requirements


### Requirement: Reviewed local processing before model inference
The service SHALL support an optional versioned source-scoped local-processing
policy with dated review evidence. The policy SHALL preserve full article content,
required attachments, originals, literal evidence and page provenance. Policies
SHALL partition processing and equivalence identities, without automatic
historical reprocessing or changes to source/public activation.

#### Scenario: Reviewed article containers match
- **WHEN** every configured simple selector matches one nonempty article container
- **THEN** interpretation uses their combined text in document order and retains attachment content and originals
- **AND** missing, empty or ambiguous containers fall back to full-page text

#### Scenario: A reviewed text-only PDF has complete native text
- **WHEN** a PDF in an explicitly reviewed text-only directory has readable native text on every numbered page, no raster images and no encryption
- **THEN** local extraction persists text with resource/page and extractor provenance without a provider call or charge
- **AND** empty pages, scans, mixed content, images, malformed extraction or unavailable tools fall back to the original complete-resource OCR path; meaningful vector/graphical products are outside text-only scope

#### Scenario: A recognized regional format establishes relevance
- **WHEN** a reviewed regional format passes strict deterministic projection with complete content and literal evidence
- **THEN** a positive regional relevance result is recorded locally without classification calls
- **AND** unsupported, incomplete, ordinary or ambiguous content continues through the existing classification behavior without a keyword-based negative decision

#### Scenario: Local processing policy changes
- **WHEN** selectors, PDF scopes or structured-format policy change
- **THEN** processing input and preflight identities differ and incompatible cached results are not reused
- **AND** previous originals/results remain available and historical work is not automatically replayed

#### Scenario: A recognized vigilance table lists no zones
- **WHEN** the strict regional parser recognizes all six phenomenon rows and both dates, no explicit zone facts are listed, the regional bulletin title is literal original-page evidence, and required content is complete
- **THEN** local classification establishes only the bulletin's regional relevance without a model call
- **AND** no all-clear, alert color or local measure is inferred; a title without the recognized table or complete evidence retains normal fallback

#### Scenario: A local-processing development trial is explicitly bounded
- **WHEN** an operator requests a trial on fixed retained versions before broader activation
- **THEN** a separate worker queue may persist evaluated candidate results with immutable processing and evidence identities, a fixed model-call cap and existing provider gates
- **AND** candidate policies apply only to the selected trial inputs, model calls and zero-call paths remain distinguishable, the trial worker exits, and current source policies, ordinary worker routing and public activation remain unchanged
