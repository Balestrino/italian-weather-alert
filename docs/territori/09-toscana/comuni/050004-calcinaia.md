---
tipo: "comune"
codice_regione: "09"
codice_istat: "050004"
ultima_revisione: "2026-10-02"
---

# Calcinaia

## Identità e ambito

Comune di Calcinaia, provincia di Pisa, ISTAT `050004`. Fonte IWA: `calcinaia-municipal`. Collegamenti: [Toscana](../README.md), [stato pubblico documentato](../../../coverage.md#region-09). La scheda descrive il perimetro di ricerca, senza certificare completezza o attivazione.

## Mappa delle fonti

| Fonte | Ruolo | Riferimento e limiti |
| --- | --- | --- |
| Sito comunale | Primaria | [Notizie](https://www.comune.calcinaia.pi.it/tipi-di-notizia/notizie), [Avvisi](https://www.comune.calcinaia.pi.it/tipi-di-notizia/avvisi), [Comunicati](https://www.comune.calcinaia.pi.it/tipi-di-notizia/comunicati); sezioni da verificare nella configurazione |
| Cittadino Informato | Diagnostica secondaria | Ruolo definito dalla specifica; non stabilisce da solo completezza, date o misure |
| Albo pretorio | Atti di supporto | Cercare gli atti citati nelle notizie primarie; il riferimento al singolo atto va verificato |

La mappa consolida la [specifica di acquisizione](../../../../openspec/changes/define-toscana-alert-service/specs/official-source-ingestion/spec.md). La sezione Notizie è stata consultata il 2026-10-02; ciò non verifica tutte le altre sezioni o i canali secondari.

## Guida operativa

Confrontare le sezioni configurate e la paginazione: un elenco generale non dimostra l'esaustività degli elenchi tematici. Distinguere data di pubblicazione, aggiornamento della pagina e validità di una misura. Acquisire e interpretare gli atti necessari prima di determinare durata o cessazione di chiusure.

Per documenti 404/410 usare la [procedura di recupero](../../../operations/acquisition-recovery.md): assenza dagli elenchi completamente traversati, riferimenti e misure ancora aperte vanno verificati prima di una esclusione. Conservare lo storico; un'assenza non prova revoca né equivalenza a un altro URL.

## Regole ed eccezioni

| ID | Ambito | Regola corrente | Stato dell'intervento |
| --- | --- | --- | --- |
| CLN-001 | Gerarchia delle fonti | Sito comunale primario; Cittadino Informato come confronto; albo per atti citati | Regola di progetto documentata; verifiche di copertura separate |
| CLN-002 | Target 404/410 | Esclusione solo con decisione motivata per fonte/configurazione/URL; rediscovery ripristina il controllo | Procedura e codice presenti; applicazione al singolo target da verificare |

## Problemi aperti

Trial e accettazione restano quelli del coverage tracker. Le lacune dei canali secondari vanno registrate separatamente dalle omissioni nel perimetro primario. Consultare il registro privato per errori e recuperi dell'ambiente; non considerarli una certificazione pubblica.

## Registro delle scoperte

### CLN-001 — Il canale secondario non dimostra la copertura primaria

- **Data:** consolidamento 2026-10-02; data originaria non disponibile.
- **Ultima verifica:** 2026-10-02, documentale sulla specifica e sul coverage tracker.
- **Ambito:** `calcinaia-municipal`, canali primario e secondari.
- **Conoscenza:** confermato come regola di progetto. **Intervento:** non necessario per questa documentazione.
- **Osservazione:** la specifica assegna al sito comunale il ruolo primario e limita l'uso dei canali secondari.
- **Evidenza:** [perimetro comunale nella specifica](../../../../openspec/changes/define-toscana-alert-service/specs/official-source-ingestion/spec.md).
- **Conseguenza:** una discrepanza del canale secondario avvia una verifica sulla fonte primaria, senza inventare misure o completezza.
- **Prossima verifica:** confrontare campioni, sezioni e periodo del prossimo percorso di accettazione.

### CLN-002 — Un documento indisponibile richiede revisione prima dell'esclusione

- **Data:** consolidamento 2026-10-02; data originaria non disponibile.
- **Ultima verifica:** 2026-10-02, documentale sulla procedura.
- **Ambito:** target municipali 404/410; non è un'autorizzazione a escludere un URL specifico.
- **Conoscenza:** confermato. **Intervento:** procedura applicabile, decisione sul target separata.
- **Osservazione:** la procedura richiede revisione degli elenchi e dei riferimenti prima di escludere uno specifico target indisponibile.
- **Evidenza:** [recupero dell'acquisizione](../../../operations/acquisition-recovery.md).
- **Conseguenza:** conservare gli errori e valutare presenza negli elenchi, riferimenti e misure prima della disposizione.
- **Prossima verifica:** nuova acquisizione completa dopo la decisione; dettagli operativi nel registro privato.
