# Acquisizione delimitata di Cittadino Informato

I task 26.3 e 26.4 implementano acquisizione delimitata, ricevute persistenti di
verifica multipla e ammissione dei campi sostenuti dalla primaria. Il controllo
delimitato del percorso completo in development chiude la 26.2 il 3 ottobre 2026. Vedere
[piano](../../openspec/changes/define-toscana-alert-service/tasks.md) e
[guida della piattaforma](../fonti/piattaforme/cittadino-informato.md).

## Identità e selezione

Registrare una fonte distinta da quella comunale primaria, con prodotto
`municipal`, territorio ISTAT e canale esterno `cittadino-informato`. La
configurazione usa `access_method: cittadino-informato-api` e un contratto
`cittadino_informato` esplicito. Il registro rifiuta contratti assegnati ad altri
Comuni o canali, e revisioni che rimuovono il contratto da questo canale.

| Campo del contratto | Significato |
| --- | --- |
| `municipality_istat: "050004"` | Identità territoriale di Calcinaia associata dal registro e dal referral |
| `municipality_slug: calcinaia` | Percorso comunale e parametro API; lo slug restituito viene verificato a ogni passaggio |
| `publisher: comune_calcinaia` | Pubblicatore selezionato nell’elenco e verificato nel dettaglio |
| `updates`, `risks` | Sezioni abilitate esplicitamente, almeno una |
| `page_size` | Elementi per pagina, da 1 a 100 |
| `updates_since` | Eventuale filtro assoluto `YYYY-MM-DD` sulla data di visualizzazione |
| `attachment_paths` | Eventuali percorsi PDF consentiti, sullo stesso dominio e delimitati esplicitamente |
| `external_attachments` | Directory PDF su origini HTTPS revisionate, con referral e policy per ciascuna dipendenza |

Il referral deve indicare evidenza, destinazione, sezioni, prodotto e ISTAT dello
stesso Comune. `policy.evidence` registra la scelta dell’operatore e il suo
perimetro: questo flusso non richiede un documento di licenza. Raccolta e
conservazione sono scelte esplicite; `copies_permitted` deve essere falso.
Creare o revisionare la configurazione non attiva la raccolta programmata.

## Accesso e limiti

L’adattatore usa HTTP diretto con controllo dell’URL iniziale e dei redirect,
senza chiamate ai provider. Per Calcinaia accede alle route del namespace
`https://cittadinoinformato.it/calcinaia/wp-json/cittadino/v2/`:

- `comune?nome=calcinaia` per verificare l’identità;
- `aggiornamenti` con `comune`, `ente`, `page`, `per_page` ed eventuale `data_inizio`;
- `aggiornamenti/{id}?comune=calcinaia` per l’originale del singolo avviso;
- `rischi/oggi` e `rischi/domani`, quando selezionati.

`discovery.max_pages_per_section` limita la paginazione e
`discovery.max_documents` il lavoro per passaggio, inclusi i due documenti rischi.
`bootstrap_days` governa il piano delle revisioni sulla data di pubblicazione
effettiva: gli avvisi di data ignota e le misure già marcate ongoing/unresolved
restano nel piano. Il filtro API `updates_since` ha un significato diverso e non
certifica la copertura della finestra di pubblicazione. È assoluto: non avanza
automaticamente con il calendario; modificarlo richiede una nuova configurazione.
L’elenco contiene anche comunicazioni di pubblica utilità, senza classificazione
automatica come misura di protezione civile.

Conteggi, pagine, identità, duplicati e forma delle risposte sono verificati. Un
limite raggiunto, uno schema non riconosciuto o un cambio dei conteggi durante la
paginazione rende incompleto il passaggio, senza dichiarare assenza di avvisi.
Le risposte `429` conservano `Retry-After`; i dettagli `404/410` restano tracciati
con tentativi differiti e recupero esplicito.

## Originali, date e dipendenze

Il documento conservato usa l’URL API della risposta originale. I metadati degli
avvisi conservano separatamente `official_link`, `api_url`, pubblicatore, ID,
espressione di pubblicazione e inizio/fine della visualizzazione. Soltanto
`data_pubblicazione` valida alimenta la data di pubblicazione del piano. Le date
di visualizzazione e del contenitore WordPress non diventano validità operativa.

Elenchi, dettagli e rischi conservano JSON originale, hash e versioni. Un controllo
identico non duplica le versioni; modifiche al dettaglio o alle dipendenze creano
una versione nuova e mantengono la precedente. I rischi vengono ricontrollati a
ogni passaggio, senza cutoff sulla pubblicazione. Valori CFR mancanti restano
mancanti; non si inferiscono prodotto, zona, emissione o validità dalle etichette
oggi/domani. Tutti gli avvisi/rischi hanno `verification_state: pending`.

I PDF collegati nel corpo dell’avviso o già registrati come dipendenze si scaricano
solo nei percorsi selezionati. Il percorso condiviso `/app/uploads/` intero è
rifiutato; occorre un sotto-percorso esplicito. Il contenuto deve superare la
validazione PDF. Una dipendenza obbligatoria fuori perimetro, indisponibile o
invalida lascia un originale conservato ma incompleto, senza conservare il corpo
di errore come PDF. Nessun percorso o permesso si trasferisce ad altri Comuni.

## Verifiche del 3 ottobre 2026

Fixture sintetiche e PostgreSQL usa e getta verificano limiti, isolamento dei
Comuni e dei canali, conservazione byte per byte, ripetibilità, revisioni,
dipendenze mancanti e recuperate, `Retry-After`, dettagli scomparsi e attesa prima
del tentativo successivo. I test sono in
[acquisizione](../../internal/backend/acquisition/cittadino_informato_test.go),
[persistenza](../../internal/backend/acquisition/cittadino_informato_integration_test.go)
e [contratto](../../internal/backend/registry/cittadino_informato_test.go).

Una prova HTTP delimitata sul canale Calcinaia, con database e archivio oggetti
isolati, ha attraversato la finestra di visualizzazione selezionata dal primo
ottobre e i rischi di oggi/domani. Due passaggi sono riusciti senza duplicazioni
o revisioni spurie al secondo. Le evidenze `CIN-ACQUIRE-20261003-01` restano private;
il repository contiene fixture redistribuibili. Nessuna fonte dell’ambiente live
è stata attivata o modificata. Questa prova chiude la 26.3; confronti, accettazione,
pubblicazione e trial programmato restano aperti.


## Verifica persistente e ammissione — task 26.4

Il metodo `domain.Store.VerifyMultiSource` confronta selezioni interpretate di
originali già conservati. I comandi privati, riservati al ruolo `admin`, sono:

```sh
iwa verify-candidate /percorso/privato/candidato.json
iwa verification-receipt 123
iwa verification-history platform-source chiusura-sintetica 0
```

Usare la configurazione dell’ambiente scelto e il database aggiornato con `iwa
migrate`: la migrazione aggiuntiva `068_multi_source_verification` segue quelle
del dominio e delle query pubbliche. La lettura dello storico restituisce fino a
100 ricevute dopo l’ID richiesto. Il verificatore legge PostgreSQL e gli originali
nell’archivio oggetti; non avvia discovery, provider o una coda di approvazione.

L’input contiene `request_id`, `candidate_key`, `kind` (`local_measure`,
`operational_phase` o `regional_record`), `municipality_istat`, `candidate` e due
`checks` per gli altri ruoli. Un riferimento disponibile contiene `source_id`,
`version_id` e `fields`. Ogni campo seleziona `resource_url`, `start_byte`,
`end_byte` esclusivo e `locator`, più `json_pointer` per uno scalare JSON oppure
`ocr_run_id` e `page` per testo OCR già completo. Il valore non viene fornito:
il verificatore ricava il passaggio dai byte conservati, verifica proprietà,
hash, perimetro e limiti, quindi applica normalizzazioni esplicite. Gli offset
riguardano il testo HTML/scalare/OCR ripulito, con spazi consecutivi ridotti a uno;
script, stile e template non costituiscono evidenza. PDF e immagini richiedono
la pagina OCR completa appartenente alla stessa versione e risorsa, insieme
all’originale integro; non vengono interpretati come testo grezzo.

I controlli dichiarano `role` (`regional`, `municipal`, `platform`), `state`
(`available`, `missing`, `unavailable`, `not_applicable`), `reason` e `checked_at`.
Gli ultimi tre stati non possono fornire campi o una versione; `missing` e
`unavailable` identificano comunque la fonte. `not_applicable` è ammesso per il
CFR di una misura locale e per il rilancio municipale di un fatto regionale;
la primaria necessaria e il controllo della piattaforma non possono essere
scartati come non applicabili. Lo stato descrive il lavoro di discovery/check
che precede la verifica; questo comando non cerca autonomamente le controparti.

Per confronto locale servono riferimento dell’atto e sua edizione espliciti;
azione e oggetto, oppure fase, devono essere sostenuti dal Comune/atto. Per
rilanci regionali la fonte originaria deve appartenere al perimetro Toscana
(`09`); devono coincidere prodotto, rischio, zona, emissione e validità
esplicita assoluta. «Oggi» e «domani» non stabiliscono questa comparabilità.
Normalizzazioni di maiuscole, spazi, etichette e offset temporali espliciti non
introducono timezone, emissioni o durate. Date sole restano date; validità
ambigue restano sconosciute. Il monitoraggio conserva un livello nullo.

Le ricevute immutabili conservano tutti e tre i ruoli, fonti/versioni/configurazioni,
hash, passaggi e selettori, tempi, direzione dei confronti e risultati per campo.
Gli esiti sono `corroborated`, `missing_evidence`, `unavailable`,
`non_comparable`, `not_applicable` e `conflict`. Solo valori comparabili diversi
producono conflitto; `unknown` non è un colore corroborato. Ripetere lo stesso
`request_id` e input restituisce la ricevuta precedente; cambiarne l’input
fallisce. Per una rivalutazione usare un nuovo ID: lo storico resta consultabile.

La transazione proietta esclusivamente i valori della primaria, con la sua
attribuzione. `admitted` e `domain_record_id` descrivono il candidato verificato;
`primary_record_id` può conservare separatamente una pubblicazione primaria
supportata quando il rilancio è discordante o non comparabile. Senza i campi
primari necessari non viene creato un fatto. Campi opzionali mancanti non sono
copiati dalla piattaforma. Un avviso soltanto comunale è ammesso senza un rilancio;
Comune e piattaforma concordi non prevalgono sul colore CFR; senza la primaria
i rilanci non risultano corroborati neppure nel dettaglio dei controlli. Le scritture dirette
dello store per fonti Cittadino Informato sono rifiutate.

Ripetizioni e più candidati per la stessa primaria riusano la misura esistente,
anche se inserita da un percorso precedente: vengono aggiunte soltanto le
valutazioni temporali e di interpretazione necessarie, senza ereditare ore
inferite da un’espressione condizionale o date prive di ora.
Le revisioni conservano fatti e ricevute precedenti; le query correnti scelgono
la revisione verificata dello stesso riferimento/oggetto/luogo e fonte entro il
confine di conoscenza richiesto. Un ritorno a un hash precedente riusa il fatto e
segue l’ordine delle acquisizioni. Non si inferiscono revoche da scomparse o
relazioni fra atti diversi. La retention preserva le versioni consultate quando
una delle prove sostiene ancora una misura ongoing/unresolved; quando tutto il
grafo è scaduto, elimina anche ricevute e riferimenti senza dipendenze pendenti.

La verifica opera su selezioni interpretate: controlla riscontro letterale,
identità, comparabilità e ammissione, senza certificare da sola la correttezza
semantica dell’interpretazione. Il task 26.5 comprende la valutazione con casi
reali revisionati e l’attribuzione API/MCP estesa. Il normale worker di estrazione
non viene convertito da questo comando in un orchestratore dei tre canali:
CLN-007 resta pertinente al percorso generico; la nuova proiezione verificata ha
un ingresso programmatico distinto. Orchestrazione continuativa e attivazione programmata restano separate; la successiva 26.6 verifica un trial delimitato, descritto sotto;
la prova delimitata della 26.2 è descritta sotto. Il collector continua a
conservare il metadato iniziale `verification_state: pending`; le rivalutazioni
sono ricevute separate, non riscritture dell’originale.

I [test di verifica](../../internal/backend/domain/verification_integration_test.go)
provano persistenza, concorrenza/idempotenza, revisioni e ritorni, campo primario
mancante, storage indisponibile, sospensione dell’interpretazione, edizioni e
validità non comparabili, conflitto senza maggioranza, fasi locali, monitoraggio,
atti PDF con OCR e retention. Il [test delle query](../../internal/backend/publicquery/query_integration_test.go)
verifica la proiezione primaria e la selezione corrente/storica senza duplicati.
Sono fixture sintetiche su PostgreSQL usa e getta: nessun nuovo caso reale,
rollout, accettazione o copertura continuativa viene dichiarato dalla 26.4.

## Prova completa in development — task 26.2

Il 3 ottobre 2026 è stato eseguito il percorso acquisizione–selezione delle
evidenze–verifica–proiezione sui servizi PostgreSQL e RustFS di development,
con database e bucket dedicati. Le ricevute persistenti, gli originali integri,
le configurazioni, gli input e un dump del database restano nell’archivio
privato `CIN-DEV-20261003-01`. Nessuna fonte del trial ha raccolta programmata,
accettazione o pubblicazione abilitate; i controlli delle fonti esistenti e i
container sono stati confrontati prima e dopo e risultano preservati.

La raccolta primaria è stata eseguita prima di quella della piattaforma: una
pagina dell’elenco Notizie comunale, con 24 documenti, e un prodotto CFR di
criticità, con le risorse richieste. Due passaggi della piattaforma sulla finestra
di visualizzazione dal 1° ottobre hanno acquisito quattro avvisi e i rischi
oggi/domani, riusando le versioni invariate. Un avviso locale individuato
nell’elenco più ampio è stato acquisito anche tramite il suo esatto dettaglio
API; il PDF richiamato dalla pubblicazione primaria è stato scaricato nel
perimetro comunale, conservato e validato con Poppler. Le selezioni del confronto
locale usano il testo dell’avviso, senza dichiarare una nuova interpretazione OCR
o l’accettazione semantica dell’atto e della sua autorità emittente.

Il primo passaggio della piattaforma, dal 23 settembre, si era fermato su un PDF
comunale fuori dal contratto della piattaforma. Il fallimento e l’originale
incompleto sono conservati: la finestra ampia non è dichiarata completa e non
sono stati estesi i permessi del canale. Anche il dettaglio locale selezionato
resta incompleto per la sua dipendenza esterna; la primaria possiede il proprio
originale completo. La successiva finestra riuscita non sostituisce quel risultato.

Dieci ricevute sono state persistite, rilette e controllate per ruoli,
fonti/versioni/configurazioni, hash, passaggi e selettori, tempi, direzioni e campi:

| Caso | Esito verificato |
| --- | --- |
| Rilancio locale reale | `non_comparable` per identità/edizione non stabilite; dipendenza del canale `unavailable`, candidato non ammesso, primaria conservata separatamente |
| Pubblicazione locale primaria indipendente | `corroborated`, misura ammessa dalla primaria; piattaforma `missing_evidence` nel solo perimetro temporale riuscito, CFR `not_applicable` |
| Rilancio regionale reale | `non_comparable`, diagnostico: prodotto/zona/emissione/validità non tutti sostenuti nel rilancio e colore CFR non verificato; nessun verde dedotto dalla legenda o da `NESSUNA` |
| Sette controlli sintetici distinti | Primaria mancante e indisponibile, corroborazione locale e regionale, edizione diversa, conflitto locale e due rilanci regionali concordi contro il CFR |

Le date di pubblicazione dei due avvisi locali non sono state trasformate in
equivalenza dell’edizione dell’atto o validità operativa. Il caso primario conserva
l’espressione di validità senza inferire un’ora di cessazione. Il risultato
`missing_evidence` del controllo temporale non dichiara assenza dalla piattaforma
nel suo complesso. I conflitti della prova sono sintetici, non conflitti osservati
fra fonti reali; i due rilanci sintetici non prevalgono sul livello CFR primario.

Ripetere ciascun input ha restituito lo stesso ID; riusare l’ID con input diverso
è stato rifiutato. Gli storici conservano tutte le rivalutazioni e i due percorsi
locali reali riusano una sola misura primaria. I candidati senza primaria
necessaria non producono fatti. Tutti gli originali del trial sono stati riletti
dall’archivio oggetti con verifica d’integrità. I test di acquisizione, registro,
verificatore e CLI, le integrazioni con PostgreSQL usa e getta e i cinque gruppi
delle query pubbliche sono passati; lo smoke development ha verificato readiness
e separazione public/admin. Il percorso non ha chiamato provider.

Per ripetere una prova, selezionare esplicitamente ambiente e perimetro, conservare
configurazioni e stato iniziale e usare dati dedicati. Eseguire le acquisizioni
primarie indipendentemente dalla piattaforma; selezionare i passaggi dagli
originali conservati e usare `verify-candidate` oppure `VerifyMultiSource` sul
database scelto. Rileggere ogni ricevuta e lo storico, verificare idempotenza,
campi primari e assenza di fatti diagnostici, quindi confrontare lo stato finale.
Il runner del collaudo usa l’ingresso programmatico e resta privato insieme ai
suoi input reali; la CLI ordinaria usa il database configurato dall’applicazione.
Conservare ricevute, dump e originali prima di rimuovere i dati dedicati.

Questa prova chiude soltanto la 26.2. La valutazione estesa con esiti reali
revisionati, l’attribuzione API/MCP delle ricevute (26.5), l’orchestrazione e il
trial programmato (26.6), la diagnosi del worker generico CLN-007 e l’accettazione
restano separati. Le copie della piattaforma restano link-only.


## Valutazione e trial delimitato — task 26.5/26.6

Verifica del 3 ottobre 2026. L’operatore ha delegato all’assistente la revisione
tramite gli originali PDF. Le nuove valutazioni indicano revisore, delega, hash,
pagine e limiti; conservano le precedenti conferme umane. Le date di pubblicazione
dei due avvisi sui lavori non provano l’edizione dell’atto: le selezioni corrette
le omettono e aggiungono la validità letterale presente nella piattaforma.
Il confronto documentale trova disposizioni operative concordanti; il verificatore
mantiene `non_comparable` quando identità/edizione non sono stabilite e `unavailable`
per la dipendenza esterna mancante. Il PDF è conservato dalla fonte comunale.
Il caso sui lavori stradali verifica il meccanismo di confronto, senza diventare
una misura meteo. Le ricevute precedenti restano immutabili.

Documenti, misure, fasi e prodotti regionali possono ora riportare fino a cento
ricevute pubblicabili al confine di conoscenza richiesto. Espongono tre ruoli,
tempi, esiti, campi confrontati, direzione e passaggi visibili. Le prove di canali
privati sono oscurate insieme a identificativi, hash e valori; rimangono esito e
limite del controllo. Selezione del comune e visibilità della fonte sono obbligatorie.
Gli ID delle richieste e i selettori privati restano interni. Il limite di cento
ricevute è dichiarato quando raggiunto, senza cambiare lo storico amministrativo.
Il [contratto](../../api/public/contratto.schema.json), il
[test delle query](../../internal/backend/publicquery/query_integration_test.go)
e i [test con client MCP](../../internal/backend/transport/httpapi/public_test.go)
verificano equivalenza, conflitti e attribuzione, storico e assenza di duplicati.

Il trial 26.6 usa dati privati su un database development separato. I preview
indipendenti conservano una pagina comunale, un prodotto CFR e la finestra della
piattaforma dal 1° ottobre con rischi oggi/domani. Tre casi reali conservati sono
rivalutati separatamente, dichiarando il riuso degli originali. API e client MCP
reale restituiscono la stessa attribuzione. L’abilitazione temporanea riguarda
Calcinaia; un altro comune non eredita visibilità. La revoca rende nuovamente
inaccessibile il documento, disabilita i controlli del trial e conserva lo storico.
Nessun worker programmato è collegato al database della prova. Le copie restano
link-only; nessuna accettazione delle fonti è prodotta dal trial.
Registro privato: `CIN-TRIAL-20261003-01`.

La revisione per l’accettazione rimane distinta. La regressione conserva tutte
le tredici aspettative originali e il riuso dichiarato dell’OCR storico; una
suite aggiuntiva registra il fallimento corrente della riapertura parziale,
senza cancellarlo con i tredici confronti passati. L’interpretazione grafica CFR
e la proiezione generica CLN-007 restano aperte. Gli esiti di revisione mancanti
non vengono sostituiti da attestazioni positive. Il
[coverage tracker](../coverage.md) conserva lo stato pending per fonte.

## Raccolta e confronti continuativi — task 31.3

Il worker ordinario include un controllo periodico delle fonti piattaforma abilitate,
con interrogazione degli originali già conservati dalle primarie. Rispetta i gate di
fonte, sospensione interpretativa e territorio; PostgreSQL serializza le repliche.
Le richieste pubbliche non avviano raccolta o inferenza.

Il controllo individua controparti tramite risorsa condivisa, titolo esatto o corpo
completo. Consulta versioni, estrazioni ordinarie e OCR completo; l'identità dell'atto
richiede una clausola esplicita numero/data nella stessa evidenza della misura.
Non usa date di elenco, pubblicazione o visualizzazione come edizioni. I rischi
conservano i valori espliciti della piattaforma; emissione, zona e validità CFR
mancanti impediscono un confronto dichiarato equivalente anche se i colori coincidono.

Le ricevute automatiche usano `comparison_only`: conservano fatti selezionati,
fonti/versioni, motivi, tempi e risultati senza modificare le proiezioni primarie.
Stati consecutivi identici non duplicano ricevute; cambiamenti, indisponibilità e
recuperi conservano ogni transizione, anche quando ritorna un hash già osservato.
Le copie piattaforma restano link-only e l'accettazione rimane indipendente.

`cittadino_informato.external_attachments` dichiara le directory PDF esterne con
origine HTTPS, referral e policy separata di raccolta/conservazione. I nomi codificati
sono ammessi entro il percorso canonico; traversal, separatori codificati, altre
origini e redirect fuori confine sono rifiutati. I PDF passano il parser reale prima
della conservazione. Preview e collector applicano gli stessi controlli.

Prima di attivare una fonte, conservare stato e revisione di rollback, verificare
referral/identità, finestra di discovery e tutti gli allegati, poi eseguire un preview
completo. Registrare attivazione esplicita, intervalli effettivi e lista delle fonti
nelle configurazioni dei runner. Controllare acquisizioni ordinarie ripetute,
originali e ricevute, errori/backoff, isolamento dei comuni e API/MCP. Non riutilizzare
un trial separato come prova della raccolta ordinaria. L'adozione descritta sotto usa questi stessi controlli; evidenze private
`CIN-CONTINUOUS-20261004`.


## Adozione ordinaria verificata — 4 ottobre 2026

La sorgente `calcinaia-cittadino-informato`, revisione 1, è registrata e abilitata
nel development ordinario dopo preview completo. Il perimetro comprende 36 avvisi
nella finestra API di visualizzazione dal `2026-09-04`, una risposta elenco e i
rischi oggi/domani: 39 originali/versioni, 55 risorse obbligatorie, 16 PDF. Tutte
le risorse sono state rilette dallo storage tramite il normale verificatore di
hash e dimensione, senza dipendenze mancanti. La directory esterna revisionata
è `https://www.comune.calcinaia.pi.it/sites/default/files/`; non consente una
discovery autonoma né download da altre origini. Il filtro assoluto resta distinto
dalla finestra di pubblicazione e dalla validità delle misure.

La cadenza effettiva della raccolta è 600 secondi, con soglia di ritardo 1.800
secondi; i confronti vengono eseguiti ogni minuto senza provider. Due controlli ordinari
completi hanno acquisito ciascuno 38 documenti e un elenco; il secondo è partito
600,22 secondi dopo la fine del primo, senza nuove versioni spurie. Le prime 50
ricevute ordinarie sono diagnostiche: 24 `missing_evidence` e 26 `non_comparable`,
inclusi i rilanci regionali senza metadati CFR espliciti. Non sono cinquanta
misure corroborate: conservano la ricerca delle controparti, i motivi e gli
originali/versioni consultati. Le estrazioni ordinarie successive possono
produrre nuove ricevute; nessun risultato di classificazione con citazioni
invalide viene riparato a mano o considerato valido. Gli errori di citazione
osservati rimangono un limite dell'interpretazione, distinto dalla raccolta.
Al termine dei due passaggi, 33 dei 38 job di classificazione sono falliti per
`classification_output_quotation` e cinque sono riusciti; tutti i 16 job OCR
sono riusciti. Questi conteggi non certificano misure o avvisi meteo corroborati.

I test sintetici PostgreSQL/race verificano idempotenza, nuova versione primaria,
conflitto confrontabile, indisponibilità, recupero e gate disabilitati; il
recupero conserva anche il ritorno a uno stato precedente. Gli originali primari
archiviati restano consultabili come evidenza; i candidati piattaforma archiviati
sono esclusi dall'esecuzione. Queste prove di errore sono sintetiche, senza
manomettere le fonti reali. Le prove di acquisizione verificano retry/backoff e
recupero delle dipendenze; le query pubbliche rispettano i confini storici alla
precisione di PostgreSQL.

Sei confronti effettivi con client SDK verificano API/MCP della situazione,
isolamento municipale, vista corrente salvata, fatti primari al precedente
confine di conoscenza, ricevute del documento e copertura della sorgente. La vista di situazione mostra il canale `available`,
mentre l'altro Comune verificato resta `not_collected`. Aggiungere/attivare una
sorgente cambia la revisione del perimetro development: le viste precedenti
rispondono `410` secondo il contratto, senza riscrivere i fatti storici. Le copie
restano link-only: il download dell’originale restituisce `403` e `copy_url`
è nullo. Accettazione, abilitazione pubblica ordinaria e produzione
sono indipendenti.

La revisione applicativa pulita `8545385` è adottata su public/admin, sei worker
e backup applicativo. Readiness, controllo delle immagini e smoke non distruttivo
passano; database, storage, crawler e frontend esistenti sono preservati. Le
configurazioni delle primarie municipale/CFR mantengono i rispettivi perimetri e
intervalli; non dipendono dalla raccolta piattaforma. Evidenze operative,
configurazione precedente e rollback restano privati in `CIN-CONTINUOUS-20261004`.
La verifica del backup/ripristino reale rimane rinviata su richiesta dell'operatore.
