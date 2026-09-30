## Purpose

Avoid repeated inference on equivalent evidence while preserving original documents, complete interpretation coverage and verifiable result provenance.

## ADDED Requirements

### Requirement: Reusable OCR with traceable evidence
The service SHALL reuse successful OCR for identical page input under the same provider, model and processing configuration across retained versions. Reused results SHALL retain the current document, URL and page association and identify the originating inference result. Reuse SHALL NOT record the original provider charge as a new charge. Concurrent requests for the same reusable input SHALL coordinate ownership; uncertain interrupted calls SHALL remain explicitly uncertain rather than promising exactly-once provider billing.

#### Scenario: An unchanged attachment occurs in a new version
- **WHEN** a retained version contains page images with successful compatible cached OCR
- **THEN** those pages require no new provider calls and their evidence resolves to the current version and original OCR provenance

#### Scenario: OCR behavior changes
- **WHEN** the model, prompt, image preprocessing or inference settings change
- **THEN** incompatible cached OCR is not silently reused

#### Scenario: Two workers encounter the same uncached page
- **WHEN** both request OCR under the same configuration
- **THEN** one owns the active inference and the other waits for its durable result or bounded ownership recovery

### Requirement: Meaningful evidence controls automatic interpretation
The service SHALL retain permitted original bytes and their revision history independently of interpretation equivalence. Automatic inference SHALL be skipped only when all required interpretive inputs are proven equivalent under a versioned policy. New issuance, validity, operative text, necessary graphics, attachment contents, completeness or interpretation configuration SHALL invalidate equivalence. Uncertain equivalence SHALL require processing or expose a processing failure. Reused interpretation SHALL preserve evidence references and SHALL NOT create a new publication time or duplicate public effects.

#### Scenario: A regenerated PDF has only metadata changes
- **WHEN** raw bytes differ but the complete rendered pages, relevant metadata and other required evidence are identical
- **THEN** the new raw evidence remains retained and no fresh OCR, classification or extraction is needed for equivalent content

#### Scenario: A map or ordinance changes while HTML stays identical
- **WHEN** a required graphic or attachment changes meaningful contents
- **THEN** the new evidence is processed and any prior state is handled according to existing supersession rules

### Requirement: Explicit resource inference eligibility
The service SHALL distinguish confirmed decorative resources from necessary document and bulletin evidence through versioned source rules. Exclusion SHALL retain a reviewable reason and SHALL apply consistently to scheduling, completeness and interpretation inputs. Size, file extension or a generic image-name heuristic alone SHALL NOT exclude necessary evidence.

#### Scenario: A reviewed decorative icon accompanies a bulletin
- **WHEN** a configured rule identifies it as decorative
- **THEN** it produces no OCR request and does not block interpretation completion

#### Scenario: A small graphic carries a warning legend
- **WHEN** the graphic supplies meaningful bulletin evidence
- **THEN** it remains eligible regardless of its dimensions or similarity to an icon

### Requirement: Reuse respects retention and explicit reprocessing
Reusable evidence and dependent results SHALL remain resolvable within applicable retention and source access policies. Processing changes SHALL NOT automatically replay history. Explicit selected reprocessing SHALL honor the chosen configuration and disclose any compatible reused stages.

#### Scenario: Retention removes an originating version
- **WHEN** a retained result still depends on its reusable OCR artifact
- **THEN** cleanup preserves the required artifact or safely materializes its provenance before removing the origin

### Requirement: Coalesce equivalent pending revisions without paid fallback
The service SHALL distinguish raw revisions from contiguous equivalent interpretation inputs before provider calls, using complete local evidence and a versioned processing policy. Equivalent copies SHALL wait for a valid representative without consuming retry budgets, then use validated local result reuse with no provider fallback. Administration SHALL expose the grouping and retain access to each raw version and any failed reuse. Archived work SHALL NOT be reactivated by grouping.

#### Scenario: Repeated technical changes arrive during an account hold
- **WHEN** complete new raw revisions have identical preflight inputs to the immediately preceding unarchived version
- **THEN** they form one pending interpretation group, original evidence remains retained and no duplicate provider work is admitted

#### Scenario: Content returns to an older value
- **WHEN** meaningful content changes from A to B and then returns to A
- **THEN** the final transition is a new interpretation group rather than a duplicate of the first A

#### Scenario: An equivalent copy cannot materialize reusable evidence
- **WHEN** a representative completed but compatible reuse or literal provenance validation fails
- **THEN** the copy exposes a local processing failure and makes no provider request
