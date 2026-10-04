# Campagna osservativa Toscana e Calcinaia

La 9.3 richiede almeno sette intervalli completi di 24 ore di controlli cadenzati
su vigilanza, criticità/allerta, monitoraggio e Calcinaia. Ogni fonte conserva
configurazione, sezioni, controlli giornalieri, ritardi, confronti degli originali
e allegati, autore ed evidenze delle prove di errore. Il tempo trascorso da solo
non conclude la campagna. Consultare il [coverage tracker](../coverage.md#region-09)
per l’accettazione e lo [stato dei task](toscana-readiness.md) per gli altri gate.

## Revisioni correttive

Il report privato di `/admin/observation-campaigns/{id}/report` distingue gli esiti
correnti dalle revisioni precedenti. Una correzione si registra con
`POST /admin/observation-campaigns/{id}/reviews`, indicando `actor` e `review`:
`id`, `source_id`, `kind`, `status`, riferimento all’originale (`version_id`),
controllo (`check_id`) o caso (`case_id`), `evidence`, `supersedes` e `correction`.
L’evidenza identifica URL, riscontro e data di osservazione; il motivo descrive
la correzione verificata. La revisione assistita deve essere attribuita alla
precisa delega dell’operatore.

`supersedes` deve riferirsi a una revisione fallita o irrisolta della stessa
campagna, fonte, tipo e identico originale/controllo/caso. Serve evidenza più
recente. Link, autore, motivo e storico sono immutabili; sono ammesse catene di
correzioni, mentre biforcazioni, ambiti diversi e correzioni dopo la chiusura
sono rifiutati atomicamente. `superseded_by` viene calcolato dal servizio.

Il report conta l’ultimo esito della catena e conserva tutte le revisioni. Un
report con `through` precedente alla nuova evidenza conserva il fallimento;
le valutazioni già registrate conservano il loro contenuto originario. Le
correzioni non cambiano la durata minima, i ritardi ammessi, gli allegati richiesti
o i controlli delle fonti. Registrare la valutazione con
`POST /admin/observation-campaigns/{id}/assess`: il servizio calcola lo stato,
serializzando valutazione e inserimento delle revisioni. Non accetta uno stato
`complete` dichiarato dal chiamante.

## Perimetro delle evidenze

Un confronto di completezza deve conservare il contratto revisionato. Il replay
delle risposte conservate e un risultato fresco del worker sono evidenze
distinte; ogni controllo riusato conserva la sua provenienza. Un nuovo fallimento
resta nel registro anche dopo la successiva correzione.

Per eventi o guasti assenti durante la finestra usare casi conservati, indicando
esplicitamente osservazioni reali, fixture sintetiche e simulazioni. I test del
software non sono guasti osservati sulla fonte. Una situazione senza evento di
monitoraggio non dimostra il comportamento del prodotto durante un evento.

Le evidenze operative, originali, hash completi, ricevute, job e configurazioni
sono conservati nell’archivio privato `TOSCANA-TRIAL-CLOSURE-20261004`. Le fixture
pubbliche restano sintetiche. La campagna interna precedente conserva i suoi
ritardi e le valutazioni `extended`; nessun margine viene modificato retroattivamente.

## Esito verificato del 4 ottobre 2026

La campagna MVP iniziata il 24 settembre ha una valutazione persistita `complete`
al 4 ottobre, dopo 239,94 ore: nove intervalli completi di 24 ore, oltre il minimo
richiesto di sette. Ogni fonte ha undici date locali di osservazione. Il limite
originario dei ritardi è 10.800 secondi; nessuna risorsa richiesta risulta mancante.

| Ambito | Controlli | Completi | Falliti/incompleti | Intervallo massimo, secondi |
| --- | ---: | ---: | ---: | ---: |
| Calcinaia | 242 | 184 | 58 | 4.801 |
| Criticità/allerta | 238 | 238 | 0 | 3.697 |
| Monitoraggio | 238 | 238 | 0 | 3.697 |
| Vigilanza | 238 | 238 | 0 | 3.697 |

Il report conserva le tre sezioni comunali Notizie/Avvisi/Comunicati e gli endpoint
specifici dei tre prodotti CFR, con configurazioni e controlli giornalieri.
Le revisioni correttive su delega confrontano gli originali e gli allegati:
364 campi per prodotto grafico, maschere di applicabilità, bande giornaliere/totali
0–10 e simboli operativi distinti dalle legende. Il monitoraggio conserva il
confronto della pagina senza evento, l’allegato non applicabile e un caso di evento
conservato. Le prove CFR di errore sono fixture sintetiche conservate, per assenza
di guasti osservati nella finestra; Calcinaia conserva il guasto reale revisionato.

La nuova esecuzione ordinaria comunale v27 produce una riapertura, due chiusure
mantenute, quattro divieti con luoghi letterali distinti e COC attivo. La regressione
passa 14/14 sul contratto immutato, con le altre verifiche riusate dichiarate;
non sono quattordici nuove chiamate al modello. Il fallimento v26 sull’inciso
retrospettivo e quello v25 sui divieti rimangono nel registro.

API/MCP coincidono sulla stessa vista, anche con paginazione su quattordici pagine
verificate; la situazione corrente contiene solo l’ultima interpretazione ordinaria
proiettata della versione corretta. Le vecchie interpretazioni restano nelle viste
salvate e nei confini di conoscenza precedenti. Archivi, job, risultati, revisioni
e valutazioni precedenti sono verificati immutati; l’altro Comune resta isolato.

La 9.3 è completata. Il collaudo 10.1, l’accettazione delle fonti e la readiness di
produzione restano separati. L’adozione è limitata a development, con sei worker,
dipendenze e controlli delle fonti conservati; produzione ferma.
