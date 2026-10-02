---
tipo: "comune"
codice_regione: "09"
codice_istat: "049009"
ultima_revisione: "2026-10-02"
---

# Livorno

## Identità e ambito

Comune di Livorno, provincia di Livorno, ISTAT `049009`. Fonte IWA: `livorno-municipal`. Collegamenti: [Toscana](../README.md), [Municipium](../../../fonti/piattaforme/municipium.md), [stato pubblico documentato](../../../coverage.md#region-09). Le scoperte qui descritte non certificano la raccolta di tutti i canali municipali.

## Mappa delle fonti

| Fonte | Ruolo | Riferimento e limiti |
| --- | --- | --- |
| Sito comunale | Primaria per le notizie esaminate | [Categoria Protezione Civile](https://www.comune.livorno.it/it/news-category/133640); verificare paginazione e perimetro nell'ambiente |
| Risorse Municipium | Allegati collegati dal Comune | `livorno-api.cloud.municipiumapp.it`; riconoscimento, accesso e riuso da valutare per il perimetro richiesto |
| Altri portali comunali | Ricerca separata | Un collegamento nell'interfaccia non ne include automaticamente tutti gli atti nella raccolta |

La [notizia comunale esaminata](https://www.comune.livorno.it/it/news/133640/allerta-arancio-per-forti-temporali-con-rischio-idrogeologico-e-idraulico-del-reticolo-minore) collega l'ordinanza n. 303 del 20 agosto 2026 nella sezione Allegati e una guida alla navigazione nel footer. I due PDF sono su domini Municipium distinti da quello della pagina. Consultazione documentale: 2026-10-02; il contenuto può cambiare.

## Guida operativa

Distinguere gli allegati della notizia dai documenti generici del sito. Verificare ogni dominio esterno rispetto al riferimento ufficiale e ai vincoli di raccolta/riuso. Una risorsa `forbidden` può essere esclusa dal collector prima di qualsiasi richiesta HTTP: non prova un 403 o un file rimosso.

Nel comportamento attuale, [retainPlanned e linkedPDFs](../../../../internal/backend/acquisition/preview.go) raccolgono i link PDF dall'intera pagina, li rendono obbligatori e rifiutano quelli con schema o host diverso dal documento. Il testo principale può essere conservato mentre il controllo rimane incompleto con `required_attachment_unavailable`.

Il ricontrollo del 2026-10-02 ha rivalutato LIV-001 e LIV-002 mediante pagina ufficiale, codice e un nuovo giro programmato; dettagli in LIV-003. Il rifiuto di un allegato necessario interrompe il giro prima dei documenti successivi: distinguere target scoperti e originali conservati.

LIV-004 individua come perimetro proposto il solo host HTTPS `livorno-api.cloud.municipiumapp.it` e `/s3/3612/allegati/`, per gli allegati pertinenti collegati dalle notizie comunali. LIV-005 verifica il download dell'ordinanza con browser ordinario e `DirectHTTP`, correggendo il limite della prima prova con un altro client. La causa della differenza fra client rimane indeterminata; perimetro del collector, persistenza e selezione degli allegati restano da implementare e verificare. Le [note legali comunali](https://www.comune.livorno.it/it/legal_notices) dichiarano CC BY 4.0 salvo eccezioni e contenuti di terzi; il riferimento esterno non concede da solo diritti su ogni risorsa di un host.

## Regole ed eccezioni

| ID | Ambito | Regola corrente | Stato dell'intervento |
| --- | --- | --- | --- |
| LIV-001 | PDF della pagina esaminata | PDF esterni registrati come `forbidden`, senza download | Comportamento applicato dal collector; modifica proposta |
| LIV-002 | Discovery degli allegati | La scansione dell'intera pagina può includere il PDF generico del footer | Comportamento applicato; selezione degli allegati pertinenti proposta |

L'eventuale ammissione di risorse Municipium deve essere esplicita e limitata alla fonte e ai percorsi verificati. La nota di piattaforma non costituisce una autorizzazione generale.

## Problemi aperti

| ID | Problema | Intervento proposto | Criterio di chiusura |
| --- | --- | --- | --- |
| LIV-001 | L'ordinanza collegata è su un host esterno escluso dal collector | Verificare il perimetro di accesso/riuso e supportare le risorse autorizzate | Test di confine e acquisizione completa della fonte verificata |
| LIV-002 | Il PDF del footer viene trattato come allegato obbligatorio | Distinguere gli allegati della notizia dai collegamenti generici | Fixture con ordinanza e footer, poi confronto con la pagina reale |

Modifica, issue e verifiche della correzione: non disponibili. I retry non cambiano questi vincoli. Trial e accettazione restano separati.

## Registro delle scoperte

### LIV-001 — Ordinanza pubblicata su risorse Municipium

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, pagina ufficiale e codice del collector; disponibilità del PDF non verificata qui.
- **Ambito:** `livorno-municipal`, notizia esaminata e host `livorno-api.cloud.municipiumapp.it`.
- **Conoscenza:** confermato per il collegamento e il vincolo nel codice. **Intervento:** modifica proposta.
- **Osservazione:** l'allegato della notizia è su un host diverso; il collector rifiuta queste risorse prima del download.
- **Evidenza:** notizia collegata nella mappa; [collector](../../../../internal/backend/acquisition/preview.go) e [fixture sugli allegati esterni](../../../../internal/backend/acquisition/attachments_test.go).
- **Conseguenza:** il riferimento esterno necessario può rendere incompleto il controllo anche quando la pagina è leggibile.
- **Prossima verifica:** accesso/riuso, ammissione limitata e acquisizione del PDF con la modifica valutata.

### LIV-002 — Guida del sito inclusa fra gli allegati obbligatori

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, pagina ufficiale e codice.
- **Ambito:** footer della stessa notizia; `cloud-ita.municipiumapp.it`.
- **Conoscenza:** confermato per la pagina esaminata. **Intervento:** correzione proposta.
- **Osservazione:** la pagina contiene un PDF nel footer, che la scansione generale dei link tratta come dipendenza obbligatoria.
- **Evidenza:** link Guida alla navigazione della notizia e scansione dell'intero HTML in `linkedPDFs`.
- **Conseguenza:** un documento generico estraneo alla misura diventa una dipendenza obbligatoria del controllo.
- **Prossima verifica:** valutare una selezione degli allegati pertinenti, conservando quelli necessari alla misura; dati delle esecuzioni nel registro privato.

### LIV-003 — Rivalutazione dei due problemi degli allegati

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, pagina ufficiale e codice; nuovo ricontrollo operativo conservato privatamente.
- **Ambito:** stessa notizia di LIV-001 e LIV-002, percorso programmato del collector.
- **Conoscenza:** confermato per il collegamento e il comportamento del codice. **Intervento:** correzioni ancora proposte, non applicate.
- **Osservazione:** i collegamenti all'ordinanza e al PDF del footer restano presenti. Il rifiuto dell'allegato rende incompleta la versione e interrompe `scheduledAcquire` prima dei target successivi.
- **Evidenza:** [notizia ufficiale](https://www.comune.livorno.it/it/news/133640/allerta-arancio-per-forti-temporali-con-rischio-idrogeologico-e-idraulico-del-reticolo-minore), [collector](../../../../internal/backend/acquisition/preview.go); dettagli del nuovo giro nel registro privato.
- **Conseguenza:** confermare il download della pagina principale o la discovery degli elenchi non basta a chiudere LIV-001 e LIV-002.
- **Prossima verifica:** conservare i criteri di chiusura originari e verificare tutti i documenti pianificati dopo una eventuale correzione.

### LIV-004 — Perimetro proposto per gli allegati esterni e accesso ancora aperto

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, pagina ufficiale, note legali e codice; prova di accesso conservata privatamente.
- **Ambito:** ordinanza n. 303 collegata dalla stessa notizia di LIV-001, host `livorno-api.cloud.municipiumapp.it`, percorso `/s3/3612/allegati/`; nessuna autorizzazione generale ad altri host o comuni.
- **Conoscenza:** confermato per il riferimento, la posizione dei link e l'assenza di un'opzione del collector; accessibilità corrente non certificata. **Intervento:** perimetro e modifica proposti, non applicati.
- **Osservazione:** l'ordinanza è collegata nella parte `article-content`, mentre la guida alla navigazione è nel footer. `retainPlanned` impone ancora lo stesso schema/host della pagina; `DirectHTTP` impedisce redirect fuori dall'origine del file richiesto. Il solo cambiamento della configurazione della fonte non può superare il controllo attuale.
- **Evidenza:** [notizia ufficiale](https://www.comune.livorno.it/it/news/133640/allerta-arancio-per-forti-temporali-con-rischio-idrogeologico-e-idraulico-del-reticolo-minore), [note legali](https://www.comune.livorno.it/it/legal_notices), [collector](../../../../internal/backend/acquisition/preview.go), [trasporto delle risorse](../../../../internal/backend/acquisition/direct_http.go). Risposta dettagliata del fetch nel registro privato.
- **Conseguenza:** valutare una regola per fonte/revisione, host esatto e percorso degli allegati, con selezione della parte pertinente della pagina; la raggiungibilità del file rimane una verifica distinta dall'ammissione tecnica.
- **Prossima verifica:** validazione del perimetro con fixture di confine, revisione delle eventuali eccezioni del documento, download ordinario riuscito e controllo completo della fonte. Chiudere LIV-001 e LIV-002 soltanto dopo i rispettivi criteri di chiusura.
- **Collegamenti:** LIV-001, LIV-002 e LIV-003; modifica e issue non disponibili.

### LIV-005 — Il download va verificato con il trasporto effettivo

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, normale browser, `DirectHTTP` del checkout e validazione del file; prove dettagliate private.
- **Ambito:** solo URL dell'ordinanza n. 303 collegata dalla notizia di LIV-001; client del checkout eseguito sul nodo e nel container Crawl4AI di development. Nessun nuovo giro del collector o modifica dei domini ammessi.
- **Conoscenza:** confermato per l'accesso al file verificato; causa della differenza fra client non determinata. **Intervento:** verifica eseguita, integrazione proposta e non applicata.
- **Osservazione:** il click ordinario sul link della pagina ha prodotto un download completato; il file è un PDF con pagine e testo estraibile. Anche `DirectHTTP` ha acquisito il PDF. La prima prova registrata in LIV-004 aveva usato un altro client HTTP e non certificava il risultato del trasporto applicativo.
- **Evidenza:** [notizia ufficiale e allegato](https://www.comune.livorno.it/it/news/133640/allerta-arancio-per-forti-temporali-con-rischio-idrogeologico-e-idraulico-del-reticolo-minore), [DirectHTTP](../../../../internal/backend/acquisition/direct_http.go). Risposte, file originale, hash e validazione conservati privatamente.
- **Conseguenza:** supera il limite di verifica dell'accessibilità in LIV-004 per questo file e questi tentativi; restano aperti LIV-001 e LIV-002, la persistenza attraverso il collector e l'accesso futuro. Il successo non certifica tutti gli allegati dello stesso host.
- **Prossima verifica:** integrare il perimetro autorizzato e il filtro dei link pertinenti, verificare gli originali conservati in un nuovo controllo completo e osservare la continuità del recupero.
- **Collegamenti:** LIV-004; probe privato senza modifica al codice applicativo, issue e integrazione non disponibili.
