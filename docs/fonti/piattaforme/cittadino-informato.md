---
tipo: "piattaforma"
ultima_revisione: "2026-10-04"
---

# Cittadino Informato

## Identità e ambito

Piattaforma web/app di protezione civile presentata da ANCI Toscana in
collaborazione con Regione Toscana. Separare operatore, pubblicatore ed emittente
del messaggio. Il riconoscimento del servizio non riconosce automaticamente ogni
pagina comunale. Primo perimetro investigato: [Calcinaia, ISTAT 050004](../../territori/09-toscana/comuni/050004-calcinaia.md).
Il [coverage tracker](../../coverage.md) resta il riferimento per l'accettazione.

## Mappa delle fonti

| Riferimento | Ambito | Verifica |
| --- | --- | --- |
| [Progetto](https://cittadinoinformato.it/il-progetto/) e [notizia della Regione](https://www.toscana-notizie.it/-/l-app-cittadino-informato-si-rinnova-protocollo-d-intesa-tra-regione-e-anci) | Presentazione e collaborazione istituzionale | Consultazione web 2026-10-03 |
| [Homepage comunale](https://comune.calcinaia.pi.it/) → [Calcinaia](https://cittadinoinformato.it/calcinaia/) | Referral per allerte e piano comunale | Consultazione web 2026-10-03 |
| [Homepage della piattaforma](https://cittadinoinformato.it/) | Link «Note legali e copyright» osservato con destinazione `#` | Lettura HTTP 2026-10-03; licenza del flusso sistematico non individuata nella pagina |

## Guida operativa

CIN-014 registra l'attivazione ordinaria di Calcinaia in development dopo il
preview completo: avvisi con filtro di visualizzazione dal 4 settembre, rischi
oggi/domani e PDF comunali collegati nel percorso revisionato. Raccolta ogni
600 secondi, soglia 1.800 secondi, confronti ogni minuto; accettazione pending.
Le ricevute distinguono riscontro mancante e non comparabilità; errori di citazione
dell'interpretazione ordinaria restano rifiutati. Perimetri primari indipendenti;
consultare la [procedura continuativa](../../operations/cittadino-informato.md#raccolta-e-confronti-continuativi--task-313).


CIN-012 registra il nuovo audit di Calcinaia: identità e referral osservati,
elenco delimitato dalle date di visualizzazione verificato, ma nessuna fonte
della piattaforma registrata nel development ordinario ispezionato. La
[matrice corrente](../../operations/calcinaia-coverage.md) separa questa lacuna
dai fallimenti di interpretazione della primaria. Una vecchia motivazione
documentale nella configurazione live non reintroduce il prerequisito superato;
per attestare continuità servono configurazione applicata, controlli recenti
e ricevute automatiche, con dipendenze e controparti effettivamente verificate.

La [vista di situazione API/MCP](../../backend/public-api-mcp.md#situation-response)
include una sintesi separata `cittadino_informato`. `not_collected` significa che
la vista non dispone di acquisizioni pubblicabili del canale: non significa che
la piattaforma non abbia avvisi. Il trial dedicato non alimenta automaticamente
il database operativo. Identità del canale, raccolta, confronti e accettazione
restano separati, come verificato in CLN-018 della scheda Calcinaia.


La [specifica](../../../openspec/changes/define-toscana-alert-service/specs/official-source-ingestion/spec.md#requirement-scoped-cittadino-informato-acquisition-and-primary-verification)
include la piattaforma come canale aggiuntivo di acquisizione/discovery per comuni
selezionati, con referral propri e verifica multipla Regione/CFR–Comune–piattaforma.
La revisione dell’utente del 2026-10-03 sostituisce il precedente requisito di base
verificata: ottenere licenze, accordi o documentazione giuridica non è il criterio
di chiusura del task 26.2. Occorre un sistema implementato e verificato con confronti
persistenti. La decisione non attiva un collector.
Le pubblicazioni generano candidati con URL/versione, emittente, territorio ed
espressioni temporali. Verificare misure locali nel sito comunale o nell'atto
richiamato; confrontare rilanci regionali con prodotto CFR, rischio, zona, emissione
e validità corrispondenti. Una misura locale può esistere senza allerta regionale.
Conservare entrambe le evidenze e l'esito datato del confronto, distinguendo
mancato riscontro, controllo indisponibile e conflitto reale. Non trasferire date
della piattaforma alla validità operativa né usare la somiglianza come equivalenza.

Mantenere discovery comunale e CFR indipendente per gli avvisi assenti dalla
piattaforma. Il task 26.3 implementa ora le route API degli aggiornamenti e dei
rischi, paginazione delimitata, originali versionati e dipendenze PDF esplicite.
La [procedura](../../operations/cittadino-informato.md) documenta il contratto,
i test sintetici di persistenza e la prova HTTP isolata su Calcinaia del 2026-10-03.
La [revisione preliminare](cittadino-informato-review.md) conserva l’inventario
anteriore al collaudo; CIN-007 aggiorna l’acquisizione e CIN-008 la verifica
persistente del task 26.4, provata con fixture sintetiche. CIN-009 registra la
prova delimitata completa della 26.2 in development, con ricevute reali e
controlli sintetici distinti. CIN-010 completa la valutazione estesa e il trial delimitato; cadenza operativa
e copertura continuativa restano da verificare.
Feed degli avvisi e intervalli di polling non sono stati individuati nelle risorse
consultate. Preferire un accesso API/feed verificato tecnicamente quando disponibile.
I metadati WordPress della pagina non datano i bollettini dinamici.

Ogni ricevuta deve identificare il candidato, i tre ruoli di canale, fonti e versioni
consultate, passaggi probatori, tempi dei controlli, campi confrontati ed esito.
Distinguere conferma, evidenza primaria non trovata, controllo indisponibile,
contenuti non comparabili, controllo non applicabile e conflitto reale. Per una
misura soltanto locale il CFR può essere non applicabile; per un fatto regionale
fa riferimento il prodotto CFR originario anche senza un rilancio comunale.
Non richiedere unanimità fra tre pubblicazioni né contare i rilanci come voti
indipendenti. Un candidato senza il necessario sostegno primario resta diagnostico.

Il task 26.2 si chiude dopo l’implementazione dei task 26.3/26.4 e una verifica
in development del percorso acquisizione–confronto con ricevute ispezionabili per
casi regionali e locali, inclusi errori e conflitti. Raccolta indipendente comunale
e CFR, copie pubbliche link-only, accettazione e trial programmato restano distinti.
Le osservazioni originarie sulle condizioni sono conservate nel registro datato;
CIN-006 ne supera l’uso come prerequisito del piano.

## Regole ed eccezioni

| ID | Ambito | Regola e stato |
| --- | --- | --- |
| CIN-001 | Calcinaia | Referral verificato; altri comuni richiedono evidenze proprie |
| CIN-002 | Condizioni | Link copyright senza documento nella homepage consultata; osservazione conservata, precedente gate superato da CIN-006 |
| CIN-003 | Nuovo flusso IWA | Piano originario; acquisizione e verifica implementate da CIN-007/CIN-008, adozione operativa separata |
| CIN-004 | Revisione originaria Calcinaia | Osservazioni conservate; criterio documentale di chiusura superato da CIN-006 |
| CIN-005 | Accesso tecnico Calcinaia | Indice REST osservato; date del contenitore non operative, route collaudate da CIN-007, cadenza continuativa da verificare |
| CIN-006 | Criterio corrente del task 26.2 | Sistema di verifica multipla implementato e verificato, con ricevute persistenti; nessun gate di ottenimento licenze/documentazione |
| CIN-007 | Acquisizione Calcinaia | Contratto API e guardie implementati; prove sintetiche e due passaggi HTTP isolati verificati, candidati inizialmente pending |
| CIN-008 | Verifica multipla | Ricevute, confronti e proiezione primaria implementati e provati con fixture; CIN-009 aggiunge il percorso development |
| CIN-009 | Percorso development | Dieci ricevute reali/sintetiche distinte ispezionate; perimetro riuscito separato da dipendenze mancanti, adozione programmata e accettazione |

## Problemi aperti

I [task 26.2–26.6](../../../openspec/changes/define-toscana-alert-service/tasks.md)
registrano la 26.2–26.6 completate nel perimetro dichiarato; attivazione
programmata e accettazione restano separate (CIN-010). Il [dossier](cittadino-informato-review.md)
conserva l’inventario tecnico e i criteri correnti. CIN-009 verifica il percorso
completo nel solo perimetro development dichiarato: dipendenze esterne non
consentite e dati regionali non comparabili restano diagnostici, con il primo
passaggio più ampio esplicitamente incompleto.
La verifica implementata consuma selezioni interpretate e risultati dei controlli:
non costituisce da sola discovery delle controparti o accettazione semantica.
La lettura HTTP non certifica aggiornamento continuo, completezza o assenza di avvisi.

## Registro delle scoperte

### CIN-013 — Percorso periodico e dipendenze esterne

- **Data:** 2026-10-04; implementazione e verifiche sintetiche PostgreSQL/race.
- **Ambito:** fonti piattaforma esplicitamente selezionate, primarie conservate indipendenti e allegati PDF con origine/directory revisionate.
- **Intervento:** worker ordinario con ricerca delle controparti, ricevute immutabili e confronto separato dalla proiezione; recuperi e ritorni di versione preservano lo storico. Il contratto consente nomi PDF codificati entro il perimetro esterno verificato.
- **Limiti:** titoli/risorse condivise identificano controparti da consultare, senza dimostrare edizione, emissione o validità. Nessuna acquisizione live continuativa attestata dalla sola implementazione.
- **Evidenza e prossima verifica:** `CIN-CONTINUOUS-20261004`; completare preview della finestra più ampia e osservare controlli ordinari/ricevute prima di chiudere 31.3. Copie link-only e accettazione separate.


### CIN-012 — Audit di Calcinaia e raccolta ordinaria non registrata

- **Data e ultima verifica:** 2026-10-04, homepage ufficiale, identità REST e lista aggiornamenti della piattaforma, registro development in sola lettura.
- **Ambito:** solo Calcinaia `050004`; lista con pubblicatore `comune_calcinaia`, cento elementi massimi per pagina e filtro di visualizzazione dal 4 settembre 2026. Nessun riconoscimento o attivazione di altri comuni.
- **Osservazione confermata:** identità e referral concordano; la lista espone 36 comunicazioni su una pagina. Le date di visualizzazione non certificano la pubblicazione o validità degli avvisi. Nel registro ordinario ispezionato non è presente una fonte Cittadino Informato; le precedenti prove isolate non dimostrano acquisizione e confronti continuativi.
- **Intervento:** audit verificato, nessuna modifica dei controlli live. Pubblicata la [matrice di copertura](../../operations/calcinaia-coverage.md); il task 31.3 mantiene adozione programmata, discovery delle controparti, dipendenze e ricevute fra le verifiche aperte. La motivazione documentale storica nella configurazione primaria è superata dal piano 26.2, non da un'abilitazione implicita.
- **Evidenza:** referral dalla [homepage comunale](https://www.comune.calcinaia.pi.it/), [canale riconosciuto](https://cittadinoinformato.it/calcinaia/), registro privato `CLN-COVERAGE-20261004` con risposte e manifest; vedi CLN-027 nella guida comunale.
- **Prossima verifica:** configurazione delimitata, preview completo con dipendenze, raccolta cadenzata e confronti persistenti di nuove versioni. Copertura, accettazione, validità e pubblicazione rimangono distinte.

### CIN-010 — Confronti revisionati e attribuzione pubblica verificati

- **Data e ultima verifica:** 2026-10-03; originali conservati, codice corrente e prova development delimitata.
- **Conoscenza:** confermata nel perimetro delle evidenze.
- **Osservazione:** Le nuove selezioni dei casi reali omettono le date di pubblicazione dal campo edizione e includono la validità letterale disponibile. Il documento primario resta completo; il canale piattaforma conserva la dipendenza esterna mancante e l’identità non provata. La concordanza documentale dei contenuti operativi non forza l’ammissione automatica.
- **Intervento e verifica:** task 26.5/26.6 completati; ricevute pubbliche con tre ruoli, evidenze visibili e limiti, equivalenza API/MCP con client reale, prove sintetiche per conflitti, indisponibilità, revisioni e canali privati. Il trial development separato verifica acquisizioni indipendenti, scelta del solo comune e revoca. Registro privato `CIN-TRIAL-20261003-01`; [procedura](../../operations/cittadino-informato.md#valutazione-e-trial-delimitato--task-265266).
- **Limiti e prossima verifica:** nessuna nuova approvazione umana inventata, copia pubblica, fonte accettata o raccolta programmata. Discovery delle controparti, orchestrazione continuativa e accettazione restano separati.


### CIN-009 — Percorso completo delimitato verificato in development

- **Data e ultima verifica:** 2026-10-03, acquisizioni HTTP reali, ricevute persistenti, controlli sintetici e smoke development.
- **Ambito:** Calcinaia `050004`, una pagina Notizie primaria, criticità CFR, aggiornamenti dal 1° ottobre e rischi oggi/domani della piattaforma; dettaglio locale esatto selezionato separatamente.
- **Conoscenza:** comportamento confermato nel perimetro provato. **Intervento:** task 26.2 verificato su PostgreSQL/RustFS development con dati dedicati; nessuna raccolta programmata o pubblicazione del trial.
- **Osservazione ed evidenza:** dieci ricevute con tre ruoli, originali/versioni, passaggi, tempi e risultati; primaria locale indipendente riusata, rilanci reali non comparabili e candidati incompleti diagnostici. Sette controlli sintetici distinguono riscontro, indisponibilità e conflitto senza maggioranza. Il primo passaggio dal 23 settembre resta incompleto per un PDF comunale esterno al contratto; nessuna estensione dei permessi. [Procedura e limiti](../../operations/cittadino-informato.md#prova-completa-in-development--task-262), registro privato `CIN-DEV-20261003-01`.
- **Conseguenza e prossima verifica:** completa la prova mancante di CIN-008; non certifica colori CFR, semantica degli atti, copertura della finestra ampia o attribuzione API/MCP delle ricevute. Completare 26.5 prima dell’adozione programmata 26.6 e dell’accettazione separata.

### CIN-008 — Ricevute persistenti e ammissione primaria verificate

- **Data e ultima verifica:** 2026-10-03, codice, fixture sintetiche e PostgreSQL isolato; nessun nuovo fetch del canale reale.
- **Ambito:** verificatore comune ai tre ruoli Regione/CFR–Comune–piattaforma, contratto comunale selezionato; prove redistribuibili con identità sintetiche.
- **Conoscenza:** comportamento confermato nei casi provati. **Intervento:** task 26.4 applicato al codice e verificato; nessuna migrazione o attivazione live.
- **Osservazione ed evidenza:** passaggi da originali integri o OCR completo, ricevute immutabili, confronto per campo senza maggioranza, proiezione esclusivamente primaria, idempotenza, revisioni/ritorni e retention delle controparti. [Procedura e test](../../operations/cittadino-informato.md#verifica-persistente-e-ammissione--task-264).
- **Conseguenza e prossima verifica:** completa l’ingresso programmatico della 26.4; non certifica semantica delle selezioni, copertura reale o orchestrazione continuativa. Eseguire il percorso development della 26.2 e la valutazione 26.5 prima del trial 26.6.

### CIN-007 — Acquisizione API delimitata e persistenza verificate

- **Data e ultima verifica:** 2026-10-03, codice, fixture sintetiche/PostgreSQL isolato e due controlli HTTP delimitati.
- **Ambito:** percorso Calcinaia, aggiornamenti del pubblicatore selezionato e rischi oggi/domani; nessuna estensione ad altri Comuni.
- **Conoscenza:** confermato nel perimetro provato. **Intervento:** adattatore applicato al codice e verificato in isolamento; nessuna attivazione live.
- **Osservazione ed evidenza:** identità comunale e API dettaglio verificate; originali JSON versionati, link pubblici separati, date di pubblicazione distinte da visualizzazione/validità. Test e limiti nella [procedura](../../operations/cittadino-informato.md); ricevuta privata `CIN-ACQUIRE-20261003-01`.
- **Conseguenza e prossima verifica:** aggiorna il collaudo mancante di CIN-005 e completa 26.3. I candidati restano pending; implementare 26.4 prima di chiudere 26.2. Non dimostra copertura, accettazione o confronti con Comune/CFR.

### CIN-001 — Referral istituzionale di Calcinaia

- **Data e ultima verifica:** 2026-10-03, consultazione web.
- **Ambito:** percorso `/calcinaia/` e referral comunale.
- **Conoscenza:** confermato. **Intervento:** verifica documentale.
- **Osservazione ed evidenza:** il sito comunale collega il canale; riferimenti e limiti in [CLN-009](../../territori/09-toscana/comuni/050004-calcinaia.md#cln-009--riconoscimento-istituzionale-del-canale-di-calcinaia).
- **Conseguenza e prossima verifica:** riconoscimento circoscritto; verificare separatamente ulteriori comuni e contenuti.

### CIN-002 — Link copyright senza documento operativo

Il prerequisito proposto in questa voce è superato da CIN-006; l’osservazione HTTP resta conservata.

- **Data e ultima verifica:** 2026-10-03, lettura HTTP ordinaria e consultazione web.
- **Ambito:** homepage e link «Note legali e copyright».
- **Conoscenza:** destinazione osservata confermata; condizioni da verificare. **Intervento:** diagnosi documentata.
- **Osservazione ed evidenza:** il link punta a `#`; licenza del flusso sistematico non individuata nelle pagine consultate. Non dimostra divieto o assenza di basi applicabili. Dettaglio nel registro privato `CIN-REVIEW-20261003-01`.
- **Conseguenza e prossima verifica:** non usare il footer come evidenza di permesso; identificare titolarità, condizioni e base applicabile prima dell'attivazione.

### CIN-003 — Acquisizione con verifica primaria pianificata

- **Data e ultima verifica:** 2026-10-03, decisione dell'utente e coerenza documentale.
- **Ambito:** comuni selezionati del servizio Toscana, primo perimetro Calcinaia.
- **Conoscenza:** requisito confermato. **Intervento:** pianificato, non applicato al runtime.
- **Osservazione ed evidenza:** proposta, design, tre specifiche e task 26 del [change Toscana](../../../openspec/changes/define-toscana-alert-service/proposal.md) descrivono candidati e riscontro comunale/CFR.
- **Conseguenza e prossima verifica:** completare presupposti, implementazione e trial; il piano non certifica accettazione, diritti o pubblicazione.

### CIN-004 — Revisione distinta per accesso, conservazione, provider e riuso

Il criterio documentale di chiusura di questa voce è superato da CIN-006; le osservazioni originarie restano conservate.

- **Data e ultima verifica:** 2026-10-03, letture HTTP delimitate di progetto, contatti, privacy e cookie policy.
- **Ambito:** canale Calcinaia; nessuna estensione ad altri comuni.
- **Conoscenza:** ANCI Toscana identificata come titolare del trattamento dei visitatori; diritti sui contenuti e sulla banca dati da accertare. **Intervento:** revisione preliminare documentata, nessuna attivazione.
- **Osservazione ed evidenza:** la privacy policy non è una licenza dei contenuti; il footer copyright resta senza documento raggiungibile. Il [dossier](cittadino-informato-review.md) distingue le quattro operazioni IWA e le possibili basi da valutare. Registro privato `CIN-REVIEW-20261003-02`.
- **Conseguenza e prossima verifica:** chiudere la matrice con licenza, base normativa applicabile o accordo verificati; conservare link-only, verifiche primarie e accettazione separati. Nessun divieto generale dedotto dall'assenza di licenza individuata.

### CIN-005 — API esposta e metadati del contenitore non equivalenti agli avvisi

- **Data e ultima verifica:** 2026-10-03, pagina Calcinaia, indice REST, metadati WordPress e risposte API di identità/progetto.
- **Ambito:** `/calcinaia/` e collegamenti REST dichiarati dal suo HTML.
- **Conoscenza:** accessibilità delle risorse consultate confermata; contratto degli avvisi non collaudato. **Intervento:** diagnosi documentata.
- **Osservazione ed evidenza:** l'indice espone route di rischi e aggiornamenti con filtri comunali; i metadati della pagina risalgono al 2017–2018, il corpo REST è vuoto e l'HTML mostra un bollettino del 2026. Percorsi e limiti nel [dossier](cittadino-informato-review.md); ricevute private `CIN-REVIEW-20261003-02`.
- **Conseguenza e prossima verifica:** non acquisire il contenitore come se fosse un avviso né usarne le date come emissione; verificare le route dei contenuti, cadenza e condizioni prima del collector. La precedente assenza di un endpoint certificato non significa assenza di un'API tecnicamente esposta.

### CIN-006 — Verifica multipla come criterio di chiusura del task 26.2

- **Data e ultima verifica:** 2026-10-03, decisione esplicita dell’utente e revisione documentale; nessun nuovo fetch o test runtime.
- **Ambito:** flusso IWA Regione Toscana/CFR–Comune–cittadinoinformato.it, primo comune Calcinaia.
- **Conoscenza:** requisito confermato; gate documentale delle voci CIN-002/CIN-004 superato. **Intervento:** applicato a proposta, design, specifiche, checklist e guide; sistema da implementare.
- **Osservazione ed evidenza:** il [task 26.2](../../../openspec/changes/define-toscana-alert-service/tasks.md) ora richiede un percorso funzionante acquisizione–verifica con ricevute persistenti e verifica delimitata in development, senza ottenere licenze, accordi o documentazione della base giuridica.
- **Conseguenza:** CFR primario per dati regionali, Comune/atto primario per misure locali; non applicabilità, mancato riscontro, indisponibilità, non comparabilità e conflitto restano distinti. Nessun voto a maggioranza o obbligo di tre pubblicazioni concordi.
- **Prossima verifica:** implementare 26.3/26.4 e verificare il nuovo 26.2; poi valutazione 26.5 e trial programmato 26.6. Copie pubbliche link-only e accettazione restano separati.

### CIN-011 — Disponibilità del canale distinta nella situazione

- **Data e ultima verifica:** 2026-10-03; test sintetici e risposta API/MCP development, senza nuovi fetch della piattaforma.
- **Ambito:** sintesi di situazione del database operativo per Calcinaia, distinta dal trial isolato CIN-010.
- **Conoscenza e intervento:** comportamento confermato; presentazione verificata dal task 28.
- **Osservazione ed evidenza:** identità del canale ricavata dalla piattaforma configurata, senza deduzione dal nome della fonte. Una fonte configurata ma senza dati pubblicabili è indisponibile; senza canale nella vista la sintesi è `not_collected`. Nessuna importazione dei risultati privati del trial o conferma implicita dei candidati. [CLN-018](../../territori/09-toscana/comuni/050004-calcinaia.md#cln-018--situazione-apimcp-centrata-su-conclusioni-e-fonti).
- **Conseguenza e prossima verifica:** rendere esplicita l'assenza dei dati operativi; raccolta continuativa, riscontro dei contenuti e accettazione restano separati.


### CIN-014 — Configurazione ordinaria e confronti periodici di Calcinaia

- **Data:** 2026-10-04; ambiente development, solo Calcinaia `050004`.
- **Osservazione confermata:** sorgente distinta `calcinaia-cittadino-informato`, revisione 1, raccolta abilitata dopo preview completo di 36 avvisi, un elenco e due risposte rischi. I 39 originali/versioni e le 55 risorse obbligatorie, inclusi 16 PDF, sono completi e riletti con verifica degli hash.
- **Comportamento verificato:** worker periodico con ricevute automatiche, confronto degli originali primari conservati e controlli indipendenti delle tre fonti. Raccolta effettiva 600 secondi, soglia di ritardo 1.800 secondi; confronti ogni minuto.
- **Limiti:** filtro assoluto sulla visualizzazione dal `2026-09-04`, nessuna copertura storica universale; identità/edizioni locali non stabilite e metadati CFR mancanti rimangono diagnostici. Errori di citazione impediscono alcune interpretazioni ordinarie; una raccolta completa non certifica correttezza semantica o fonte accettata.
- **Evidenza:** `CIN-CONTINUOUS-20261004`, [adozione e verifiche](../../operations/cittadino-informato.md#adozione-ordinaria-verificata--4-ottobre-2026); originali, configurazioni e audit operativi privati.
- **Prossima verifica:** seguire i risultati ordinari e le controparti, correggere soltanto difetti sostenuti da evidenza e verificare l'accettazione nel coverage tracker.
