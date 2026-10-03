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
selezionati, con referral e condizioni propri. La decisione non attiva un collector.
Le pubblicazioni generano candidati con URL/versione, emittente, territorio ed
espressioni temporali. Verificare misure locali nel sito comunale o nell'atto
richiamato; confrontare rilanci regionali con prodotto CFR, rischio, zona, emissione
e validità corrispondenti. Una misura locale può esistere senza allerta regionale.
Conservare entrambe le evidenze e l'esito datato del confronto, distinguendo
mancato riscontro, controllo indisponibile e conflitto reale. Non trasferire date
della piattaforma alla validità operativa né usare la somiglianza come equivalenza.

Mantenere discovery comunale e CFR indipendente per gli avvisi assenti dalla
piattaforma. Sezioni, paginazione, dipendenze, aggiornamenti e cadenza effettiva sono
da collaudare; questa indagine non certifica endpoint API/feed o intervalli di
polling. Preferire un accesso documentato e ammesso quando disponibile.

Prima della raccolta automatica identificare una base verificata per accesso,
conservazione, trattamento tramite provider e riuso previsto: licenza, disposizione
applicabile o accordo. Non presumere un obbligo generale di permesso individuale,
né un'autorizzazione illimitata dalla sola accessibilità o attribuzione della fonte.
Il riscontro primario verifica i fatti e non sana un'acquisizione priva dei
presupposti richiesti. Le copie pubbliche restano link-only nella specifica IWA;
diritti sui contenuti e sulla banca dati, condizioni tecniche ed eventuali dati
personali richiedono valutazioni distinte.

Le [linee guida AgID](https://www.agid.gov.it/sites/agid/files/2024-05/lg-open-data_v.1.0_1.pdf)
descrivono apertura per impostazione predefinita nei casi applicabili e la distinta
posizione dei testi degli atti ufficiali. L'applicabilità ai titolari e ai contenuti
di questo canale va accertata. La [direttiva sulle banche dati, articoli 7–8](https://eur-lex.europa.eu/legal-content/IT/ALL/?uri=celex%3A31996L0009)
disciplina anche estrazioni sostanziali e, alle condizioni previste, ripetute e
sistematiche; pochi campi e confronto successivo non costituiscono una verifica
dei diritti. Questi riferimenti orientano la revisione, senza concluderla.

## Regole ed eccezioni

| ID | Ambito | Regola e stato |
| --- | --- | --- |
| CIN-001 | Calcinaia | Referral verificato; altri comuni richiedono evidenze proprie |
| CIN-002 | Condizioni | Link copyright senza documento nella homepage consultata; base del flusso da verificare |
| CIN-003 | Nuovo flusso IWA | Acquisizione aggiuntiva con verifica primaria pianificata; implementazione e attivazione aperte |

## Problemi aperti

I [task 26.2–26.6](../../../openspec/changes/define-toscana-alert-service/tasks.md)
richiedono condizioni/accesso verificati, collector separato, confronti persistenti,
valutazione e trial delimitato. Una pagina senza licenza individuata non dimostra
che manchino ovunque condizioni o altre basi applicabili. La lettura HTTP non
certifica aggiornamento continuo, completezza o assenza di avvisi.

## Registro delle scoperte

### CIN-001 — Referral istituzionale di Calcinaia

- **Data e ultima verifica:** 2026-10-03, consultazione web.
- **Ambito:** percorso `/calcinaia/` e referral comunale.
- **Conoscenza:** confermato. **Intervento:** verifica documentale.
- **Osservazione ed evidenza:** il sito comunale collega il canale; riferimenti e limiti in [CLN-009](../../territori/09-toscana/comuni/050004-calcinaia.md#cln-009--riconoscimento-istituzionale-del-canale-di-calcinaia).
- **Conseguenza e prossima verifica:** riconoscimento circoscritto; verificare separatamente ulteriori comuni e contenuti.

### CIN-002 — Link copyright senza documento operativo

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
