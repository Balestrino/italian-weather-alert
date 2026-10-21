# Development acquisition and interpretation recovery

HTTP readiness verifies dependencies and listeners, not source collection or
interpretation progress. Select development explicitly with
`scripts/compose-env.sh development`. Keep database dumps, captures, policies and
call ledgers private under ignored `.local/operations/`.

## Unavailable municipal documents

Inspect `acquisition_source_status`, recent `acquisition_checks`, and the
unavailable targets of the source:

```sh
scripts/compose-env.sh development run --rm --no-deps --pull never admin source-unavailable-targets <source-id>
```

The same list appears under **URL non disponibili** on the source page in
administration. Each unrecovered 404/410 target is classified on every check:

- **withdrawn**: absent from the latest complete traversal of the configured
  listings, with no explicit reference and no `ongoing` or `unresolved`
  maintained measure state (see below). It is
  retried with the bounded backoff (1h, 2h, … up to 24h) but does not make the
  check incomplete. `acquisition_checks.withdrawn_unavailable_count` records how
  many were seen. A renamed notice needs no manual action.
- **blocking**: still listed, explicitly referenced, or with an `ongoing` or
  `unresolved` maintained measure state. It keeps checks incomplete during
  retries.
- **excluded**: covered by a reviewed `exclude_stale_unavailable` decision.

A later HTTP 200 recovers any target and retains a new version. Rediscovery in
a listing makes a still-unavailable target blocking again and lifts an earlier
exclusion. Classification never asserts equivalence to another URL or
cancellation of a provision.

Record a reviewed decision only for a blocking target that has stopped needing
checks, or to resume an excluded one. Traverse every page of every configured
section and retain dated evidence privately first. Confirm that the URL remains
unavailable and is absent from the complete traversal. Confirm also that any
derived measure no longer needs checking: excluding a measure-bearing target
requires an explicit acknowledgement. Use the administration form or:

```sh
scripts/compose-env.sh development run --rm -T --no-deps --pull never admin \
  source-unavailable-decide <source-id> <configuration> exclude_stale_unavailable <operator> < decision.json
```

`decision.json` holds `url`, the `observed_last_seen_at` printed by the list
command, `reason`, a non-empty `evidence` object and, when needed,
`acknowledge_measures`. Use `resume` instead of `exclude_stale_unavailable` to
return an excluded target to checking. The decision is rejected without changes
when the target was seen again, recovered, belongs to an inactive configuration
or was already decided. The error is a stable code such as `target_changed`,
`target_recovered` or `measure_acknowledgement_required`. Reload the list before
retrying. Decisions preserve originals, unavailable-target records and failed
checks, and never start a check or change publication.

To verify a decision, bring forward only that source's `next_check_at` when no
check is claimed, and let its normal worker run. A fresh check must have
`complete=true`, an empty error code and a new `last_complete_at`.
Reachability alone is insufficient.

## Followed notices and maintained measure state

At the start of every scheduled check of a listing-driven municipal source, the
measure, relevance and reference state of each tracked notice is derived again
from validated interpretation: projected municipal measures, their validity,
applied linked updates and the latest classification. No model call, crawl or
re-interpretation is involved.

| Measure state | Meaning | Rechecked beyond the revision window |
| --- | --- | --- |
| `ongoing` | A measure has a supported end that has not passed | Yes, until the end passes |
| `unresolved` | A measure has an unknown, start-only or conditional end and the notice is within the limit | Yes |
| `aged_unresolved` | As `unresolved`, but older than the limit | No |
| `ceased` | Every measure ended or was cancelled by a linked later measure | No |
| `none` | No derived measure | No |

Partial reopenings and extensions never end a measure by themselves. The limit
is `discovery.unresolved_measure_days` (default 90, valid 1–366); changing it
needs an ordinary configuration revision. A notice is marked explicitly
referenced when a linked later measure from another document targets one of its
measures; this keeps it in bootstrap plans and protects it from withdrawal, and
is never cleared automatically.

List followed notices, with the reason each is or is no longer rechecked:

```sh
scripts/compose-env.sh development run --rm --no-deps --pull never admin source-tracked-targets <source-id>
```

The same list appears under **Documenti seguiti** on the source page. Both read
the state derived by the last check. A failed derivation makes the check
incomplete with `interpretation_state_failed` and keeps the previous states.
Excluding a notice that ever produced a measure, even one `ceased` or
`aged_unresolved`, still requires the operator acknowledgement.

## Temporary interpretation recovery limits

Provider gates and `processing_recovery_limits` are separate controls. A closed
provider gate permits requests, while a persistent recovery restriction can still
defer ordinary jobs. Successful incident jobs do not remove that restriction.

Stop the development worker cleanly before changing the policy. Preserve a
verified private database dump, the recovery row, call ledger and relevant
job/attempt state. Confirm no inference jobs remain running. Inspect incident
results and diagnose unresolved provider or validation failures before expanding
the scope.

Use a bounded canary with selected versions, a recorded start time and a call
cap. Carry forward calls already spent when replacing its policy. Keep retry
budgets, archives, equivalence/reuse rules and provider holds intact. Include
relevant and irrelevant notices; exercise extraction and embedding/linking
dependencies. Child jobs inherit document scope through their extraction run.
Inspect results and run attempts: a successful queue job can still contain an
`uninterpreted` extraction result.

Rejected classification and extraction responses are captured privately in
`processing_invalid_outputs` with their exact request/response, window, attempt,
finish reason and diagnostic code. These are not successful checkpoints. Capture
failure stops processing with an explicit error. Enable source-scoped output
corrections after regression checks. Use the private `reprocess-interpretation`
action with narrow acquisition bounds for selected uninterpreted versions,
preserving earlier runs. Reprocessing run keys retain all identity components in
a deterministic digest when they exceed the store's identifier bound.

Once the representative chain succeeds, preserve the final canary policy and
ledger and explicitly remove the obsolete recovery row in a guarded transaction.
Ordinary admission resumes under existing retry, archive and provider controls;
failed/archived jobs are not relaunched and public sources are not enabled.
Verify fresh successes, queue progress and worker stability without new
authentication/quota rejection. Finish with:

```sh
python3 scripts/check-deployment-config.py --environment development --runtime
python3 scripts/smoke-compose.py development
```

## Provider entitlement failures

An HTTP 402 quota rejection can report an expired trial rather than a transient
capacity error. Inspect the durable call's `provider_code` and the credential-wide
`*` gate. Keep a quota/authentication hold in place until the provider account or
credential has been restored; application restarts and repeated resumes cannot
repair that entitlement.

After entitlement restoration, use the private `/admin/inference/resume` action
with the recorded scope, model `*` and an operator actor. This reopens each model
for one recovery probe. Relaunch only the reviewed failed job through its private
`/admin/jobs/{id}/relaunch` action with the current `after_attempt`; preserve the
failed receipt and earlier attempts. Verify that OCR, classification, extraction
and their dependencies progress without a new rejection. Collection can remain
healthy while interpretation waits for the provider.

Production activation, source acceptance and longer observation gates remain
separate from development recovery.

### Verifica del recupero provider — 4 ottobre 2026

Il task 13.8 è verificato con una sonda Qwen autenticata, ripresa privata dello
scope credenziale e una sola sonda di recupero ammessa per modello. La credenziale
esistente è tornata utilizzabile; nessun credito, piano o nuovo secret è stato
dedotto dalla risposta. I gate Qwen/OCR e quello globale risultano chiusi dopo
le chiamate riuscite. Gli embedding restano disabilitati e non sono collaudati
da questa verifica.

Il job OCR originariamente fallito era già stato rilanciato e completato dal
fallback locale il 1 ottobre. La ripetizione dello stesso comando di rilancio è
idempotente: conserva quel successo e gli errori precedenti. Un canary selezionato
sullo stesso documento e sulla stessa risorsa ha poi completato il run OCR remoto
prima fallito, con pagine e ricevute persistite; il limite del canary era un solo
tentativo. Non è stato riscritto lo stato del job storico.

Un primo caso ordinario ha prodotto chiamate valide ma una citazione non conforme,
rifiutata dalla validazione. Il fallimento rimane registrato e non costituisce un
nuovo rifiuto quota. Un secondo caso municipale ha completato classificazione,
estrazione e linking tramite i worker ordinari, con chiamate remote riuscite e
senza nuovi rifiuti. Questa verifica delimitata completa il recupero del provider;
non certifica tutte le interpretazioni o la suite di accettazione.

Sonde, comandi/ricevute di ripresa e rilancio, canary, risultati OCR, tentativi,
ledger e stato finale sono nell'archivio privato `PROVIDER-RECOVERY-20261004`.
I servizi esistenti e i controlli di fonte/pubblicazione sono stati preservati.
