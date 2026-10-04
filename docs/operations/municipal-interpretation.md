# Interpretazione della riapertura parziale di Calcinaia

Verifica del 4 ottobre 2026, avanzamento dei task 9.3 e 10.1 della
[checklist Toscana](../../openspec/changes/define-toscana-alert-service/tasks.md).
La correzione del software supera il fallimento di merge del caso conservato;
la successiva [chiusura della campagna 9.3](observational-trial.md) verifica il
percorso ordinario corretto; accettazione e 10.1 restano separati.

## Causa e correzione

Il modello restituisce un apostrofo diritto nel luogo di un divieto; il testo
conservato usa quello tipografico. La tabella delle citazioni restituisce già il
testo originale attraverso un confronto contiguo canonico, ma la validazione
dell'ambito continua a usare il luogo del modello. Il luogo diventa sconosciuto
e confligge nel merge con gli altri divieti aventi lo stesso soggetto generale.

La validazione ora conserva la grafia originale del luogo prima del controllo
dell'ambito. Il controllo del merge per conflitti fra luogo noto e sconosciuto resta
attivo: una citazione inesistente non viene corretta per somiglianza.

Un completamento locale limitato riconosce una frase autonoma che dichiara una
riapertura già avvenuta alla circolazione, con soggetto esplicito e frase di
competenza del core. Un aggiornamento generico dello stesso soggetto sostenuto
da quella frase diventa una riapertura, con la citazione operativa precisa.
In assenza della misura viene aggiunta soltanto la riapertura sostenuta.
Negazioni, ipotesi, futuro, citazioni riportate, intestazioni di ordinanza e
frasi appartenenti al solo contesto non producono questa integrazione.
Altri formati continuano a richiedere l'interpretazione ordinaria.

L'inciso sulla precedente chiusura e l'aggiornamento della pagina non datano
la riapertura. Nel formato riconosciuto i tempi rimangono indeterminati.
Le altre restrizioni e l'attivazione del COC mantengono ambiti ed evidenze propri.

## Verifiche e limiti

Il replay delle risposte conservate produce una riapertura, due chiusure,
quattro divieti distinti e un'attivazione. Mantiene il conflitto temporale
agosto/settembre dell'altro caso e l'eccezione del sottopasso nel caso delle
strade liberate. Nessuna nuova chiamata ai modelli; originali e risposte del
precedente replay rimangono invariati. Evidenza privata: `CLN-PARTIAL-REOPEN-20261004`.

Fixture sintetiche verificano casi positivi/negativi, grafia dei luoghi, tempi
indeterminati, deduplicazione e compatibilità legacy. PostgreSQL isolato verifica
il percorso del runner, la rilettura di tutte le misure/citazioni e la conservazione
dell'estrazione precedente. Passano test race, integrazione dei pacchetti
interessati, vet e build.

La configurazione di estrazione v25, logica v20, distingue la nuova semantica
per entrambi i provider; prompt, selezione delle fonti e cataloghi precedenti
sono conservati. Il cambio di catalogo non rielabora automaticamente lo storico.

Le due campagne lette in development restano `extended`; i report persistenti
precedenti non sono stati sostituiti dal replay locale. Il seguito descritto sotto
verifica adozione, stesso contratto e recupero esplicito dei casi selezionati con
confronto persistente, storico e API/MCP. Restano da completare confronti, prove di
errore e ritardi ancora mancanti prima di chiudere 9.3/10.1. La revisione del
software non conclude gli altri [gate operativi](toscana-readiness.md).

## Recupero selettivo delle versioni archiviate

Il seguito del 4 ottobre registra una rivalutazione riuscita dello stesso
contratto di completezza: quattordici confronti, aspettative e corpus invariati,
precedente fallimento conservato. Discovery e parser vengono rieseguiti; OCR,
classificazione e linking usano risposte conservate, senza nuove chiamate.
L’adozione di v25 è verificata in development. Il caso non archiviato del
conflitto di date completa classificazione ed estrazione sul worker.

I casi storici di riapertura ed eccezione sono invece archiviati. Su scelta
esplicita dell’operatore i task 29.1/29.2 aggiungono un recupero per versioni esatte.
La migrazione additiva `096_selected_archive_recovery` conserva il marker di
archiviazione e aggiunge audit immutabile, fonte/revisione, attore, data,
evidenza, regressione e selezioni. Richiede la revisione attiva, acquisizioni
conservate complete e le ultime regressioni riuscite. Una richiesta non modifica
risultati o tentativi precedenti e non abilita raccolta o pubblicazione.

La sola nuova classificazione della selezione e le sue estrazioni/link derivati
possono attraversare il guard di archiviazione. Il percorso ordinario, le altre
richieste e gli altri documenti restano esclusi. Restano efficaci i controlli
territoriali, le sospensioni e i gate dei provider. Il retry della stessa richiesta
riprende enqueue mancanti senza duplicare i job; cambiare la selezione sotto lo
stesso ID è un conflitto. Un errore dopo il salvataggio dell’audit si recupera
ripetendo l’identico corpo, con identiche evidenze.

Usare esclusivamente l’API amministrativa privata, con ID di versioni e
regressione letti dall’ambiente interessato. Esempio sintetico:

```http
POST /admin/sources/example-municipal/archive-recoveries
Content-Type: application/json
Accept: application/json

{
  "id": "selected-recovery-001",
  "source_id": "example-municipal",
  "revision": 1,
  "actor": "operator",
  "version_ids": [101, 102],
  "regression_id": "reviewed-correction-001",
  "evidence": {
    "url": "https://municipality.example/review",
    "locator": "Retained correction and same-contract evaluation",
    "observed_at": "2026-10-04T10:00:00Z"
  }
}
```

`GET /admin/sources/{id}/archive-recoveries` restituisce l’audit; i job sono
rileggibili dalle selezioni di reprocessing e dalle viste operative private.
Non esiste un’operazione pubblica API/MCP corrispondente. Prove sintetiche con
PostgreSQL verificano rifiuti atomici, regressioni fallite/superate, immutabilità,
outage della coda e ripresa, classificazione/estrazione reali del runner e
ammissione dei discendenti. Test race, backoffice, vet e build passano.
Il task 29.3 verifica l’adozione e il recupero live descritti sotto;
accettazione e campagne rimangono separate.

## Esito del recupero live

La revisione pulita `74acab8` è adottata in development su public/admin, sei worker
e backup applicativo, con la migrazione additiva eseguita esplicitamente tramite
`scripts/compose-env.sh development run --rm --no-deps --pull never admin migrate`.
Dipendenze, volumi e controlli delle fonti sono conservati; produzione ferma.
Il registro privato `CLN-CONTINUATION-20261004` conserva immagine/configurazione
precedenti, dump, richiesta esatta e ricevute per verifica e recupero.

I due casi archiviati selezionati completano classificazione, estrazione e linking
sul worker ordinario. La stessa richiesta ripetuta restituisce audit e data
identici senza duplicare i job. Marker di archivio, vecchi job, run e risultati
sono riletti e confrontati senza modifiche. La riapertura è sostenuta dalla propria
citazione e conserva i tempi indeterminati; l’aggiornamento delle strade rimane
separato dalla chiusura del sottopasso in via Maremmana.

API e MCP effettivi, con client SDK, restituiscono gli stessi fatti sulla stessa
vista salvata. La verifica comprende confine di conoscenza anteriore al recupero,
vista precedente immutata, proiezione dei nuovi fatti, isolamento da Cascina e
quattordici pagine da due elementi senza duplicazioni o omissioni.

Il nuovo output della riapertura conserva una riapertura, due chiusure, un divieto,
un’attivazione del COC e un aggiornamento generico. Il contratto revisionato richiede
quattro distinti ambiti di divieto: tre sono omessi. Questa nuova risposta del
provider non riproduce la completezza del replay 14/14. La regressione persistente
successiva registra l’omissione con corpus e aspettative invariati, distinguendo
i quattro nuovi confronti worker dagli altri dieci controlli conservati.
Il successo del replay e i fallimenti precedenti rimangono consultabili.
L’ultima regressione fallita impedisce nuovi recuperi e l’accettazione finché una
rivalutazione corretta dello stesso contratto non riesce.

Entrambe le campagne sono rivalutate e restano `extended`: confronti degli
originali, allegati, prove di errore e alcuni ritardi rimangono irrisolti nei
rispettivi report. Il recupero selettivo è verificato; la completezza municipale,
9.3/10.1 e gli altri sei gate operativi restano aperti.

## Correzione ordinaria v27 e chiusura della 9.3

Il seguito conserva gli esiti intermedi sopra e supera la loro omissione: v26
recupera quattro divieti ma aggiunge una chiusura dall’inciso retrospettivo. Il
fallimento viene registrato; v27 esclude il solo inciso, preservando clausole
operative indipendenti. Il nuovo worker ordinario produce una riapertura, due
chiusure mantenute, quattro divieti distinti e COC attivo. Nessuna data di pagina,
allerta o precedente chiusura viene trasferita alla riapertura.

Il contratto invariato passa 14/14, con riscontri riusati attribuiti e nuovo caso
selezionato fresco. API/MCP, paginazione, ambito municipale, viste salvate e limiti
storici passano; vecchi archivi, job e risultati restano immutati. La vista corrente
sceglie l’ultima interpretazione ordinaria proiettata della stessa versione.
La [valutazione MVP completa](observational-trial.md#esito-verificato-del-4-ottobre-2026)
chiude 9.3 con revisioni correttive tracciate, senza cancellare l’altra campagna
`extended` o completare 10.1, accettazione e produzione.
