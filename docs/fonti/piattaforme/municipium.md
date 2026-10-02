# Municipium — note di piattaforma

Ultima revisione documentale: 2026-10-02. Ambito verificato qui: collegamenti nella pagina comunale di [Livorno](../../territori/09-toscana/comuni/049009-livorno.md). Questa nota non certifica il comportamento o i permessi di altri comuni che usano la piattaforma.

## Caratteristiche osservate

La pagina di Livorno esaminata nella scheda comunale collega risorse PDF su `livorno-api.cloud.municipiumapp.it` e `cloud-ita.municipiumapp.it`, mentre la notizia è sul dominio del Comune. Distingue il PDF dell'ordinanza nella sezione Allegati dalla guida alla navigazione nel footer. Il riconoscimento dei collegamenti non verifica la disponibilità dei PDF o i loro diritti di riuso.

La notizia COC di Cascina esaminata in CAS-002 collega invece nel footer un Piano di miglioramento in PDF su `municipium-images-production-autopilot.s3.eu-south-1.amazonaws.com`. L'osservazione conferma un secondo caso di PDF generico esterno; non stabilisce una regola comune per tutti i siti Municipium.

## Regole del collector da considerare

Il [collector attuale](../../../internal/backend/acquisition/preview.go) identifica i link PDF sull'intera pagina e li considera obbligatori. Rifiuta una risorsa con schema o host diverso da quello del documento, marcandola `forbidden` prima del download. La [fixture degli allegati esterni](../../../internal/backend/acquisition/attachments_test.go) documenta la conservazione del riferimento e l'esito incompleto.

Per ogni comune verificare separatamente riferimento ufficiale, emittente/pubblicatore, domini, percorsi, allegati pertinenti, accesso e riuso. La presenza del nome Municipium non autorizza tutti i suoi host o tutti i loro contenuti. Eventuali eccezioni del collector devono avere un perimetro esplicito e testare il rifiuto delle risorse fuori da quel perimetro.

## Problemi e interventi proposti

La scansione generica può includere PDF di navigazione che non documentano la notizia; il vincolo sul dominio può escludere un vero allegato ufficialmente collegato. Vedere `LIV-001` e `LIV-002` per il caso concreto, i criteri di chiusura e la distinzione fra comportamento presente e correzione proposta. Non è stata applicata una modifica al collector con queste guide.

CAS-002 documenta lo stesso rischio di selezione del footer su Cascina. Nel percorso programmato l'errore dell'allegato ferma anche l'acquisizione dei documenti successivi: controllare originali conservati e completezza, oltre al numero dei target scoperti.

LIV-004 precisa il perimetro esterno proposto per Livorno e distingue il filtro locale del collector dall'accessibilità del file. Il collegamento nel footer a `cloud-ita.municipiumapp.it` non giustifica l'abilitazione di quel dominio come fonte di allegati pertinenti. LIV-005 verifica il download del singolo PDF con browser ordinario e `DirectHTTP`; la prima prova con un altro client non certificava indisponibilità per il trasporto applicativo. L'ammissione nel collector resta proposta.

## Registro delle scoperte

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
