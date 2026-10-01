package extraction

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
)

func TestOperationalWindowsAreDeterministicBoundedAndCoverCoresOnce(t *testing.T) {
	raw := "## **Premessa**  " + strings.Repeat("dato amministrativo; ", 20) + "ORDINA la chiusura del ponte Verde dalle 18. " + strings.Repeat("segue relazione tecnica. ", 20)
	view := classification.NormalizeOCR(raw)
	content := classification.Content{Complete: true, Sections: []classification.ContentSection{{ResourceURL: "https://comune.example/ordinanza.pdf", Role: "attachment", Page: 2, Text: view.Text, OffsetMap: view.OffsetMap}}}
	first, err := buildOperationalWindows(77, content, 128, 48)
	if err != nil {
		t.Fatal(err)
	}
	second, err := buildOperationalWindows(77, content, 128, 48)
	if err != nil || len(first) < 3 || len(first) != len(second) {
		t.Fatalf("operational windows unavailable: %d %d %v", len(first), len(second), err)
	}
	coreCursor := 0
	owners := 0
	for index, window := range first {
		other := second[index]
		if window.Hash != other.Hash || window.ContextStartByte != other.ContextStartByte || window.ContextEndByte != other.ContextEndByte || window.CoreStartByte < coreCursor || strings.TrimSpace(view.Text[coreCursor:window.CoreStartByte]) != "" || window.CoreEndByte <= window.CoreStartByte || window.ContextStartByte > window.CoreStartByte || window.ContextEndByte < window.CoreEndByte || len(window.Text) > 128+2*48 || window.SourceEndByte <= window.SourceStartByte || len(window.OffsetMap) == 0 {
			t.Fatalf("invalid deterministic window %d: %#v / %#v", index, window, other)
		}
		coreCursor = window.CoreEndByte
		if window.ownsOperativeQuote("ORDINA la chiusura") {
			owners++
		}
	}
	if coreCursor != len(view.Text) || owners != 1 {
		t.Fatalf("cores did not cover normalized page once or predicate has %d owners: %d/%d", owners, coreCursor, len(view.Text))
	}
	predicate := strings.Index(view.Text, "ORDINA")
	for _, window := range first {
		if predicate >= window.ContextStartByte && predicate < window.ContextEndByte && predicate < window.CoreStartByte && window.ownsOperativeQuote("ORDINA la chiusura") {
			t.Fatal("context-only overlap acquired operative ownership")
		}
	}
}

func TestParseWindowAllowsLocalContextAndRejectsCrossWindowEvidence(t *testing.T) {
	text := "ponte Verde nel Comune di Calcinaia. ORDINA la chiusura dalle ore 18 fino a revoca. relazione successiva"
	predicate := strings.Index(text, "ORDINA")
	window := Window{Ordinal: 2, Total: 3, DocumentVersionID: 8, ResourceURL: "https://comune.example/o", Role: "attachment", Page: 1, ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: predicate, CoreEndByte: len(text) - len(" relazione successiva"), Text: text, Hash: strings.Repeat("a", 64)}
	raw := `{"envelope_version":"compact-evidence-v1","window_ordinal":2,"evidence":["ORDINA la chiusura","ponte Verde","Comune di Calcinaia","dalle ore 18","fino a revoca"],"measures":[{"kind":"closure","subject":"ponte Verde","place":"Comune di Calcinaia","valid_from":"dalle ore 18","valid_until":"fino a revoca","evidence_refs":{"kind":[0],"subject":[1],"place":[2],"valid_from":[3],"valid_until":[4]}}]}`
	measures, err := ParseWindow(raw, window)
	if err != nil || len(measures) != 1 || measures[0].Evidence[0].SegmentOrdinal != 2 || len(measures[0].TemporalCandidates) != 2 || measures[0].TemporalCandidates[0].OriginalExpression != "dalle ore 18" || measures[0].TemporalCandidates[0].ConflictIdentity != nil {
		t.Fatalf("same-window context was not accepted: %#v %v", measures, err)
	}

	contextOnly := window
	contextOnly.CoreStartByte = predicate + len("ORDINA la chiusura")
	if _, err = ParseWindow(raw, contextOnly); !errors.Is(err, ErrEvidence) {
		t.Fatalf("context-only predicate acquired ownership: %v", err)
	}
	crossWindow := strings.Replace(raw, `"window_ordinal":2`, `"window_ordinal":1`, 1)
	if _, err = ParseWindow(crossWindow, window); !errors.Is(err, ErrInvalid) {
		t.Fatalf("cross-window evidence was accepted: %v", err)
	}
}

func TestCompactWindowRejectsMalformedDuplicateOutOfRangeAndNonLiteralReferences(t *testing.T) {
	text := "ORDINA la chiusura del ponte Verde"
	window := Window{Ordinal: 3, Total: 4, DocumentVersionID: 8, ResourceURL: "https://comune.example/o", Role: "original", ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: 0, CoreEndByte: len(text), Text: text, Hash: strings.Repeat("b", 64)}
	valid := `{"envelope_version":"compact-evidence-v1","window_ordinal":3,"evidence":["ORDINA la chiusura","ponte Verde"],"measures":[{"kind":"closure","subject":"ponte Verde","place":null,"valid_from":null,"valid_until":null,"evidence_refs":{"kind":[0],"subject":[1],"place":[],"valid_from":[],"valid_until":[]}}]}`
	cases := []struct {
		name string
		raw  string
		err  error
	}{
		{"malformed", strings.Replace(valid, `"subject":[1]`, `"subject":null`, 1), ErrInvalid},
		{"duplicate-table-entry", strings.Replace(valid, `"ponte Verde"]`, `"ORDINA la chiusura"]`, 1), ErrEvidence},
		{"duplicate-reference", strings.Replace(valid, `"subject":[1]`, `"subject":[1,1]`, 1), ErrEvidence},
		{"out-of-range", strings.Replace(valid, `"subject":[1]`, `"subject":[2]`, 1), ErrEvidence},
		{"cross-window", strings.Replace(valid, `"window_ordinal":3`, `"window_ordinal":2`, 1), ErrInvalid},
		{"non-literal", strings.Replace(valid, `"ponte Verde"]`, `"ponte Rosso"]`, 1), ErrEvidence},
		{"non-null-empty", strings.Replace(valid, `"place":null`, `"place":" "`, 1), ErrInvalid},
		{"response-budget", valid + strings.Repeat(" ", MaxCompactResponseBytes), ErrInvalid},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ParseWindow(test.raw, window); !errors.Is(err, test.err) {
				t.Fatalf("invalid compact reference accepted: %v", err)
			}
		})
	}
	measures, err := ParseWindow(valid, window)
	if err != nil || len(measures) != 1 || !slices.Equal(measures[0].IndeterminateFields, []string{"place", "valid_from", "valid_until"}) {
		t.Fatalf("local null-to-indeterminate projection failed: %#v %v", measures, err)
	}
	withUnused := strings.Replace(valid, `"ponte Verde"]`, `"ponte Verde","chiusura"]`, 1)
	if measures, err = ParseWindow(withUnused, window); err != nil || len(measures) != 1 {
		t.Fatalf("unused literal evidence discarded supported measures: %#v %v", measures, err)
	}
}

func TestCompactWindowDropsUnsupportedOptionalValueWithoutReference(t *testing.T) {
	text := "Il sottopasso di via Maremmana resta chiuso"
	window := Window{Ordinal: 1, Total: 1, DocumentVersionID: 4, ResourceURL: "https://comune.example/avviso", Role: "original", ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: 0, CoreEndByte: len(text), Text: text}
	raw := `{"envelope_version":"compact-evidence-v1","window_ordinal":1,"evidence":["Il sottopasso di via Maremmana resta chiuso"],"measures":[{"kind":"restriction","subject":"sottopasso di via Maremmana","place":"Calcinaia","valid_from":null,"valid_until":null,"evidence_refs":{"kind":[0],"subject":[0],"place":[],"valid_from":[],"valid_until":[]}}]}`
	measures, err := ParseWindow(raw, window)
	if err != nil || len(measures) != 1 || measures[0].Place != nil || !slices.Contains(measures[0].IndeterminateFields, "place") {
		t.Fatalf("unsupported optional value was not projected to unknown: %#v %v", measures, err)
	}
}

func TestCompactWindowV2ProjectsExplicitTemporalConflict(t *testing.T) {
	text := "ORDINA la chiusura dei cimiteri dal 10 settembre. Il dispositivo ordina la chiusura dei cimiteri dal 10 agosto."
	window := Window{Ordinal: 1, Total: 1, DocumentVersionID: 9, ResourceURL: "https://comune.example/o", Role: "original", ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: 0, CoreEndByte: len(text), Text: text}
	raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["ORDINA la chiusura dei cimiteri dal 10 settembre","Il dispositivo ordina la chiusura dei cimiteri dal 10 agosto"],"measures":[{"kind":"closure","subject":"cimiteri","place":null,"evidence_refs":{"kind":[0,1],"subject":[0,1],"place":[]},"temporal_candidates":[{"field":"valid_from","original_expression":"10 settembre","evidence_refs":[0]},{"field":"valid_from","original_expression":"10 agosto","evidence_refs":[1]}]}]}`
	measures, err := ParseWindow(raw, window)
	if err != nil {
		t.Fatal(err)
	}
	if len(measures) != 1 || measures[0].ValidFrom != nil || len(measures[0].TemporalCandidates) != 2 || measures[0].TemporalCandidates[0].ConflictIdentity == nil || *measures[0].TemporalCandidates[0].ConflictIdentity != *measures[0].TemporalCandidates[1].ConflictIdentity || !slices.Contains(measures[0].IndeterminateFields, "valid_from") {
		t.Fatalf("explicit temporal candidates were not projected as one conflict: %#v", measures)
	}
	request, err := WindowRequest("qwen3.8-27b", window)
	if err != nil || !strings.Contains(string(request.ResponseFormat), CompactEnvelopeVersionV2) {
		t.Fatalf("production request did not select compact v2: %v", err)
	}
}

func TestCompactWindowV2FindsLocalConflictOmittedByModel(t *testing.T) {
	text := "Emessa ordinanza di chiusura cimiteri a partire dalle 0.00 di Giovedì 10 Settembre. Allerta meteo a partire dalle ore 15.00 di mercoledì 9 settembre. Il Sindaco dispone, a partire dalle ore 0.00 di giovedì 10 agosto, la chiusura dei cimiteri."
	window := Window{Ordinal: 1, Total: 1, DocumentVersionID: 9, ResourceURL: "https://comune.example/o", Role: "original", ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: 0, CoreEndByte: len(text), Text: text}
	raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["Il Sindaco dispone, a partire dalle ore 0.00 di giovedì 10 agosto, la chiusura dei cimiteri"],"measures":[{"kind":"closure","subject":"cimiteri","place":null,"evidence_refs":{"kind":[0],"subject":[0],"place":[]},"temporal_candidates":[]}]}`
	measures, err := ParseWindow(raw, window)
	if err != nil || len(measures) != 1 || measures[0].ValidFrom != nil || len(measures[0].TemporalCandidates) != 2 || measures[0].TemporalCandidates[0].ConflictIdentity == nil || strings.Contains(strings.Join([]string{measures[0].TemporalCandidates[0].OriginalExpression, measures[0].TemporalCandidates[1].OriginalExpression}, " "), "15.00") {
		t.Fatalf("local temporal conflict was not isolated from alert metadata: %#v %v", measures, err)
	}
	content := classification.Content{Complete: true, Sections: []classification.ContentSection{{ResourceURL: window.ResourceURL, Role: window.Role, Text: text}}}
	if merged, mergeErr := Merge([][]Measure{measures}, content, 1); mergeErr != nil || len(merged) != 1 {
		t.Fatalf("locally completed temporal evidence did not survive merge: %#v %v", merged, mergeErr)
	}
}

func TestCompactWindowV2RejectsOnlyUnsupportedCandidate(t *testing.T) {
	text := "ORDINA la chiusura del ponte Verde. DISPONE l'attivazione delle seguenti funzioni di supporto: volontariato."
	window := Window{Ordinal: 1, Total: 1, DocumentVersionID: 10, ResourceURL: "https://comune.example/o", Role: "original", ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: 0, CoreEndByte: len(text), Text: text}
	raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["ORDINA la chiusura del ponte Verde","DISPONE l'attivazione delle seguenti funzioni di supporto: volontariato"],"measures":[{"kind":"closure","subject":"ponte Verde","place":null,"evidence_refs":{"kind":[0],"subject":[0],"place":[]},"temporal_candidates":[]},{"kind":"activation","subject":"funzioni: volontariato","place":null,"evidence_refs":{"kind":[1],"subject":[1],"place":[]},"temporal_candidates":[]}]}`
	measures, err := ParseWindow(raw, window)
	if err != nil || len(measures) != 2 || measures[0].Subject != "ponte Verde" || measures[0].Ordinal != 1 || measures[1].Subject == "funzioni: volontariato" {
		t.Fatalf("unsupported candidate erased supported window facts: %#v %v", measures, err)
	}
}

func TestCompactWindowV2CompletesExplicitMissingOperativeKinds(t *testing.T) {
	text := "ORDINA la chiusura dei cimiteri; il divieto di qualunque attività all'aperto: - nei parchi; - negli orti; - la sospensione delle attività didattiche educative; DISPONE l'attivazione delle seguenti funzioni di supporto: volontariato."
	window := Window{Ordinal: 1, Total: 1, DocumentVersionID: 10, ResourceURL: "https://comune.example/o", Role: "original", ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: 0, CoreEndByte: len(text), Text: text}
	raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["ORDINA la chiusura dei cimiteri"],"measures":[{"kind":"closure","subject":"cimiteri","place":null,"evidence_refs":{"kind":[0],"subject":[0],"place":[]},"temporal_candidates":[]}]}`
	measures, err := ParseWindow(raw, window)
	kinds := map[string]bool{}
	for _, measure := range measures {
		kinds[measure.Kind] = true
	}
	if err != nil || !kinds["closure"] || !kinds["prohibition"] || !kinds["suspension"] || !kinds["activation"] {
		t.Fatalf("explicit operative kinds were omitted: %#v %v", measures, err)
	}
}

func TestCompactWindowV2RejectsOnlyCandidateUsingNonLiteralEvidence(t *testing.T) {
	text := "ORDINA la chiusura del ponte Verde. Dispone la sospensione delle lezioni."
	window := Window{Ordinal: 1, Total: 1, DocumentVersionID: 10, ResourceURL: "https://comune.example/o", Role: "original", ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: 0, CoreEndByte: len(text), Text: text}
	raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["ORDINA la chiusura del ponte Verde","Dispone la esposizione delle lezioni"],"measures":[{"kind":"closure","subject":"ponte Verde","place":null,"evidence_refs":{"kind":[0],"subject":[0],"place":[]},"temporal_candidates":[]},{"kind":"suspension","subject":"lezioni","place":null,"evidence_refs":{"kind":[1],"subject":[1],"place":[]},"temporal_candidates":[]}]}`
	measures, err := ParseWindow(raw, window)
	if err != nil || len(measures) != 1 || measures[0].Subject != "ponte Verde" || measures[0].Ordinal != 1 {
		t.Fatalf("non-literal evidence erased an independent supported candidate: %#v %v", measures, err)
	}
}

func TestCompactWindowV2KeepsMeasureWithOneValidKindReference(t *testing.T) {
	text := "ORDINA la chiusura del ponte Verde dalle ore 18."
	window := Window{Ordinal: 1, Total: 1, DocumentVersionID: 10, ResourceURL: "https://comune.example/o", Role: "original", ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: 0, CoreEndByte: len(text), Text: text}
	raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["ORDINO la chiusura","ordina la chiusura del ponte Verde","dalle ore diciotto"],"measures":[{"kind":"closure","subject":"ponte Verde","place":null,"evidence_refs":{"kind":[0,1],"subject":[1],"place":[]},"temporal_candidates":[{"field":"valid_from","original_expression":"dalle ore diciotto","evidence_refs":[2]}]}]}`
	measures, err := ParseWindow(raw, window)
	if err != nil || len(measures) != 1 || measures[0].Subject != "ponte Verde" || measures[0].ValidFrom != nil || len(measures[0].Evidence) != 2 || measures[0].Evidence[0].Quote != "ORDINA la chiusura del ponte Verde" {
		t.Fatalf("invalid optional references erased supported evidence: %#v %v", measures, err)
	}
}

func TestCompactWindowV2RecoversNearbyLiteralSubjectContext(t *testing.T) {
	text := "Il sottopasso di via Maremmana, chiuso ieri, è stato riaperto alla circolazione."
	window := Window{Ordinal: 1, Total: 1, DocumentVersionID: 10, ResourceURL: "https://comune.example/o", Role: "original", ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: 0, CoreEndByte: len(text), Text: text}
	raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["è stato riaperto alla circolazione"],"measures":[{"kind":"reopening","subject":"sottopasso di via Maremmana","place":"via Maremmana","evidence_refs":{"kind":[0],"subject":[0],"place":[0]},"temporal_candidates":[]}]}`
	measures, err := ParseWindow(raw, window)
	if err != nil || len(measures) != 1 || measures[0].Kind != "reopening" || measures[0].Place == nil || *measures[0].Place != "via Maremmana" || !valueInEvidence("sottopasso di via Maremmana", measures[0].Evidence) {
		t.Fatalf("nearby literal subject context was not retained: %#v %v", measures, err)
	}
}

func TestCompactWindowV2RejectsDistantSubjectContext(t *testing.T) {
	text := "Il sottopasso di via Maremmana è menzionato qui. " + strings.Repeat("x", 600) + " è stato riaperto alla circolazione."
	window := Window{Ordinal: 1, Total: 1, DocumentVersionID: 10, ResourceURL: "https://comune.example/o", Role: "original", ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: 0, CoreEndByte: len(text), Text: text}
	raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["è stato riaperto alla circolazione"],"measures":[{"kind":"reopening","subject":"sottopasso di via Maremmana","place":null,"evidence_refs":{"kind":[0],"subject":[0],"place":[]},"temporal_candidates":[]}]}`
	if measures, err := ParseWindow(raw, window); !errors.Is(err, ErrEvidence) || measures != nil {
		t.Fatalf("distant subject context was accepted: %#v %v", measures, err)
	}
}

func TestCompactWindowV2RejectsKindWithoutMatchingPredicate(t *testing.T) {
	text := "La situazione resta monitorata. L'attività all'aperto continua. Il COC resta attivo."
	window := Window{Ordinal: 1, Total: 1, DocumentVersionID: 10, ResourceURL: "https://comune.example/o", Role: "original", ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: 0, CoreEndByte: len(text), Text: text}
	raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["La situazione resta monitorata","L'attività all'aperto continua","Il COC resta attivo"],"measures":[{"kind":"activation","subject":"situazione","place":null,"evidence_refs":{"kind":[0],"subject":[0],"place":[]},"temporal_candidates":[]},{"kind":"activation","subject":"attività all'aperto","place":null,"evidence_refs":{"kind":[1],"subject":[1],"place":[]},"temporal_candidates":[]},{"kind":"activation","subject":"COC","place":null,"evidence_refs":{"kind":[2],"subject":[2],"place":[]},"temporal_candidates":[]}]}`
	measures, err := ParseWindow(raw, window)
	if err != nil || len(measures) != 1 || measures[0].Subject != "COC" {
		t.Fatalf("kind/predicate mismatch was accepted or erased the valid fact: %#v %v", measures, err)
	}
}

func TestCompactWindowV2ProjectsUnsupportedOptionalPlaceToUnknown(t *testing.T) {
	text := "Il sottopasso di via Maremmana resta chiuso."
	window := Window{Ordinal: 1, Total: 1, DocumentVersionID: 10, ResourceURL: "https://comune.example/o", Role: "original", ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: 0, CoreEndByte: len(text), Text: text}
	raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["Il sottopasso di via Maremmana resta chiuso"],"measures":[{"kind":"closure","subject":"sottopasso di via Maremmana","place":"Calcinaia","evidence_refs":{"kind":[0],"subject":[0],"place":[]},"temporal_candidates":[]}]}`
	measures, err := ParseWindow(raw, window)
	if err != nil || len(measures) != 1 || measures[0].Place != nil || !slices.Contains(measures[0].IndeterminateFields, "place") {
		t.Fatalf("unsupported optional place erased the measure: %#v %v", measures, err)
	}
}

func TestCompactWindowV2DoesNotApplyAlertValidityToLocalMeasure(t *testing.T) {
	text := "Allerta meteo in vigore fino alle ore 18. Il sottopasso resta chiuso."
	window := Window{Ordinal: 1, Total: 1, DocumentVersionID: 11, ResourceURL: "https://comune.example/o", Role: "original", ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: 0, CoreEndByte: len(text), Text: text}
	raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["Allerta meteo in vigore fino alle ore 18","Il sottopasso resta chiuso"],"measures":[{"kind":"closure","subject":"sottopasso","place":null,"evidence_refs":{"kind":[1],"subject":[1],"place":[]},"temporal_candidates":[{"field":"valid_until","original_expression":"fino alle ore 18","evidence_refs":[0]}]}]}`
	measures, err := ParseWindow(raw, window)
	if err != nil || len(measures) != 1 || measures[0].ValidUntil != nil || len(measures[0].TemporalCandidates) != 0 || !slices.Contains(measures[0].IndeterminateFields, "valid_until") {
		t.Fatalf("regional alert validity leaked into a local measure: %#v %v", measures, err)
	}
}

func TestCompactWindowV2DoesNotApplyOrdinanceIntervalToSeparateActivation(t *testing.T) {
	text := "ORDINA dalle ore 18 e fino all'emergenza la chiusura dei cimiteri. DISPONE l'attivazione del volontariato."
	window := Window{Ordinal: 1, Total: 1, DocumentVersionID: 11, ResourceURL: "https://comune.example/o", Role: "original", ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: 0, CoreEndByte: len(text), Text: text}
	raw := `{"envelope_version":"compact-evidence-v2","window_ordinal":1,"evidence":["ORDINA dalle ore 18 e fino all'emergenza la chiusura dei cimiteri","DISPONE l'attivazione del volontariato"],"measures":[{"kind":"activation","subject":"volontariato","place":null,"evidence_refs":{"kind":[1],"subject":[1],"place":[]},"temporal_candidates":[{"field":"valid_from","original_expression":"dalle ore 18","evidence_refs":[0]},{"field":"valid_until","original_expression":"fino all'emergenza","evidence_refs":[0]}]}]}`
	measures, err := ParseWindow(raw, window)
	if err != nil || len(measures) != 1 || measures[0].ValidFrom != nil || measures[0].ValidUntil != nil || len(measures[0].TemporalCandidates) != 0 {
		t.Fatalf("separate activation inherited ordinance interval: %#v %v", measures, err)
	}
}

func TestOperativePredicateAtEveryCoreBoundaryHasOneOwner(t *testing.T) {
	var source strings.Builder
	for index := 0; index < 30; index++ {
		fmt.Fprintf(&source, "frase-%02d relazione tecnica senza effetti. ", index)
	}
	text := strings.TrimSpace(source.String())
	content := classification.Content{Complete: true, Sections: []classification.ContentSection{{ResourceURL: "https://comune.example/o", Role: "original", Text: text}}}
	windows, err := buildOperationalWindows(9, content, 96, 40)
	if err != nil || len(windows) < 3 {
		t.Fatalf("boundary fixture did not create windows: %d %v", len(windows), err)
	}
	for _, owner := range windows {
		end := min(owner.CoreStartByte+24, owner.CoreEndByte)
		quote := strings.TrimSpace(text[owner.CoreStartByte:end])
		if quote == "" {
			t.Fatal("empty boundary quote")
		}
		owned := 0
		for _, candidate := range windows {
			if candidate.ownsOperativeQuote(quote) {
				owned++
			}
		}
		if owned != 1 {
			t.Fatalf("core boundary %d has %d owners for %q", owner.CoreStartByte, owned, quote)
		}
	}
}

func TestOperationalWindowsRejectMalformedOCRMap(t *testing.T) {
	content := classification.Content{Complete: true, Sections: []classification.ContentSection{{
		ResourceURL: "https://comune.example/o.pdf", Role: "attachment", Page: 1, Text: "ORDINA la chiusura",
		OffsetMap: []classification.OffsetSpan{{NormalizedStart: 1, NormalizedEnd: 19, SourceStart: 0, SourceEnd: 18}},
	}}}
	if _, err := BuildOperationalWindows(10, content); !errors.Is(err, ErrInvalid) {
		t.Fatalf("malformed retained OCR mapping accepted: %v", err)
	}
}

func TestWindowRequestKeepsPerCallBudget(t *testing.T) {
	text := "ORDINA la chiusura del ponte"
	window := Window{Ordinal: 1, Total: 1, DocumentVersionID: 4, ResourceURL: "https://comune.example/o", Role: "original", ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: 0, CoreEndByte: len(text), Text: text}
	request, err := WindowRequest("qwen3.8-27b", window)
	if err != nil || request.MaxCompletionTokens != MaxWindowCompletionTokens || !strings.Contains(string(request.Messages[1].Content), `\"core_start_byte\":0`) || !strings.Contains(string(request.ResponseFormat), CompactEnvelopeVersionV2) || strings.Contains(string(request.ResponseFormat), "indeterminate_fields") {
		t.Fatalf("window request changed bounded completion contract: %#v %v", request, err)
	}
	window.ContextEndByte = MaxWindowCoreBytes + 2*MaxWindowContextBytes + 1
	window.CoreEndByte = MaxWindowCoreBytes
	window.Text = strings.Repeat("x", window.ContextEndByte)
	if _, err := WindowRequest("qwen3.8-27b", window); !errors.Is(err, ErrInvalid) {
		t.Fatalf("oversized operational input was accepted: %v", err)
	}
}

func TestRetainedLongCaseCompactEnvelopeStaysWithinConfiguredBudgets(t *testing.T) {
	subjects := []string{"sottopasso di via Maremmana", "cimiteri", "area di sgambamento cani", "parchi e giardini pubblici", "aree verdi", "ciclopiste fluviali in riva d'Arno", "orti sociali di Fornacette", "Centro Operativo Comunale (COC)"}
	var source strings.Builder
	source.WriteString("Restano al momento in vigore: ")
	for _, subject := range subjects {
		source.WriteString(subject)
		source.WriteString("; ")
	}
	text := source.String()
	window := Window{Ordinal: 1, Total: 1, DocumentVersionID: 22, ResourceURL: "https://www.comune.calcinaia.pi.it/novita/monitoraggio-situazione-del-territorio", Role: "original", ContextStartByte: 0, ContextEndByte: len(text), CoreStartByte: 0, CoreEndByte: len(text), Text: text, Hash: strings.Repeat("c", 64)}
	evidence := append([]string{"Restano al momento in vigore"}, subjects...)
	measures := make([]compactWireMeasure, 0, len(subjects))
	for index, subject := range subjects {
		kind := "closure"
		if index == len(subjects)-1 {
			kind = "operational_update"
		}
		measures = append(measures, compactWireMeasure{Kind: kind, Subject: subject, Evidence: compactEvidenceRefs{Kind: []int{0}, Subject: []int{index + 1}, Place: []int{}, ValidFrom: []int{}, ValidUntil: []int{}}})
	}
	raw, err := json.Marshal(compactResponse{EnvelopeVersion: CompactEnvelopeVersion, WindowOrdinal: 1, Evidence: evidence, Measures: measures})
	if err != nil {
		t.Fatal(err)
	}
	request, err := WindowRequest("qwen3.8-27b", window)
	parsed, parseErr := ParseWindow(string(raw), window)
	if err != nil || parseErr != nil || len(parsed) != len(subjects) || len(window.Text) > MaxWindowCoreBytes+2*MaxWindowContextBytes || len(raw) > MaxCompactResponseBytes || request.MaxCompletionTokens != MaxWindowCompletionTokens {
		t.Fatalf("retained long case exceeded compact budgets: input=%d output=%d parsed=%d request=%v parse=%v", len(window.Text), len(raw), len(parsed), err, parseErr)
	}
}
