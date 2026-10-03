# Municipium — note di piattaforma

Ultima revisione documentale: 2026-10-02. Ambito verificato qui: collegamenti nelle pagine comunali di [Livorno](../../territori/09-toscana/comuni/049009-livorno.md) e [Cascina](../../territori/09-toscana/comuni/050008-cascina.md). Questa nota non certifica il comportamento o i permessi di altri comuni che usano la piattaforma.

## Caratteristiche osservate

La pagina di Livorno esaminata nella scheda comunale collega risorse PDF su `livorno-api.cloud.municipiumapp.it` e `cloud-ita.municipiumapp.it`, mentre la notizia è sul dominio del Comune. Distingue il PDF dell'ordinanza nella sezione Allegati dalla guida alla navigazione nel footer. Il riconoscimento dei collegamenti non verifica la disponibilità dei PDF o i loro diritti di riuso.

La notizia COC di Cascina esaminata in CAS-002 collega invece nel footer un Piano di miglioramento in PDF su `municipium-images-production-autopilot.s3.eu-south-1.amazonaws.com`. L'osservazione conferma un secondo caso di PDF generico esterno; non stabilisce una regola comune per tutti i siti Municipium.

## Regole del collector da considerare

Il [collector](../../../internal/backend/acquisition/preview.go) mantiene per le fonti prive di policy `attachments` la discovery generale dei PDF e il rifiuto delle risorse esterne prima del download. La policy revisionata permette di selezionare una classe del contenuto e autorizzare solo origini HTTPS e percorsi con referral ufficiale e riuso verificato; non abilita nuovi canali di discovery. `validate_pdf` controlla tipo, struttura e leggibilità con Poppler. Preview e percorso programmato applicano gli stessi controlli alle fonti con la policy. Vedere le [fixture storiche](../../../internal/backend/acquisition/attachments_test.go), i [test del perimetro](../../../internal/backend/acquisition/scoped_attachments_test.go) e la [verifica di persistenza](../../../internal/backend/acquisition/scoped_attachments_integration_test.go).

Per ogni comune verificare separatamente riferimento ufficiale, emittente/pubblicatore, domini, percorsi, allegati pertinenti, accesso e riuso. La presenza del nome Municipium non autorizza tutti i suoi host o tutti i loro contenuti. Eventuali eccezioni del collector devono avere un perimetro esplicito e testare il rifiuto delle risorse fuori da quel perimetro.

Il replay interpretativo di Cascina e Livorno del 2026-10-02 ha usato
`main#main-content` per conservare titolo, corpo e metadati controllati. La classe
scelta per scoprire allegati non dimostra da sola copertura di tutta la notizia.
MUN-005 precisa il campione e la distinzione fra trial e policy continuativa.

## Problemi e interventi proposti

La scansione generica può includere PDF di navigazione che non documentano la notizia; il vincolo sul dominio può escludere un vero allegato ufficialmente collegato. Vedere `LIV-001` e `LIV-002` per il caso concreto, i criteri di chiusura e la distinzione fra comportamento presente e correzione proposta. La correzione del collector è successiva alle guide iniziali: LIV-006 e MUN-003 documentano il comportamento opt-in implementato.

CAS-002 documenta la diagnosi precedente dello stesso rischio di selezione del footer su Cascina; CAS-004 e MUN-004 registrano la successiva correzione separata. Nel percorso programmato l'errore dell'allegato ferma anche l'acquisizione dei documenti successivi: controllare originali conservati e completezza, oltre al numero dei target scoperti.

LIV-004 precisa il perimetro esterno proposto per Livorno e distingue il filtro locale del collector dall'accessibilità del file. Il collector ora offre una regola esplicita per fonte/revisione; l'ammissione tecnica non garantisce che il download sia consentito dal server. Il collegamento nel footer a `cloud-ita.municipiumapp.it` non giustifica l'abilitazione di quel dominio come fonte di allegati pertinenti.

LIV-005 aggiorna la verifica di accesso per il singolo PDF di Livorno: browser ordinario e `DirectHTTP` hanno acquisito il file. La prima prova con un diverso client non certificava indisponibilità per il trasporto applicativo; la causa della differenza rimane da determinare. Non estendere l'esito ad altri file, host o comuni, né considerarlo una modifica del collector.

## Registro delle scoperte

MUN-001 e MUN-002 registrano il comportamento osservato prima della correzione opt-in di MUN-003. Cascina e gli altri perimetri non ricevono automaticamente la policy di Livorno.

### MUN-001 — Collegamenti esterni e PDF generici nella stessa pagina

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, documentale su pagina comunale e codice.
- **Ambito:** pagina Livorno collegata nella scheda; estensione ad altri comuni da verificare.
- **Conoscenza:** confermato per il caso esaminato. **Intervento:** modifiche proposte, non applicate.
- **Osservazione:** allegato della notizia e documento generico del footer sono entrambi PDF collegati a host esterni.
- **Evidenza:** [caso e riferimento ufficiale](../../territori/09-toscana/comuni/049009-livorno.md), codice e fixture collegati sopra.
- **Conseguenza:** verificare posizione e funzione dei PDF prima di definirli dipendenze necessarie della notizia.
- **Prossima verifica:** valutare selezione degli allegati e risorse esterne ammesse su fixture e fonte specifica; ogni altro comune richiede la propria evidenza.

### MUN-002 — Il caso del footer è osservato anche su Cascina

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, pagina ufficiale e codice.
- **Ambito:** sola notizia COC di Cascina collegata in CAS-002; nessuna estensione dei domini ammessi.
- **Conoscenza:** confermato. **Intervento:** selezione degli allegati pertinenti proposta, non applicata.
- **Osservazione:** il footer collega il PDF Piano di miglioramento su un host S3 esterno; il collector lo considera obbligatorio e interrompe il controllo programmato dopo il rifiuto.
- **Evidenza:** [CAS-002 e pagina ufficiale](../../territori/09-toscana/comuni/050008-cascina.md), [collector](../../../internal/backend/acquisition/preview.go).
- **Conseguenza:** il problema di selezione non riguarda soltanto la pagina Livorno di MUN-001; i risultati per ambiente restano privati.
- **Prossima verifica:** distinguere PDF generici e allegati pertinenti su fixture e sui due perimetri specifici, mantenendo espliciti i veri allegati mancanti.


### MUN-003 — Policy esplicita per gli allegati, verificata sul perimetro di Livorno

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, codice, fixture redistribuibili, parser applicativo e verifica operativa privata del perimetro Livorno.
- **Ambito:** capacità opt-in per fonti municipali crawl-based; Livorno e solo gli allegati pertinenti del percorso comunale verificato. Nessuna regola trasferita a Cascina o agli altri comuni.
- **Conoscenza:** confermato. **Intervento:** implementato e verificato sul perimetro Livorno; risultati dettagliati per ambiente restano privati.
- **Osservazione:** ogni revisione può limitare la discovery dei PDF al contenitore del documento e autorizzare origini/percorso con referral e riuso espliciti. I PDF del footer sono esclusi dal filtro; i riferimenti necessari non recuperati restano visibili e impediscono un controllo completo.
- **Evidenza:** [LIV-006](../../territori/09-toscana/comuni/049009-livorno.md), [policy](../../../internal/backend/registry/attachments.go), [fixture](../../../internal/backend/acquisition/scoped_attachments_test.go) e [procedura](../../operations/municipal-attachments.md).
- **Conseguenza:** aggiorna l'intervento proposto in MUN-001 per Livorno; MUN-002 descriveva il perimetro di Cascina allora privo della policy; MUN-004 documenta la successiva verifica indipendente. La capacità non autorizza automaticamente domini Municipium o host S3 di altri enti.
- **Prossima verifica:** raccogliere evidenze e configurare separatamente ogni ulteriore fonte; osservare continuità del download e cambiamenti nella struttura del sito.

### MUN-004 — Cascina richiede contenitore e tenant verificati separatamente

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, pagine ufficiali di Cascina, trasporto applicativo, parser PDF e fixture.
- **Ambito:** pagine configurate di `cascina-municipal`, classe `page-content` e soli PDF municipali collegati su `cascina-api.cloud.municipiumapp.it` nella directory `/s3/1520/allegati/`.
- **Conoscenza:** confermato per i casi esaminati. **Intervento:** revisione separata implementata e verificata in sviluppo; esiti e configurazioni dettagliati privati.
- **Osservazione:** contenuto e sezioni degli allegati usano `page-content`, diversamente da `article-content` di Livorno. Il referral Sogefarm collega atti municipali nel tenant API Cascina; le note legali e le eccezioni sono state riesaminate. Il PDF generico del footer appartiene a un'origine S3 diversa, non autorizzata. Il normale client applicativo scarica i due PDF esaminati e Poppler li legge; una precedente prova con altro client era fallita.
- **Evidenza:** [CAS-004, referral e condizioni](../../territori/09-toscana/comuni/050008-cascina.md), [policy](../../../internal/backend/registry/attachments.go), [fixture](../../../internal/backend/acquisition/scoped_attachments_test.go), [verifica operativa](../../operations/municipal-attachments.md).
- **Conseguenza:** supera per il perimetro revisionato l'intervento proposto in MUN-002; non trasferisce il permesso di Livorno, né autorizza tutti gli host Municipium. Non verifica formati diversi dai PDF o tutti i documenti del tenant.
- **Prossima verifica:** osservare struttura, download, condizioni ed eccezioni documentali; verificare indipendentemente ogni ulteriore comune o directory.


### MUN-005 — Il perimetro interpretativo comprende titolo e metadati

- **Data:** 2026-10-02. **Ultima verifica:** 2026-10-02, originali conservati di Cascina/Livorno e trial limitato in development.
- **Ambito:** solo i due comuni esaminati; selettore `main#main-content`, distinto dalle classi per la discovery PDF.
- **Conoscenza:** confermato nel campione. **Intervento:** replay e confronti con modello applicati; nessuna policy comune attivata.
- **Osservazione:** il contenitore principale conserva titolo, corpo di `article-content` e metadati controllati, riducendo il testo di navigazione/footer. I quattro confronti municipalità baseline/candidato concordano sulla pertinenza; allegati e originali rimangono separati e completi quando disponibili.
- **Evidenza:** [CAS-006](../../territori/09-toscana/comuni/050008-cascina.md), [LIV-007](../../territori/09-toscana/comuni/049009-livorno.md), [trial e limiti](../../operations/local-processing.md). Evidenza dettagliata privata.
- **Conseguenza:** verificare contenitore e layout per ogni comune. PDF tecnicamente estraibili non stabiliscono un permesso di attivare un’intera directory di documenti come solo testo.
- **Prossima verifica:** casi significativi, cambiamenti del layout e qualità downstream prima del rollout della singola fonte.
