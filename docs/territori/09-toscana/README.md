---
tipo: "regione"
codice_regione: "09"
ultima_revisione: "2026-10-02"
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

## Regole ed eccezioni

| ID | Ambito | Regola corrente | Stato dell'intervento |
| --- | --- | --- | --- |
| TOS-001 | Tre prodotti CFR | Accettazione e risorse necessarie sono distinte per prodotto; il monitoraggio non emette livelli desunti dalla narrativa | Applicato nel parser; accettazione reale separata |
| TOS-002 | Stato del controllo | Completezza della raccolta e pubblicazione attesa sono dimensioni separate | Applicato nello store dei controlli |
| TOS-003 | Formati CFR con policy locale esplicita | Pertinenza positiva da formato rigorosamente riconosciuto; risorse complete ed evidenza letterale obbligatorie, fallback sui casi non riconosciuti | Implementato; TOS-004 aggiunge replay e trial limitato |
| TOS-004 | Originali CFR conservati e trial development | Tabella di vigilanza senza zone pertinente solo dopo parsing rigoroso e titolo letterale; nessun all-clear desunto | Replay e classificazioni persistite verificati; rollout continuativo separato |

Non derivare colori delle mappe dal testo circostante. Le pubblicazioni comunali che rilanciano un bollettino non costituiscono automaticamente una nuova allerta o una misura locale.

## Problemi aperti

Accettazione formale e osservazione dei prodotti rimangono quelle indicate nel coverage tracker e nei [task del servizio Toscana](../../../openspec/changes/define-toscana-alert-service/tasks.md). Verificare sul canale operativo il monitoraggio e i casi senza evento. In presenza di `missing`, verificare il percorso che alimenta la data della pubblicazione: gli esiti per ambiente appartengono all'archivio privato.

## Comuni documentati

| Comune | ISTAT | Scheda |
| --- | --- | --- |
| Calcinaia | `050004` | [Guida comunale](comuni/050004-calcinaia.md) |
| Cascina | `050008` | [Guida comunale](comuni/050008-cascina.md) |
| Livorno | `049009` | [Guida comunale](comuni/049009-livorno.md) |
| Pisa | `050026` | [Guida comunale](comuni/050026-pisa.md) |
| Pontedera | `050029` | [Guida comunale](comuni/050029-pontedera.md) |

## Registro delle scoperte

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
