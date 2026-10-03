---
tipo: "piattaforma"
ultima_revisione: "2026-10-03"
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
anteriore al collaudo; CIN-007 aggiorna lo stato tecnico. Confronti persistenti,
cadenza operativa e copertura continuativa restano da verificare.
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
| CIN-003 | Nuovo flusso IWA | Piano originario; acquisizione implementata da CIN-007, confronti e attivazione ancora aperti |
| CIN-004 | Revisione originaria Calcinaia | Osservazioni conservate; criterio documentale di chiusura superato da CIN-006 |
| CIN-005 | Accesso tecnico Calcinaia | Indice REST osservato; date del contenitore non operative, route degli avvisi e cadenza da collaudare |
| CIN-006 | Criterio corrente del task 26.2 | Sistema di verifica multipla implementato e verificato, con ricevute persistenti; nessun gate di ottenimento licenze/documentazione |
| CIN-007 | Acquisizione Calcinaia | Contratto API e guardie implementati; prove sintetiche e due passaggi HTTP isolati verificati, candidati pending |

## Problemi aperti

I [task 26.2–26.6](../../../openspec/changes/define-toscana-alert-service/tasks.md)
registrano la 26.3 completata e richiedono ancora confronti persistenti, verifica del percorso in
development, valutazione estesa e trial programmato. Il [dossier](cittadino-informato-review.md)
conserva l’inventario tecnico e i criteri correnti. Il task 26.2 resta aperto perché
il confronto e l’ammissione dei candidati non sono ancora implementati e verificati.
La lettura HTTP non certifica aggiornamento continuo, completezza o assenza di avvisi.

## Registro delle scoperte

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
