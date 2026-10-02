---
tipo: "comune"
codice_regione: "09"
codice_istat: "050008"
ultima_revisione: "2026-10-02"
---

# Cascina

## Identità e ambito

Comune di Cascina, provincia di Pisa, ISTAT `050008`. Fonte IWA: `cascina-municipal`. Collegamenti: [Toscana](../README.md), [Municipium](../../../fonti/piattaforme/municipium.md), [stato pubblico documentato](../../../coverage.md#region-09). La candidatura e la preview indicate nel tracker non certificano disponibilità continua o completezza.

## Mappa delle fonti

| Fonte | Ruolo | Riferimento e limiti |
| --- | --- | --- |
| Sito comunale | Punto di partenza ufficiale | [Sito del Comune](https://www.comune.cascina.pi.it/) |
| Archivio Avvisi | Sezione da verificare | [Avvisi, pagina 1](https://www.comune.cascina.pi.it/it/news?type=3&page=1); verificare le pagine effettivamente configurate e la sezione tematica |

La prima consultazione web del 2026-10-02 non aveva restituito il contenuto dell'archivio. La successiva consultazione dello stesso giorno ha restituito gli Avvisi e la [notizia Allerta arancione, aperto il COC](https://www.comune.cascina.pi.it/it/news/119346/allerta-arancione-aperto-il-coc). Quest'ultima collega nel footer un Piano di miglioramento in PDF su un host S3 esterno. La nuova osservazione supera il limite della prima consultazione, senza dimostrare disponibilità continua del crawler.

## Guida operativa

Verificare raggiungibilità, riconoscimento dei contenuti e completezza come passaggi distinti. Se Crawl4AI segnala anti-bot, confrontare risposta, stato HTTP e contenuto nel registro privato. Una preview storica non esclude un blocco successivo. Rispettare retry e vincoli della fonte; rivalutare il normale accesso prima di cambiare configurazione.

Controllare separatamente archivi generali e tematici, date e paginazione. Una sezione senza notizie recenti non dimostra che il Comune non abbia pubblicato altrove.

Distinguere i documenti della notizia dai PDF generici del footer. Il collector attuale considera anche questi ultimi allegati obbligatori e registra quelli esterni come `forbidden` prima del download. Nel percorso programmato questo errore interrompe il giro dei documenti: target scoperti negli elenchi non equivalgono a originali acquisiti. Un accesso riuscito agli elenchi non risolve il problema degli allegati; verificare entrambe le fasi.

## Regole ed eccezioni

| ID | Ambito | Regola corrente | Stato dell'intervento |
| --- | --- | --- | --- |
| CAS-001 | Diagnostica di accesso | Il solo `source_http_error` non identifica una causa precisa; servono risposta e diagnostica del crawler | Procedura diagnostica; causa operativa nel registro privato |
| CAS-002 | Footer della notizia COC esaminata | Il Piano di miglioramento esterno viene trattato come allegato obbligatorio e rifiutato dal collector | Comportamento applicato; selezione degli allegati pertinenti proposta |

Nessuna eccezione ai domini, agli allegati o ai controlli anti-bot viene introdotta da questa scheda.

## Problemi aperti

Verificare l'accesso dell'acquisitore e la riconoscibilità degli elenchi nell'ambiente interessato, includendo eventuali protezioni anti-bot. Chiudere il problema solo con un controllo completo sul perimetro dichiarato. Paginazione, atti originali, aggiornamenti e accettazione rimangono da valutare separatamente.

CAS-002 richiede una selezione dei PDF pertinenti alla notizia, verificata con fixture e poi nel perimetro effettivo della fonte. L'esclusione di un documento generico non deve nascondere un vero allegato necessario. Dettagli e alternanza degli esiti operativi restano nel registro privato; nessuna correzione è stata applicata durante il ricontrollo.

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

### CAS-002 — Un PDF del footer può interrompere l'acquisizione dei documenti

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, pagina ufficiale e codice; esiti operativi conservati privatamente.
- **Ambito:** `cascina-municipal`, notizia COC collegata nella mappa, PDF `piano_miglioramento_anthesi.pdf` sul percorso S3 `/s3/1520/allegati/`; percorso di acquisizione programmata.
- **Conoscenza:** confermato per il collegamento e il comportamento del codice. **Intervento:** correzione proposta, non applicata.
- **Osservazione:** il footer collega un Piano di miglioramento esterno. `linkedPDFs` esamina tutta la pagina; `retainPlanned` rende il PDF obbligatorio, lo registra come `forbidden` e restituisce un errore che interrompe `scheduledAcquire` prima dei target successivi.
- **Evidenza:** [notizia ufficiale](https://www.comune.cascina.pi.it/it/news/119346/allerta-arancione-aperto-il-coc), [collector](../../../../internal/backend/acquisition/preview.go), [fixture degli allegati esterni](../../../../internal/backend/acquisition/attachments_test.go). Il riferimento non verifica il contenuto o il riuso del PDF.
- **Conseguenza:** la raccolta può rimanere incompleta anche quando gli elenchi sono riconosciuti; CAS-001 non descrive tutte le possibili cause del fallimento.
- **Prossima verifica:** fixture con PDF del footer e allegato pertinente, quindi controllo completo della fonte e confronto fra target pianificati e originali conservati.
- **Collegamenti:** CAS-001 e MUN-002; modifica e issue non disponibili.
