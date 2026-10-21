---
tipo: "comune"
codice_regione: "09"
codice_istat: "049009"
ultima_revisione: "2026-10-21"
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

Per timeout OCR ricorrenti, confrontare anche hash del PDF, input delle chiamate
e versioni HTML (LIV-010). Nel campione confrontato cambiano soltanto i token
CSRF: una nuova versione conservata non dimostra una nuova pubblicazione o un
nuovo allegato. Verificare le liste effettive di riuso e preflight del worker;
la policy indipendente `livorno-html-v1` ignora soltanto le due forme CSRF
esatte verificate in LIV-011. Il riuso è stato adottato nel perimetro development
revisionato; conservare gli originali e distinguere copie equivalenti, cache
riuscite e tentativi falliti. Le variazioni di testo, date, link, metadati,
allegati o completezza mantengono una nuova identità interpretativa.

Per `source_http_error`, distinguere il controllo complessivo dai singoli target
e dai log del crawler (LIV-009). Dal 2026-10-21 un target 404/410 uscito dagli
elenchi completi, senza riferimenti né misure ricavate, è ritirato: viene
ritentato ma non rende incompleto il giro (LIV-012). Restano bloccanti i target
ancora elencati, citati o con misure. Il messaggio strutturale «anti-bot» di Crawl4AI non dimostra da solo
un blocco del sito: verificare stato HTTP e contenuto restituito. Gli errori di
elaborazione del modello hanno ricevute e retry separati dalla raccolta.

Il precollaudo del 2026-10-04 produce un report individuale **pending**,
con valutazione, osservazione, accettazione e pubblicazione distinte. La
[campagna municipale](../../../operations/municipal-acceptance.md) mantiene
sette intervalli completi di 24 ore e le prove di originali, allegati ed errori.
LIV-008 registra i prerequisiti; il report non certifica copertura completa.

Distinguere gli allegati della notizia dai documenti generici del sito. Verificare ogni dominio esterno rispetto al riferimento ufficiale e ai vincoli di raccolta/riuso. Una risorsa `forbidden` può essere esclusa dal collector prima di qualsiasi richiesta HTTP: non prova un 403 o un file rimosso.

Il [collector](../../../../internal/backend/acquisition/preview.go) supporta ora una policy `attachments` per fonte/revisione: `content_class` limita la discovery dei PDF alla parte pertinente del documento, `external` ammette origini HTTPS e prefissi di percorso verificati e `validate_pdf` richiede un PDF valido. Le fonti senza questa policy mantengono la scansione generale e il rifiuto degli allegati esterni. Un allegato necessario non recuperato mantiene il controllo incompleto e il riferimento conservato.

Il ricontrollo del 2026-10-02 ha rivalutato LIV-001 e LIV-002 mediante pagina ufficiale, codice e un nuovo giro programmato. In quella fase nessuna correzione era stata introdotta; LIV-006 registra la successiva implementazione e verifica. Nell'acquisizione programmata l'errore ferma il giro prima dei documenti successivi: distinguere target scoperti e originali conservati; esiti per ambiente nel registro privato. Vedere LIV-003.

La selezione interpretativa del testo richiede un perimetro più ampio del
contenitore usato per scoprire gli allegati. Il replay del 2026-10-02 usa
`main#main-content`, conservando titolo, corpo e metadati controllati, oltre agli
allegati già raccolti. LIV-007 documenta il campione di Livorno e il trial
limitato; non estende la stessa policy ad altri comuni o a tutti i PDF del tenant.
Vedere [elaborazione locale](../../../operations/local-processing.md).

## Regole ed eccezioni

| ID | Ambito | Regola corrente | Stato dell'intervento |
| --- | --- | --- | --- |
| LIV-001 | PDF della pagina esaminata | Allegati esterni ammessi solo nel perimetro HTTPS/percorso revisionato; altri riferimenti restano `forbidden` | Correzione implementata e verificata sul perimetro; LIV-006 |
| LIV-002 | Discovery degli allegati | Selezione entro `article-content`, con esclusione di footer, navigazione e intestazioni | Correzione implementata e verificata sul perimetro; LIV-006 |
| LIV-007 | Testo delle notizie conservate di Livorno | `main#main-content` conserva titolo/corpo e rimuove navigazione/footer; PDF locali revisionati individualmente | Replay e trial limitato verificati; policy continuativa separata |
| LIV-011 | Equivalenza di versioni complete | `livorno-html-v1` normalizza soltanto i due campi CSRF esatti; PDF, testo, metadati e completezza restano significativi | Test e adozione development verificati; osservazione prolungata separata |

L'eventuale ammissione di risorse Municipium deve essere esplicita e limitata alla fonte e ai percorsi verificati. La nota di piattaforma non costituisce una autorizzazione generale.

La rivalutazione LIV-004 ha identificato come perimetro tecnico il solo host HTTPS `livorno-api.cloud.municipiumapp.it`, percorso `/s3/3612/allegati/`, per gli allegati pertinenti collegati dalle notizie comunali. Il riferimento della pagina esaminata compare dentro `article-content`; la guida generica è nel footer. `cloud-ita.municipiumapp.it` non richiede un'abilitazione per raccogliere quell'ordinanza. Sono da verificare separatamente l'effettivo accesso al file e le condizioni di ciascun documento; il collector dispone ora di una policy esplicita per autorizzare questo perimetro, con referral e riuso separati.

## Problemi aperti

LIV-001 e LIV-002 hanno una correzione nel checkout con test di confine, confronto del filtro con la pagina reale e verifica della conservazione; vedere LIV-006. Restano da verificare continuità dei download, eventuali nuovi host o eccezioni dei documenti e copertura oltre la prima pagina configurata. Trial, interpretazione e accettazione restano separati.

LIV-005 ha corretto il limite di accesso registrato in LIV-004: la prima prova HTTP non dimostrava indisponibilità per tutti i client. Il file esatto è stato acquisito con un normale browser e con `DirectHTTP`, e validato come PDF leggibile; dettagli dei trasporti ed esiti nell'archivio privato. Le [note legali comunali](https://www.comune.livorno.it/it/legal_notices) dichiarano CC BY 4.0 salvo eccezioni e contenuti di terzi; la clausola sui collegamenti esterni non concede da sola i diritti su ogni risorsa di un host.

Per verificare il recupero usare il trasporto dell'applicazione nell'ambiente interessato, conservare il risultato e validare il PDF, oltre allo stato HTTP. Un esito del browser o di un altro client non sostituisce la verifica di `DirectHTTP`. Prima di considerare completo un giro verificare anche persistenza, riferimenti e tutti gli allegati necessari. Gestire i fallimenti temporanei con retry limitati e cadenze della fonte; un rifiuto persistente richiede diagnosi o un canale ufficiale alternativo verificato.

## Registro delle scoperte

### LIV-013 — Notizie con misure aperte di nuovo ricontrollate

- **Data e ambito:** 2026-10-21; `livorno-municipal` dopo l'adozione della change `maintain-target-measure-state`.
- **Comportamento verificato:** otto notizie con misure risultano `unresolved`, cinque delle quali riferite da misure successive collegate. Le notizie più vecchie della finestra recente sono state riscaricate nel primo controllo dopo l'adozione, che resta completo con il target ritirato di LIV-012.
- **Limiti:** gli stati derivano da misure e collegamenti già validati; una notizia non ancora interpretata non ha stato misura. Le misure dalla fine ignota escono dai ricontrolli dopo 90 giorni.
- **Evidenza:** comando `source-tracked-targets` e registro privato `target-state-20261021`.

### LIV-012 — Notizia sostituita e giro di nuovo completo

- **Data e ambito:** 2026-10-21; `livorno-municipal` in development, target `allerta-gialla-per-forti-temporali-e-rischio-idrogeologico-e-idraulico-del-reticolo-minore`.
- **Osservazione confermata:** dal 7 ottobre il target restituisce 410 e manca dagli elenchi; la notizia sostitutiva sull'allerta arancio del 7–8 ottobre è comparsa nello stesso giro. Le 25 versioni conservate sono classificate come rilancio dell'allerta regionale, senza misure comunali estratte. Il target era l'unico non disponibile e teneva incompleti tutti i giri: 342 fallimenti consecutivi e un ritardo segnalato per due settimane.
- **Intervento verificato:** con la change `handle-withdrawn-unavailable-targets` il target è classificato ritirato. Il primo giro successivo all'adozione è completo, con un target ritirato conteggiato; i fallimenti consecutivi sono azzerati e il ritardo segnalato risulta chiuso. Originali, storico dei 410 e giri falliti sono conservati; nessuna equivalenza o revoca è dedotta.
- **Prossima verifica:** osservare i giri successivi e i nuovi target indisponibili; accettazione e pubblicazione restano separate.
- **Evidenza:** registro privato `withdrawn-targets-20261021`; [procedura target indisponibili](../../../operations/acquisition-recovery.md).

### LIV-011 — Equivalenza CSRF e riuso adottati nel perimetro revisionato

- **Data e ultima verifica:** 2026-10-09, fixture sintetiche, PostgreSQL isolato con race detector, preparazione locale di originali conservati e worker development; nessun nuovo fetch ufficiale manuale.
- **Ambito:** soltanto `livorno-municipal`, con policy indipendente `livorno-html-v1`. Campione di versioni complete e pagine OCR; configurazione, backup verificato, immagini e osservazioni dettagliate nel registro privato `livorno-reuse-20261009`.
- **Conoscenza confermata:** le forme esatte `<meta name="csrf-token" content="[A-Za-z0-9]{40}">` e `<input type="hidden" name="_token" value="[A-Za-z0-9]{40}">` possono essere normalizzate nel manifest senza modificare gli originali. I test conservano cambi di testo, data, orario, link e completezza; forme non revisionate e altre fonti non ereditano la regola. I gruppi sono contigui e rispettano i confini di archiviazione.
- **Intervento:** policy implementata e riuso OCR/interpretativo con preflight attivato sul singolo worker development. La preparazione locale ha riconosciuto 18 copie fra 36 versioni di Livorno senza chiamate al provider; un job copia già in coda resta differito con zero tentativi. Timeout locale esteso a 600 secondi come trial di capacità separato; contesto, GPU e modello non modificati. Nel primo campione ordinario tre pagine OCR sono complete e persistite; la più lunga termina in 305 secondi, oltre il vecchio limite di 180, senza nuovi timeout nel campione.
- **Evidenza:** [policy e fixture sintetiche](../../../../internal/backend/classification/manifest_test.go), [gruppi e riuso senza chiamate](../../../../internal/backend/interpretation/preflight_integration_test.go), [riuso delle pagine](../../../../internal/backend/ocr/reuse.go) e [trial operativo](../../../operations/local-llm-fallback.md). Originali, ricevute e misure operative restano privati.
- **Prossima verifica:** osservare cache ordinaria, documenti lunghi, cambi significativi e qualità per almeno la finestra prevista dal canary. La verifica delimitata non chiude i fallimenti storici, LIV-009, accettazione o copertura della fonte; stato della GPU e capacità continuativa restano da verificare.
- **Collegamenti:** aggiorna la proposta LIV-010; task 11 della [specifica di efficienza](../../../../openspec/changes/reduce-regolo-token-waste/tasks.md).

### LIV-010 — Versioni CSRF, OCR ripetuto e pausa condivisa del provider

- **Data e ultima verifica:** 2026-10-09, ricevute recenti, configurazione effettiva e confronto di due originali già conservati; nessun nuovo fetch ufficiale o chiamata manuale al modello.
- **Ambito:** elaborazione ordinaria `livorno-municipal`; coppia di versioni della notizia collegata all'ordinanza n. 303 e hash degli allegati nelle ricevute. Dettagli e originali nel registro privato `JOB-DIAGNOSIS-20261009`.
- **Conoscenza confermata:** la coppia HTML ha metadati e PDF identici e differisce soltanto nelle due forme CSRF esatte già documentate per Cascina. Più versioni accodano OCR dello stesso PDF/input; nel worker ispezionato Livorno è esclusa dalle liste di riuso OCR, interpretativo e preflight. I timeout al limite HTTP esauriscono i tentativi del singolo job, mentre nuove versioni possono creare nuovi job. Il gate locale dello stesso modello è condiviso fra OCR e fasi testuali; la sua pausa può rinviare entrambe.
- **Intervento:** diagnosi verificata; normalizzazione, adozione del riuso per Livorno e regolazione della capacità restano proposte. Nessuna modifica dei provider, rilancio, archiviazione o attivazione della fonte.
- **Evidenza:** [identità dei documenti](../../../../internal/backend/documents/model.go), [policy HTML](../../../../internal/backend/classification/manifest.go), [identità dei job OCR](../../../../internal/backend/ocr/runner.go), [fallback e gate condiviso](../../../../cmd/iwa/local_fallback.go) e [operazioni locali](../../../operations/local-llm-fallback.md). Originali, hash e ricevute dettagliate restano privati.
- **Prossima verifica:** verificare su fixture sintetiche una policy Livorno che ignori soltanto le forme CSRF osservate e preservi cambi reali; valutare riuso OCR/preflight con un trial delimitato. Verificare separatamente tempo dei documenti più lunghi, timeout e capacità del modello. La coppia esaminata non certifica equivalenza di ogni pagina o completezza della raccolta.

### LIV-009 — Diagnosi distinta di target indisponibili e rifiuto del crawler

- **Data e ultima verifica:** 2026-10-08, codice e ricevute/log conservati; nessun nuovo fetch ufficiale o chiamata al modello.
- **Ambito:** `livorno-municipal`, categoria Protezione Civile e target già tracciati. Ambiente, revisione, URL ed esiti dettagliati restano nel registro privato `PROCESSING-FAILURES-20261008`.
- **Conoscenza confermata:** i target 404/410 e quelli in backoff rendono incompleto il giro senza impedire tutti i download successivi. Un rifiuto strutturale del crawler può essere riportato come `source_http_error` anche se la pagina è raggiungibile; la dicitura «anti-bot» non prova la causa remota.
- **Intervento:** sola diagnosi verificata; nessuna esclusione di target, modifica dei retry, riattivazione provider o correzione applicata.
- **Evidenza:** [collector programmato](../../../../internal/backend/acquisition/preview.go), [adattatore Crawl4AI](../../../../internal/backend/acquisition/crawl4ai.go) e [procedura di recupero](../../../operations/acquisition-recovery.md). Ricevute e log originali conservati privatamente.
- **Prossima verifica:** confrontare stato HTTP/contenuto del target rifiutato con il trasporto effettivo; prima di escludere un URL indisponibile, completare la traversata e la revisione previste dalla procedura. Non dedurre rimozione definitiva, revoca di una misura o completezza dalla sola raggiungibilità; accettazione e pubblicazione restano separate.

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


### LIV-007 — Contenuto interpretativo e trial locale limitato

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, originali conservati, parser applicativo, modello locale e risultati persistiti in development.
- **Ambito:** 16 versioni di notizie Livorno nel replay; due confronti municipalità baseline/candidato nel trial. I PDF rimangono risorse separate.
- **Conoscenza:** confermato sul campione. **Intervento:** verifica applicata nel reader del trial; nessuna policy continuativa della fonte attivata.
- **Osservazione:** `main#main-content` conserva titolo, testo completo di `article-content`, metadati controllati e testo degli allegati disponibili, riducendo del 45,3% i byte HTML normalizzati del campione. I due confronti con lo stesso modello concordano sulla pertinenza. Una classe scelta soltanto per gli allegati non basta a dimostrare copertura del titolo e dei metadati.
- **Evidenza:** [selettore e fallback](../../../../internal/backend/classification/html_body.go), [operazioni e risultati del trial](../../../operations/local-processing.md); originali, selezioni, token e ricevute restano nel registro privato.
- **Conseguenza:** la riduzione del testo non autorizza nuovi download o una directory PDF di solo testo. PDF con immagini mantengono il fallback; significato dei grafici vettoriali e qualità dell’estrazione delle misure richiedono verifica separata.
- **Prossima verifica:** ampliare casi pertinenti e non pertinenti, variazioni del layout e interpretazione downstream prima di una policy continuativa.


### LIV-008 — Report individuale pendente e collaudo indipendente

- **Data:** 2026-10-04. **Ultima verifica:** 2026-10-04, registro dell'ambiente, apertura della campagna e test sintetici con PostgreSQL isolato; nessuna nuova consultazione web della fonte.
- **Ambito:** fonte municipale della scheda. Protezione Civile e allegati revisionati; copertura oltre la prima pagina e altri portali ancora da verificare.
- **Conoscenza:** confermato per i prerequisiti e il comportamento del servizio. **Intervento:** percorso indipendente implementato e report pendente consegnato; accettazione non completata.
- **Osservazione:** lo scope esplicito `municipality` valuta solo questa fonte senza richiedere i tre prodotti regionali. Non riduce durata, cadenza, confronti con originali/allegati o prove di errore. I prerequisiti irrisolti impediscono l'apertura della campagna; la raccolta eventualmente già attiva non li annulla.
- **Evidenza:** [report e procedura](../../../operations/municipal-acceptance.md), [campagne](../../../../internal/backend/observation/store.go), [prove isolate](../../../../internal/backend/observation/observation_integration_test.go). Report completo, ricevuta del tentativo e stato dei controlli nell'archivio privato `ROLLOUT-GATES-20261004`.
- **Conseguenza:** il report mantiene lo stato pending e i controlli esistenti; non estende il perimetro né dichiara copertura dei cinque comuni.
- **Prossima verifica:** risolvere i prerequisiti dell'ambito, completare valutazione e osservazione, quindi accettare e abilitare separatamente la fonte verificata.
- **Collegamenti:** LIV-006 e LIV-007; task 11.5 della [specifica Toscana](../../../../openspec/changes/define-toscana-alert-service/tasks.md).
