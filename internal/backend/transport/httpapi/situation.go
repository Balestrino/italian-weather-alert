package httpapi

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/publicquery"
)

// situationData is the concise public presentation of a pinned domain situation.
// Detailed quality, receipts and document histories remain in search/document/coverage.
type situationData struct {
	Municipality    situationMunicipality `json:"municipality"`
	ProcessedData   processedSituation    `json:"processed_data"`
	SourceSummaries situationSources      `json:"source_summaries"`
}
type situationMunicipality struct {
	ISTAT              string   `json:"istat"`
	Name               string   `json:"name"`
	Zones              []string `json:"zones"`
	MappingVersion     *string  `json:"mapping_version,omitempty"`
	MappingLimitations []string `json:"mapping_limitations,omitempty"`
}
type processedSituation struct {
	Summary           string               `json:"summary"`
	RegionalAlerts    []situationFact      `json:"regional_alerts"`
	LocalMeasures     []situationFact      `json:"local_measures"`
	OperationalPhases []situationFact      `json:"operational_phases"`
	References        []situationReference `json:"references"`
}
type situationFact struct {
	Weather          *acquisition.VigilanceWeather `json:"weather,omitempty"`
	ID               string                        `json:"id"`
	Product          string                        `json:"product,omitempty"`
	Risk             string                        `json:"risk,omitempty"`
	RiskLabel        string                        `json:"risk_label,omitempty"`
	Zone             string                        `json:"zone,omitempty"`
	Level            string                        `json:"level,omitempty"`
	Action           string                        `json:"action,omitempty"`
	Subject          string                        `json:"subject,omitempty"`
	Place            *string                       `json:"place,omitempty"`
	TerritorialScope string                        `json:"territorial_scope,omitempty"`
	Authority        string                        `json:"authority,omitempty"`
	Phase            string                        `json:"phase,omitempty"`
	Status           string                        `json:"status"`
	Reliability      string                        `json:"reliability"`
	Validity         situationValidity             `json:"validity"`
	Limitations      []string                      `json:"limitations,omitempty"`
	ReferenceIDs     []string                      `json:"reference_ids"`
}
type situationValidity struct {
	Precision  string     `json:"precision"`
	Original   string     `json:"original,omitempty"`
	Instant    *time.Time `json:"instant,omitempty"`
	Date       *string    `json:"date,omitempty"`
	EndInstant *time.Time `json:"end_instant,omitempty"`
	Timezone   *string    `json:"timezone,omitempty"`
	Assumption *string    `json:"assumption,omitempty"`
}
type situationReference struct {
	ID         string `json:"id"`
	DocumentID string `json:"document_id"`
	VersionID  string `json:"version_id"`
	URL        string `json:"url"`
	Page       *int   `json:"page,omitempty"`
}
type situationSources struct {
	Regional           sourceSummary `json:"regional"`
	Municipal          sourceSummary `json:"municipal"`
	CittadinoInformato sourceSummary `json:"cittadino_informato"`
}
type sourceSummary struct {
	Name     string                 `json:"name"`
	Status   string                 `json:"status"`
	Summary  string                 `json:"summary"`
	Products []sourceProductSummary `json:"products"`
}
type sourceProductSummary struct {
	SourceID       string     `json:"source_id"`
	Product        string     `json:"product"`
	Status         string     `json:"status"`
	Summary        string     `json:"summary"`
	LastCheckedAt  *time.Time `json:"last_checked_at,omitempty"`
	CoverageStatus string     `json:"coverage_status"`
	ReferenceIDs   []string   `json:"reference_ids"`
}

func summarizeSituation(value publicquery.Situation) situationData {
	result := situationData{Municipality: situationMunicipality{ISTAT: value.Municipality.ISTAT, Name: value.Municipality.Name, Zones: value.Municipality.Zones, MappingVersion: value.Municipality.MappingVersion, MappingLimitations: value.Municipality.MappingLimitations},
		ProcessedData:   processedSituation{RegionalAlerts: []situationFact{}, LocalMeasures: []situationFact{}, OperationalPhases: []situationFact{}, References: []situationReference{}},
		SourceSummaries: situationSources{Regional: emptySourceSummary("Regione Toscana / CFR"), Municipal: emptySourceSummary("Comune di " + value.Municipality.Name), CittadinoInformato: emptySourceSummary("Cittadino Informato")}}
	references := map[string]string{}
	cite := func(evidence []publicquery.Evidence) []string {
		ids := []string{}
		for _, e := range evidence {
			page := 0
			if e.Page != nil {
				page = *e.Page
			}
			key := fmt.Sprintf("%s/%s/%s/%d", e.DocumentID, e.VersionID, e.SourceURL, page)
			id, ok := references[key]
			if !ok {
				id = fmt.Sprintf("ref-%d", len(references)+1)
				references[key] = id
				result.ProcessedData.References = append(result.ProcessedData.References, situationReference{ID: id, DocumentID: e.DocumentID, VersionID: e.VersionID, URL: e.SourceURL, Page: e.Page})
			}
			if !slices.Contains(ids, id) {
				ids = append(ids, id)
			}
		}
		return ids
	}
	makeFact := func(id, status string, validity publicquery.Temporal, quality publicquery.Quality, evidence []publicquery.Evidence) situationFact {
		limits := situationLimitations(quality.Interpretation.Limitations)
		reliability := quality.Interpretation.State
		if reliability == "" {
			reliability = "not_processed"
		}
		if quality.Provenance.State != "verified" {
			if reliability == "supported" {
				reliability = "partial"
			}
			limits = append(limits, "Provenienza della fonte non verificata.")
		}
		if status == "undetermined" {
			limits = append(limits, "Validità attuale non determinabile: non considerare questa misura confermata in vigore.")
		}
		if validity.Assumption != nil && !slices.Contains(limits, *validity.Assumption) {
			limits = append(limits, *validity.Assumption)
		}
		return situationFact{ID: id, Status: status, Reliability: reliability, Validity: situationValidity{Precision: validity.Precision, Original: validity.Original, Instant: validity.Instant, Date: validity.Date, EndInstant: validity.EndInstant, Timezone: validity.Timezone, Assumption: validity.Assumption}, Limitations: limits, ReferenceIDs: cite(evidence)}
	}
	for _, r := range value.RegionalProducts {
		f := makeFact(r.ID, r.Status, r.Validity, r.Quality, r.Evidence)
		f.Product, f.Risk, f.RiskLabel, f.Zone, f.Level = r.Product, r.Risk, r.OfficialRiskLabel, r.Zone, r.Level
		f.Weather = r.Weather
		result.ProcessedData.RegionalAlerts = append(result.ProcessedData.RegionalAlerts, f)
	}
	active, undetermined := 0, 0
	for _, m := range value.LocalMeasures {
		f := makeFact(m.ID, m.Status, m.Validity, m.Quality, m.Evidence)
		f.Action, f.Subject, f.Place, f.TerritorialScope = m.Action, m.Subject, m.Place, m.TerritorialScope
		if len(m.NewerUninterpretedDocumentIDs) > 0 {
			f.Limitations = append(f.Limitations, "Esistono pubblicazioni successive non ancora interpretate; potrebbero aggiornare questa misura.")
		}
		if m.Issuer != nil {
			f.Authority = *m.Issuer
		}
		result.ProcessedData.LocalMeasures = append(result.ProcessedData.LocalMeasures, f)
		if m.Status == "current" {
			active++
		}
		if m.Status == "undetermined" {
			undetermined++
		}
	}
	for _, p := range value.OperationalPhases {
		f := makeFact(p.ID, "documented", p.Validity, p.Quality, p.Evidence)
		f.Authority, f.Phase = p.Authority, p.Phase
		result.ProcessedData.OperationalPhases = append(result.ProcessedData.OperationalPhases, f)
	}
	regionalText := regionalConclusion(value.RegionalProducts)
	municipalText := fmt.Sprintf("%d misure comunali documentate; %d con validità attuale determinata, %d con validità non determinabile.", len(value.LocalMeasures), active, undetermined)
	if len(value.LocalMeasures) == 0 {
		municipalText = "Nessuna misura comunale dedotta dai dati disponibili; questo non dimostra l'assenza di provvedimenti."
	}
	result.ProcessedData.Summary = regionalText + " " + municipalText
	municipalContents := []string{}
	actionLabels := map[string]string{"closure": "chiusura", "suspension": "sospensione", "prohibition": "divieto", "activation": "attivazione", "reopening": "riapertura"}
	for _, measure := range value.LocalMeasures {
		label := measure.Action
		if translated, ok := actionLabels[label]; ok {
			label = translated
		}
		if measure.Subject != "" {
			label += " — " + measure.Subject
		}
		if !slices.Contains(municipalContents, label) {
			municipalContents = append(municipalContents, label)
		}
	}
	for _, c := range value.Coverage {
		group := &result.SourceSummaries.Regional
		if c.Platform == "cittadino-informato" {
			group = &result.SourceSummaries.CittadinoInformato
		} else if c.Product == "municipal" {
			group = &result.SourceSummaries.Municipal
		}
		status := "unavailable"
		if c.PublicState == "enabled" || c.DevelopmentPublication {
			status = c.Quality.Updating.State
			if status == "ok" {
				status = "available"
			}
			if status == "" {
				status = "unavailable"
			}
		}
		text := "Dati non disponibili nella vista richiesta."
		evidence := []publicquery.Evidence{}
		if status != "unavailable" {
			evidence = c.Quality.Interpretation.Evidence
			switch {
			case c.Platform == "cittadino-informato":
				text = "Canale aggiuntivo: i contenuti richiedono riscontro nelle fonti primarie comunali o CFR; non stabiliscono da soli misure o colori."
			case c.Product == "municipal":
				text = municipalText
				if len(value.DocumentsRequiringAttention) > 0 {
					text += " L'elaborazione degli avvisi acquisiti è incompleta; non è verificata la completezza dei provvedimenti."
				}
			case c.Product == "criticality":
				text = regionalText
			case c.Product == "vigilance":
				text = "Bollettino di vigilanza acquisito: descrive fenomeni in valutazione, distinti dai livelli di criticità."
			case c.Product == "monitoring":
				text = "Il monitoraggio non assegna un nuovo colore di allerta."
			}
			for _, l := range c.Quality.Interpretation.Limitations {
				if c.Product == "vigilance" && strings.Contains(l, "graphical details are retained but not interpreted") {
					text += " Fasce di pioggia e dettagli grafici non interpretati."
				}
				if strings.HasPrefix(l, "Retained bulletin statement: ") && c.Product == "monitoring" {
					text += " " + strings.TrimPrefix(l, "Retained bulletin statement: ")
				}
			}
			if c.Quality.Interpretation.State != "supported" && c.Product != "municipal" {
				text += " L'interpretazione del prodotto è parziale o non completata."
			}
		}
		group.Products = append(group.Products, sourceProductSummary{SourceID: c.SourceID, Product: c.Product, Status: status, Summary: text, LastCheckedAt: c.Quality.Updating.LastCompleteAt, CoverageStatus: c.CoverageStatus, ReferenceIDs: cite(evidence)})
	}
	for _, g := range []*sourceSummary{&result.SourceSummaries.Regional, &result.SourceSummaries.Municipal, &result.SourceSummaries.CittadinoInformato} {
		if len(g.Products) == 0 {
			continue
		}
		usable := 0
		fresh := 0
		for _, p := range g.Products {
			if p.Status != "unavailable" && p.Status != "unverified" {
				usable++
			}
			if p.Status == "available" {
				fresh++
			}
		}
		g.Status = "unavailable"
		if usable > 0 {
			g.Status = "partial"
		}
		if fresh == len(g.Products) {
			g.Status = "available"
		}
		if g == &result.SourceSummaries.Regional {
			g.Summary = regionalText
		} else if g == &result.SourceSummaries.Municipal {
			g.Summary = municipalText
			if len(municipalContents) > 0 {
				g.Summary = "Il Comune riporta: " + strings.Join(municipalContents, "; ") + ". " + municipalText
			}
			if len(value.DocumentsRequiringAttention) > 0 {
				g.Summary += " Elaborazione degli avvisi incompleta: la completezza dei provvedimenti non è verificata."
			}
		} else {
			g.Summary = "Canale aggiuntivo da confrontare con Comune e CFR; la disponibilità dei dati non dimostra che ogni contenuto sia stato confermato."
		}
		if usable == 0 {
			g.Summary = "Fonte configurata, ma dati non disponibili nella vista richiesta; non è possibile dedurre l'assenza di avvisi."
		}
	}
	return result
}

func situationLimitations(values []string) []string {
	result := []string{}
	labels := map[string]string{
		"place": "Luogo non determinato.", "place_not_established": "Luogo non determinato.",
		"valid_from": "Decorrenza non determinata.", "valid_until": "Termine non determinato.",
		"validity_not_established": "Validità temporale non determinata.",
	}
	for _, value := range values {
		if strings.HasPrefix(value, "Development publication:") {
			continue // The envelope already identifies development limitations.
		}
		if label, ok := labels[value]; ok {
			value = label
		}
		if !slices.Contains(result, value) {
			result = append(result, value)
		}
	}
	return result
}

func emptySourceSummary(name string) sourceSummary {
	return sourceSummary{Name: name, Status: "not_collected", Summary: "Nessun dato di questa fonte acquisito e pubblicabile nella vista richiesta; non è possibile dedurre l'assenza di avvisi.", Products: []sourceProductSummary{}}
}

func regionalConclusion(values []publicquery.RegionalWarning) string {
	current, green, elevated, unknown, excluded := 0, 0, 0, 0, 0
	for _, r := range values {
		if r.Product != "criticality" || r.Status != "current" {
			continue
		}
		current++
		if r.Quality.Interpretation.State != "supported" || r.Quality.Provenance.State != "verified" {
			unknown++
			continue
		}
		switch r.Level {
		case "green":
			green++
		case "yellow", "orange", "red":
			elevated++
		case "not_applicable":
			excluded++
		default:
			unknown++
		}
	}
	if current == 0 {
		return "Criticità regionale attuale non determinabile dai dati disponibili."
	}
	if elevated > 0 {
		return fmt.Sprintf("La criticità regionale riporta %d livelli gialli, arancioni o rossi nelle zone applicabili; consultare rischi e validità.", elevated)
	}
	if unknown > 0 {
		return fmt.Sprintf("Criticità regionale interpretata solo in parte: %d livelli verdi, %d rischi non applicabili e %d livelli non determinabili.", green, excluded, unknown)
	}
	exclusionLabel := "rischi non applicabili"
	if excluded == 1 {
		exclusionLabel = "rischio non applicabile"
	}
	return fmt.Sprintf("La criticità regionale attuale riporta %d livelli verdi e %d %s nelle zone del Comune.", green, excluded, exclusionLabel)
}
