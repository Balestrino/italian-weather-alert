package extraction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
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

type fakeClassifications struct{ result classification.Result }

func (f fakeClassifications) Get(context.Context, int64) (classification.Result, bool, error) {
	return f.result, true, nil
}

type fakeProcessing struct {
	lastConfiguration string
	run               processing.Run
	attempts          int
	finished          []processing.AttemptFinish
}

func (f *fakeProcessing) StartRun(_ context.Context, request processing.RunRequest) (processing.Run, error) {
	f.lastConfiguration = request.ConfigurationVersionID
	if request.Stage != "extraction" || request.DocumentVersionID == nil {
		return processing.Run{}, errors.New("bad run")
	}
	if f.run.ID == 0 {
		f.run = processing.Run{ID: 81}
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
	err      error
	requests []inference.Request
	complete func(inference.Request) inference.Response
}

func (f *fakeAdapter) Name() string { return "fixture" }
func (f *fakeAdapter) Complete(_ context.Context, request inference.Request) (inference.Response, error) {
	f.requests = append(f.requests, request)
	if f.complete != nil {
		return f.complete(request), nil
	}
	return f.response, f.err
}

func int64Pointer(value int64) *int64 { return &value }

func fixtureRunner(t *testing.T, complete bool, response string) (*Runner, *memoryResults, *fakeAdapter, *fakeProcessing) {
	t.Helper()
	url := "https://comune.example/notizia"
	docs := &fakeDocuments{version: documents.Version{ID: 12, Complete: complete, Resources: []documents.Reference{{URL: url, Role: "original", Required: true, SourceID: "calcinaia", Configuration: 1, MediaType: "text/html", Hash: strings.Repeat("a", 64)}}}, bodies: map[string][]byte{url: []byte(`<main>Il Comune dispone la chiusura del sottopasso di via Maremmana fino al perdurare dell'emergenza.</main>`)}}
	relevant := true
	classes := fakeClassifications{classification.Result{RunID: 61, DocumentVersionID: 12, Status: "classified", Relevant: &relevant}}
	adapter := &fakeAdapter{response: inference.Response{ID: "qwen-extraction", Model: "qwen3.8-27b", Content: response, Usage: inference.Usage{InputTokens: int64Pointer(150), OutputTokens: int64Pointer(80)}}}
	results, process := &memoryResults{}, &fakeProcessing{}
	now := time.Date(2026, 9, 17, 16, 0, 0, 0, time.UTC)
	runner := &Runner{Documents: docs, OCR: fakeOCR{}, Classifications: classes, Processing: process, Results: results, Adapter: adapter, Model: "qwen3.8-27b", ConfigurationVersion: "extraction-config", DisableThinking: true, Now: func() time.Time { return now }}
	return runner, results, adapter, process
}

func runJob(t *testing.T, runner *Runner) (jobs.Result, error) {
	t.Helper()
	payload, _ := json.Marshal(Payload{DocumentVersionID: 12, ClassificationRunID: 61, Workload: "evaluation"})
	return runner.Handler()(context.Background(), jobs.Job{ID: 90, Attempt: 1, MaxAttempts: 3, Payload: payload})
}

func TestRunnerPersistsSupportedMeasureAndUsage(t *testing.T) {
	response := `{"envelope_version":"compact-evidence-v1","window_ordinal":1,"evidence":["dispone la chiusura","sottopasso di via Maremmana","via Maremmana","fino al perdurare dell'emergenza"],"measures":[{"kind":"closure","subject":"sottopasso di via Maremmana","place":"via Maremmana","valid_from":null,"valid_until":"fino al perdurare dell'emergenza","evidence_refs":{"kind":[0],"subject":[1],"place":[2],"valid_from":[],"valid_until":[3]}}]}`
	runner, results, adapter, process := fixtureRunner(t, true, response)
	output, err := runJob(t, runner)
	if err != nil {
		t.Fatal(err)
	}
	if results.value == nil || results.value.Status != "extracted" || len(results.value.Measures) != 1 || results.value.Measures[0].ValidFrom != nil || len(adapter.requests) != 1 {
		t.Fatalf("supported extraction lost: %#v", results.value)
	}
	if adapter.requests[0].ChatTemplateKwargs == nil || adapter.requests[0].ChatTemplateKwargs.EnableThinking {
		t.Fatal("evaluated Qwen non-thinking configuration was not applied")
	}
	if len(output.Effects) != 1 || output.Effects[0].Kind != "extraction_ready" || len(process.finished) != 1 || process.finished[0].Usage.Status != "reported" {
		t.Fatalf("extraction completion/accounting lost: %#v %#v", output, process.finished)
	}
	version, _ := runner.Documents.Version(context.Background(), 12)
	content, e := classification.GatherContent(context.Background(), runner.Documents, runner.OCR, version)
	if e != nil {
		t.Fatal(e)
	}
	windows, e := BuildOperationalWindows(13, content)
	if e != nil {
		t.Fatal(e)
	}
	base := Result{RunID: 99, DocumentVersionID: 13, ClassificationRunID: 62, CreatedAt: runner.Now()}
	mapped, ok := RemapResult(*results.value, windows, content, base)
	if !ok || mapped.DocumentVersionID != 13 || mapped.InputTokens != nil {
		t.Fatal("valid mapping lost")
	}
	windows[0].SourceEndByte++
	if _, ok = RemapResult(*results.value, windows, content, base); ok {
		t.Fatal("changed source offsets accepted")
	}
	windows[0].SourceEndByte--
	results.value.Measures[0].Evidence[0].Quote = "unsupported quotation"
	if _, ok = RemapResult(*results.value, windows, content, base); ok {
		t.Fatal("unsupported quotation reused")
	}

	truncated, truncatedResults, truncatedAdapter, _ := fixtureRunner(t, true, response)
	truncatedAdapter.response.FinishReason = "length"
	if _, e := runJob(t, truncated); e != nil {
		t.Fatal(e)
	}
	if truncatedResults.value.Status != "uninterpreted" || len(truncatedResults.value.Measures) != 0 || len(truncatedAdapter.requests) != 1 {
		t.Fatal("truncated output accepted or repair call sent")
	}

}

func TestInvalidEvidenceRemainsAcquiredButUninterpreted(t *testing.T) {
	response := `{"envelope_version":"compact-evidence-v1","window_ordinal":1,"evidence":["riapertura","tutte le strade"],"measures":[{"kind":"reopening","subject":"tutte le strade","place":null,"valid_from":null,"valid_until":null,"evidence_refs":{"kind":[0],"subject":[1],"place":[],"valid_from":[],"valid_until":[]}}]}`
	runner, results, _, process := fixtureRunner(t, true, response)
	output, err := runJob(t, runner)
	if err != nil {
		t.Fatal(err)
	}
	if results.value == nil || results.value.Status != "uninterpreted" || results.value.ReasonCode != "extraction_evidence_invalid" || len(results.value.Measures) != 0 {
		t.Fatalf("unsupported output was not isolated: %#v", results.value)
	}
	if output.Effects[0].Kind != "extraction_uninterpreted" || process.finished[0].Outcome != "failed" {
		t.Fatalf("uninterpreted visibility or failed attempt lost: %#v %#v", output, process.finished)
	}
}

func TestIncompleteRequiredContentDoesNotCallModel(t *testing.T) {
	runner, results, adapter, process := fixtureRunner(t, false, `{}`)
	output, err := runJob(t, runner)
	if err != nil {
		t.Fatal(err)
	}
	if results.value.Status != "uninterpreted" || results.value.ReasonCode != "incomplete_required_content" || len(adapter.requests) != 0 || process.finished[0].Usage.Status != "not_applicable" || output.Effects[0].Kind != "extraction_uninterpreted" {
		t.Fatalf("incomplete content was inferred or hidden: %#v", results.value)
	}
}

func TestRunnerSegmentsLongInputAndMergesEvidence(t *testing.T) {
	runner, results, adapter, process := fixtureRunner(t, true, `{}`)
	docs := runner.Documents.(*fakeDocuments)
	url := "https://comune.example/notizia"
	docs.bodies[url] = []byte("<main>Il Comune dispone la chiusura del ponte Verde. " + strings.Repeat("aggiornamento amministrativo senza nuove misure ", 800) + "Il Comune dispone la riapertura del sottopasso Blu.</main>")
	adapter.complete = func(request inference.Request) inference.Response {
		input := string(request.Messages[1].Content)
		var encoded string
		_ = json.Unmarshal(request.Messages[1].Content, &encoded)
		var envelope struct {
			Window struct {
				Ordinal int `json:"ordinal"`
			} `json:"window"`
		}
		_ = json.Unmarshal([]byte(encoded), &envelope)
		response := fmt.Sprintf(`{"envelope_version":"compact-evidence-v1","window_ordinal":%d,"evidence":[],"measures":[]}`, envelope.Window.Ordinal)
		if strings.Contains(input, "chiusura del ponte Verde") {
			response = fmt.Sprintf(`{"envelope_version":"compact-evidence-v1","window_ordinal":%d,"evidence":["dispone la chiusura","ponte Verde"],"measures":[{"kind":"closure","subject":"ponte Verde","place":null,"valid_from":null,"valid_until":null,"evidence_refs":{"kind":[0],"subject":[1],"place":[],"valid_from":[],"valid_until":[]}}]}`, envelope.Window.Ordinal)
		} else if strings.Contains(input, "riapertura del sottopasso Blu") {
			response = fmt.Sprintf(`{"envelope_version":"compact-evidence-v1","window_ordinal":%d,"evidence":["dispone la riapertura","sottopasso Blu"],"measures":[{"kind":"reopening","subject":"sottopasso Blu","place":null,"valid_from":null,"valid_until":null,"evidence_refs":{"kind":[0],"subject":[1],"place":[],"valid_from":[],"valid_until":[]}}]}`, envelope.Window.Ordinal)
		}
		return inference.Response{ID: "segmented-extraction", Model: "qwen3.8-27b", Content: response, Usage: inference.Usage{InputTokens: int64Pointer(60), OutputTokens: int64Pointer(20)}}
	}
	if _, err := runJob(t, runner); err != nil {
		t.Fatal(err)
	}
	if results.value == nil || results.value.Status != "extracted" || len(results.value.Measures) != 2 || len(results.value.Segments) < 2 || len(adapter.requests) != len(results.value.Segments) {
		t.Fatalf("segmented extraction/merge failed: %#v calls=%d", results.value, len(adapter.requests))
	}
	for _, measure := range results.value.Measures {
		for _, evidence := range measure.Evidence {
			if evidence.SegmentOrdinal < 1 || evidence.SegmentOrdinal > len(results.value.Segments) {
				t.Fatalf("evidence lost segment provenance: %#v", evidence)
			}
		}
	}
	wantInput := int64(60 * len(results.value.Segments))
	if results.value.InputTokens == nil || *results.value.InputTokens != wantInput || process.finished[0].Usage.InputTokens == nil || *process.finished[0].Usage.InputTokens != wantInput {
		t.Fatalf("segmented extraction usage not aggregated: %#v %#v", results.value, process.finished)
	}
}

func TestRunnerPersistsReversibleOCRWindowCoordinates(t *testing.T) {
	response := `{"envelope_version":"compact-evidence-v1","window_ordinal":1,"evidence":["ORDINA la chiusura","ponte Verde","dalle ore 18"],"measures":[{"kind":"closure","subject":"ponte Verde","place":null,"valid_from":"dalle ore 18","valid_until":null,"evidence_refs":{"kind":[0],"subject":[1],"place":[],"valid_from":[2],"valid_until":[]}}]}`
	runner, results, _, _ := fixtureRunner(t, true, response)
	docs := runner.Documents.(*fakeDocuments)
	url := "https://comune.example/ordinanza.pdf"
	docs.version.Resources = []documents.Reference{{URL: url, Role: "attachment", Required: true, SourceID: "calcinaia", Configuration: 1, MediaType: "application/pdf", Hash: strings.Repeat("b", 64)}}
	rawOCR := "## **ORDINA**\nla chiusura del ponte Verde dalle ore 18"
	runner.OCR = fakeOCR{
		pages:     []ocr.PageResult{{RunID: 7, DocumentVersionID: 12, ResourceURL: url, PageNumber: 2, Status: "complete", ExtractedText: rawOCR}},
		resources: []ocr.ResourceResult{{RunID: 7, DocumentVersionID: 12, ResourceURL: url, Status: "complete", PageCount: 1}},
	}
	if _, err := runJob(t, runner); err != nil {
		t.Fatal(err)
	}
	if results.value == nil || results.value.Status != "extracted" || len(results.value.Segments) != 1 {
		t.Fatalf("OCR extraction unavailable: %#v", results.value)
	}
	window := results.value.Segments[0]
	if window.Page != 2 || window.SourceEndByte <= window.SourceStartByte || len(window.NormalizationMap) == 0 || !validWindowOffsetMap(window.Page, window.StartByte, window.EndByte, window.SourceStartByte, window.SourceEndByte, window.NormalizationMap) {
		t.Fatalf("retained OCR coordinates were not persisted: %#v", window)
	}
	if rawOCR[window.SourceStartByte:window.SourceEndByte] != "ORDINA**\nla chiusura del ponte Verde dalle ore 18" {
		t.Fatalf("window does not resolve to verbatim retained OCR: %q", rawOCR[window.SourceStartByte:window.SourceEndByte])
	}
}
