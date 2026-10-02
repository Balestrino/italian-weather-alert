# Local processing before model inference

The optional `local_processing` policy reduces classification inputs and avoids
model calls where reviewed formats provide sufficient evidence. It belongs to a
source configuration and requires dated review evidence. It does not grant
collection/reuse permissions, activate a source, or reprocess historical work.
Existing configurations preserve their previous behavior.

## Article-body selection

For an original HTML notice, declare the complete article containers after
checking representative notices, updates, exceptions, dates and attached acts.
Supported selectors are a single tag, `#id`, `.class`, `tag#id` or `tag.class`.
Compound CSS expressions, wildcards and arbitrary selectors are rejected.

Every selector must match exactly one nonempty container. The text of distinct
containers is joined in document order; selected descendants do not duplicate
an ancestor's text. Missing, empty or ambiguous matches fall back to full-page
text. All attachment text remains included, including HTML attachments that
happen to match a parent's selector. Retained original bytes and attachment
acquisition remain unchanged. Source review is necessary: a matching container
alone does not establish that all operative content occurs inside it.

This synthetic fragment illustrates a policy, not a reviewed live source:

```json
{
  "local_processing": {
    "version": "local-processing-v1",
    "evidence": {
      "url": "https://municipal.example/review",
      "locator": "Reviewed complete title/body/date containers and text-only ordinance directory; synthetic example",
      "observed_at": "2026-10-02T00:00:00Z"
    },
    "html_body_selectors": ["#notice-title", "article.notice-body"],
    "text_pdf_path_prefixes": ["/text-only-ordinances/"]
  }
}
```

## Native PDF text

`text_pdf_path_prefixes` declares canonical directory paths containing reviewed
text-only PDF documents. No root grant, encoded/traversing path or wildcard is
accepted. The policy is applied only to retained resources already governed by
that source configuration; it does not authorize a new host, redirect or download.

The existing application image supplies `pdfinfo`, `pdfimages` and `pdftotext`.
The extractor uses private temporary files, a 20-second context and bounded
outputs. It requires an unencrypted PDF with readable native text on every
numbered page, complete page delimiters, valid Unicode, at least 20 letters per
page and no raster images. Otherwise the entire resource follows normal rendered
image/OCR processing. Missing tools also fall back. No partial native-text result
is accepted as a complete document.

Text extraction cannot reliably identify meaningful vector drawings or maps.
Those documents must stay outside declared text-only scopes. Regional graphical
products reject native-text scope configuration. Scanned or mixed PDFs and
image-bearing documents use the model OCR fallback even when they have a text
layer. This deliberately favors complete evidence over the maximum call reduction.

Local results retain the original PDF URL/hash and page number, with
`application/pdf` input media and `poppler-pdf-text-v1` extractor identity. Usage
is `not_applicable`; no provider receipt, token count or remote price is fabricated.
An interrupted run preserves its chosen text/model path so later retries cannot
mix incompatible page evidence.

## Deterministic regional relevance

The optional `structured_format` is one of `cfr-criticality-v1`,
`cfr-vigilance-v1` or `cfr-monitoring-v1`, matching that source's existing regional
product contract. It cannot be combined with article selectors. For example, a
reviewed monitoring configuration can set:

```json
{
  "local_processing": {
    "version": "local-processing-v1",
    "evidence": {
      "url": "https://regional.example/review",
      "locator": "Synthetic review of explicit monitoring no-event layout",
      "observed_at": "2026-10-02T00:00:00Z"
    },
    "structured_format": "cfr-monitoring-v1"
  }
}
```

The classifier reuses the strict regional HTML projection. A validated explicit
fact/table or monitoring no-event statement can establish positive regional
relevance with literal original-page evidence. A strictly recognized vigilance
table with no listed zones remains relevant when its regional bulletin title is
literal evidence; this does not establish an all-clear. `structured-regional-v1` identifies
the local result, with no classification call or remote charge. Required content
must still be complete. Unsupported layouts, missing evidence and unresolved
content keep existing classification behavior. No title/keyword negative filter
is applied. Downstream extraction and graphical interpretation remain separate;
relevance alone does not establish a new alert color or a local measure.

## Activation, accounting and rollback

Append a complete candidate source revision through the existing audited
configuration controls, preserving permissions and other fields. Record the
source-specific review in its territorial guide and private register; compare
representative originals, attachments and model-fallback cases before activation.
Existing preview and collection activation controls remain required. This
implementation does not activate a live source policy or certify acceptance.

Workers derive separate immutable local-first processing configurations from
their selected fallback model configurations. Source-policy identities partition
classification inputs, OCR runs, preflight configuration/fingerprints and result
manifests. Exact originals and previous results remain available. Policy/catalog
changes do not automatically replay already processed documents; explicit
historical reprocessing remains selected and evaluated. Provider admission and
pre-claim holds remain effective, so local-first jobs can still wait behind a
provider hold. Account recovery and local model routing remain independent.

Migration `066_ocr_native_text` expands the existing page-result media constraint
to include native PDF inputs; existing image rows and prior migration checksums
are preserved. For rollback, append and activate a reviewed source revision
without `local_processing` and restore the previous compatible worker image.
Keep the additive migration and evidence. Existing native-text results are
retained; rollback does not remove history or replay jobs.

## Repository verification

Synthetic tests cover selector drift/order, late operative content, misleading
keywords, HTML/text attachment coverage, policy partitions, complete native text,
scanned/mixed/image/unreadable fallback and interrupted-write recovery. Disposable
PostgreSQL tests verify native page media, durable policy loading, separate
immutable configurations, zero-call/zero-price accounting, deterministic result
reuse and downstream extraction admission. Real Poppler verifies generated text,
blank-page, image-bearing and malformed PDFs.

```sh
go test -race ./...
go vet ./...
go build ./cmd/...
go test -tags=integration ./internal/backend/ocr ./internal/backend/interpretation -run 'Test(NativePDFPersistsWithoutProviderCalls|StructuredClassificationPersistsReusesAndSchedulesExtraction)$' -v
```

Run `TestPDFTextUsesRealPoppler` in the application image if Poppler is absent on
the host. These fixture checks do not establish live omission rates, source
acceptance, deployment or monetary savings; measure those after scoped activation.
