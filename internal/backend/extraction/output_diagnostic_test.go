package extraction

import "testing"

func TestWindowOutputReasons(t *testing.T) {
	w := Window{Ordinal: 1, ResourceURL: "https://source.example", Text: "chiusura parco"}
	legacy := w
	legacy.legacyLiteral = true
	typographic := Window{Ordinal: 1, ResourceURL: "https://source.example", Text: "Un’ordinanza vieta il transito nel tratto tra il civico 38 e il civico 42 ."}
	for _, tc := range []struct {
		raw, finish, want string
		window            Window
	}{
		{`{`, "length", "truncated", w},
		{`{"envelope_version":"unknown"}`, "stop", "envelope_version", w},
		{`{"envelope_version":"compact-evidence-v1","window_ordinal":2}`, "stop", "window_reference", w},
		{`{"envelope_version":"compact-evidence-v1","window_ordinal":1,"evidence":["chiusura"],"measures":[{"evidence_refs":{"kind":[9]}}]}`, "stop", "evidence_reference", w},
		{`{"envelope_version":"compact-evidence-v1","window_ordinal":1,"evidence":["**chiusura**"],"measures":[]}`, "stop", "normalization", legacy},
		{`{"envelope_version":"compact-evidence-v1","window_ordinal":1,"evidence":["riapertura"],"measures":[]}`, "stop", "quotation", w},
		// Accepted typography is not reported; the later non-literal quote is.
		{`{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["Un'ordinanza vieta il transito","tra il civico 38 e il civico 42."],"measures":[]}`, "stop", "quotation", typographic},
		// Every quote is accepted, so the parser's evidence failure is reported.
		{`{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["Un'ordinanza vieta il transito"],"measures":[]}`, "stop", "evidence_reference", typographic},
	} {
		if got := InvalidOutputReason(tc.raw, tc.finish, tc.window, ErrEvidence); got != "extraction_output_"+tc.want {
			t.Fatalf("%s: got %s want %s", tc.raw, got, tc.want)
		}
	}
}
