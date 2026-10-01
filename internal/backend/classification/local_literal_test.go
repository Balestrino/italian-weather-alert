package classification

import (
	"encoding/json"
	"testing"
)

func TestLocalDoubleQuotePresentationReturnsVerbatimEvidence(t *testing.T) {
	source := `Titolo. L'attività è riservata alla scuola “G. Esempio”. Contatti.`
	quote := `L'attività è riservata alla scuola "G. Esempio".`
	raw, _ := json.Marshal(Decision{Relevant: false, ReasonCode: "not_relevant", EvidenceQuote: quote})
	decision, err := ParseLocalDecision(string(raw), Content{Text: source})
	if err != nil || decision.EvidenceQuote != `L'attività è riservata alla scuola “G. Esempio”.` {
		t.Fatal("local presentation did not return original contiguous source", decision, err)
	}
	if _, err := ParseDecision(string(raw), Content{Text: source}); err == nil {
		t.Fatal("remote parsing contract changed")
	}
	for _, invalid := range []string{
		`L'attività ... scuola "G. Esempio".`,
		`L'attività è riservata alla scuola "A. Inventato".`,
		`L'attività è riservata alla scuola "G. Esempio". Contatti inesistenti.`,
	} {
		raw, _ := json.Marshal(Decision{Relevant: false, ReasonCode: "not_relevant", EvidenceQuote: invalid})
		if _, err := ParseLocalDecision(string(raw), Content{Text: source}); err == nil {
			t.Fatal("unsupported quotation accepted", invalid)
		}
	}
	if got, ok := CanonicalLocalLiteral("Prima. “İSTANBUL” dopo.", `"istanbul"`); !ok || got != "“İSTANBUL”" {
		t.Fatal("Unicode source offsets lost", got, ok)
	}
}
