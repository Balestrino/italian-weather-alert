# Tasks

## 1. Selection

- [x] 1.1 Extend the replacement predicate of `publicquery.measures()` with the equivalent-version branch of design decision 1: fingerprints within the cutoff, later `(first_acquired_at, id)`, validated complete non-evaluation extraction with a newer binding. Keep the same-version branch unchanged. Verify with `go vet` and the existing publicquery unit tests.
- [x] 1.2 Add publicquery PostgreSQL integration tests for every scenario of the delta spec:
  - template-only reinterpretation with reworded subjects;
  - a genuine revision omitting a closure;
  - a newer equivalent result that is not relevant, has no measures, or failed;
  - boundaries before the replacing projection and before each fingerprint;
  - same-version replacement unchanged.

  Also assert that reads write no row.
- [x] 1.3 Verify bounds and interfaces: a municipality with at least 40 template-only versions keeps a fixed number of measure-selection round trips (query-count harness), and HTTP/MCP return the identical selection for measure search and the municipality situation (transport equivalence test).

## 2. Verification and adoption

- [x] 2.1 Run unit, race, vet and `-tags=integration` for publicquery, transport and cmd, plus `openspec validate --strict`, and record the results.

  Results on 2026-10-23: unit and `-race` pass for publicquery, transport/httpapi and every `cmd` package. `go vet ./...` passes with and without `-tags=integration`. The `-tags=integration` suites pass for publicquery (18.9 s), transport/httpapi and cmd/iwa (6.8 s). `openspec validate --strict` passes. The new integration test fails against the previous predicate, which returned both measures of the template-only document at 45 s. The suites were rerun with the same results after the predicate became set-based (design decision 4).
- [x] 2.2 Adopt the public service in development with a settings backup. Verify through API and MCP:
  - Calcinaia documents 60 and 58 return only the measures of their latest equivalent projection;
  - documents 43, 15668 and 34667 keep their measures;
  - a boundary before the 21–22 October projections returns the earlier selection.

  Record the private evidence.

  Results on 2026-10-23: the public service runs `a4458c5` (set-based predicate); the first adoption of `79f6f6e` produced the same selections. Runtime readiness and smoke pass. API and MCP are identical for measure search (every page) and the municipality situation in four municipalities at both boundaries.
  - At the pre-adoption boundary, Calcinaia goes from 29 to 24 current measures. Only document 60 version 4291 (3 measures) and document 58 version 325 (2) leave the selection, so the documents show only versions 28212 (5) and 28211 (2).
  - Documents 43, 15668 and 34667 are unchanged. Versions 3812 and 4287 of document 43 share a fingerprint, but neither was in the public selection before or after the change.
  - At 2026-10-21T21:00Z the selection is identical (23 measures). No other municipality has public measures.
- [x] 2.3 Update the Calcinaia guide (new dated entry following CLN-039), the operations guide on equivalent interpretations (measure selection and the conservative tracked-target union) and the changelog with the verified results.
