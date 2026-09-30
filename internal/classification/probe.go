package classification

import (
	"context"
	"errors"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/inference"
)

type ProbeCaseResult struct {
	ID, ReasonCode     string
	Expected, Relevant bool
	InputTokens        *int64
	OutputTokens       *int64
}

type ProbeResult struct {
	Adapter, RequestedModel string
	ObservedAt              time.Time
	Cases                   []ProbeCaseResult
}

func ProbeReviewedCases(ctx context.Context, adapter inference.Adapter, model string, observedAt time.Time) (ProbeResult, error) {
	if adapter == nil || model == "" || observedAt.IsZero() {
		return ProbeResult{}, ErrInvalid
	}
	cases := []struct {
		id, text string
		expected bool
	}{
		{"CAL-ORD34-relevant", "Comune di Calcinaia. Ordinanza 34/2026. Per allerta meteo e rischio idraulico, dalle ore 18 del 20 agosto sono chiusi i cimiteri comunali e l'area di sgambamento cani fino al perdurare dell'emergenza.", true},
		{"CAL-NONRELEVANT", "Museo Coccapani: apertura della nuova mostra. Scacchi in riva all'Arno: iscrizioni aperte per il torneo di domenica.", false},
		{"ordinary-closure", "La biblioteca comunale resterà chiusa lunedì per manutenzione dell'impianto elettrico.", false},
		{"misleading-keyword", "La biblioteca presenta il romanzo Allerta meteo. Ingresso libero e nessuna modifica ai servizi.", false},
		{"empty-source-category", "Categoria della fonte: non valorizzata. A causa dell'allerta meteo e del rischio idraulico il Comune dispone la chiusura del sottopasso dalle ore 18.", true},
	}
	result := ProbeResult{Adapter: adapter.Name(), RequestedModel: model, ObservedAt: observedAt.UTC()}
	for _, test := range cases {
		content := fullContent{Complete: true, Sections: []contentSection{{ResourceURL: "https://fixture.example/" + test.id, Role: "original", Text: test.text}}, Text: test.text}
		request, err := Request(model, 1, content)
		if err != nil {
			return ProbeResult{}, err
		}
		response, err := adapter.Complete(ctx, request)
		if err != nil {
			return ProbeResult{}, err
		}
		decision, err := ParseDecision(response.Content, content)
		if err != nil {
			return ProbeResult{}, err
		}
		caseResult := ProbeCaseResult{ID: test.id, Expected: test.expected, Relevant: decision.Relevant, ReasonCode: decision.ReasonCode, InputTokens: response.Usage.InputTokens, OutputTokens: response.Usage.OutputTokens}
		result.Cases = append(result.Cases, caseResult)
		if decision.Relevant != test.expected {
			return result, errors.New("classification_mismatch")
		}
	}
	return result, nil
}
