//go:build llm

package inference_test

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/linking"
)

//go:embed testdata/local-llm-notice.png
var localOCRNotice []byte

type nonThinkingTestAdapter struct{ inference.Adapter }

func (a nonThinkingTestAdapter) Complete(ctx context.Context, request inference.Request) (inference.Response, error) {
	request = inference.LocalChatRequest(request)
	return a.Adapter.Complete(ctx, request)
}

func localModel(t *testing.T) (*inference.OpenAIChat, string) {
	t.Helper()
	endpoint, model := os.Getenv("IWA_LOCAL_LLM_ENDPOINT"), os.Getenv("IWA_LOCAL_LLM_MODEL")
	if endpoint == "" || model == "" {
		t.Skip("set IWA_LOCAL_LLM_ENDPOINT and IWA_LOCAL_LLM_MODEL to opt into real model calls")
	}
	adapter, err := inference.NewOpenAIChat("local-openai-chat", endpoint, "local-no-key", &http.Client{Timeout: 180 * time.Second})
	if err != nil {
		t.Fatal("invalid local test endpoint")
	}
	return adapter, model
}

func localCompletion(t *testing.T, adapter inference.Adapter, request inference.Request) inference.Response {
	t.Helper()
	request = inference.LocalChatRequest(request)
	ctx, cancel := context.WithTimeout(t.Context(), 180*time.Second)
	defer cancel()
	response, err := adapter.Complete(ctx, request)
	if err != nil {
		t.Fatal("local provider call failed:", err)
	}
	if response.FinishReason != "stop" || response.Model == "" {
		t.Fatal("incomplete response or missing model identity")
	}
	t.Logf("response_sha256=%x returned_model_sha256=%x", sha256.Sum256([]byte(response.Content)), sha256.Sum256([]byte(response.Model)))
	if directory := os.Getenv("IWA_LOCAL_LLM_REVIEW_DIR"); directory != "" {
		if err := os.MkdirAll(directory, 0700); err != nil {
			t.Fatal("review directory unavailable")
		}
		body, _ := json.Marshal(struct {
			Request  inference.Request
			Response inference.Response
		}{request, response})
		name := strings.NewReplacer("/", "-", " ", "-").Replace(t.Name()) + ".json"
		if err := os.WriteFile(filepath.Join(directory, name), body, 0600); err != nil {
			t.Fatal("review capture unavailable")
		}
	}
	return response
}

func TestLocalLLMClassification(t *testing.T) {
	adapter, model := localModel(t)
	for _, test := range []struct {
		name, text string
		relevant   bool
	}{
		{"weather closure", "Per allerta meteo e rischio idraulico, il Comune dispone la chiusura dei cimiteri comunali dalle ore 18.00 fino al termine dell'emergenza.", true},
		{"civil protection", "Per l'allerta meteo il COC (Centro Operativo Comunale) sarà aperto dalle ore 06.00.", true},
		{"ordinary maintenance", "La biblioteca comunale resta chiusa lunedì per manutenzione dell'impianto elettrico.", false},
		{"misleading title", "La biblioteca presenta il romanzo Allerta meteo. Ingresso libero e nessuna modifica ai servizi.", false},
		{"long ordinary notice", "Comune di Esempio. " + strings.Repeat("Amministrazione Servizi Notizie Contatti. ", 75) + "Sono aperte le iscrizioni ai laboratori pomeridiani per bambini. " + strings.Repeat("Privacy Accessibilità Contatti. ", 75), false},
		{"typographic quotes", "Sono aperte le iscrizioni ai laboratori della scuola “G. Esempio” per bambini.", false},
		{"unrelated opening", "L'ufficio postale sarà aperto sabato dalle ore 09.00.", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			segment := classification.Segment{DocumentVersionID: 1, Ordinal: 1, Total: 1, ResourceURL: "https://fixture.example/notice", Role: "original", Text: test.text, EndByte: len(test.text)}
			request, err := classification.LocalSegmentRequest(model, segment)
			if err != nil {
				t.Fatal(err)
			}
			response := localCompletion(t, adapter, request)
			decision, err := classification.ParseLocalDecision(response.Content, segment.Content())
			if err != nil || decision.Relevant != test.relevant {
				t.Fatal("classification or literal evidence mismatch", classification.InvalidOutputReason(response.Content, response.FinishReason, segment.Content()), err)
			}
		})
	}
}

func TestLocalLLMExtraction(t *testing.T) {
	adapter, model := localModel(t)
	for _, test := range []struct {
		name, text, kind, subject string
		missingTimes              bool
	}{
		{"closure", "A causa dell'allerta meteo il Comune dispone la chiusura dei cimiteri comunali dalle ore 18.00 del 2 ottobre 2026 fino alle ore 08.00 del 3 ottobre 2026.", "closure", "cimiteri", false},
		{"civil protection without alert time borrowing", "Allerta meteo dalle ore 06.00 alle ore 18.00 del 2 ottobre 2026. Il COC (Centro Operativo Comunale) sarà aperto durante l'emergenza.", "activation", "Centro Operativo Comunale", true},
		{"partial reopening", "Il sottopasso di via Esempio è stato riaperto. Restano chiusi i cimiteri comunali per il rischio idraulico.", "reopening", "sottopasso", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			content := classification.Content{Complete: true, Text: test.text, Sections: []classification.ContentSection{{ResourceURL: "https://fixture.example/notice", Role: "original", Text: test.text}}}
			windows, err := extraction.BuildOperationalWindows(1, content)
			if err != nil || len(windows) != 1 {
				t.Fatal("fixture windows", err)
			}
			request, err := extraction.LocalWindowRequest(model, windows[0])
			if err != nil {
				t.Fatal(err)
			}
			response := localCompletion(t, adapter, request)
			measures, err := extraction.ParseWindow(response.Content, windows[0])
			if err != nil {
				t.Fatal("invalid extraction evidence", err)
			}
			var found bool
			for _, measure := range measures {
				if measure.Kind == test.kind && strings.Contains(strings.ToLower(measure.Subject), strings.ToLower(test.subject)) {
					found = true
					if test.missingTimes && (measure.ValidFrom != nil || measure.ValidUntil != nil) {
						t.Fatal("times borrowed from unrelated evidence")
					}
					if !test.missingTimes && (measure.ValidFrom == nil || measure.ValidUntil == nil) {
						t.Fatal("explicit measure times omitted")
					}
				}
			}
			if !found {
				t.Fatal("expected operative fact omitted")
			}
			if test.name == "partial reopening" {
				var retained bool
				for _, m := range measures {
					retained = retained || m.Kind == "closure" && strings.Contains(m.Subject, "cimiteri")
				}
				if !retained {
					t.Fatal("retained closure omitted")
				}
			}
		})
	}
}

func TestLocalLLMCivilProtectionInLongWindow(t *testing.T) {
	adapter, model := localModel(t)
	text := "Comune di Esempio. " + strings.Repeat("Amministrazione Servizi Notizie Contatti. ", 30) + "Allerta regionale arancio dalle ore 06.00 alle ore 18.00 del 4 ottobre 2026. Il COC (Centro Operativo Comunale) sarà aperto in concomitanza con l'inizio dell'allerta (dalle ore 06.00 del 4 ottobre 2026). " + strings.Repeat("Privacy Accessibilità Contatti. ", 30)
	content := classification.Content{Complete: true, Text: text, Sections: []classification.ContentSection{{ResourceURL: "https://fixture.example/notice", Role: "original", Text: text}}}
	windows, err := extraction.BuildOperationalWindows(1, content)
	if err != nil || len(windows) != 1 {
		t.Fatal("fixture windows", err)
	}
	request, err := extraction.LocalWindowRequest(model, windows[0])
	if err != nil {
		t.Fatal(err)
	}
	response := localCompletion(t, adapter, request)
	measures, err := extraction.ParseWindow(response.Content, windows[0])
	if err != nil {
		t.Fatal("invalid literal evidence", err)
	}
	for _, measure := range measures {
		if measure.Kind == "activation" && (strings.Contains(measure.Subject, "COC") || strings.Contains(measure.Subject, "Centro Operativo Comunale")) {
			if measure.ValidFrom == nil || measure.ValidUntil != nil {
				t.Fatal("opening start omitted or alert end borrowed")
			}
			return
		}
	}
	t.Fatal("civil protection opening omitted in long notice")
}

func TestLocalLLMLinking(t *testing.T) {
	adapter, model := localModel(t)
	current := linking.MeasureContext{RunID: 2, DocumentVersionID: 2, Ordinal: 1, SourceID: "fixture", Municipality: "050004", Kind: "reopening", Subject: "sottopasso", Place: "via Esempio", OfficialURL: "https://fixture.example/reopening", Evidence: []extraction.Evidence{{Field: "kind", Quote: "Il sottopasso di via Esempio è stato riaperto."}}}
	previous := linking.Candidate{MeasureContext: linking.MeasureContext{RunID: 1, DocumentVersionID: 1, Ordinal: 1, SourceID: "fixture", Municipality: "050004", Kind: "closure", Subject: "sottopasso", Place: "via Esempio", OfficialURL: "https://fixture.example/closure", Evidence: []extraction.Evidence{{Field: "kind", Quote: "Per il rischio idraulico è chiuso il sottopasso di via Esempio."}}}}
	request, err := linking.Request(model, current, []linking.Candidate{previous})
	if err != nil {
		t.Fatal(err)
	}
	response := localCompletion(t, adapter, request)
	result, err := linking.Parse(response.Content, current, []linking.Candidate{previous})
	if err != nil || result.Status != "linked" || result.Relation == nil || *result.Relation != "reopens" {
		t.Fatal("literal update relation lost", err)
	}
}

func TestLocalLLMVisionCapability(t *testing.T) {
	localModel(t)
	endpoint := os.Getenv("IWA_LOCAL_LLM_ENDPOINT")
	base := strings.TrimSuffix(endpoint, "/v1/chat/completions")
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, base+"/props", nil)
	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(request)
	if err != nil {
		t.Fatal("cannot inspect local model capabilities")
	}
	defer response.Body.Close()
	var properties struct {
		Modalities struct {
			Vision bool `json:"vision"`
		} `json:"modalities"`
	}
	if response.StatusCode != 200 || json.NewDecoder(response.Body).Decode(&properties) != nil {
		t.Fatal("llama.cpp capability response unavailable")
	}
	if !properties.Modalities.Vision {
		t.Skip("server has no vision projector; OCR fallback is not validated")
	}
	t.Log("vision is enabled")
}

func TestLocalLLMOCR(t *testing.T) {
	adapter, model := localModel(t)
	TestLocalLLMVisionCapability(t)
	ctx, cancel := context.WithTimeout(t.Context(), 180*time.Second)
	defer cancel()
	result, err := inference.ProbeOCR(ctx, nonThinkingTestAdapter{adapter}, model, "image/png", localOCRNotice, []inference.OCRCheck{
		{Name: "authority", Text: "COMUNE DI ESEMPIO"},
		{Name: "weather", Text: "ALLERTA METEO"},
		{Name: "measure", Text: "CIMITERI COMUNALI CHIUSI"},
		{Name: "start", Text: "18.00 DEL 02/10/2026"},
		{Name: "end", Text: "08.00 DEL 03/10/2026"},
	}, time.Now())
	if err != nil || result.FinishReason != "stop" || len(result.MissingChecks) > 0 || len(result.MatchedChecks) != 5 {
		t.Fatal("synthetic OCR notice mismatch", result.MissingChecks, err)
	}
	t.Logf("synthetic OCR passed: output_bytes=%d output_sha256=%s", result.OutputBytes, result.OutputSHA256)
}
