# Reviewed municipal PDF attachments

The crawl-based municipal collector supports an optional `attachments` policy in
each immutable source configuration. Use it only after reviewing the official
referral, document context and applicable reuse conditions. The
[Livorno guide](../territori/09-toscana/comuni/049009-livorno.md) and
[Cascina guide](../territori/09-toscana/comuni/050008-cascina.md) record separate source
research; [Municipium notes](../fonti/piattaforme/municipium.md) describe the
observed platform cases without granting permissions to other municipalities.

## Configuration and boundaries

`content_class` is one HTML class token. PDF links must occur inside that content
container and outside footer, navigation and header elements. A missing container
fails acquisition visibly; a recognized notice with no PDF links is valid.
Explicitly registered dependencies remain required even outside the discovery
container. Leaving the class empty retains whole-page PDF discovery.

Each `external` rule requires an exact HTTPS origin, a canonical directory prefix,
official referral evidence and a separate reuse policy allowing collection and
retention. Wildcards, a root directory grant, encoded or traversing paths and
external URL queries are rejected. This authorizes linked attachments only;
it does not add listing pages or discovery channels. The ordinary same-origin
behavior is preserved. Sources without `attachments` retain their prior checks.

For example, this synthetic configuration fragment requires completing the
placeholder evidence before registration:

```json
{
  "attachments": {
    "content_class": "article-content",
    "validate_pdf": true,
    "external": [{
      "origin": "https://assets.example",
      "path_prefix": "/s3/42/allegati/",
      "referral": {
        "url": "https://municipal.example/news/order",
        "locator": "Reviewed municipal attachment in the official notice",
        "observed_at": "2026-10-02T00:00:00Z"
      },
      "policy": {
        "evidence": {
          "url": "https://municipal.example/reuse",
          "locator": "Reviewed document-specific collection and retention terms",
          "observed_at": "2026-10-02T00:00:00Z"
        },
        "collection_permitted": true,
        "retention_permitted": true,
        "publication_permitted": false,
        "copies_permitted": false,
        "conditions": "Only reviewed municipal documents in the stated directory."
      }
    }]
  }
}
```

## Preview, activation and verification

Preserve the current source revision, service images, private environment settings
and database evidence before an environment change. Append a candidate revision
through the existing admin configuration endpoint, preserving the source's other
fields. Run its explicit preview before using the audited collection activation
control. Preview and scheduled collection share attachment validation for sources
with this policy, including the original publication metadata.

`DirectHTTP` checks external requests and redirects before sending them. Redirects
must remain on the file's initial origin and within the reviewed rule. With
`validate_pdf`, response type, PDF framing and Poppler parsing must pass; scanned
PDFs do not need readable text. Invalid responses are retained as an unavailable
reference, with `invalid_attachment_pdf` on a failed scheduled check. Required
missing resources prevent a complete result. Existing rate-limit and backoff
handling remains in force. Download validation does not guarantee future server
availability.

After activation, verify a fresh ordinary worker receipt, all planned documents,
version completeness and each necessary attachment. The document store verifies
object length and SHA-256 both after storage and on read. Re-read representative
retained PDFs through the internal store and parse them, rather than relying only
on HTTP status or listing discovery. Keep source captures, dumps, active settings,
resource hashes and detailed receipts under ignored territorial operations paths.

For rollback, preserve history and append a new candidate configuration copied
from the previous revision through the audited configuration control. Preview and
activate that newest revision; a historical revision that is no longer active
cannot be selected directly. Stop the worker before replacing incompatible
binaries, restore its recorded image/settings and resume it only after readiness
checks. Restoring the old configuration can reintroduce its attachment failures. Publication permissions, source acceptance and production
release remain independent operations.

## Repository verification

Synthetic tests cover exact origin/path boundaries, redirects denied before
requests, footer exclusion, missing containers, explicit dependencies, invalid PDF
responses and preview/collection checks. The isolated PostgreSQL test verifies
failure retention, later recovery, distinct metadata identities and rejection of
out-of-scope stored attachments. The application image supplies Poppler for the
real-parser test.

```sh
go test ./...
go vet ./internal/backend/acquisition ./internal/backend/registry ./internal/backend/documents
go test -tags=integration ./internal/backend/acquisition -run TestScopedAttachmentsPersistFailureAndRecover -v
```

Run `TestPDFValidationUsesRealParser` in the application image when Poppler is not
installed on the host. Passing fixture tests does not certify live source coverage.

## Cascina date repair and interpretation reuse

The separately reviewed Cascina revision uses `2 Jan 2006` with Italian month
normalization, `page-content` for news and attachment sections, and its own exact
API origin/tenant directory grant. One/two-digit date fixtures and persistent
tracking tests verify that a repaired old date leaves the recent acquisition
window; unknown dates and explicit-reference/ongoing/unresolved exceptions remain
eligible. No unavailable-target exclusion is needed for a reachable historical
page correctly dated outside the window.

`cascina-html-v1` removes only the 40-character alphanumeric values in the exact
reviewed CSRF meta/hidden-input forms from interpretation fingerprints. Retained
bytes remain unchanged. Enable Cascina explicitly in `IWA_PREFLIGHT_SOURCES`,
`IWA_OCR_REUSE_SOURCES` and `IWA_INTERPRETATION_REUSE_SOURCES` only after checking
the policy. Keep provider holds and local fallback scope independent. The isolated
preflight test verifies contiguous grouping, incomplete evidence, real changes,
archive/configuration boundaries and validated reuse without a second model call.

For a reviewed historical backlog, stop the worker and record exact version/job
IDs and the corrected official date privately. A scoped transaction must assert
the active revision, date, absent protection/references, unchanged version set,
no claimed check, no runs and no attempted/running jobs. Archive only those jobs
and interpretation versions with an actor and snapshot cutoff; preserve originals,
trigger identities, results and accounting. Global archive commands would affect
unrelated work and are inappropriate for this targeted repair. Resume the worker
and confirm the historical version is absent from the default pending admin view.
Archival records a decision to retire work, not a successful interpretation.

Development verification included a successful scoped preview, two complete
ordinary checks, corrected stored dates and real parser reads of retained PDFs.
The second check reused unchanged retained versions; no provider calls were made
for Cascina during these checks. That observation demonstrates unchanged-content
handling; the isolated test supplies the separate validated-result reuse evidence.
Fresh non-equivalent work can still await a held provider. Capture receipts,
rollback images/settings and the scoped archival operation privately. Restoring a
source configuration/image does not undo archives; any future historical
reprocessing is a separately selected action preserving prior history.

```sh
go test -tags=integration ./internal/backend/acquisition ./internal/backend/interpretation -run 'Test(RevisionTrackingBootstrapAndRecurringSelection|PreflightContiguousGroups|ScopedAttachmentsPersistFailureAndRecover)$' -v
```
