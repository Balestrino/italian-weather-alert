# Stato dei punti Toscana — 4 ottobre 2026

La [checklist](../../openspec/changes/define-toscana-alert-service/tasks.md)
ha 159 task completati su 166. In questo giro sono verificati:

| Task | Risultato |
| --- | --- |
| 9.4 | [Consuntivi separati](trial-costs.md), metriche indisponibili e prezzi storici non verificati espliciti; budget pendente |
| 11.5 | [Campagne indipendenti e quattro report pending](municipal-acceptance.md); stessi gate del pilota |
| 13.8 | [Recupero provider verificato](acquisition-recovery.md#verifica-del-recupero-provider--4-ottobre-2026), OCR e pipeline ordinaria riusciti; fallimento di citazione preservato |
| 9.3 e 30.1/30.2 | [Campagna MVP complete con audit correttivo](observational-trial.md#esito-verificato-del-4-ottobre-2026); durata, quattro ambiti, confronti e prove attribuiti |
| 29.1–29.3 | [Recupero selettivo con audit](municipal-interpretation.md#esito-del-recupero-live), test e worker ordinari verificati; omissione live e campagne extended nell’esito iniziale, superato dalla chiusura 9.3 |

I sette gate originari seguenti restano aperti per gli esiti specifici mancanti:

| Task | Condizione ancora da verificare |
| --- | --- |
| 1.8 | Dimensionamento previsto della VM, SMTP/mailbox, PBS e telemetria continuativa delle soglie |
| 8.2 | Backup PBS whole-VM off-host, piano adottato e allarme effettivamente consegnato |
| 8.3 | Ripristino PBS in target isolato con dati/originali, configurazioni e recupero job |
| 10.1 | Collaudo reale dei quattro ambiti MVP, sette rischi/API-MCP/scansioni/storico/errori; accettazione e abilitazione separate |
| 10.2 | Readiness operativa della produzione, capacità concorrente, proxy/SMTP/provider, recupero host e rollback |
| 12.4 | Immagine approvata su revisione pulita, staging verificato, accesso GHCR operativo, pubblicazione e pull del digest |
| 12.5 | Piano PBS ora documentato; backup, allarmi e ripristino cronometrato reali prima di attestare RPO/RTO |

Il [runbook PBS](pbs-recovery.md) rende concreti obiettivi, cadenza, retention
proposta, controllo età e prove di recupero. I controlli di configurazione degli
ambienti passano, mentre configurazione e precedenti consegne all'operatore non
sono risultati di collaudo dell'infrastruttura. Le evidenze operative e i dettagli
di accesso restano privati, riferimento `REMAINING-TOSCANA-20261004`.

I test del software non autorizzano l'avvio della produzione. Il
[coverage tracker](../coverage.md#region-09) conserva accettazione pendente e
copertura municipale incompleta, indipendentemente dal conteggio dei task.


## Adozione in development

La successiva [correzione della riapertura parziale](municipal-interpretation.md)
supera il fallimento del merge nel replay conservato e nelle fixture sintetiche
con PostgreSQL, con catalogo di estrazione v25. Il seguito descritto sotto verifica
anche adozione e rielaborazione ordinaria selettiva, con un limite di completezza
distinto dal problema del merge.
La lettura aggiornata delle due campagne conserva `extended`. I task 9.3 e 10.1
rimangono aperti anche dopo questo avanzamento.

La revisione applicativa `c69a10e` è stata costruita dal checkout pulito e adottata
in development dopo i controlli del software e le migrazioni. Public, admin, le
sei repliche del worker e il backup già attivo condividevano l'immagine;
database, storage e crawler conservano i contenitori e i volumi esistenti.
Il backup applicativo resta attivo finché la protezione PBS non è verificata.

Passano la verifica delle immagini/configurazioni runtime, readiness HTTP,
isolamento amministrativo e confronto effettivo API/MCP della situazione con
client SDK sulla stessa vista salvata. Le due campagne storiche restano MVP e
l'API privata rifiuta i quattro avvii municipali con prerequisiti irrisolti,
senza nuove campagne o cambiamenti dei controlli delle fonti. Il registro
operativo conserva immagine/configurazione precedenti e materiali per il rollback.
L'adozione non conclude i task aperti sopra e la produzione resta ferma.

## Seguito: regressione e recupero selettivo

La correzione v25 è stata adottata in development con sei worker, dipendenze e
controlli delle fonti conservati. API/MCP sulla stessa vista, runtime e smoke
passano. Il confronto di completezza passa 14/14 con lo stesso contratto,
conservando il fallimento precedente; si tratta di replay delle risposte
conservate, distinto dalle nuove esecuzioni ordinarie.

I task 29.1–29.3 implementano e verificano il recupero tracciato delle versioni
archiviate scelto dall’operatore. La revisione pulita `74acab8` e la migrazione
additiva sono adottate su public/admin, sei worker e backup applicativo.
Classificazione, estrazione e linking dei due casi scelti riescono; retry, audit,
archivi, risultati precedenti, viste salvate e confini di conoscenza sono verificati.
API/MCP coincidono anche sulle pagine successive e i dati restano nel Comune.

Il nuovo output conserva riapertura ed eccezione, ma un solo ambito di divieto
invece dei quattro revisionati. La nuova regressione fallita, con contratto
invariato e verifiche riusate dichiarate, conserva e supera il successo del replay;
ulteriori recuperi richiedono una rivalutazione riuscita. Le due campagne rivalutate
restano `extended`. Controlli delle fonti e dipendenze sono conservati, produzione
ferma. Nessuno degli otto gate originari viene chiuso da questo passaggio.

## Correzioni preparate per la 9.3

I task 30.1/30.2 verificano estrazione v26 per i luoghi indipendenti dei divieti,
sostituzione della sola interpretazione proiettata corrente della stessa versione
e revisioni correttive con legami immutabili. Le valutazioni storiche conservano
il precedente esito; durata, ritardi, allegati e controlli delle fonti restano gate.
I confronti delegati sui due originali PDF CFR conservati verificano 364 campi
per prodotto, maschere di applicabilità e quantità senza trasformarle in allerte.
Un simbolo curvo sopra le mappe è escluso solo con hull interamente esterno;
le forme non supportate nei pannelli continuano a bloccare la lettura.
La chiusura 9.3 richiede ancora l’esito ordinario e la valutazione persistita.

Il primo recupero ordinario v26 conserva i quattro divieti, ma aggiunge una
chiusura ricavata dall’inciso sullo stato precedente alla riapertura. Il nuovo
fallimento dello stesso contratto rimane nel registro. V27 verifica l’esclusione
del solo inciso retrospettivo, conservando clausole di chiusura indipendenti;
la prova ordinaria resta necessaria prima della chiusura 9.3 (CLN-024).

## Chiusura della 9.3

La [campagna MVP](observational-trial.md#esito-verificato-del-4-ottobre-2026)
ha una valutazione persistita `complete`: 239,94 ore, nove intervalli completi di
24 ore e undici date locali osservate su tutti e quattro gli ambiti. Confronti
originali/allegati, prove di errore e caso di evento assente sono attribuiti;
i ritardi sono entro i limiti originari e nessuna risorsa richiesta manca.
Le revisioni correttive e lo storico sono immutabili. La campagna interna precedente
resta `extended` per i ritardi storici; questi non sono stati condonati.

La revisione pulita `ce941ed` (estrazione v27) è adottata in development su
public/admin, sei worker e backup applicativo. Il nuovo recupero ordinario conserva
una riapertura, due chiusure mantenute, quattro divieti distinti e COC attivo,
escludendo l’inciso di chiusura precedente. La regressione passa 14/14 con contratto
immutato e riuso dichiarato; i fallimenti v25/v26 restano registrati. API/MCP
coincidono anche con paginazione, e viste/limiti storici conservano i vecchi fatti.

Runtime, smoke, test unit/race e PostgreSQL passano. Archivi, job e risultati
precedenti, controlli delle fonti e dipendenze sono conservati; produzione ferma.
La 9.3 è chiusa, 10.1 e gli altri sei gate restano aperti. I paragrafi precedenti
conservano gli esiti intermedi datati, superati da questa valutazione.
