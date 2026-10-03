# Cittadino Informato — revisione del canale Calcinaia

Revisione ed evidenze HTTP: **3 ottobre 2026**. Ambito: Comune di Calcinaia,
ISTAT `050004`, pagina `https://cittadinoinformato.it/calcinaia/` e riferimenti
pubblicati dalla piattaforma. Questa è la revisione preliminare del
[task 26.2](../../../openspec/changes/define-toscana-alert-service/tasks.md),
completato successivamente dalla prova development descritta sotto. Le
[note correnti](cittadino-informato.md) conservano le regole
applicabili; risposte originali e ricevute sono nel registro privato
`CIN-REVIEW-20261003-02`.

## Esito originario e revisione del criterio

Referral e accessibilità tecnica sono confermati nel perimetro osservato.
La revisione originaria non aveva individuato una base verificata per l’intero
flusso e proponeva una chiusura documentale. La successiva decisione dell’utente
sostituisce quel criterio: il task 26.2 deve essere chiuso mediante un sistema
funzionante di verifica multipla Regione Toscana/CFR–Comune–cittadinoinformato.it,
senza ottenere licenze, accordi o documentazione della base giuridica.

La successiva 26.4 implementa confronti persistenti e ammissione; la prova
delimitata in development della 26.2 verifica il percorso completo con dieci
ricevute reali/sintetiche distinte, conservate nel registro privato
`CIN-DEV-20261003-01`. Vedere
[procedura e limiti](../../operations/cittadino-informato.md#prova-completa-in-development--task-262).
La [implementazione della 26.3](../../operations/cittadino-informato.md) verifica
l’acquisizione delimitata; ricevuta privata `CIN-ACQUIRE-20261003-01`.
Nessuna attivazione è stata effettuata durante questa revisione; le copie pubbliche
rimangono link-only. Le osservazioni originarie sotto riportate restano evidenze
storiche e non costituiscono il gate di chiusura corrente.

## Soggetti e diritti

| Aspetto | Evidenza osservata | Limite e verifica restante |
| --- | --- | --- |
| Canale comunale | [Homepage del Comune](https://comune.calcinaia.pi.it/) con referral alla pagina Calcinaia | Riconosce il canale, senza concedere diritti sull'intera piattaforma |
| Servizio | [Presentazione](https://cittadinoinformato.it/il-progetto/) e [contatti](https://cittadinoinformato.it/contatti/) ANCI Toscana | Distinguere gestione tecnica, pubblicatore e autore di ogni contenuto |
| Trattamento dei visitatori | [Privacy policy](https://www.iubenda.com/privacy-policy/85024701), ultima modifica dichiarata 19 maggio 2026, indica ANCI Toscana come titolare | Non identifica tutti i titolari dei diritti sui contenuti o sulla banca dati e non autorizza il trattamento IWA tramite provider |
| Condizioni dei contenuti | Footer della homepage e della pagina comunale: «Note legali e copyright» con destinazione `#` | Nessun documento di licenza raggiunto da quel link; titolarità, eventuali riserve e condizioni del flusso restano da accertare |
| Emittenti | La presentazione descrive avvisi delle amministrazioni e contenuti regionali | Identificare emittente e diritti del singolo avviso; non trasferire quelli di un atto a immagini, mappe, materiali incorporati o raccolta complessiva |

## Accesso tecnico osservato

Questo inventario riguarda la revisione preliminare. Il successivo collaudo
26.3 aggiorna lo stato delle route aggiornamenti/rischi come registrato in
[CIN-007](cittadino-informato.md#cin-007--acquisizione-api-delimitata-e-persistenza-verificate).

Le letture sono consultazioni manuali delimitate, senza login, scansione di altri
comuni, polling, chiamate a provider o operazioni di scrittura sulla piattaforma.
Le risposte osservate non costituiscono un contratto di disponibilità o di riuso.

| Riferimento | Riscontro HTTP del 3 ottobre | Significato e limiti |
| --- | --- | --- |
| [Pagina Calcinaia](https://cittadinoinformato.it/calcinaia/) | `200`, HTML; navigazione verso `/calcinaia/aggiornamenti/` | Sezione candidata; paginazione, allegati e completezza degli aggiornamenti non collaudati |
| [Indice REST comunale](https://cittadinoinformato.it/calcinaia/wp-json/) | `200`, JSON, collegato nell'HTML; espone i namespace `cittadino/v1`, `cittadino/v2` e `wp/v2` | Descrizione tecnica pubblicata dal server; non è una licenza o un impegno di stabilità |
| [Identità comunale](https://cittadinoinformato.it/calcinaia/wp-json/cittadino/v2/comune?nome=calcinaia) | `200`, JSON con nome e slug Calcinaia | L'identificativo interno della piattaforma non sostituisce ISTAT; equivalenza territoriale da vincolare alla configurazione |
| [Presentazione via API](https://cittadinoinformato.it/calcinaia/wp-json/cittadino/v2/info-progetto) | `200`, JSON con progetto e contatti | Non individuata qui una licenza del flusso IWA |
| Route aggiornamenti nell'indice REST | `GET /cittadino/v2/aggiornamenti`, parametro stringa `comune` obbligatorio; `per_page`, `page`, `ente`, `tipologia`, `data_inizio`, `data_fine`; dettaglio con ID | Route dichiarate, risposte degli avvisi non acquisite in questa revisione; semantica dei filtri, ordinamento, totalità e dipendenze da verificare |
| Route rischi nell'indice REST | Route `cittadino/v2/rischi`, con varianti per oggi/domani e tipologia | Non collaudate; niente equivalenza presunta con un prodotto CFR o con misure comunali |
| [Metadati WordPress della pagina](https://cittadinoinformato.it/calcinaia/wp-json/wp/v2/pages/64) | `200`; creazione nel 2017, modifica nel 2018, `content.rendered` vuoto | La pagina HTML espone un bollettino del 2026: i metadati del contenitore non datano gli avvisi dinamici e questo endpoint non ne fornisce il corpo |
| [robots.txt](https://cittadinoinformato.it/robots.txt) | `404` | Non dimostra permesso, divieto o assenza di altre riserve |

Non sono stati individuati nelle risorse consultate un feed ufficiale degli avvisi,
una cadenza documentata, limiti di polling o garanzie di conservazione storica.
Non equivale a certificarne l'assenza. I dieci minuti del collector IWA sono una
configurazione interna, non una cadenza documentata dalla piattaforma.

## Sistema richiesto e criteri correnti di chiusura

| Ruolo | Verifica | Esito da conservare |
| --- | --- | --- |
| Regione Toscana/CFR | Rischio, zona, prodotto, emissione/versione e validità delle affermazioni regionali | Campi sostenuti o conflitto comparabile; non applicabile per una misura soltanto locale |
| Comune/atto richiamato | Azione, luogo, emittente ed espressioni di validità delle misure locali | Sostegno primario dei singoli campi, assenza di riscontro o controllo indisponibile; rilancio comunale non necessario per confermare il CFR |
| Cittadino Informato | Candidato, URL/versione, territorio, tipo di comunicazione e significato delle date | Corroborazione per campo, non comparabilità, conflitto o candidato diagnostico senza sostegno primario |

Ogni ricevuta conserva identità del candidato e dei tre ruoli, fonti/versioni e
passaggi quando disponibili, tempi dei controlli, campi confrontati ed esiti con
motivazione. L’assenza di una pubblicazione non prova assenza di rischio o revoca.
Tre pubblicazioni concordi non sono richieste e due rilanci non prevalgono per
maggioranza sul prodotto originario.

I criteri della 26.2, verificati successivamente alla revisione preliminare, sono:

1. Acquisizione delimitata e identità separate del task 26.3: implementate e
   verificate successivamente alla revisione preliminare.
2. Confronti persistenti e ammissione dei soli campi sostenuti implementati dalla
   26.4, mantenendo indipendenti raccolta comunale e CFR.
3. Controllo delimitato in development del percorso completo per casi regionali
   e locali, con ricevute persistite e rilette; reali non comparabili/diagnostici
   distinti dai controlli sintetici positivi e di conflitto.
4. Mancato riscontro, indisponibilità, non comparabilità, non applicabilità e
   conflitto verificati; candidati senza sostegno primario restano diagnostici.

L’inventario HTTP e la revisione preliminare non chiudevano il task runtime.
La successiva prova completa chiude la 26.2, conservando il fallimento della
finestra ampia della piattaforma su un PDF comunale esterno al suo contratto;
soltanto il perimetro successivo riuscito è dichiarato completo.
La valutazione estesa del task 26.5 e il successivo trial programmato del task 26.6
restano separati, come accettazione delle fonti e pubblicazione. La chiusura della
26.2 non è subordinata a licenze, accordi o documentazione giuridica.
