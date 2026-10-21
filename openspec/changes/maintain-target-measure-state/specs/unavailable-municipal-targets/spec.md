# Spec Delta

## MODIFIED Requirements

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
