# regional-criticality-validity Specification

## Purpose

Defines how long CFR regional criticality facts are valid. Official alert rows, including alerts already in progress at issuance, bound the published validity. Daily map peaks must not extend an alert beyond its stated end.

## Requirements

### Requirement: Ongoing criticality rows keep their stated end
The service SHALL interpret a criticality table row whose time expression reads «in corso fino alle ore HH.MM di <weekday>, DD <month> YYYY» as an explicit alert interval:
- It SHALL start at the edition's literal issuance instant and end at the stated instant, both in Europe/Rome.
- It SHALL keep the original expression.
- It SHALL record an assumption that the alert was already in effect at issuance and that its earlier start belongs to preceding editions.

The service SHALL treat as unsupported a row whose end is missing, ambiguous, not later than the issuance instant, or falls in a nonexistent or repeated local time. Such a row SHALL NOT be projected with a whole-day or open-ended validity.

#### Scenario: Ongoing orange alert ends in the afternoon
- **WHEN** an edition issued at 13.17 lists zone A4 «in corso fino alle ore 17.00» at orange and a yellow row from 17.00 to the next day at 14.00
- **THEN** zone A4 is orange from the issuance instant until 17.00 and yellow from 17.00, and orange is not current after 17.00

#### Scenario: Ongoing row with an end before issuance
- **WHEN** an ongoing row states an end instant that is not later than the edition's issuance
- **THEN** that row is unsupported, the edition reports the limitation, and no validity is invented for that row

### Requirement: Tables with ongoing rows keep every explicit interval
A recognized criticality table SHALL retain the precise validity of all its rows when some of them are ongoing rows. The service SHALL merge these intervals with the edition's recognized daily maps under the same rules as other explicit intervals:
- the daily peak SHALL equal the highest overlapping interval level;
- contradictory overlapping intervals and genuine map/table conflicts SHALL remain failures that keep the supported fallback.

The presence of an ongoing row SHALL NOT cause the whole table to be discarded.

#### Scenario: Mixed ongoing and scheduled rows
- **WHEN** a table contains one ongoing row and two rows with explicit start and end times for the same risk
- **THEN** all three intervals are published with their precise validity and the daily maps do not replace them with whole-day facts

### Requirement: Explicit green outside listed alert intervals
When a zone/risk/day has at least one explicit alert interval, the service SHALL derive explicit green intervals for the parts of that day, in Europe/Rome, not covered by any listed interval for the same zone and risk. This SHALL apply only when:
- the edition's table is recognized, including its ongoing rows;
- the complete daily map set is recognized;
- the map reports the zone/risk as applicable.

Each derived green interval SHALL:
- cite the table and map evidence;
- state that it is derived from the absence of a listed alert for that period;
- remain distinguishable from a green read directly from the map.

The service SHALL NOT derive green when the map marks the zone/risk not applicable or unresolved, when the table or maps are unsupported, or for days without explicit intervals. Those days keep the daily map fact.

#### Scenario: Evening alert leaves the morning green
- **WHEN** zone A4 has a yellow thunderstorm interval from 18.00 to 23.59 and the daily map shows yellow
- **THEN** zone A4 thunderstorms is green from 00.00 to 18.00 with a derivation statement, yellow from 18.00, and the API reports green at 15.00

#### Scenario: No derivation without a supported table
- **WHEN** the edition's table is unsupported and only daily maps are available
- **THEN** no derived green interval is published and the daily map facts keep day precision

### Requirement: Corrected criticality projections preserve history
Corrected criticality validity SHALL be published as a new immutable projection logic version, recorded at its actual knowledge time. Earlier projection records SHALL remain unchanged and readable at their own knowledge boundaries. Reprojection of retained editions SHALL be explicit and bounded to selected versions. Views known before the reprojection SHALL return the earlier validity, and API and MCP SHALL return the same facts for the same view.

#### Scenario: Historical view before reprojection
- **WHEN** a retained edition is reprojected today and a client queries an evaluation and knowledge time before the reprojection
- **THEN** the earlier facts and validity are returned unchanged, and the corrected facts appear only for knowledge times after the reprojection
