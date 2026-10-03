---
tipo: "comune"
codice_regione: "09"
codice_istat: "050004"
ultima_revisione: "2026-10-03"
---

# Calcinaia

## Identità e ambito

Comune di Calcinaia, provincia di Pisa, ISTAT `050004`. Fonte IWA: `calcinaia-municipal`. Collegamenti: [Toscana](../README.md), [stato pubblico documentato](../../../coverage.md#region-09). La scheda descrive il perimetro di ricerca, senza certificare completezza o attivazione.

## Mappa delle fonti

| Fonte | Ruolo | Riferimento e limiti |
| --- | --- | --- |
| Sito comunale | Primaria | [Notizie](https://www.comune.calcinaia.pi.it/tipi-di-notizia/notizie), [Avvisi](https://www.comune.calcinaia.pi.it/tipi-di-notizia/avvisi), [Comunicati](https://www.comune.calcinaia.pi.it/tipi-di-notizia/comunicati); sezioni da verificare nella configurazione |
| Cittadino Informato | Canale riconosciuto dal Comune; diagnostica secondaria nella specifica IWA | [Pagina Calcinaia](https://cittadinoinformato.it/calcinaia/), collegata dalla homepage comunale; riconoscimento istituzionale verificato il 2026-10-03, completezza e aggiornamento da collaudare |
| Albo pretorio | Atti di supporto | Cercare gli atti citati nelle notizie primarie; il riferimento al singolo atto va verificato |

La mappa consolida la [specifica di acquisizione](../../../../openspec/changes/define-toscana-alert-service/specs/official-source-ingestion/spec.md). La sezione Notizie è stata consultata il 2026-10-02; ciò non verifica tutte le altre sezioni o i canali secondari.

Cittadino Informato è previsto come sentinella di confronto nella specifica,
ma questo non certifica un controllo periodico attivo. CLN-008 distingue il ruolo
documentato dall'acquisizione configurata: verificare una fonte/sezione dedicata
e ricevute recenti prima di dichiarare il confronto automatico operativo. La
limitazione registrata rinvia l'acquisizione attiva alla revisione delle condizioni
di riuso; questa scheda non ne rivaluta le condizioni correnti.

CLN-009 conferma il riconoscimento istituzionale del canale di Calcinaia: il
Comune lo collega e lo raccomanda, mentre Regione e ANCI documentano il servizio.
Il ruolo secondario assegnato dalla specifica IWA descrive la gerarchia di
acquisizione; non è un giudizio di mancata ufficialità. Provenienza riconosciuta,
affidabilità tecnica misurata e condizioni di acquisizione/riuso sono verifiche
distinte. Attribuire ogni comunicazione al suo emittente e conservare il prodotto
CFR originario per livelli regionali; non trasferire l'autorevolezza a tutti i
contenuti del dominio o dedurre assenza di avvisi da un elenco vuoto.

## Guida operativa

Confrontare le sezioni configurate e la paginazione: un elenco generale non dimostra l'esaustività degli elenchi tematici. Distinguere data di pubblicazione, aggiornamento della pagina e validità di una misura. Acquisire e interpretare gli atti necessari prima di determinare durata o cessazione di chiusure.

Per documenti 404/410 usare la [procedura di recupero](../../../operations/acquisition-recovery.md): assenza dagli elenchi completamente traversati, riferimenti e misure ancora aperte vanno verificati prima di una esclusione. Conservare lo storico; un'assenza non prova revoca né equivalenza a un altro URL.

Nel ricontrollo del 2026-10-02 è stato seguito il percorso programmato del collector, confrontando esito del controllo, target verificati e presenza degli oggetti originali conservati. La verifica è limitata alle sezioni e alla finestra della configurazione effettiva; risultati e configurazione per ambiente sono nell'archivio privato. Vedere CLN-003.

La situazione API combina fonti comunali e prodotti regionali ammessi alla
pubblicazione. Verificare la copertura di ciascuna fonte: `pending` con controlli
recenti non equivale ad assenza di allerte. Confrontare separatamente originali,
fatti regionali, provvedimenti locali e abilitazione pubblica; una classificazione
o estrazione riuscita non dimostra che un provvedimento sia consolidato nel
dominio. CLN-004 e [TOS-005](../README.md#tos-005--situazione-pubblica-vuota-e-livelli-non-determinati-hanno-cause-distinte)
descrivono questi criteri, senza cambiare l'accettazione della fonte.

La [pubblicazione manuale in development](../../../operations/development-publication.md) consente di scegliere singoli
comuni per la consultazione API/MCP, includendo i prodotti regionali applicabili.
La scelta è revocabile e distinta dalla raccolta e dall’accettazione delle fonti.
Gli stati pendenti e i livelli non verificati restano espliciti; staging e produzione
mantengono i requisiti della specifica. Vedere CLN-006.

## Regole ed eccezioni

| ID | Ambito | Regola corrente | Stato dell'intervento |
| --- | --- | --- | --- |
| CLN-001 | Gerarchia delle fonti | Sito comunale primario; Cittadino Informato come confronto; albo per atti citati | Regola di progetto documentata; verifiche di copertura separate |
| CLN-002 | Target 404/410 | Esclusione solo con decisione motivata per fonte/configurazione/URL; rediscovery ripristina il controllo | Procedura e codice presenti; applicazione al singolo target da verificare |

## Problemi aperti

CLN-007 conferma un collegamento da completare fra estrazioni e dominio comunale:
il salvataggio in `extracted_measures` non inserisce misure in
`domain_local_measures` né fasi in `domain_operational_phases`. La query di
situazione legge queste ultime tabelle. Verificare separatamente i candidati
estratti, la loro proiezione con evidenza/validità e gli stati di interpretazione
richiesti dalla query; la pubblicazione manuale non esegue questo passaggio.
`documents_requiring_attention` raccoglie soltanto avvisi collegati a misure già
presenti: se queste mancano, l'array vuoto non esclude documenti non interpretati.
La ricerca API dei documenti conservati è un percorso distinto.

Trial e accettazione restano quelli del coverage tracker. Le lacune dei canali secondari vanno registrate separatamente dalle omissioni nel perimetro primario. Consultare il registro privato per errori e recuperi dell'ambiente; non considerarli una certificazione pubblica.

Prima dell'accettazione controllare l'ultima regressione della fonte: un esito
fallito richiede una rivalutazione corretta dello stesso contratto, con gli esiti
attesi revisionati e i casi non eseguiti completati. Un recupero dell'acquisizione
o un test sintetico positivo non chiude da solo quel report. Vedere CLN-005.

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

### CLN-003 — Verificare gli originali oltre all'esito del controllo

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, codice e ricontrollo operativo privato.
- **Ambito:** `calcinaia-municipal`, raccolta programmata delle sezioni configurate; dettagli dell'ambiente e della revisione nel registro privato.
- **Conoscenza:** confermato come criterio diagnostico. **Intervento:** verifica eseguita, nessuna correzione applicata.
- **Osservazione:** conteggio degli elenchi, target scoperti e documenti acquisiti descrivono fasi diverse; la presenza degli originali va controllata nelle versioni e negli oggetti conservati.
- **Evidenza:** [collector](../../../../internal/backend/acquisition/preview.go), [store documenti](../../../../internal/backend/documents/store.go) e [procedura di recupero](../../../operations/acquisition-recovery.md); esiti dettagliati privati.
- **Conseguenza:** un ricontrollo riguarda il perimetro effettivo e non completa l'accettazione o la copertura dei canali secondari di CLN-001.
- **Prossima verifica:** proseguire l'osservazione delle sezioni e confrontare pubblicazioni note e allegati necessari nel trial.

### CLN-004 — Distinguere dati interni e situazione API pubblicabile

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, codice e diagnosi operativa in sola lettura.
- **Ambito:** situazione del comune `050004`, fonte `calcinaia-municipal` e prodotti regionali applicabili; dettagli dell'ambiente nel registro privato.
- **Conoscenza:** confermato come criterio diagnostico. **Intervento:** verifica eseguita; nessuna correzione o attivazione applicata.
- **Osservazione:** la query restituisce separatamente mapping, provvedimenti, fasi, prodotti regionali e copertura. I filtri pubblici possono escludere dati esistenti nelle viste private; la copertura segnala l'accettazione pendente indipendentemente dall'esito dei controlli di acquisizione. Gli array vuoti non sono uno stato sintetico di assenza di allerte.
- **Evidenza:** [query di situazione](../../../../internal/backend/publicquery/query.go), [visibilità pubblica](../../../../internal/backend/publicquery/administrative.go), [stati di copertura](../../../../internal/backend/publicquery/store.go); riferimento operativo privato `SITUATION-20261002-01`.
- **Conseguenza:** distinguere un dato non pubblicabile da un fatto non consolidato o da una dichiarazione ufficiale di assenza di criticità; una fonte comunale pendente non determina da sola la disponibilità dei prodotti regionali.
- **Prossima verifica:** completare i controlli e l'accettazione separati delle fonti e verificare la rappresentazione dell'indisponibilità nell'API prima di dichiarare uno stato comunale corrente.
- **Collegamenti:** TOS-005, [specifica di accesso pubblico](../../../../openspec/changes/define-toscana-alert-service/specs/public-alert-access/spec.md), [task 7.3 e 10.1](../../../../openspec/changes/define-toscana-alert-service/tasks.md).

### CLN-005 — Il recupero non sostituisce la rivalutazione della regressione

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, inventario del registro operativo, report delle campagne e test su PostgreSQL isolato.
- **Ambito:** precollaudo della fonte `calcinaia-municipal`; casi ed esiti specifici restano nel dossier privato.
- **Conoscenza:** confermato come vincolo diagnostico. **Intervento:** precollaudo e valutazioni registrati; rivalutazione reale e accettazione ancora da completare.
- **Osservazione:** una regressione fallita impedisce l'accettazione finché una nuova esecuzione dello stesso contratto documenta la correzione e passa. Il confronto deve includere discovery, omissioni negli aggiornamenti e affermazioni temporali non supportate, distinguendo casi reali e simulazioni.
- **Evidenza:** [contratto di valutazione](../../../../internal/backend/evaluation/report.go), [vincolo sulle regressioni](../../../../internal/backend/registry/regression.go), [test persistente del blocco e della rivalutazione](../../../../internal/backend/registry/regression_integration_test.go). Riferimento privato `ACCEPTANCE-PRECHECK-20261002-01`.
- **Conseguenza:** non marcare una regressione superata sulla base del solo deployment o della classificazione riuscita; non inserire un'accettazione con attestazioni prive di evidenza.
- **Prossima verifica:** recuperare o ricostruire il corpus con esiti attesi revisionati, completare la rivalutazione e i confronti del trial, poi verificare l'accettazione e l'abilitazione separate.
- **Collegamenti:** CLN-003, CLN-004, TOS-006 e [task 9.3, 10.1 e 24](../../../../openspec/changes/define-toscana-alert-service/tasks.md).

### CLN-006 — Scelta comunale di pubblicazione per development

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, fixture, PostgreSQL isolato, controllo nativo nel browser e confronto API/MCP in development; dettagli operativi conservati separatamente.
- **Ambito:** Pubblicazione manuale di Calcinaia in development.
- **Conoscenza:** comportamento implementato; non certifica accettazione reale. **Intervento:** controlli separati per ambiente e comune.
- **Osservazione:** un operatore può pubblicare dati comunali conservati e fatti regionali applicabili senza completare le verifiche di produzione. La scelta non si estende agli altri comuni della stessa zona, non cambia l’accettazione né autorizza copie pubbliche. Revoca e cambio di ambito fanno scadere le viste API/MCP conservate.
- **Evidenza:** [controlli operativi](../../../operations/development-publication.md) e [specifica pubblica](../../../../openspec/changes/define-toscana-alert-service/specs/public-alert-access/spec.md#requirement-manual-municipality-publication-in-development).
- **Conseguenza:** in development dati consultabili possono accompagnare `public_state=pending` e `development_publication=true`; nessuna misura locale o colore regionale viene dedotto per colmare dati mancanti.
- **Prossima verifica:** usare i controlli separati per la consultazione di sviluppo; completare le revisioni e le rivalutazioni registrate prima della pubblicazione verificata in produzione.

### CLN-007 — Estrazioni conservate senza proiezione comunale nella situazione

- **Data:** 2026-10-03. **Ultima verifica:** 2026-10-03, codice, database e API development in sola lettura; nessuna nuova acquisizione o inferenza avviata.
- **Ambito:** estrazioni di `calcinaia-municipal`, dominio comunale e situazione API del comune `050004`.
- **Conoscenza:** confermato sul percorso esaminato. **Intervento:** diagnosi registrata; collegamento e rappresentazione dei documenti pendenti da implementare e validare.
- **Osservazione:** `extraction.Store.Put` conserva risultati, misure candidate ed evidenze nelle tabelle di estrazione senza alimentare le tabelle del dominio lette da `measures` e `phases`. I metodi di inserimento del dominio esistono, ma non sono collegati al salvataggio delle estrazioni. La situazione costruisce i documenti da attenzionare soltanto dagli avvisi delle misure già restituite. La copertura comunale in development conserva inoltre lo stato iniziale `not_processed` con motivazione di accettazione pendente, senza riassumere le estrazioni interne.
- **Evidenza:** [salvataggio estrazioni](../../../../internal/backend/extraction/store.go), [store del dominio](../../../../internal/backend/domain/store.go), [query dei fatti](../../../../internal/backend/publicquery/facts.go), [situazione e ricerca documenti](../../../../internal/backend/publicquery/query.go), [copertura](../../../../internal/backend/publicquery/store.go). Conteggi, versioni e risposte development restano nel registro privato `SITUATION-20261003-01`.
- **Conseguenza:** gli array vuoti non dipendono necessariamente dall'accettazione o dall'assenza di comunicazioni; estrazione riuscita e pubblicazione development non dimostrano una proiezione comunale completa. Un'attivazione estratta non determina automaticamente una fase operativa.
- **Prossima verifica:** collegare le estrazioni validate a una proiezione comunale versionata e idempotente, conservando evidenze, campi indeterminati e storico; verificare il percorso worker–dominio–API/MCP e un replay esplicitamente delimitato. Verificare anche documenti pendenti senza misure precedenti e qualità distinta dall'accettazione.
- **Collegamenti:** CLN-004, CLN-006 e [TOS-008](../README.md#tos-008--la-consultazione-in-development-non-completa-la-lettura-delle-mappe).

### CLN-008 — Sentinella secondaria prevista e raccolta attiva sono distinte

- **Data:** 2026-10-03. **Ultima verifica:** 2026-10-03, specifiche, codice e configurazione development in sola lettura; nessuna consultazione del sito secondario.
- **Ambito:** ruolo di Cittadino Informato per Calcinaia e perimetro del collector municipale.
- **Conoscenza:** confermato sul perimetro verificato. **Intervento:** diagnosi registrata; nessuna attivazione.
- **Osservazione:** proposta, design e specifica prevedono un confronto diagnostico secondario; il task storico 1.11 non dimostra un controllo continuativo corrente. La configurazione esaminata dichiara esclusione dell'acquisizione attiva in attesa delle condizioni di riuso e limita le sezioni ai tre elenchi del sito comunale. Il controllo del registro e degli originali conservati non ha individuato un canale acquisito del dominio secondario nell'ambiente esaminato.
- **Evidenza:** [specifica di acquisizione](../../../../openspec/changes/define-toscana-alert-service/specs/official-source-ingestion/spec.md), [design](../../../../openspec/changes/define-toscana-alert-service/design.md), [task storico](../../../../openspec/changes/define-toscana-alert-service/tasks.md). Riscontro operativo: registro privato `SITUATION-20261003-01`.
- **Conseguenza:** la presenza del nome nelle limitazioni non significa che il worker confronti periodicamente i due siti; il canale secondario non stabilisce misure, date o completezza senza riscontro primario.
- **Prossima verifica:** revisionare separatamente il perimetro e le condizioni del canale secondario, poi implementare/configurare e verificare il confronto diagnostico prima di dichiararlo operativo.
- **Collegamenti:** CLN-001 e CLN-007.

### CLN-009 — Riconoscimento istituzionale del canale di Calcinaia

- **Data:** 2026-10-03. **Ultima verifica:** 2026-10-03, consultazione web di fonti istituzionali e della presentazione della piattaforma; nessun collaudo continuativo o attivazione del collector.
- **Ambito:** referral comunale a `https://cittadinoinformato.it/calcinaia/` e collaborazione Regione Toscana–ANCI sul servizio.
- **Conoscenza:** confermato per il riconoscimento del canale; affidabilità operativa da misurare. **Intervento:** guida aggiornata; policy IWA invariata.
- **Osservazione:** la homepage comunale collega direttamente la pagina Calcinaia nella sezione dei siti tematici per allerte e piano di protezione civile; una comunicazione meteo del Comune invita a usare l'app. Regione e ANCI documentano il protocollo di collaborazione sul servizio. Queste evidenze sostengono il riconoscimento come canale istituzionale aggiuntivo, senza dimostrare esaustività, tempestività o la provenienza di ogni singolo messaggio.
- **Evidenza:** [homepage del Comune](https://comune.calcinaia.pi.it/), [comunicazione comunale](https://comune.calcinaia.pi.it/novita/allerta-meteo-giovedi-17-settembre), [notizia della Regione](https://www.toscana-notizie.it/-/l-app-cittadino-informato-si-rinnova-protocollo-d-intesa-tra-regione-e-anci), [conferma ANCI](https://ancitoscana.it/protezione-civile-anci-toscana-e-regione-toscana-siglano-un-protocollo-dintesa-per-lo-sviluppo-e-la-diffusione-dellapp-cittadino-informato/), [presentazione della piattaforma](https://cittadinoinformato.it/il-progetto/).
- **Conseguenza:** l'esclusione dalla raccolta attiva di CLN-008 non è una dichiarazione di fonte non autentica. Il riconoscimento istituzionale non completa l'accettazione tecnica né determina le condizioni di riuso.
- **Prossima verifica:** identificare l'emittente dei messaggi nel perimetro comunale e confrontare contenuti, date, aggiornamenti e omissioni con gli originali; verificare separatamente il contratto di acquisizione prima dell'attivazione.
- **Collegamenti:** CLN-001, CLN-008 e [specifica di acquisizione](../../../../openspec/changes/define-toscana-alert-service/specs/official-source-ingestion/spec.md).
