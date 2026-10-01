# Development acquisition and interpretation recovery

HTTP readiness verifies dependencies and listeners, not source collection or
interpretation progress. Select development explicitly with
`scripts/compose-env.sh development`. Keep database dumps, captures, policies and
call ledgers private under ignored `.local/operations/`.

## Unavailable municipal documents

Inspect `acquisition_source_status`, recent `acquisition_checks`,
`acquisition_targets` and `acquisition_unavailable_targets`. A document HTTP
404/410 allows later documents to be acquired, but keeps checks incomplete during
bounded retries and backoff.

Traverse every page of every configured section and retain dated evidence before
excluding a stale target. Confirm its URL remains unavailable, is absent from the
complete traversal, and has no explicit reference or unresolved/ongoing measure
requiring continued checking. Do not infer equivalence to another URL or
cancellation of a provision.

Record an `exclude_stale_unavailable` row in
`acquisition_target_dispositions`, scoped to the active source, configuration and
exact URL. Include actor, reason, evidence and observed `last_seen_at`. Use a
guarded transaction asserting the reviewed target has not changed and exactly
one decision is recorded. Preserve originals, unavailable-target records and
failed checks. Rediscovery after the reviewed `last_seen_at` automatically restores
checking; a later `resume` decision also restores it.

When no check is claimed, bring only that source's `next_check_at` forward and
let its normal worker run. Verify a fresh check has `complete=true`, an empty
error code and a new `last_complete_at`. Reachability alone is insufficient.

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
