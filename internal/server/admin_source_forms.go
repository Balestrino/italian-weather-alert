package server

import (
	"encoding/json"
	"github.com/Balestrino/italian-weather-alert/internal/registry"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type guidedField struct {
	Name, Label, Kind, Value, Hint string
	Required, Readonly             bool
}

func field(name, label, kind string, required bool) guidedField {
	return guidedField{Name: name, Label: label, Kind: kind, Required: required}
}
func sourceFields(action string, v adminSourceView) []guidedField {
	revision := v.State.LatestRevision
	if action != "preview" && action != "enable-collection" && v.State.ActiveRevision != nil {
		revision = *v.State.ActiveRevision
	}
	rev := field("revision", "Revisione selezionata", "number", true)
	rev.Value = strconv.Itoa(revision)
	rev.Readonly = true
	actor := field("actor", "Operatore", "text", true)
	fields := []guidedField{rev, actor}
	switch action {
	case "preview", "enable-collection", "suspend-collection", "enable-public", "suspend-public":
	case "intervals":
		fields[0].Name = "expected_revision"
		fields[0].Label = "Revisione intervalli"
		fields[0].Value = strconv.Itoa(v.Intervals.Revision)
		check := field("check_seconds", "Controllo ogni (secondi)", "number", true)
		check.Value = strconv.Itoa(v.EffectiveCheckSeconds)
		delay := field("delay_seconds", "Soglia ritardo (secondi)", "number", true)
		delay.Value = strconv.Itoa(v.EffectiveDelaySeconds)
		fields = append(fields, check, delay)
	case "review-acceptance":
		fields = append(fields, evidenceFields("evidence", "Revisione finale")...)
	case "suspend-interpretation":
		fields = append(fields, evidenceFields("defect", "Difetto confermato")...)
	case "accept":
		fields = append(fields, evidenceFields("acceptance.report", "Rapporto di accettazione")...)
		fields = append(fields, field("acceptance.period_start", "Inizio osservazione", "time", true), field("acceptance.period_end", "Fine osservazione", "time", true), field("acceptance.sections", "Sezioni verificate", "lines", true), field("acceptance.risk_coverage", "Rischi coperti", "lines", false), field("acceptance.coverage_status", "Copertura", "coverage", true), field("acceptance.coverage_limitations", "Limitazioni dichiarate", "lines", false))
		for _, check := range []struct{ name, label string }{{"extraction_verified", "Estrazione verificata"}, {"updates_verified", "Aggiornamenti verificati"}, {"attachments_verified", "Allegati verificati"}, {"scanned_attachments_verified", "Scansioni verificate"}, {"history_verified", "Storico verificato"}, {"failure_behavior_verified", "Comportamento in errore verificato"}, {"interfaces_equivalent", "Equivalenza API/MCP verificata"}} {
			fields = append(fields, field("acceptance."+check.name, check.label, "bool", true))
		}
		fields = append(fields, field("acceptance.known_omissions", "Omissioni note", "number", true), field("acceptance.unsupported_assertions", "Asserzioni non supportate", "number", true))
	case "reprocess-interpretation":
		fields = append(fields, field("from", "Acquisiti dal (incluso)", "time", false), field("through", "Acquisiti fino al", "time", false), field("error_code", "Codice errore da selezionare", "text", false), field("processing_evaluation_id", "Valutazione elaborazione (se applicabile)", "text", false))
	case "resume-interpretation":
		fields = append(fields, field("recovery.suspension_id", "ID sospensione", "number", true))
		fields = append(fields, evidenceFields("recovery.correction", "Correzione")...)
		fields = append(fields, evidenceFields("recovery.validation", "Validazione")...)
		fields = append(fields, field("recovery.reprocessing_id", "ID rielaborazione completata", "text", true), field("recovery.known_omissions", "Omissioni note", "number", true), field("recovery.unsupported_assertions", "Asserzioni non supportate", "number", true), field("recovery.verified", "Validazione verificata", "bool", true))
	default:
		return nil
	}
	for i := range fields {
		switch fields[i].Kind {
		case "time":
			fields[i].Hint = "Data e ora con fuso, es. 2026-09-18T14:30:00Z"
		case "lines":
			fields[i].Hint = "Un valore per riga"
		}
	}
	return fields
}
func sourceGuidedPayload(r *http.Request) (string, error) {
	action := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	fields := sourceFields(action, adminSourceView{})
	if len(fields) == 0 || !strings.HasPrefix(r.URL.Path, "/admin/sources/") {
		return "", registry.ErrInvalid
	}
	values := map[string]any{}
	allowed := map[string]bool{}
	for _, f := range fields {
		allowed[f.Name] = true
		entries := r.PostForm[f.Name]
		if len(entries) > 1 || (f.Required && len(entries) != 1) {
			return "", registry.ErrInvalid
		}
		raw := strings.TrimSpace(r.PostForm.Get(f.Name))
		if f.Required && raw == "" {
			return "", registry.ErrInvalid
		}
		var value any = raw
		switch f.Kind {
		case "number":
			n, err := strconv.ParseInt(raw, 10, 64)
			if err != nil || n < 0 {
				return "", registry.ErrInvalid
			}
			value = n
		}
		if f.Kind == "bool" {
			if raw != "true" && raw != "false" {
				return "", registry.ErrInvalid
			}
			value = raw == "true"
		}
		if f.Kind == "time" {
			if raw == "" {
				value = nil
			} else {
				if _, err := time.Parse(time.RFC3339, raw); err != nil {
					return "", registry.ErrInvalid
				}
			}
		}
		if f.Kind == "lines" {
			lines := []string{}
			for _, line := range strings.Split(raw, "\n") {
				if line = strings.TrimSpace(line); line != "" {
					lines = append(lines, line)
				}
			}
			value = lines
		}
		target := values
		parts := strings.Split(f.Name, ".")
		for _, part := range parts[:len(parts)-1] {
			if target[part] == nil {
				target[part] = map[string]any{}
			}
			target = target[part].(map[string]any)
		}
		target[parts[len(parts)-1]] = value
	}
	for key := range r.PostForm {
		if !allowed[key] {
			return "", registry.ErrInvalid
		}
	}
	if action == "reprocess-interpretation" && r.PostForm.Get("from") == "" && r.PostForm.Get("through") == "" && strings.TrimSpace(r.PostForm.Get("error_code")) == "" {
		return "", registry.ErrInvalid
	}
	b, err := json.Marshal(values)
	return string(b), err
}

func evidenceFields(prefix, label string) []guidedField {
	return []guidedField{field(prefix+".url", label+" · URL", "text", true), field(prefix+".locator", label+" · Riferimento", "text", true), field(prefix+".observed_at", label+" · Osservato il", "time", true)}
}
