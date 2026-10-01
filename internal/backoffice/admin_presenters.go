package backoffice

import "strings"

func columnLabel(key string) string {
	if v, ok := map[string]string{"evidence_role": "Ruolo della pagina", "equivalent_versions": "Copie equivalenti in riuso", "equivalent_to_version": "Versione di riferimento", "job_id": "Job", "source_id": "Fonte", "kind": "Tipo", "state": "Stato", "attempt_count": "Tentativi", "last_error_code": "Ultimo errore", "available_at": "Disponibile dal", "document_version_id": "Versione documento", "run_id": "Elaborazione", "updating_state": "Aggiornamento", "collection_enabled": "Raccolta", "public_enabled": "Pubblicazione", "last_complete_at": "Ultimo controllo completo", "classification_status": "Classificazione", "extraction_status": "Estrazione", "first_acquired_at": "Acquisito il", "recorded_at": "Registrato il", "detail": "Descrizione", "outcome": "Esito", "duration_ms": "Durata (ms)", "estimated_cost_microunits": "Costo stimato (milionesimi)", "known_cost_microunits": "Subtotale noto (milionesimi)", "cost_complete": "Costo completo", "pricing_provenance_url": "Provenienza prezzo", "currency": "Valuta", "workload": "Attività", "stage": "Fase", "model": "Modello", "provider": "Provider", "attempt": "Tentativo", "usage_status": "Disponibilità consumi", "cost_status": "Disponibilità costo", "regional_source_id": "Fonte regionale", "dpc_source_id": "Fonte DPC", "compared_at": "Confrontato il", "scope": "Perimetro"}[key]; ok {
		return v
	}
	return strings.ReplaceAll(key, "_", " ")
}
func primaryColumns(section string, available []string) []string {
	wanted := map[string][]string{"sources": {"source_id", "updating_state", "collection_enabled", "public_enabled", "last_complete_at"}, "documents": {"document_version_id", "source_id", "classification_status", "extraction_status", "equivalent_versions", "first_acquired_at"}, "findings": {"kind", "source_id", "state", "detail", "recorded_at"}, "dpc-comparisons": {"id", "regional_source_id", "dpc_source_id", "scope", "compared_at"},
		"usage": {"run_id", "source_id", "workload", "stage", "model", "attempt", "outcome", "duration_ms", "cost_status"},
		"jobs":  {"job_id", "source_id", "kind", "state", "attempt_count", "last_error_code"}}[section]
	if len(wanted) == 0 {
		return available
	}
	var result []string
	for _, k := range wanted {
		for _, a := range available {
			if a == k {
				result = append(result, k)
				break
			}
		}
	}
	if len(result) == 0 {
		return available
	}
	return result
}
func statusLabel(s string) string {
	if v, ok := map[string]string{"failed": "Fallito", "retry_wait": "Nuovo tentativo in attesa", "queued": "In coda", "running": "In corso", "succeeded": "Completato", "delayed": "In ritardo", "current": "Aggiornato", "not_yet_verified": "Non ancora verificato"}[s]; ok {
		return v
	}
	return s
}
