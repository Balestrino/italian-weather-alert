# Collaudo indipendente dei comuni Toscana

Verifica del 4 ottobre 2026, task 11.5 della
[specifica Toscana](../../openspec/changes/define-toscana-alert-service/tasks.md).
I quattro comuni successivi al pilota hanno un report individuale **pending**.
Il percorso di collaudo è implementato e verificato; nessun comune aggiuntivo è
qui dichiarato accettato o coperto integralmente.

| Comune e fonte | Ambito documentato | Esito e condizioni da chiudere |
| --- | --- | --- |
| [Livorno](../territori/09-toscana/comuni/049009-livorno.md), `livorno-municipal` | Categoria Protezione Civile e allegati nel perimetro revisionato | Pending: paginazione, portali degli atti e perimetro dei documenti; valutazione e osservazione complete |
| [Pisa](../territori/09-toscana/comuni/050026-pisa.md), `pisa-municipal` | Notizie | Pending: altri canali, atti/allegati e paginazione; valutazione e osservazione complete |
| [Pontedera](../territori/09-toscana/comuni/050029-pontedera.md), `pontedera-municipal` | Avvisi, pagine esplicitamente dichiarate | Pending: contratto attivo e preview, paginazione e documenti; valutazione e osservazione complete |
| [Cascina](../territori/09-toscana/comuni/050008-cascina.md), `cascina-municipal` | Avvisi e categoria storica dichiarata | Pending: continuità della sezione, atti, allegati e paginazione; valutazione e osservazione complete |

I report completi conservano identità della fonte, revisione, sezioni, limiti,
condizioni irrisolte, stato di raccolta, valutazioni, campagna e controlli di
accettazione/pubblicazione. Restano nell'archivio operativo privato con riferimento
`ROLLOUT-GATES-20261004`; il confronto prima/dopo verifica che i controlli delle
fonti non sono cambiati. I tentativi di apertura delle quattro campagne sono
stati rifiutati per prerequisiti incompleti, senza creare campagne fittizie.

## Avvio e raccolta delle evidenze

L'API privata `POST /admin/observation-campaigns` accetta `scope: "municipality"`
con **una sola fonte municipale** di Calcinaia, Livorno, Pisa, Pontedera o Cascina.
Un payload di esempio, da compilare con una data reale e una fonte pronta:

```json
{
  "id": "pisa-trial-<data>",
  "actor": "<operatore>",
  "scope": "municipality",
  "started_at": "<inizio-reale-RFC3339>",
  "source_ids": ["pisa-municipal"]
}
```

La fonte deve avere contratto di raccolta/conservazione e provenienza verificati,
sezioni dichiarate, revisione attiva dopo preview e raccolta abilitata, senza
prerequisiti irrisolti. Le condizioni pendenti vanno risolte con evidenze o con
una revisione esplicita e veritiera dell'ambito prima di iniziare la campagna.

Il campo omesso conserva `scope: "mvp"`: servono ancora vigilanza, criticità,
monitoraggio e Calcinaia per completare il MVP. La migrazione aggiuntiva assegna
questo valore alle campagne precedenti senza riscrivere valutazioni o checksum.
Lo scope municipale esclude solo la richiesta dei tre prodotti regionali dal
report di quel comune; mantiene tutti gli altri criteri.

Per ogni comune conservare sette intervalli completi di 24 ore, controlli delle
sezioni e ritardi misurati, confronti attribuibili con originali e allegati,
errori osservati o casi conservati, eventi significativi assenti e valutazioni
delle omissioni, affermazioni non supportate e campi indeterminati. Comprendere
aggiornamenti, cancellazioni, riaperture parziali, date conflittuali, scansioni,
contenuti irrilevanti, storico ed equivalenza API/MCP. Usare i report e le
revisioni attribuibili dell'API privata; nessun campo permette al chiamante di
dichiarare direttamente il completamento.

`running` indica durata insufficiente; `extended` conserva ritardi, confronti
mancanti e revisioni fallite/irrisolte. `complete` richiede tutte le evidenze.
Valutazione, campagna, accettazione e abilitazione pubblica sono passaggi distinti
per ciascuna fonte. Il successo di un comune non completa altri comuni o il MVP;
la copertura completa dei cinque comuni resta pendente nel
[coverage tracker](../coverage.md#region-09).

## Verifica del software

Fixture sintetiche con PostgreSQL isolato verificano i quattro comuni aggiuntivi,
esclusione di un comune non pianificato, scope misti/multipli rifiutati, campagna
incompleta, confronto dell'originale obbligatorio, allegato necessario non
sostituibile con `not_applicable`, durata minima e completamento indipendente.
L'upgrade conserva le campagne MVP e i vincoli di immutabilità. I test di registro
verificano separatamente regressioni, accettazione e abilitazione; i controlli
API verificano il passaggio esplicito dello scope e rifiutano un verdetto fornito
dal chiamante. Queste prove non simulano giorni di osservazione reali.

```sh
go test -race -tags=integration ./internal/backend/observation ./internal/backend/registry
go test -race ./internal/backoffice
```
