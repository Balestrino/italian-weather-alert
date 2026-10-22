package extraction

import (
	"encoding/json"
	"strings"
	"testing"
)

func prohibitionWindowResponse(t *testing.T, text, quote string) (Window, string) {
	t.Helper()
	if !strings.Contains(text, quote) {
		t.Fatal("synthetic quote is not literal")
	}
	b, _ := json.Marshal(map[string]any{"envelope_version": "compact-evidence-v2", "window_ordinal": 1, "evidence": []string{quote},
		"measures": []any{map[string]any{"kind": "prohibition", "subject": "transito veicolare e pedonale", "place": nil, "evidence_refs": map[string]any{"kind": []int{0}, "subject": []int{0}, "place": []int{}}, "temporal_candidates": []any{}}}})
	return Window{Ordinal: 1, Total: 1, Text: text, ResourceURL: "https://example.test/notice", CoreEndByte: len(text)}, string(b)
}

func TestFiniteProhibitionVerbSupportsProhibition(t *testing.T) {
	for _, quote := range []string{
		"un'ordinanza che vieta il transito veicolare e pedonale nel tratto allagato",
		"È vietato il transito veicolare e pedonale lungo l'argine",
		"sono vietati il transito veicolare e pedonale e la sosta",
	} {
		w, raw := prohibitionWindowResponse(t, "Per il rischio idraulico è stata emessa "+quote+".", quote)
		got, err := ParseWindow(raw, w)
		if err != nil || len(got) != 1 || got[0].Kind != "prohibition" || got[0].Subject != "transito veicolare e pedonale" {
			t.Fatal("finite prohibition verb rejected", quote, got, err)
		}
		for _, e := range got[0].Evidence {
			if !strings.Contains(w.Text, e.Quote) {
				t.Fatal("non-literal evidence", e)
			}
		}
		w.legacyLiteral = true
		if _, err = ParseWindow(raw, w); err == nil {
			t.Fatal("frozen legacy kind predicate changed", quote)
		}
	}
}

func TestProhibitionVerbRejectsNegationAndUnrelatedWords(t *testing.T) {
	for _, quote := range []string{
		"non è vietato il transito veicolare e pedonale",
		"l'ordinanza non vieta il transito veicolare e pedonale",
		"non sono vietati il transito veicolare e pedonale e la sosta",
		"il transito veicolare e pedonale prosegue verso il Vietnam",
	} {
		w, raw := prohibitionWindowResponse(t, "Si informa che "+quote+".", quote)
		if got, err := ParseWindow(raw, w); err == nil {
			t.Fatal("unsupported prohibition accepted", quote, got)
		}
	}
	heading := "Oggetto: si vieta il transito veicolare e pedonale"
	w, raw := prohibitionWindowResponse(t, heading+". VISTO il bollettino regionale.", heading)
	if got, err := ParseWindow(raw, w); len(got) != 0 {
		t.Fatal("ordinance heading established a prohibition", got, err)
	}
}
