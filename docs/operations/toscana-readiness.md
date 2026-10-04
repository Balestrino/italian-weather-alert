# Stato dei punti Toscana — 4 ottobre 2026

La [checklist](../../openspec/changes/define-toscana-alert-service/tasks.md)
ha 156 task completati su 164. In questo giro sono verificati:

| Task | Risultato |
| --- | --- |
| 9.4 | [Consuntivi separati](trial-costs.md), metriche indisponibili e prezzi storici non verificati espliciti; budget pendente |
| 11.5 | [Campagne indipendenti e quattro report pending](municipal-acceptance.md); stessi gate del pilota |
| 13.8 | [Recupero provider verificato](acquisition-recovery.md#verifica-del-recupero-provider--4-ottobre-2026), OCR e pipeline ordinaria riusciti; fallimento di citazione preservato |
| 29.1–29.3 | [Recupero selettivo con audit](municipal-interpretation.md#esito-del-recupero-live), test e worker ordinari verificati; omissione live registrata e campagne ancora extended |

Gli otto gate originari seguenti restano aperti per gli esiti specifici mancanti:

| Task | Condizione ancora da verificare |
| --- | --- |
| 1.8 | Dimensionamento previsto della VM, SMTP/mailbox, PBS e telemetria continuativa delle soglie |
| 8.2 | Backup PBS whole-VM off-host, piano adottato e allarme effettivamente consegnato |
| 8.3 | Ripristino PBS in target isolato con dati/originali, configurazioni e recupero job |
| 9.3 | Campagne esistenti ancora extended: confronti/prove mancanti o falliti e ritardi; tempo trascorso sufficiente non equivale a completamento |
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
