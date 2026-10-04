---
tipo: "regione"
codice_regione: "09"
ultima_revisione: "2026-10-04"
---

# Toscana

## Identità e ambito

Regione Toscana, codice `09`. Questa scheda distingue i prodotti del Centro Funzionale Regionale (CFR) dalle pubblicazioni dei singoli comuni. Il [coverage tracker](../../coverage.md#region-09) descrive lo stato pubblico documentato; questa guida non certifica accettazione o disponibilità live.

## Mappa delle fonti

| Prodotto | Identificativo IWA | Riferimento | Ambito |
| --- | --- | --- | --- |
| Vigilanza meteorologica | `cfr-vigilance` | [Bollettino CFR](https://cfr.toscana.it/index.php?IDS=2&IDSS=71) | Prodotto regionale distinto dalla criticità |
| Valutazione delle criticità | `cfr-criticality` | [Bollettino CFR](https://cfr.toscana.it/index.php?IDS=2&IDSS=76) | Livelli, rischi, zone e validità secondo le evidenze del prodotto |
| Monitoraggio evento | `cfr-monitoring` | [Canale CFR](https://www.cfr.toscana.it/mobile3/avviso-criticita/) | Canale previsto dalla specifica; verifica operativa separata |

Gli identificativi aiutano a riconoscere le fonti nel registro dell'ambiente; non ne indicano l'attivazione. Consultazione documentale: 2026-10-02. Il canale di monitoraggio è riferito dalla [specifica di acquisizione](../../../openspec/changes/define-toscana-alert-service/specs/official-source-ingestion/spec.md); la consultazione web usata per questa scheda non ne ha verificato il contenuto.

## Guida operativa

TOS-019 verifica la chiusura della campagna MVP 9.3 sui quattro ambiti: confronti
originali/allegati corretti con audit, cadenza e ritardi documentati e prove di
errore/evento assente attribuite. La campagna interna precedente conserva i suoi
ritardi; 10.1 e accettazione restano separati.


Il [consuntivo delle campagne](../../operations/trial-costs.md) distingue costi
noti, metriche mancanti e prezzi non verificati per il periodo. I quattro comuni
successivi hanno [report individuali pending](../../operations/municipal-acceptance.md)
e campagne indipendenti con gli stessi criteri del pilota. TOS-016 registra la
verifica del percorso; non completa osservazione o accettazione delle fonti.

Verificare separatamente pagina, risorse grafiche e versione stampabile richieste dal contratto del prodotto. Conservare emissione, validità e acquisizione come informazioni distinte. Il [parser regionale](../../../internal/backend/acquisition/regional.go) richiede contenuto riconoscibile e una espressione di emissione, salvo lo stato esplicito di assenza di evento del monitoraggio.

Consultare lo [stato dei controlli](../../../internal/backend/acquisition/schedule.go) per distinguere raggiungibilità, completezza, ritardo e pubblicazione attesa. `publication_state=missing` significa che la pubblicazione attesa non è stata osservata secondo la configurazione; non dimostra da solo che il CFR non abbia pubblicato. Controllare cadenza configurata, data osservata e parsing prima di attribuire il problema alla fonte.

La capacità opzionale [elaborazione locale](../../operations/local-processing.md)
può riconoscere la pertinenza di un formato regionale già revisionato, usando
il parser rigoroso delle tabelle o lo stato esplicito senza evento. Richiede una
policy `local_processing` del prodotto; non interpreta nuove mappe, non inventa
livelli e non rende completa una risorsa mancante. Le policy regionali non possono
dichiarare PDF grafici come documenti di solo testo. Le verifiche del 2026-10-02 comprendono fixture, replay di originali conservati
e un trial development limitato. La vigilanza con tabella riconosciuta ma senza
zone elencate resta pertinente, senza dedurre un all-clear. TOS-004 distingue
queste prove dall’attivazione di una policy continuativa e dall’accettazione.

La [pubblicazione manuale in development](../../operations/development-publication.md) consente di scegliere singoli
comuni per la consultazione API/MCP, includendo i prodotti regionali applicabili.
La scelta è revocabile e distinta dalla raccolta e dall’accettazione delle fonti.
Gli stati pendenti e i livelli non verificati restano espliciti; staging e produzione
mantengono i requisiti della specifica. Vedere TOS-007.

La [lettura grafica CFR](../../operations/cfr-graphics.md) usa i PDF conservati e
la stessa edizione HTML. La criticità legge livelli e maschere di applicabilità;
la vigilanza conserva bande di pioggia media sull’area, cumulati distinti e simboli
dei fenomeni, con livello d’allerta non applicabile. `not_depicted` descrive la
grafica, senza attestare assenza di rischio. Risorse, simboli o associazioni non
riconosciuti restano irrisolti. TOS-015 registra il collaudo del software e degli
originali disponibili; TOS-017 registra la successiva adozione in development,
mentre l’accettazione resta separata.

## Regole ed eccezioni

| ID | Ambito | Regola corrente | Stato dell'intervento |
| --- | --- | --- | --- |
| TOS-001 | Tre prodotti CFR | Accettazione e risorse necessarie sono distinte per prodotto; il monitoraggio non emette livelli desunti dalla narrativa | Applicato nel parser; accettazione reale separata |
| TOS-002 | Stato del controllo | Completezza della raccolta e pubblicazione attesa sono dimensioni separate | Applicato nello store dei controlli |
| TOS-003 | Formati CFR con policy locale esplicita | Pertinenza positiva da formato rigorosamente riconosciuto; risorse complete ed evidenza letterale obbligatorie, fallback sui casi non riconosciuti | Implementato; TOS-004 aggiunge replay e trial limitato |
| TOS-004 | Originali CFR conservati e trial development | Tabella di vigilanza senza zone pertinente solo dopo parsing rigoroso e titolo letterale; nessun all-clear desunto | Replay e classificazioni persistite verificati; rollout continuativo separato |
| TOS-015 | Mappe CFR conservate | Edizione/date concordanti, 26 zone e sette rischi; vigilanza distinta dai colori di criticità, cumulati distinti dai giorni | Verificato sui quattro campioni conservati e controlli sintetici; accettazione pending |

Non derivare colori delle mappe dal testo circostante. Le pubblicazioni comunali che rilanciano un bollettino non costituiscono automaticamente una nuova allerta o una misura locale.

Il flusso di [Cittadino Informato](../../fonti/piattaforme/cittadino-informato.md)
aggiunge discovery per i comuni selezionati e verifica primaria; i rilanci
regionali si confrontano con rischio, zona, emissione e validità del prodotto CFR
originario. Le raccolte CFR e comunali restano indipendenti e una misura locale
non richiede un'allerta regionale. TOS-009 registra il piano; TOS-010 documenta
il verificatore implementato e le prove sintetiche; TOS-011 verifica il percorso
development delimitato senza adozione programmata.

Per diagnosticare una situazione API vuota, controllare prima `coverage.public_state`
e `coverage_status`: controlli di acquisizione recenti e `updating.state=ok` non
abilitano la pubblicazione. La query pubblica richiede `public_enabled` e una
acquisizione collegata a una revisione con evento di accettazione; le viste private
territoriali hanno un percorso distinto. `interpretation.state=not_processed`
con la limitazione `source acceptance is pending` descrive questa barriera e non
dimostra che nessuna elaborazione interna sia avvenuta. Separatamente, il parser
conserva la dichiarazione di assenza di criticità con livelli per zona `unknown`
quando i colori delle mappe non sono verificati. Vedere TOS-005.

## Problemi aperti

TOS-015 estende il collaudo di TOS-014 alle 26 zone e ai sette rischi dei prodotti
conservati: `cfr-graphics-v4` risolve i posizionamenti revisionati A6/I e i
riempimenti separati, interpreta la vigilanza con le proprie legende e preserva
lo storico. Le maschere reticolo principale/mareggiate restano distinte dal verde;
gli intervalli precisi HTML prevalgono sulla mappa giornaliera. La sola
dichiarazione `NESSUNA` non produce colori.

Restano da acquisire per l’accettazione ulteriori casi reali positivi di
vento/mare/neve/ghiaccio, verificati qui come controlli sintetici. Nuove legende,
mappe raster, geometrie non supportate e sigle ambigue restano irrisolte o usano
il fallback HTML sostenuto. Il collaudo non certifica tutti i formati futuri, la
completezza delle misure comunali o l’accettazione. Le valutazioni di provenienza
di TOS-013 restano distinte; questo intervento non modifica servizi o controlli live.

Accettazione formale e osservazione dei prodotti rimangono quelle indicate nel coverage tracker e nei [task del servizio Toscana](../../../openspec/changes/define-toscana-alert-service/tasks.md). Verificare sul canale operativo il monitoraggio e i casi senza evento. In presenza di `missing`, verificare il percorso che alimenta la data della pubblicazione: gli esiti per ambiente appartengono all'archivio privato.

Il precollaudo di TOS-006 richiede di consultare il report della campagna oltre
al tempo trascorso: confronti con gli originali, revisione delle risorse necessarie
e prove di errore sono evidenze separate. Per il monitoraggio senza nuovi eventi,
documentare anche un caso conservato di evento assente dal trial. Le verifiche
tecniche del software sono distinte dagli esiti attesi revisionati; TOS-012
registra la revisione dell’assistente su delega esplicita dell’operatore.

## Comuni documentati

| Comune | ISTAT | Scheda |
| --- | --- | --- |
| Calcinaia | `050004` | [Guida comunale](comuni/050004-calcinaia.md) |
| Cascina | `050008` | [Guida comunale](comuni/050008-cascina.md) |
| Livorno | `049009` | [Guida comunale](comuni/049009-livorno.md) |
| Pisa | `050026` | [Guida comunale](comuni/050026-pisa.md) |
| Pontedera | `050029` | [Guida comunale](comuni/050029-pontedera.md) |

## Registro delle scoperte

### TOS-015 — Grafica CFR per tutte le zone dei campioni conservati

- **Data e ultima verifica:** 2026-10-04; replay di due PDF di criticità e due di vigilanza conservati, controlli sintetici, PostgreSQL isolato e API/MCP con client reale.
- **Ambito e conoscenza:** comportamento verificato per 26 zone, sette rischi e due giorni in ciascun campione. Il PDF storico con criticità gialla include pagine preliminari; il suo HTML è una trascrizione revisionata della tabella, non una nuova acquisizione. La vigilanza reale comprende temporali presenti e un’edizione senza simboli; i casi positivi degli altri quattro fenomeni sono sintetici.
- **Osservazione ed evidenza:** la versione precedente perdeva riempimenti separati, sigle A6/I al margine e mappe dopo pagine narrative. Il renderer SVG fallisce su un’edizione storica; una normalizzazione vettoriale locale dei font consente la lettura conservando la pagina fisica. [Parser](../../../internal/backend/acquisition/regional_vector.go), [vigilanza](../../../internal/backend/acquisition/regional_vigilance_vector.go), [prove sintetiche](../../../internal/backend/acquisition/regional_vigilance_vector_test.go) e [procedura](../../operations/cfr-graphics.md). Originali, hash e confronto per zona/rischio/giorno restano privati: `CFR-GRAPHICS-20261004-01`.
- **Intervento e verifica:** interpretazione con legende del prodotto, periodi giornalieri/cumulati separati e maschere di applicabilità; 364 valori determinati per campione, inclusi i valori gialli del campione storico. Proiezione append-only, idempotenza, fallback su risorse mancanti e metadati weather equivalenti API/MCP verificati. Nessun nuovo colore deriva dalla vigilanza.
- **Limiti e prossima verifica:** completa il task software 27.4 nei formati riconosciuti e supera la lacuna grafica registrata in TOS-008/TOS-014 per il campione. Servono adozione live distinta, casi reali aggiuntivi e collaudo della campagna prima dell’accettazione; nessuna raccolta o pubblicazione viene abilitata.

### TOS-014 — Criticità della zona A4 determinata dai poligoni PDF

- **Data e ultima verifica:** 2026-10-03; originali conservati, lettura diretta delle mappe, fixture sintetiche e controllo HTTPS API/MCP in development.
- **Ambito e conoscenza:** criticità corrente per Calcinaia/A4, sette rischi per oggi/domani. Il comportamento è verificato nel perimetro indicato, senza estenderlo ai prodotti grafici non interpretati.
- **Intervento e verifica:** lettura dei contorni/colori vettoriali con sigle e date del PDF; concordanza dell’emissione HTML/PDF, provenienza alla pagina, storico senza retrodatazione e API/MCP equivalenti. I poligoni noti producono valori determinati; colori sconosciuti, sigle ambigue, risorse mancanti o discordanti conservano l’incertezza. Gli intervalli precisi delle righe esplicite non diventano allerte dalla mezzanotte. Verifiche [software](../../../internal/backend/acquisition/regional_vector_test.go) e [procedura](../../operations/development-publication.md).
- **Limiti e prossima verifica:** alcune sigle sono esterne o al margine dei poligoni; questi campi restano unknown, senza contaminare la valutazione dei fatti A4 sostenuti. Il campione non completa tutti i posizionamenti, la vigilanza grafica, la validità delle misure comunali o l’accettazione. Originali, risultati e rollback restano privati.


### TOS-013 — Provenienza configurata collegata alla qualità API

- **Data e ultima verifica:** 2026-10-03; fixture sintetiche PostgreSQL e servizio development verificato tramite HTTPS API/MCP.
- **Ambito e conoscenza:** valutazione della provenienza dei tre prodotti CFR configurati e della primaria Calcinaia, separata dalla lettura delle mappe.
- **Intervento e verifica:** registrazione append-only dell’evidenza ufficiale già presente nella configurazione attiva, con tempo di valutazione effettivo; valutazioni esplicite precedenti sono preservate. Storico e controlli delle fonti invariati. L’integrazione comunale è descritta da [CLN-017](comuni/050004-calcinaia.md#cln-017--integrazione-ordinaria-delle-misure-verificata-in-development).
- **Limiti e prossima verifica:** la provenienza verificata non certifica livelli, rischi, applicabilità o validità grafica. TOS-008 resta aperto e nessun livello `unknown` viene sostituito con verde. Accettazione dei prodotti pending.


### TOS-012 — Revisione degli originali CFR su delega e limiti confermati

- **Data e ultima verifica:** 2026-10-03; originali conservati, codice corrente e prova development delimitata.
- **Conoscenza:** confermata nel perimetro delle evidenze.
- **Osservazione:** Confrontati gli originali conservati e tutte le pagine dei PDF di criticità e vigilanza, con emissioni distinte e risorse integre. La vigilanza conserva quantità di pioggia e fenomeni, non colori di allerta. Il monitoraggio conserva la dichiarazione no_event; l’esempio dell’Allegato5 resta datato marzo 2025. La lettura manuale delle mappe non modifica i livelli unknown del parser live TOS-008.
- **Intervento e verifica:** revisione attribuita all’assistente su delega esplicita, con esiti fail/unresolved dove l’estrazione resta incompleta. Il trial 26.6 verifica controlli regionali e comunali indipendenti, attribuzione API/MCP e rollback; non accetta i prodotti CFR. Le configurazioni e i controlli live rimangono preservati.
- **Limiti e prossima verifica:** accettazione per prodotto pending; completare interpretazione grafica, prove di errore e valutazioni previste nei task 9.3/10.1. Un originale invariato può sostenere la revisione di una nuova configurazione soltanto con acquisizione finalizzata precedente che dimostri quel riuso; i test rifiutano riuso non provato o futuro.


### TOS-011 — Rilancio regionale verificato come diagnostico in development

- **Data e ultima verifica:** 2026-10-03, acquisizione CFR HTTP indipendente, originali/versioni e confronto persistente sullo sviluppo con dati dedicati.
- **Ambito:** un prodotto di criticità originario e il rilancio dei rischi oggi/domani nel canale Calcinaia della piattaforma; confronto regionale distinto dalla misura locale.
- **Conoscenza:** non comparabilità confermata nei campi disponibili. **Intervento:** percorso delimitato della 26.2 verificato, nessuna nuova lettura dei colori o attivazione programmata.
- **Osservazione ed evidenza:** la ricevuta conserva il CFR e la piattaforma senza inferire prodotto/zona/emissione/validità mancanti o verde dalle etichette e dalla legenda. Il caso reale resta diagnostico; i controlli sintetici regionali verificano corroborazione e conflitto, con il livello primario preservato contro due rilanci concordi. [Procedura](../../operations/cittadino-informato.md#prova-completa-in-development--task-262); registro privato `CIN-DEV-20261003-01`.
- **Conseguenza e prossima verifica:** completa la prova dell’ingresso implementato in TOS-010, senza risolvere l’interpretazione grafica TOS-008 o l’accettazione. Completare valutazione estesa e attribuzione API/MCP delle ricevute in 26.5 prima della successiva adozione 26.6.

### TOS-010 — Confronti dei rilanci con primaria regionale verificati

- **Data e ultima verifica:** 2026-10-03, codice e fixture sintetiche/PostgreSQL isolato; nessun nuovo fetch CFR.
- **Ambito:** confronto di prodotti regionali con rilanci comunali/piattaforma per il territorio selezionato, distinto dalle misure locali.
- **Conoscenza:** comportamento software confermato. **Intervento:** task 26.4 implementato e testato; nessuna attivazione o accettazione dei prodotti CFR.
- **Osservazione ed evidenza:** identità di prodotto/rischio/zona/emissione/validità esplicita obbligatoria per comparare; periodi diversi e `unknown` non producono colori corroborati. Due rilanci concordi in conflitto con la primaria conservano il conflitto e il livello CFR; il monitoraggio non inventa livelli. [Verificatore](../../../internal/backend/domain/verification_store.go), [test](../../../internal/backend/domain/verification_integration_test.go) e [procedura](../../operations/cittadino-informato.md#verifica-persistente-e-ammissione--task-264).
- **Conseguenza e prossima verifica:** implementa il confronto pianificato in TOS-009; non completa l’interpretazione grafica di TOS-008 o la correttezza semantica delle selezioni. Verificare il percorso reale della 26.2 e la valutazione 26.5 prima del trial 26.6.

### TOS-001 — Prodotti regionali con semantiche distinte

- **Data:** consolidamento 2026-10-02; data originaria della scoperta non disponibile.
- **Ultima verifica:** 2026-10-02, documentale sul codice e sulle specifiche.
- **Ambito:** `cfr-vigilance`, `cfr-criticality`, `cfr-monitoring`.
- **Conoscenza:** confermato. **Intervento:** applicato; verifica operativa e accettazione separate.
- **Osservazione:** il parser distingue i tre prodotti e non ricava nuovi livelli dalla narrativa del monitoraggio.
- **Evidenza:** [parser e osservazioni regionali](../../../internal/backend/acquisition/regional.go), [specifica di acquisizione](../../../openspec/changes/define-toscana-alert-service/specs/official-source-ingestion/spec.md).
- **Conseguenza:** un prodotto non eredita permessi, accettazione o livelli di un altro.
- **Prossima verifica:** confrontare ciascun prodotto con originali e risorse richieste nel suo percorso di accettazione.

### TOS-002 — Fetch riuscito e pubblicazione osservata sono distinti

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, documentale.
- **Ambito:** stato di acquisizione e cadenza dei prodotti regionali.
- **Conoscenza:** confermato. **Intervento:** applicato nello store; eventuali difetti nella data osservata restano da verificare.
- **Osservazione:** completezza ed evidenza della pubblicazione attesa sono persistite separatamente.
- **Evidenza:** [ScheduleStore e publicationStateAt](../../../internal/backend/acquisition/schedule.go).
- **Conseguenza:** un controllo completo può coesistere con `publication_state=missing`; occorre diagnosticare le due dimensioni separatamente.
- **Prossima verifica:** controllare configurazione ed evidenza di emissione nell'ambiente interessato; non dedurre l'assenza di un bollettino dal solo stato.


### TOS-003 — Pertinenza regionale da formato strutturato revisionato

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, codice e fixture sintetiche.
- **Ambito:** capacità per `cfr-vigilance`, `cfr-criticality`, `cfr-monitoring` con una policy `local_processing` esplicita; nessuna verifica web o attivazione live.
- **Conoscenza:** confermato sul codice; equivalenza del formato live da verificare. **Intervento:** implementato e verificato in repository; rollout non eseguito.
- **Osservazione:** il riconoscimento rigoroso di fatti espliciti o dello stato di monitoraggio senza evento permette una decisione positiva locale senza chiamata di classificazione. Formati non riconosciuti mantengono il percorso ordinario; dati incompleti restano indeterminati.
- **Evidenza:** [classificatore strutturato](../../../internal/backend/classification/structured.go), [verifica persistente sintetica](../../../internal/backend/interpretation/local_processing_integration_test.go), [specifica di acquisizione](../../../openspec/changes/define-toscana-alert-service/specs/official-source-ingestion/spec.md).
- **Conseguenza:** si riducono le chiamate sui formati revisionati senza dedurre colori dalle mappe; policy e risultati hanno identità separate. Non si dichiara un risparmio misurato o accettazione della fonte.
- **Prossima verifica:** confrontare formati e risorse del singolo prodotto con originali, revisionare la policy e misurare il canary prima dell'attivazione.


### TOS-004 — Replay regionale e trial limitato con evidenza reale

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, originali conservati, fixture, PostgreSQL e trial development limitato; nessun nuovo fetch certificato dalla prova.
- **Ambito:** quindici versioni conservate dei tre prodotti CFR nel replay; una versione per prodotto nel trial di classificazione persistente.
- **Conoscenza:** confermato sul campione. **Intervento:** verifica applicata e correzione della vigilanza implementata; policy continuative e accettazione separate.
- **Osservazione:** le quindici versioni complete producono pertinenza locale; tredici hanno confronto storico concordante. La prima prova rileva fallback non necessario quando la tabella di vigilanza riconosciuta non elenca zone. La correzione richiede parsing rigoroso di fenomeni/date e titolo del bollettino come evidenza letterale. Le tre classificazioni del trial sono persistite senza chiamate al modello.
- **Evidenza:** [classificatore e confini](../../../internal/backend/classification/structured.go), [regressioni sintetiche](../../../internal/backend/classification/local_processing_test.go), [risultati e limiti del trial](../../operations/local-processing.md). Selezione e originali restano nel registro privato.
- **Conseguenza:** il prodotto resta pertinente anche senza zone elencate; nessun nuovo livello, all-clear o misura locale è desunto. Il worker diagnostico termina e conserva i risultati; non attiva una policy generale della fonte.
- **Prossima verifica:** estendere il confronto a eventi significativi e alle interpretazioni downstream prima di un rollout continuativo.
- **Collegamenti:** TOS-003 e task 20–21 del servizio Toscana.

### TOS-005 — Situazione pubblica vuota e livelli non determinati hanno cause distinte

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, codice, consultazione del bollettino ufficiale e controlli operativi in sola lettura.
- **Ambito:** query di situazione pubblica per un comune toscano e proiezione del prodotto `cfr-criticality`; configurazioni ed esiti per ambiente restano privati.
- **Conoscenza:** confermato sul codice. **Intervento:** verifica diagnostica eseguita; nessuna attivazione o correzione applicata.
- **Osservazione:** la visibilità pubblica dipende da pubblicazione e acquisizioni accettate, mentre la freschezza dei controlli è separata. La dichiarazione testuale di assenza di criticità è conservata senza assegnare colori alle singole zone. La pagina ufficiale consultata riportava emissione del 2 ottobre 2026 alle 11:56 e validità del 2 e 3 ottobre; questa osservazione non certifica nuove capacità grafiche.
- **Evidenza:** [filtro pubblico e query amministrative](../../../internal/backend/publicquery/administrative.go), [copertura](../../../internal/backend/publicquery/store.go), [proiezione regionale](../../../internal/backend/acquisition/regional_facts.go), [bollettino CFR](https://www.cfr.toscana.it/index.php?IDS=2&IDSS=76). Riferimento operativo privato: `SITUATION-20261002-01`.
- **Conseguenza:** un array vuoto non è una dichiarazione di assenza di allerte; abilitare una fonte non risolve da solo i livelli grafici non determinati.
- **Prossima verifica:** completare la verifica e accettazione del prodotto; valutare separatamente una rappresentazione esplicita dello stato non disponibile e della dichiarazione ufficiale con validità e provenienza.
- **Collegamenti:** [specifica di accesso pubblico](../../../openspec/changes/define-toscana-alert-service/specs/public-alert-access/spec.md), [task 7.3 e 10.1](../../../openspec/changes/define-toscana-alert-service/tasks.md); CLN-004.

### TOS-006 — Il tempo di osservazione non completa il collaudo delle fonti

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, report delle campagne e test su fixture sintetiche con PostgreSQL isolato.
- **Ambito:** precollaudo dei tre prodotti CFR e del pilota Calcinaia; report, revisioni attive ed esiti operativi sono conservati privatamente.
- **Conoscenza:** confermato sul contratto e sul codice. **Intervento:** valutazioni della campagna registrate; accettazione e abilitazione non completate.
- **Osservazione:** le campagne distinguono durata, cadenza, originali, allegati, errori ed eventi assenti. Una campagna con durata sufficiente può restare `extended` per le revisioni mancanti. Il software può essere verificato con fixture senza certificare il comportamento delle fonti reali.
- **Evidenza:** [report e valutazioni delle campagne](../../../internal/backend/observation/store.go), [vincoli di accettazione](../../../internal/backend/registry/store.go), [regressioni](../../../internal/backend/registry/regression.go), [requisito di qualità misurata](../../../openspec/changes/define-toscana-alert-service/specs/official-source-ingestion/spec.md). Riferimento privato: `ACCEPTANCE-PRECHECK-20261002-01`.
- **Conseguenza:** conservare gli esiti incompleti e predisporre il dossier delle verifiche mancanti prima di presentare una fonte come accettata; un permesso di pubblicazione non è un risultato di collaudo.
- **Prossima verifica:** completare confronti, prove di errore e casi conservati; verificare sette rischi, API/MCP e storico negli ambiti dichiarati prima della revisione finale e dell'abilitazione indipendente.
- **Collegamenti:** CLN-005 e [task 9.3, 10.1 e 24](../../../openspec/changes/define-toscana-alert-service/tasks.md).

### TOS-007 — Scelta comunale di pubblicazione per development

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, fixture, PostgreSQL isolato, controllo nativo nel browser e confronto API/MCP in development; dettagli operativi conservati separatamente.
- **Ambito:** Comuni toscani selezionati in development e prodotti CFR applicabili.
- **Conoscenza:** comportamento implementato; non certifica accettazione reale. **Intervento:** controlli separati per ambiente e comune.
- **Osservazione:** un operatore può pubblicare dati comunali conservati e fatti regionali applicabili senza completare le verifiche di produzione. La scelta non si estende agli altri comuni della stessa zona, non cambia l’accettazione né autorizza copie pubbliche. Revoca e cambio di ambito fanno scadere le viste API/MCP conservate.
- **Evidenza:** [controlli operativi](../../operations/development-publication.md) e [specifica pubblica](../../../openspec/changes/define-toscana-alert-service/specs/public-alert-access/spec.md#requirement-manual-municipality-publication-in-development).
- **Conseguenza:** in development dati consultabili possono accompagnare `public_state=pending` e `development_publication=true`; nessuna misura locale o colore regionale viene dedotto per colmare dati mancanti.
- **Prossima verifica:** usare i controlli separati per la consultazione di sviluppo; completare le revisioni e le rivalutazioni registrate prima della pubblicazione verificata in produzione.

### TOS-008 — La consultazione in development non completa la lettura delle mappe

- **Data:** 2026-10-03. **Ultima verifica:** 2026-10-03, codice e diagnosi development in sola lettura; nessun nuovo fetch ufficiale.
- **Ambito:** proiezione `cfr-criticality` con parser `cfr-html-v2` e dimensioni di qualità della query pubblica.
- **Conoscenza:** confermato sul codice. **Intervento:** diagnosi registrata; nessuna nuova interpretazione grafica o accettazione applicata.
- **Osservazione:** `ProjectRegionalHTML` espande la dichiarazione esplicita di assenza di criticità sulle date, i rischi e le zone riconosciuti con livello `unknown`. Le righe tabellari esplicite possono fornire livelli; i colori delle mappe restano fuori dalla proiezione. `quality` legge una valutazione di provenienza separata e restituisce `unresolved` se manca.
- **Evidenza:** [proiezione HTML](../../../internal/backend/acquisition/regional_facts.go), [proiezione persistente](../../../internal/backend/domain/regional_projection.go), [qualità e copertura](../../../internal/backend/publicquery/store.go). Ricevute e risposte development: registro privato `SITUATION-20261003-01`.
- **Conseguenza:** una risposta con quattordici combinazioni per due date e sette rischi non dimostra lettura dei colori; controlli recenti e pubblicazione manuale non colmano i livelli sconosciuti o le valutazioni di provenienza mancanti.
- **Prossima verifica:** definire e validare separatamente la lettura delle evidenze grafiche e la rappresentazione della dichiarazione ufficiale; preservare validità e casi non supportati. Completare le valutazioni di provenienza con evidenza revisionata.
- **Collegamenti:** TOS-005, TOS-007 e [CLN-007](comuni/050004-calcinaia.md#cln-007--estrazioni-conservate-senza-proiezione-comunale-nella-situazione).

### TOS-009 — Riscontro CFR dei rilanci scoperti sul canale aggiuntivo

- **Data e ultima verifica:** 2026-10-03, decisione dell'utente e coerenza delle specifiche; nessun nuovo fetch CFR o attivazione.
- **Ambito:** candidati regionali da Cittadino Informato per comuni selezionati.
- **Conoscenza:** requisito confermato. **Intervento:** pianificato; confronto persistente da implementare.
- **Osservazione ed evidenza:** il [nuovo requisito](../../../openspec/changes/define-toscana-alert-service/specs/official-source-ingestion/spec.md#requirement-scoped-cittadino-informato-acquisition-and-primary-verification) conserva origine CFR e confronto per prodotto, rischio, zona, emissione e validità. Dati non comparabili e fonte primaria indisponibile restano distinti da conflitti.
- **Conseguenza e prossima verifica:** la piattaforma non colma per inferenza i colori CFR `unknown`; implementare e collaudare i confronti nei task 26, mantenendo separati misure locali, condizioni di riuso e accettazione. Vedere [CLN-010](comuni/050004-calcinaia.md#cln-010--canale-aggiuntivo-pianificato-con-riscontro-primario).


### TOS-016 — Consuntivo e report municipali senza accettazione implicita

- **Data:** 2026-10-04. **Ultima verifica:** 2026-10-04, report privati, registro e test sintetici/PostgreSQL; nessuna nuova attestazione sulla disponibilità dei siti.
- **Ambito:** due campagne MVP esistenti e quattro comuni successivi: Livorno, Pisa, Pontedera e Cascina.
- **Conoscenza:** confermato sul servizio e sui report. **Intervento:** scope municipale implementato, consuntivi e report individuali pending consegnati.
- **Osservazione:** le campagne precedenti restano MVP e incomplete. Le nuove campagne per singolo comune mantengono durata e prove, mentre condizioni irrisolte impediscono l'avvio. Il consuntivo distingue i subtotali noti dalle metriche indisponibili; un listino corrente non verifica retroattivamente i prezzi storici.
- **Evidenza:** [collaudo municipale](../../operations/municipal-acceptance.md), [consuntivo](../../operations/trial-costs.md), [migrazione e vincoli](../../../internal/backend/observation/migrate.go). Rapporti, ricevute e manifest negli archivi privati `ROLLOUT-GATES-20261004` e `TRIAL-COSTS-20261004`.
- **Conseguenza:** chiude i task di rendicontazione e predisposizione dei gate indipendenti; budget, collaudo reale, accettazione e copertura completa restano pendenti.
- **Prossima verifica:** completare i confronti e le prove mancanti per ogni ambito, poi riesaminare report di accettazione e budget.
- **Collegamenti:** TOS-006, LIV-008, PIS-003, PON-002 e CAS-007; task 9.3, 9.4, 10.1 e 11.5.


### TOS-017 — Adozione development e confini del collaudo

- **Data:** 2026-10-04. **Ultima verifica:** 2026-10-04, immagini/configurazioni runtime, API privata e client SDK API/MCP sulla stessa vista salvata.
- **Ambito:** servizio development, scope delle campagne e software grafico CFR della revisione applicativa verificata; nessuna nuova attestazione di disponibilità o completezza delle fonti.
- **Conoscenza:** confermato sul runtime. **Intervento:** adozione verificata, con materiali di rollback privati.
- **Osservazione:** le applicazioni attive condividono l'immagine della revisione pulita. Le campagne precedenti conservano lo scope MVP; i quattro tentativi municipali restano rifiutati per prerequisiti irrisolti. Il confronto API/MCP usa una stessa vista salvata, senza dedurre livelli mancanti o accettazione dai controlli HTTP.
- **Evidenza:** [adozione e controlli](../../operations/toscana-readiness.md#adozione-in-development), [gates municipali](../../operations/municipal-acceptance.md). Configurazioni, ricevute e confronti privati `REMAINING-TOSCANA-20261004`.
- **Conseguenza:** supera il rinvio dell'adozione in TOS-015 per development, preservando limiti dei formati e controlli delle fonti. PBS, prove mancanti e accettazione restano pendenti.
- **Prossima verifica:** osservare il percorso ordinario sui prodotti/casi dichiarati e completare ogni dossier reale prima dell'accettazione e dell'abilitazione indipendente.
- **Collegamenti:** TOS-006, TOS-015 e TOS-016; task 9.3, 10.1, 11.5 e 27.4.

### TOS-018 — Confronti degli originali e correzioni tracciate

- **Data:** 2026-10-04.
- **Ambito:** originali conservati CFR di criticità/vigilanza, tutte le 26 zone, sette rischi e due giornate per prodotto.
- **Osservazione confermata:** il PDF di criticità include un simbolo curvo sopra le mappe che il lettore precedente rifiutava; nessuna forma interna ai pannelli viene ignorata.
- **Correzione verificata:** esclusione solo se tutti i punti di controllo trasformati della curva sono strettamente sopra le mappe. Test sintetici rifiutano curve interne, attraversamenti e comandi sconosciuti.
- **Confronto delegato:** 364 campi per prodotto confrontati con gli originali; criticità verde/applicabilità distinta, vigilanza quantità e assenza di simboli operativi senza nuovi colori di allerta.
- **Evidenza:** materiali privati `TOSCANA-TRIAL-CLOSURE-20261004`; [stato](../../operations/toscana-readiness.md#correzioni-preparate-per-la-93).
- **Prossima verifica:** adozione ordinaria e valutazione di campagna con audit correttivo; accettazione resta indipendente.

### TOS-019 — Campagna MVP completa con audit correttivo

- **Data:** 2026-10-04.
- **Ambito:** vigilanza, criticità/allerta, monitoraggio CFR e Calcinaia, sezioni/configurazioni della campagna MVP iniziata il 24 settembre.
- **Comportamento verificato:** valutazione persistita `complete`, 239,94 ore e nove intervalli completi di 24 ore; undici date locali per fonte. Tutti i ritardi entro 10.800 secondi, nessuna risorsa richiesta mancante.
- **Confronti:** due originali grafici verificati su 364 campi ciascuno, monitoraggio senza evento con caso conservato, allegati/non-applicabilità e guasti reali/sintetici distinguibili. Correzioni, motivi e delega conservano revisioni e valutazioni precedenti.
- **Evidenza:** [rapporto osservativo](../../operations/observational-trial.md#esito-verificato-del-4-ottobre-2026), materiali privati `TOSCANA-TRIAL-CLOSURE-20261004`; CLN-025 per worker e API/MCP.
- **Conseguenza:** chiusura della sola 9.3. La campagna interna resta `extended` per ritardi storici; controlli delle fonti e produzione non vengono abilitati dalla valutazione.
- **Prossima verifica:** collaudo 10.1 e gate operativi; accettazione resta nel [coverage tracker](../../coverage.md#region-09).
