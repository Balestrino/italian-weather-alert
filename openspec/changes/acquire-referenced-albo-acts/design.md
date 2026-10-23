# Design

## Context

- Ordinary municipal processing runs in this order: acquisition, OCR, classification, extraction, then `ProjectMunicipalExtraction`. Projection writes `domain_local_measures` with bindings, a validity row in `domain_temporal_values` and a row in `domain_interpretation_events`.
- Validity rows are append-only. Public reads take the latest row within `registry_interpretation_cutoff`. `municipalValidity` returns `unknown` for conflicting temporal candidates. It recognises an explicit start only in the notice pattern "dalle ore HH.MM di [giorno] D mese".
- Public measure evidence comes from `extraction_evidence` through the binding (`publicquery.measureEvidence`).
- Retained attachments are already gathered into a version's content, including OCR pages. The 20 August notice retains Ordinanza 34/2026 this way.
- The Calcinaia albo runs on JCityGov "papca" at `calcinaia.trasparenza-valutazione-merito.it`. The 23 October 2026 private review observed:
  - the "Storico atti" list at `/web/trasparenza/papca-ap/-/papca/igrid/248878/3791`, server-rendered with 20 rows per page, ordered by descending registry number;
  - search by object and by exact registry year and number works;
  - date filters returned nothing in all attempts;
  - record details at `/papca/display/<id>?p_p_state=pop_up` list category, act number and date, registry number, object and attachments;
  - attachment download URLs are base64-encoded in the detail markup, and the unsigned PDF downloads without a session;
  - an "Esporta in OpenFormat" list export exists, not yet assessed.
- The `CittadinoInformato` registry configuration and acquisition adapter are the precedent for a separately registered, policy-scoped municipal channel.

## Goals / Non-Goals

**Goals:**
- Reference handling that is deterministic, model-free and explainable, with append-only state.
- A bounded number of albo requests per reference.
- Clarification that is reproducible from retained literal text only.

**Non-Goals:**
- Model-based reading of acts.
- Acts that modify or revoke measures of other notices.
- Albi of other platforms or municipalities, beyond a configuration shape that allows them later.
- Backfilling all historical notices. Only notices with projected measures in the source's normal window are processed, plus explicitly named development cases.

## Decisions

1. **Separate albo source with a JCityGov access method.**
   - **What:** `registry.Configuration` gains `AccessMethod = "jcitygov-albo"` and a `JCityGovAlbo` block. The block holds the origin, list path, detail path prefix, allowed categories, the page bound per reference (default 6), and a link to the primary source it serves. Validation mirrors `CittadinoInformato`: origin and paths must be HTTPS and the reviewed policy must be complete.
   - **Policy:** the source policy permits collection and retention and forbids copies. It has no listing schedule and no bootstrap.
   - **Albo documents:** the official URL is the detail page. The detail HTML is the original resource; the main attachment PDF is required.
   - **Why:** keeping the albo out of the primary source preserves that source's version identity and coverage, and keeps its policy reviewable. *Alternative:* add albo PDFs as attachments of the notice version. Rejected, because it would create new notice versions and mix two publishers' scopes.
2. **Model-free reference detection.**
   - **Patterns:** a literal pattern set over interpreted content sections. It uses kind words (ordinanza, decreto, delibera/deliberazione, determina/determinazione), optional "n./n°/numero" N with "/AAAA" or "del DD/MM/AAAA", and unnumbered forms only in operative sentences ("apposita ordinanza", "ordinanza comunale", "emessa ordinanza", "mediante ordinanza").
   - **Exclusions:** site chrome. The navigation block is identified by a reviewed per-source fingerprint ("Albo pretorio Meccanismo di feedback…") and is excluded.
   - **Synthetic fixtures** cover:
     - positive: numbered with year; numbered with date; unnumbered operative forms;
     - negative: footer link; legal boilerplate ("deliberazione di Giunta comunale n. 129 del 24/10/2023" inside the act); negated or hypothetical forms.
   - **Scope:** references are detected only on the notice, never on acts, so "Visto…" citations inside acts are never followed.
   - **Why:** the citation forms observed in Calcinaia are few and literal. A model call would add nondeterminism, as seen in task 31.2.
3. **Resolution order and candidate rule.**
   - **Order:** `attached`, then `numbered`, then `unique_candidate`, as the spec requires.
   - **Bounded search:**
     - registry numbers increase with publication date, so the client finds the page range covering the date window with a bounded binary search over list pages ordered by registry number;
     - it then reads at most `page_bound` pages;
     - exact registry search confirms a candidate's detail.
   - **Matching:**
     - category: the albo category must equal the kind (ORDINANZE for "ordinanza");
     - date: the act date must fall in [publication − 7 days, publication];
     - subject: disposition terms are normalised head nouns of projected measure subjects (for example "cimiteri", "parchi", "sottopasso"), plus each measure kind's Italian keyword; at least one must occur in the object.

   *Alternative:* the date filter of the search form. Rejected because it was not reliable in the live review. Task 2.1 checks the OpenFormat export and adopts it only if it is complete and stable.
4. **Acts are evidence, not measures.**
   - **Pipeline:** albo versions go through OCR only. The existing gathered content provides page text with offsets.
   - **Clause parser:** a literal parser for the dispositive clause of municipal emergency ordinances:
     - "ORDINA … a partire dalle ore HH.MM del giorno D mese AAAA, fino al/alle … :" followed by bullet dispositions;
     - each disposition keeps a kind keyword: chiusura → closure, divieto → prohibition, sospensione → suspension, attivazione → activation;
     - each disposition keeps its literal items.
   - **Matching a measure:** a projected notice measure matches a disposition when the kinds are equal and the measure subject's head term occurs in that disposition's items.
   - **Why:** the spec forbids duplicate measures, and the act's dispositive clause is the authoritative literal. *Alternative:* interpret the act as its own primary document. Rejected by the operator decision, because of duplicates and changed provenance.
5. **Append-only clarification storage.**
   - **Tables:** migration `101_referenced_acts` adds four tables:
     - `act_references` (notice version, ordinal, kind, number, year, date, passage, locator, detected_at);
     - `act_resolutions` (reference, outcome, method, candidates JSON, act document version, reason, recorded_at);
     - `act_clarifications` (local measure, act version, disposition ordinal, passages, outcome `clarified|conflict|not_matched|clarification_unavailable`, recorded_at);
     - `act_albo_records`, mapping albo record id to retained document.
   - **Writes for a clarified measure:** one `domain_temporal_values` row (`evidence_document_version_id` = act version) and one `domain_interpretation_events` row with reason `clarified_by_referenced_act`. The condition stays as a limitation.
   - **Integrity:** every table uses `processing_catalog_immutable()` triggers and retention participation like other interpretation tables.
6. **Jobs and bounds.**
   - **Job chain:** `resolve_act_references` (after a successful projection with at least one measure), then `acquire_albo_act` (per new albo record), then the existing OCR job, then `clarify_act_measures` (when the act's content is complete).
   - **Limits:** at most 3 references per notice version. Albo requests per reference stay within the page bound and pass through the existing retry/backoff and provider-gate style limits.
   - **Idempotency:** idempotency keys use version, reference ordinal and albo record.
7. **Public and operator surfaces.**
   - **Public evidence:** `measureEvidence` appends act evidence items from `act_clarifications` recorded within the cutoff. Each item carries the official detail URL, the page and the passage, with locator `referenced_act:<kind> <number>/<year>`. Validity and interpretation come from the appended rows, so no new read path is needed.
   - **Operator surfaces:** the admin source page and the CLI command `source-act-references <source>` list references read-only.

## Risks / Trade-offs

- [An unnumbered match picks the wrong act] → The match requires a single candidate on kind, a 7-day window and a subject term. Ambiguity is recorded, never resolved, and every clarification shows the act link for review. Synthetic fixtures cover an ordinance and a decree with the same object.
- [The albo markup changes] → The parser and client fail closed to `unavailable` or `clarification_unavailable`. A live preview task records the reviewed page shapes, and changes surface in the operator list.
- [Load on a third-party portal] → Requests are bounded per reference, made only for relevant notices with measures, and use the existing retry discipline. Nothing is crawled.
- [The literal parser covers only one template] → Other templates record `clarification_unavailable`, an explicit limitation. Widening the parser requires new reviewed fixtures.
- [A conditional end stays open] → This is intended. "fino al perdurare dell'emergenza" is not an end date, and the measure stays conditional rather than ceased.

## Migration Plan

1. Apply migration `101` through the admin role.
2. Register and preview `calcinaia-albo` in development with a reviewed policy and evidence.
3. Deploy the admin, worker and public services from one revision.
4. Run resolution for the acceptance notices through bounded reprocessing of their projections.
5. Verify the documents through API and MCP, and at earlier boundaries.

**Rollback:** restore the previous images and disable collection for the albo source. The additive tables stay.

Appended validity and interpretation rows are facts recorded at their time. Older readers still return them as the latest validity, but without the act evidence item. Rolling back therefore stops new clarifications and leaves existing ones in place; it does not remove them. Removing a clarification would mean rewriting history, which the spec forbids.

Tracked-target measure state reads the latest validity too. A conditional validity keeps a target `unresolved`, so no target becomes `ceased` through a clarification.

## Open Questions

- Whether the "Esporta in OpenFormat" export can replace list paging, as task 2.1 assesses. Either way the bounded search contract is unchanged.
