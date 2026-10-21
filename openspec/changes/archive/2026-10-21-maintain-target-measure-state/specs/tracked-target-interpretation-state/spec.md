# Spec Delta

## Purpose

Keep each tracked municipal target's measure, relevance and reference state truthful and current, derived from validated interpretation, so revision checks keep following notices whose measures may still be in force and stop following them on evidence or after a bounded period.

## ADDED Requirements

### Requirement: Measure state derived from validated interpretation
For listing-driven municipal sources, the service SHALL derive each tracked target's measure state from the municipal domain measures projected from any retained version of the same source and URL, their recorded validity, and the linked later measures applied to the domain. It SHALL NOT call a model, re-interpret documents or fetch sources to derive state. A measure SHALL be **closed** when its validity has a supported end instant earlier than the evaluation time, or when an applied linked later measure cancels it. Partial reopenings, extensions, corrections and updates SHALL NOT close a measure by themselves. Conflicting or undetermined validity SHALL count as an open end. The state SHALL be:
- `none` when no measure exists;
- `ongoing` when some unclosed measure has a supported end at or after the evaluation time;
- `unresolved` when no measure is ongoing and some unclosed measure has an unknown, start-only or conditional end, and the notice's age is within the source's open-ended measure limit;
- `aged_unresolved` in the same case once that limit has passed;
- `ceased` when every measure is closed.

The age SHALL be measured from the source publication date, or from first discovery when the date is unknown. The limit SHALL default to 90 days and SHALL be configurable per source from 1 to 366 days.

#### Scenario: COC activation with unknown end
- **WHEN** a notice published 40 days ago produced a COC activation whose end is not stated, under the default limit
- **THEN** its target state is `unresolved` and the notice stays in the recurring revision plan

#### Scenario: Open-ended measure passes the limit
- **WHEN** the same notice becomes 91 days old without a supported end, cancellation or rediscovered change
- **THEN** its target state becomes `aged_unresolved`, it leaves the recurring revision plan, and the state remains visible as no longer rechecked

#### Scenario: Closure with a supported end
- **WHEN** a closure valid until a supported instant tomorrow is the only measure of a notice older than the revision window
- **THEN** the target is `ongoing` and rechecked until that instant passes, after which it becomes `ceased`

#### Scenario: Partial reopening keeps the remaining restriction open
- **WHEN** a later measure reopens one underpass named in an earlier closure that also closed other places
- **THEN** the earlier measure is not closed by the reopening alone and its target state does not become `ceased`

#### Scenario: Cancellation by a linked later measure
- **WHEN** an applied linked later measure cancels every measure of an earlier notice
- **THEN** the earlier target becomes `ceased`

### Requirement: Relevance and explicit reference state
The service SHALL set a tracked target's relevance state from the latest successful classification of its latest retained version: `relevant`, `irrelevant`, or `unknown` when none exists. It SHALL mark a target explicitly referenced when an applied linked later measure, from any document of a source in the same municipality, targets a measure derived from that target's document. Explicit references SHALL NOT be cleared automatically. Existing explicit references SHALL be preserved.

#### Scenario: A later notice extends an earlier closure
- **WHEN** a linked later measure extending a closure is applied to the domain
- **THEN** the earlier notice's target is marked explicitly referenced and is included in a bootstrap plan regardless of age

#### Scenario: Classification changes on a new version
- **WHEN** the latest version of a notice is classified irrelevant after an earlier relevant version
- **THEN** its relevance state becomes `irrelevant` while its measure state continues to follow the measures derived from any version

### Requirement: Timely refresh before planning
Every scheduled check of a listing-driven municipal source SHALL refresh the derived states of that source's tracked targets before selecting its revision plan, using the check's evaluation time. The refresh SHALL be atomic per source and SHALL record when each target's state was last derived. When domain or classification data are unavailable, tracked states SHALL remain unchanged. A refresh failure SHALL make the check fail visibly rather than plan with partially updated states.

#### Scenario: Time passes without new content
- **WHEN** no new document arrives but a measure's supported end passes between two checks
- **THEN** the later check refreshes the target to `ceased` before planning and no longer selects it for age-based rechecks

#### Scenario: Refresh cannot complete
- **WHEN** the state refresh fails during a check
- **THEN** the check is incomplete with a stable error code and the previous states are preserved

### Requirement: Visibility of followed notices
Internal administration and a dedicated CLI command scoped to an explicitly selected environment SHALL list a source's tracked targets whose measure state is not `none`, or that are explicitly referenced. The list SHALL show measure state, relevance, explicit reference, publication date, age, the open-ended limit, when the state was last derived, and whether the target is currently selected for rechecks. Listing SHALL NOT start checks, change state or expose credentials and captures.

#### Scenario: Operator reviews followed notices
- **WHEN** an operator opens the source page or runs the list command
- **THEN** each followed notice shows its state, why it is or is not rechecked, and the time of the last derivation
