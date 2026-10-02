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
