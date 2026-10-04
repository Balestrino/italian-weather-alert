---
tipo: "comune"
codice_regione: "09"
codice_istat: "050029"
ultima_revisione: "2026-10-04"
---

# Pontedera

## Identità e ambito

Comune di Pontedera, provincia di Pisa, ISTAT `050029`. Fonte candidata IWA: `pontedera-municipal`. Collegamenti: [Toscana](../README.md), [stato pubblico documentato](../../../coverage.md#region-09). Questa scheda avvia il registro di ricerca; non certifica attivazione o acquisizione programmata.

## Mappa delle fonti

| Fonte | Ruolo | Riferimento e limiti |
| --- | --- | --- |
| Sito comunale, Avvisi | Archivio ufficiale da valutare | [Avvisi](https://www.comune.pontedera.pi.it/tipi_notizia/avviso/); pubblicazioni e collegamenti di paginazione osservati il 2026-10-02 |
| Notizie e Comunicati | Canali distinti | Collegati dal footer degli Avvisi; contenuti e raccolta da verificare separatamente |
| Unione Valdera e portali degli atti | Fonti da investigare separatamente | Referral nel sito comunale; competenza, perimetro, accesso e riuso da verificare |

## Guida operativa

Il precollaudo del 2026-10-04 produce un report individuale **pending**,
con valutazione, osservazione, accettazione e pubblicazione distinte. La
[campagna municipale](../../../operations/municipal-acceptance.md) mantiene
sette intervalli completi di 24 ore e le prove di originali, allegati ed errori.
PON-002 registra i prerequisiti; il report non certifica copertura completa.

Verificare regione, comune, fonte e worker come passaggi distinti. L'abilitazione del comune non avvia la raccolta di una fonte disattivata. Lo stato live si legge nel registro dell'ambiente; una preview storica nel coverage tracker non sostituisce un nuovo controllo programmato.

L'archivio Avvisi mostra date con anno a due cifre e collegamenti a pagine successive. Confrontare formato delle date, sezioni e paginazione reali con la configurazione prima di dichiarare completo un periodo. Distinguere rilanci CFR e misure comunali, verificando gli atti e gli allegati necessari. Una consultazione web non verifica il percorso Crawl4AI, il bootstrap o la conservazione degli originali.

## Regole ed eccezioni

| ID | Ambito | Regola corrente | Stato dell'intervento |
| --- | --- | --- | --- |
| PON-001 | Abilitazioni territoriali e della fonte | Comune abilitato e fonte abilitata sono prerequisiti distinti della raccolta | Applicato nel controllo di ammissione; stato effettivo nel registro privato |

Non sono introdotte eccezioni ai domini, agli allegati o ai vincoli di accesso/riuso.

## Problemi aperti

Acquisizione programmata corrente, paginazione, altri canali, allegati e atti originali richiedono verifica nel perimetro autorizzato. Trial e accettazione restano quelli del coverage tracker. Una fonte disattivata va riportata come tale, senza trattarla come un fetch fallito o attivarla durante una verifica.

## Registro delle scoperte

### PON-001 — Distinguere abilitazione del comune e raccolta della fonte

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, documentazione, codice e consultazione degli Avvisi; stato operativo conservato privatamente.
- **Ambito:** `pontedera-municipal`, prerequisiti di raccolta e archivio Avvisi.
- **Conoscenza:** confermato. **Intervento:** non necessario; nessuna attivazione o modifica applicata.
- **Osservazione:** il collector ammette il lavoro solo con i prerequisiti territoriali e della fonte. L'archivio ufficiale è consultabile e contiene paginazione, ma questo non dimostra l'esecuzione della raccolta programmata.
- **Evidenza:** [Avvisi ufficiali](https://www.comune.pontedera.pi.it/tipi_notizia/avviso/), [regole di attivazione](../../../operations/environments.md), [scheduler](../../../../internal/backend/acquisition/schedule.go). Stato del registro e configurazione effettiva nell'archivio privato.
- **Conseguenza:** nei ricontrolli distinguere fonti attive, disattivate e controlli falliti; la preview storica rimane separata.
- **Prossima verifica:** quando la raccolta sarà abilitata, verificare un nuovo giro programmato, originali conservati e limiti di paginazione prima del trial.
- **Collegamenti:** task 11.3 e 11.5 del [servizio Toscana](../../../../openspec/changes/define-toscana-alert-service/tasks.md); ulteriori issue e modifiche non disponibili.


### PON-002 — Report individuale pendente e collaudo indipendente

- **Data:** 2026-10-04. **Ultima verifica:** 2026-10-04, registro dell'ambiente, apertura della campagna e test sintetici con PostgreSQL isolato; nessuna nuova consultazione web della fonte.
- **Ambito:** fonte municipale della scheda. Avvisi dichiarati; contratto attivo, preview e raccolta sono prerequisiti oltre a paginazione e documenti.
- **Conoscenza:** confermato per i prerequisiti e il comportamento del servizio. **Intervento:** percorso indipendente implementato e report pendente consegnato; accettazione non completata.
- **Osservazione:** lo scope esplicito `municipality` valuta solo questa fonte senza richiedere i tre prodotti regionali. Non riduce durata, cadenza, confronti con originali/allegati o prove di errore. I prerequisiti irrisolti impediscono l'apertura della campagna; la raccolta eventualmente già attiva non li annulla.
- **Evidenza:** [report e procedura](../../../operations/municipal-acceptance.md), [campagne](../../../../internal/backend/observation/store.go), [prove isolate](../../../../internal/backend/observation/observation_integration_test.go). Report completo, ricevuta del tentativo e stato dei controlli nell'archivio privato `ROLLOUT-GATES-20261004`.
- **Conseguenza:** il report mantiene lo stato pending e i controlli esistenti; non estende il perimetro né dichiara copertura dei cinque comuni.
- **Prossima verifica:** risolvere i prerequisiti dell'ambito, completare valutazione e osservazione, quindi accettare e abilitare separatamente la fonte verificata.
- **Collegamenti:** PON-001; task 11.5 della [specifica Toscana](../../../../openspec/changes/define-toscana-alert-service/tasks.md).
