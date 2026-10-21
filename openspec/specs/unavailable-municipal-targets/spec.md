# unavailable-municipal-targets Specification

## Purpose

Keep municipal check outcomes truthful when a tracked notice URL becomes unavailable after being renamed or withdrawn from listings, while preserving evidence and letting operators review and record dispositions without manual database edits.

## Requirements

### Requirement: Classification of unavailable municipal targets
For listing-driven municipal sources, a tracked target whose latest fetch returned HTTP 404 or 410, and that has not since recovered, SHALL be classified on every check as either **blocking** or **withdrawn**. A target SHALL be withdrawn only when all of these hold: the current check traversed every configured listing page without error; the target was not discovered in that traversal; the target has no explicit reference; and the target supports no ongoing or unresolved measure. Every other unavailable target SHALL be blocking. A target SHALL be treated as supporting an ongoing or unresolved measure when its maintained measure state, refreshed before planning as specified by `tracked-target-interpretation-state`, is `ongoing` or `unresolved`. A target whose measures are `ceased` or `aged_unresolved` SHALL NOT be protected by them. A reviewed `exclude_stale_unavailable` disposition SHALL continue to remove its exact source, configuration and URL from the revision plan until rediscovery or a `resume` decision. Classification SHALL NOT infer equivalence to another URL, cancellation of a provision or what a missing original contained. Retained originals, versions, interpretations and failure history SHALL be preserved.

#### Scenario: Renamed notice disappears from listings
- **WHEN** a previously discovered notice URL returns 404, is absent from the complete traversal, and has no explicit reference and no ongoing or unresolved measure
- **THEN** the target is classified withdrawn and its retained originals, interpretations and failure history remain unchanged

#### Scenario: Unavailable notice is still listed
- **WHEN** a target returns 404 or 410 but still appears in the current complete traversal
- **THEN** the target is blocking

#### Scenario: Unavailable target supports an open measure
- **WHEN** an unavailable target is absent from listings but is explicitly referenced or its maintained measure state is `ongoing` or `unresolved`
- **THEN** the target is blocking

#### Scenario: Withdrawn notice produced a municipal measure
- **WHEN** an unavailable target absent from listings produced a municipal measure whose maintained state is `aged_unresolved` or `ceased`, and it has no explicit reference
- **THEN** the target is classified withdrawn, and an operator exclusion of it still requires the measure acknowledgement

#### Scenario: Traversal does not complete
- **WHEN** a listing page fails, is unrecognized or the page limit is reached
- **THEN** no target is classified withdrawn on the basis of that traversal and the check is incomplete as before

### Requirement: Check completeness with withdrawn targets
A withdrawn target SHALL NOT make a check incomplete, whether it is fetched again in that check or waiting for its next permitted attempt. A blocking target SHALL keep the check incomplete during retry deferral, as before. A check that is otherwise complete SHALL be recorded as complete, SHALL advance the last complete check time, and SHALL record the number of withdrawn targets it observed, separately from collection failure codes. Withdrawn targets SHALL remain in the revision plan for as long as their revision window or measure state allows, retried under the existing bounded backoff. A later successful fetch SHALL clear the unavailable state and retain a new version without overwriting earlier evidence. Rediscovery in a listing SHALL make a still-unavailable target blocking. Other HTTP errors, rate limits and service failures SHALL keep their current effect on the check.

#### Scenario: Only withdrawn targets remain unavailable
- **WHEN** all listings and planned documents succeed except targets classified withdrawn
- **THEN** the check is complete with no error code, the source's consecutive failures reset, and the check records the withdrawn count

#### Scenario: A blocking and a withdrawn target coexist
- **WHEN** one unavailable target is withdrawn and another is blocking
- **THEN** the check is incomplete because of the blocking target, and the withdrawn target is still counted

#### Scenario: A withdrawn URL is restored
- **WHEN** a withdrawn target later returns HTTP 200
- **THEN** its unavailable state is marked recovered, a new version is retained when the content differs, and earlier failures remain in history

#### Scenario: Source delay incident follows completeness
- **WHEN** checks complete while only withdrawn targets remain unavailable
- **THEN** the source-delay notification does not open or remain open because of those targets

### Requirement: Operator visibility and reviewed dispositions
Internal administration and dedicated CLI commands scoped to an explicitly selected environment SHALL list a source's unavailable targets with URL, HTTP status, first and last failure, next attempt, last seen in listings, current classification (blocking, withdrawn or excluded) and disposition history. Operators SHALL be able to record an `exclude_stale_unavailable` or `resume` decision for one exact source, configuration and URL, with operator identity, reason, evidence and the observed last-seen time. A decision SHALL be accepted only if the target is still unavailable and not recovered, its last-seen time equals the observed value, it belongs to the active configuration, and exclusion targets have no explicit reference. An exclusion of a target that supports a measure SHALL additionally require an explicit operator acknowledgement that the measure no longer needs checking, retained with the decision. Otherwise the decision SHALL be rejected atomically, with a stable, private-detail-free error code. A decision SHALL NOT start a check, call a provider, change publication or accept a source. Administration forms SHALL work without JavaScript and preserve source context after an action.

#### Scenario: Operator lists unavailable targets
- **WHEN** an operator opens a source's unavailable targets in administration or runs the list command
- **THEN** each unavailable target shows its status, timing, classification and disposition history without exposing credentials or captures

#### Scenario: Reviewed exclusion is recorded
- **WHEN** an operator submits an exclusion with matching observed state, reason and evidence
- **THEN** exactly one decision is recorded with operator attribution and the target leaves the revision plan until rediscovery or resumption

#### Scenario: Target changed since review
- **WHEN** the target was seen in a listing, recovered or gained an open measure after the operator's observation
- **THEN** the decision is rejected without changes and the operator must reload before retrying

#### Scenario: Exclusion of a target with a derived measure
- **WHEN** an operator submits an exclusion for a target that supports a municipal measure without the acknowledgement
- **THEN** the decision is rejected; with the acknowledgement, reason and evidence it is recorded with that acknowledgement

#### Scenario: Excluded target is resumed
- **WHEN** an operator records a `resume` decision for an excluded target
- **THEN** the target returns to the revision plan and is classified on the next check
