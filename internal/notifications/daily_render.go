package notifications

import (
	"fmt"
	"github.com/Balestrino/italian-weather-alert/internal/publicquery"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const MaxReportBytes = 256 * 1024

type ReportMail struct {
	Subject, Body string
	ObservedAt    time.Time
}

func RenderReport(s ReportSnapshot, timezone string) (ReportMail, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return ReportMail{}, ErrConfig
	}
	local := s.ObservedAt.In(loc)
	var b strings.Builder
	line := func(format string, args ...any) { fmt.Fprintf(&b, format+"\n", args...) }
	municipalities, sources := 0, 0
	for _, r := range s.Regions {
		municipalities += len(r.Municipalities)
		sources += len(r.Sources)
	}
	line("IWA — REPORT SERALE ALLERTE")
	line("Situazione al %s (%s)", local.Format("02/01/2006 15:04:05"), timezone)
	line("Ambito: %d regioni abilitate, %d comuni in raccolta, %d fonti.", len(s.Regions), municipalities, sources)
	line("Report operativo privato dai dati conservati; le fonti non pubblicate restano soggette alla valutazione di accettazione.")
	if len(s.Regions) == 0 {
		line("Nessun territorio abilitato.")
	}
	for _, r := range s.Regions {
		line("\nREGIONE: %s", cleanReportText(r.Name))
		if len(r.Sources) == 0 {
			line("Copertura non disponibile: nessuna fonte in raccolta compatibile.")
		}
		for _, source := range r.Sources {
			checked := "non disponibile"
			if source.LastChecked != nil {
				checked = source.LastChecked.In(loc).Format("02/01/2006 15:04")
			}
			state := "aggiornata"
			if source.LastChecked == nil {
				state = "mai verificata"
			} else if !s.ObservedAt.Before(source.LastChecked.Add(time.Duration(source.DelaySeconds) * time.Second)) {
				state = "in ritardo"
			}
			if source.Error != nil {
				state = "ultimo controllo con errore"
			}
			line("Fonte %s (%s): %s; ultimo controllo completo %s.", source.ID, reportProduct(source.Product), state, checked)
			if !source.Supported {
				line("Profilo non supportato per questa regione: stato delle allerte non disponibile.")
			}
			if !source.Public {
				line("Fonte in raccolta privata; pubblicazione non abilitata.")
			}
			if source.Product != "municipal" && source.Documents > len(source.Bulletins) {
				line("Bollettini: anteprima dei %d più recenti su %d documenti; i dati interpretati sono riepilogati sotto.", len(source.Bulletins), source.Documents)
			}
			for _, v := range source.Bulletins {
				line("Bollettino: %s; emissione %s; acquisizione %s.", v.URL, reportIssuance(v.Observation.IssuanceExpression), v.AcquiredAt.In(loc).Format("02/01/2006 15:04"))
				if len(v.Observation.ValidityExpressions) > 0 {
					line("Validità documentata: %s.", cleanReportText(strings.Join(v.Observation.ValidityExpressions, "; ")))
				}
				if v.Observation.TextSaysNoCriticality {
					line("Il testo del bollettino riporta: Criticità previste: NESSUNA. I colori non verificati restano indeterminati.")
				}
			}
		}
		groups := map[string][]publicquery.RegionalWarning{}
		for _, w := range r.Warnings {
			if omittedStatus(w.Status) {
				continue
			}
			period := reportPeriod(w.Validity, s.ObservedAt, loc)
			label := reportLevel(w.Level)
			if w.Level == "unknown" {
				label = "non determinato (colore non verificato)"
			}
			key := period + " | " + reportProduct(w.Product) + " | " + w.OfficialRiskLabel + " | " + label
			groups[key] = append(groups[key], w)
		}
		keys := []string{}
		for k := range groups {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		if len(keys) == 0 {
			line("Nessun livello regionale interpretato disponibile: non equivale ad assenza di allerta.")
		}
		for _, k := range keys {
			zones := map[string]bool{}
			evidence := map[string]bool{}
			limits := map[string]bool{}
			for _, w := range groups[k] {
				zones[w.Zone] = true
				for _, e := range w.Evidence {
					evidence[e.SourceURL] = true
				}
				collectReportLimits(limits, w.Quality)
			}
			line("%s — zone: %s.", cleanReportText(k), strings.Join(sortedReportKeys(zones), ", "))
			for _, url := range sortedReportKeys(evidence) {
				line("  Fonte: %s", cleanReportText(url))
			}
			for _, lim := range sortedReportKeys(limits) {
				line("  Limite: %s", cleanReportText(lim))
			}
		}
		for _, m := range r.Municipalities {
			line("\nCOMUNE: %s (ISTAT %s)", cleanReportText(m.Name), m.ISTAT)
			if len(m.Zones) == 0 {
				line("Associazione alle zone non disponibile; livello comunale non determinabile.")
			} else {
				line("Zone regionali associate: %s. I livelli restano quelli documentati per ciascun rischio/zona.", strings.Join(m.Zones, ", "))
				if m.PartialMapping {
					line("Associazione parziale: le zone non coprono uniformemente tutto il comune.")
				}
			}
			for _, source := range r.Sources {
				if source.Municipality != nil && *source.Municipality == m.ISTAT {
					line("Fonte %s: %d ultime versioni di documenti, %d senza classificazione, %d classificate senza estrazione disponibile. Non sono conteggi di allerte.", source.ID, source.Documents, source.Pending, source.Uninterpreted)
				}
			}
			n := 0
			for _, measure := range r.Measures {
				if measure.MunicipalityISTAT != m.ISTAT || omittedStatus(measure.Status) {
					continue
				}
				n++
				line("Provvedimento: %s; %s; ambito %s.", reportAction(measure.Action), reportPeriod(measure.Validity, s.ObservedAt, loc), measure.TerritorialScope)
				subject := ""
				for _, e := range measure.Evidence {
					if e.Locator != nil && *e.Locator == "subject" && e.Passage != nil {
						subject = *e.Passage
						break
					}
				}
				if subject != "" {
					line("  Soggetto documentato: %s", cleanReportText(subject))
				}
				if measure.Place != nil {
					line("  Luogo: %s", cleanReportText(*measure.Place))
				}
				renderReportTemporal(line, measure.Validity)
				for _, e := range measure.Evidence {
					line("  Fonte: %s", cleanReportText(e.SourceURL))
				}
				limits := map[string]bool{}
				collectReportLimits(limits, measure.Quality)
				if len(measure.NewerUninterpretedDocumentIDs) > 0 {
					limits["Documenti successivi non interpretati: situazione non completamente aggiornata."] = true
				}
				for _, lim := range sortedReportKeys(limits) {
					line("  Limite: %s", cleanReportText(lim))
				}
			}
			for _, phase := range r.Phases {
				if phase.MunicipalityISTAT != m.ISTAT || reportTemporalExpired(phase.Validity, s.ObservedAt, loc) {
					continue
				}
				line("Fase operativa documentata: %s (%s). È distinta dai colori regionali.", cleanReportText(phase.Phase), reportPeriod(phase.Validity, s.ObservedAt, loc))
				renderReportTemporal(line, phase.Validity)
				for _, e := range phase.Evidence {
					line("  Fonte: %s", cleanReportText(e.SourceURL))
				}
				limits := map[string]bool{}
				collectReportLimits(limits, phase.Quality)
				for _, lim := range sortedReportKeys(limits) {
					line("  Limite: %s", cleanReportText(lim))
				}
			}
			if n == 0 {
				line("Nessun provvedimento comunale interpretato disponibile. Non dimostra assenza di chiusure o restrizioni.")
			}
		}
	}
	line("\nValidità incerta, dati mancanti e nuove versioni non interpretate non costituiscono revoca o cessato allarme. Una riapertura riguarda solo il soggetto esplicitamente riaperto.")
	line("Consultare le fonti ufficiali. Generazione senza nuove chiamate a modelli, OCR o crawler.")
	body := b.String()
	if len(body) > MaxReportBytes {
		suffix := "\nREPORT TRONCATO: altri dettagli omessi per il limite email. I conteggi iniziali coprono l’intero ambito; usare l’anteprima privata per il dettaglio.\n"
		end := MaxReportBytes - len(suffix)
		for !utf8.RuneStart(body[end]) {
			end--
		}
		body = body[:end] + suffix
	}
	return ReportMail{Subject: "[IWA] Report allerte — " + local.Format("02/01/2006"), Body: body, ObservedAt: s.ObservedAt}, nil
}
func cleanReportText(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "\x00", "")), " ")
}
func omittedStatus(s string) bool { return s == "expired" || s == "cancelled" || s == "superseded" }
func sortedReportKeys(m map[string]bool) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func reportAction(s string) string {
	if label := map[string]string{"closure": "chiusura", "reopening": "riapertura", "prohibition": "divieto", "restriction": "limitazione", "activation": "attivazione", "deactivation": "disattivazione", "suspension": "sospensione", "operational_update": "aggiornamento operativo", "observation": "osservazione"}[s]; label != "" {
		return label
	}
	return cleanReportText(s)
}
func collectReportLimits(m map[string]bool, q publicquery.Quality) {
	for _, d := range []publicquery.Dimension{q.Provenance, q.Interpretation} {
		for _, l := range d.Limitations {
			m[l] = true
		}
		if d.State != "supported" && d.State != "" {
			m["Valutazione: "+reportState(d.State)] = true
		}
	}
	for _, l := range q.Updating.Limitations {
		m[l] = true
	}
	if q.Updating.State != "ok" && q.Updating.State != "" {
		m["Aggiornamento: "+reportState(q.Updating.State)] = true
	}
}
func renderReportTemporal(line func(string, ...any), v publicquery.Temporal) {
	if v.Original != "" {
		line("  Validità originale: %s", cleanReportText(v.Original))
	}
	if v.Precision == "unknown" || v.Precision == "condition" || v.Precision == "" {
		line("  Applicabilità temporale indeterminata; nessuna durata inferita.")
	}
}
func reportPeriod(v publicquery.Temporal, at time.Time, loc *time.Location) string {
	today := at.In(loc).Format(time.DateOnly)
	tomorrow := at.In(loc).AddDate(0, 0, 1).Format(time.DateOnly)
	if v.Precision == "date" && v.Date != nil {
		if *v.Date == today {
			return "Oggi"
		}
		if *v.Date == tomorrow {
			return "Domani"
		}
		return "Data documentata " + *v.Date
	}
	if v.Precision == "interval" && v.Instant != nil && v.EndInstant != nil {
		if !at.Before(*v.Instant) && at.Before(*v.EndInstant) {
			return "In corso"
		}
		if v.Instant.In(loc).Format(time.DateOnly) == tomorrow {
			return "Domani"
		}
		return "Intervallo documentato"
	}
	if v.Precision == "instant" && v.Instant != nil {
		return "Data/ora documentata " + v.Instant.In(loc).Format("02/01/2006 15:04")
	}
	return "Validità incerta"
}

func reportProduct(s string) string {
	if label := map[string]string{"criticality": "criticità / allerta", "vigilance": "vigilanza meteorologica", "monitoring": "monitoraggio", "municipal": "provvedimenti comunali", "dpc_comparison": "confronto DPC"}[s]; label != "" {
		return label
	}
	return cleanReportText(s)
}
func reportLevel(s string) string {
	if label := map[string]string{"green": "verde", "yellow": "giallo", "orange": "arancione", "red": "rosso", "not_applicable": "non applicabile"}[s]; label != "" {
		return label
	}
	return cleanReportText(s)
}
func reportState(s string) string {
	if label := map[string]string{"unresolved": "non verificata", "partial": "parziale", "unreliable": "non affidabile", "unknown": "non determinata", "delayed": "in ritardo", "failed": "fallito", "not_verified": "non verificato", "suspended": "sospeso"}[s]; label != "" {
		return label
	}
	return cleanReportText(s)
}

func reportIssuance(s string) string {
	if cleanReportText(s) == "" {
		return "non determinata"
	}
	return cleanReportText(s)
}
func reportTemporalExpired(v publicquery.Temporal, at time.Time, loc *time.Location) bool {
	if v.Precision == "date" && v.Date != nil {
		return *v.Date < at.In(loc).Format(time.DateOnly)
	}
	return v.Precision == "interval" && v.EndInstant != nil && !at.Before(*v.EndInstant)
}
