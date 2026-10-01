package classification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/processing"
)

type fakeDocuments struct {
	version documents.Version
	bodies  map[string][]byte
}

func (f *fakeDocuments) Version(context.Context, int64) (documents.Version, error) {
	return f.version, nil
}
func (f *fakeDocuments) Read(_ context.Context, _ int64, raw string) ([]byte, error) {
	body, ok := f.bodies[raw]
	if !ok {
		return nil, errors.New("missing fixture")
	}
	return body, nil
}

type fakeOCR struct {
	pages     []ocr.PageResult
	resources []ocr.ResourceResult
}

func (f fakeOCR) PagesForVersion(context.Context, int64) ([]ocr.PageResult, error) {
	return f.pages, nil
}
func (f fakeOCR) ResourcesForVersion(context.Context, int64) ([]ocr.ResourceResult, error) {
	return f.resources, nil
}

type fakeProcessing struct {
	invalidOutputs    []processing.InvalidOutput
	captureErr        error
	lastConfiguration string
	run               processing.Run
	attempts          int
	finished          []processing.AttemptFinish
}

func (f *fakeProcessing) RecordInvalidOutput(_ context.Context, v processing.InvalidOutput) error {
	if f.captureErr != nil {
		return f.captureErr
	}
	f.invalidOutputs = append(f.invalidOutputs, v)
	return nil
}
func (f *fakeProcessing) StartRun(_ context.Context, request processing.RunRequest) (processing.Run, error) {
	if len(request.IdempotencyKey) > 200 {
		return processing.Run{}, processing.ErrInvalid
	}
	f.lastConfiguration = request.ConfigurationVersionID
	if request.Stage != "classification" || request.DocumentVersionID == nil {
		return processing.Run{}, errors.New("bad run")
	}
	if f.run.ID == 0 {
		f.run = processing.Run{ID: 51}
	}
	return f.run, nil
}
func (f *fakeProcessing) StartAttempt(_ context.Context, start processing.AttemptStart) (processing.Attempt, error) {
	f.attempts++
	return processing.Attempt{RunID: start.RunID, Number: f.attempts, StartedAt: start.StartedAt}, nil
}
func (f *fakeProcessing) FinishAttempt(_ context.Context, finish processing.AttemptFinish) (processing.Attempt, error) {
	f.finished = append(f.finished, finish)
	return processing.Attempt{RunID: finish.RunID, Number: finish.Number}, nil
}

type memoryResults struct{ value *Result }

func (m *memoryResults) Put(_ context.Context, value Result) error { m.value = &value; return nil }
func (m *memoryResults) Get(context.Context, int64) (Result, bool, error) {
	if m.value == nil {
		return Result{}, false, nil
	}
	return *m.value, true, nil
}

type fakeAdapter struct {
	response inference.Response
	requests []inference.Request
	complete func(inference.Request) inference.Response
}

func (f *fakeAdapter) Name() string { return "fixture" }
func (f *fakeAdapter) Complete(_ context.Context, request inference.Request) (inference.Response, error) {
	f.requests = append(f.requests, request)
	if f.complete != nil {
		return f.complete(request), nil
	}
	return f.response, nil
}

func int64Pointer(value int64) *int64 { return &value }

func runFixture(t *testing.T, html string, metadata json.RawMessage, response string) (*memoryResults, *fakeAdapter, *fakeProcessing) {
	t.Helper()
	url := "https://comune.example/notizia"
	documentsStore := &fakeDocuments{version: documents.Version{ID: 8, Complete: true, Metadata: metadata, Resources: []documents.Reference{{URL: url, Role: "original", Required: true, SourceID: "calcinaia", Configuration: 1, MediaType: "text/html", Hash: strings.Repeat("b", 64)}}}, bodies: map[string][]byte{url: []byte(html)}}
	adapter := &fakeAdapter{response: inference.Response{ID: "qwen-response", Model: "qwen3.8-27b", Content: response, Usage: inference.Usage{InputTokens: int64Pointer(100), OutputTokens: int64Pointer(30)}}}
	results, process := &memoryResults{}, &fakeProcessing{}
	now := time.Date(2026, 9, 17, 14, 0, 0, 0, time.UTC)
	runner := &Runner{Documents: documentsStore, OCR: fakeOCR{}, Processing: process, Results: results, Adapter: adapter, Model: "qwen3.8-27b", ConfigurationVersion: "classification-config", DisableThinking: true, Now: func() time.Time { return now }}
	payload, _ := json.Marshal(Payload{DocumentVersionID: 8, Workload: "evaluation"})
	if _, err := runner.Handler()(context.Background(), jobs.Job{ID: 70, Attempt: 1, Payload: payload}); err != nil {
		t.Fatal(err)
	}
	return results, adapter, process
}

func TestFullContentClassificationIgnoresEmptySourceCategory(t *testing.T) {
	html := `<html><head><style>.hidden{}</style><script>ignore me</script></head><body><h1>Ordinanza per allerta meteo</h1><p>Dalle ore 18 chiusi i cimiteri comunali per rischio idraulico.</p></body></html>`
	results, adapter, process := runFixture(t, html, json.RawMessage(`{"source_category":""}`), `{"relevant":true,"reason_code":"local_weather_measure","evidence_quote":"chiusi i cimiteri comunali per rischio idraulico"}`)
	if results.value == nil || results.value.Relevant == nil || !*results.value.Relevant || results.value.ReasonCode != "local_weather_measure" || !results.value.ContentComplete {
		t.Fatalf("relevant content with empty category was lost: %#v", results.value)
	}
	if len(adapter.requests) != 1 || len(adapter.requests[0].Messages) != 2 || !strings.Contains(string(adapter.requests[0].Messages[1].Content), "Dalle ore 18") || strings.Contains(string(adapter.requests[0].Messages[1].Content), "ignore me") || strings.Contains(string(adapter.requests[0].Messages[1].Content), "source_category") {
		t.Fatal("full visible content contract or category independence lost")
	}
	if adapter.requests[0].ChatTemplateKwargs == nil || adapter.requests[0].ChatTemplateKwargs.EnableThinking {
		t.Fatal("evaluated Qwen non-thinking configuration was not applied")
	}
	var userContent string
	if json.Unmarshal(adapter.requests[0].Messages[1].Content, &userContent) != nil || !json.Valid([]byte(userContent)) {
		t.Fatal("classification content is not an OpenAI-compatible JSON string")
	}
	if len(process.finished) != 1 || process.finished[0].Usage.Status != "reported" || *process.finished[0].Usage.InputTokens != 100 {
		t.Fatalf("classification usage not recorded: %#v", process.finished)
	}
}

func TestUnrelatedAndMisleadingKeywordsRemainNotRelevant(t *testing.T) {
	cases := []struct{ name, text, quote string }{
		{"ordinary closure", "Il Museo Coccapani resterà chiuso lunedì per riallestimento della mostra.", "chiuso lunedì per riallestimento"},
		{"misleading keyword", "La biblioteca presenta il romanzo Allerta meteo. Ingresso libero e nessuna modifica ai servizi.", "presenta il romanzo Allerta meteo"},
		{"sports", "Scacchi in riva all'Arno: iscrizioni aperte per il torneo di domenica.", "iscrizioni aperte per il torneo"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			body := "<main><p>" + test.text + "</p></main>"
			response, _ := json.Marshal(Decision{Relevant: false, ReasonCode: "not_relevant", EvidenceQuote: test.quote})
			results, _, _ := runFixture(t, body, json.RawMessage(`{}`), string(response))
			if results.value.Relevant == nil || *results.value.Relevant || results.value.ReasonCode != "not_relevant" {
				t.Fatalf("unrelated content classified relevant: %#v", results.value)
			}
		})
	}
}

func TestMissingRequiredOCRContentIsUndeterminedWithoutModelCall(t *testing.T) {
	url, attachment := "https://comune.example/notizia", "https://comune.example/ordinanza.pdf"
	documentsStore := &fakeDocuments{version: documents.Version{ID: 9, Complete: false, Resources: []documents.Reference{
		{URL: url, Role: "original", Required: true, SourceID: "calcinaia", Configuration: 1, MediaType: "text/html", Hash: strings.Repeat("c", 64)},
		{URL: attachment, Role: "attachment", Required: true, SourceID: "calcinaia", Configuration: 1, MediaType: "", Missing: "unavailable"},
	}}, bodies: map[string][]byte{url: []byte(`<main>Avviso che collega una ordinanza non disponibile.</main>`)}}
	adapter, results, process := &fakeAdapter{}, &memoryResults{}, &fakeProcessing{}
	now := time.Date(2026, 9, 17, 14, 0, 0, 0, time.UTC)
	runner := &Runner{Documents: documentsStore, OCR: fakeOCR{}, Processing: process, Results: results, Adapter: adapter, Model: "qwen3.8-27b", ConfigurationVersion: "classification-config", Now: func() time.Time { return now }}
	payload, _ := json.Marshal(Payload{DocumentVersionID: 9, Workload: "ordinary"})
	if _, err := runner.Handler()(context.Background(), jobs.Job{ID: 71, Attempt: 1, Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if results.value.Status != "undetermined" || results.value.ReasonCode != "incomplete_required_content" || len(adapter.requests) != 0 || process.finished[0].Usage.Status != "not_applicable" {
		t.Fatalf("missing content was inferred or billed: %#v", results.value)
	}
}

func TestDecisionRequiresConsistentLiteralEvidence(t *testing.T) {
	content := fullContent{Complete: true, Text: "chiusura dei cimiteri per rischio idraulico", Sections: []contentSection{{Text: "chiusura dei cimiteri per rischio idraulico"}}}
	for _, raw := range []string{
		`{"relevant":true,"reason_code":"not_relevant","evidence_quote":"chiusura dei cimiteri"}`,
		`{"relevant":true,"reason_code":"local_weather_measure","evidence_quote":"chiusura di tutte le scuole"}`,
		`{"relevant":true,"reason_code":"local_weather_measure","evidence_quote":"chiusura dei cimiteri","extra":1}`,
	} {
		if _, err := ParseDecision(raw, content); !errors.Is(err, ErrInvalid) {
			t.Fatalf("invalid evidence accepted: %s", raw)
		}
	}
}

func TestExplicitOperativeFormulaOverridesFalseNegative(t *testing.T) {
	content := Content{Complete: true, Text: "Premessa. ORDINA in via contingibile e urgente la chiusura dei cimiteri."}
	raw := `{"relevant":false,"reason_code":"not_relevant","evidence_quote":"Premessa"}`
	decision, err := ParseDecision(raw, content)
	if err != nil || !decision.Relevant || decision.ReasonCode != "local_weather_measure" || decision.EvidenceQuote != "ORDINA in via contingibile e urgente" {
		t.Fatalf("explicit operative formula did not override false negative: %#v %v", decision, err)
	}
}

func TestWeatherWordsAloneDoNotOverrideFalseNegative(t *testing.T) {
	content := Content{Complete: true, Text: "Mostra fotografica sul maltempo nei giardini pubblici."}
	raw := `{"relevant":false,"reason_code":"not_relevant","evidence_quote":"Mostra fotografica"}`
	decision, err := ParseDecision(raw, content)
	if err != nil || decision.Relevant {
		t.Fatalf("non-operative weather words overrode negative decision: %#v %v", decision, err)
	}
}

func TestLongDocumentClassificationEvaluatesEveryBoundedSegment(t *testing.T) {
	url := "https://comune.example/notizia-lunga"
	text := strings.Repeat("comunicazione amministrativa ordinaria ", 800) + "Per rischio idraulico il Comune dispone la chiusura del ponte."
	documentsStore := &fakeDocuments{version: documents.Version{ID: 18, Complete: true, Resources: []documents.Reference{{URL: url, Role: "original", Required: true, SourceID: "calcinaia", Configuration: 1, MediaType: "text/html", Hash: strings.Repeat("d", 64)}}}, bodies: map[string][]byte{url: []byte("<main>" + text + "</main>")}}
	adapter := &fakeAdapter{}
	adapter.complete = func(request inference.Request) inference.Response {
		content := string(request.Messages[1].Content)
		decision := `{"relevant":false,"reason_code":"not_relevant","evidence_quote":"comunicazione amministrativa ordinaria"}`
		if strings.Contains(content, "rischio idraulico") {
			decision = `{"relevant":true,"reason_code":"local_weather_measure","evidence_quote":"dispone la chiusura del ponte"}`
		}
		return inference.Response{ID: "segment-response", Model: "qwen3.8-27b", Content: decision, Usage: inference.Usage{InputTokens: int64Pointer(50), OutputTokens: int64Pointer(10)}}
	}
	results, process := &memoryResults{}, &fakeProcessing{}
	runner := &Runner{Documents: documentsStore, OCR: fakeOCR{}, Processing: process, Results: results, Adapter: adapter, Model: "qwen3.8-27b", ConfigurationVersion: "classification-segmented", Now: func() time.Time { return time.Date(2026, 9, 17, 14, 0, 0, 0, time.UTC) }}
	payload, _ := json.Marshal(Payload{DocumentVersionID: 18, Workload: "evaluation"})
	if _, err := runner.Handler()(context.Background(), jobs.Job{ID: 78, Attempt: 1, Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if results.value == nil || results.value.Relevant == nil || !*results.value.Relevant || len(results.value.Segments) < 2 || len(adapter.requests) != len(results.value.Segments) {
		t.Fatalf("full segmented classification lost late relevant content: result=%#v calls=%d", results.value, len(adapter.requests))
	}
	for index, segment := range results.value.Segments {
		if segment.Ordinal != index+1 || segment.Total != len(results.value.Segments) || segment.ResourceURL != url || segment.EndByte-segment.StartByte > MaxSegmentTextBytes {
			t.Fatalf("invalid retained segment reference: %#v", segment)
		}
		if adapter.requests[index].MaxCompletionTokens != 4096 {
			t.Fatal("segmentation changed timeout/output budget instead of bounding each request")
		}
	}
	wantInput := int64(50 * len(results.value.Segments))
	if results.value.InputTokens == nil || *results.value.InputTokens != wantInput || process.finished[0].Usage.InputTokens == nil || *process.finished[0].Usage.InputTokens != wantInput {
		t.Fatalf("segmented usage was not aggregated: %#v %#v", results.value, process.finished)
	}
}

func TestInvalidQuotationCapturedWithoutAcceptingOrRetrying(t *testing.T) {
	for _, brokenStorage := range []bool{false, true} {
		t.Run(fmt.Sprint(brokenStorage), func(t *testing.T) {
			url := "https://comune.example/festa"
			doc := &fakeDocuments{version: documents.Version{ID: 8, Complete: true, Resources: []documents.Reference{{URL: url, Role: "original", Required: true, MediaType: "text/html"}}}, bodies: map[string][]byte{url: []byte("<p>Una festa di quartiere.</p>")}}
			raw := `{"relevant":false,"reason_code":"not_relevant","evidence_quote":"Una frase non presente"}`
			adapter := &fakeAdapter{response: inference.Response{Content: raw, FinishReason: "stop", Usage: inference.Usage{InputTokens: int64Pointer(17), OutputTokens: int64Pointer(8)}}}
			process := &fakeProcessing{}
			if brokenStorage {
				process.captureErr = errors.New("storage unavailable")
			}
			results := &memoryResults{}
			runner := &Runner{Documents: doc, OCR: fakeOCR{}, Processing: process, Results: results, Adapter: adapter, Model: "fixture", ConfigurationVersion: "fixture"}
			payload, _ := json.Marshal(Payload{DocumentVersionID: 8, Workload: "ordinary"})
			_, err := runner.Handler()(context.Background(), jobs.Job{ID: 70, Attempt: 1, Payload: payload})
			var he *jobs.HandlerError
			if !errors.As(err, &he) || he.Failure.Temporary || len(adapter.requests) != 1 || results.value != nil {
				t.Fatal("invalid result accepted or retried", err)
			}
			if len(process.finished) != 1 || *process.finished[0].Usage.InputTokens != 17 {
				t.Fatal("lost billed usage")
			}
			if brokenStorage {
				if he.Failure.Code != "classification_diagnostic_unavailable" {
					t.Fatal("capture failure treated as document-only", err)
				}
			} else {
				if he.Failure.Code != "classification_output_quotation" || len(process.invalidOutputs) != 1 {
					t.Fatal("invalid quotation not retained", err)
				}
				capture := process.invalidOutputs[0]
				req, _ := json.Marshal(adapter.requests[0])
				if string(capture.Request) != string(req) || string(capture.Response) != raw || capture.RunID != 51 || capture.AttemptNumber != 1 || capture.SegmentOrdinal != 1 || capture.FinishReason != "stop" {
					t.Fatal("capture differs from actual evidence")
				}
			}
		})
	}
}
