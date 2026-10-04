---
tipo: "comune"
codice_regione: "09"
codice_istat: "050004"
ultima_revisione: "2026-10-04"
---

# Calcinaia

## Identità e ambito

Comune di Calcinaia, provincia di Pisa, ISTAT `050004`. Fonte IWA: `calcinaia-municipal`. Collegamenti: [Toscana](../README.md), [stato pubblico documentato](../../../coverage.md#region-09). La scheda descrive il perimetro di ricerca, senza certificare completezza o attivazione.

## Mappa delle fonti

| Fonte | Ruolo | Riferimento e limiti |
| --- | --- | --- |
| Sito comunale | Primaria | [Notizie](https://www.comune.calcinaia.pi.it/tipi-di-notizia/notizie), [Avvisi](https://www.comune.calcinaia.pi.it/tipi-di-notizia/avvisi), [Comunicati](https://www.comune.calcinaia.pi.it/tipi-di-notizia/comunicati); sezioni da verificare nella configurazione |
| Cittadino Informato | Canale riconosciuto dal Comune; acquisizione/discovery aggiuntiva con verifica primaria nel piano IWA | [Pagina Calcinaia](https://cittadinoinformato.it/calcinaia/), collegata dalla homepage comunale; riconoscimento verificato il 2026-10-03; acquisizione e verificatore implementati, percorso development verificato da CLN-015, adozione programmata separata |
| Albo pretorio | Atti di supporto | Cercare gli atti citati nelle notizie primarie; il riferimento al singolo atto va verificato |

La mappa consolida la [specifica di acquisizione](../../../../openspec/changes/define-toscana-alert-service/specs/official-source-ingestion/spec.md). La sezione Notizie è stata consultata il 2026-10-02; ciò non verifica tutte le altre sezioni o i canali secondari.

Cittadino Informato, già previsto come sentinella di confronto, è incluso dalla
decisione CLN-010 come canale aggiuntivo di acquisizione/discovery con verifica
primaria. Il nuovo piano non certifica un controllo periodico attivo. CLN-008 distingue il ruolo
documentato dall'acquisizione configurata: verificare una fonte/sezione dedicata
e ricevute recenti prima di dichiarare il confronto automatico operativo. La
limitazione storica CLN-008 rinviava l’acquisizione alla revisione delle condizioni
di riuso. CLN-012 supera quel prerequisito documentale; CLN-013/CLN-014 registrano
il codice e i collaudi successivi; CLN-015 verifica il percorso delimitato in
development con dati dedicati.

CLN-009 conferma il riconoscimento istituzionale del canale di Calcinaia: il
Comune lo collega e lo raccomanda, mentre Regione e ANCI documentano il servizio.
Il ruolo secondario assegnato dalla specifica IWA descrive la gerarchia di
acquisizione; non è un giudizio di mancata ufficialità. Provenienza riconosciuta,
affidabilità tecnica misurata e condizioni di acquisizione/riuso sono verifiche
distinte. Attribuire ogni comunicazione al suo emittente e conservare il prodotto
CFR originario per livelli regionali; non trasferire l'autorevolezza a tutti i
contenuti del dominio o dedurre assenza di avvisi da un elenco vuoto.

## Guida operativa

CLN-022 verifica il recupero selettivo live con audit e conservazione dello storico.
La riapertura e l’eccezione sono proiettate, ma il nuovo output omette tre ambiti
di divieto del contratto revisionato. Consultare l’ultima regressione fallita;
il replay riuscito non attesta completezza del worker né chiude le campagne.

Il recupero provider del 2026-10-04 verifica nuovamente OCR remoto e una pipeline
ordinaria di Calcinaia, dopo ripresa controllata del blocco credenziale. Una
citazione non valida rimane rifiutata e conservata. CLN-019 distingue questo
recupero dall'accettazione della fonte e dal completamento della regressione.

La vista API/MCP di situazione presenta ora `processed_data` e
`source_summaries` (CLN-018): conclusioni e fatti separati dalle sintesi
Regione/CFR, Comune e Cittadino Informato. Le misure documentate con validità
`undetermined` non sono confermate in vigore; il verde regionale non risolve le
date comunali. Un canale `not_collected` indica assenza di dati pubblicabili nella
vista, non assenza di avvisi. Documenti, ricevute e qualità dettagliata restano
negli endpoint dedicati. Vedere [contratto](../../../backend/public-api-mcp.md#situation-response).


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

Consultare le [note della piattaforma](../../../fonti/piattaforme/cittadino-informato.md)
prima del nuovo flusso: il task 26.2 ora richiede un sistema funzionante di verifica
multipla Regione/CFR–Comune–piattaforma, senza un gate di ottenimento licenze, accordi
o documentazione della base giuridica. Mantenere controlli comunali
indipendenti per gli avvisi assenti dalla piattaforma; confrontare con il CFR soltanto
le affermazioni regionali comparabili, senza richiedere un'allerta regionale per ogni
provvedimento locale. I task 26.2–26.6 sono completati nei rispettivi perimetri delimitati,
con il coverage tracker invariato.

La [revisione preliminare del canale](../../../fonti/piattaforme/cittadino-informato-review.md)
identifica l'indice REST e i limiti delle date del contenitore WordPress. Il
trattamento dei visitatori dichiarato da ANCI Toscana non risolve i diritti sul
flusso IWA. Il precedente criterio documentale è superato da CLN-012; occorrono
ricevute persistenti con fonti/versioni, tempi, campi ed esiti distinti. CLN-013
verifica le route API, i limiti e la persistenza dell’acquisizione; CLN-014 aggiunge
ricevute e proiezione primaria con prove sintetiche. CLN-015 aggiunge il
percorso development delimitato; cadenza operativa e valutazione estesa restano
da collaudare. Il CFR può essere
non applicabile a una misura locale; non è richiesto un accordo unanime di tre canali.

| ID | Ambito | Regola corrente | Stato dell'intervento |
| --- | --- | --- | --- |
| CLN-001 | Gerarchia delle fonti | Sito comunale primario; Cittadino Informato come confronto; albo per atti citati | Regola di progetto documentata; verifiche di copertura separate |
| CLN-002 | Target 404/410 | Esclusione solo con decisione motivata per fonte/configurazione/URL; rediscovery ripristina il controllo | Procedura e codice presenti; applicazione al singolo target da verificare |

## Problemi aperti

CLN-017 supera la lacuna di integrazione del worker descritta da CLN-007:
l’estrazione primaria completa ora alimenta le misure del dominio con legami alle
evidenze, validità letterale e confine di conoscenza. Gli aggiornamenti collegati
sono applicati solo con entrambe le misure sostenute; un’attivazione del COC non
inventa una fase operativa. Valutazioni, canali piattaforma, risorse mancanti e
fonti sospese sono esclusi da questo ingresso. La situazione segnala anche i
primi avvisi non interpretati, senza richiedere una misura precedente, e distingue
l’elaborazione dall’accettazione pendente. Il recupero delimitato dei risultati
conservati non certifica l’intero corpus né corregge i falsi positivi del modello.

La criticità CFR corrente per la zona A4 è ora determinata dalle mappe PDF
conservate, con sette rischi per oggi/domani e pagine di evidenza; vedere
[TOS-014](../README.md#tos-014--criticità-della-zona-a4-determinata-dai-poligoni-pdf).
Questa verifica regionale non risolve automaticamente la validità delle singole
misure comunali: un atto diverso per data/oggetto non chiarisce per inferenza il
conflitto agosto/settembre di un’altra comunicazione.

CLN-020 supera il fallimento di merge della riapertura parziale di CLN-016 nel
replay delle risposte conservate e nelle fixture sintetiche con PostgreSQL:
grafia originale dei luoghi, riapertura, restrizioni e COC mantengono evidenze
separate. Catalogo v25 adottato in development; rivalutazione dello stesso contratto
verificata con 14/14 confronti. CLN-021 distingue il recupero selettivo delle
versioni archiviate dalla prova ordinaria ancora da completare.
La correttezza semantica dell'intero corpus resta da completare. Date in conflitto, termini condizionali senza
inizio stabilito e ambiti territoriali mancanti rimangono indeterminati; una
misura visibile non viene dichiarata automaticamente vigente oggi.

Trial e accettazione restano quelli del coverage tracker. Le lacune dei canali secondari vanno registrate separatamente dalle omissioni nel perimetro primario. Consultare il registro privato per errori e recuperi dell'ambiente; non considerarli una certificazione pubblica.

Prima dell'accettazione controllare l'ultima regressione della fonte: un esito
fallito richiede una rivalutazione corretta dello stesso contratto, con gli esiti
attesi revisionati e i casi non eseguiti completati. Un recupero dell'acquisizione
o un test sintetico positivo non chiude da solo quel report. Vedere CLN-005.

## Registro delle scoperte

### CLN-022 — Recupero live riuscito e omissione distinta dal replay

- **Data e ultima verifica:** 2026-10-04; worker development, PostgreSQL, API/MCP con client SDK e rivalutazione delle due campagne. Nessun nuovo fetch delle fonti richiesto dal recupero; nuove inferenze sui testi conservati.
- **Ambito e conoscenza:** due versioni selezionate di `calcinaia-municipal`, riapertura parziale ed eccezione delle strade. Classificazione, estrazione e linking completati; audit idempotente, marker e risultati precedenti invariati. La nuova estrazione conserva un solo ambito di divieto invece dei quattro revisionati: tre omissioni confermate nel perimetro del confronto.
- **Intervento verificato:** adozione della revisione pulita e migrazione additiva; storico e viste salvate conservati, riapertura con citazione e tempi indeterminati, eccezione distinta, API/MCP equivalenti anche in paginazione, isolamento da Cascina. Registrata una regressione fallita con lo stesso contratto e controlli riusati attribuiti, conservando il successo del replay e i fallimenti precedenti.
- **Evidenza:** [esito operativo](../../../operations/municipal-interpretation.md#esito-del-recupero-live), task 29.3 e registro privato `CLN-CONTINUATION-20261004`. Le campagne rivalutate restano `extended`; controlli delle fonti conservati e produzione ferma.
- **Prossima verifica:** correggere e rivalutare la completezza dei distinti ambiti nella nuova risposta; completare confronti degli originali/allegati, prove di errore e ritardi delle campagne. Nessuna accettazione o chiusura di 9.3/10.1.

### CLN-021 — Recupero selettivo con audit delle interpretazioni archiviate

- **Data e ultima verifica:** 2026-10-04; archivio e regressioni development, test sintetici e PostgreSQL isolato; nessun nuovo fetch delle fonti per questo intervento.
- **Ambito e conoscenza:** `calcinaia-municipal`, casi conservati di riapertura ed eccezione. Il worker ordinario li esclude perché le versioni sono archiviate, anche se richieste dal reprocessing. La rivalutazione di CLN-020 passa 14/14 con lo stesso contratto, conservando il fallimento precedente e dichiarando il riuso delle risposte dei provider.
- **Intervento:** l’operatore sceglie un recupero selettivo con audit. I task 29.1/29.2 implementano richiesta privata per versioni esatte, revisione attiva, evidenza conservata e ultime regressioni riuscite. Archivi, job, tentativi e risultati precedenti restano conservati; la deroga è limitata alla nuova selezione e ai discendenti validati.
- **Verifica:** rifiuti atomici, retry e recupero da outage della coda, audit immutabile, classificazione/estrazione del runner, ammissione dei discendenti ed esclusione del lavoro ordinario. [Procedura](../../../operations/municipal-interpretation.md#recupero-selettivo-delle-versioni-archiviate); evidenza privata `CLN-CONTINUATION-20261004`.
- **Seguito:** CLN-022 verifica adozione, recupero live 29.3, risultati persistenti, viste storiche API/MCP e rivalutazione delle campagne; registra l’omissione della nuova risposta. Questo intervento non conclude 9.3/10.1 o l’accettazione.


### CLN-020 — Riapertura parziale e luoghi tipografici nel replay

- **Data e ultima verifica:** 2026-10-04; risposte e testo del caso conservato, senza nuovo fetch o chiamate ai modelli.
- **Ambito e conoscenza:** `calcinaia-municipal`, comunicazione primaria di monitoraggio del territorio; confermata la perdita del luogo per differenza di apostrofo fra valore del modello e citazione canonica. Il successivo conflitto noto/sconosciuto bloccava il merge completo. Il modello usava inoltre un aggiornamento generico per la riapertura esplicita.
- **Intervento:** applicato nel checkout e verificato nel replay/parser e con fixture sintetiche PostgreSQL. La grafia contigua originale viene conservata prima del controllo dell'ambito; una formula positiva delimitata di riapertura già avvenuta conserva il proprio soggetto e la citazione. Tempi della precedente chiusura o della pagina non datano la riapertura. Catalogo di estrazione v25/logica v20, con percorso legacy preservato.
- **Verifica:** una riapertura, due chiusure, quattro divieti e un'attivazione del COC; conflitto temporale dell'altro caso ed eccezione delle strade liberate conservati. Prove negative per negazioni, ipotesi, futuro, citazioni, intestazioni e solo contesto. Evidenza privata `CLN-PARTIAL-REOPEN-20261004`; [procedura e limiti](../../../operations/municipal-interpretation.md).
- **Prossima verifica:** adozione e rielaborazione selezionata del percorso ordinario, confronto persistente/API-MCP e rivalutazione delle campagne. Supera il fallimento tecnico riportato da CLN-016 nel perimetro del replay; non dichiara risolta la completezza del corpus, non sostituisce i report live e non conclude 9.3/10.1 o l'accettazione.

### CLN-017 — Integrazione ordinaria delle misure verificata in development

- **Data e ultima verifica:** 2026-10-03; codice, fixture sintetiche PostgreSQL e controllo HTTPS API/MCP sul servizio development.
- **Ambito e conoscenza:** comportamento verificato per estrazioni primarie complete e recupero delimitato di risultati conservati; supera la lacuna tecnica di CLN-007 nel perimetro provato.
- **Intervento e verifica:** proiezione con binding atomico e ripetizione idempotente, esclusione delle valutazioni e sospensioni, storico senza retrodatazione, evidenze e soggetto consultabili. La provenienza configurata alimenta una valutazione separata senza sostituire esiti espliciti. Pubblicazione manuale e controlli delle fonti sono preservati. API/MCP concordano; paginazione e isolamento di un altro comune verificati. Corretto anche il conflitto dei cataloghi locali fra revisioni, conservando le configurazioni storiche; worker riavviati con successo.
- **Limiti e prossima verifica:** recupero senza nuove chiamate ai modelli; un falso positivo non meteo individuato nel campione è escluso dal replay e conservato nell’evidenza privata. Questo non risolve tutti i falsi positivi futuri, la completezza comunale, la riapertura parziale o l’interpretazione grafica CFR. Accettazione pending nel coverage tracker. Procedura di [pubblicazione development](../../../operations/development-publication.md).


### CLN-016 — Revisione delegata e trial con visibilità revocata

- **Data e ultima verifica:** 2026-10-03; originali conservati, codice corrente e prova development delimitata.
- **Conoscenza:** confermata nel perimetro delle evidenze.
- **Osservazione:** Revisione documentale eseguita dall’assistente su delega esplicita dell’operatore. L’ordinanza distingue chiusure, divieti e sospensione didattica; il conflitto agosto/settembre della pagina resta irrisolto. Le date di pubblicazione non provano l’edizione dell’atto. Il confronto dei due avvisi sui lavori usa la validità letterale presente in entrambi, mantenendo distinta la dipendenza mancante del canale piattaforma.
- **Intervento e verifica:** task 26.5/26.6 completati con ricevute corrette, API/MCP equivalenti, trial development su dati separati e rollback della visibilità Calcinaia. Le ricevute precedenti e gli storici restano conservati; gli altri comuni non sono abilitati. Registro privato `CIN-TRIAL-20261003-01`.
- **Limiti e prossima verifica:** il caso lavori stradali è tecnico, non meteo. Il replay corrente conserva le strade liberate e il conflitto temporale, ma fallisce il merge della riapertura parziale; CLN-007 non è risolto né distribuito. Accettazione pending nel coverage tracker.


### CLN-015 — Acquisizione e verifica delimitate in development

- **Data e ultima verifica:** 2026-10-03, controlli HTTP, originali conservati e ricevute persistenti su servizi development con dati dedicati.
- **Ambito:** una pagina Notizie comunale e 24 documenti, confronto di un avviso locale con il dettaglio della piattaforma selezionato, PDF richiamato dalla pubblicazione primaria; nessuna traversata completa di tutte le sezioni.
- **Conoscenza:** comportamento confermato nel campione. **Intervento:** prova della 26.2 completata; nessuna attivazione delle fonti del trial, controlli e container esistenti preservati.
- **Osservazione ed evidenza:** la primaria locale viene acquisita indipendentemente e produce una sola misura dai campi sostenuti. Il rilancio resta non comparabile, con dipendenza comunale esterna non consentita nel contratto della piattaforma; il PDF è conservato e validato dalla primaria. Date di pubblicazione diverse non provano equivalenza di edizione o validità. Il controllo CFR è non applicabile alla misura locale. Ricevute, errori sintetici e limiti nella [procedura](../../../operations/cittadino-informato.md#prova-completa-in-development--task-262); registro privato `CIN-DEV-20261003-01`.
- **Conseguenza e prossima verifica:** completa la prova runtime dell’ingresso separato di CLN-014. L’assenza dal perimetro temporale riuscito non è assenza dall’intera piattaforma; la finestra più ampia fallita resta incompleta. CLN-007, interpretazione/autorità dell’atto, valutazione estesa 26.5, adozione 26.6 e accettazione restano separati.

### CLN-014 — Proiezione con riscontro primario implementata

- **Data e ultima verifica:** 2026-10-03, codice e fixture sintetiche/PostgreSQL isolato; nessun nuovo confronto di avvisi reali di Calcinaia.
- **Ambito:** ingresso programmatico di verifica per il Comune selezionato, misure locali e fasi operative distinte dai rilanci regionali.
- **Conoscenza:** comportamento software confermato nei casi provati. **Intervento:** 26.4 implementato e testato; nessun cambiamento al runtime live o all’accettazione.
- **Osservazione ed evidenza:** i campi vengono ricavati dagli originali/versioni posseduti, inclusi atti PDF con OCR completo; ricevute dei tre ruoli e ammissione atomica da valori comunali, senza obbligo di allerta regionale. Ripetizioni riusano la misura; revisioni e ritorni conservano lo storico e la selezione per conoscenza. [Procedura](../../../operations/cittadino-informato.md#verifica-persistente-e-ammissione--task-264) e [test](../../../../internal/backend/domain/verification_integration_test.go).
- **Conseguenza e prossima verifica:** crea il percorso di proiezione verificata richiesto dalla 26.4, separato dal worker generico diagnosticato in CLN-007. Eseguire 26.2 con originali reali in development; 26.5/26.6 verificano valutazione estesa e successiva adozione programmata. Nessuna copertura o pubblicazione dedotta dai test.

### CLN-013 — Acquisizione Cittadino Informato delimitata verificata

- **Data e ultima verifica:** 2026-10-03, implementazione, fixture sintetiche/PostgreSQL isolato e due controlli HTTP.
- **Ambito:** canale aggiuntivo Calcinaia, ISTAT `050004`, aggiornamenti del pubblicatore comunale selezionato e rischi oggi/domani.
- **Conoscenza:** confermato nel perimetro verificato. **Intervento:** codice applicato e verificato in isolamento; nessuna modifica o attivazione della fonte live.
- **Osservazione ed evidenza:** originali API JSON e link pubblici separati; date di pubblicazione/visualizzazione conservate con significati distinti, controlli di paginazione, dipendenze e retry. La ripetizione invariata riusa le versioni. [Procedura e test](../../../operations/cittadino-informato.md); evidenza privata `CIN-ACQUIRE-20261003-01`.
- **Conseguenza e prossima verifica:** aggiorna lo stato tecnico di CLN-011 e completa 26.3; i candidati rimangono pending. Implementare confronti persistenti e ammissione in 26.4, compreso il collegamento al dominio municipale di CLN-007, prima di chiudere 26.2.

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

### CLN-010 — Canale aggiuntivo pianificato con riscontro primario

Il precedente prerequisito documentale di questa voce è superato da CLN-012; le evidenze storiche restano conservate.

- **Data e ultima verifica:** 2026-10-03, decisione dell'utente, consultazione web delle condizioni e coerenza documentale.
- **Ambito:** Calcinaia come primo perimetro; altri comuni selezionati richiedono evidenze proprie.
- **Conoscenza:** requisito confermato; base di acquisizione/riuso da accertare. **Intervento:** pianificato, collector non attivato.
- **Osservazione:** il nuovo requisito estende il ruolo diagnostico: candidati verificati nei siti comunali/atti richiamati e, per rilanci regionali, nei prodotti CFR comparabili. Controlli indipendenti restano necessari; mancato riscontro e indisponibilità non diventano conflitti o assenza di allerte.
- **Evidenza:** [specifica aggiornata](../../../../openspec/changes/define-toscana-alert-service/specs/official-source-ingestion/spec.md#requirement-scoped-cittadino-informato-acquisition-and-primary-verification), [note della piattaforma](../../../fonti/piattaforme/cittadino-informato.md), [task 26](../../../../openspec/changes/define-toscana-alert-service/tasks.md). Registro privato `CIN-REVIEW-20261003-01`.
- **Conseguenza:** il riscontro primario non autorizza retroattivamente l'acquisizione e non completa accettazione o pubblicazione; originali e condizioni dei canali restano separati.
- **Prossima verifica:** task 26.2 per titolarità e base applicabile, poi implementazione, valutazione e trial delimitato. CLN-008 resta il riscontro datato del runtime precedente.

### CLN-011 — API candidata e presupposti ancora da chiudere

Il criterio documentale di chiusura di questa voce è superato da CLN-012; le osservazioni tecniche restano conservate.

- **Data e ultima verifica:** 2026-10-03, consultazioni HTTP delimitate del canale e dei riferimenti pubblicati; nessun polling o trattamento tramite provider.
- **Ambito:** pagina Calcinaia e relativo indice REST, ISTAT `050004`.
- **Conoscenza:** accessibilità tecnica e date del contenitore confermate; base del flusso e contratto degli avvisi da accertare. **Intervento:** revisione preliminare documentata; nessuna attivazione effettuata.
- **Osservazione:** l'indice dichiara route comunali per aggiornamenti e rischi. La risposta WordPress della pagina ha corpo vuoto e date 2017–2018, distinte dal bollettino 2026 nell'HTML. La privacy indica ANCI Toscana come titolare del trattamento, senza risolvere diritti sui singoli contenuti o sulla banca dati.
- **Evidenza:** [dossier del task 26.2](../../../fonti/piattaforme/cittadino-informato-review.md), scoperte CIN-004 e CIN-005 nelle [note della piattaforma](../../../fonti/piattaforme/cittadino-informato.md). Ricevute e originali nel registro privato `CIN-REVIEW-20261003-02`.
- **Conseguenza:** non usare date del contenitore per emissione/validità e non considerare l'API esposta come autorizzazione del flusso. CLN-010 resta un requisito pianificato; il task 26.2 non è concluso.
- **Prossima verifica:** accertare base applicabile o condizioni concordate per accesso ricorrente, originali, provider e risultati; poi collaudare il contratto tecnico e il trial Calcinaia previsto, mantenendo primaria comunale/CFR e copie pubbliche link-only.

### CLN-012 — Task 26.2 orientato al sistema di verifica multipla

- **Data e ultima verifica:** 2026-10-03, decisione dell’utente e verifica documentale; nessuna nuova acquisizione o verifica runtime.
- **Ambito:** Calcinaia `050004`, Comune/atti richiamati, Regione Toscana/CFR e cittadinoinformato.it.
- **Conoscenza:** nuovo criterio confermato; gate documentale precedente superato. **Intervento:** specifiche e checklist aggiornate; codice e trial ancora da eseguire.
- **Osservazione ed evidenza:** la [specifica](../../../../openspec/changes/define-toscana-alert-service/specs/official-source-ingestion/spec.md#requirement-scoped-cittadino-informato-acquisition-and-primary-verification) richiede confronti persistenti, esiti per campo e una verifica delimitata del percorso completo per chiudere 26.2; ottenere licenze, accordi o documentazione giuridica non è un prerequisito.
- **Conseguenza:** usare CFR per le affermazioni regionali e Comune/atto per le misure locali; registrare controlli non applicabili, indisponibili, mancati riscontri, non comparabilità e conflitti. Nessuna maggioranza fra rilanci o obbligo di presenza su tutti i canali; candidati senza sostegno primario restano diagnostici.
- **Prossima verifica:** implementare acquisizione 26.3 e confronti 26.4, poi chiudere 26.2 con ricevute del controllo development; completare 26.5/26.6 separatamente. Copie pubbliche link-only, copertura e accettazione restano distinte.
- **Collegamenti:** CIN-006 nelle [note della piattaforma](../../../fonti/piattaforme/cittadino-informato.md), [dossier aggiornato](../../../fonti/piattaforme/cittadino-informato-review.md).

### CLN-018 — Situazione API/MCP centrata su conclusioni e fonti

- **Data e ultima verifica:** 2026-10-03, test sintetici/PostgreSQL e risposta development HTTPS; nessuna nuova acquisizione richiesta dalla vista.
- **Ambito:** situazione di Calcinaia `050004`, fonti pubblicabili nel database operativo e conoscenza fissata; il trial separato di Cittadino Informato non è importato.
- **Conoscenza:** comportamento confermato nel perimetro provato. **Intervento:** presentazione applicata e verificata, task 28; raccolta e accettazione restano distinte.
- **Osservazione ed evidenza:** conclusioni, quattordici fatti regionali e sette misure comunali documentate, riferimenti deduplicati e tre sintesi; la piattaforma senza dati pubblicabili è dichiarata separatamente. Il backlog comunale produce un limite di completezza anziché centinaia di righe e ulteriori pagine. [Contratto e limiti](../../../backend/public-api-mcp.md#situation-response), [verifica development](../../../operations/development-publication.md#situazione-sintetica--3-ottobre-2026); registro privato `CLN-SUMMARY-20261003-01`.
- **Conseguenza e prossima verifica:** consultare i fatti per rischio/validità e gli endpoint documentali per dettagli; non considerare confermate in vigore le misure con date non determinabili e non trasformare l'assenza dei dati della piattaforma in assenza di avvisi. Accettazione, completezza municipale e raccolta continuativa rimangono da verificare.


### CLN-019 — Recupero del provider e pipeline ordinaria verificati

- **Data:** 2026-10-04. **Ultima verifica:** 2026-10-04, sonda autenticata, gate e ledger, risultati OCR e pipeline dei worker; nessuna nuova ricerca web della fonte.
- **Ambito:** originale e risorsa del job OCR storico revisionato e un avviso meteo comunale già conservato, nel solo perimetro della fonte `calcinaia-municipal`.
- **Conoscenza:** confermato sulle esecuzioni delimitate. **Intervento:** ripresa provider applicata e verificata; accettazione non completata.
- **Osservazione:** il rilancio storico già riuscito tramite fallback resta idempotente. Un canary selezionato completa il precedente run OCR remoto fallito; un altro caso ordinario completa classificazione, estrazione e linking senza nuovi rifiuti provider. Il primo caso di classificazione fallisce per citazione non conforme e rimane tale.
- **Evidenza:** [procedura e verifica](../../../operations/acquisition-recovery.md#verifica-del-recupero-provider--4-ottobre-2026); originali, pagine, job, tentativi e ricevute private `PROVIDER-RECOVERY-20261004`.
- **Conseguenza:** rimuove il blocco quota osservato nel periodo precedente, senza trasformare una citazione errata in prova valida o cambiare i controlli della fonte. Embedding e copertura integrale non sono certificati dal recupero.
- **Prossima verifica:** completare regressione e osservazione con confronti attribuibili, conservando i fallimenti e verificando l'accesso del provider nel tempo.
- **Collegamenti:** CLN-005, CLN-017 e CLN-018; task 13.8, 9.3 e 10.1.
