# equivalent-interpretation-carry-forward Specification

## Purpose

Keep public document interpretation consistent across retained versions that differ only in presentation. When a newer equivalent version's own interpretation failed or is pending, the earlier validated interpretation is shown, with explicit provenance and limitations, without rewriting history or hiding newer results. Current public measures likewise follow the latest validated projection among equivalent versions of a document, so identical text interpreted more than once is not presented twice.

## Requirements

### Requirement: Configuration-independent equivalence of interpretation inputs
The service SHALL treat two retained versions as having an equivalent interpretation input only when all of the following hold:
- they belong to the same document and source;
- both inputs are complete and reusable;
- their interpreted text sections are identical, with the same resource URL, role and page;
- their metadata, issuer and completeness are identical;
- their resources have the same URLs, roles, media types, required and missing states, inference eligibility and processing policies.

The processing configuration SHALL NOT affect equivalence. Raw bytes of an HTML, text or JSON resource SHALL NOT affect equivalence when its interpreted text is identical. OCR resources SHALL be equivalent only when their page inputs and outputs are identical. An input lacking a retained manifest SHALL NOT be equivalent to any other input.

#### Scenario: Template change with identical text
- **WHEN** two versions of a municipal notice differ only in page markup outside the interpreted text, and their manifests were recorded under different classification configurations
- **THEN** the two inputs are equivalent

#### Scenario: Changed notice text
- **WHEN** a newer version changes one character in the interpreted text, a publication date in metadata, or an attachment's required or missing state
- **THEN** the inputs are not equivalent

#### Scenario: Different OCR output
- **WHEN** two versions share a scanned attachment URL but its retained OCR page output differs
- **THEN** the inputs are not equivalent

#### Scenario: Incomplete input
- **WHEN** either version's interpretation input is incomplete or marked not reusable
- **THEN** the inputs are not equivalent

### Requirement: Carried-forward interpretation for failed or pending equivalent versions
In public document reads, when a version's own interpretation at the knowledge boundary is `failed` or `not_processed`, the service SHALL look for earlier versions of the same document that meet all of the following:
- they were first acquired no later than this version;
- their input is equivalent to this version's input;
- they have a validated interpretation (`supported` or `partial`) recorded within the source's interpretation cutoff at that boundary.

If such versions exist, the service SHALL present the most recent such interpretation for this version with its original state. It SHALL add an evidence item and limitations identifying the following:
- the originating version, run and configuration;
- that the interpretation is carried forward from an equivalent version;
- this version's own outcome (`failed` with its stable failure code, or not yet processed).

The response SHALL keep this version's identity, hash and acquisition time. It SHALL NOT carry forward an `unreliable` interpretation, and SHALL NOT use another document or source.

#### Scenario: Newer template version failed under a new configuration
- **WHEN** a notice's earlier version has a supported extraction and its latest equivalent version failed extraction with `extraction_evidence_invalid` under a newer configuration
- **THEN** the latest version is presented as `supported`, with evidence pointing to the earlier version's run and configuration and limitations stating the carry-forward and the `extraction_evidence_invalid` outcome

#### Scenario: Newer version not yet processed
- **WHEN** an equivalent latest version has a recorded interpretation input but no interpretation result at the boundary, for example after a rejected classification, and an earlier version is supported
- **THEN** the latest version is presented as supported, carried forward, with a limitation that its own interpretation is not yet available

#### Scenario: No validated equivalent exists
- **WHEN** a failed latest version has no earlier equivalent version with a validated interpretation
- **THEN** its own `failed` or `not_processed` state and limitations are presented unchanged

#### Scenario: Earlier interpretation was unreliable
- **WHEN** the only earlier equivalent interpretation is `unreliable` because of a confirmed source defect
- **THEN** nothing is carried forward and the version's own state is presented

### Requirement: A version's own result takes precedence
The service SHALL present a version's own interpretation whenever one exists at the knowledge boundary other than `failed` or `not_processed`, including results that are `supported`, `partial`, `unreliable` or not relevant, even when an earlier equivalent version concluded differently. Carry-forward SHALL NOT create, copy, rewrite or delete retained interpretation results, receipts, domain measures or projections, and SHALL NOT change processing schedules, configuration identities or model calls.

#### Scenario: Later own result differs
- **WHEN** an equivalent latest version is later reinterpreted successfully and found not relevant, while the earlier version was relevant with measures
- **THEN** from that result's recording time the latest version presents its own not-relevant interpretation, and earlier knowledge boundaries still show the carried-forward one

#### Scenario: No retained data is written
- **WHEN** a public read presents a carried-forward interpretation
- **THEN** no interpretation, receipt, measure or projection row is created or modified

### Requirement: Consistent, historical and bounded carry-forward reads
The carried-forward presentation SHALL follow the same knowledge boundary as other public reads, using only versions, results and cutoffs known at that boundary. Equivalent API and MCP operations, document detail, the versions list, document search and saved views SHALL present the same state, evidence and limitations for the same data version and evaluation time. Document reads SHALL determine carry-forward for all requested versions in a bounded number of database round trips, independent of the number of versions in a page.

#### Scenario: Historical boundary before the earlier result
- **WHEN** a query's knowledge boundary precedes the earlier version's validated result
- **THEN** no interpretation is carried forward for that boundary

#### Scenario: API and MCP agree
- **WHEN** the same document is read through the HTTP API and MCP at the same saved view
- **THEN** both return the identical carried-forward state, evidence and limitations

#### Scenario: Page of many versions
- **WHEN** a document detail includes all retained versions of a notice with dozens of template-only versions
- **THEN** carry-forward is resolved for the whole page without one query per version

### Requirement: Current measures follow the latest equivalent projection
Public measure reads SHALL exclude a municipal measure projected from document version A when, at the query's knowledge boundary, a replacing extraction exists. The replacing extraction:
- is a validated, complete, non-evaluation extraction of a version B of the same document and source;
- has at least one projected measure recorded within the source interpretation cutoff;
- belongs to a version B that is either A itself with a later run, or a version first acquired after A whose interpretation input is equivalent to A's.

The equivalence SHALL be established by fingerprints of both versions recorded within that cutoff. Versions with different interpretation inputs SHALL NOT replace each other's measures. A newer equivalent result that is not relevant, has no projected measure, or failed SHALL NOT replace earlier measures. The service SHALL NOT create, rewrite or delete measures, bindings, extractions or receipts to apply this selection. Equivalent API and MCP operations SHALL apply the same selection.

#### Scenario: Template-only version reinterpreted
- **WHEN** version A of a notice has projected closures and a later version B with an equivalent input is extracted and projected again with differently worded subjects
- **THEN** current measure search returns only B's measures for that notice, and the evidence identifies version B

#### Scenario: Genuine revision of the notice
- **WHEN** a later version changes the interpreted text and its extraction omits one of A's closures
- **THEN** A's measures remain current unless linking applies a supported relation, because disappearance does not establish cessation

#### Scenario: Newer equivalent result is not relevant or failed
- **WHEN** the latest equivalent version's own extraction is not relevant, has no projected measure, or failed
- **THEN** the earlier projected measures remain current

#### Scenario: Historical boundary before the replacement
- **WHEN** a query's knowledge boundary precedes the projection of B's measures, or precedes the recording of either fingerprint
- **THEN** A's measures are returned as they were known at that boundary

#### Scenario: Same-version replacement preserved
- **WHEN** a later validated extraction of the same version projects measures
- **THEN** measures of the earlier run of that version are excluded exactly as before

#### Scenario: Interfaces agree
- **WHEN** the same municipality's measures are read through the HTTP API and MCP at the same saved view
- **THEN** both return the identical selection
