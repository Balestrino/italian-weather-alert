# Runbook PBS: protezione della VM e ripristino

Procedura predisposta il 4 ottobre 2026 per i task 1.8, 8.2, 8.3, 10.2 e 12.5.
Configurazione, backup, consegna degli allarmi e ripristino reale sono ancora
**da verificare**. Gli obiettivi non sono dichiarati raggiunti.

## Piano iniziale

| Voce | Obiettivo o impostazione proposta |
| --- | --- |
| Protezione | Intera VM, datastore PBS su host fisicamente separato |
| Dati inclusi | Tutti i volumi PostgreSQL/RustFS dei tre ambienti, checkout/configurazioni, secret e materiali necessari al recupero |
| RPO | Perdita massima di dati: 6 ore |
| RTO | Dalla rilevazione del guasto a servizio API/MCP verificato: 1 ora |
| Cadenza | Ogni 3 ore; orari iniziali 00:00, 03:00, 06:00, 09:00, 12:00, 15:00, 18:00, 21:00 UTC |
| Retention proposta | `keep-last=8`, `keep-daily=7`, `keep-weekly=4`, `keep-monthly=3` |
| Controllo età | Watchdog fuori dalla VM ogni 5 minuti; warning a 4 ore, critico a 5 ore, violazione a 6 ore |
| Capacità | Telemetria continuativa; segnalare disco/RAM stabilmente oltre il 70% |

Registrare privatamente nodo/VMID, datastore/namespace, dischi inclusi o esclusi,
versioni PVE/PBS, fuso del job, modalità di consistenza, credenziali di recupero,
chiavi di cifratura e target isolato. Verificare che ogni volume e configurazione
necessaria sia effettivamente nel backup e che l'accesso di recupero non dipenda
dalla VM guasta. I materiali privati devono essere recuperabili da un secondo
host prima di dichiarare operativa la protezione.

La retention è una proposta da dimensionare e registrare prima dell'adozione;
le regole possono selezionare gli stessi snapshot. Pianificare prune, garbage
collection e verifica secondo la [documentazione PBS](https://pbs.proxmox.com/docs/maintenance.html),
misurando spazio e durata. Non avviare prune sui backup esistenti durante una
semplice verifica del runbook.

## Backup, età e allarmi

1. Configurare il job whole-VM in PVE verso PBS, includendo tutti i dischi
   necessari e la consistenza PostgreSQL/RustFS. Conservare configurazione,
   log, identificativo snapshot, punto di recupero e risultato della verifica.
2. Eseguire un backup reale. Il recupero si misura dal punto dati dello snapshot,
   senza azzerare l'età al termine del trasferimento. Un job incompleto o uno
   snapshot non recuperabile non aggiorna l'ultimo punto valido.
3. Instradare errori PVE di backup e gli eventi PBS pertinenti verso il percorso
   operativo SMTP esistente. I [target e matcher PBS](https://pbs.proxmox.com/docs/notifications.html)
   gestiscono le notifiche PBS; il watchdog esterno controlla anche l'assenza di
   esecuzioni e l'età. Una VM o un PBS spento non deve silenziare il watchdog.
4. Verificare test di consegna e un errore controllato con un job dedicato,
   conservando ricezione, apertura dell'incidente, promemoria e recupero.
   La configurazione SMTP senza ricezione effettiva non completa la prova.
5. Provare il watchdog con snapshot di prova vecchio, assente e non valido,
   mantenendo i backup reali. Verificare warning, escalation prima delle 6 ore,
   recupero e allarme sul mancato funzionamento del watchdog stesso.
6. Su job fallito avviare diagnosi e retry entro il margine disponibile. Una
   cadenza di 3 ore non garantisce da sola RPO 6 ore: un intero giro saltato,
   durata del backup e tempo di risposta possono superarlo. Ridurre la cadenza
   quando misure e margini non bastano.

Lasciare il worker di backup applicativo disabilitato quando PBS whole-VM è
verificato e attivo, salvo un'esigenza distinta autorizzata. Un backup applicativo
esistente non certifica PBS e non va rimosso come conseguenza della sola
documentazione del nuovo piano.

## Ripristino isolato e tempi

1. Registrare `t0` alla rilevazione del guasto simulato e il punto dati scelto.
   Preparare una VM di recupero isolata e impedire, **prima del primo avvio**,
   uscita verso fonti, provider e SMTP. Evitare conflitti di IP, DNS e volumi
   con l'installazione esistente. La produzione resta ferma.
2. Ripristinare lo snapshot PBS nel target isolato. Annotare tempi di provisioning,
   download, ripristino e primo avvio, errori e interventi manuali.
3. Verificare secret/configurazioni, PostgreSQL e RustFS; risolvere dal database
   gli oggetti originali e gli allegati, confrontando hash, versioni, riferimenti,
   interpretazioni e storico. Comprendere dati dei tre ambienti.
4. Verificare recupero/idempotenza dei job incompleti con un campione controllato
   che non generi richieste esterne. Le ultime acquisizioni e i controlli devono
   conservare le date pre-backup; l'avvio non li rende freschi.
5. Verificare readiness, isolamento dell'amministrazione e richieste API/MCP
   equivalenti, con riferimenti agli originali e stato di aggiornamento veritiero.
   Annotare `t1` solo quando tutte queste prove sono riuscite.
6. Calcolare RTO `t1-t0` e RPO rispetto al punto dati recuperato, rendicontando
   separatamente la durata del job di backup. Se uno supera l'obiettivo,
   conservare l'esito fallito e correggere/ripetere la procedura.

Conservare privatamente cronologia firmata dall'operatore, log, snapshot,
controlli oggetti, esiti API/MCP, prova allarmi, capacità e decisione finale.
Ripetere la prova dopo modifiche rilevanti alla topologia o alla consistenza.
Il rollback applicativo usa il precedente digest compatibile; il recupero di
dati richiede uno snapshot coordinato PostgreSQL/RustFS e la valutazione della
perdita successiva al backup. Vedere [rilasci](releases.md).

## Esiti necessari per chiudere i task

| Task | Evidenza reale richiesta |
| --- | --- |
| 1.8 | Risorse VM previste, SMTP ricevuto, piano PBS adottato, telemetria e soglie verificate |
| 8.2 | Backup whole-VM off-host riuscito, inclusioni/retention e allarme consegnato |
| 8.3 | Ripristino isolato, originali/versioni/configurazioni e recupero job verificati |
| 10.2 | Capacità concorrente, HTTPS/proxy e amministrazione privata, provider/SMTP, recupero da guasto host e rollback |
| 12.5 | Piano registrato e prove dipendenti effettive prima di attestare RPO/RTO |

La consegna precedente dei controlli all'operatore è conservata nello
[stato della verifica](environment-verification.md); non sostituisce questi esiti.
