package ocr

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
	"github.com/Balestrino/italian-weather-alert/internal/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/processing"
)

type fakeDocuments struct {
	version documents.Version
	body    []byte
}

func (f *fakeDocuments) Version(context.Context, int64) (documents.Version, error) {
	return f.version, nil
}
func (f *fakeDocuments) Read(context.Context, int64, string) ([]byte, error) { return f.body, nil }

type fakeRenderer struct {
	pages []PageImage
	err   error
}

func (f fakeRenderer) Render(context.Context, []byte) ([]PageImage, error) { return f.pages, f.err }

type memoryResults struct {
	pages    map[int]PageResult
	resource *ResourceResult
}

func (m *memoryResults) PutPage(_ context.Context, page PageResult) error {
	if m.pages == nil {
		m.pages = map[int]PageResult{}
	}
	m.pages[page.PageNumber] = page
	return nil
}
func (m *memoryResults) Page(_ context.Context, _ int64, page int) (PageResult, bool, error) {
	result, ok := m.pages[page]
	return result, ok, nil
}
func (m *memoryResults) PutResource(_ context.Context, result ResourceResult) error {
	m.resource = &result
	return nil
}
func (m *memoryResults) Resource(context.Context, int64) (ResourceResult, bool, error) {
	if m.resource == nil {
		return ResourceResult{}, false, nil
	}
	return *m.resource, true, nil
}

type fakeProcessing struct {
	run      processing.Run
	attempts int
	finishes []processing.AttemptFinish
}

func (f *fakeProcessing) StartRun(_ context.Context, request processing.RunRequest) (processing.Run, error) {
	if request.Stage != "ocr" || request.DocumentVersionID == nil {
		return processing.Run{}, errors.New("bad run")
	}
	if f.run.ID == 0 {
		f.run = processing.Run{ID: 41}
	}
	return f.run, nil
}
func (f *fakeProcessing) StartAttempt(_ context.Context, start processing.AttemptStart) (processing.Attempt, error) {
	f.attempts++
	return processing.Attempt{RunID: start.RunID, Number: f.attempts, StartedAt: start.StartedAt}, nil
}
func (f *fakeProcessing) FinishAttempt(_ context.Context, finish processing.AttemptFinish) (processing.Attempt, error) {
	f.finishes = append(f.finishes, finish)
	return processing.Attempt{RunID: finish.RunID, Number: finish.Number}, nil
}

type fakeAdapter struct {
	responses []inference.Response
	requests  []inference.Request
}

func (f *fakeAdapter) Name() string { return "fixture" }
func (f *fakeAdapter) Complete(_ context.Context, request inference.Request) (inference.Response, error) {
	f.requests = append(f.requests, request)
	response := f.responses[len(f.requests)-1]
	return response, nil
}

func pointer(value int64) *int64 { return &value }

func fixtureRunner(mediaType string, missing string, adapter *fakeAdapter, renderer Renderer) (*Runner, *memoryResults, *fakeProcessing) {
	now := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	results := &memoryResults{}
	process := &fakeProcessing{}
	version := documents.Version{ID: 7, Resources: []documents.Reference{{URL: "https://comune.example/ordinanza.pdf", Role: "attachment", SourceID: "calcinaia", Configuration: 1, MediaType: mediaType, Hash: strings.Repeat("a", 64), Missing: missing}}}
	return &Runner{Documents: &fakeDocuments{version: version, body: []byte("retained PDF")}, Processing: process, Results: results, Adapter: adapter, Renderer: renderer, Model: "deepseek-ocr-2", ConfigurationVersion: "ocr-config", Now: func() time.Time { return now }}, results, process
}

func TestScannedPDFIsRasterizedPerPageAndKeepsEvidence(t *testing.T) {
	adapter := &fakeAdapter{responses: []inference.Response{
		{ID: "response-1", Model: "deepseek-ocr-2", Content: "ORDINA dal 20 agosto 2026 alle ore 18.00", Usage: inference.Usage{InputTokens: pointer(10), OutputTokens: pointer(20)}},
		{ID: "response-2", Model: "deepseek-ocr-2", Content: "fino al perdurare dello stato di emergenza; area cani e cimiteri", Usage: inference.Usage{InputTokens: pointer(11), OutputTokens: pointer(21)}},
	}}
	renderer := fakeRenderer{pages: []PageImage{{Number: 1, MediaType: "image/png", Bytes: []byte("page one")}, {Number: 2, MediaType: "image/png", Bytes: []byte("page two")}}}
	runner, results, process := fixtureRunner("application/pdf", "", adapter, renderer)
	payload, _ := json.Marshal(Payload{DocumentVersionID: 7, ResourceURL: "https://comune.example/ordinanza.pdf", Workload: "evaluation"})
	if _, err := runner.Handler()(context.Background(), jobs.Job{ID: 9, Attempt: 1, Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if results.resource == nil || results.resource.Status != "complete" || results.resource.PageCount != 2 || len(results.pages) != 2 {
		t.Fatalf("page evidence not completed: %#v %#v", results.resource, results.pages)
	}
	if results.pages[1].DocumentVersionID != 7 || results.pages[1].ResourceURL == "" || results.pages[1].PageNumber != 1 || !strings.Contains(results.pages[1].ExtractedText, "20 agosto") {
		t.Fatalf("original/page/text evidence lost: %#v", results.pages[1])
	}
	for _, request := range adapter.requests {
		wire := string(request.Messages[0].Content)
		if strings.Contains(wire, "retained PDF") || !strings.Contains(wire, "data:image/png;base64,") || !strings.Contains(wire, PromptBody) {
			t.Fatal("raw PDF was sent or rasterized page contract was lost")
		}
	}
	if len(process.finishes) != 1 || process.finishes[0].Usage.Status != "reported" || *process.finishes[0].Usage.InputTokens != 21 || process.finishes[0].Usage.OtherUnits["requests"] != 2 {
		t.Fatalf("provider usage not aggregated: %#v", process.finishes)
	}
}

func TestUnreadablePageAndMissingAttachmentStayExplicit(t *testing.T) {
	t.Run("empty OCR output", func(t *testing.T) {
		adapter := &fakeAdapter{responses: []inference.Response{{ID: "empty", Model: "deepseek-ocr-2", Content: "   ", Usage: inference.Usage{InputTokens: pointer(4), OutputTokens: pointer(1)}}}}
		runner, results, _ := fixtureRunner("image/png", "", adapter, fakeRenderer{})
		payload, _ := json.Marshal(Payload{DocumentVersionID: 7, ResourceURL: "https://comune.example/ordinanza.pdf", Workload: "ordinary"})
		if _, err := runner.Handler()(context.Background(), jobs.Job{ID: 10, Attempt: 1, Payload: payload}); err != nil {
			t.Fatal(err)
		}
		if results.resource.Status != "unreadable" || results.resource.ErrorCode != "all_pages_unreadable" || results.pages[1].ExtractedText != "" {
			t.Fatalf("unreadable page invented content: %#v %#v", results.resource, results.pages[1])
		}
	})
	t.Run("missing attachment", func(t *testing.T) {
		adapter := &fakeAdapter{}
		runner, results, process := fixtureRunner("", "unavailable", adapter, fakeRenderer{})
		payload, _ := json.Marshal(Payload{DocumentVersionID: 7, ResourceURL: "https://comune.example/ordinanza.pdf", Workload: "ordinary"})
		if _, err := runner.Handler()(context.Background(), jobs.Job{ID: 11, Attempt: 1, Payload: payload}); err != nil {
			t.Fatal(err)
		}
		if results.resource.Status != "missing" || results.resource.ErrorCode != "attachment_unavailable" || results.resource.PageCount != 0 || len(adapter.requests) != 0 {
			t.Fatalf("missing attachment not preserved: %#v", results.resource)
		}
		if process.finishes[0].Usage.Status != "not_applicable" {
			t.Fatal("missing attachment was charged as a model call")
		}
	})
}
