# Matrice di copertura di Calcinaia

Audit e adozione del 4 ottobre 2026, ISTAT `050004`, task 31.1 e 31.3 del
[servizio Toscana](../../openspec/changes/define-toscana-alert-service/tasks.md).
La matrice distingue discovery, acquisizione, interpretazione, confronto e
accettazione. Il [coverage tracker](../coverage.md#region-09) conserva lo stato
pubblico: fonte municipale e prodotti CFR ancora `pending`.

## Perimetro e metodo

La verifica legge la configurazione e gli intervalli effettivi del development,
attraversa le tre sezioni dichiarate e confronta separatamente l'elenco generale
`/novita`. La finestra recente usa le **date delle schede nell'elenco**, dal
4 settembre al 4 ottobre 2026; non rappresenta la validità dei provvedimenti.
Sono distinti gli URL scoperti, quelli con un originale conservato e quelli con
un'interpretazione disponibile nelle interfacce.

Le quattro pagine di ricerca documentale API appartengono alla stessa vista
salvata. Il confronto con PostgreSQL legge versioni, acquisizioni, risorse e
tentativi; non scarica nuovamente gli oggetti dallo storage e non ne certifica
l'integrità. Le verifiche API/MCP già concluse nella
[campagna osservativa](observational-trial.md) restano evidenze precedenti,
senza presentarle come nuovi confronti MCP di questo audit.

## Matrice corrente

| Canale o ambito | Osservazione confermata | Limite e criterio di chiusura |
| --- | --- | --- |
| [Notizie](https://www.comune.calcinaia.pi.it/tipi-di-notizia/notizie) | Dieci pagine attraversate; 234 URL di avvisi | Il confronto copre gli elenchi osservati, non tutte le pubblicazioni del Comune |
| [Avvisi](https://www.comune.calcinaia.pi.it/tipi-di-notizia/avvisi) | Tre pagine attraversate; 49 URL | Conservare controlli indipendenti della sezione e paginazione |
| [Comunicati](https://www.comune.calcinaia.pi.it/tipi-di-notizia/comunicati) | Una pagina attraversata; tre URL | Un elenco breve non dimostra assenza di altri canali |
| [Elenco generale](https://www.comune.calcinaia.pi.it/novita) | 91 pagine e 1.085 URL; 799 non presenti nelle tre sezioni. Nessuno dei 799 ha una data di elenco nella finestra recente; nessuna data di scheda mancante nel confronto | Discovery storica più ampia distinta dal recupero iniziale di 30 giorni. Non inferire copertura storica completa |
| Discovery e originali municipali | I 286 URL delle tre sezioni sono tutti nel tracker. I 57 avvisi della finestra recente hanno un originale completo. Gli otto URL senza originale conservato sono datati maggio–luglio, fuori dalla finestra iniziale e senza eccezioni di riferimento esplicito o misura protetta nel tracker | L'assenza degli otto originali non è un'omissione dimostrata nella finestra recente. Nuovi riferimenti espliciti o misure irrisolte richiedono recupero indipendente |
| Allegati municipali | Nessun riferimento obbligatorio mancante nelle ultime versioni ispezionate; 64 riferimenti ad allegati censiti | Completezza delle risorse registrate distinta dalla scoperta di atti non collegati, dal readback degli oggetti e dall'accuratezza OCR delle ordinanze scansionate |
| Interpretazione della finestra recente | Nella stessa vista API, 43 dei 57 avvisi sono `supported`; 14 sono `not_processed`. Per questi ultimi, tredici job di classificazione sono falliti per citazione non valida e uno conserva un risultato indeterminato per contenuto richiesto incompleto | Correggere la causa dai request/response conservati, senza accettare parafrasi o marcare a mano le notizie come irrilevanti. Verificare nuove esecuzioni ordinarie e proiezioni. Un job riuscito non basta; i casi 14/14 già revisionati non certificano tutto il corpus |
| [Cittadino Informato](https://cittadinoinformato.it/calcinaia/) | Referral municipale e identità API nuovamente osservati. L'elenco API delimitato dalla data di visualizzazione dal 4 settembre restituisce 36 comunicazioni su una pagina. Il successivo seguito 31.3 registra e abilita una sorgente distinta in development: preview e due controlli ordinari completi, 16 PDF necessari riletti con hash, raccolta 600/1.800 secondi e confronti ogni minuto | Le prime 50 ricevute restano diagnostiche: 24 riscontri mancanti e 26 non comparabilità. Errori di citazione dell'interpretazione ordinaria rimangono rifiutati; atti/edizioni e metadati CFR mancanti non sono inferiti. Il filtro di visualizzazione non prova copertura per data di pubblicazione o validità. La motivazione storica di esclusione per termini di riuso nella configurazione primaria è superata dal piano 26.2, ma non da un'attivazione automatica |
| Albo e atti richiamati | Il sito comunale espone un referral all'albo; non risulta una raccolta generale configurata | Recuperare gli atti esplicitamente citati quando necessari a misura, luogo o validità. La specifica non richiede l'acquisizione indiscriminata dell'albo |
| Date e aggiornamenti | La pagina sulle disposizioni meteo conserva ancora il conflitto settembre/agosto nel testo ufficiale | Cercare l'atto richiamato o una rettifica riferita alla stessa misura. Se nessuna evidenza chiarisce il campo, mantenerlo indeterminato; le date di pagina o del CFR non possono correggerlo |
| Prodotti CFR per Calcinaia/A4 | Pipeline e casi conservati verificati nei rispettivi perimetri; dossier di accettazione revisionati | Casi positivi reali di vigilanza e bollettino operativo reale di monitoraggio restano condizioni dei dossier. Una nuova verifica degli elenchi comunali non le chiude |

Le date degli elenchi e i conteggi sono osservazioni di questo audit. Le
classificazioni `supported` indicano un risultato validato disponibile, non una
nuova revisione semantica di ciascun documento o un'accettazione della fonte.
L'ispezione runtime/readiness passa; produzione e abilitazione pubblica ordinaria
restano ferme. Cadenza e soglia di ritardo vanno lette dagli **intervalli effettivi**,
non soltanto dai valori iniziali della configurazione.

## Lavoro successivo e prove richieste

1. **Interpretazioni pendenti — 31.2.** Conservare l'elenco esatto delle versioni
   e i tentativi falliti. Separare cambi di presentazione, citazioni inventate,
   contenuti OCR incompleti e input equivalenti. Applicare soltanto correzioni
   verificate con casi positivi/negativi redistribuibili; rielaborare in modo
   delimitato. Controllare pertinenza, evidenze e risultati pubblici ordinari,
   mantenendo precedenti fallimenti e confini storici.
2. **Canale aggiuntivo continuativo — 31.3 conclusa.** Il perimetro Calcinaia
   separato dalla primaria è attivo nel development ordinario dopo preview,
   dipendenze e due controlli riusciti. Continuare a verificare controparti,
   versioni, ricevute e interpretazioni secondo la [procedura](cittadino-informato.md#adozione-ordinaria-verificata--4-ottobre-2026). Un passaggio più stretto che evita un PDF necessario non
   chiude il perimetro fallito. Nessun candidato senza primaria diventa misura
   confermata; indisponibilità, non comparabilità e conflitto restano distinti.
3. **Atti e tempi — 31.4.** Censire riferimenti espliciti degli avvisi pertinenti,
   recuperare le risorse necessarie entro un perimetro revisionato e verificare
   lo stesso atto/versione. Conservare assenze, dinieghi e conflitti non chiariti;
   non dedurre revoche o validità attuale dall'età del documento.
4. **Accettazione — 31.5.** Rivalutare i dossier sul perimetro effettivamente
   collaudato, con le nuove esecuzioni e gli originali. Registrare separatamente
   accettazione e revisione del relativo report; se restano limitazioni,
   dichiarare copertura parziale. Il successo di Calcinaia non accetta i CFR.

I sei task infrastrutturali/di rilascio già aperti restano nello
[stato Toscana](toscana-readiness.md). L'operatore ha richiesto in questa sessione
di **rinviare backup e ripristino**: la richiesta non li completa, non li elimina
dai requisiti di produzione e non autorizza a dichiarare raggiunti RPO/RTO.
La preparazione di un rilascio deve anche verificare capacità disponibile,
telemetria, percorso SMTP, staging e digest approvato prima dell'avvio.

## Evidenze e manutenzione

Configurazione live, inventari, HTML degli elenchi, risposte API, errori,
diagnostica e manifest di integrità restano nell'archivio privato territoriale,
riferimento `CLN-COVERAGE-20261004`. La presente matrice pubblica conclusioni e
limiti, senza copie dei contenuti o dettagli di accesso. Nessuna configurazione,
fonte, job o stato di accettazione è stato modificato dall'audit 31.1. Il seguito
31.3 attiva soltanto la sorgente piattaforma selezionata nel development;
configurazione e verifiche restano private in `CIN-CONTINUOUS-20261004`.

Ripetere il confronto dopo cambi di sezioni/configurazione o nuove evidenze.
Registrare data, ambiente, revisione attiva, intervalli effettivi, inizio/fine
dell'osservazione, finestre di pubblicazione e visualizzazione, pagine percorse,
confronti con originali/allegati, esiti delle interpretazioni e lacune ancora
aperte. Conservare l'audit precedente; aggiornare guida, registro datato,
specifica/checklist e coverage soltanto per ciò che le nuove prove stabiliscono.
