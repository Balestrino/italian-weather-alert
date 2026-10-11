# Notifiche di elaborazione e lavoro arretrato

Gli incidenti `processing_failed` conservano gli errori irrisolti nel backoffice.
Dopo la prima email, una nuova notifica di fase richiede un fallimento definitivo
o un rilancio esplicito successivo allo snapshot già consegnato e il decorso
dell’intervallo configurato. Un vecchio errore immutato non genera altri promemoria
di fase. I successi di job diversi non attestano il recupero dei job falliti.

Il rilevatore crea un incidente `processing_backlog` per coda quando almeno un
job irrisolto non archiviato è in attesa/esecuzione da 24 ore, oppure è fallito
definitivamente da 24 ore. Il riepilogo conta tutti i job falliti e pendenti della
coda, con il sottoinsieme pendente che riporta `provider_scope_held`. Questo è
un conteggio di lavoro, non di documenti mancanti: più job possono riferirsi allo
stesso documento. I promemoria del riepilogo hanno un intervallo minimo di 24 ore;
un intervallo configurato più lungo resta efficace e zero disabilita i promemoria.
I ritardi delle fonti e gli esiti dei backup mantengono i criteri precedenti.

Ogni email conserva i conteggi del proprio snapshot. I conteggi degli incidenti
aperti vengono aggiornati separatamente e sono visibili in `/admin/notifications`.
I messaggi anteriori alla migrazione non dispongono dello snapshot: il campo
`available` è falso e non se ne inventano conteggi. Un invio SMTP ritardato non
consuma la possibilità di notificare fallimenti avvenuti dopo lo snapshot.

La chiusura di un riepilogo attesta soltanto l’assenza di lavoro arretrato di
almeno 24 ore. Può dipendere da completamenti o dall’archiviazione revisionata;
non certifica disponibilità del provider, completezza interpretativa o accettazione
di una fonte. Il monitor non consuma tentativi, rilancia job o rimuove evidenze.
I job falliti definitivamente non fanno parte della coda eseguibile; i job
differiti mantengono il budget e attendono il recupero della dipendenza.

## Verifiche e adozione

La verifica dell’11 ottobre 2026 usa PostgreSQL disposable e un relay SMTP locale
con fixture sintetiche: raggruppamento, errori immutati, nuovi fallimenti, rilanci,
invii ritardati, restart, snapshot durevoli, soglia di 24 ore, intervalli disabilitati
o più lunghi, archiviazione e mancato recupero da successi estranei. I test
dell’amministrazione e dei template Compose verificano il reader e la selezione
della sola immagine admin in development; staging/production mantengono
l’immagine di release condivisa.

Per l’adozione development conservare backup del database, immagini/configurazione
precedenti e identificativi dei container nell’archivio privato. Prima di ritirare
un job ridondante, verificare la sua esatta versione e un risultato successivo
valido per lo stesso stadio/risorsa. Un risultato di valutazione, un semplice
`state=succeeded`, un originale più nuovo o un’equivalenza senza materializzazione
validata non bastano. Per la selezione delimitata di classificazioni ordinarie
considerare l’ultimo run non di valutazione e richiedere `classified`, contenuto
completo e timestamp successivo al fallimento. Registrare job, run sostitutivo,
attore e data privatamente; conservare originali, tentativi e risultati.

Usare `IWA_WORKER_IMAGE` e `IWA_ADMIN_IMAGE` in `.local/development.env`. Prima del riavvio applicare la migrazione con
`scripts/compose-env.sh development run --rm --no-deps --pull never admin migrate`, quindi
`scripts/compose-env.sh development up -d --no-deps --no-build --pull never --wait admin worker`
per sostituire soltanto quei servizi. Verificare migrazioni additive, snapshot
incidenti, outbox, immagini reali, readiness e identità degli altri container.
Il backup privato e il rollback software non costituiscono una prova di restore
infrastrutturale o di readiness production.

L’adozione development dell’11 ottobre è verificata dopo l’applicazione esplicita
della migrazione. Solo worker/admin sono sostituiti; gli altri container conservano
identità, immagini e tempi di avvio. Una selezione delimitata di fallimenti di
classificazione coperti dall’ultimo risultato valido della stessa versione viene
archiviata con attore e data, senza eliminare tentativi. La lettura del primo
riepilogo verifica i conteggi persistiti e l’accettazione SMTP; gli errori storici
immutati non producono nuove email di fase. Il lavoro ancora scoperto e i blocchi
dei provider restano espliciti, senza rilanci automatici o certificazioni di
accettazione. Evidenze e rollback sono conservati privatamente.
