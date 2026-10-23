## Purpose

Defines the updating state that saved public views report. Current views stay truthful as time passes, and historical views describe source freshness at the requested evaluation time.

## ADDED Requirements

### Requirement: Updating state follows the view's evaluation time
A public API or MCP view whose evaluation time is the request time, or later, SHALL refresh each item's updating state at every delivery. When the last complete source check is older than the source's delay threshold at delivery time, the state SHALL become `delayed`.

A view requested for an explicit evaluation time before its creation SHALL report the updating state computed at that evaluation time. It SHALL NOT turn `delayed` merely because the view is served later.

Equivalent API and MCP views SHALL report the same updating state.

#### Scenario: Historical view shortly after a source check
- **WHEN** a client requests a view for an evaluation and knowledge time five minutes after a complete source check whose delay threshold is thirty minutes
- **THEN** the updating state is `ok` in both API and MCP, even though the view is served days later

#### Scenario: Current view served after the threshold
- **WHEN** a current view is created before the delay threshold elapses and a later page is delivered after it
- **THEN** that page reports `delayed` with the delay limitation
