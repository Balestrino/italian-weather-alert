# Acquisizione delimitata di Cittadino Informato

Il task 26.3 implementa l’acquisizione degli aggiornamenti e dei rischi del Comune
selezionato. Confronti persistenti e ammissione restano nel task 26.4; la chiusura
della 26.2 richiede il successivo controllo del percorso completo. Vedere
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
