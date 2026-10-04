# Consuntivo delle campagne Toscana

Il consuntivo del 4 ottobre 2026 chiude il task 9.4: due report separati delle
campagne esistenti riepilogano token, OCR, embedding, tentativi, retry e costi
stimati per fase e carico. Entrambe le campagne restano incomplete e il budget
operativo resta **non proponibile**. Una rendicontazione con lacune esplicite non
certifica la conclusione del trial o la completezza della fatturazione.

I dati e le ricevute sono privati, riferimento `TRIAL-COSTS-20261004`. Per ogni
campagna il consuntivo comprende:

- periodo di osservazione e data limite unica per i confronti;
- `bootstrap`, `ordinary`, `evaluation` e `reprocessing` separati;
- conteggi di run, tentativi e retry per fase/modello/provider;
- token input/output/cache disponibili e gruppi/tentativi senza misure;
- altre unità rendicontate, incluso il numero di richieste OCR/embedding;
- subtotali di costo stimato noto e dei retry, con valuta e prezzi attribuibili;
- hosting e storage per bootstrap e attività ordinaria, entrambi esplicitamente
  `unavailable` in assenza di fatture, allocazione e misure attribuibili;
- prezzi storici non verificati, metriche mancanti e `budget_ready: false`.

Le campagne si sovrappongono: sommarle conterebbe più volte gli stessi run.
Un token o costo omesso resta sconosciuto; un subtotale noto non diventa il costo
totale. L'inferenza locale senza addebito provider rendicontato conserva il costo
di hosting indisponibile. Gli embedding disabilitati non dimostrano il prezzo
o consumo di un futuro percorso attivo.

## Prezzi e validità temporale

Il [listino ufficiale Regolo](https://regolo.ai/pricing/), consultato il
4 ottobre 2026, espone prezzi EUR IVA esclusa:

| Modello | Unità | Prezzo corrente |
| --- | --- | ---: |
| Qwen3.8 27B | 1 milione token input / output | €0,50 / €2,10 |
| DeepSeek OCR 2 | Richiesta | €0,02 |
| Qwen3 Embedding 8B | Richiesta | €0,001 |

Il riscontro corrente concorda con le tariffe Qwen/OCR già registrate, ma non
verifica il piano dell'account o la decorrenza durante le campagne. Le ipotesi
storiche restano `verified_for_trial: false`: nessun prezzo precedente è stato
riscritto o retrodatato. L'entitlement del provider richiede una verifica distinta.

## Ripetizione e criteri per il budget

Leggere `GET /admin/observation-campaigns/{id}/cost-report?through=<RFC3339>`
dall'amministrazione privata. Riportare i gruppi `processing_breakdown`, le
`pricing_assumptions`, gli `infrastructure_inputs` e tutti i `missing_metrics`.
Registrare le misure mancanti con `POST .../{id}/cost-inputs`, stato
`unavailable`, motivazione, attore ed evidenza datata; successivi dati reali
aggiungono revisioni, senza sostituire lo storico. Conservare ricevute, report
JSON, sintesi leggibile e manifest SHA-256 nell'archivio privato.

Proporre il budget solo dopo osservazione completata alla stessa data limite,
prezzi validi per il periodo, contabilizzazione dei tentativi/costi e input
hosting/storage completi. Il servizio rifiuta una proposta incompleta. Il
[test isolato](../../internal/backend/trialcost/trialcost_integration_test.go)
verifica questo blocco insieme alla proiezione mensile delle sole attività
ordinarie e ai costi bootstrap/evaluation/reprocessing separati.
