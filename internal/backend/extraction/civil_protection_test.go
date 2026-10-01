package extraction

import (
	"encoding/json"
	"testing"
)

func TestCivilProtectionOpeningRequiresLiteralOperativeScope(t *testing.T) {
	for _, test := range []struct {
		name, text, subject string
		accepted            bool
	}{
		{"centre", "Il COC (Centro Operativo Comunale) sarà aperto dalle ore 06.00 di martedì.", "COC (Centro Operativo Comunale)", true},
		{"expanded name", "Il Centro Operativo Comunale resta aperto durante l'evento.", "Centro Operativo Comunale", true},
		{"unrelated office", "L'ufficio postale sarà aperto dalle ore 06.00.", "ufficio postale", false},
		{"negated", "Il COC (Centro Operativo Comunale) non sarà aperto dalle ore 06.00.", "COC (Centro Operativo Comunale)", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			window := Window{DocumentVersionID: 1, Ordinal: 1, Total: 1, ResourceURL: "https://example.org/notice", Role: "original", Text: test.text, CoreStartByte: 0, CoreEndByte: len(test.text), ContextEndByte: len(test.text)}
			raw, _ := json.Marshal(map[string]any{"envelope_version": CompactEnvelopeVersionV2, "window_ordinal": 1, "evidence": []string{test.text}, "measures": []any{map[string]any{"kind": "activation", "subject": test.subject, "place": nil, "evidence_refs": map[string][]int{"kind": {0}, "subject": {0}, "place": {}}, "temporal_candidates": []any{}}}})
			got, err := ParseWindow(string(raw), window)
			if test.accepted && (err != nil || len(got) != 1 || got[0].Kind != "activation" || got[0].ValidFrom != nil) {
				t.Fatalf("literal centre opening lost or alert time invented: %v %+v", err, got)
			}
			if !test.accepted && err == nil && len(got) > 0 {
				t.Fatal("unsupported activation accepted")
			}
			window.legacyLiteral = true
			if got, err := ParseWindow(string(raw), window); err == nil && len(got) > 0 {
				t.Fatal("legacy contract changed")
			}
		})
	}
}
