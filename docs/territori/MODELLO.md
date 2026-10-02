# Modello di scheda territoriale

Copiare la struttura seguente nel percorso descritto nella [guida](README.md). Sostituire i segnaposto; lasciare espliciti gli elementi non ancora verificati. Per una regione omettere `codice_istat`, elencare separatamente i prodotti regionali e aggiungere i collegamenti alle schede comunali.

```yaml
---
tipo: "comune"
codice_regione: "09"
codice_istat: "000000"
ultima_revisione: "AAAA-MM-GG"
---
```

Il blocco YAML va all'inizio della nuova scheda, senza delimitatori di codice. `ultima_revisione` riguarda il documento; le verifiche delle singole scoperte hanno date proprie.

## Identità e ambito

- Nome, regione, provincia e codice ISTAT ove applicabile.
- Identificativi delle fonti: da verificare nell'ambiente interessato.
- Ambito documentato e limiti conosciuti.
- Collegamenti al coverage tracker, alla regione e alle piattaforme pertinenti.

## Mappa delle fonti

| Fonte o prodotto | Ruolo | URL ufficiale | Sezioni e limiti | Evidenza e data |
| --- | --- | --- | --- | --- |
| Da compilare | Primaria / secondaria / candidata | Da verificare | Da verificare | Non disponibile |

Registrare separatamente emittente, pubblicatore e piattaforma quando differiscono. Per un canale esterno indicare il riferimento ufficiale, il suo ambito e le verifiche di accesso/riuso ancora necessarie.

## Guida operativa

Descrivere discovery, paginazione, date, allegati, dominio/percorso, frequenza documentata, riconoscimento del contenuto e gestione dei fallimenti. Collegare le procedure comuni. Indicare ciò che resta da verificare senza presumere completezza o attivazione.

## Regole ed eccezioni

| ID scoperta | Ambito preciso | Regola corrente | Stato dell'intervento | Riferimento |
| --- | --- | --- | --- | --- |
| PREFISSO-001 | Fonte/prodotto e percorso | Da verificare | Proposto | Non disponibile |

Separare le regole effettivamente applicate dalle modifiche desiderate. Richiamare le note della piattaforma senza ereditare automaticamente autorizzazioni o eccezioni.

## Problemi aperti

| ID scoperta | Sintomo o lacuna | Causa e stato della conoscenza | Prossima verifica | Criterio di chiusura |
| --- | --- | --- | --- | --- |
| PREFISSO-001 | Da compilare | Ipotesi / confermato | Da compilare | Evidenza necessaria |

## Registro delle scoperte

### PREFISSO-001 — Titolo della scoperta

- **Data della scoperta:** AAAA-MM-GG, oppure data originaria non disponibile.
- **Ultima verifica:** AAAA-MM-GG; precisare se documentale, con fixture o operativa privata.
- **Ambito:** territorio, fonte/prodotto, sezione o dominio; ambiente/revisione se pertinenti.
- **Conoscenza:** ipotesi / confermato / superato.
- **Intervento:** proposto / applicato / verificato / non necessario.
- **Osservazione:** fatto constatato, separato dall'interpretazione.
- **Evidenza:** link ufficiali e riferimenti pubblicabili; dettagli riservati nel registro privato.
- **Conseguenza:** effetto su discovery, raccolta, interpretazione o copertura.
- **Prossima verifica:** controllo necessario e criterio di chiusura.
- **Collegamenti:** issue, modifica/configurazione, test e precedenti ID; non disponibile quando assenti.

Una rivalutazione aggiunge una nuova voce collegata all'ID originario, conservandone il testo. Non descrivere una proposta o un test sintetico come una correzione verificata sulla fonte reale.
