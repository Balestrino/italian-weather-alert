# Interpretazione della riapertura parziale di Calcinaia

Verifica del 4 ottobre 2026, avanzamento dei task 9.3 e 10.1 della
[checklist Toscana](../../openspec/changes/define-toscana-alert-service/tasks.md).
La correzione del software supera il fallimento di merge del caso conservato;
campagne, accettazione e verifica del percorso ordinario live restano pendenti.

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
precedenti non sono stati sostituiti dal replay locale. Per proseguire occorre
adottare la revisione verificata, rivalutare lo stesso contratto senza indebolire
le aspettative, rielaborare esplicitamente i soli casi selezionati e confrontare
risultati persistenti, storico e API/MCP. Completare anche confronti, prove di
errore e ritardi ancora mancanti prima di chiudere 9.3/10.1. La revisione del
software non conclude gli altri [gate operativi](toscana-readiness.md).
