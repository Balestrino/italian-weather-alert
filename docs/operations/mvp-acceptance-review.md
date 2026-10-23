# Revisione dei quattro ambiti MVP — 4 ottobre 2026

Il task 10.1 è concluso: sono stati revisionati quattro dossier separati,
con riferimenti alle evidenze e limiti per ciascuna fonte, e verificati i controlli
di accettazione, abilitazione indipendente e copertura parziale. L'esito di
accettazione resta `pending` per tutti e quattro gli ambiti. La revisione dei
rapporti non registra un'accettazione positiva, non abilita le fonti e non attesta
la readiness di produzione. Il [coverage tracker](../coverage.md#region-09)
conserva lo stato pubblico; il [rapporto osservativo](observational-trial.md)
documenta separatamente la campagna completa del task 9.3.

| Ambito revisionato | Evidenza e perimetro | Limiti conservati |
| --- | --- | --- |
| Vigilanza CFR | Originali conservati e confronto delegato dei 364 campi; quantità di pioggia, simboli di temporale e applicabilità; sette filtri di rischio API/MCP | Casi positivi reali di vento, mareggiate, neve e ghiaccio ancora da confrontare; i controlli sintetici non ne attestano l'accettazione semantica |
| Criticità/allerta CFR | Originali conservati: verde/non applicabile e caso storico giallo idrogeologico/temporali; livelli, zone, date e sette filtri API/MCP | Edizioni vettoriali riconosciute; altre combinazioni positive e geometrie verificate con fixture; formati non supportati mantengono fallback e incertezza |
| Monitoraggio CFR | Canale ordinario senza evento, controlli persistiti della campagna e confini rispetto agli altri prodotti | Il caso con evento è un esempio normativo conservato: manca il confronto di un bollettino operativo reale con parser e API/MCP; nessun nuovo colore dedotto |
| Calcinaia | Tre sezioni dichiarate, periodo della campagna, allegati, scoperta/paginazione e contratto invariato di quattordici confronti; riapertura, chiusure, quattro divieti, COC ed eccezione | Verifiche precedenti riusate dichiarate; nessuna completezza dell'intero Comune o dei canali aggiuntivi; conflitti di data e validità indeterminata conservati |

## Verifiche delle interfacce, allegati e storico

Ventidue confronti effettivi tra HTTP e client SDK MCP passano sulle stesse
viste del development. Coprono i cinque gruppi di operazioni: ricerca dei Comuni,
situazione, ricerca dei documenti/fatti, dettaglio documentale e copertura delle
fonti. Comprendono i sette filtri di rischio regionali, copertura e ricerca
storica per ciascuna fonte e isolamento delle route amministrative. Il dettaglio
conserva versioni ed evidenze; le richieste anteriori alla conoscenza del servizio
restituiscono il limite dell'archivio, senza inventare uno storico precedente.
I confronti già conservati di paginazione, viste precedenti e confini di conoscenza
sono attribuiti come verifiche riusate.

La scansione dell'ordinanza è una simulazione dichiarata, con corrispondenza fra
pagina originale e raster sintetico. Test OCR verificano immagini per pagina,
citazioni, risorse mancanti/illeggibili e persistenza della provenienza su
PostgreSQL isolato. La precedente esecuzione remota su due pagine di un allegato
comunale verifica il provider e il runtime: quell'allegato non è un'ordinanza
meteo e non dimostra l'accuratezza su tutte le scansioni. I dossier mantengono
queste distinzioni anziché presentare fixture o canary come nuovi casi operativi.

I fallimenti comunali reali e i risultati precedenti restano conservati. I guasti
regionali assenti dalla campagna sono esercitati da controlli sintetici attribuiti
su parsing, risposta 429, retry/backoff, lease scadute e PDF mancanti. Un test con
client MCP effettivo verifica inoltre che guasti nella creazione o lettura di una
vista producano `service_unavailable`, come HTTP 503, senza dati di assenza di
allerte. Questi controlli non comportano interruzioni delle dipendenze attive.

## Gate indipendenti e risultato della revisione

Un nuovo test con quattro fonti sintetiche e PostgreSQL temporaneo verifica:

- rifiuto dell'abilitazione senza accettazione e revisione del relativo report;
- rifiuto di evidenze mancanti per sette rischi, scansioni, storico, errori o
  equivalenza delle interfacce;
- abilitazione esplicita di una fonte alla volta, senza attivazione implicita
  durante la revisione e senza cambiamenti alle altre fonti;
- stato `partial` con uno, due o tre ambiti accettati, e anche con quattro ambiti
  se uno conserva limitazioni di copertura;
- stato `ready` solo con tutti e quattro gli ambiti pienamente accettati;
- sospensione della pubblicazione di una fonte preservando raccolta e altre fonti.

Le fonti reali conservano raccolta attiva, accettazione pendente e abilitazione
pubblica ordinaria disattivata. Le scelte separate di consultazione in development
non cambiano questo esito. La readiness MVP reale resta `pending`.

Prima di un'accettazione positiva occorre chiudere le condizioni del singolo
dossier, dichiarare il perimetro e tutte le limitazioni e registrare accettazione
e relativa revisione. Per la vigilanza rimangono i casi positivi reali indicati;
per il monitoraggio rimane il bollettino operativo. Eventuali ambiti accettati con
limitazioni conservano copertura parziale. PBS, capacità, SMTP, staging e rilascio
restano gate operativi separati, riportati nello [stato Toscana](toscana-readiness.md).

Originali, risposte, quattro dossier con manifest di integrità e log dei controlli
restano nel registro operativo privato `TOSCANA-ACCEPTANCE-101-20261004`.
Questa sintesi pubblica non redistribuisce catture o dettagli dell'ambiente.

## Rivalutazione del 23 ottobre 2026 (task 31.5)

L'operatore ha scelto di rivalutare la vigilanza CFR, limitatamente ai casi reali disponibili, e Calcinaia. Criticità e monitoraggio non sono stati rivalutati in questa sessione. Per tutti e quattro gli ambiti l'esito resta `pending` e il registro non contiene alcuna accettazione.

| Ambito | Verifica | Esito |
| --- | --- | --- |
| Vigilanza CFR | Edizioni del 20 e 21 ottobre con vento e mare reali. Le 416 voci zona/giorno di vento, mare, neve e ghiaccio coincidono con le tabelle degli originali: le zone elencate risultano «in valutazione», senza intensità dedotte, e le altre non rappresentate. API e MCP coincidono | Pending: mancano casi reali di neve e ghiaccio |
| Calcinaia | Stesso contratto di quattordici controlli sulla configurazione corrente (estrazione v29, modello locale), registrato nel registro di development. Classificazione 7 su 7 e linking 2 su 2 superati; discovery e mappatura OCR riusate con provenienza dichiarata | Pending: omissioni nell'estrazione dell'aggiornamento con strade liberate e della riapertura parziale |

Per la riapertura parziale, l'elaborazione ordinaria dello stesso giorno, con la stessa configurazione, estrae le misure attese con un divieto aggregato in più. L'esito del modello locale quindi non è riproducibile. Finché la suite non passa, i recuperi d'archivio della fonte restano bloccati e la regressione fallita impedisce l'accettazione.

Altri limiti di Calcinaia:
- I limiti di provenienza nella configurazione attiva della fonte sono superati: indicano ancora Cittadino Informato e l'albo come esclusi.
- Lo stesso vale per una revisione della configurazione, che richiederebbe una nuova accettazione.

Per la criticità esistono ora edizioni reali non verdi: arancione per temporali il 7 e il 20 ottobre; gialle per vento, mareggiate e reticolo principale. Per il monitoraggio esistono bollettini con evento del 5, 7–8 e 20 ottobre, proiettati come narrativa parziale. Entrambi restano da confrontare in una sessione dedicata.

Evidenze private: `ACCEPTANCE-31-5-20261023`.

