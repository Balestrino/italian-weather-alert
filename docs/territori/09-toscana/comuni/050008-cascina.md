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
| Categoria Protezione Civile | Elenco tematico con date | [Categoria 119346](https://www.comune.cascina.pi.it/it/news-category/119346?type=3); il 2026-10-02 mostra anche la notizia COC dell'8 ottobre 2024 |

La prima consultazione web del 2026-10-02 non aveva restituito il contenuto dell'archivio. La successiva consultazione dello stesso giorno ha restituito gli Avvisi e la [notizia Allerta arancione, aperto il COC](https://www.comune.cascina.pi.it/it/news/119346/allerta-arancione-aperto-il-coc). Quest'ultima collega nel footer un Piano di miglioramento in PDF su un host S3 esterno. La nuova osservazione supera il limite della prima consultazione, senza dimostrare disponibilità continua del crawler.

## Guida operativa

Verificare raggiungibilità, riconoscimento dei contenuti e completezza come passaggi distinti. Se Crawl4AI segnala anti-bot, confrontare risposta, stato HTTP e contenuto nel registro privato. Una preview storica non esclude un blocco successivo. Rispettare retry e vincoli della fonte; rivalutare il normale accesso prima di cambiare configurazione.

Controllare separatamente archivi generali e tematici, date e paginazione. Una sezione senza notizie recenti non dimostra che il Comune non abbia pubblicato altrove.

Usare il layout revisionato `2 Jan 2006` con normalizzazione dei mesi italiani per le schede con giorni a una o due cifre. La notizia COC mostra `8 ottobre 2024`: la data ufficiale persistita permette di escluderla dalla finestra recente. Le date sconosciute rimangono candidati; riferimenti espliciti e misure ongoing/unresolved mantengono le eccezioni del piano. La data di acquisizione non è quella di pubblicazione. CAS-003 conserva la diagnosi precedente; CAS-004 documenta la correzione verificata.

Selezionare i PDF dentro `page-content`, escludendo footer e navigazione. Gli allegati pertinenti sono anche in contenitori separati del medesimo tipo. Per i PDF municipali collegati dalle pagine ufficiali è stato esaminato separatamente `https://cascina-api.cloud.municipiumapp.it/s3/1520/allegati/`: non autorizza il diverso host S3 del Piano di miglioramento, altri tenant o portali esterni. Rispettare le eccezioni documentali e di terzi delle [note legali](https://www.comune.cascina.pi.it/it/legal_notices); nessuna pubblicazione o copia pubblica è abilitata da questa verifica. Un vero PDF richiesto ma mancante impedisce ancora un controllo completo. Vedere CAS-004 e la [procedura sugli allegati](../../../operations/municipal-attachments.md).

Controllare separatamente raccolta, originali, trigger, coda e provider. Versioni identiche dello stesso documento vengono riusate dal salvataggio; originali diversi rimangono conservati. La policy interpretativa `cascina-html-v1` normalizza soltanto i valori alfanumerici di 40 caratteri nelle forme esatte di meta CSRF e input nascosto verificate; date, testo operativo, collegamenti e allegati restano significativi. Preflight e riuso dei risultati richiedono opt-in della fonte, prove complete, compatibilità e un risultato validato; le copie equivalenti aspettano la loro versione di riferimento senza duplicare l'analisi. Gli input incompleti restano distinti. Vedere CAS-005 e il [piano di efficienza](../../../../openspec/changes/reduce-regolo-token-waste/design.md). Le impostazioni e i blocchi del provider per ambiente rimangono privati.

## Regole ed eccezioni

| ID | Ambito | Regola corrente | Stato dell'intervento |
| --- | --- | --- | --- |
| CAS-001 | Diagnostica di accesso | Il solo `source_http_error` non identifica una causa precisa; servono risposta e diagnostica del crawler | Procedura diagnostica; causa operativa nel registro privato |
| CAS-002 | Footer della notizia COC esaminata | Diagnosi della discovery generale precedente | Superata per il perimetro revisionato da CAS-004; storia conservata |
| CAS-003 | Data della scheda COC nell'elenco tematico | Diagnosi del layout precedente con giorno a due cifre | Superata dalla verifica di CAS-004; date sconosciute ancora ammesse |
| CAS-004 | Elenchi e PDF pertinenti di Cascina | Layout con giorno a una o due cifre, contenitore e permesso API del tenant esaminato | Implementato e verificato; dettagli di sviluppo privati |
| CAS-005 | Riuso interpretativo HTML di Cascina | Normalizzare soltanto le due forme CSRF esaminate | Implementato; test di equivalenza e riuso senza chiamate |

Il permesso per gli allegati esaminati appartiene alla revisione della sola fonte Cascina; questa scheda non abilita fonti, pubblicazione o eccezioni anti-bot.

## Problemi aperti

Verificare l'accesso dell'acquisitore e la riconoscibilità degli elenchi nell'ambiente interessato, includendo eventuali protezioni anti-bot. Chiudere il problema solo con un controllo completo sul perimetro dichiarato. Paginazione, atti originali, aggiornamenti e accettazione rimangono da valutare separatamente.

CAS-004 risolve date e selezione degli allegati nel perimetro esaminato. Monitorare cambiamenti dei contenitori, download e note legali. La discovery PDF non acquisisce automaticamente moduli `.doc` o altri formati: una loro eventuale necessità interpretativa richiede verifica separata. Non inferire completezza degli atti o cancellazione delle misure dalla selezione temporale.

CAS-005 non risolve blocchi o entitlement del provider: un documento nuovo e non equivalente richiede la propria elaborazione. Il riuso con risultato validato è verificato su fixture e PostgreSQL isolato; controlli ordinari senza nuovi contenuti non costituiscono una prova di interpretazione semantica della fonte.

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

### CAS-003 — Il giorno a una cifra può lasciare una notizia storica nel piano

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, browser sulle due pagine ufficiali, parser applicativo e implementazione locale di `time.Parse`.
- **Ambito:** `cascina-municipal`, notizia COC e sua scheda nella categoria 119346; parser con layout `02 Jan 2006` e pianificazione dei target senza data. Nessuna estensione ad altri comuni.
- **Conoscenza:** confermato per data e incompatibilità del layout; stato del tracker per ambiente nel registro privato. **Intervento:** layout alternativo proposto, non applicato.
- **Osservazione:** pagina e scheda mostrano `8 ottobre 2024`; il mese viene tradotto, ma `02` richiede due cifre. `discoverDocuments` conserva il collegamento con data sconosciuta se il parsing fallisce; `Plan` include questi target indipendentemente dall'età.
- **Evidenza:** [notizia ufficiale](https://www.comune.cascina.pi.it/it/news/119346/allerta-arancione-aperto-il-coc), [elenco tematico](https://www.comune.cascina.pi.it/it/news-category/119346?type=3), [parser](../../../../internal/backend/acquisition/listing.go), [pianificazione](../../../../internal/backend/acquisition/tracking.go). I test esistenti degli elenchi coprono giorni a due cifre; il loro esito non verifica la correzione proposta.
- **Conseguenza:** una notizia storica può essere acquisita nuovamente senza essere una nuova pubblicazione. Salvataggio, elaborazione e riuso sono stati distinti; il fallimento degli allegati non implica assenza di un job accodato.
- **Prossima verifica:** fixture con giorni a una e due cifre, revisione della configurazione e confronto della data persistita con l'originale; confronto degli HTML conservati prima di proporre una normalizzazione per Cascina.
- **Collegamenti:** CAS-002 e [specifica di acquisizione](../../../../openspec/changes/define-toscana-alert-service/specs/official-source-ingestion/spec.md); modifica e issue non disponibili.

### CAS-004 — Date e allegati pertinenti nel perimetro revisionato

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, pagine ufficiali, client applicativo, Poppler, fixture e verifica operativa privata.
- **Ambito:** `cascina-municipal`, primi due elenchi Avvisi e categoria storica; PDF comunali collegati dentro `page-content`, origine API Cascina e directory `/s3/1520/allegati/`.
- **Conoscenza:** confermato per i casi esaminati. **Intervento:** configurazione revisionata applicata e verificata in sviluppo; nessuna accettazione pubblica o attivazione di produzione.
- **Osservazione:** il layout `2 Jan 2006` riconosce il giorno COC e i giorni a due cifre; la data persistita esclude il vecchio target senza modificare le eccezioni del piano. Il filtro elimina il PDF del footer ma conserva i veri allegati in sezioni `page-content` separate. La notizia Sogefarm collega decreto sindacale e avviso municipale sul tenant API esaminato; `DirectHTTP` e Poppler ne verificano download e struttura. Una prova con diverso client non aveva scaricato i file e non certificava indisponibilità applicativa.
- **Evidenza:** [notizia COC](https://www.comune.cascina.pi.it/it/news/119346/allerta-arancione-aperto-il-coc), [referral Sogefarm](https://www.comune.cascina.pi.it/it/news/amministratore-unico-sogefarm-aperta-la-manifestazione-d-interesse), [condizioni con eccezioni](https://www.comune.cascina.pi.it/it/legal_notices), [fixture date](../../../../internal/backend/acquisition/listing_test.go), [fixture allegati](../../../../internal/backend/acquisition/scoped_attachments_test.go), [pianificazione persistente](../../../../internal/backend/acquisition/schedule_integration_test.go).
- **Conseguenza:** supera gli interventi proposti in CAS-002 e CAS-003 nel solo perimetro revisionato. Preview, controlli programmati completi e rilettura dei PDF conservati sono verificati privatamente; non garantiscono disponibilità futura o interpretazione.
- **Prossima verifica:** osservare continuità, contenitori e condizioni; valutare separatamente eventuali allegati non PDF, pagine ulteriori e atti originali mancanti.
- **Collegamenti:** CAS-002, CAS-003, CAS-005 e MUN-004; dettagli di configurazione e ripristino nel registro privato.

### CAS-005 — Riuso senza analizzare di nuovo i soli valori CSRF

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, confronto degli originali conservati, fixture sintetiche e PostgreSQL isolato.
- **Ambito:** policy `cascina-html-v1` della sola `cascina-municipal`; nessuna normalizzazione estesa agli altri comuni.
- **Conoscenza:** confermato per le due forme esatte esaminate. **Intervento:** implementato in preflight e manifest interpretativi; opt-in di sviluppo verificato privatamente.
- **Osservazione:** due originali conservati differiscono soltanto nei valori alfanumerici di 40 caratteri del meta `csrf-token` e dell'input hidden `_token`. La policy rimuove soltanto questi valori dall'identità interpretativa e conserva ogni originale. Il test persistente distingue cambi reali e sequenze A/B/A, confini di configurazione/archivio e risorse mancanti; il riuso validato non richiama il provider.
- **Evidenza:** [policy e manifest](../../../../internal/backend/classification/manifest.go), [preflight](../../../../internal/backend/interpretation/preflight.go), [regressioni](../../../../internal/backend/classification/manifest_test.go), [test PostgreSQL](../../../../internal/backend/interpretation/preflight_integration_test.go). Confronti operativi e valori originali rimangono privati.
- **Conseguenza:** evita nuovi job per copie complete equivalenti mentre attendono il rappresentante; dopo un risultato compatibile validato ne riusa l'interpretazione. Un input nuovo, incompleto o diverso resta lavoro indipendente. Non sblocca il provider.
- **Prossima verifica:** monitorare forme dei tag e risultati di riuso; qualsiasi nuova forma di normalizzazione richiede evidenza e una nuova policy.
- **Collegamenti:** CAS-004 e [verifica operativa](../../../operations/municipal-attachments.md).
