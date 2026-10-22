# Spec Delta

## ADDED Requirements

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
