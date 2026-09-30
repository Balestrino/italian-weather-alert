package classification

import (
	"strings"
	"testing"
)

func TestCanonicalLiteralPreservesOriginalAndRejectsAssertions(t *testing.T) {
	for _, tc := range []struct{ source, quote, want string }{
		{"fino al perdurare dell’emergenza", "fino al perdurare dell'emergenza", "fino al perdurare dell’emergenza"},
		{"**AVVERTE** che si riserva, al termine dell’emergenza", "AVVERTE che si riserva, al termine dell'emergenza", "AVVERTE** che si riserva, al termine dell’emergenza"},
		{"İ AVVISO DI CHIUSURA", "avviso di chiusura", "AVVISO DI CHIUSURA"},
	} {
		literal, ok := CanonicalLiteral(tc.source, tc.quote)
		if !ok || literal != tc.want || !strings.Contains(tc.source, literal) {
			t.Fatalf("literal mapping failed: %q", literal)
		}
	}
	for _, quote := range []string{"chiusura ... riapertura", "chiusura il 19 settembre", "riapertura il 18 settembre", "chiusura perchè", ""} {
		if _, ok := CanonicalLiteral("chiusura il 18 settembre perché persiste acqua", quote); ok {
			t.Fatalf("unsupported quote accepted: %q", quote)
		}
	}
}

func TestDecisionRequiresExplicitBooleanAndContiguousEvidence(t *testing.T) {
	content := Content{Complete: true, Text: "Avviso ordinario"}
	for _, raw := range []string{
		`{"reason_code":"not_relevant","evidence_quote":"Avviso"}`,
		`{"relevant":null,"reason_code":"not_relevant","evidence_quote":"Avviso"}`,
		`{"relevant":false,"reason_code":"not_relevant","evidence_quote":"Avviso ... ordinario"}`,
	} {
		if _, err := ParseDecision(raw, content); err == nil {
			t.Fatalf("unsupported decision accepted: %s", raw)
		}
	}
}
