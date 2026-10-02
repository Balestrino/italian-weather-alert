---
tipo: "comune"
codice_regione: "09"
codice_istat: "049009"
ultima_revisione: "2026-10-02"
---

# Livorno

## Identità e ambito

Comune di Livorno, provincia di Livorno, ISTAT `049009`. Fonte IWA: `livorno-municipal`. Collegamenti: [Toscana](../README.md), [Municipium](../../../fonti/piattaforme/municipium.md), [stato pubblico documentato](../../../coverage.md#region-09). Le scoperte qui descritte non certificano la raccolta di tutti i canali municipali.

## Mappa delle fonti

| Fonte | Ruolo | Riferimento e limiti |
| --- | --- | --- |
| Sito comunale | Primaria per le notizie esaminate | [Categoria Protezione Civile](https://www.comune.livorno.it/it/news-category/133640); verificare paginazione e perimetro nell'ambiente |
| Risorse Municipium | Allegati collegati dal Comune | `livorno-api.cloud.municipiumapp.it`, solo `/s3/3612/allegati/`; regola esplicita per fonte/revisione, senza ammissione di altri host |
| Altri portali comunali | Ricerca separata | Un collegamento nell'interfaccia non ne include automaticamente tutti gli atti nella raccolta |

La [notizia comunale esaminata](https://www.comune.livorno.it/it/news/133640/allerta-arancio-per-forti-temporali-con-rischio-idrogeologico-e-idraulico-del-reticolo-minore) collega l'ordinanza n. 303 del 20 agosto 2026 nella sezione Allegati e una guida alla navigazione nel footer. I due PDF sono su domini Municipium distinti da quello della pagina. Consultazione documentale: 2026-10-02; il contenuto può cambiare.

## Guida operativa

Distinguere gli allegati della notizia dai documenti generici del sito. Verificare ogni dominio esterno rispetto al riferimento ufficiale e ai vincoli di raccolta/riuso. Una risorsa `forbidden` può essere esclusa dal collector prima di qualsiasi richiesta HTTP: non prova un 403 o un file rimosso.

Il [collector](../../../../internal/backend/acquisition/preview.go) supporta ora una policy `attachments` per fonte/revisione: `content_class` limita la discovery dei PDF alla parte pertinente del documento, `external` ammette origini HTTPS e prefissi di percorso verificati e `validate_pdf` richiede un PDF valido. Le fonti senza questa policy mantengono la scansione generale e il rifiuto degli allegati esterni. Un allegato necessario non recuperato mantiene il controllo incompleto e il riferimento conservato.

Il ricontrollo del 2026-10-02 ha rivalutato LIV-001 e LIV-002 mediante pagina ufficiale, codice e un nuovo giro programmato. In quella fase nessuna correzione era stata introdotta; LIV-006 registra la successiva implementazione e verifica. Nell'acquisizione programmata l'errore ferma il giro prima dei documenti successivi: distinguere target scoperti e originali conservati; esiti per ambiente nel registro privato. Vedere LIV-003.

## Regole ed eccezioni

| ID | Ambito | Regola corrente | Stato dell'intervento |
| --- | --- | --- | --- |
| LIV-001 | PDF della pagina esaminata | Allegati esterni ammessi solo nel perimetro HTTPS/percorso revisionato; altri riferimenti restano `forbidden` | Correzione implementata e verificata sul perimetro; LIV-006 |
| LIV-002 | Discovery degli allegati | Selezione entro `article-content`, con esclusione di footer, navigazione e intestazioni | Correzione implementata e verificata sul perimetro; LIV-006 |

L'eventuale ammissione di risorse Municipium deve essere esplicita e limitata alla fonte e ai percorsi verificati. La nota di piattaforma non costituisce una autorizzazione generale.

La rivalutazione LIV-004 ha identificato come perimetro tecnico il solo host HTTPS `livorno-api.cloud.municipiumapp.it`, percorso `/s3/3612/allegati/`, per gli allegati pertinenti collegati dalle notizie comunali. Il riferimento della pagina esaminata compare dentro `article-content`; la guida generica è nel footer. `cloud-ita.municipiumapp.it` non richiede un'abilitazione per raccogliere quell'ordinanza. Sono da verificare separatamente l'effettivo accesso al file e le condizioni di ciascun documento; il collector dispone ora di una policy esplicita per autorizzare questo perimetro, con referral e riuso separati.

## Problemi aperti

LIV-001 e LIV-002 hanno una correzione nel checkout con test di confine, confronto del filtro con la pagina reale e verifica della conservazione; vedere LIV-006. Restano da verificare continuità dei download, eventuali nuovi host o eccezioni dei documenti e copertura oltre la prima pagina configurata. Trial, interpretazione e accettazione restano separati.

LIV-005 ha corretto il limite di accesso registrato in LIV-004: la prima prova HTTP non dimostrava indisponibilità per tutti i client. Il file esatto è stato acquisito con un normale browser e con `DirectHTTP`, e validato come PDF leggibile; dettagli dei trasporti ed esiti nell'archivio privato. Le [note legali comunali](https://www.comune.livorno.it/it/legal_notices) dichiarano CC BY 4.0 salvo eccezioni e contenuti di terzi; la clausola sui collegamenti esterni non concede da sola i diritti su ogni risorsa di un host.

Per verificare il recupero usare il trasporto dell'applicazione nell'ambiente interessato, conservare il risultato e validare il PDF, oltre allo stato HTTP. Un esito del browser o di un altro client non sostituisce la verifica di `DirectHTTP`. Prima di considerare completo un giro verificare anche persistenza, riferimenti e tutti gli allegati necessari. Gestire i fallimenti temporanei con retry limitati e cadenze della fonte; un rifiuto persistente richiede diagnosi o un canale ufficiale alternativo verificato.

## Registro delle scoperte

Le voci LIV-001–LIV-005 descrivono gli stati osservati nelle rispettive fasi del 2026-10-02; LIV-006 aggiorna il comportamento corrente dopo la correzione.

### LIV-001 — Ordinanza pubblicata su risorse Municipium

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, pagina ufficiale e codice del collector; disponibilità del PDF non verificata qui.
- **Ambito:** `livorno-municipal`, notizia esaminata e host `livorno-api.cloud.municipiumapp.it`.
- **Conoscenza:** confermato per il collegamento e il vincolo nel codice. **Intervento:** modifica proposta.
- **Osservazione:** l'allegato della notizia è su un host diverso; il collector rifiuta queste risorse prima del download.
- **Evidenza:** notizia collegata nella mappa; [collector](../../../../internal/backend/acquisition/preview.go) e [fixture sugli allegati esterni](../../../../internal/backend/acquisition/attachments_test.go).
- **Conseguenza:** il riferimento esterno necessario può rendere incompleto il controllo anche quando la pagina è leggibile.
- **Prossima verifica:** accesso/riuso, ammissione limitata e acquisizione del PDF con la modifica valutata.

### LIV-002 — Guida del sito inclusa fra gli allegati obbligatori

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, pagina ufficiale e codice.
- **Ambito:** footer della stessa notizia; `cloud-ita.municipiumapp.it`.
- **Conoscenza:** confermato per la pagina esaminata. **Intervento:** correzione proposta.
- **Osservazione:** la pagina contiene un PDF nel footer, che la scansione generale dei link tratta come dipendenza obbligatoria.
- **Evidenza:** link Guida alla navigazione della notizia e scansione dell'intero HTML in `linkedPDFs`.
- **Conseguenza:** un documento generico estraneo alla misura diventa una dipendenza obbligatoria del controllo.
- **Prossima verifica:** valutare una selezione degli allegati pertinenti, conservando quelli necessari alla misura; dati delle esecuzioni nel registro privato.

### LIV-003 — Rivalutazione dei due problemi degli allegati

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, pagina ufficiale e codice; nuovo ricontrollo operativo conservato privatamente.
- **Ambito:** stessa notizia di LIV-001 e LIV-002, percorso programmato del collector.
- **Conoscenza:** confermato per il collegamento e il comportamento del codice. **Intervento:** correzioni ancora proposte, non applicate.
- **Osservazione:** i collegamenti all'ordinanza e al PDF del footer restano presenti. Il rifiuto dell'allegato rende incompleta la versione e interrompe `scheduledAcquire` prima dei target successivi.
- **Evidenza:** [notizia ufficiale](https://www.comune.livorno.it/it/news/133640/allerta-arancio-per-forti-temporali-con-rischio-idrogeologico-e-idraulico-del-reticolo-minore), [collector](../../../../internal/backend/acquisition/preview.go); dettagli del nuovo giro nel registro privato.
- **Conseguenza:** confermare il download della pagina principale o la discovery degli elenchi non basta a chiudere LIV-001 e LIV-002.
- **Prossima verifica:** conservare i criteri di chiusura originari e verificare tutti i documenti pianificati dopo una eventuale correzione.

### LIV-004 — Perimetro proposto per gli allegati esterni e accesso ancora aperto

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, pagina ufficiale, note legali e codice; prova di accesso conservata privatamente.
- **Ambito:** ordinanza n. 303 collegata dalla stessa notizia di LIV-001, host `livorno-api.cloud.municipiumapp.it`, percorso `/s3/3612/allegati/`; nessuna autorizzazione generale ad altri host o comuni.
- **Conoscenza:** confermato per il riferimento, la posizione dei link e l'assenza di un'opzione del collector; accessibilità corrente non certificata. **Intervento:** perimetro e modifica proposti, non applicati.
- **Osservazione:** l'ordinanza è collegata nella parte `article-content`, mentre la guida alla navigazione è nel footer. `retainPlanned` impone ancora lo stesso schema/host della pagina; `DirectHTTP` impedisce redirect fuori dall'origine del file richiesto. Il solo cambiamento della configurazione della fonte non può superare il controllo attuale.
- **Evidenza:** [notizia ufficiale](https://www.comune.livorno.it/it/news/133640/allerta-arancio-per-forti-temporali-con-rischio-idrogeologico-e-idraulico-del-reticolo-minore), [note legali](https://www.comune.livorno.it/it/legal_notices), [collector](../../../../internal/backend/acquisition/preview.go), [trasporto delle risorse](../../../../internal/backend/acquisition/direct_http.go). Risposta dettagliata del fetch nel registro privato.
- **Conseguenza:** valutare una regola per fonte/revisione, host esatto e percorso degli allegati, con selezione della parte pertinente della pagina; la raggiungibilità del file rimane una verifica distinta dall'ammissione tecnica.
- **Prossima verifica:** validazione del perimetro con fixture di confine, revisione delle eventuali eccezioni del documento, download ordinario riuscito e controllo completo della fonte. Chiudere LIV-001 e LIV-002 soltanto dopo i rispettivi criteri di chiusura.
- **Collegamenti:** LIV-001, LIV-002 e LIV-003; modifica e issue non disponibili.

### LIV-005 — Il download va verificato con il trasporto effettivo

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, normale browser, `DirectHTTP` del checkout e validazione del file; prove dettagliate private.
- **Ambito:** solo URL dell'ordinanza n. 303 collegata dalla notizia di LIV-001; client del checkout eseguito sul nodo e nel container Crawl4AI di development. Nessun nuovo giro del collector o modifica dei domini ammessi.
- **Conoscenza:** confermato per l'accesso al file verificato; causa della differenza fra client non determinata. **Intervento:** verifica eseguita, integrazione proposta e non applicata.
- **Osservazione:** il click ordinario sul link della pagina ha prodotto un download completato; il file è un PDF con pagine e testo estraibile. Anche `DirectHTTP` ha acquisito il PDF. La prima prova registrata in LIV-004 aveva usato un altro client HTTP e non certificava il risultato del trasporto applicativo.
- **Evidenza:** [notizia ufficiale e allegato](https://www.comune.livorno.it/it/news/133640/allerta-arancio-per-forti-temporali-con-rischio-idrogeologico-e-idraulico-del-reticolo-minore), [DirectHTTP](../../../../internal/backend/acquisition/direct_http.go). Risposte, file originale, hash e validazione conservati privatamente.
- **Conseguenza:** supera il limite di verifica dell'accessibilità in LIV-004 per questo file e questi tentativi; restano aperti LIV-001 e LIV-002, la persistenza attraverso il collector e l'accesso futuro. Il successo non certifica tutti gli allegati dello stesso host.
- **Prossima verifica:** integrare il perimetro autorizzato e il filtro dei link pertinenti, verificare gli originali conservati in un nuovo controllo completo e osservare la continuità del recupero.
- **Collegamenti:** LIV-004; probe privato senza modifica al codice applicativo, issue e integrazione non disponibili.


### LIV-006 — Allegati revisionati e filtro del contenuto verificati

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, fixture sintetiche, parser PDF dell'immagine applicativa, PostgreSQL isolato e verifica operativa privata.
- **Ambito:** fonte `livorno-municipal`, prima pagina della categoria configurata e allegati PDF pertinenti entro `article-content`; origine HTTPS `livorno-api.cloud.municipiumapp.it` e solo `/s3/3612/allegati/`. Ambiente e revisione attivati sono documentati nel registro privato.
- **Conoscenza:** confermato per questo perimetro. **Intervento:** implementato e verificato; continuità, interpretazione e accettazione rimangono separate.
- **Osservazione:** la policy della revisione autorizza gli allegati esterni verificati, esclude la guida del footer e richiede tipo/struttura PDF e parsing Poppler. Preview e percorso programmato condividono i controlli e i metadati di pubblicazione. Le risorse mancanti impediscono la completezza; il percorso esterno e i redirect sono limitati prima della richiesta.
- **Evidenza:** [policy e confini](../../../../internal/backend/registry/attachments.go), [collector](../../../../internal/backend/acquisition/preview.go), [fixture del perimetro](../../../../internal/backend/acquisition/scoped_attachments_test.go), [persistenza e recupero](../../../../internal/backend/acquisition/scoped_attachments_integration_test.go), [procedura di verifica](../../../operations/municipal-attachments.md). Configurazione attiva, ricevute, originali e verifiche di rilettura conservati privatamente.
- **Conseguenza:** chiude i criteri di correzione di LIV-001 e LIV-002 sul perimetro verificato; gli altri comuni e i domini generici non ereditano la regola. Un download riuscito non garantisce l'accessibilità futura o la copertura di altri canali.
- **Prossima verifica:** osservare la continuità del recupero; rivalutare nuovi host, percorsi, eccezioni dei documenti o cambiamenti dell'HTML. Paginazione completa, altri portali, trial e accettazione restano da verificare.
- **Collegamenti:** LIV-001–LIV-005; [specifica](../../../../openspec/changes/define-toscana-alert-service/specs/official-source-ingestion/spec.md) e task 17; issue non disponibile.
