# Contributing to Italian Weather Alert

IWA aims to make official regional weather and hydrogeological alerts, and related municipal notices and measures, accessible across Italy. Developers, source researchers, and citizens who compare results with official publications can all help. The service is still in an internal pilot. Use this repository's issues and pull requests to report reproducible problems or propose changes.

## Report a discrepancy

Check the original authority's page first. A missing result may mean a source is not yet onboarded, its check failed, or the document was not interpreted correctly. The [coverage tracker](docs/coverage.md) shows documented onboarding status; it does not certify live completeness.

For a report, include:

1. Municipality name and ISTAT code; CAP if it helped you find the municipality. For a regional bulletin, name the region and alert zone if known.
2. Official source URL, issuing authority, and the document's publication and validity times, including the timezone if shown.
3. When you checked IWA, what it returned, and what you expected to see.
4. The exact missing or incorrect fact, with a document section, page, or short excerpt that supports it.
5. Whether the problem concerns a regional warning, a municipal republication, or a municipality's own measure.

Use public official URLs when possible. Do not include private contact details, credentials, or personal information from an alert document. A screenshot can help, but the original URL and observation time are more useful.

## Contribute code or source research

Start with a focused issue or pull request. Keep regional warnings and municipal measures separately attributed. New sources need an official referral, a defined scope, access and reuse checks, representative examples, and explicit limitations; a source candidate is not accepted public coverage. Link to the authority's original publication and explain what the proposed source would cover.

Read the [OpenSpec plans and progress](openspec/README.md) before changing an existing capability. Update the relevant change specification and task checklist with the implementation, or propose a new change when the behavior is new. Checked tasks are a dated project snapshot, not a claim that a source is publicly accepted.

For Go changes, run:

```sh
go test ./...
go vet ./...
go build ./cmd/...
go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...
```

The [developer setup guide](docs/developer-setup.md) explains installation, local configuration, and stack checks. Use synthetic fixtures for ordinary tests. Include real source material only when its provenance and reuse terms are recorded. Describe what changed, how you checked it, and any known gaps in your pull request.

Changes intended for a deployed release are checked on the isolated staging stack described in [Environments](docs/environments.md). Production uses only a manually approved image digest after the [release checks](deploy/release-operations.md); a passing pull request or staging test does not activate an alert source.

## What happens to a source report

A report helps prioritize investigation. A source is configured, previewed, observed, reviewed, and explicitly enabled in separate steps. Finding one matching bulletin or municipal notice does not establish complete coverage for a territory. Never interpret “no result” as an official all-clear.
