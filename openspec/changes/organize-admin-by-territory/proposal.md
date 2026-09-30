# Proposal

## Why

Il pannello separa fonti, documenti, elaborazioni e risultati in strumenti globali: per capire cosa accade in un territorio occorre ricostruirne ogni volta il contesto. L'operatore richiede un percorso Regioni → Regione → Comune, comprendendo anche configurazione e abilitazione di nuove regioni.

## What Changes

- Rendere Regioni la pagina iniziale: stato di abilitazione, configurazione, copertura, ultimi risultati e problemi per territorio.
- Introdurre un catalogo regionale e associazioni territoriali esplicite, con configurazioni versionate, importazione verificabile dell'anagrafica completa dei comuni e abilitazione/disabilitazione tracciata.
- Offrire pagine regionali con Risultati, Comuni, Configurazione e Storico; pagine comunali con Dati, Configurazione e Storico. Includere comuni privi di fonti e distinguere dati mancanti, configurazione incompleta e servizio disabilitato.
- Consentire l'onboarding di nuove regioni e delle loro fonti compatibili, senza applicare implicitamente parser o regole della Toscana a territori diversi.
- Contestualizzare documenti, job e configurazioni esistenti; raccogliere gli strumenti trasversali in Operazioni e Sistema.
- Separare abilitazione territoriale, raccolta, interpretazione e pubblicazione delle fonti; preservare evidenze, revisioni, protezioni amministrative e compatibilità dei client esistenti.

## Capabilities

### New Capabilities

- `territorial-administration`: catalogo e ciclo di vita delle regioni, anagrafiche e associazioni versionate, configurazione e controlli territoriali.
- `territorial-admin-navigation`: navigazione gerarchica, risultati territoriali, elenchi completi e storico consultabile.

### Modified Capabilities

Nessuna specifica principale presente (`openspec list --specs`: nessun risultato). Le nuove capacità estendono le proposte esistenti `modernize-admin-dashboard` e `define-toscana-alert-service`, mantenendone i vincoli di accessibilità, provenienza, incertezza e compatibilità. L'ampliamento amministrativo oltre la Toscana non estende automaticamente il perimetro pubblico del primo rilascio.

## Impact

Schema e migrazioni PostgreSQL; `internal/domain`, `internal/registry`, acquisizione e pianificazione, `internal/operations`, `internal/publicquery` per isolamento territoriale, wiring in `cmd/iwa`, route/template/asset in `internal/server`. Nuovi endpoint amministrativi privati; URL e contratti JSON esistenti preservati. Test di migrazione, isolamento fra due regioni, ciclo di vita, storico e navigazione browser.

## Non-goals

Non attivare in produzione regioni o fonti durante la migrazione, non acquisire automaticamente dati visitando le pagine, non implementare adattatori per ogni portale regionale italiano, non dichiarare copertura completa o pubblicabilità sulla sola base dell'abilitazione territoriale.
