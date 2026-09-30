package extraction

import "testing"

func TestWindowOutputReasons(t *testing.T) {
	w := Window{Ordinal: 1, ResourceURL: "https://source.example", Text: "chiusura parco"}
	for _, tc := range []struct{ raw, finish, want string }{
		{`{`, "length", "truncated"},
		{`{"envelope_version":"unknown"}`, "stop", "envelope_version"},
		{`{"envelope_version":"compact-evidence-v1","window_ordinal":2}`, "stop", "window_reference"},
		{`{"envelope_version":"compact-evidence-v1","window_ordinal":1,"evidence":["chiusura"],"measures":[{"evidence_refs":{"kind":[9]}}]}`, "stop", "evidence_reference"},
		{`{"envelope_version":"compact-evidence-v1","window_ordinal":1,"evidence":["**chiusura**"],"measures":[]}`, "stop", "normalization"},
		{`{"envelope_version":"compact-evidence-v1","window_ordinal":1,"evidence":["riapertura"],"measures":[]}`, "stop", "quotation"},
	} {
		if got := InvalidOutputReason(tc.raw, tc.finish, w, ErrEvidence); got != "extraction_output_"+tc.want {
			t.Fatalf("got %s want %s", got, tc.want)
		}
	}
}
