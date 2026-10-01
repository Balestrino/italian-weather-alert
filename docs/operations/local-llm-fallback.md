# Optional local Qwen fallback

The worker can select a separate local OpenAI-compatible chat configuration when
an eligible remote provider is held or its circuit is cooling down. It is disabled
by default and requires an explicit source list. A healthy remote provider, or an
available recovery probe, keeps preference. Authentication, quota, availability,
transport and rate-limit holds qualify; invalid evidence, schema errors and
storage failures do not initiate fallback.

## Evaluate the local server

Use a llama.cpp server with a Qwen3.5 chat model and structured JSON support.
Disable thinking for these tasks. For OCR, load the matching multimodal projector
and enable image input. Verify `/v1/models` and `/props`, then run the opt-in suite:

```sh
IWA_LOCAL_LLM_ENDPOINT=http://127.0.0.1:8081/v1/chat/completions \
IWA_LOCAL_LLM_MODEL=qwen3.5-9b-q8_0 \
go test -tags=llm ./internal/backend/inference -run TestLocalLLM -count=1 -v
```

Use a reachable endpoint and the model identifier accepted by your server. These
are real network calls using original synthetic Italian notices, production
prompts and strict evidence parsers. Ordinary `go test ./...` makes no such calls;
the live suite skips when endpoint/model variables are absent. The image test
checks vision capability and skips if unsupported; a skip does not establish OCR
acceptance. Optional `IWA_LOCAL_LLM_REVIEW_DIR` retains private request/response
captures; use an absolute path under ignored `.local/operations/`.

The local policy uses temperature 0, seed 42 and
`chat_template_kwargs.enable_thinking=false`. The configuration records this
policy and keeps a distinct prompt/model identity. Local classification tolerates
straight/typographic double-quote presentation while returning the exact contiguous
source quotation; invented or stitched evidence remains rejected. Local extraction
uses concise instructions for operative measures and explicit civil-protection
centre openings, independently versioned from the remote prompt. Sampling controls improve
repeatability; all output still has to pass the normal validators. Consult the
[llama.cpp server documentation](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md)
for model loading, multimodal projectors and OpenAI-compatible API behavior.

## Configure development

Put reviewed settings in ignored `.local/development.env`:

```dotenv
IWA_LOCAL_FALLBACK_ENABLED=true
IWA_LOCAL_FALLBACK_ENDPOINT=http://local-server:8081/v1/chat/completions
IWA_LOCAL_FALLBACK_MODEL=qwen3.5-9b-q8_0
IWA_LOCAL_FALLBACK_SOURCES=calcinaia-municipal
IWA_OUTPUT_FIX_SOURCES=calcinaia-municipal
IWA_LOCAL_FALLBACK_TIMEOUT_SECONDS=180
IWA_LOCAL_FALLBACK_OCR_ENABLED=false
```

The endpoint must be reachable from the worker container. Keep your existing
output-fix source selection when adding a source; fallback sources must use the
evaluated output contracts. Enable OCR only after the image test passes. The
endpoint must end in `/v1/chat/completions`; URL credentials and query parameters
are rejected. Source wildcards are rejected. Timeout accepts 1–600 seconds.

The `local-openai-chat` adapter has its own provider gate, immutable model and
stage configurations, request checkpoints and call receipts. It does not inherit
the remote API key or remote tariff. An unauthenticated server uses a non-secret
placeholder; an authenticated gateway can use `IWA_LOCAL_FALLBACK_API_KEY_FILE`
through a private Compose override that mounts and exposes the file to the worker.
Keep keys out of environment files and command arguments.

A first remote rejection preserves its failed run and schedules the local choice
on a remaining ordinary attempt. It does not extend retry budgets. Pending jobs
outside the source list, and jobs whose local provider is also held, remain
subject to pre-claim deferral. Equivalent copies can use a compatible primary or
local cache while their transports remain blocked. Archive, suspension, territory
and bounded recovery controls continue to apply.

Qwen chat/image support does not supply semantic embeddings. Embedding work keeps
its separately configured provider and gates. Linking can use the local chat
model when its required retrieval inputs are available; a held embedding provider
can still delay new semantic linking. Do not interpret local text/OCR success as
completion of those dependencies or public source acceptance.

## Deploy and roll back

Run the ordinary Go checks and the processing integration tests on disposable
PostgreSQL. Preserve a verified database dump, the previous image/settings and
private canary evidence. Build the development image and recreate the worker with
the selected image. Use the [acquisition recovery procedure](acquisition-recovery.md)
for a bounded representative canary. Inspect successful run attempts, strict
classification/extraction results, actual returned model, call receipts and
service stability; queue success alone is insufficient. Leave the remote account
hold in place until its entitlement is restored.

To disable routing, set `IWA_LOCAL_FALLBACK_ENABLED=false` and recreate the worker.
For code rollback, restore the prior development image and settings and recreate
the affected application services, retaining data volumes and additive records.
Earlier local runs and remote failed attempts remain available for audit. No
production deployment or public source activation is implied.
