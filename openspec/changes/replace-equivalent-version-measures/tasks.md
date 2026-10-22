# Tasks

## 1. Selection

- [ ] 1.1 Extend the replacement predicate of `publicquery.measures()` with the equivalent-version branch of design decision 1: fingerprints within the cutoff, later `(first_acquired_at, id)`, validated complete non-evaluation extraction with a newer binding. Keep the same-version branch unchanged. Verify with `go vet` and the existing publicquery unit tests.
- [ ] 1.2 Add publicquery PostgreSQL integration tests for every scenario of the delta spec:
  - template-only reinterpretation with reworded subjects;
  - a genuine revision omitting a closure;
  - a newer equivalent result that is not relevant, has no measures, or failed;
  - boundaries before the replacing projection and before each fingerprint;
  - same-version replacement unchanged.

  Also assert that reads write no row.
- [ ] 1.3 Verify bounds and interfaces: a municipality with at least 40 template-only versions keeps a fixed number of measure-selection round trips (query-count harness), and HTTP/MCP return the identical selection for measure search and the municipality situation (transport equivalence test).

## 2. Verification and adoption

- [ ] 2.1 Run unit, race, vet and `-tags=integration` for publicquery, transport and cmd, plus `openspec validate --strict`, and record the results.
- [ ] 2.2 Adopt the public service in development with a settings backup. Verify through API and MCP:
  - Calcinaia documents 60 and 58 return only the measures of their latest equivalent projection;
  - documents 43, 15668 and 34667 keep their measures;
  - a boundary before the 21–22 October projections returns the earlier selection.

  Record the private evidence.
- [ ] 2.3 Update the Calcinaia guide (new dated entry following CLN-039), the operations guide on equivalent interpretations (measure selection and the conservative tracked-target union) and the changelog with the verified results.
