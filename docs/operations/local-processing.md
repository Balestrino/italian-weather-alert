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


## Retained-source replay — 2 October 2026

The bounded development replay used 50 retained versions: 35 municipal notices
from Cascina/Livorno and 15 regional bulletins from the three CFR products. The
candidate municipal selector `main#main-content` retained the article title,
complete `article-content`, checked metadata and all collected attachment text.
Across these municipal originals, normalized HTML input fell from 188,933 to
90,979 bytes (51.8%). This is input reduction, not a measured token or monetary
saving. Missing and ambiguous selector probes retained full-page fallback.

All 15 complete regional inputs produced positive local relevance; the 13 with
historical classified results agreed on relevance. The two other cases lacked
that historical comparison. The first replay exposed unnecessary fallback on
vigilance tables with no listed zones; the narrow recognized-table/title case
was corrected and regression-tested before the passing replay. No alert colors
or municipal measures were inferred.

Of 15 unique retained municipal PDFs, six passed native extraction (11 pages).
The nine others contained raster images and kept whole-resource OCR fallback.
The reader currently rejects even small embedded logos. Eight selected municipal
versions had incomplete interpretation evidence and remained incomplete. These
checks made no provider calls and used read-only database access. Original
captures, exact case selections, outputs and reproducible probe programs remain
in the private operational archive. Observations concern retained snapshots,
not a fresh source crawl, continuous availability or source acceptance.

## Fixed development trial

An explicitly selected trial uses a separate queue on eight fixed retained
versions, with one attempt per job and a cap of twenty calls to the existing
local model. Candidate policies are attached only inside that trial reader;
current source configurations and ordinary worker routing are preserved. Real
queue claims, immutable configurations, input manifests, processing results and
call receipts make the evaluated results visible in the existing backoffice.
Four municipal baseline/candidate pairs use the same model. Three regional
classifications and four individually reviewed native PDF resources exercise
local zero-call paths. The trial worker exits after the selected queue; no
automatic downstream interpretation or publication is scheduled.

For inspection, open **Operazioni → Job** and filter the trial queue, then
inspect **Consumi** and the documents for the selected sources. Detailed case
IDs, queue name, model responses, snapshot and rollback evidence belong to the
private register. Completion of this bounded trial does not enable a continuing
local-processing source policy; a broader trial must review directory scopes,
model quality and downstream extraction separately.

The completed development trial persisted sixteen successful jobs with no failures:
eight candidate classifications, four baseline classifications and four native
PDF resources (seven pages). The three regional classifications and all four
PDF resources made no model calls; other classification work made sixteen local
model calls in total, below the twenty-call cap. All four baseline/candidate
municipal pairs agreed on relevance. Their reported input tokens fell from
6,767 to 4,277 (36.8%); this small sample is not an omission-rate evaluation.
Local model attempts reported usage with unknown cost rather than inventing a
remote price. Native/structured attempts recorded usage and cost as not
applicable. Real database receipts and the private browser view confirmed all
sixteen jobs completed, and the temporary worker exited. Ordinary acquisition
and source configurations were preserved.
