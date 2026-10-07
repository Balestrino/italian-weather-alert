package classification

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDecodedJSONVisibleScalarsRetainPointersAndLiteralEvidence(t *testing.T) {
	body := []byte(`{"contenuto":"<p>Il centro \\u0026 la biblioteca: l&#8217;attività è aperta.</p>","a/b":{"~value":["<strong>Evento ordinario</strong>",12,true,null]}}`)
	text, scalars, err := decodedJSONContent(body)
	if err != nil || len(scalars) != 5 || strings.Contains(text, "<p>") || !strings.Contains(text, "l’attività") {
		t.Fatal(text, scalars, err)
	}
	for _, scalar := range scalars {
		if scalar.EndByte <= scalar.StartByte || scalar.EndByte > len(text) {
			t.Fatal("bad scalar provenance", scalar)
		}
	}
	if scalars[0].Pointer != "/a~1b/~0value/0" {
		t.Fatal("JSON pointer escaping", scalars)
	}
	again, _, _ := decodedJSONContent(body)
	if text != again {
		t.Fatal("nondeterministic JSON preparation")
	}
	quote := text[scalars[0].StartByte:scalars[0].EndByte]
	raw, _ := json.Marshal(Decision{Relevant: false, ReasonCode: "not_relevant", EvidenceQuote: quote})
	if _, err := ParseDecision(string(raw), Content{Text: text}); err != nil {
		t.Fatal("decoded literal rejected", err)
	}
	if _, _, err := decodedJSONContent([]byte(`{"x":1} trailing`)); err == nil {
		t.Fatal("malformed JSON accepted")
	}
}

func TestEvidenceOptionsConstrainGenerationAndPreserveStrictValidator(t *testing.T) {
	text := strings.Repeat("Evento ordinario: attività per famiglie. ", 20) + "Il COC sarà aperto dalle ore 17.00 per l'allerta meteo."
	segment := Segment{DocumentVersionID: 1, Ordinal: 1, Total: 1, ResourceURL: "https://fixture.example", EndByte: len(text), Text: text}
	request, err := SegmentRequest("fixture", segment)
	if err != nil {
		t.Fatal(err)
	}
	var format struct {
		JSONSchema struct {
			Schema struct {
				Properties map[string]struct {
					Enum []string `json:"enum"`
				} `json:"properties"`
			} `json:"schema"`
		} `json:"json_schema"`
	}
	if err := json.Unmarshal(request.ResponseFormat, &format); err != nil {
		t.Fatal(err)
	}
	options := format.JSONSchema.Schema.Properties["evidence_quote"].Enum
	if len(options) < 2 {
		t.Fatal("missing bounded quote choices")
	}
	for _, option := range options {
		if !strings.Contains(text, option) || len(option) > 120 {
			t.Fatal("invented quotation option", option)
		}
	}
	if _, err := ParseDecision(`{"relevant":true,"reason_code":"local_weather_measure","evidence_quote":"Il centro operativo chiuderà alle 18"}`, segment.Content()); err == nil {
		t.Fatal("paraphrase accepted")
	}
}
