package classification

import "testing"

func TestInvalidOutputReasonsRemainStable(t *testing.T) {
	content := Content{Text: "Comune dispone chiusura", Complete: true}
	for _, tc := range []struct{ raw, finish, want string }{
		{`{"relevant":`, "length", "truncated"},
		{`{"relevant":`, "stop", "json_syntax"},
		{`{}`, "stop", "schema"},
		{`{"relevant":true,"reason_code":"invented","evidence_quote":"Comune"}`, "stop", "reason"},
		{`{"relevant":true,"reason_code":"local_weather_measure","evidence_quote":"**Comune**"}`, "stop", "normalization"},
		{`{"relevant":true,"reason_code":"local_weather_measure","evidence_quote":"Comune ... chiusura"}`, "stop", "quotation"},
	} {
		if got := InvalidOutputReason(tc.raw, tc.finish, content); got != "classification_output_"+tc.want {
			t.Fatalf("got %s want %s", got, tc.want)
		}
	}
}
