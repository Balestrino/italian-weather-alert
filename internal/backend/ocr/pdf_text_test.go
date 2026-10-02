package ocr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

type fakeTextExtractor struct {
	pages []TextPage
	err   error
	calls int
}

func (f *fakeTextExtractor) Extract(context.Context, []byte) ([]TextPage, error) {
	f.calls++
	return f.pages, f.err
}

func TestNativePDFTextKeepsPageEvidenceWithoutCalls(t *testing.T) {
	adapter := &fakeAdapter{}
	runner, results, process := fixtureRunner("application/pdf", "", adapter, fakeRenderer{err: errors.New("renderer must not run")})
	reader := &fakeTextExtractor{pages: []TextPage{{1, "Il Comune dispone la chiusura del ponte per rischio idraulico."}, {2, "La disposizione resta valida fino alla conclusione dell'emergenza."}}}
	runner.TextExtractor = reader
	runner.LocalConfigurationVersion = "text-first-config"
	docs := runner.Documents.(*fakeDocuments)
	docs.version.Resources[0].URL = "https://comune.example/orders/ordinanza.pdf"
	docs.version.Resources[0].LocalProcessing = &registry.LocalProcessing{Version: registry.LocalProcessingVersion, TextPDFPathPrefixes: []string{"/orders/"}}
	payload, _ := json.Marshal(Payload{DocumentVersionID: 7, ResourceURL: docs.version.Resources[0].URL, Workload: "evaluation"})
	if _, err := runner.Handler()(context.Background(), jobs.Job{ID: 9, Attempt: 1, Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if len(adapter.requests) != 0 || len(results.pages) != 2 || results.resource.Status != "complete" || process.finishes[0].Usage.Status != "not_applicable" || process.finishes[0].PriceVersionID != nil {
		t.Fatal("native text consumed model/price", results, process.finishes)
	}
	for n, p := range results.pages {
		if p.PageNumber != n || p.DocumentVersionID != 7 || p.ResourceURL == "" || p.MediaType != "application/pdf" || len(p.InputSHA256) != 64 || p.ReturnedModel != PDFTextVersion || p.ProviderResponseID != "local:"+PDFTextVersion || p.InputTokens != nil {
			t.Fatal("page provenance", p)
		}
	}
	if _, err := runner.Handler()(context.Background(), jobs.Job{ID: 9, Attempt: 2, Payload: payload}); err != nil || reader.calls != 1 || process.attempts != 1 {
		t.Fatal("completed native job repeated", err)
	}
}

func TestNativePDFFallsBackForWholeResource(t *testing.T) {
	for _, tc := range []struct {
		name  string
		pages []TextPage
		err   error
		scope string
	}{
		{"empty page", []TextPage{{1, "Il Comune dispone la chiusura del ponte."}, {2, ""}}, nil, "/orders/"},
		{"missing page", []TextPage{{1, "Il Comune dispone la chiusura del ponte."}, {3, "Il Comune dispone la riapertura del ponte."}}, nil, "/orders/"},
		{"invalid Unicode", []TextPage{{1, "Il Comune dispone la chiusura \xff del ponte."}}, nil, "/orders/"},
		{"unavailable tools", nil, ErrNativeText, "/orders/"},
		{"graphical outside reviewed scope", nil, nil, "/text-only/"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			adapter := &fakeAdapter{responses: []inference.Response{{Content: "OCR page one", Model: "fixture"}, {Content: "OCR page two", Model: "fixture"}}}
			runner, results, _ := fixtureRunner("application/pdf", "", adapter, fakeRenderer{pages: []PageImage{{1, "image/png", []byte("scan1")}, {2, "image/png", []byte("scan2")}}})
			reader := &fakeTextExtractor{pages: tc.pages, err: tc.err}
			runner.TextExtractor = reader
			runner.LocalConfigurationVersion = "text-first"
			docs := runner.Documents.(*fakeDocuments)
			docs.version.Resources[0].URL = "https://comune.example/orders/ordinanza.pdf"
			docs.version.Resources[0].LocalProcessing = &registry.LocalProcessing{Version: registry.LocalProcessingVersion, TextPDFPathPrefixes: []string{tc.scope}}
			payload, _ := json.Marshal(Payload{DocumentVersionID: 7, ResourceURL: docs.version.Resources[0].URL, Workload: "evaluation"})
			if _, err := runner.Handler()(context.Background(), jobs.Job{ID: 9, Attempt: 1, Payload: payload}); err != nil {
				t.Fatal(err)
			}
			if len(adapter.requests) != 2 || results.resource.PageCount != 2 || results.pages[1].ReturnedModel == PDFTextVersion {
				t.Fatal("partial native output accepted", results)
			}
			if tc.scope != "/orders/" && reader.calls != 0 {
				t.Fatal("unreviewed PDF sent to native extractor")
			}
		})
	}
}

type interruptedNativeResults struct {
	memoryResults
	fail bool
}

func (m *interruptedNativeResults) PutPage(ctx context.Context, p PageResult) error {
	if p.PageNumber == 2 && m.fail {
		return errors.New("fixture interrupted write")
	}
	return m.memoryResults.PutPage(ctx, p)
}
func TestNativePDFResumeDoesNotMixModelEvidence(t *testing.T) {
	runner, _, _ := fixtureRunner("application/pdf", "", &fakeAdapter{}, fakeRenderer{})
	docs := runner.Documents.(*fakeDocuments)
	docs.version.Resources[0].URL = "https://comune.example/orders/ordinanza.pdf"
	docs.version.Resources[0].LocalProcessing = &registry.LocalProcessing{Version: registry.LocalProcessingVersion, TextPDFPathPrefixes: []string{"/orders/"}}
	reader := &fakeTextExtractor{pages: []TextPage{{1, "Il Comune dispone la chiusura del ponte."}, {2, "Il Comune dispone la riapertura del ponte."}}}
	runner.TextExtractor = reader
	runner.LocalConfigurationVersion = "text-first"
	results := &interruptedNativeResults{fail: true}
	runner.Results = results
	payload, _ := json.Marshal(Payload{DocumentVersionID: 7, ResourceURL: docs.version.Resources[0].URL, Workload: "evaluation"})
	if _, err := runner.Handler()(context.Background(), jobs.Job{ID: 9, Attempt: 1, Payload: payload}); err == nil {
		t.Fatal("interrupted write accepted")
	}
	reader.err = ErrNativeText
	results.fail = false
	_, err := runner.Handler()(context.Background(), jobs.Job{ID: 9, Attempt: 2, Payload: payload})
	var failure *jobs.HandlerError
	if !errors.As(err, &failure) || failure.Failure.Code != "native_text_retry_unavailable" || len(runner.Adapter.(*fakeAdapter).requests) != 0 {
		t.Fatal("mixed native and model pages", err)
	}
	reader.err = nil
	if _, err := runner.Handler()(context.Background(), jobs.Job{ID: 9, Attempt: 3, Payload: payload}); err != nil || results.resource == nil || len(results.pages) != 2 {
		t.Fatal("native resume failed", err)
	}
}

func TestNativePDFRejectsImagesAndIncompleteToolOutput(t *testing.T) {
	header := "page   num  type   width height color comp bpc enc interp object ID x-ppi y-ppi size ratio\n--------------------------------------------------------------------------------------------\n"
	if !noPDFImages([]byte(header)) || noPDFImages([]byte(header+"1 0 image 1 1 gray 1 8 image no 6 0 72 72 1B 100%\n")) || noPDFImages([]byte("unexpected output")) {
		t.Fatal("image list failed closed")
	}
	dir := t.TempDir()
	script := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"+body), 0700); err != nil {
			t.Fatal(err)
		}
	}
	script("pdfinfo", "printf 'Pages: 2\nEncrypted: no\n'")
	script("pdfimages", "printf '"+header+"'")
	script("pdftotext", "printf 'Il Comune dispone la chiusura del ponte.\fIl Comune dispone la riapertura del ponte.\f'")
	t.Setenv("PATH", dir)
	if pages, err := (PopplerText{}).Extract(context.Background(), []byte("fixture")); err != nil || len(pages) != 2 {
		t.Fatal("native tool envelope", pages, err)
	}
	script("pdfinfo", "printf 'Pages: 2\nEncrypted: yes (print:yes)\n'")
	if _, err := (PopplerText{}).Extract(context.Background(), []byte("fixture")); err == nil {
		t.Fatal("encrypted document accepted")
	}
	script("pdfinfo", "printf 'Pages: 2\nEncrypted: no\n'")
	script("pdftotext", "printf 'Il Comune dispone la chiusura del ponte.\f\f'")
	if _, err := (PopplerText{}).Extract(context.Background(), []byte("fixture")); err == nil {
		t.Fatal("empty page silently omitted")
	}
	script("pdftotext", "printf 'Il Comune dispone la chiusura del ponte.\f'")
	if _, err := (PopplerText{}).Extract(context.Background(), []byte("fixture")); err == nil {
		t.Fatal("wrong page count accepted")
	}
	if err := os.Remove(filepath.Join(dir, "pdfinfo")); err != nil {
		t.Fatal(err)
	}
	if _, err := (PopplerText{}).Extract(context.Background(), []byte("fixture")); err == nil {
		t.Fatal("missing tools accepted")
	}
	out := &boundedOutput{limit: 5}
	if _, err := out.Write([]byte("abcdef")); err == nil || out.Len() != 0 {
		t.Fatal("output limit bypassed")
	}
}

// syntheticTextPDF is redistributable, with a real page tree and optional raster
// image. It exercises Poppler itself rather than treating fake PDF bytes as valid.
func syntheticTextPDF(texts []string, image bool) []byte {
	objects := []string{"", "", "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"}
	var kids []string
	for _, text := range texts {
		page := len(objects) + 1
		stream := page + 1
		kids = append(kids, fmt.Sprintf("%d 0 R", page))
		imageResource := ""
		imageCommand := ""
		if image {
			imageResource = " /XObject << /Im1 4 0 R >>"
			imageCommand = "q 10 0 0 10 20 20 cm /Im1 Do Q\n"
		}
		escaped := strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)").Replace(text)
		data := fmt.Sprintf("BT /F1 12 Tf 20 180 Td (%s) Tj ET\n%s", escaped, imageCommand)
		objects = append(objects, fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 300 200] /Resources << /Font << /F1 3 0 R >>%s >> /Contents %d 0 R >>", imageResource, stream), fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(data), data))
	}
	// Image fixtures have one page; its image object follows its content stream.
	if image {
		objects[3] = strings.Replace(objects[3], "/Im1 4 0 R", "/Im1 6 0 R", 1)
		objects = append(objects, "<< /Type /XObject /Subtype /Image /Width 1 /Height 1 /ColorSpace /DeviceGray /BitsPerComponent 8 /Length 1 >>\nstream\n\x00\nendstream")
	}
	objects[0] = "<< /Type /Catalog /Pages 2 0 R >>"
	objects[1] = fmt.Sprintf("<< /Type /Pages /Count %d /Kids [%s] >>", len(texts), strings.Join(kids, " "))
	var b strings.Builder
	b.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for i, obj := range objects {
		offsets = append(offsets, b.Len())
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	start := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&b, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, start)
	return []byte(b.String())
}

func TestPDFTextUsesRealPoppler(t *testing.T) {
	for _, command := range []string{"pdfinfo", "pdfimages", "pdftotext"} {
		if _, err := exec.LookPath(command); err != nil {
			t.Skip("real Poppler unavailable; run this test in the application image")
		}
	}
	text := "Il Comune dispone la chiusura del ponte per rischio idraulico."
	pages, err := (PopplerText{}).Extract(context.Background(), syntheticTextPDF([]string{text, text}, false))
	if err != nil || len(pages) != 2 || !strings.Contains(pages[0].Text, "chiusura del ponte") {
		t.Fatal("real native text", pages, err)
	}
	for _, pdf := range [][]byte{syntheticTextPDF([]string{text, ""}, false), syntheticTextPDF([]string{text}, true), []byte("%PDF-invalid")} {
		if _, err := (PopplerText{}).Extract(context.Background(), pdf); err == nil {
			t.Fatal("real mixed/image/invalid PDF accepted")
		}
	}
}

func TestNativePDFResumeKeepsStartedModelPath(t *testing.T) {
	adapter := &fakeAdapter{responses: []inference.Response{{Content: "OCR second page", Model: "fixture"}}}
	runner, results, _ := fixtureRunner("application/pdf", "", adapter, fakeRenderer{pages: []PageImage{{1, "image/png", []byte("scan1")}, {2, "image/png", []byte("scan2")}}})
	docs := runner.Documents.(*fakeDocuments)
	docs.version.Resources[0].URL = "https://comune.example/orders/ordinanza.pdf"
	docs.version.Resources[0].LocalProcessing = &registry.LocalProcessing{Version: registry.LocalProcessingVersion, TextPDFPathPrefixes: []string{"/orders/"}}
	results.pages = map[int]PageResult{1: {PageNumber: 1, Status: "complete", ReturnedModel: "fixture", ExtractedText: "OCR first page"}}
	reader := &fakeTextExtractor{pages: []TextPage{{1, "Il Comune dispone la chiusura del ponte."}, {2, "Il Comune dispone la riapertura del ponte."}}}
	runner.TextExtractor = reader
	runner.LocalConfigurationVersion = "text-first"
	payload, _ := json.Marshal(Payload{DocumentVersionID: 7, ResourceURL: docs.version.Resources[0].URL, Workload: "evaluation"})
	if _, err := runner.Handler()(context.Background(), jobs.Job{ID: 9, Attempt: 2, Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if reader.calls != 0 || len(adapter.requests) != 1 || results.pages[1].ExtractedText != "OCR first page" || results.pages[2].ReturnedModel == PDFTextVersion {
		t.Fatal("resumed model evidence replaced by native pages")
	}
}
