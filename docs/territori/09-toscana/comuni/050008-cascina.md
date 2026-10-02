---
tipo: "comune"
codice_regione: "09"
codice_istat: "050008"
ultima_revisione: "2026-10-02"
---

# Cascina

## Identità e ambito

Comune di Cascina, provincia di Pisa, ISTAT `050008`. Fonte IWA: `cascina-municipal`. Collegamenti: [Toscana](../README.md), [stato pubblico documentato](../../../coverage.md#region-09). La candidatura e la preview indicate nel tracker non certificano disponibilità continua o completezza.

## Mappa delle fonti

| Fonte | Ruolo | Riferimento e limiti |
| --- | --- | --- |
| Sito comunale | Punto di partenza ufficiale | [Sito del Comune](https://www.comune.cascina.pi.it/) |
| Archivio Avvisi | Sezione da verificare | [Avvisi, pagina 1](https://www.comune.cascina.pi.it/it/news?type=3&page=1); verificare le pagine effettivamente configurate e la sezione tematica |

La consultazione web del 2026-10-02 non ha restituito il contenuto dell'archivio. Questo esito dello strumento non stabilisce da solo la causa del problema o un'indisponibilità generale del sito.

## Guida operativa

Verificare raggiungibilità, riconoscimento dei contenuti e completezza come passaggi distinti. Se Crawl4AI segnala anti-bot, confrontare risposta, stato HTTP e contenuto nel registro privato. Una preview storica non esclude un blocco successivo. Rispettare retry e vincoli della fonte; rivalutare il normale accesso prima di cambiare configurazione.

Controllare separatamente archivi generali e tematici, date e paginazione. Una sezione senza notizie recenti non dimostra che il Comune non abbia pubblicato altrove.

## Regole ed eccezioni

| ID | Ambito | Regola corrente | Stato dell'intervento |
| --- | --- | --- | --- |
| CAS-001 | Diagnostica di accesso | Il solo `source_http_error` non identifica una causa precisa; servono risposta e diagnostica del crawler | Procedura diagnostica; causa operativa nel registro privato |

Nessuna eccezione ai domini, agli allegati o ai controlli anti-bot viene introdotta da questa scheda.

## Problemi aperti

Verificare l'accesso dell'acquisitore e la riconoscibilità degli elenchi nell'ambiente interessato, includendo eventuali protezioni anti-bot. Chiudere il problema solo con un controllo completo sul perimetro dichiarato. Paginazione, atti originali, aggiornamenti e accettazione rimangono da valutare separatamente.

## Registro delle scoperte

### CAS-001 — Separare l'errore di raccolta dalla sua causa

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, documentale e tentativo di consultazione web non conclusivo.
- **Ambito:** `cascina-municipal`, archivio Avvisi e diagnostica del crawler.
- **Conoscenza:** limite della verifica confermato; disponibilità e causa correnti non certificate pubblicamente.
- **Intervento:** verifica operativa aperta; nessuna correzione applicata con queste schede.
- **Osservazione:** il tentativo di consultazione non ha restituito il contenuto; una diagnosi del crawler richiede evidenze ulteriori.
- **Evidenza:** [CrawlFailure e client Crawl4AI](../../../../internal/backend/acquisition/crawl4ai.go), [procedura sui controlli completi](../../../operations/acquisition-recovery.md). Diagnostica dettagliata conservata nel registro privato.
- **Conseguenza:** non equiparare un errore di consultazione a sito dismesso, assenza di avvisi o errore HTTP specifico.
- **Prossima verifica:** ispezionare la risposta del normale percorso di acquisizione e verificare un controllo completo, rispettando i vincoli della fonte.
