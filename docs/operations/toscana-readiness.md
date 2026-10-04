# Stato dei punti Toscana — 4 ottobre 2026

La [checklist](../../openspec/changes/define-toscana-alert-service/tasks.md)
ha 153 task completati su 161. In questo giro sono verificati:

| Task | Risultato |
| --- | --- |
| 9.4 | [Consuntivi separati](trial-costs.md), metriche indisponibili e prezzi storici non verificati espliciti; budget pendente |
| 11.5 | [Campagne indipendenti e quattro report pending](municipal-acceptance.md); stessi gate del pilota |
| 13.8 | [Recupero provider verificato](acquisition-recovery.md#verifica-del-recupero-provider--4-ottobre-2026), OCR e pipeline ordinaria riusciti; fallimento di citazione preservato |

Gli otto task seguenti restano aperti per gli esiti specifici mancanti:

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
con PostgreSQL. È verificata nel checkout, con catalogo di estrazione v25;
adozione e rielaborazione ordinaria live di questa correzione restano da eseguire.
La lettura aggiornata delle due campagne conserva `extended`. I task 9.3 e 10.1
rimangono aperti anche dopo questo avanzamento.

La revisione applicativa `c69a10e` è stata costruita dal checkout pulito e adottata
in development dopo i controlli del software e le migrazioni. Public, admin, le
sei repliche del worker e il backup già attivo ora condividono l'immagine;
database, storage e crawler conservano i contenitori e i volumi esistenti.
Il backup applicativo resta attivo finché la protezione PBS non è verificata.

Passano la verifica delle immagini/configurazioni runtime, readiness HTTP,
isolamento amministrativo e confronto effettivo API/MCP della situazione con
client SDK sulla stessa vista salvata. Le due campagne storiche restano MVP e
l'API privata rifiuta i quattro avvii municipali con prerequisiti irrisolti,
senza nuove campagne o cambiamenti dei controlli delle fonti. Il registro
operativo conserva immagine/configurazione precedenti e materiali per il rollback.
L'adozione non conclude i task aperti sopra e la produzione resta ferma.
