# Guide di regioni e comuni

Queste schede conservano le conoscenze sulle fonti: dove cercare, come raccogliere i documenti, quali eccezioni rispettare e quali problemi restano aperti. La guida corrente precede il registro datato delle scoperte.

Il [coverage tracker](../coverage.md) resta il riferimento pubblico per stato e accettazione delle fonti. Le schede non sono un feed di allerte né un inventario dello stato live: configurazioni, attivazioni ed esiti delle esecuzioni si verificano nell'ambiente interessato. Una raccolta riuscita non dimostra copertura completa, accettazione o pubblicazione pubblica.

## Indice

| Regione | Codice | Scheda | Comuni documentati |
| --- | --- | --- | --- |
| Toscana | `09` | [Guida regionale](09-toscana/README.md) | [Calcinaia](09-toscana/comuni/050004-calcinaia.md), [Cascina](09-toscana/comuni/050008-cascina.md), [Livorno](09-toscana/comuni/049009-livorno.md), [Pisa](09-toscana/comuni/050026-pisa.md), [Pontedera](09-toscana/comuni/050029-pontedera.md) |

L'indice contiene i territori già investigati. L'assenza di una scheda significa che qui non abbiamo ancora raccolto conoscenze, non che manchino pubblicazioni ufficiali. L'elenco nazionale completo rimane nel coverage tracker.

## Struttura e identità

Usare `docs/territori/<codice-regione>-<nome>/README.md` per la regione e `comuni/<codice-ISTAT>-<nome>.md` per ciascun comune. I codici sono stringhe con zeri iniziali: due cifre per la regione, sei per il comune. Il nome rende leggibile il percorso; il codice identifica il territorio. Cambi di identità o confini richiedono riferimenti al registro geografico e alle schede precedenti, senza trasferire automaticamente regole.

Creare le schede dal [modello](MODELLO.md) quando inizia la ricerca. Ogni scheda contiene identità e ambito, mappa delle fonti, guida operativa, regole ed eccezioni, problemi aperti e registro delle scoperte. La scheda regionale aggiunge l'indice dei comuni documentati e distingue i prodotti regionali.

Descrivere le caratteristiche condivise delle piattaforme in `docs/fonti/piattaforme/`; iniziamo con [Municipium](../fonti/piattaforme/municipium.md). I comuni referenziano queste note e documentano le proprie differenze. Una scoperta su Livorno non certifica il comportamento di un altro sito Municipium. Le regole regionali non estendono automaticamente canali, permessi o copertura municipale.

## Come registrare una scoperta

Prima di investigare una fonte leggere la scheda territoriale, le note della piattaforma e gli eventuali problemi collegati. Dopo l'indagine:

1. Aggiungere una voce datata con identificativo stabile, fonte o prodotto, osservazione, evidenza, conseguenza e prossima verifica.
2. Distinguere lo stato della conoscenza (`ipotesi`, `confermato`, `superato`) da quello dell'intervento (`proposto`, `applicato`, `verificato`, oppure `non necessario`). Una causa può essere confermata mentre la correzione è ancora proposta.
3. Aggiornare la guida corrente e i problemi aperti quando la nuova evidenza cambia le indicazioni applicabili.
4. Collegare issue, configurazione, modifica e verifica quando esistono. Usare `non disponibile` per un riferimento mancante; non inventare ticket o risultati.
5. Registrare la rivalutazione con una nuova voce che richiama l'ID precedente. Conservare la scoperta originaria e indicare nella guida quale conclusione è stata superata.

Usare ID progressivi per scheda, per esempio `TOS-001`, `CLN-001`, `CAS-001`, `LIV-001`, `PIS-001` e `MUN-001`. Un ID non cambia quando cambia l'ordine delle voci. Separare la data della scoperta, l'ultima verifica dell'evidenza e la revisione del documento. Quest'ultima non certifica un nuovo fetch.

Precisare l'ambito: fonte/prodotto, sezione, dominio o percorso e, per un comportamento operativo, ambiente e revisione della configurazione. Distinguere il comportamento presente dal comportamento desiderato. Una regola proposta diventa applicata solo con un riferimento alla modifica effettiva; diventa verificata con evidenza adeguata al suo ambito.

## Evidenze pubblicabili e archivio privato

Nel repository conservare conclusioni pubblicabili, URL ufficiali, riferimenti al codice e verifiche con fixture redistribuibili. Riassumere il necessario senza copiare pagine, PDF o log non autorizzati alla redistribuzione. Un link ufficiale dimostra il collegamento osservato, non una licenza generale né l'autorizzazione ad acquisire tutti i contenuti dello stesso dominio.

Log, risposte originali, dump, esiti dettagliati di esecuzioni e impostazioni operative restano nell'archivio privato. Su questo checkout la posizione prevista è `.local/operations/territori/`, già esclusa da Git, con una gerarchia territoriale corrispondente. Referenziare le evidenze riservate tramite ID nel registro privato, senza inserire link pubblici a file assenti dal repository. Non conservare credenziali nelle schede, nemmeno private. La cartella locale richiede una conservazione e un backup privati separati; Git non la protegge.

Quando una scoperta è sostenuta solo da evidenza privata, conservarne lì il dettaglio e mantenere esplicito nella scheda pubblica ciò che deve ancora essere verificato. L'archivio privato può contenere osservazioni per ambiente; la scheda pubblica non deve trasformarle in una dichiarazione di disponibilità o accettazione.

## Manutenzione

Le note di [Cittadino Informato](../fonti/piattaforme/cittadino-informato.md)
distinguono riconoscimento istituzionale, nuovo flusso pianificato con verifica
primaria e condizioni di acquisizione ancora da verificare.

Aggiornare le schede durante l'indagine o la modifica relativa alla fonte, prima di dichiarare conclusa la verifica. Rivalutare le indicazioni quando cambiano sito, piattaforma, configurazione, diritti d'uso o comportamento del collector. Una nuova evidenza contraria riapre il problema.

Verificare link locali, codici territoriali, identificativi delle scoperte, coerenza tra guida e registro e distinzione fra proposte e regole applicate. Collegare specifica e task OpenSpec quando cambia una capacità; registrare gli aggiornamenti nel [changelog](../../CHANGELOG.md). Le verifiche documentali non richiedono fetch o chiamate ai modelli.
