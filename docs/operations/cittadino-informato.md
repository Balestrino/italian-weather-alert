# Acquisizione delimitata di Cittadino Informato

I task 26.3 e 26.4 implementano acquisizione delimitata, ricevute persistenti di
verifica multipla e ammissione dei campi sostenuti dalla primaria. La chiusura
della 26.2 richiede il successivo controllo in development del percorso completo. Vedere
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
un ingresso programmatico distinto. Collegamento operativo, prova completa della
26.2 e trial programmato 26.6 restano da eseguire. Il collector continua a
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
