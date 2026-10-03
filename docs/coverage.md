# Italian weather alert coverage tracker

**Snapshot: 30 September 2026.** This is a planning and source-verification inventory, not a live alert feed or a claim of public coverage. The project aims to publish official regional weather and hydrogeological alerts and related municipal notices and measures throughout Italy. Regional and municipal sources must be checked and accepted separately before their results can be made public.

Each of the 20 regional tables starts with a **region row** summarizing regional-source implementation and checks. The remaining rows list every municipality in that region alphabetically. In the municipality rows, *implementation* means onboarding that municipality’s source, not the existence of a generic API or of an ISTAT registry entry. *Checks* means source research, preview, trial and acceptance; it does not describe postal validation. “Not checked” never means “no alerts.” In municipality rows, **—** means no source integration or source check is recorded in this inventory; it does not mean that the municipality has issued no alerts.

Linked territory names open [source knowledge guides](territori/README.md) with scoped rules, exceptions and discoveries. Those guides may be revised after this inventory's snapshot; their existence or revision does not change the source-status assessments below.

**Subsequent acceptance precheck — 2 October 2026:** the three CFR products and
Calcinaia remain pending acceptance and public enablement. Campaign assessments
and a source-scoped evidence inventory were prepared; missing reviewed comparisons,
failure/event cases and the municipal regression must be resolved before acceptance.
Technical tests use synthetic fixtures and do not establish real-source acceptance.
See [TOS-006](territori/09-toscana/README.md#tos-006--il-tempo-di-osservazione-non-completa-il-collaudo-delle-fonti)
and [CLN-005](territori/09-toscana/comuni/050004-calcinaia.md#cln-005--il-recupero-non-sostituisce-la-rivalutazione-della-regressione).
The original dated inventory below is preserved; detailed operational evidence remains private.

**Subsequent delegated review and bounded trial — 3 October 2026:** original
PDF/notice review was performed by the assistant on explicit operator delegation,
with historical human confirmations preserved. Tasks 26.5/26.6 now verify evidence
attribution through API/MCP and isolated Calcinaia visibility with rollback.
Source acceptance remains pending: the current municipal completeness suite,
generic municipal fact projection and regional graphical interpretation still
require successful verification. No source acceptance or public enablement was
inferred from the document review or development trial. See
[the procedure](operations/cittadino-informato.md#valutazione-e-trial-delimitato--task-265266).

**Subsequent municipal integration verification — 3 October 2026:** primary
municipal extraction-to-domain integration and configured provenance history are
now verified in development, including a bounded retained-result replay and
actual API/MCP equivalence. This supersedes the generic projection gap above in
the tested scope. Municipal completeness, false-positive handling and graphical
CFR interpretation still require successful verification; source acceptance and
production publication remain pending. See [CLN-017](territori/09-toscana/comuni/050004-calcinaia.md#cln-017--integrazione-ordinaria-delle-misure-verificata-in-development).

The CAP column lists every municipality-level postcode in the retained Garda Informatica dataset. A municipality may have several CAPs and a CAP may belong to more than one municipality. A CAP does not identify an alert zone or guarantee an address-level match. Leading zeroes are significant.

## Data and verification

- Municipalities: [ISTAT-derived national registry](nazionale/registri/comuni-italia.csv), captured 15 September 2026; 7,894 municipalities in 20 regions. [Official ISTAT registry](https://www.istat.it/classificazione/codici-dei-comuni-delle-province-e-delle-regioni/) and [reuse terms](https://www.istat.it/dati/open-data/). IWA selected and normalized the registry fields; see [attribution](../THIRD_PARTY_NOTICES.md).
- CAPs: [Garda Informatica archive](prerequisiti-mvp/evidenze/cap-garda-2026-09-11.zip), version 11 September 2026, MIT-licensed according to its included README; 8,456 municipality–CAP pairs. [Publisher page](https://www.gardainformatica.it/database-comuni-italiani/). The CAP data are not an official ISTAT or Poste Italiane postcode registry.
- Join checks for this document: every ISTAT municipality has at least one CAP in this archive; all identifiers, municipality names and region codes match the ISTAT-derived registry; every CAP is five digits; the archive’s two CAP tables agree. These checks do **not** verify each postal assignment or its effective date outside the previously reviewed Tuscany scope.
- Source-status evidence comes from dated Tuscany research and acceptance-precheck snapshots retained privately while their contents are reviewed for redistribution. The status rows are IWA's assessment, not an official certification; no listed source is represented here as publicly accepted.

| Dataset | SHA-256 |
| --- | --- |
| ISTAT-derived CSV | `428bc75a6272923a33a9a4ad4a0fd4e6c4845af1663b5c0db8fa11eedd20620c` |
| Garda Informatica ZIP | `b7837b44477ba901426690782a126e7d44c8611adc91b80413375c995ff34ec1` |

## Regions

- [Abruzzo](#region-13)
- [Basilicata](#region-17)
- [Calabria](#region-18)
- [Campania](#region-15)
- [Emilia-Romagna](#region-08)
- [Friuli-Venezia Giulia](#region-06)
- [Lazio](#region-12)
- [Liguria](#region-07)
- [Lombardia](#region-03)
- [Marche](#region-11)
- [Molise](#region-14)
- [Piemonte](#region-01)
- [Puglia](#region-16)
- [Sardegna](#region-20)
- [Sicilia](#region-19)
- [Toscana](#region-09)
- [Trentino-Alto Adige/Südtirol](#region-04)
- [Umbria](#region-10)
- [Valle d'Aosta/Vallée d'Aoste](#region-02)
- [Veneto](#region-05)

<a id="region-13"></a>

### Abruzzo (305 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Abruzzo — region** | 13 | — | Not started | Regional sources not checked |
| Abbateggio | 068001 | 65020 | — | — |
| Acciano | 066001 | 67020 | — | — |
| Aielli | 066002 | 67041 | — | — |
| Alanno | 068002 | 65020 | — | — |
| Alba Adriatica | 067001 | 64011 | — | — |
| Alfedena | 066003 | 67030 | — | — |
| Altino | 069001 | 66040 | — | — |
| Ancarano | 067002 | 64010 | — | — |
| Anversa degli Abruzzi | 066004 | 67030 | — | — |
| Archi | 069002 | 66044 | — | — |
| Ari | 069003 | 66010 | — | — |
| Arielli | 069004 | 66030 | — | — |
| Arsita | 067003 | 64031 | — | — |
| Ateleta | 066005 | 67030 | — | — |
| Atessa | 069005 | 66041 | — | — |
| Atri | 067004 | 64032 | — | — |
| Avezzano | 066006 | 67051 | — | — |
| Balsorano | 066007 | 67052 | — | — |
| Barete | 066008 | 67010 | — | — |
| Barisciano | 066009 | 67021 | — | — |
| Barrea | 066010 | 67030 | — | — |
| Basciano | 067005 | 64030 | — | — |
| Bellante | 067006 | 64020 | — | — |
| Bisegna | 066011 | 67050 | — | — |
| Bisenti | 067007 | 64033 | — | — |
| Bolognano | 068003 | 65020 | — | — |
| Bomba | 069006 | 66042 | — | — |
| Borrello | 069007 | 66040 | — | — |
| Brittoli | 068004 | 65010 | — | — |
| Bucchianico | 069008 | 66011 | — | — |
| Bugnara | 066012 | 67030 | — | — |
| Bussi sul Tirino | 068005 | 65022 | — | — |
| Cagnano Amiterno | 066013 | 67012 | — | — |
| Calascio | 066014 | 67020 | — | — |
| Campli | 067008 | 64012 | — | — |
| Campo di Giove | 066015 | 67030 | — | — |
| Campotosto | 066016 | 67013 | — | — |
| Canistro | 066017 | 67050 | — | — |
| Canosa Sannita | 069010 | 66010 | — | — |
| Cansano | 066018 | 67030 | — | — |
| Canzano | 067009 | 64020 | — | — |
| Capestrano | 066019 | 67022 | — | — |
| Capistrello | 066020 | 67053 | — | — |
| Capitignano | 066021 | 67014 | — | — |
| Caporciano | 066022 | 67020 | — | — |
| Cappadocia | 066023 | 67060 | — | — |
| Cappelle sul Tavo | 068006 | 65010 | — | — |
| Caramanico Terme | 068007 | 65023 | — | — |
| Carapelle Calvisio | 066024 | 67020 | — | — |
| Carpineto della Nora | 068008 | 65010 | — | — |
| Carpineto Sinello | 069011 | 66030 | — | — |
| Carsoli | 066025 | 67061 | — | — |
| Carunchio | 069012 | 66050 | — | — |
| Casacanditella | 069013 | 66010 | — | — |
| Casalanguida | 069014 | 66031 | — | — |
| Casalbordino | 069015 | 66021 | — | — |
| Casalincontrada | 069016 | 66012 | — | — |
| Casoli | 069017 | 66043 | — | — |
| Castel Castagna | 067010 | 64030 | — | — |
| Castel del Monte | 066026 | 67023 | — | — |
| Castel di Ieri | 066027 | 67020 | — | — |
| Castel di Sangro | 066028 | 67031 | — | — |
| Castel Frentano | 069018 | 66032 | — | — |
| Castelguidone | 069019 | 66040 | — | — |
| Castellafiume | 066029 | 67050 | — | — |
| Castellalto | 067011 | 64020 | — | — |
| Castelli | 067012 | 64041 | — | — |
| Castelvecchio Calvisio | 066030 | 67020 | — | — |
| Castelvecchio Subequo | 066031 | 67024 | — | — |
| Castiglione a Casauria | 068009 | 65020 | — | — |
| Castiglione Messer Marino | 069020 | 66033 | — | — |
| Castiglione Messer Raimondo | 067013 | 64034 | — | — |
| Castilenti | 067014 | 64035 | — | — |
| Catignano | 068010 | 65011 | — | — |
| Celano | 066032 | 67043 | — | — |
| Celenza sul Trigno | 069021 | 66050 | — | — |
| Cellino Attanasio | 067015 | 64036 | — | — |
| Cepagatti | 068011 | 65012 | — | — |
| Cerchio | 066033 | 67044 | — | — |
| Cermignano | 067016 | 64037 | — | — |
| Chieti | 069022 | 66100 | — | — |
| Città Sant'Angelo | 068012 | 65013 | — | — |
| Civita d'Antino | 066034 | 67050 | — | — |
| Civitaluparella | 069023 | 66040 | — | — |
| Civitaquana | 068013 | 65010 | — | — |
| Civitella Alfedena | 066035 | 67030 | — | — |
| Civitella Casanova | 068014 | 65010 | — | — |
| Civitella del Tronto | 067017 | 64010 | — | — |
| Civitella Messer Raimondo | 069024 | 66010 | — | — |
| Civitella Roveto | 066036 | 67054 | — | — |
| Cocullo | 066037 | 67030 | — | — |
| Collarmele | 066038 | 67040 | — | — |
| Collecorvino | 068015 | 65010 | — | — |
| Colledara | 067018 | 64042 | — | — |
| Colledimacine | 069025 | 66010 | — | — |
| Colledimezzo | 069026 | 66040 | — | — |
| Collelongo | 066039 | 67050 | — | — |
| Collepietro | 066040 | 67020 | — | — |
| Colonnella | 067019 | 64010 | — | — |
| Controguerra | 067020 | 64010 | — | — |
| Corfinio | 066041 | 67030 | — | — |
| Corropoli | 067021 | 64013 | — | — |
| Cortino | 067022 | 64040 | — | — |
| Corvara | 068016 | 65020 | — | — |
| Crecchio | 069027 | 66014 | — | — |
| Crognaleto | 067023 | 64043 | — | — |
| Cugnoli | 068017 | 65020 | — | — |
| Cupello | 069028 | 66051 | — | — |
| Dogliola | 069029 | 66050 | — | — |
| Elice | 068018 | 65010 | — | — |
| Fagnano Alto | 066042 | 67020 | — | — |
| Fallo | 069104 | 66040 | — | — |
| Fano Adriano | 067024 | 64044 | — | — |
| Fara Filiorum Petri | 069030 | 66010 | — | — |
| Fara San Martino | 069031 | 66015 | — | — |
| Farindola | 068019 | 65010 | — | — |
| Filetto | 069032 | 66030 | — | — |
| Fontecchio | 066043 | 67020 | — | — |
| Fossa | 066044 | 67020 | — | — |
| Fossacesia | 069033 | 66022 | — | — |
| Fraine | 069034 | 66050 | — | — |
| Francavilla al Mare | 069035 | 66023 | — | — |
| Fresagrandinaria | 069036 | 66050 | — | — |
| Frisa | 069037 | 66030 | — | — |
| Furci | 069038 | 66050 | — | — |
| Gagliano Aterno | 066045 | 67020 | — | — |
| Gamberale | 069039 | 66040 | — | — |
| Gessopalena | 069040 | 66010 | — | — |
| Gioia dei Marsi | 066046 | 67055 | — | — |
| Gissi | 069041 | 66052 | — | — |
| Giuliano Teatino | 069042 | 66010 | — | — |
| Giulianova | 067025 | 64021 | — | — |
| Goriano Sicoli | 066047 | 67030 | — | — |
| Guardiagrele | 069043 | 66016 | — | — |
| Guilmi | 069044 | 66050 | — | — |
| Introdacqua | 066048 | 67030 | — | — |
| Isola del Gran Sasso d'Italia | 067026 | 64045 | — | — |
| L'Aquila | 066049 | 67100 | — | — |
| Lama dei Peligni | 069045 | 66010 | — | — |
| Lanciano | 069046 | 66034 | — | — |
| Lecce nei Marsi | 066050 | 67050 | — | — |
| Lentella | 069047 | 66050 | — | — |
| Lettomanoppello | 068020 | 65020 | — | — |
| Lettopalena | 069048 | 66010 | — | — |
| Liscia | 069049 | 66050 | — | — |
| Loreto Aprutino | 068021 | 65014 | — | — |
| Luco dei Marsi | 066051 | 67056 | — | — |
| Lucoli | 066052 | 67045 | — | — |
| Magliano de' Marsi | 066053 | 67062 | — | — |
| Manoppello | 068022 | 65024 | — | — |
| Martinsicuro | 067047 | 64014 | — | — |
| Massa d'Albe | 066054 | 67050 | — | — |
| Miglianico | 069050 | 66010 | — | — |
| Molina Aterno | 066055 | 67020 | — | — |
| Montazzoli | 069051 | 66030 | — | — |
| Montebello di Bertona | 068023 | 65010 | — | — |
| Montebello sul Sangro | 069009 | 66040 | — | — |
| Monteferrante | 069052 | 66040 | — | — |
| Montefino | 067027 | 64030 | — | — |
| Montelapiano | 069053 | 66040 | — | — |
| Montenerodomo | 069054 | 66010 | — | — |
| Monteodorisio | 069055 | 66050 | — | — |
| Montereale | 066056 | 67015 | — | — |
| Montesilvano | 068024 | 65015 | — | — |
| Montorio al Vomano | 067028 | 64046 | — | — |
| Morino | 066057 | 67050 | — | — |
| Morro d'Oro | 067029 | 64020 | — | — |
| Mosciano Sant'Angelo | 067030 | 64023 | — | — |
| Moscufo | 068025 | 65010 | — | — |
| Mozzagrogna | 069056 | 66030 | — | — |
| Navelli | 066058 | 67020 | — | — |
| Nereto | 067031 | 64015 | — | — |
| Nocciano | 068026 | 65010 | — | — |
| Notaresco | 067032 | 64024 | — | — |
| Ocre | 066059 | 67040 | — | — |
| Ofena | 066060 | 67025 | — | — |
| Opi | 066061 | 67030 | — | — |
| Oricola | 066062 | 67063 | — | — |
| Orsogna | 069057 | 66036 | — | — |
| Ortona | 069058 | 66026 | — | — |
| Ortona dei Marsi | 066063 | 67050 | — | — |
| Ortucchio | 066064 | 67050 | — | — |
| Ovindoli | 066065 | 67046 | — | — |
| Pacentro | 066066 | 67030 | — | — |
| Paglieta | 069059 | 66020 | — | — |
| Palena | 069060 | 66017 | — | — |
| Palmoli | 069061 | 66050 | — | — |
| Palombaro | 069062 | 66010 | — | — |
| Penna Sant'Andrea | 067033 | 64039 | — | — |
| Pennadomo | 069063 | 66040 | — | — |
| Pennapiedimonte | 069064 | 66010 | — | — |
| Penne | 068027 | 65017 | — | — |
| Perano | 069065 | 66040 | — | — |
| Pereto | 066067 | 67064 | — | — |
| Pescara | 068028 | 65121, 65122, 65123, 65124, 65125, 65126, 65127, 65128, 65129 | — | — |
| Pescasseroli | 066068 | 67032 | — | — |
| Pescina | 066069 | 67057 | — | — |
| Pescocostanzo | 066070 | 67033 | — | — |
| Pescosansonesco | 068029 | 65020 | — | — |
| Pettorano sul Gizio | 066071 | 67034 | — | — |
| Pianella | 068030 | 65019 | — | — |
| Picciano | 068031 | 65010 | — | — |
| Pietracamela | 067034 | 64047 | — | — |
| Pietraferrazzana | 069103 | 66040 | — | — |
| Pietranico | 068032 | 65020 | — | — |
| Pineto | 067035 | 64025 | — | — |
| Pizzoferrato | 069066 | 66040 | — | — |
| Pizzoli | 066072 | 67017 | — | — |
| Poggio Picenze | 066073 | 67026 | — | — |
| Poggiofiorito | 069067 | 66030 | — | — |
| Pollutri | 069068 | 66020 | — | — |
| Popoli Terme | 068033 | 65026 | — | — |
| Prata d'Ansidonia | 066074 | 67020 | — | — |
| Pratola Peligna | 066075 | 67035 | — | — |
| Pretoro | 069069 | 66010 | — | — |
| Prezza | 066076 | 67030 | — | — |
| Quadri | 069070 | 66040 | — | — |
| Raiano | 066077 | 67027 | — | — |
| Rapino | 069071 | 66010 | — | — |
| Ripa Teatina | 069072 | 66010 | — | — |
| Rivisondoli | 066078 | 67036 | — | — |
| Rocca di Botte | 066080 | 67066 | — | — |
| Rocca di Cambio | 066081 | 67047 | — | — |
| Rocca di Mezzo | 066082 | 67048 | — | — |
| Rocca Pia | 066083 | 67030 | — | — |
| Rocca San Giovanni | 069074 | 66020 | — | — |
| Rocca Santa Maria | 067036 | 64010 | — | — |
| Roccacasale | 066079 | 67030 | — | — |
| Roccamontepiano | 069073 | 66010 | — | — |
| Roccamorice | 068034 | 65020 | — | — |
| Roccaraso | 066084 | 67037 | — | — |
| Roccascalegna | 069075 | 66040 | — | — |
| Roccaspinalveti | 069076 | 66050 | — | — |
| Roio del Sangro | 069077 | 66040 | — | — |
| Rosciano | 068035 | 65020 | — | — |
| Rosello | 069078 | 66040 | — | — |
| Roseto degli Abruzzi | 067037 | 64026 | — | — |
| Salle | 068036 | 65020 | — | — |
| San Benedetto dei Marsi | 066085 | 67058 | — | — |
| San Benedetto in Perillis | 066086 | 67020 | — | — |
| San Buono | 069079 | 66050 | — | — |
| San Demetrio ne' Vestini | 066087 | 67028 | — | — |
| San Giovanni Lipioni | 069080 | 66050 | — | — |
| San Giovanni Teatino | 069081 | 66020 | — | — |
| San Martino sulla Marrucina | 069082 | 66010 | — | — |
| San Pio delle Camere | 066088 | 67020 | — | — |
| San Salvo | 069083 | 66050 | — | — |
| San Valentino in Abruzzo Citeriore | 068038 | 65020 | — | — |
| San Vincenzo Valle Roveto | 066092 | 67050 | — | — |
| San Vito Chietino | 069086 | 66038 | — | — |
| Sant'Egidio alla Vibrata | 067038 | 64016 | — | — |
| Sant'Eufemia a Maiella | 068037 | 65020 | — | — |
| Sant'Eusanio del Sangro | 069085 | 66037 | — | — |
| Sant'Eusanio Forconese | 066090 | 67020 | — | — |
| Sant'Omero | 067039 | 64027 | — | — |
| Santa Maria Imbaro | 069084 | 66030 | — | — |
| Sante Marie | 066089 | 67067 | — | — |
| Santo Stefano di Sessanio | 066091 | 67020 | — | — |
| Scafa | 068039 | 65027 | — | — |
| Scanno | 066093 | 67038 | — | — |
| Scerni | 069087 | 66020 | — | — |
| Schiavi di Abruzzo | 069088 | 66045 | — | — |
| Scontrone | 066094 | 67030 | — | — |
| Scoppito | 066095 | 67019 | — | — |
| Scurcola Marsicana | 066096 | 67068 | — | — |
| Secinaro | 066097 | 67029 | — | — |
| Serramonacesca | 068040 | 65025 | — | — |
| Silvi | 067040 | 64028 | — | — |
| Spoltore | 068041 | 65010 | — | — |
| Sulmona | 066098 | 67039 | — | — |
| Tagliacozzo | 066099 | 67069 | — | — |
| Taranta Peligna | 069089 | 66018 | — | — |
| Teramo | 067041 | 64100 | — | — |
| Tione degli Abruzzi | 066100 | 67020 | — | — |
| Tocco da Casauria | 068042 | 65028 | — | — |
| Tollo | 069090 | 66010 | — | — |
| Torano Nuovo | 067042 | 64010 | — | — |
| Torino di Sangro | 069091 | 66020 | — | — |
| Tornareccio | 069092 | 66046 | — | — |
| Tornimparte | 066101 | 67049 | — | — |
| Torre de' Passeri | 068043 | 65029 | — | — |
| Torrebruna | 069093 | 66050 | — | — |
| Torrevecchia Teatina | 069094 | 66010 | — | — |
| Torricella Peligna | 069095 | 66019 | — | — |
| Torricella Sicura | 067043 | 64010 | — | — |
| Tortoreto | 067044 | 64018 | — | — |
| Tossicia | 067045 | 64049 | — | — |
| Trasacco | 066102 | 67059 | — | — |
| Treglio | 069096 | 66030 | — | — |
| Tufillo | 069097 | 66050 | — | — |
| Turrivalignani | 068044 | 65020 | — | — |
| Vacri | 069098 | 66010 | — | — |
| Valle Castellana | 067046 | 64010 | — | — |
| Vasto | 069099 | 66054 | — | — |
| Vicoli | 068045 | 65010 | — | — |
| Villa Celiera | 068046 | 65010 | — | — |
| Villa Sant'Angelo | 066105 | 67020 | — | — |
| Villa Santa Lucia degli Abruzzi | 066104 | 67020 | — | — |
| Villa Santa Maria | 069102 | 66047 | — | — |
| Villalago | 066103 | 67030 | — | — |
| Villalfonsina | 069100 | 66020 | — | — |
| Villamagna | 069101 | 66010 | — | — |
| Villavallelonga | 066106 | 67050 | — | — |
| Villetta Barrea | 066107 | 67030 | — | — |
| Vittorito | 066108 | 67030 | — | — |

<a id="region-17"></a>

### Basilicata (131 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Basilicata — region** | 17 | — | Not started | Regional sources not checked |
| Abriola | 076001 | 85010 | — | — |
| Accettura | 077001 | 75011 | — | — |
| Acerenza | 076002 | 85011 | — | — |
| Albano di Lucania | 076003 | 85010 | — | — |
| Aliano | 077002 | 75010 | — | — |
| Anzi | 076004 | 85010 | — | — |
| Armento | 076005 | 85010 | — | — |
| Atella | 076006 | 85020 | — | — |
| Avigliano | 076007 | 85021 | — | — |
| Balvano | 076008 | 85050 | — | — |
| Banzi | 076009 | 85010 | — | — |
| Baragiano | 076010 | 85050 | — | — |
| Barile | 076011 | 85022 | — | — |
| Bella | 076012 | 85051 | — | — |
| Bernalda | 077003 | 75012 | — | — |
| Brienza | 076013 | 85050 | — | — |
| Brindisi Montagna | 076014 | 85010 | — | — |
| Calciano | 077004 | 75010 | — | — |
| Calvello | 076015 | 85010 | — | — |
| Calvera | 076016 | 85030 | — | — |
| Campomaggiore | 076017 | 85010 | — | — |
| Cancellara | 076018 | 85010 | — | — |
| Carbone | 076019 | 85030 | — | — |
| Castelgrande | 076021 | 85050 | — | — |
| Castelluccio Inferiore | 076022 | 85040 | — | — |
| Castelluccio Superiore | 076023 | 85040 | — | — |
| Castelmezzano | 076024 | 85010 | — | — |
| Castelsaraceno | 076025 | 85031 | — | — |
| Castronuovo di Sant'Andrea | 076026 | 85030 | — | — |
| Cersosimo | 076027 | 85030 | — | — |
| Chiaromonte | 076028 | 85032 | — | — |
| Cirigliano | 077005 | 75010 | — | — |
| Colobraro | 077006 | 75021 | — | — |
| Corleto Perticara | 076029 | 85012 | — | — |
| Craco | 077007 | 75010 | — | — |
| Episcopia | 076030 | 85033 | — | — |
| Fardella | 076031 | 85034 | — | — |
| Ferrandina | 077008 | 75013 | — | — |
| Filiano | 076032 | 85020 | — | — |
| Forenza | 076033 | 85023 | — | — |
| Francavilla in Sinni | 076034 | 85034 | — | — |
| Gallicchio | 076035 | 85010 | — | — |
| Garaguso | 077009 | 75010 | — | — |
| Genzano di Lucania | 076036 | 85013 | — | — |
| Ginestra | 076099 | 85020 | — | — |
| Gorgoglione | 077010 | 75010 | — | — |
| Grassano | 077011 | 75014 | — | — |
| Grottole | 077012 | 75010 | — | — |
| Grumento Nova | 076037 | 85050 | — | — |
| Guardia Perticara | 076038 | 85010 | — | — |
| Irsina | 077013 | 75022 | — | — |
| Lagonegro | 076039 | 85042 | — | — |
| Latronico | 076040 | 85043 | — | — |
| Laurenzana | 076041 | 85014 | — | — |
| Lauria | 076042 | 85044 | — | — |
| Lavello | 076043 | 85024 | — | — |
| Maratea | 076044 | 85046 | — | — |
| Marsico Nuovo | 076045 | 85052 | — | — |
| Marsicovetere | 076046 | 85050 | — | — |
| Maschito | 076047 | 85020 | — | — |
| Matera | 077014 | 75100 | — | — |
| Melfi | 076048 | 85025 | — | — |
| Miglionico | 077015 | 75010 | — | — |
| Missanello | 076049 | 85010 | — | — |
| Moliterno | 076050 | 85047 | — | — |
| Montalbano Jonico | 077016 | 75023 | — | — |
| Montemilone | 076051 | 85020 | — | — |
| Montemurro | 076052 | 85053 | — | — |
| Montescaglioso | 077017 | 75024 | — | — |
| Muro Lucano | 076053 | 85054 | — | — |
| Nemoli | 076054 | 85040 | — | — |
| Noepoli | 076055 | 85035 | — | — |
| Nova Siri | 077018 | 75020 | — | — |
| Oliveto Lucano | 077019 | 75010 | — | — |
| Oppido Lucano | 076056 | 85015 | — | — |
| Palazzo San Gervasio | 076057 | 85026 | — | — |
| Paterno | 076100 | 85050 | — | — |
| Pescopagano | 076058 | 85020 | — | — |
| Picerno | 076059 | 85055 | — | — |
| Pietragalla | 076060 | 85016 | — | — |
| Pietrapertosa | 076061 | 85010 | — | — |
| Pignola | 076062 | 85010 | — | — |
| Pisticci | 077020 | 75015 | — | — |
| Policoro | 077021 | 75025 | — | — |
| Pomarico | 077022 | 75016 | — | — |
| Potenza | 076063 | 85100 | — | — |
| Rapolla | 076064 | 85027 | — | — |
| Rapone | 076065 | 85020 | — | — |
| Rionero in Vulture | 076066 | 85028 | — | — |
| Ripacandida | 076067 | 85020 | — | — |
| Rivello | 076068 | 85040 | — | — |
| Roccanova | 076069 | 85036 | — | — |
| Rotonda | 076070 | 85048 | — | — |
| Rotondella | 077023 | 75026 | — | — |
| Ruoti | 076071 | 85056 | — | — |
| Ruvo del Monte | 076072 | 85020 | — | — |
| Salandra | 077024 | 75017 | — | — |
| San Chirico Nuovo | 076073 | 85010 | — | — |
| San Chirico Raparo | 076074 | 85030 | — | — |
| San Costantino Albanese | 076075 | 85030 | — | — |
| San Fele | 076076 | 85020 | — | — |
| San Giorgio Lucano | 077025 | 75027 | — | — |
| San Martino d'Agri | 076077 | 85030 | — | — |
| San Mauro Forte | 077026 | 75010 | — | — |
| San Paolo Albanese | 076020 | 85030 | — | — |
| San Severino Lucano | 076078 | 85030 | — | — |
| Sant'Angelo Le Fratte | 076079 | 85050 | — | — |
| Sant'Arcangelo | 076080 | 85037 | — | — |
| Sarconi | 076081 | 85050 | — | — |
| Sasso di Castalda | 076082 | 85050 | — | — |
| Satriano di Lucania | 076083 | 85050 | — | — |
| Savoia di Lucania | 076084 | 85050 | — | — |
| Scanzano Jonico | 077031 | 75020 | — | — |
| Senise | 076085 | 85038 | — | — |
| Spinoso | 076086 | 85039 | — | — |
| Stigliano | 077027 | 75018 | — | — |
| Teana | 076087 | 85032 | — | — |
| Terranova di Pollino | 076088 | 85030 | — | — |
| Tito | 076089 | 85050 | — | — |
| Tolve | 076090 | 85017 | — | — |
| Tramutola | 076091 | 85057 | — | — |
| Trecchina | 076092 | 85049 | — | — |
| Tricarico | 077028 | 75019 | — | — |
| Trivigno | 076093 | 85018 | — | — |
| Tursi | 077029 | 75028 | — | — |
| Vaglio Basilicata | 076094 | 85010 | — | — |
| Valsinni | 077030 | 75029 | — | — |
| Venosa | 076095 | 85029 | — | — |
| Vietri di Potenza | 076096 | 85058 | — | — |
| Viggianello | 076097 | 85040 | — | — |
| Viggiano | 076098 | 85059 | — | — |

<a id="region-18"></a>

### Calabria (404 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Calabria — region** | 18 | — | Not started | Regional sources not checked |
| Acquaformosa | 078001 | 87010 | — | — |
| Acquappesa | 078002 | 87020 | — | — |
| Acquaro | 102001 | 89832 | — | — |
| Acri | 078003 | 87041 | — | — |
| Africo | 080001 | 89030 | — | — |
| Agnana Calabra | 080002 | 89040 | — | — |
| Aiello Calabro | 078004 | 87031 | — | — |
| Aieta | 078005 | 87020 | — | — |
| Albi | 079002 | 88055 | — | — |
| Albidona | 078006 | 87070 | — | — |
| Alessandria del Carretto | 078007 | 87070 | — | — |
| Altilia | 078008 | 87040 | — | — |
| Altomonte | 078009 | 87042 | — | — |
| Amantea | 078010 | 87032 | — | — |
| Amaroni | 079003 | 88050 | — | — |
| Amato | 079004 | 88040 | — | — |
| Amendolara | 078011 | 87071 | — | — |
| Andali | 079005 | 88050 | — | — |
| Anoia | 080003 | 89020 | — | — |
| Antonimina | 080004 | 89040 | — | — |
| Aprigliano | 078012 | 87051 | — | — |
| Ardore | 080005 | 89031 | — | — |
| Arena | 102002 | 89832 | — | — |
| Argusto | 079007 | 88060 | — | — |
| Badolato | 079008 | 88060 | — | — |
| Bagaladi | 080006 | 89060 | — | — |
| Bagnara Calabra | 080007 | 89011 | — | — |
| Belcastro | 079009 | 88050 | — | — |
| Belmonte Calabro | 078013 | 87033 | — | — |
| Belsito | 078014 | 87030 | — | — |
| Belvedere di Spinello | 101001 | 88824 | — | — |
| Belvedere Marittimo | 078015 | 87021 | — | — |
| Benestare | 080008 | 89030 | — | — |
| Bianchi | 078016 | 87050 | — | — |
| Bianco | 080009 | 89032 | — | — |
| Bisignano | 078017 | 87043 | — | — |
| Bivongi | 080010 | 89040 | — | — |
| Bocchigliero | 078018 | 87060 | — | — |
| Bonifati | 078019 | 87020 | — | — |
| Borgia | 079011 | 88021 | — | — |
| Botricello | 079012 | 88070 | — | — |
| Bova | 080011 | 89033 | — | — |
| Bova Marina | 080013 | 89035 | — | — |
| Bovalino | 080012 | 89034 | — | — |
| Brancaleone | 080014 | 89036 | — | — |
| Briatico | 102003 | 89817 | — | — |
| Brognaturo | 102004 | 89822 | — | — |
| Bruzzano Zeffirio | 080015 | 89030 | — | — |
| Buonvicino | 078020 | 87020 | — | — |
| Caccuri | 101002 | 88833 | — | — |
| Calanna | 080016 | 89050 | — | — |
| Calopezzati | 078021 | 87060 | — | — |
| Caloveto | 078022 | 87060 | — | — |
| Camini | 080017 | 89040 | — | — |
| Campana | 078023 | 87061 | — | — |
| Campo Calabro | 080018 | 89052 | — | — |
| Candidoni | 080019 | 89020 | — | — |
| Canna | 078024 | 87070 | — | — |
| Canolo | 080020 | 89040 | — | — |
| Capistrano | 102005 | 89818 | — | — |
| Caraffa del Bianco | 080021 | 89030 | — | — |
| Caraffa di Catanzaro | 079017 | 88050 | — | — |
| Cardeto | 080022 | 89060 | — | — |
| Cardinale | 079018 | 88062 | — | — |
| Careri | 080023 | 89030 | — | — |
| Carfizzi | 101003 | 88817 | — | — |
| Cariati | 078025 | 87062 | — | — |
| Carlopoli | 079020 | 88040 | — | — |
| Carolei | 078026 | 87030 | — | — |
| Carpanzano | 078027 | 87050 | — | — |
| Casabona | 101004 | 88822 | — | — |
| Casali del Manco | 078156 | 87059 | — | — |
| Casignana | 080024 | 89030 | — | — |
| Cassano all'Ionio | 078029 | 87011 | — | — |
| Castelsilano | 101005 | 88834 | — | — |
| Castiglione Cosentino | 078030 | 87040 | — | — |
| Castrolibero | 078031 | 87040 | — | — |
| Castroregio | 078032 | 87070 | — | — |
| Castrovillari | 078033 | 87012 | — | — |
| Catanzaro | 079023 | 88100 | — | — |
| Caulonia | 080025 | 89041 | — | — |
| Celico | 078034 | 87053 | — | — |
| Cellara | 078035 | 87050 | — | — |
| Cenadi | 079024 | 88067 | — | — |
| Centrache | 079025 | 88067 | — | — |
| Cerchiara di Calabria | 078036 | 87070 | — | — |
| Cerenzia | 101006 | 88833 | — | — |
| Cerisano | 078037 | 87044 | — | — |
| Cerva | 079027 | 88050 | — | — |
| Cervicati | 078038 | 87010 | — | — |
| Cerzeto | 078039 | 87040 | — | — |
| Cessaniti | 102006 | 89816 | — | — |
| Cetraro | 078040 | 87022 | — | — |
| Chiaravalle Centrale | 079029 | 88064 | — | — |
| Cicala | 079030 | 88040 | — | — |
| Ciminà | 080026 | 89040 | — | — |
| Cinquefrondi | 080027 | 89021 | — | — |
| Cirò | 101007 | 88813 | — | — |
| Cirò Marina | 101008 | 88811 | — | — |
| Cittanova | 080028 | 89022 | — | — |
| Civita | 078041 | 87010 | — | — |
| Cleto | 078042 | 87030 | — | — |
| Colosimi | 078043 | 87050 | — | — |
| Condofuri | 080029 | 89030 | — | — |
| Conflenti | 079033 | 88040 | — | — |
| Corigliano-Rossano | 078157 | 87064 | — | — |
| Cortale | 079034 | 88020 | — | — |
| Cosenza | 078045 | 87100 | — | — |
| Cosoleto | 080030 | 89050 | — | — |
| Cotronei | 101009 | 88836 | — | — |
| Cropalati | 078046 | 87060 | — | — |
| Cropani | 079036 | 88051 | — | — |
| Crosia | 078047 | 87060 | — | — |
| Crotone | 101010 | 88900 | — | — |
| Crucoli | 101011 | 88812 | — | — |
| Curinga | 079039 | 88022 | — | — |
| Cutro | 101012 | 88842 | — | — |
| Dasà | 102007 | 89832 | — | — |
| Davoli | 079042 | 88060 | — | — |
| Decollatura | 079043 | 88041 | — | — |
| Delianuova | 080031 | 89012 | — | — |
| Diamante | 078048 | 87023 | — | — |
| Dinami | 102008 | 89833 | — | — |
| Dipignano | 078049 | 87045 | — | — |
| Domanico | 078050 | 87030 | — | — |
| Drapia | 102009 | 89862 | — | — |
| Fabrizia | 102010 | 89823 | — | — |
| Fagnano Castello | 078051 | 87013 | — | — |
| Falconara Albanese | 078052 | 87030 | — | — |
| Falerna | 079047 | 88042 | — | — |
| Feroleto Antico | 079048 | 88040 | — | — |
| Feroleto della Chiesa | 080032 | 89050 | — | — |
| Ferruzzano | 080033 | 89030 | — | — |
| Figline Vegliaturo | 078053 | 87050 | — | — |
| Filadelfia | 102011 | 89814 | — | — |
| Filandari | 102012 | 89841 | — | — |
| Filogaso | 102013 | 89843 | — | — |
| Firmo | 078054 | 87010 | — | — |
| Fiumara | 080034 | 89050 | — | — |
| Fiumefreddo Bruzio | 078055 | 87030 | — | — |
| Fossato Serralta | 079052 | 88050 | — | — |
| Francavilla Angitola | 102014 | 89815 | — | — |
| Francavilla Marittima | 078056 | 87072 | — | — |
| Francica | 102015 | 89851 | — | — |
| Frascineto | 078057 | 87010 | — | — |
| Fuscaldo | 078058 | 87024 | — | — |
| Gagliato | 079055 | 88060 | — | — |
| Galatro | 080035 | 89054 | — | — |
| Gasperina | 079056 | 88060 | — | — |
| Gerace | 080036 | 89040 | — | — |
| Gerocarne | 102016 | 89831 | — | — |
| Giffone | 080037 | 89020 | — | — |
| Gimigliano | 079058 | 88045 | — | — |
| Gioia Tauro | 080038 | 89013 | — | — |
| Gioiosa Ionica | 080039 | 89042 | — | — |
| Girifalco | 079059 | 88024 | — | — |
| Gizzeria | 079060 | 88040 | — | — |
| Grimaldi | 078059 | 87034 | — | — |
| Grisolia | 078060 | 87020 | — | — |
| Grotteria | 080040 | 89043 | — | — |
| Guardavalle | 079061 | 88065 | — | — |
| Guardia Piemontese | 078061 | 87020 | — | — |
| Isca sullo Ionio | 079063 | 88060 | — | — |
| Isola di Capo Rizzuto | 101013 | 88841 | — | — |
| Jacurso | 079065 | 88020 | — | — |
| Jonadi | 102017 | 89851 | — | — |
| Joppolo | 102018 | 89863 | — | — |
| Laganadi | 080041 | 89050 | — | — |
| Lago | 078062 | 87035 | — | — |
| Laino Borgo | 078063 | 87014 | — | — |
| Laino Castello | 078064 | 87015 | — | — |
| Lamezia Terme | 079160 | 88046 | — | — |
| Lappano | 078065 | 87050 | — | — |
| Lattarico | 078066 | 87010 | — | — |
| Laureana di Borrello | 080042 | 89023 | — | — |
| Limbadi | 102019 | 89844 | — | — |
| Locri | 080043 | 89044 | — | — |
| Longobardi | 078067 | 87030 | — | — |
| Longobucco | 078068 | 87066 | — | — |
| Lungro | 078069 | 87010 | — | — |
| Luzzi | 078070 | 87040 | — | — |
| Magisano | 079068 | 88050 | — | — |
| Maida | 079069 | 88025 | — | — |
| Maierà | 078071 | 87020 | — | — |
| Maierato | 102020 | 89843 | — | — |
| Malito | 078072 | 87030 | — | — |
| Malvito | 078073 | 87010 | — | — |
| Mammola | 080044 | 89045 | — | — |
| Mandatoriccio | 078074 | 87060 | — | — |
| Mangone | 078075 | 87050 | — | — |
| Marano Marchesato | 078076 | 87040 | — | — |
| Marano Principato | 078077 | 87040 | — | — |
| Marcedusa | 079071 | 88050 | — | — |
| Marcellinara | 079072 | 88044 | — | — |
| Marina di Gioiosa Ionica | 080045 | 89046 | — | — |
| Maropati | 080046 | 89020 | — | — |
| Martirano | 079073 | 88040 | — | — |
| Martirano Lombardo | 079074 | 88040 | — | — |
| Martone | 080047 | 89040 | — | — |
| Marzi | 078078 | 87050 | — | — |
| Melicuccà | 080048 | 89020 | — | — |
| Melicucco | 080049 | 89020 | — | — |
| Melissa | 101014 | 88814 | — | — |
| Melito di Porto Salvo | 080050 | 89063 | — | — |
| Mendicino | 078079 | 87040 | — | — |
| Mesoraca | 101015 | 88838 | — | — |
| Miglierina | 079077 | 88040 | — | — |
| Mileto | 102021 | 89852 | — | — |
| Molochio | 080051 | 89010 | — | — |
| Monasterace | 080052 | 89040 | — | — |
| Mongiana | 102022 | 89823 | — | — |
| Mongrassano | 078080 | 87040 | — | — |
| Montalto Uffugo | 078081 | 87046 | — | — |
| Montauro | 079080 | 88060 | — | — |
| Montebello Jonico | 080053 | 89064 | — | — |
| Montegiordano | 078082 | 87070 | — | — |
| Montepaone | 079081 | 88060 | — | — |
| Monterosso Calabro | 102023 | 89819 | — | — |
| Morano Calabro | 078083 | 87016 | — | — |
| Mormanno | 078084 | 87026 | — | — |
| Motta San Giovanni | 080054 | 89065 | — | — |
| Motta Santa Lucia | 079083 | 88040 | — | — |
| Mottafollone | 078085 | 87010 | — | — |
| Nardodipace | 102024 | 89824 | — | — |
| Nicotera | 102025 | 89844 | — | — |
| Nocara | 078086 | 87070 | — | — |
| Nocera Terinese | 079087 | 88047 | — | — |
| Olivadi | 079088 | 88067 | — | — |
| Oppido Mamertina | 080055 | 89014 | — | — |
| Oriolo | 078087 | 87073 | — | — |
| Orsomarso | 078088 | 87020 | — | — |
| Palermiti | 079089 | 88050 | — | — |
| Palizzi | 080056 | 89038 | — | — |
| Pallagorio | 101016 | 88818 | — | — |
| Palmi | 080057 | 89015 | — | — |
| Paludi | 078089 | 87060 | — | — |
| Panettieri | 078090 | 87050 | — | — |
| Paola | 078091 | 87027 | — | — |
| Papasidero | 078092 | 87020 | — | — |
| Parenti | 078093 | 87040 | — | — |
| Parghelia | 102026 | 89861 | — | — |
| Paterno Calabro | 078094 | 87040 | — | — |
| Pazzano | 080058 | 89040 | — | — |
| Pedivigliano | 078096 | 87050 | — | — |
| Pentone | 079092 | 88050 | — | — |
| Petilia Policastro | 101017 | 88837 | — | — |
| Petrizzi | 079094 | 88060 | — | — |
| Petronà | 079095 | 88050 | — | — |
| Piane Crati | 078097 | 87050 | — | — |
| Pianopoli | 079096 | 88040 | — | — |
| Pietrafitta | 078098 | 87050 | — | — |
| Pietrapaola | 078099 | 87060 | — | — |
| Pizzo | 102027 | 89812 | — | — |
| Pizzoni | 102028 | 89834 | — | — |
| Placanica | 080059 | 89040 | — | — |
| Plataci | 078100 | 87070 | — | — |
| Platania | 079099 | 88040 | — | — |
| Platì | 080060 | 89039 | — | — |
| Polia | 102029 | 89813 | — | — |
| Polistena | 080061 | 89024 | — | — |
| Portigliola | 080062 | 89040 | — | — |
| Praia a Mare | 078101 | 87028 | — | — |
| Reggio di Calabria | 080063 | 89121, 89122, 89123, 89124, 89125, 89126, 89127, 89128, 89129, 89131, 89132, 89133, 89134, 89135 | — | — |
| Rende | 078102 | 87036 | — | — |
| Riace | 080064 | 89040 | — | — |
| Ricadi | 102030 | 89866 | — | — |
| Rizziconi | 080065 | 89016 | — | — |
| Rocca di Neto | 101019 | 88821 | — | — |
| Rocca Imperiale | 078103 | 87074 | — | — |
| Roccabernarda | 101018 | 88835 | — | — |
| Roccaforte del Greco | 080066 | 89060 | — | — |
| Roccella Ionica | 080067 | 89047 | — | — |
| Roggiano Gravina | 078104 | 87017 | — | — |
| Roghudi | 080068 | 89060 | — | — |
| Rogliano | 078105 | 87054 | — | — |
| Rombiolo | 102031 | 89841 | — | — |
| Rosarno | 080069 | 89025 | — | — |
| Rose | 078106 | 87040 | — | — |
| Roseto Capo Spulico | 078107 | 87070 | — | — |
| Rota Greca | 078109 | 87010 | — | — |
| Rovito | 078110 | 87050 | — | — |
| Samo | 080070 | 89030 | — | — |
| San Basile | 078111 | 87010 | — | — |
| San Benedetto Ullano | 078112 | 87040 | — | — |
| San Calogero | 102032 | 89842 | — | — |
| San Cosmo Albanese | 078113 | 87060 | — | — |
| San Costantino Calabro | 102033 | 89851 | — | — |
| San Demetrio Corone | 078114 | 87069 | — | — |
| San Donato di Ninea | 078115 | 87010 | — | — |
| San Ferdinando | 080097 | 89026 | — | — |
| San Fili | 078116 | 87037 | — | — |
| San Floro | 079108 | 88021 | — | — |
| San Giorgio Albanese | 078118 | 87060 | — | — |
| San Giorgio Morgeto | 080071 | 89017 | — | — |
| San Giovanni di Gerace | 080072 | 89040 | — | — |
| San Giovanni in Fiore | 078119 | 87055 | — | — |
| San Gregorio d'Ippona | 102034 | 89853 | — | — |
| San Lorenzo | 080073 | 89069 | — | — |
| San Lorenzo Bellizzi | 078120 | 87070 | — | — |
| San Lorenzo del Vallo | 078121 | 87040 | — | — |
| San Luca | 080074 | 89030 | — | — |
| San Lucido | 078122 | 87038 | — | — |
| San Mango d'Aquino | 079110 | 88040 | — | — |
| San Marco Argentano | 078123 | 87018 | — | — |
| San Martino di Finita | 078124 | 87010 | — | — |
| San Mauro Marchesato | 101020 | 88831 | — | — |
| San Nicola Arcella | 078125 | 87020 | — | — |
| San Nicola da Crissa | 102035 | 89821 | — | — |
| San Nicola dell'Alto | 101021 | 88817 | — | — |
| San Pietro a Maida | 079114 | 88025 | — | — |
| San Pietro Apostolo | 079115 | 88040 | — | — |
| San Pietro di Caridà | 080075 | 89020 | — | — |
| San Pietro in Amantea | 078126 | 87030 | — | — |
| San Pietro in Guarano | 078127 | 87047 | — | — |
| San Procopio | 080076 | 89020 | — | — |
| San Roberto | 080077 | 89050 | — | — |
| San Sostene | 079116 | 88060 | — | — |
| San Sosti | 078128 | 87010 | — | — |
| San Vincenzo La Costa | 078135 | 87030 | — | — |
| San Vito sullo Ionio | 079122 | 88067 | — | — |
| Sangineto | 078117 | 87020 | — | — |
| Sant'Agata del Bianco | 080079 | 89030 | — | — |
| Sant'Agata di Esaro | 078131 | 87010 | — | — |
| Sant'Alessio in Aspromonte | 080080 | 89050 | — | — |
| Sant'Andrea Apostolo dello Ionio | 079118 | 88060 | — | — |
| Sant'Eufemia d'Aspromonte | 080081 | 89027 | — | — |
| Sant'Ilario dello Ionio | 080082 | 89040 | — | — |
| Sant'Onofrio | 102036 | 89843 | — | — |
| Santa Caterina Albanese | 078129 | 87010 | — | — |
| Santa Caterina dello Ionio | 079117 | 88060 | — | — |
| Santa Cristina d'Aspromonte | 080078 | 89056 | — | — |
| Santa Domenica Talao | 078130 | 87020 | — | — |
| Santa Maria del Cedro | 078132 | 87020 | — | — |
| Santa Severina | 101022 | 88832 | — | — |
| Santa Sofia d'Epiro | 078133 | 87048 | — | — |
| Santo Stefano di Rogliano | 078134 | 87056 | — | — |
| Santo Stefano in Aspromonte | 080083 | 89057 | — | — |
| Saracena | 078136 | 87010 | — | — |
| Satriano | 079123 | 88060 | — | — |
| Savelli | 101023 | 88825 | — | — |
| Scala Coeli | 078137 | 87060 | — | — |
| Scalea | 078138 | 87029 | — | — |
| Scandale | 101024 | 88831 | — | — |
| Scido | 080084 | 89010 | — | — |
| Scigliano | 078139 | 87057 | — | — |
| Scilla | 080085 | 89058 | — | — |
| Sellia | 079126 | 88050 | — | — |
| Sellia Marina | 079127 | 88050 | — | — |
| Seminara | 080086 | 89028 | — | — |
| Serra d'Aiello | 078140 | 87030 | — | — |
| Serra San Bruno | 102037 | 89822 | — | — |
| Serrastretta | 079129 | 88040 | — | — |
| Serrata | 080087 | 89020 | — | — |
| Sersale | 079130 | 88054 | — | — |
| Settingiano | 079131 | 88040 | — | — |
| Siderno | 080088 | 89048 | — | — |
| Simbario | 102038 | 89822 | — | — |
| Simeri Crichi | 079133 | 88050 | — | — |
| Sinopoli | 080089 | 89020 | — | — |
| Sorbo San Basile | 079134 | 88050 | — | — |
| Sorianello | 102039 | 89831 | — | — |
| Soriano Calabro | 102040 | 89831 | — | — |
| Soverato | 079137 | 88068 | — | — |
| Soveria Mannelli | 079138 | 88049 | — | — |
| Soveria Simeri | 079139 | 88050 | — | — |
| Spadola | 102041 | 89822 | — | — |
| Spezzano Albanese | 078142 | 87019 | — | — |
| Spezzano della Sila | 078143 | 87058 | — | — |
| Spilinga | 102042 | 89864 | — | — |
| Squillace | 079142 | 88069 | — | — |
| Staiti | 080090 | 89030 | — | — |
| Stalettì | 079143 | 88071 | — | — |
| Stefanaconi | 102043 | 89843 | — | — |
| Stignano | 080091 | 89040 | — | — |
| Stilo | 080092 | 89049 | — | — |
| Strongoli | 101025 | 88816 | — | — |
| Tarsia | 078145 | 87040 | — | — |
| Taurianova | 080093 | 89029 | — | — |
| Taverna | 079146 | 88055 | — | — |
| Terranova da Sibari | 078146 | 87010 | — | — |
| Terranova Sappo Minulio | 080094 | 89010 | — | — |
| Terravecchia | 078147 | 87060 | — | — |
| Tiriolo | 079147 | 88056 | — | — |
| Torano Castello | 078148 | 87010 | — | — |
| Torre di Ruggiero | 079148 | 88060 | — | — |
| Tortora | 078149 | 87020 | — | — |
| Trebisacce | 078150 | 87075 | — | — |
| Tropea | 102044 | 89861 | — | — |
| Umbriatico | 101026 | 88823 | — | — |
| Vaccarizzo Albanese | 078152 | 87060 | — | — |
| Vallefiorita | 079151 | 88050 | — | — |
| Vallelonga | 102045 | 89821 | — | — |
| Varapodio | 080095 | 89010 | — | — |
| Vazzano | 102046 | 89834 | — | — |
| Verbicaro | 078153 | 87020 | — | — |
| Verzino | 101027 | 88819 | — | — |
| Vibo Valentia | 102047 | 89900 | — | — |
| Villa San Giovanni | 080096 | 89018 | — | — |
| Villapiana | 078154 | 87076 | — | — |
| Zaccanopoli | 102048 | 89867 | — | — |
| Zagarise | 079157 | 88050 | — | — |
| Zambrone | 102049 | 89868 | — | — |
| Zumpano | 078155 | 87040 | — | — |
| Zungri | 102050 | 89867 | — | — |

<a id="region-15"></a>

### Campania (550 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Campania — region** | 15 | — | Not started | Regional sources not checked |
| Acerno | 065001 | 84042 | — | — |
| Acerra | 063001 | 80011 | — | — |
| Afragola | 063002 | 80021 | — | — |
| Agerola | 063003 | 80051 | — | — |
| Agropoli | 065002 | 84043 | — | — |
| Aiello del Sabato | 064001 | 83020 | — | — |
| Ailano | 061001 | 81010 | — | — |
| Airola | 062001 | 82011 | — | — |
| Albanella | 065003 | 84044 | — | — |
| Alfano | 065004 | 84040 | — | — |
| Alife | 061002 | 81011 | — | — |
| Altavilla Irpina | 064002 | 83011 | — | — |
| Altavilla Silentina | 065005 | 84045 | — | — |
| Alvignano | 061003 | 81012 | — | — |
| Amalfi | 065006 | 84011 | — | — |
| Amorosi | 062002 | 82031 | — | — |
| Anacapri | 063004 | 80071 | — | — |
| Andretta | 064003 | 83040 | — | — |
| Angri | 065007 | 84012 | — | — |
| Apice | 062003 | 82021 | — | — |
| Apollosa | 062004 | 82030 | — | — |
| Aquara | 065008 | 84020 | — | — |
| Aquilonia | 064004 | 83041 | — | — |
| Ariano Irpino | 064005 | 83031 | — | — |
| Arienzo | 061004 | 81021 | — | — |
| Arpaia | 062005 | 82011 | — | — |
| Arpaise | 062006 | 82010 | — | — |
| Arzano | 063005 | 80022 | — | — |
| Ascea | 065009 | 84046 | — | — |
| Atena Lucana | 065010 | 84030 | — | — |
| Atrani | 065011 | 84010 | — | — |
| Atripalda | 064006 | 83042 | — | — |
| Auletta | 065012 | 84031 | — | — |
| Avella | 064007 | 83021 | — | — |
| Avellino | 064008 | 83100 | — | — |
| Aversa | 061005 | 81031 | — | — |
| Bacoli | 063006 | 80070 | — | — |
| Bagnoli Irpino | 064009 | 83043 | — | — |
| Baia e Latina | 061006 | 81010 | — | — |
| Baiano | 064010 | 83022 | — | — |
| Barano d'Ischia | 063007 | 80072 | — | — |
| Baronissi | 065013 | 84081 | — | — |
| Baselice | 062007 | 82020 | — | — |
| Battipaglia | 065014 | 84091 | — | — |
| Bellizzi | 065158 | 84092 | — | — |
| Bellona | 061007 | 81041 | — | — |
| Bellosguardo | 065015 | 84020 | — | — |
| Benevento | 062008 | 82100 | — | — |
| Bisaccia | 064011 | 83044 | — | — |
| Bonea | 062009 | 82013 | — | — |
| Bonito | 064012 | 83032 | — | — |
| Boscoreale | 063008 | 80041 | — | — |
| Boscotrecase | 063009 | 80042 | — | — |
| Bracigliano | 065016 | 84082 | — | — |
| Brusciano | 063010 | 80031 | — | — |
| Bucciano | 062010 | 82010 | — | — |
| Buccino | 065017 | 84021 | — | — |
| Buonabitacolo | 065018 | 84032 | — | — |
| Buonalbergo | 062011 | 82020 | — | — |
| Caggiano | 065019 | 84030 | — | — |
| Caianello | 061008 | 81059 | — | — |
| Caiazzo | 061009 | 81013 | — | — |
| Cairano | 064013 | 83040 | — | — |
| Caivano | 063011 | 80023 | — | — |
| Calabritto | 064014 | 83040 | — | — |
| Calitri | 064015 | 83045 | — | — |
| Calvanico | 065020 | 84080 | — | — |
| Calvi | 062012 | 82018 | — | — |
| Calvi Risorta | 061010 | 81042 | — | — |
| Calvizzano | 063012 | 80012 | — | — |
| Camerota | 065021 | 84059 | — | — |
| Camigliano | 061011 | 81050 | — | — |
| Campagna | 065022 | 84022 | — | — |
| Campolattaro | 062013 | 82020 | — | — |
| Campoli del Monte Taburno | 062014 | 82030 | — | — |
| Campora | 065023 | 84040 | — | — |
| Camposano | 063013 | 80030 | — | — |
| Cancello ed Arnone | 061012 | 81030 | — | — |
| Candida | 064016 | 83040 | — | — |
| Cannalonga | 065024 | 84040 | — | — |
| Capaccio Paestum | 065025 | 84047 | — | — |
| Capodrise | 061013 | 81020 | — | — |
| Caposele | 064017 | 83040 | — | — |
| Capri | 063014 | 80073 | — | — |
| Capriati a Volturno | 061014 | 81014 | — | — |
| Capriglia Irpina | 064018 | 83010 | — | — |
| Capua | 061015 | 81043 | — | — |
| Carbonara di Nola | 063015 | 80030 | — | — |
| Cardito | 063016 | 80024 | — | — |
| Carife | 064019 | 83040 | — | — |
| Carinaro | 061016 | 81032 | — | — |
| Carinola | 061017 | 81030 | — | — |
| Casagiove | 061018 | 81022 | — | — |
| Casal di Principe | 061019 | 81033 | — | — |
| Casal Velino | 065028 | 84040 | — | — |
| Casalbore | 064020 | 83034 | — | — |
| Casalbuono | 065026 | 84030 | — | — |
| Casalduni | 062015 | 82027 | — | — |
| Casaletto Spartano | 065027 | 84030 | — | — |
| Casalnuovo di Napoli | 063017 | 80013 | — | — |
| Casaluce | 061020 | 81030 | — | — |
| Casamarciano | 063018 | 80032 | — | — |
| Casamicciola Terme | 063019 | 80074 | — | — |
| Casandrino | 063020 | 80025 | — | — |
| Casapesenna | 061103 | 81030 | — | — |
| Casapulla | 061021 | 81020 | — | — |
| Casavatore | 063021 | 80020 | — | — |
| Caselle in Pittari | 065029 | 84030 | — | — |
| Caserta | 061022 | 81100 | — | — |
| Casola di Napoli | 063022 | 80050 | — | — |
| Casoria | 063023 | 80026 | — | — |
| Cassano Irpino | 064021 | 83040 | — | — |
| Castel Baronia | 064022 | 83040 | — | — |
| Castel Campagnano | 061023 | 81010 | — | — |
| Castel di Sasso | 061024 | 81040 | — | — |
| Castel Morrone | 061026 | 81020 | — | — |
| Castel San Giorgio | 065034 | 84083 | — | — |
| Castel San Lorenzo | 065035 | 84049 | — | — |
| Castel Volturno | 061027 | 81030 | — | — |
| Castelcivita | 065030 | 84020 | — | — |
| Castelfranci | 064023 | 83040 | — | — |
| Castelfranco in Miscano | 062016 | 82022 | — | — |
| Castellabate | 065031 | 84048 | — | — |
| Castellammare di Stabia | 063024 | 80053 | — | — |
| Castello del Matese | 061025 | 81016 | — | — |
| Castello di Cisterna | 063025 | 80030 | — | — |
| Castelnuovo Cilento | 065032 | 84040 | — | — |
| Castelnuovo di Conza | 065033 | 84020 | — | — |
| Castelpagano | 062017 | 82024 | — | — |
| Castelpoto | 062018 | 82030 | — | — |
| Castelvenere | 062019 | 82037 | — | — |
| Castelvetere in Val Fortore | 062020 | 82023 | — | — |
| Castelvetere sul Calore | 064024 | 83040 | — | — |
| Castiglione del Genovesi | 065036 | 84090 | — | — |
| Cautano | 062021 | 82030 | — | — |
| Cava de' Tirreni | 065037 | 84013 | — | — |
| Celle di Bulgheria | 065038 | 84040 | — | — |
| Cellole | 061102 | 81030 | — | — |
| Centola | 065039 | 84051 | — | — |
| Ceppaloni | 062022 | 82014 | — | — |
| Ceraso | 065040 | 84052 | — | — |
| Cercola | 063026 | 80040 | — | — |
| Cerreto Sannita | 062023 | 82032 | — | — |
| Cervinara | 064025 | 83012 | — | — |
| Cervino | 061028 | 81023 | — | — |
| Cesa | 061029 | 81030 | — | — |
| Cesinali | 064026 | 83020 | — | — |
| Cetara | 065041 | 84010 | — | — |
| Chianche | 064027 | 83010 | — | — |
| Chiusano di San Domenico | 064028 | 83040 | — | — |
| Cicciano | 063027 | 80033 | — | — |
| Cicerale | 065042 | 84053 | — | — |
| Cimitile | 063028 | 80030 | — | — |
| Ciorlano | 061030 | 81010 | — | — |
| Circello | 062024 | 82020 | — | — |
| Colle Sannita | 062025 | 82024 | — | — |
| Colliano | 065043 | 84020 | — | — |
| Comiziano | 063029 | 80030 | — | — |
| Conca dei Marini | 065044 | 84010 | — | — |
| Conca della Campania | 061031 | 81044 | — | — |
| Contrada | 064029 | 83020 | — | — |
| Controne | 065045 | 84020 | — | — |
| Contursi Terme | 065046 | 84024 | — | — |
| Conza della Campania | 064030 | 83040 | — | — |
| Corbara | 065047 | 84010 | — | — |
| Corleto Monforte | 065048 | 84020 | — | — |
| Crispano | 063030 | 80020 | — | — |
| Cuccaro Vetere | 065049 | 84050 | — | — |
| Curti | 061032 | 81040 | — | — |
| Cusano Mutri | 062026 | 82033 | — | — |
| Domicella | 064031 | 83020 | — | — |
| Dragoni | 061033 | 81010 | — | — |
| Dugenta | 062027 | 82030 | — | — |
| Durazzano | 062028 | 82015 | — | — |
| Eboli | 065050 | 84025 | — | — |
| Ercolano | 063064 | 80056 | — | — |
| Faicchio | 062029 | 82030 | — | — |
| Falciano del Massico | 061101 | 81030 | — | — |
| Felitto | 065051 | 84055 | — | — |
| Fisciano | 065052 | 84084 | — | — |
| Flumeri | 064032 | 83040 | — | — |
| Foglianise | 062030 | 82030 | — | — |
| Foiano di Val Fortore | 062031 | 82020 | — | — |
| Fontanarosa | 064033 | 83040 | — | — |
| Fontegreca | 061034 | 81014 | — | — |
| Forchia | 062032 | 82011 | — | — |
| Forino | 064034 | 83020 | — | — |
| Forio | 063031 | 80075 | — | — |
| Formicola | 061035 | 81040 | — | — |
| Fragneto l'Abate | 062033 | 82020 | — | — |
| Fragneto Monforte | 062034 | 82020 | — | — |
| Francolise | 061036 | 81050 | — | — |
| Frasso Telesino | 062035 | 82030 | — | — |
| Frattamaggiore | 063032 | 80027 | — | — |
| Frattaminore | 063033 | 80020 | — | — |
| Frigento | 064035 | 83040 | — | — |
| Frignano | 061037 | 81030 | — | — |
| Furore | 065053 | 84010 | — | — |
| Futani | 065054 | 84050 | — | — |
| Gallo Matese | 061038 | 81010 | — | — |
| Galluccio | 061039 | 81044 | — | — |
| Gesualdo | 064036 | 83040 | — | — |
| Giano Vetusto | 061040 | 81042 | — | — |
| Giffoni Sei Casali | 065055 | 84090 | — | — |
| Giffoni Valle Piana | 065056 | 84095 | — | — |
| Ginestra degli Schiavoni | 062036 | 82020 | — | — |
| Gioi | 065057 | 84056 | — | — |
| Gioia Sannitica | 061041 | 81010 | — | — |
| Giugliano in Campania | 063034 | 80014 | — | — |
| Giungano | 065058 | 84050 | — | — |
| Gragnano | 063035 | 80054 | — | — |
| Grazzanise | 061042 | 81046 | — | — |
| Greci | 064037 | 83030 | — | — |
| Gricignano di Aversa | 061043 | 81030 | — | — |
| Grottaminarda | 064038 | 83035 | — | — |
| Grottolella | 064039 | 83010 | — | — |
| Grumo Nevano | 063036 | 80028 | — | — |
| Guardia Lombardi | 064040 | 83040 | — | — |
| Guardia Sanframondi | 062037 | 82034 | — | — |
| Ischia | 063037 | 80077 | — | — |
| Ispani | 065059 | 84050 | — | — |
| Lacco Ameno | 063038 | 80076 | — | — |
| Lacedonia | 064041 | 83046 | — | — |
| Lapio | 064042 | 83030 | — | — |
| Laureana Cilento | 065060 | 84050 | — | — |
| Laurino | 065061 | 84057 | — | — |
| Laurito | 065062 | 84050 | — | — |
| Lauro | 064043 | 83023 | — | — |
| Laviano | 065063 | 84020 | — | — |
| Letino | 061044 | 81010 | — | — |
| Lettere | 063039 | 80050 | — | — |
| Liberi | 061045 | 81040 | — | — |
| Limatola | 062038 | 82030 | — | — |
| Lioni | 064044 | 83047 | — | — |
| Liveri | 063040 | 80030 | — | — |
| Luogosano | 064045 | 83040 | — | — |
| Lusciano | 061046 | 81030 | — | — |
| Lustra | 065064 | 84050 | — | — |
| Macerata Campania | 061047 | 81047 | — | — |
| Maddaloni | 061048 | 81024 | — | — |
| Magliano Vetere | 065065 | 84050 | — | — |
| Maiori | 065066 | 84010 | — | — |
| Manocalzati | 064046 | 83030 | — | — |
| Marano di Napoli | 063041 | 80016 | — | — |
| Marcianise | 061049 | 81025 | — | — |
| Mariglianella | 063042 | 80030 | — | — |
| Marigliano | 063043 | 80034 | — | — |
| Marzano Appio | 061050 | 81035 | — | — |
| Marzano di Nola | 064047 | 83020 | — | — |
| Massa di Somma | 063092 | 80040 | — | — |
| Massa Lubrense | 063044 | 80061 | — | — |
| Melito di Napoli | 063045 | 80017 | — | — |
| Melito Irpino | 064048 | 83030 | — | — |
| Melizzano | 062039 | 82030 | — | — |
| Mercato San Severino | 065067 | 84085 | — | — |
| Mercogliano | 064049 | 83013 | — | — |
| Meta | 063046 | 80062 | — | — |
| Mignano Monte Lungo | 061051 | 81049 | — | — |
| Minori | 065068 | 84010 | — | — |
| Mirabella Eclano | 064050 | 83036 | — | — |
| Moiano | 062040 | 82010 | — | — |
| Moio della Civitella | 065069 | 84060 | — | — |
| Molinara | 062041 | 82020 | — | — |
| Mondragone | 061052 | 81034 | — | — |
| Montaguto | 064051 | 83030 | — | — |
| Montano Antilia | 065070 | 84060 | — | — |
| Monte di Procida | 063047 | 80070 | — | — |
| Monte San Giacomo | 065075 | 84030 | — | — |
| Montecalvo Irpino | 064052 | 83037 | — | — |
| Montecorice | 065071 | 84060 | — | — |
| Montecorvino Pugliano | 065072 | 84090 | — | — |
| Montecorvino Rovella | 065073 | 84096 | — | — |
| Montefalcione | 064053 | 83030 | — | — |
| Montefalcone di Val Fortore | 062042 | 82025 | — | — |
| Monteforte Cilento | 065074 | 84060 | — | — |
| Monteforte Irpino | 064054 | 83024 | — | — |
| Montefredane | 064055 | 83030 | — | — |
| Montefusco | 064056 | 83030 | — | — |
| Montella | 064057 | 83048 | — | — |
| Montemarano | 064058 | 83040 | — | — |
| Montemiletto | 064059 | 83038 | — | — |
| Montesano sulla Marcellana | 065076 | 84033 | — | — |
| Montesarchio | 062043 | 82016 | — | — |
| Monteverde | 064060 | 83049 | — | — |
| Montoro | 064121 | 83025 | — | — |
| Morcone | 062044 | 82026 | — | — |
| Morigerati | 065077 | 84030 | — | — |
| Morra De Sanctis | 064063 | 83040 | — | — |
| Moschiano | 064064 | 83020 | — | — |
| Mugnano del Cardinale | 064065 | 83027 | — | — |
| Mugnano di Napoli | 063048 | 80018 | — | — |
| Napoli | 063049 | 80121, 80122, 80123, 80124, 80125, 80126, 80127, 80128, 80129, 80131, 80132, 80133, 80134, 80135, 80136, 80137, 80138, 80139, 80141, 80142, 80143, 80144, 80145, 80146, 80147 | — | — |
| Nocera Inferiore | 065078 | 84014 | — | — |
| Nocera Superiore | 065079 | 84015 | — | — |
| Nola | 063050 | 80035 | — | — |
| Novi Velia | 065080 | 84060 | — | — |
| Nusco | 064066 | 83051 | — | — |
| Ogliastro Cilento | 065081 | 84061 | — | — |
| Olevano sul Tusciano | 065082 | 84062 | — | — |
| Oliveto Citra | 065083 | 84020 | — | — |
| Omignano | 065084 | 84060 | — | — |
| Orria | 065085 | 84060 | — | — |
| Orta di Atella | 061053 | 81030 | — | — |
| Ospedaletto d'Alpinolo | 064067 | 83014 | — | — |
| Ottati | 065086 | 84020 | — | — |
| Ottaviano | 063051 | 80044 | — | — |
| Padula | 065087 | 84034 | — | — |
| Paduli | 062045 | 82020 | — | — |
| Pagani | 065088 | 84016 | — | — |
| Pago del Vallo di Lauro | 064068 | 83020 | — | — |
| Pago Veiano | 062046 | 82020 | — | — |
| Palma Campania | 063052 | 80036 | — | — |
| Palomonte | 065089 | 84020 | — | — |
| Pannarano | 062047 | 82017 | — | — |
| Paolisi | 062048 | 82011 | — | — |
| Parete | 061054 | 81030 | — | — |
| Parolise | 064069 | 83050 | — | — |
| Pastorano | 061055 | 81050 | — | — |
| Paternopoli | 064070 | 83052 | — | — |
| Paupisi | 062049 | 82030 | — | — |
| Pellezzano | 065090 | 84080 | — | — |
| Perdifumo | 065091 | 84060 | — | — |
| Perito | 065092 | 84060 | — | — |
| Pertosa | 065093 | 84030 | — | — |
| Pesco Sannita | 062050 | 82020 | — | — |
| Petina | 065094 | 84020 | — | — |
| Petruro Irpino | 064071 | 83010 | — | — |
| Piaggine | 065095 | 84065 | — | — |
| Piana di Monte Verna | 061056 | 81013 | — | — |
| Piano di Sorrento | 063053 | 80063 | — | — |
| Piedimonte Matese | 061057 | 81016 | — | — |
| Pietradefusi | 064072 | 83030 | — | — |
| Pietramelara | 061058 | 81051 | — | — |
| Pietraroja | 062051 | 82030 | — | — |
| Pietrastornina | 064073 | 83015 | — | — |
| Pietravairano | 061059 | 81040 | — | — |
| Pietrelcina | 062052 | 82020 | — | — |
| Pignataro Maggiore | 061060 | 81052 | — | — |
| Pimonte | 063054 | 80050 | — | — |
| Pisciotta | 065096 | 84066 | — | — |
| Poggiomarino | 063055 | 80040 | — | — |
| Polla | 065097 | 84035 | — | — |
| Pollena Trocchia | 063056 | 80040 | — | — |
| Pollica | 065098 | 84068 | — | — |
| Pomigliano d'Arco | 063057 | 80038 | — | — |
| Pompei | 063058 | 80045 | — | — |
| Ponte | 062053 | 82030 | — | — |
| Pontecagnano Faiano | 065099 | 84098 | — | — |
| Pontelandolfo | 062054 | 82027 | — | — |
| Pontelatone | 061061 | 81040 | — | — |
| Portici | 063059 | 80055 | — | — |
| Portico di Caserta | 061062 | 81050 | — | — |
| Positano | 065100 | 84017 | — | — |
| Postiglione | 065101 | 84026 | — | — |
| Pozzuoli | 063060 | 80078 | — | — |
| Praiano | 065102 | 84010 | — | — |
| Prata di Principato Ultra | 064074 | 83030 | — | — |
| Prata Sannita | 061063 | 81010 | — | — |
| Pratella | 061064 | 81010 | — | — |
| Pratola Serra | 064075 | 83039 | — | — |
| Presenzano | 061065 | 81050 | — | — |
| Prignano Cilento | 065103 | 84060 | — | — |
| Procida | 063061 | 80079 | — | — |
| Puglianello | 062055 | 82030 | — | — |
| Quadrelle | 064076 | 83020 | — | — |
| Qualiano | 063062 | 80019 | — | — |
| Quarto | 063063 | 80010 | — | — |
| Quindici | 064077 | 83020 | — | — |
| Ravello | 065104 | 84010 | — | — |
| Raviscanina | 061066 | 81017 | — | — |
| Recale | 061067 | 81020 | — | — |
| Reino | 062056 | 82020 | — | — |
| Riardo | 061068 | 81053 | — | — |
| Ricigliano | 065105 | 84020 | — | — |
| Rocca d'Evandro | 061069 | 81040 | — | — |
| Rocca San Felice | 064079 | 83050 | — | — |
| Roccabascerana | 064078 | 83016 | — | — |
| Roccadaspide | 065106 | 84069 | — | — |
| Roccagloriosa | 065107 | 84060 | — | — |
| Roccamonfina | 061070 | 81035 | — | — |
| Roccapiemonte | 065108 | 84086 | — | — |
| Roccarainola | 063065 | 80030 | — | — |
| Roccaromana | 061071 | 81051 | — | — |
| Rocchetta e Croce | 061072 | 81042 | — | — |
| Rofrano | 065109 | 84070 | — | — |
| Romagnano al Monte | 065110 | 84020 | — | — |
| Roscigno | 065111 | 84020 | — | — |
| Rotondi | 064080 | 83017 | — | — |
| Rutino | 065112 | 84070 | — | — |
| Ruviano | 061073 | 81010 | — | — |
| Sacco | 065113 | 84070 | — | — |
| Sala Consilina | 065114 | 84036 | — | — |
| Salento | 065115 | 84070 | — | — |
| Salerno | 065116 | 84121, 84122, 84123, 84124, 84125, 84126, 84127, 84128, 84129, 84131, 84132, 84133, 84134, 84135 | — | — |
| Salvitelle | 065117 | 84020 | — | — |
| Salza Irpina | 064081 | 83050 | — | — |
| San Bartolomeo in Galdo | 062057 | 82028 | — | — |
| San Cipriano d'Aversa | 061074 | 81036 | — | — |
| San Cipriano Picentino | 065118 | 84099 | — | — |
| San Felice a Cancello | 061075 | 81027 | — | — |
| San Gennaro Vesuviano | 063066 | 80040 | — | — |
| San Giorgio a Cremano | 063067 | 80046 | — | — |
| San Giorgio del Sannio | 062058 | 82018 | — | — |
| San Giorgio La Molara | 062059 | 82020 | — | — |
| San Giovanni a Piro | 065119 | 84070 | — | — |
| San Giuseppe Vesuviano | 063068 | 80047 | — | — |
| San Gregorio Magno | 065120 | 84020 | — | — |
| San Gregorio Matese | 061076 | 81010 | — | — |
| San Leucio del Sannio | 062060 | 82010 | — | — |
| San Lorenzello | 062061 | 82030 | — | — |
| San Lorenzo Maggiore | 062062 | 82034 | — | — |
| San Lupo | 062063 | 82034 | — | — |
| San Mango Piemonte | 065121 | 84090 | — | — |
| San Mango sul Calore | 064082 | 83050 | — | — |
| San Marcellino | 061077 | 81030 | — | — |
| San Marco dei Cavoti | 062064 | 82029 | — | — |
| San Marco Evangelista | 061104 | 81020 | — | — |
| San Martino Sannita | 062065 | 82010 | — | — |
| San Martino Valle Caudina | 064083 | 83018 | — | — |
| San Marzano sul Sarno | 065122 | 84010 | — | — |
| San Mauro Cilento | 065123 | 84070 | — | — |
| San Mauro la Bruca | 065124 | 84070 | — | — |
| San Michele di Serino | 064084 | 83020 | — | — |
| San Nazzaro | 062066 | 82018 | — | — |
| San Nicola Baronia | 064085 | 83050 | — | — |
| San Nicola la Strada | 061078 | 81020 | — | — |
| San Nicola Manfredi | 062067 | 82010 | — | — |
| San Paolo Bel Sito | 063069 | 80030 | — | — |
| San Pietro al Tanagro | 065125 | 84030 | — | — |
| San Pietro Infine | 061079 | 81049 | — | — |
| San Potito Sannitico | 061080 | 81016 | — | — |
| San Potito Ultra | 064086 | 83050 | — | — |
| San Prisco | 061081 | 81054 | — | — |
| San Rufo | 065126 | 84030 | — | — |
| San Salvatore Telesino | 062068 | 82030 | — | — |
| San Sebastiano al Vesuvio | 063070 | 80040 | — | — |
| San Sossio Baronia | 064087 | 83050 | — | — |
| San Tammaro | 061085 | 81050 | — | — |
| San Valentino Torio | 065132 | 84010 | — | — |
| San Vitaliano | 063075 | 80030 | — | — |
| Sant'Agata de' Goti | 062070 | 82019 | — | — |
| Sant'Agnello | 063071 | 80065 | — | — |
| Sant'Anastasia | 063072 | 80048 | — | — |
| Sant'Andrea di Conza | 064089 | 83053 | — | — |
| Sant'Angelo a Cupolo | 062071 | 82010 | — | — |
| Sant'Angelo a Fasanella | 065128 | 84027 | — | — |
| Sant'Angelo a Scala | 064091 | 83010 | — | — |
| Sant'Angelo all'Esca | 064090 | 83050 | — | — |
| Sant'Angelo d'Alife | 061086 | 81017 | — | — |
| Sant'Angelo dei Lombardi | 064092 | 83054 | — | — |
| Sant'Antimo | 063073 | 80029 | — | — |
| Sant'Antonio Abate | 063074 | 80057 | — | — |
| Sant'Arcangelo Trimonte | 062078 | 82021 | — | — |
| Sant'Arpino | 061087 | 81030 | — | — |
| Sant'Arsenio | 065129 | 84037 | — | — |
| Sant'Egidio del Monte Albino | 065130 | 84010 | — | — |
| Santa Croce del Sannio | 062069 | 82020 | — | — |
| Santa Lucia di Serino | 064088 | 83020 | — | — |
| Santa Maria a Vico | 061082 | 81028 | — | — |
| Santa Maria Capua Vetere | 061083 | 81055 | — | — |
| Santa Maria la Carità | 063090 | 80050 | — | — |
| Santa Maria la Fossa | 061084 | 81050 | — | — |
| Santa Marina | 065127 | 84067 | — | — |
| Santa Paolina | 064093 | 83030 | — | — |
| Santo Stefano del Sole | 064095 | 83050 | — | — |
| Santomenna | 065131 | 84020 | — | — |
| Sanza | 065133 | 84030 | — | — |
| Sapri | 065134 | 84073 | — | — |
| Sarno | 065135 | 84087 | — | — |
| Sassano | 065136 | 84038 | — | — |
| Sassinoro | 062072 | 82026 | — | — |
| Saviano | 063076 | 80039 | — | — |
| Savignano Irpino | 064096 | 83030 | — | — |
| Scafati | 065137 | 84018 | — | — |
| Scala | 065138 | 84010 | — | — |
| Scampitella | 064097 | 83050 | — | — |
| Scisciano | 063077 | 80030 | — | — |
| Senerchia | 064098 | 83050 | — | — |
| Serino | 064099 | 83028 | — | — |
| Serramezzana | 065139 | 84070 | — | — |
| Serrara Fontana | 063078 | 80081 | — | — |
| Serre | 065140 | 84028 | — | — |
| Sessa Aurunca | 061088 | 81037 | — | — |
| Sessa Cilento | 065141 | 84074 | — | — |
| Siano | 065142 | 84088 | — | — |
| Sicignano degli Alburni | 065143 | 84029 | — | — |
| Sirignano | 064100 | 83020 | — | — |
| Solofra | 064101 | 83029 | — | — |
| Solopaca | 062073 | 82036 | — | — |
| Somma Vesuviana | 063079 | 80049 | — | — |
| Sorbo Serpico | 064102 | 83050 | — | — |
| Sorrento | 063080 | 80067 | — | — |
| Sparanise | 061089 | 81056 | — | — |
| Sperone | 064103 | 83020 | — | — |
| Stella Cilento | 065144 | 84070 | — | — |
| Stio | 065145 | 84075 | — | — |
| Striano | 063081 | 80040 | — | — |
| Sturno | 064104 | 83055 | — | — |
| Succivo | 061090 | 81030 | — | — |
| Summonte | 064105 | 83010 | — | — |
| Taurano | 064106 | 83020 | — | — |
| Taurasi | 064107 | 83030 | — | — |
| Teano | 061091 | 81057 | — | — |
| Teggiano | 065146 | 84039 | — | — |
| Telese Terme | 062074 | 82037 | — | — |
| Teora | 064108 | 83056 | — | — |
| Terzigno | 063082 | 80040 | — | — |
| Teverola | 061092 | 81030 | — | — |
| Tocco Caudio | 062075 | 82030 | — | — |
| Tora e Piccilli | 061093 | 81044 | — | — |
| Torchiara | 065147 | 84076 | — | — |
| Torella dei Lombardi | 064109 | 83057 | — | — |
| Torraca | 065148 | 84030 | — | — |
| Torre Annunziata | 063083 | 80058 | — | — |
| Torre del Greco | 063084 | 80059 | — | — |
| Torre Le Nocelle | 064110 | 83030 | — | — |
| Torre Orsaia | 065149 | 84077 | — | — |
| Torrecuso | 062076 | 82030 | — | — |
| Torrioni | 064111 | 83010 | — | — |
| Tortorella | 065150 | 84030 | — | — |
| Tramonti | 065151 | 84010 | — | — |
| Trecase | 063091 | 80040 | — | — |
| Trentinara | 065152 | 84070 | — | — |
| Trentola Ducenta | 061094 | 81038 | — | — |
| Trevico | 064112 | 83058 | — | — |
| Tufino | 063085 | 80030 | — | — |
| Tufo | 064113 | 83010 | — | — |
| Vairano Patenora | 061095 | 81058 | — | — |
| Vallata | 064114 | 83059 | — | — |
| Valle Agricola | 061096 | 81010 | — | — |
| Valle dell'Angelo | 065153 | 84070 | — | — |
| Valle di Maddaloni | 061097 | 81020 | — | — |
| Vallesaccarda | 064115 | 83050 | — | — |
| Vallo della Lucania | 065154 | 84078 | — | — |
| Valva | 065155 | 84020 | — | — |
| Venticano | 064116 | 83030 | — | — |
| Vibonati | 065156 | 84079 | — | — |
| Vico Equense | 063086 | 80069 | — | — |
| Vietri sul Mare | 065157 | 84019 | — | — |
| Villa di Briano | 061098 | 81030 | — | — |
| Villa Literno | 061099 | 81039 | — | — |
| Villamaina | 064117 | 83050 | — | — |
| Villanova del Battista | 064118 | 83030 | — | — |
| Villaricca | 063087 | 80010 | — | — |
| Visciano | 063088 | 80030 | — | — |
| Vitulano | 062077 | 82038 | — | — |
| Vitulazio | 061100 | 81041 | — | — |
| Volla | 063089 | 80040 | — | — |
| Volturara Irpina | 064119 | 83050 | — | — |
| Zungoli | 064120 | 83030 | — | — |

<a id="region-08"></a>

### Emilia-Romagna (330 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Emilia-Romagna — region** | 08 | — | Not started | Regional sources not checked |
| Agazzano | 033001 | 29010 | — | — |
| Albareto | 034001 | 43051 | — | — |
| Albinea | 035001 | 42020 | — | — |
| Alfonsine | 039001 | 48011 | — | — |
| Alseno | 033002 | 29010 | — | — |
| Alta Val Tidone | 033049 | 29031 | — | — |
| Alto Reno Terme | 037062 | 40046 | — | — |
| Anzola dell'Emilia | 037001 | 40011 | — | — |
| Argelato | 037002 | 40050 | — | — |
| Argenta | 038001 | 44011 | — | — |
| Bagnacavallo | 039002 | 48012 | — | — |
| Bagnara di Romagna | 039003 | 48031 | — | — |
| Bagno di Romagna | 040001 | 47021 | — | — |
| Bagnolo in Piano | 035002 | 42011 | — | — |
| Baiso | 035003 | 42031 | — | — |
| Bardi | 034002 | 43032 | — | — |
| Baricella | 037003 | 40052 | — | — |
| Bastiglia | 036001 | 41030 | — | — |
| Bedonia | 034003 | 43041 | — | — |
| Bellaria-Igea Marina | 099001 | 47814 | — | — |
| Bentivoglio | 037005 | 40010 | — | — |
| Berceto | 034004 | 43042 | — | — |
| Bertinoro | 040003 | 47032 | — | — |
| Besenzone | 033003 | 29010 | — | — |
| Bettola | 033004 | 29021 | — | — |
| Bibbiano | 035004 | 42021 | — | — |
| Bobbio | 033005 | 29022 | — | — |
| Bologna | 037006 | 40121, 40122, 40123, 40124, 40125, 40126, 40127, 40128, 40129, 40131, 40132, 40133, 40134, 40135, 40136, 40137, 40138, 40139, 40141 | — | — |
| Bomporto | 036002 | 41030 | — | — |
| Bondeno | 038003 | 44012 | — | — |
| Bore | 034005 | 43030 | — | — |
| Boretto | 035005 | 42022 | — | — |
| Borghi | 040004 | 47030 | — | — |
| Borgo Tossignano | 037007 | 40021 | — | — |
| Borgo Val di Taro | 034006 | 43043 | — | — |
| Borgonovo Val Tidone | 033006 | 29011 | — | — |
| Brescello | 035006 | 42041 | — | — |
| Brisighella | 039004 | 48013 | — | — |
| Budrio | 037008 | 40054 | — | — |
| Busseto | 034007 | 43011 | — | — |
| Cadelbosco di Sopra | 035008 | 42023 | — | — |
| Cadeo | 033007 | 29010 | — | — |
| Calderara di Reno | 037009 | 40012 | — | — |
| Calendasco | 033008 | 29010 | — | — |
| Calestano | 034008 | 43030 | — | — |
| Campagnola Emilia | 035009 | 42012 | — | — |
| Campegine | 035010 | 42040 | — | — |
| Campogalliano | 036003 | 41011 | — | — |
| Camposanto | 036004 | 41031 | — | — |
| Camugnano | 037010 | 40032 | — | — |
| Canossa | 035018 | 42026 | — | — |
| Caorso | 033010 | 29012 | — | — |
| Carpaneto Piacentino | 033011 | 29013 | — | — |
| Carpi | 036005 | 41012 | — | — |
| Carpineti | 035011 | 42033 | — | — |
| Casalecchio di Reno | 037011 | 40033 | — | — |
| Casalfiumanese | 037012 | 40020 | — | — |
| Casalgrande | 035012 | 42013 | — | — |
| Casina | 035013 | 42034 | — | — |
| Casola Valsenio | 039005 | 48032 | — | — |
| Castel Bolognese | 039006 | 48014 | — | — |
| Castel d'Aiano | 037013 | 40034 | — | — |
| Castel del Rio | 037014 | 40022 | — | — |
| Castel di Casio | 037015 | 40030 | — | — |
| Castel Guelfo di Bologna | 037016 | 40023 | — | — |
| Castel Maggiore | 037019 | 40013 | — | — |
| Castel San Giovanni | 033013 | 29015 | — | — |
| Castel San Pietro Terme | 037020 | 40024 | — | — |
| Casteldelci | 099021 | 47861 | — | — |
| Castelfranco Emilia | 036006 | 41013 | — | — |
| Castell'Arquato | 033012 | 29014 | — | — |
| Castellarano | 035014 | 42014 | — | — |
| Castello d'Argile | 037017 | 40050 | — | — |
| Castelnovo di Sotto | 035015 | 42024 | — | — |
| Castelnovo ne' Monti | 035016 | 42035 | — | — |
| Castelnuovo Rangone | 036007 | 41051 | — | — |
| Castelvetro di Modena | 036008 | 41014 | — | — |
| Castelvetro Piacentino | 033014 | 29010 | — | — |
| Castenaso | 037021 | 40055 | — | — |
| Castiglione dei Pepoli | 037022 | 40035 | — | — |
| Castrocaro Terme e Terra del Sole | 040005 | 47011 | — | — |
| Cattolica | 099002 | 47841 | — | — |
| Cavezzo | 036009 | 41032 | — | — |
| Cavriago | 035017 | 42025 | — | — |
| Cento | 038004 | 44042 | — | — |
| Cerignale | 033015 | 29020 | — | — |
| Cervia | 039007 | 48015 | — | — |
| Cesena | 040007 | 47521, 47522 | — | — |
| Cesenatico | 040008 | 47042 | — | — |
| Civitella di Romagna | 040009 | 47012 | — | — |
| Codigoro | 038005 | 44021 | — | — |
| Coli | 033016 | 29020 | — | — |
| Collecchio | 034009 | 43044 | — | — |
| Colorno | 034010 | 43052 | — | — |
| Comacchio | 038006 | 44022 | — | — |
| Compiano | 034011 | 43053 | — | — |
| Concordia sulla Secchia | 036010 | 41033 | — | — |
| Conselice | 039008 | 48017 | — | — |
| Copparo | 038007 | 44034 | — | — |
| Coriano | 099003 | 47853 | — | — |
| Corniglio | 034012 | 43021 | — | — |
| Correggio | 035020 | 42015 | — | — |
| Corte Brugnatella | 033017 | 29020 | — | — |
| Cortemaggiore | 033018 | 29016 | — | — |
| Cotignola | 039009 | 48033 | — | — |
| Crevalcore | 037024 | 40014 | — | — |
| Dovadola | 040011 | 47013 | — | — |
| Dozza | 037025 | 40060 | — | — |
| Fabbrico | 035021 | 42042 | — | — |
| Faenza | 039010 | 48018 | — | — |
| Fanano | 036011 | 41021 | — | — |
| Farini | 033019 | 29023 | — | — |
| Felino | 034013 | 43035 | — | — |
| Ferrara | 038008 | 44121, 44122, 44123, 44124 | — | — |
| Ferriere | 033020 | 29024 | — | — |
| Fidenza | 034014 | 43036 | — | — |
| Finale Emilia | 036012 | 41034 | — | — |
| Fiorano Modenese | 036013 | 41042 | — | — |
| Fiorenzuola d'Arda | 033021 | 29017 | — | — |
| Fiscaglia | 038027 | 44027 | — | — |
| Fiumalbo | 036014 | 41022 | — | — |
| Fontanelice | 037026 | 40025 | — | — |
| Fontanellato | 034015 | 43012 | — | — |
| Fontevivo | 034016 | 43010 | — | — |
| Forlì | 040012 | 47121, 47122 | — | — |
| Forlimpopoli | 040013 | 47034 | — | — |
| Formigine | 036015 | 41043 | — | — |
| Fornovo di Taro | 034017 | 43045 | — | — |
| Frassinoro | 036016 | 41044 | — | — |
| Fusignano | 039011 | 48034 | — | — |
| Gaggio Montano | 037027 | 40041 | — | — |
| Galeata | 040014 | 47010 | — | — |
| Galliera | 037028 | 40015 | — | — |
| Gambettola | 040015 | 47035 | — | — |
| Gattatico | 035022 | 42043 | — | — |
| Gatteo | 040016 | 47043 | — | — |
| Gazzola | 033022 | 29010 | — | — |
| Gemmano | 099004 | 47855 | — | — |
| Goro | 038025 | 44020 | — | — |
| Gossolengo | 033023 | 29020 | — | — |
| Gragnano Trebbiense | 033024 | 29010 | — | — |
| Granarolo dell'Emilia | 037030 | 40057 | — | — |
| Grizzana Morandi | 037031 | 40030 | — | — |
| Gropparello | 033025 | 29025 | — | — |
| Gualtieri | 035023 | 42044 | — | — |
| Guastalla | 035024 | 42016 | — | — |
| Guiglia | 036017 | 41052 | — | — |
| Imola | 037032 | 40026 | — | — |
| Jolanda di Savoia | 038010 | 44037 | — | — |
| Lagosanto | 038011 | 44023 | — | — |
| Lama Mocogno | 036018 | 41023 | — | — |
| Langhirano | 034018 | 43013 | — | — |
| Lesignano de' Bagni | 034019 | 43037 | — | — |
| Lizzano in Belvedere | 037033 | 40042 | — | — |
| Loiano | 037034 | 40050 | — | — |
| Longiano | 040018 | 47020 | — | — |
| Lugagnano Val d'Arda | 033026 | 29018 | — | — |
| Lugo | 039012 | 48022 | — | — |
| Luzzara | 035026 | 42045 | — | — |
| Maiolo | 099022 | 47862 | — | — |
| Malalbergo | 037035 | 40051 | — | — |
| Maranello | 036019 | 41053 | — | — |
| Marano sul Panaro | 036020 | 41054 | — | — |
| Marzabotto | 037036 | 40043 | — | — |
| Masi Torello | 038012 | 44020 | — | — |
| Massa Lombarda | 039013 | 48024 | — | — |
| Medesano | 034020 | 43014 | — | — |
| Medicina | 037037 | 40059 | — | — |
| Medolla | 036021 | 41036 | — | — |
| Meldola | 040019 | 47014 | — | — |
| Mercato Saraceno | 040020 | 47025 | — | — |
| Mesola | 038014 | 44026 | — | — |
| Minerbio | 037038 | 40061 | — | — |
| Mirandola | 036022 | 41037 | — | — |
| Misano Adriatico | 099005 | 47843 | — | — |
| Modena | 036023 | 41121, 41122, 41123, 41124, 41125, 41126 | — | — |
| Modigliana | 040022 | 47015 | — | — |
| Molinella | 037039 | 40062 | — | — |
| Monchio delle Corti | 034022 | 43010 | — | — |
| Mondaino | 099006 | 47836 | — | — |
| Monghidoro | 037040 | 40063 | — | — |
| Monte San Pietro | 037042 | 40050 | — | — |
| Montecchio Emilia | 035027 | 42027 | — | — |
| Montechiarugolo | 034023 | 43022 | — | — |
| Montecopiolo | 099030 | 47868 | — | — |
| Montecreto | 036024 | 41025 | — | — |
| Montefiore Conca | 099008 | 47834 | — | — |
| Montefiorino | 036025 | 41045 | — | — |
| Montegridolfo | 099009 | 47837 | — | — |
| Monterenzio | 037041 | 40050 | — | — |
| Montescudo-Monte Colombo | 099029 | 47854 | — | — |
| Montese | 036026 | 41055 | — | — |
| Montiano | 040028 | 47020 | — | — |
| Monticelli d'Ongina | 033027 | 29010 | — | — |
| Monzuno | 037044 | 40036 | — | — |
| Morciano di Romagna | 099011 | 47833 | — | — |
| Mordano | 037045 | 40027 | — | — |
| Morfasso | 033028 | 29020 | — | — |
| Neviano degli Arduini | 034024 | 43024 | — | — |
| Noceto | 034025 | 43015 | — | — |
| Nonantola | 036027 | 41015 | — | — |
| Novafeltria | 099023 | 47863 | — | — |
| Novellara | 035028 | 42017 | — | — |
| Novi di Modena | 036028 | 41016 | — | — |
| Ostellato | 038017 | 44020 | — | — |
| Ottone | 033030 | 29026 | — | — |
| Ozzano dell'Emilia | 037046 | 40064 | — | — |
| Palagano | 036029 | 41046 | — | — |
| Palanzano | 034026 | 43025 | — | — |
| Parma | 034027 | 43121, 43122, 43123, 43124, 43125, 43126 | — | — |
| Pavullo nel Frignano | 036030 | 41026 | — | — |
| Pellegrino Parmense | 034028 | 43047 | — | — |
| Pennabilli | 099024 | 47864 | — | — |
| Piacenza | 033032 | 29121, 29122 | — | — |
| Pianello Val Tidone | 033033 | 29010 | — | — |
| Pianoro | 037047 | 40065 | — | — |
| Pieve di Cento | 037048 | 40066 | — | — |
| Pievepelago | 036031 | 41027 | — | — |
| Piozzano | 033034 | 29010 | — | — |
| Podenzano | 033035 | 29027 | — | — |
| Poggio Renatico | 038018 | 44028 | — | — |
| Poggio Torriana | 099028 | 47824 | — | — |
| Polesine Zibello | 034050 | 43016 | — | — |
| Polinago | 036032 | 41040 | — | — |
| Ponte dell'Olio | 033036 | 29028 | — | — |
| Pontenure | 033037 | 29010 | — | — |
| Portico e San Benedetto | 040031 | 47010 | — | — |
| Portomaggiore | 038019 | 44015 | — | — |
| Poviglio | 035029 | 42028 | — | — |
| Predappio | 040032 | 47016 | — | — |
| Premilcuore | 040033 | 47010 | — | — |
| Prignano sulla Secchia | 036033 | 41048 | — | — |
| Quattro Castella | 035030 | 42020 | — | — |
| Ravarino | 036034 | 41017 | — | — |
| Ravenna | 039014 | 48121, 48122, 48123, 48124, 48125 | — | — |
| Reggio nell'Emilia | 035033 | 42121, 42122, 42123, 42124 | — | — |
| Reggiolo | 035032 | 42046 | — | — |
| Riccione | 099013 | 47838 | — | — |
| Rimini | 099014 | 47921, 47922, 47923, 47924 | — | — |
| Rio Saliceto | 035034 | 42010 | — | — |
| Riolo Terme | 039015 | 48025 | — | — |
| Riolunato | 036035 | 41020 | — | — |
| Riva del Po | 038029 | 44033 | — | — |
| Rivergaro | 033038 | 29029 | — | — |
| Rocca San Casciano | 040036 | 47017 | — | — |
| Roccabianca | 034030 | 43010 | — | — |
| Rolo | 035035 | 42047 | — | — |
| Roncofreddo | 040037 | 47020 | — | — |
| Rottofreno | 033039 | 29010 | — | — |
| Rubiera | 035036 | 42048 | — | — |
| Russi | 039016 | 48026 | — | — |
| Sala Baganza | 034031 | 43038 | — | — |
| Sala Bolognese | 037050 | 40010 | — | — |
| Salsomaggiore Terme | 034032 | 43039 | — | — |
| Saludecio | 099015 | 47835 | — | — |
| San Benedetto Val di Sambro | 037051 | 40048 | — | — |
| San Cesario sul Panaro | 036036 | 41018 | — | — |
| San Clemente | 099016 | 47832 | — | — |
| San Felice sul Panaro | 036037 | 41038 | — | — |
| San Giorgio di Piano | 037052 | 40016 | — | — |
| San Giorgio Piacentino | 033040 | 29019 | — | — |
| San Giovanni in Marignano | 099017 | 47842 | — | — |
| San Giovanni in Persiceto | 037053 | 40017 | — | — |
| San Lazzaro di Savena | 037054 | 40068 | — | — |
| San Leo | 099025 | 47865 | — | — |
| San Martino in Rio | 035037 | 42018 | — | — |
| San Mauro Pascoli | 040041 | 47030 | — | — |
| San Pietro in Casale | 037055 | 40018 | — | — |
| San Pietro in Cerro | 033041 | 29010 | — | — |
| San Polo d'Enza | 035038 | 42020 | — | — |
| San Possidonio | 036038 | 41039 | — | — |
| San Prospero | 036039 | 41030 | — | — |
| San Secondo Parmense | 034033 | 43017 | — | — |
| Sant'Agata Bolognese | 037056 | 40019 | — | — |
| Sant'Agata Feltria | 099026 | 47866 | — | — |
| Sant'Agata sul Santerno | 039017 | 48020 | — | — |
| Sant'Ilario d'Enza | 035039 | 42049 | — | — |
| Santa Sofia | 040043 | 47018 | — | — |
| Santarcangelo di Romagna | 099018 | 47822 | — | — |
| Sarmato | 033042 | 29010 | — | — |
| Sarsina | 040044 | 47027 | — | — |
| Sasso Marconi | 037057 | 40037 | — | — |
| Sassofeltrio | 099031 | 47869 | — | — |
| Sassuolo | 036040 | 41049 | — | — |
| Savignano sul Panaro | 036041 | 41056 | — | — |
| Savignano sul Rubicone | 040045 | 47039 | — | — |
| Scandiano | 035040 | 42019 | — | — |
| Serramazzoni | 036042 | 41028 | — | — |
| Sestola | 036043 | 41029 | — | — |
| Sissa Trecasali | 034049 | 43018 | — | — |
| Sogliano al Rubicone | 040046 | 47030 | — | — |
| Solarolo | 039018 | 48027 | — | — |
| Soliera | 036044 | 41019 | — | — |
| Solignano | 034035 | 43046 | — | — |
| Soragna | 034036 | 43019 | — | — |
| Sorbolo Mezzani | 034051 | 43058 | — | — |
| Spilamberto | 036045 | 41057 | — | — |
| Talamello | 099027 | 47867 | — | — |
| Terenzo | 034038 | 43040 | — | — |
| Terre del Reno | 038028 | 44047 | — | — |
| Tizzano Val Parma | 034039 | 43028 | — | — |
| Toano | 035041 | 42010 | — | — |
| Tornolo | 034040 | 43059 | — | — |
| Torrile | 034041 | 43056 | — | — |
| Traversetolo | 034042 | 43029 | — | — |
| Travo | 033043 | 29020 | — | — |
| Tredozio | 040049 | 47019 | — | — |
| Tresignana | 038030 | 44039 | — | — |
| Valmozzola | 034044 | 43050 | — | — |
| Valsamoggia | 037061 | 40053 | — | — |
| Varano de' Melegari | 034045 | 43040 | — | — |
| Varsi | 034046 | 43049 | — | — |
| Ventasso | 035046 | 42032 | — | — |
| Vergato | 037059 | 40038 | — | — |
| Verghereto | 040050 | 47028 | — | — |
| Vernasca | 033044 | 29010 | — | — |
| Verucchio | 099020 | 47826 | — | — |
| Vetto | 035042 | 42020 | — | — |
| Vezzano sul Crostolo | 035043 | 42030 | — | — |
| Viano | 035044 | 42030 | — | — |
| Vigarano Mainarda | 038022 | 44049 | — | — |
| Vignola | 036046 | 41058 | — | — |
| Vigolzone | 033045 | 29020 | — | — |
| Villa Minozzo | 035045 | 42030 | — | — |
| Villanova sull'Arda | 033046 | 29010 | — | — |
| Voghiera | 038023 | 44019 | — | — |
| Zerba | 033047 | 29020 | — | — |
| Ziano Piacentino | 033048 | 29010 | — | — |
| Zocca | 036047 | 41059 | — | — |
| Zola Predosa | 037060 | 40069 | — | — |

<a id="region-06"></a>

### Friuli-Venezia Giulia (215 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Friuli-Venezia Giulia — region** | 06 | — | Not started | Regional sources not checked |
| Aiello del Friuli | 030001 | 33041 | — | — |
| Amaro | 030002 | 33020 | — | — |
| Ampezzo | 030003 | 33021 | — | — |
| Andreis | 093001 | 33080 | — | — |
| Aquileia | 030004 | 33051 | — | — |
| Arba | 093002 | 33090 | — | — |
| Arta Terme | 030005 | 33022 | — | — |
| Artegna | 030006 | 33011 | — | — |
| Attimis | 030007 | 33040 | — | — |
| Aviano | 093004 | 33081 | — | — |
| Azzano Decimo | 093005 | 33082 | — | — |
| Bagnaria Arsa | 030008 | 33050 | — | — |
| Barcis | 093006 | 33080 | — | — |
| Basiliano | 030009 | 33031 | — | — |
| Bertiolo | 030010 | 33032 | — | — |
| Bicinicco | 030011 | 33050 | — | — |
| Bordano | 030012 | 33010 | — | — |
| Brugnera | 093007 | 33070 | — | — |
| Budoia | 093008 | 33070 | — | — |
| Buja | 030013 | 33030 | — | — |
| Buttrio | 030014 | 33042 | — | — |
| Camino al Tagliamento | 030015 | 33030 | — | — |
| Campoformido | 030016 | 33030 | — | — |
| Campolongo Tapogliano | 030138 | 33040 | — | — |
| Caneva | 093009 | 33070 | — | — |
| Capriva del Friuli | 031001 | 34070 | — | — |
| Carlino | 030018 | 33050 | — | — |
| Casarsa della Delizia | 093010 | 33072 | — | — |
| Cassacco | 030019 | 33010 | — | — |
| Castelnovo del Friuli | 093011 | 33090 | — | — |
| Castions di Strada | 030020 | 33050 | — | — |
| Cavasso Nuovo | 093012 | 33092 | — | — |
| Cavazzo Carnico | 030021 | 33020 | — | — |
| Cercivento | 030022 | 33020 | — | — |
| Cervignano del Friuli | 030023 | 33052 | — | — |
| Chions | 093013 | 33083 | — | — |
| Chiopris-Viscone | 030024 | 33048 | — | — |
| Chiusaforte | 030025 | 33010 | — | — |
| Cimolais | 093014 | 33080 | — | — |
| Cividale del Friuli | 030026 | 33043 | — | — |
| Claut | 093015 | 33080 | — | — |
| Clauzetto | 093016 | 33090 | — | — |
| Codroipo | 030027 | 33033 | — | — |
| Colloredo di Monte Albano | 030028 | 33010 | — | — |
| Comeglians | 030029 | 33023 | — | — |
| Cordenons | 093017 | 33084 | — | — |
| Cordovado | 093018 | 33075 | — | — |
| Cormons | 031002 | 34071 | — | — |
| Corno di Rosazzo | 030030 | 33040 | — | — |
| Coseano | 030031 | 33030 | — | — |
| Dignano | 030032 | 33030 | — | — |
| Doberdò del Lago | 031003 | 34070 | — | — |
| Dogna | 030033 | 33010 | — | — |
| Dolegna del Collio | 031004 | 34070 | — | — |
| Drenchia | 030034 | 33040 | — | — |
| Duino Aurisina | 032001 | 34011 | — | — |
| Enemonzo | 030035 | 33020 | — | — |
| Erto e Casso | 093019 | 33080 | — | — |
| Faedis | 030036 | 33040 | — | — |
| Fagagna | 030037 | 33034 | — | — |
| Fanna | 093020 | 33092 | — | — |
| Farra d'Isonzo | 031005 | 34072 | — | — |
| Fiume Veneto | 093021 | 33080 | — | — |
| Fiumicello Villa Vicentina | 030190 | 33059 | — | — |
| Flaibano | 030039 | 33030 | — | — |
| Fogliano Redipuglia | 031006 | 34070 | — | — |
| Fontanafredda | 093022 | 33074 | — | — |
| Forgaria nel Friuli | 030137 | 33030 | — | — |
| Forni Avoltri | 030040 | 33020 | — | — |
| Forni di Sopra | 030041 | 33024 | — | — |
| Forni di Sotto | 030042 | 33020 | — | — |
| Frisanco | 093024 | 33080 | — | — |
| Gemona del Friuli | 030043 | 33013 | — | — |
| Gonars | 030044 | 33050 | — | — |
| Gorizia | 031007 | 34170 | — | — |
| Gradisca d'Isonzo | 031008 | 34072 | — | — |
| Grado | 031009 | 34073 | — | — |
| Grimacco | 030045 | 33040 | — | — |
| Latisana | 030046 | 33053 | — | — |
| Lauco | 030047 | 33029 | — | — |
| Lestizza | 030048 | 33050 | — | — |
| Lignano Sabbiadoro | 030049 | 33054 | — | — |
| Lusevera | 030051 | 33010 | — | — |
| Magnano in Riviera | 030052 | 33010 | — | — |
| Majano | 030053 | 33030 | — | — |
| Malborghetto Valbruna | 030054 | 33010 | — | — |
| Maniago | 093025 | 33085 | — | — |
| Manzano | 030055 | 33044 | — | — |
| Marano Lagunare | 030056 | 33050 | — | — |
| Mariano del Friuli | 031010 | 34070 | — | — |
| Martignacco | 030057 | 33035 | — | — |
| Medea | 031011 | 34076 | — | — |
| Meduno | 093026 | 33092 | — | — |
| Mereto di Tomba | 030058 | 33036 | — | — |
| Moggio Udinese | 030059 | 33015 | — | — |
| Moimacco | 030060 | 33040 | — | — |
| Monfalcone | 031012 | 34074 | — | — |
| Monrupino | 032002 | 34016 | — | — |
| Montenars | 030061 | 33010 | — | — |
| Montereale Valcellina | 093027 | 33086 | — | — |
| Moraro | 031013 | 34070 | — | — |
| Morsano al Tagliamento | 093028 | 33075 | — | — |
| Mortegliano | 030062 | 33050 | — | — |
| Moruzzo | 030063 | 33030 | — | — |
| Mossa | 031014 | 34070 | — | — |
| Muggia | 032003 | 34015 | — | — |
| Muzzana del Turgnano | 030064 | 33055 | — | — |
| Nimis | 030065 | 33045 | — | — |
| Osoppo | 030066 | 33010 | — | — |
| Ovaro | 030067 | 33025 | — | — |
| Pagnacco | 030068 | 33010 | — | — |
| Palazzolo dello Stella | 030069 | 33056 | — | — |
| Palmanova | 030070 | 33057 | — | — |
| Paluzza | 030071 | 33026 | — | — |
| Pasian di Prato | 030072 | 33037 | — | — |
| Pasiano di Pordenone | 093029 | 33087 | — | — |
| Paularo | 030073 | 33027 | — | — |
| Pavia di Udine | 030074 | 33050 | — | — |
| Pinzano al Tagliamento | 093030 | 33094 | — | — |
| Pocenia | 030075 | 33050 | — | — |
| Polcenigo | 093031 | 33070 | — | — |
| Pontebba | 030076 | 33016 | — | — |
| Porcia | 093032 | 33080 | — | — |
| Pordenone | 093033 | 33170 | — | — |
| Porpetto | 030077 | 33050 | — | — |
| Povoletto | 030078 | 33040 | — | — |
| Pozzuolo del Friuli | 030079 | 33050 | — | — |
| Pradamano | 030080 | 33040 | — | — |
| Prata di Pordenone | 093034 | 33080 | — | — |
| Prato Carnico | 030081 | 33020 | — | — |
| Pravisdomini | 093035 | 33076 | — | — |
| Precenicco | 030082 | 33050 | — | — |
| Premariacco | 030083 | 33040 | — | — |
| Preone | 030084 | 33020 | — | — |
| Prepotto | 030085 | 33040 | — | — |
| Pulfero | 030086 | 33046 | — | — |
| Ragogna | 030087 | 33030 | — | — |
| Ravascletto | 030088 | 33020 | — | — |
| Raveo | 030089 | 33029 | — | — |
| Reana del Rojale | 030090 | 33010 | — | — |
| Remanzacco | 030091 | 33047 | — | — |
| Resia | 030092 | 33010 | — | — |
| Resiutta | 030093 | 33010 | — | — |
| Rigolato | 030094 | 33020 | — | — |
| Rive d'Arcano | 030095 | 33030 | — | — |
| Rivignano Teor | 030188 | 33061 | — | — |
| Romans d'Isonzo | 031015 | 34076 | — | — |
| Ronchi dei Legionari | 031016 | 34077 | — | — |
| Ronchis | 030097 | 33050 | — | — |
| Roveredo in Piano | 093036 | 33080 | — | — |
| Ruda | 030098 | 33050 | — | — |
| Sacile | 093037 | 33077 | — | — |
| Sagrado | 031017 | 34078 | — | — |
| San Canzian d'Isonzo | 031018 | 34075 | — | — |
| San Daniele del Friuli | 030099 | 33038 | — | — |
| San Dorligo della Valle | 032004 | 34018 | — | — |
| San Floriano del Collio | 031019 | 34070 | — | — |
| San Giorgio della Richinvelda | 093038 | 33095 | — | — |
| San Giorgio di Nogaro | 030100 | 33058 | — | — |
| San Giovanni al Natisone | 030101 | 33048 | — | — |
| San Leonardo | 030102 | 33040 | — | — |
| San Lorenzo Isontino | 031020 | 34070 | — | — |
| San Martino al Tagliamento | 093039 | 33098 | — | — |
| San Pier d'Isonzo | 031021 | 34070 | — | — |
| San Pietro al Natisone | 030103 | 33049 | — | — |
| San Quirino | 093040 | 33080 | — | — |
| San Vito al Tagliamento | 093041 | 33078 | — | — |
| San Vito al Torre | 030105 | 33050 | — | — |
| San Vito di Fagagna | 030106 | 33030 | — | — |
| Santa Maria la Longa | 030104 | 33050 | — | — |
| Sappada | 030189 | 33012 | — | — |
| Sauris | 030107 | 33020 | — | — |
| Savogna | 030108 | 33040 | — | — |
| Savogna d'Isonzo | 031022 | 34070 | — | — |
| Sedegliano | 030109 | 33039 | — | — |
| Sequals | 093042 | 33090 | — | — |
| Sesto al Reghena | 093043 | 33079 | — | — |
| Sgonico | 032005 | 34010 | — | — |
| Socchieve | 030110 | 33020 | — | — |
| Spilimbergo | 093044 | 33097 | — | — |
| Staranzano | 031023 | 34079 | — | — |
| Stregna | 030111 | 33040 | — | — |
| Sutrio | 030112 | 33020 | — | — |
| Taipana | 030113 | 33040 | — | — |
| Talmassons | 030114 | 33030 | — | — |
| Tarcento | 030116 | 33017 | — | — |
| Tarvisio | 030117 | 33018 | — | — |
| Tavagnacco | 030118 | 33010 | — | — |
| Terzo d'Aquileia | 030120 | 33050 | — | — |
| Tolmezzo | 030121 | 33028 | — | — |
| Torreano | 030122 | 33040 | — | — |
| Torviscosa | 030123 | 33050 | — | — |
| Tramonti di Sopra | 093045 | 33090 | — | — |
| Tramonti di Sotto | 093046 | 33090 | — | — |
| Trasaghis | 030124 | 33010 | — | — |
| Travesio | 093047 | 33090 | — | — |
| Treppo Grande | 030126 | 33010 | — | — |
| Treppo Ligosullo | 030191 | 33014 | — | — |
| Tricesimo | 030127 | 33019 | — | — |
| Trieste | 032006 | 34121, 34122, 34123, 34124, 34125, 34126, 34127, 34128, 34129, 34131, 34132, 34133, 34134, 34135, 34136, 34137, 34138, 34139, 34141, 34142, 34143, 34144, 34145, 34146, 34147, 34148, 34149, 34151 | — | — |
| Trivignano Udinese | 030128 | 33050 | — | — |
| Turriaco | 031024 | 34070 | — | — |
| Udine | 030129 | 33100 | — | — |
| Vajont | 093052 | 33080 | — | — |
| Valvasone Arzene | 093053 | 33098 | — | — |
| Varmo | 030130 | 33030 | — | — |
| Venzone | 030131 | 33010 | — | — |
| Verzegnis | 030132 | 33020 | — | — |
| Villa Santina | 030133 | 33029 | — | — |
| Villesse | 031025 | 34070 | — | — |
| Visco | 030135 | 33040 | — | — |
| Vito d'Asio | 093049 | 33090 | — | — |
| Vivaro | 093050 | 33099 | — | — |
| Zoppola | 093051 | 33080 | — | — |
| Zuglio | 030136 | 33020 | — | — |

<a id="region-12"></a>

### Lazio (378 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Lazio — region** | 12 | — | Not started | Regional sources not checked |
| Accumoli | 057001 | 02011 | — | — |
| Acquafondata | 060001 | 03040 | — | — |
| Acquapendente | 056001 | 01021 | — | — |
| Acuto | 060002 | 03010 | — | — |
| Affile | 058001 | 00021 | — | — |
| Agosta | 058002 | 00020 | — | — |
| Alatri | 060003 | 03011 | — | — |
| Albano Laziale | 058003 | 00041 | — | — |
| Allumiere | 058004 | 00051 | — | — |
| Alvito | 060004 | 03041 | — | — |
| Amaseno | 060005 | 03021 | — | — |
| Amatrice | 057002 | 02012 | — | — |
| Anagni | 060006 | 03012 | — | — |
| Anguillara Sabazia | 058005 | 00061 | — | — |
| Anticoli Corrado | 058006 | 00022 | — | — |
| Antrodoco | 057003 | 02013 | — | — |
| Anzio | 058007 | 00042 | — | — |
| Aprilia | 059001 | 04011 | — | — |
| Aquino | 060007 | 03031 | — | — |
| Arce | 060008 | 03032 | — | — |
| Arcinazzo Romano | 058008 | 00020 | — | — |
| Ardea | 058117 | 00040 | — | — |
| Ariccia | 058009 | 00072 | — | — |
| Arlena di Castro | 056002 | 01010 | — | — |
| Arnara | 060009 | 03020 | — | — |
| Arpino | 060010 | 03033 | — | — |
| Arsoli | 058010 | 00023 | — | — |
| Artena | 058011 | 00031 | — | — |
| Ascrea | 057004 | 02020 | — | — |
| Atina | 060011 | 03042 | — | — |
| Ausonia | 060012 | 03040 | — | — |
| Bagnoregio | 056003 | 01022 | — | — |
| Barbarano Romano | 056004 | 01010 | — | — |
| Bassano in Teverina | 056006 | 01030 | — | — |
| Bassano Romano | 056005 | 01030 | — | — |
| Bassiano | 059002 | 04010 | — | — |
| Bellegra | 058012 | 00030 | — | — |
| Belmonte Castello | 060013 | 03040 | — | — |
| Belmonte in Sabina | 057005 | 02020 | — | — |
| Blera | 056007 | 01010 | — | — |
| Bolsena | 056008 | 01023 | — | — |
| Bomarzo | 056009 | 01020 | — | — |
| Borbona | 057006 | 02010 | — | — |
| Borgo Velino | 057008 | 02010 | — | — |
| Borgorose | 057007 | 02021 | — | — |
| Boville Ernica | 060014 | 03022 | — | — |
| Bracciano | 058013 | 00062 | — | — |
| Broccostella | 060015 | 03030 | — | — |
| Calcata | 056010 | 01030 | — | — |
| Camerata Nuova | 058014 | 00020 | — | — |
| Campagnano di Roma | 058015 | 00063 | — | — |
| Campodimele | 059003 | 04020 | — | — |
| Campoli Appennino | 060016 | 03030 | — | — |
| Canale Monterano | 058016 | 00060 | — | — |
| Canepina | 056011 | 01030 | — | — |
| Canino | 056012 | 01011 | — | — |
| Cantalice | 057009 | 02014 | — | — |
| Cantalupo in Sabina | 057010 | 02040 | — | — |
| Canterano | 058017 | 00020 | — | — |
| Capena | 058018 | 00060 | — | — |
| Capodimonte | 056013 | 01010 | — | — |
| Capranica | 056014 | 01012 | — | — |
| Capranica Prenestina | 058019 | 00030 | — | — |
| Caprarola | 056015 | 01032 | — | — |
| Carbognano | 056016 | 01030 | — | — |
| Carpineto Romano | 058020 | 00032 | — | — |
| Casalattico | 060017 | 03030 | — | — |
| Casalvieri | 060018 | 03034 | — | — |
| Casape | 058021 | 00010 | — | — |
| Casaprota | 057011 | 02030 | — | — |
| Casperia | 057012 | 02041 | — | — |
| Cassino | 060019 | 03043 | — | — |
| Castel di Tora | 057013 | 02020 | — | — |
| Castel Gandolfo | 058022 | 00073 | — | — |
| Castel Madama | 058023 | 00024 | — | — |
| Castel San Pietro Romano | 058025 | 00030 | — | — |
| Castel Sant'Angelo | 057015 | 02010 | — | — |
| Castel Sant'Elia | 056017 | 01030 | — | — |
| Castelforte | 059004 | 04021 | — | — |
| Castelliri | 060020 | 03030 | — | — |
| Castelnuovo di Farfa | 057014 | 02031 | — | — |
| Castelnuovo di Porto | 058024 | 00060 | — | — |
| Castelnuovo Parano | 060021 | 03040 | — | — |
| Castiglione in Teverina | 056018 | 01024 | — | — |
| Castro dei Volsci | 060023 | 03020 | — | — |
| Castrocielo | 060022 | 03030 | — | — |
| Cave | 058026 | 00033 | — | — |
| Ceccano | 060024 | 03023 | — | — |
| Celleno | 056019 | 01020 | — | — |
| Cellere | 056020 | 01010 | — | — |
| Ceprano | 060025 | 03024 | — | — |
| Cerreto Laziale | 058027 | 00020 | — | — |
| Cervara di Roma | 058028 | 00020 | — | — |
| Cervaro | 060026 | 03044 | — | — |
| Cerveteri | 058029 | 00052 | — | — |
| Ciampino | 058118 | 00043 | — | — |
| Ciciliano | 058030 | 00020 | — | — |
| Cineto Romano | 058031 | 00020 | — | — |
| Cisterna di Latina | 059005 | 04012 | — | — |
| Cittaducale | 057016 | 02015 | — | — |
| Cittareale | 057017 | 02010 | — | — |
| Civita Castellana | 056021 | 01033 | — | — |
| Civitavecchia | 058032 | 00053 | — | — |
| Civitella d'Agliano | 056022 | 01020 | — | — |
| Civitella San Paolo | 058033 | 00060 | — | — |
| Colfelice | 060027 | 03030 | — | — |
| Collalto Sabino | 057018 | 02022 | — | — |
| Colle di Tora | 057019 | 02020 | — | — |
| Colle San Magno | 060029 | 03030 | — | — |
| Colleferro | 058034 | 00034 | — | — |
| Collegiove | 057020 | 02020 | — | — |
| Collepardo | 060028 | 03010 | — | — |
| Collevecchio | 057021 | 02042 | — | — |
| Colli sul Velino | 057022 | 02010 | — | — |
| Colonna | 058035 | 00030 | — | — |
| Concerviano | 057023 | 02020 | — | — |
| Configni | 057024 | 02040 | — | — |
| Contigliano | 057025 | 02043 | — | — |
| Corchiano | 056023 | 01030 | — | — |
| Coreno Ausonio | 060030 | 03040 | — | — |
| Cori | 059006 | 04010 | — | — |
| Cottanello | 057026 | 02040 | — | — |
| Esperia | 060031 | 03045 | — | — |
| Fabrica di Roma | 056024 | 01034 | — | — |
| Faleria | 056025 | 01030 | — | — |
| Falvaterra | 060032 | 03020 | — | — |
| Fara in Sabina | 057027 | 02032 | — | — |
| Farnese | 056026 | 01010 | — | — |
| Ferentino | 060033 | 03013 | — | — |
| Fiamignano | 057028 | 02023 | — | — |
| Fiano Romano | 058036 | 00065 | — | — |
| Filacciano | 058037 | 00060 | — | — |
| Filettino | 060034 | 03010 | — | — |
| Fiuggi | 060035 | 03014 | — | — |
| Fiumicino | 058120 | 00054 | — | — |
| Fondi | 059007 | 04022 | — | — |
| Fontana Liri | 060036 | 03035 | — | — |
| Fonte Nuova | 058122 | 00013 | — | — |
| Fontechiari | 060037 | 03030 | — | — |
| Forano | 057029 | 02044 | — | — |
| Formello | 058038 | 00060 | — | — |
| Formia | 059008 | 04023 | — | — |
| Frascati | 058039 | 00044 | — | — |
| Frasso Sabino | 057030 | 02030 | — | — |
| Frosinone | 060038 | 03100 | — | — |
| Fumone | 060039 | 03010 | — | — |
| Gaeta | 059009 | 04024 | — | — |
| Gallese | 056027 | 01035 | — | — |
| Gallicano nel Lazio | 058040 | 00010 | — | — |
| Gallinaro | 060040 | 03040 | — | — |
| Gavignano | 058041 | 00030 | — | — |
| Genazzano | 058042 | 00030 | — | — |
| Genzano di Roma | 058043 | 00045 | — | — |
| Gerano | 058044 | 00025 | — | — |
| Giuliano di Roma | 060041 | 03020 | — | — |
| Gorga | 058045 | 00030 | — | — |
| Gradoli | 056028 | 01010 | — | — |
| Graffignano | 056029 | 01020 | — | — |
| Greccio | 057031 | 02045 | — | — |
| Grottaferrata | 058046 | 00046 | — | — |
| Grotte di Castro | 056030 | 01025 | — | — |
| Guarcino | 060042 | 03016 | — | — |
| Guidonia Montecelio | 058047 | 00012 | — | — |
| Ischia di Castro | 056031 | 01010 | — | — |
| Isola del Liri | 060043 | 03036 | — | — |
| Itri | 059010 | 04020 | — | — |
| Jenne | 058048 | 00020 | — | — |
| Labico | 058049 | 00030 | — | — |
| Labro | 057032 | 02010 | — | — |
| Ladispoli | 058116 | 00055 | — | — |
| Lanuvio | 058050 | 00075 | — | — |
| Lariano | 058115 | 00076 | — | — |
| Latera | 056032 | 01010 | — | — |
| Latina | 059011 | 04100 | — | — |
| Lenola | 059012 | 04025 | — | — |
| Leonessa | 057033 | 02016 | — | — |
| Licenza | 058051 | 00026 | — | — |
| Longone Sabino | 057034 | 02020 | — | — |
| Lubriano | 056033 | 01020 | — | — |
| Maenza | 059013 | 04010 | — | — |
| Magliano Romano | 058052 | 00060 | — | — |
| Magliano Sabina | 057035 | 02046 | — | — |
| Mandela | 058053 | 00020 | — | — |
| Manziana | 058054 | 00066 | — | — |
| Marano Equo | 058055 | 00020 | — | — |
| Marcellina | 058056 | 00010 | — | — |
| Marcetelli | 057036 | 02020 | — | — |
| Marino | 058057 | 00047 | — | — |
| Marta | 056034 | 01010 | — | — |
| Mazzano Romano | 058058 | 00060 | — | — |
| Mentana | 058059 | 00013 | — | — |
| Micigliano | 057037 | 02010 | — | — |
| Minturno | 059014 | 04026 | — | — |
| Mompeo | 057038 | 02040 | — | — |
| Montalto di Castro | 056035 | 01014 | — | — |
| Montasola | 057039 | 02040 | — | — |
| Monte Compatri | 058060 | 00077 | — | — |
| Monte Porzio Catone | 058064 | 00078 | — | — |
| Monte Romano | 056037 | 01010 | — | — |
| Monte San Biagio | 059015 | 04020 | — | — |
| Monte San Giovanni Campano | 060044 | 03025 | — | — |
| Monte San Giovanni in Sabina | 057043 | 02040 | — | — |
| Montebuono | 057040 | 02040 | — | — |
| Montefiascone | 056036 | 01027 | — | — |
| Monteflavio | 058061 | 00010 | — | — |
| Montelanico | 058062 | 00030 | — | — |
| Monteleone Sabino | 057041 | 02033 | — | — |
| Montelibretti | 058063 | 00010 | — | — |
| Montenero Sabino | 057042 | 02040 | — | — |
| Monterosi | 056038 | 01030 | — | — |
| Monterotondo | 058065 | 00015 | — | — |
| Montopoli di Sabina | 057044 | 02034 | — | — |
| Montorio Romano | 058066 | 00010 | — | — |
| Moricone | 058067 | 00010 | — | — |
| Morlupo | 058068 | 00067 | — | — |
| Morolo | 060045 | 03017 | — | — |
| Morro Reatino | 057045 | 02010 | — | — |
| Nazzano | 058069 | 00060 | — | — |
| Nemi | 058070 | 00074 | — | — |
| Nepi | 056039 | 01036 | — | — |
| Nerola | 058071 | 00017 | — | — |
| Nespolo | 057046 | 02020 | — | — |
| Nettuno | 058072 | 00048 | — | — |
| Norma | 059016 | 04010 | — | — |
| Olevano Romano | 058073 | 00035 | — | — |
| Onano | 056040 | 01010 | — | — |
| Oriolo Romano | 056041 | 01010 | — | — |
| Orte | 056042 | 01028 | — | — |
| Orvinio | 057047 | 02035 | — | — |
| Paganico Sabino | 057048 | 02020 | — | — |
| Palestrina | 058074 | 00036 | — | — |
| Paliano | 060046 | 03018 | — | — |
| Palombara Sabina | 058075 | 00018 | — | — |
| Pastena | 060047 | 03020 | — | — |
| Patrica | 060048 | 03010 | — | — |
| Percile | 058076 | 00020 | — | — |
| Pescorocchiano | 057049 | 02024 | — | — |
| Pescosolido | 060049 | 03030 | — | — |
| Petrella Salto | 057050 | 02025 | — | — |
| Piansano | 056043 | 01010 | — | — |
| Picinisco | 060050 | 03040 | — | — |
| Pico | 060051 | 03020 | — | — |
| Piedimonte San Germano | 060052 | 03030 | — | — |
| Piglio | 060053 | 03010 | — | — |
| Pignataro Interamna | 060054 | 03040 | — | — |
| Pisoniano | 058077 | 00020 | — | — |
| Pofi | 060055 | 03026 | — | — |
| Poggio Bustone | 057051 | 02018 | — | — |
| Poggio Catino | 057052 | 02040 | — | — |
| Poggio Mirteto | 057053 | 02047 | — | — |
| Poggio Moiano | 057054 | 02037 | — | — |
| Poggio Nativo | 057055 | 02030 | — | — |
| Poggio San Lorenzo | 057056 | 02030 | — | — |
| Poli | 058078 | 00010 | — | — |
| Pomezia | 058079 | 00071 | — | — |
| Pontecorvo | 060056 | 03037 | — | — |
| Pontinia | 059017 | 04014 | — | — |
| Ponza | 059018 | 04027 | — | — |
| Ponzano Romano | 058080 | 00060 | — | — |
| Posta | 057057 | 02019 | — | — |
| Posta Fibreno | 060057 | 03030 | — | — |
| Pozzaglia Sabina | 057058 | 02030 | — | — |
| Priverno | 059019 | 04015 | — | — |
| Proceno | 056044 | 01020 | — | — |
| Prossedi | 059020 | 04010 | — | — |
| Riano | 058081 | 00060 | — | — |
| Rieti | 057059 | 02100 | — | — |
| Rignano Flaminio | 058082 | 00068 | — | — |
| Riofreddo | 058083 | 00020 | — | — |
| Ripi | 060058 | 03027 | — | — |
| Rivodutri | 057060 | 02010 | — | — |
| Rocca Canterano | 058084 | 00020 | — | — |
| Rocca d'Arce | 060059 | 03030 | — | — |
| Rocca di Cave | 058085 | 00030 | — | — |
| Rocca di Papa | 058086 | 00040 | — | — |
| Rocca Massima | 059022 | 04010 | — | — |
| Rocca Priora | 058088 | 00079 | — | — |
| Rocca Santo Stefano | 058089 | 00030 | — | — |
| Rocca Sinibalda | 057062 | 02026 | — | — |
| Roccagiovine | 058087 | 00020 | — | — |
| Roccagorga | 059021 | 04010 | — | — |
| Roccantica | 057061 | 02040 | — | — |
| Roccasecca | 060060 | 03038 | — | — |
| Roccasecca dei Volsci | 059023 | 04010 | — | — |
| Roiate | 058090 | 00030 | — | — |
| Roma | 058091 | 00118, 00119, 00120, 00121, 00122, 00123, 00124, 00125, 00126, 00127, 00128, 00131, 00132, 00133, 00134, 00135, 00136, 00137, 00138, 00139, 00141, 00142, 00143, 00144, 00145, 00146, 00147, 00148, 00149, 00151, 00152, 00153, 00154, 00155, 00156, 00157, 00158, 00159, 00161, 00162, 00163, 00164, 00165, 00166, 00167, 00168, 00169, 00171, 00172, 00173, 00174, 00175, 00176, 00177, 00178, 00179, 00181, 00182, 00183, 00184, 00185, 00186, 00187, 00188, 00189, 00191, 00192, 00193, 00195, 00196, 00197, 00198, 00199 | — | — |
| Ronciglione | 056045 | 01037 | — | — |
| Roviano | 058092 | 00027 | — | — |
| Sabaudia | 059024 | 04016 | — | — |
| Sacrofano | 058093 | 00060 | — | — |
| Salisano | 057063 | 02040 | — | — |
| Sambuci | 058094 | 00020 | — | — |
| San Biagio Saracinisco | 060061 | 03040 | — | — |
| San Cesareo | 058119 | 00030 | — | — |
| San Donato Val di Comino | 060062 | 03046 | — | — |
| San Felice Circeo | 059025 | 04017 | — | — |
| San Giorgio a Liri | 060063 | 03047 | — | — |
| San Giovanni Incarico | 060064 | 03028 | — | — |
| San Gregorio da Sassola | 058095 | 00010 | — | — |
| San Lorenzo Nuovo | 056047 | 01020 | — | — |
| San Polo dei Cavalieri | 058096 | 00010 | — | — |
| San Vito Romano | 058100 | 00030 | — | — |
| San Vittore del Lazio | 060070 | 03040 | — | — |
| Sant'Ambrogio sul Garigliano | 060065 | 03040 | — | — |
| Sant'Andrea del Garigliano | 060066 | 03040 | — | — |
| Sant'Angelo Romano | 058098 | 00010 | — | — |
| Sant'Apollinare | 060067 | 03048 | — | — |
| Sant'Elia Fiumerapido | 060068 | 03049 | — | — |
| Sant'Oreste | 058099 | 00060 | — | — |
| Santa Marinella | 058097 | 00058 | — | — |
| Santi Cosma e Damiano | 059026 | 04020 | — | — |
| Santopadre | 060069 | 03030 | — | — |
| Saracinesco | 058101 | 00020 | — | — |
| Scandriglia | 057064 | 02038 | — | — |
| Segni | 058102 | 00037 | — | — |
| Selci | 057065 | 02040 | — | — |
| Sermoneta | 059027 | 04013 | — | — |
| Serrone | 060071 | 03010 | — | — |
| Settefrati | 060072 | 03040 | — | — |
| Sezze | 059028 | 04018 | — | — |
| Sgurgola | 060073 | 03010 | — | — |
| Sonnino | 059029 | 04010 | — | — |
| Sora | 060074 | 03039 | — | — |
| Soriano nel Cimino | 056048 | 01038 | — | — |
| Sperlonga | 059030 | 04029 | — | — |
| Spigno Saturnia | 059031 | 04020 | — | — |
| Stimigliano | 057066 | 02048 | — | — |
| Strangolagalli | 060075 | 03020 | — | — |
| Subiaco | 058103 | 00028 | — | — |
| Supino | 060076 | 03019 | — | — |
| Sutri | 056049 | 01015 | — | — |
| Tarano | 057067 | 02040 | — | — |
| Tarquinia | 056050 | 01016 | — | — |
| Terelle | 060077 | 03040 | — | — |
| Terracina | 059032 | 04019 | — | — |
| Tessennano | 056051 | 01010 | — | — |
| Tivoli | 058104 | 00019 | — | — |
| Toffia | 057068 | 02039 | — | — |
| Tolfa | 058105 | 00059 | — | — |
| Torre Cajetani | 060078 | 03010 | — | — |
| Torri in Sabina | 057070 | 02049 | — | — |
| Torrice | 060079 | 03020 | — | — |
| Torricella in Sabina | 057069 | 02030 | — | — |
| Torrita Tiberina | 058106 | 00060 | — | — |
| Trevi nel Lazio | 060080 | 03010 | — | — |
| Trevignano Romano | 058107 | 00069 | — | — |
| Trivigliano | 060081 | 03010 | — | — |
| Turania | 057071 | 02020 | — | — |
| Tuscania | 056052 | 01017 | — | — |
| Vacone | 057072 | 02040 | — | — |
| Valentano | 056053 | 01018 | — | — |
| Vallecorsa | 060082 | 03020 | — | — |
| Vallemaio | 060083 | 03040 | — | — |
| Vallepietra | 058108 | 00020 | — | — |
| Vallerano | 056054 | 01030 | — | — |
| Vallerotonda | 060084 | 03040 | — | — |
| Vallinfreda | 058109 | 00020 | — | — |
| Valmontone | 058110 | 00038 | — | — |
| Varco Sabino | 057073 | 02020 | — | — |
| Vasanello | 056055 | 01030 | — | — |
| Vejano | 056056 | 01010 | — | — |
| Velletri | 058111 | 00049 | — | — |
| Ventotene | 059033 | 04031 | — | — |
| Veroli | 060085 | 03029 | — | — |
| Vetralla | 056057 | 01019 | — | — |
| Vicalvi | 060086 | 03030 | — | — |
| Vico nel Lazio | 060087 | 03010 | — | — |
| Vicovaro | 058112 | 00029 | — | — |
| Vignanello | 056058 | 01039 | — | — |
| Villa Latina | 060088 | 03040 | — | — |
| Villa San Giovanni in Tuscia | 056046 | 01010 | — | — |
| Villa Santa Lucia | 060089 | 03030 | — | — |
| Villa Santo Stefano | 060090 | 03020 | — | — |
| Viterbo | 056059 | 01100 | — | — |
| Viticuso | 060091 | 03040 | — | — |
| Vitorchiano | 056060 | 01030 | — | — |
| Vivaro Romano | 058113 | 00020 | — | — |
| Zagarolo | 058114 | 00039 | — | — |

<a id="region-07"></a>

### Liguria (234 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Liguria — region** | 07 | — | Not started | Regional sources not checked |
| Airole | 008001 | 18030 | — | — |
| Alassio | 009001 | 17021 | — | — |
| Albenga | 009002 | 17031 | — | — |
| Albisola Superiore | 009004 | 17011 | — | — |
| Albissola Marina | 009003 | 17012 | — | — |
| Altare | 009005 | 17041 | — | — |
| Ameglia | 011001 | 19031 | — | — |
| Andora | 009006 | 17051 | — | — |
| Apricale | 008002 | 18035 | — | — |
| Aquila d'Arroscia | 008003 | 18020 | — | — |
| Arcola | 011002 | 19021 | — | — |
| Arenzano | 010001 | 16011 | — | — |
| Armo | 008004 | 18026 | — | — |
| Arnasco | 009007 | 17032 | — | — |
| Aurigo | 008005 | 18020 | — | — |
| Avegno | 010002 | 16036 | — | — |
| Badalucco | 008006 | 18010 | — | — |
| Bajardo | 008007 | 18031 | — | — |
| Balestrino | 009008 | 17020 | — | — |
| Bardineto | 009009 | 17057 | — | — |
| Bargagli | 010003 | 16021 | — | — |
| Bergeggi | 009010 | 17028 | — | — |
| Beverino | 011003 | 19020 | — | — |
| Bogliasco | 010004 | 16031 | — | — |
| Boissano | 009011 | 17054 | — | — |
| Bolano | 011004 | 19020 | — | — |
| Bonassola | 011005 | 19011 | — | — |
| Bordighera | 008008 | 18012 | — | — |
| Borghetto d'Arroscia | 008009 | 18020 | — | — |
| Borghetto di Vara | 011006 | 19020 | — | — |
| Borghetto Santo Spirito | 009012 | 17052 | — | — |
| Borgio Verezzi | 009013 | 17022 | — | — |
| Borgomaro | 008010 | 18021 | — | — |
| Bormida | 009014 | 17045 | — | — |
| Borzonasca | 010005 | 16041 | — | — |
| Brugnato | 011007 | 19020 | — | — |
| Busalla | 010006 | 16012 | — | — |
| Cairo Montenotte | 009015 | 17014 | — | — |
| Calice al Cornoviglio | 011008 | 19020 | — | — |
| Calice Ligure | 009016 | 17020 | — | — |
| Calizzano | 009017 | 17057 | — | — |
| Camogli | 010007 | 16032 | — | — |
| Campo Ligure | 010008 | 16013 | — | — |
| Campomorone | 010009 | 16014 | — | — |
| Camporosso | 008011 | 18033 | — | — |
| Carasco | 010010 | 16042 | — | — |
| Caravonica | 008012 | 18020 | — | — |
| Carcare | 009018 | 17043 | — | — |
| Carro | 011009 | 19012 | — | — |
| Carrodano | 011010 | 19020 | — | — |
| Casanova Lerrone | 009019 | 17033 | — | — |
| Casarza Ligure | 010011 | 16030 | — | — |
| Casella | 010012 | 16015 | — | — |
| Castel Vittorio | 008015 | 18030 | — | — |
| Castelbianco | 009020 | 17030 | — | — |
| Castellaro | 008014 | 18011 | — | — |
| Castelnuovo Magra | 011011 | 19033 | — | — |
| Castelvecchio di Rocca Barbena | 009021 | 17034 | — | — |
| Castiglione Chiavarese | 010013 | 16030 | — | — |
| Celle Ligure | 009022 | 17015 | — | — |
| Cengio | 009023 | 17056 | — | — |
| Ceranesi | 010014 | 16014 | — | — |
| Ceriale | 009024 | 17023 | — | — |
| Ceriana | 008016 | 18034 | — | — |
| Cervo | 008017 | 18010 | — | — |
| Cesio | 008018 | 18022 | — | — |
| Chiavari | 010015 | 16043 | — | — |
| Chiusanico | 008019 | 18027 | — | — |
| Chiusavecchia | 008020 | 18027 | — | — |
| Cicagna | 010016 | 16044 | — | — |
| Cipressa | 008021 | 18017 | — | — |
| Cisano sul Neva | 009025 | 17035 | — | — |
| Civezza | 008022 | 18017 | — | — |
| Cogoleto | 010017 | 16016 | — | — |
| Cogorno | 010018 | 16030 | — | — |
| Coreglia Ligure | 010019 | 16040 | — | — |
| Cosio d'Arroscia | 008023 | 18023 | — | — |
| Cosseria | 009026 | 17017 | — | — |
| Costarainera | 008024 | 18017 | — | — |
| Crocefieschi | 010020 | 16010 | — | — |
| Davagna | 010021 | 16022 | — | — |
| Dego | 009027 | 17058 | — | — |
| Deiva Marina | 011012 | 19013 | — | — |
| Diano Arentino | 008025 | 18013 | — | — |
| Diano Castello | 008026 | 18013 | — | — |
| Diano Marina | 008027 | 18013 | — | — |
| Diano San Pietro | 008028 | 18013 | — | — |
| Dolceacqua | 008029 | 18035 | — | — |
| Dolcedo | 008030 | 18020 | — | — |
| Erli | 009028 | 17030 | — | — |
| Fascia | 010022 | 16020 | — | — |
| Favale di Malvaro | 010023 | 16040 | — | — |
| Finale Ligure | 009029 | 17024 | — | — |
| Follo | 011013 | 19020 | — | — |
| Fontanigorda | 010024 | 16023 | — | — |
| Framura | 011014 | 19014 | — | — |
| Garlenda | 009030 | 17033 | — | — |
| Genova | 010025 | 16121, 16122, 16123, 16124, 16125, 16126, 16127, 16128, 16129, 16131, 16132, 16133, 16134, 16135, 16136, 16137, 16138, 16139, 16141, 16142, 16143, 16144, 16145, 16146, 16147, 16148, 16149, 16151, 16152, 16153, 16154, 16155, 16156, 16157, 16158, 16159, 16161, 16162, 16163, 16164, 16165, 16166, 16167 | — | — |
| Giustenice | 009031 | 17027 | — | — |
| Giusvalla | 009032 | 17010 | — | — |
| Gorreto | 010026 | 16020 | — | — |
| Imperia | 008031 | 18100 | — | — |
| Isola del Cantone | 010027 | 16017 | — | — |
| Isolabona | 008032 | 18035 | — | — |
| La Spezia | 011015 | 19121, 19122, 19123, 19124, 19125, 19126, 19131, 19132, 19133, 19134, 19135, 19136, 19137 | — | — |
| Laigueglia | 009033 | 17053 | — | — |
| Lavagna | 010028 | 16033 | — | — |
| Leivi | 010029 | 16040 | — | — |
| Lerici | 011016 | 19032 | — | — |
| Levanto | 011017 | 19015 | — | — |
| Loano | 009034 | 17025 | — | — |
| Lorsica | 010030 | 16045 | — | — |
| Lucinasco | 008033 | 18020 | — | — |
| Lumarzo | 010031 | 16024 | — | — |
| Luni | 011020 | 19034 | — | — |
| Magliolo | 009035 | 17020 | — | — |
| Maissana | 011018 | 19010 | — | — |
| Mallare | 009036 | 17045 | — | — |
| Masone | 010032 | 16010 | — | — |
| Massimino | 009037 | 12071 | — | — |
| Mele | 010033 | 16010 | — | — |
| Mendatica | 008034 | 18025 | — | — |
| Mezzanego | 010034 | 16046 | — | — |
| Mignanego | 010035 | 16018 | — | — |
| Millesimo | 009038 | 17017 | — | — |
| Mioglia | 009039 | 17040 | — | — |
| Moconesi | 010036 | 16047 | — | — |
| Molini di Triora | 008035 | 18010 | — | — |
| Moneglia | 010037 | 16030 | — | — |
| Montalto Carpasio | 008068 | 18028 | — | — |
| Montebruno | 010038 | 16025 | — | — |
| Montegrosso Pian Latte | 008037 | 18025 | — | — |
| Monterosso al Mare | 011019 | 19016 | — | — |
| Montoggio | 010039 | 16026 | — | — |
| Murialdo | 009040 | 17013 | — | — |
| Nasino | 009041 | 17030 | — | — |
| Ne | 010040 | 16040 | — | — |
| Neirone | 010041 | 16040 | — | — |
| Noli | 009042 | 17026 | — | — |
| Olivetta San Michele | 008038 | 18030 | — | — |
| Onzo | 009043 | 17037 | — | — |
| Orco Feglino | 009044 | 17024 | — | — |
| Orero | 010042 | 16040 | — | — |
| Ortovero | 009045 | 17037 | — | — |
| Osiglia | 009046 | 17010 | — | — |
| Ospedaletti | 008039 | 18014 | — | — |
| Pallare | 009047 | 17043 | — | — |
| Perinaldo | 008040 | 18032 | — | — |
| Piana Crixia | 009048 | 17058 | — | — |
| Pietra Ligure | 009049 | 17027 | — | — |
| Pietrabruna | 008041 | 18010 | — | — |
| Pieve di Teco | 008042 | 18026 | — | — |
| Pieve Ligure | 010043 | 16031 | — | — |
| Pigna | 008043 | 18037 | — | — |
| Pignone | 011021 | 19020 | — | — |
| Plodio | 009050 | 17043 | — | — |
| Pompeiana | 008044 | 18015 | — | — |
| Pontedassio | 008045 | 18027 | — | — |
| Pontinvrea | 009051 | 17042 | — | — |
| Pornassio | 008046 | 18024 | — | — |
| Portofino | 010044 | 16034 | — | — |
| Portovenere | 011022 | 19025 | — | — |
| Prelà | 008047 | 18020 | — | — |
| Propata | 010045 | 16027 | — | — |
| Quiliano | 009052 | 17047 | — | — |
| Ranzo | 008048 | 18020 | — | — |
| Rapallo | 010046 | 16035 | — | — |
| Recco | 010047 | 16036 | — | — |
| Rezzo | 008049 | 18026 | — | — |
| Rezzoaglio | 010048 | 16048 | — | — |
| Rialto | 009053 | 17020 | — | — |
| Riccò del Golfo di Spezia | 011023 | 19020 | — | — |
| Riomaggiore | 011024 | 19017 | — | — |
| Riva Ligure | 008050 | 18015 | — | — |
| Roccavignale | 009054 | 17017 | — | — |
| Rocchetta di Vara | 011025 | 19020 | — | — |
| Rocchetta Nervina | 008051 | 18030 | — | — |
| Ronco Scrivia | 010049 | 16019 | — | — |
| Rondanina | 010050 | 16025 | — | — |
| Rossiglione | 010051 | 16010 | — | — |
| Rovegno | 010052 | 16028 | — | — |
| San Bartolomeo al Mare | 008052 | 18016 | — | — |
| San Biagio della Cima | 008053 | 18036 | — | — |
| San Colombano Certenoli | 010053 | 16040 | — | — |
| San Lorenzo al Mare | 008054 | 18017 | — | — |
| Sanremo | 008055 | 18038 | — | — |
| Sant'Olcese | 010055 | 16010 | — | — |
| Santa Margherita Ligure | 010054 | 16038 | — | — |
| Santo Stefano al Mare | 008056 | 18010 | — | — |
| Santo Stefano d'Aveto | 010056 | 16049 | — | — |
| Santo Stefano di Magra | 011026 | 19037 | — | — |
| Sarzana | 011027 | 19038 | — | — |
| Sassello | 009055 | 17046 | — | — |
| Savignone | 010057 | 16010 | — | — |
| Savona | 009056 | 17100 | — | — |
| Seborga | 008057 | 18012 | — | — |
| Serra Riccò | 010058 | 16010 | — | — |
| Sesta Godano | 011028 | 19020 | — | — |
| Sestri Levante | 010059 | 16039 | — | — |
| Soldano | 008058 | 18036 | — | — |
| Sori | 010060 | 16031 | — | — |
| Spotorno | 009057 | 17028 | — | — |
| Stella | 009058 | 17044 | — | — |
| Stellanello | 009059 | 17020 | — | — |
| Taggia | 008059 | 18018 | — | — |
| Terzorio | 008060 | 18010 | — | — |
| Testico | 009060 | 17020 | — | — |
| Tiglieto | 010061 | 16010 | — | — |
| Toirano | 009061 | 17055 | — | — |
| Torriglia | 010062 | 16029 | — | — |
| Tovo San Giacomo | 009062 | 17020 | — | — |
| Tribogna | 010063 | 16030 | — | — |
| Triora | 008061 | 18010 | — | — |
| Urbe | 009063 | 17048 | — | — |
| Uscio | 010064 | 16036 | — | — |
| Vado Ligure | 009064 | 17047 | — | — |
| Valbrevenna | 010065 | 16010 | — | — |
| Vallebona | 008062 | 18012 | — | — |
| Vallecrosia | 008063 | 18019 | — | — |
| Varazze | 009065 | 17019 | — | — |
| Varese Ligure | 011029 | 19028 | — | — |
| Vasia | 008064 | 18020 | — | — |
| Vendone | 009066 | 17032 | — | — |
| Ventimiglia | 008065 | 18039 | — | — |
| Vernazza | 011030 | 19018 | — | — |
| Vessalico | 008066 | 18026 | — | — |
| Vezzano Ligure | 011031 | 19020 | — | — |
| Vezzi Portio | 009067 | 17028 | — | — |
| Villa Faraldi | 008067 | 18010 | — | — |
| Villanova d'Albenga | 009068 | 17038 | — | — |
| Vobbia | 010066 | 16010 | — | — |
| Zignago | 011032 | 19020 | — | — |
| Zoagli | 010067 | 16035 | — | — |
| Zuccarello | 009069 | 17039 | — | — |

<a id="region-03"></a>

### Lombardia (1,501 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Lombardia — region** | 03 | — | Not started | Regional sources not checked |
| Abbadia Cerreto | 098001 | 26834 | — | — |
| Abbadia Lariana | 097001 | 23821 | — | — |
| Abbiategrasso | 015002 | 20081 | — | — |
| Acquafredda | 017001 | 25010 | — | — |
| Acquanegra Cremonese | 019001 | 26020 | — | — |
| Acquanegra sul Chiese | 020001 | 46011 | — | — |
| Adrara San Martino | 016001 | 24060 | — | — |
| Adrara San Rocco | 016002 | 24060 | — | — |
| Adro | 017002 | 25030 | — | — |
| Agnadello | 019002 | 26020 | — | — |
| Agnosine | 017003 | 25071 | — | — |
| Agra | 012001 | 21010 | — | — |
| Agrate Brianza | 108001 | 20864 | — | — |
| Aicurzio | 108002 | 20886 | — | — |
| Airuno | 097002 | 23881 | — | — |
| Alagna | 018001 | 27020 | — | — |
| Albairate | 015005 | 20080 | — | — |
| Albano Sant'Alessandro | 016003 | 24061 | — | — |
| Albaredo per San Marco | 014001 | 23010 | — | — |
| Albavilla | 013003 | 22031 | — | — |
| Albese con Cassano | 013004 | 22032 | — | — |
| Albiate | 108003 | 20847 | — | — |
| Albino | 016004 | 24021 | — | — |
| Albiolo | 013005 | 22070 | — | — |
| Albizzate | 012002 | 21041 | — | — |
| Albonese | 018003 | 27020 | — | — |
| Albosaggia | 014002 | 23010 | — | — |
| Albuzzano | 018004 | 27010 | — | — |
| Alfianello | 017004 | 25020 | — | — |
| Algua | 016248 | 24010 | — | — |
| Almè | 016005 | 24011 | — | — |
| Almenno San Bartolomeo | 016006 | 24030 | — | — |
| Almenno San Salvatore | 016007 | 24031 | — | — |
| Alserio | 013006 | 22040 | — | — |
| Alta Valle Intelvi | 013253 | 22024 | — | — |
| Alzano Lombardo | 016008 | 24022 | — | — |
| Alzate Brianza | 013007 | 22040 | — | — |
| Ambivere | 016009 | 24030 | — | — |
| Andalo Valtellino | 014003 | 23009 | — | — |
| Anfo | 017005 | 25070 | — | — |
| Angera | 012003 | 21021 | — | — |
| Angolo Terme | 017006 | 25040 | — | — |
| Annicco | 019003 | 26021 | — | — |
| Annone di Brianza | 097003 | 23841 | — | — |
| Antegnate | 016010 | 24051 | — | — |
| Anzano del Parco | 013009 | 22040 | — | — |
| Appiano Gentile | 013010 | 22070 | — | — |
| Aprica | 014004 | 23031 | — | — |
| Arcene | 016011 | 24040 | — | — |
| Arcisate | 012004 | 21051 | — | — |
| Arconate | 015007 | 20020 | — | — |
| Arcore | 108004 | 20862 | — | — |
| Ardenno | 014005 | 23011 | — | — |
| Ardesio | 016012 | 24020 | — | — |
| Arena Po | 018005 | 27040 | — | — |
| Arese | 015009 | 20044 | — | — |
| Argegno | 013011 | 22010 | — | — |
| Arluno | 015010 | 20004 | — | — |
| Arosio | 013012 | 22060 | — | — |
| Arsago Seprio | 012005 | 21010 | — | — |
| Artogne | 017007 | 25040 | — | — |
| Arzago d'Adda | 016013 | 24040 | — | — |
| Asola | 020002 | 46041 | — | — |
| Assago | 015011 | 20057 | — | — |
| Asso | 013013 | 22033 | — | — |
| Averara | 016014 | 24010 | — | — |
| Aviatico | 016015 | 24020 | — | — |
| Azzanello | 019004 | 26010 | — | — |
| Azzano Mella | 017008 | 25020 | — | — |
| Azzano San Paolo | 016016 | 24052 | — | — |
| Azzate | 012006 | 21022 | — | — |
| Azzio | 012007 | 21030 | — | — |
| Azzone | 016017 | 24020 | — | — |
| Badia Pavese | 018006 | 27010 | — | — |
| Bagnaria | 018007 | 27050 | — | — |
| Bagnatica | 016018 | 24060 | — | — |
| Bagnolo Cremasco | 019005 | 26010 | — | — |
| Bagnolo Mella | 017009 | 25021 | — | — |
| Bagnolo San Vito | 020003 | 46031 | — | — |
| Bagolino | 017010 | 25072 | — | — |
| Ballabio | 097004 | 23811 | — | — |
| Baranzate | 015250 | 20021 | — | — |
| Barasso | 012008 | 21020 | — | — |
| Barbariga | 017011 | 25030 | — | — |
| Barbata | 016019 | 24040 | — | — |
| Barbianello | 018008 | 27041 | — | — |
| Bardello con Malgesso e Bregano | 012144 | 21009 | — | — |
| Bareggio | 015012 | 20008 | — | — |
| Barghe | 017012 | 25070 | — | — |
| Bariano | 016020 | 24050 | — | — |
| Barlassina | 108005 | 20825 | — | — |
| Barni | 013015 | 22030 | — | — |
| Barzago | 097005 | 23890 | — | — |
| Barzana | 016021 | 24030 | — | — |
| Barzanò | 097006 | 23891 | — | — |
| Barzio | 097007 | 23816 | — | — |
| Bascapè | 018009 | 27010 | — | — |
| Basiano | 015014 | 20060 | — | — |
| Basiglio | 015015 | 20079 | — | — |
| Bassano Bresciano | 017013 | 25020 | — | — |
| Bastida Pancarana | 018011 | 27050 | — | — |
| Battuda | 018012 | 27020 | — | — |
| Bedero Valcuvia | 012010 | 21039 | — | — |
| Bedizzole | 017014 | 25081 | — | — |
| Bedulita | 016022 | 24030 | — | — |
| Belgioioso | 018013 | 27011 | — | — |
| Bellagio | 013250 | 22021 | — | — |
| Bellano | 097008 | 23822 | — | — |
| Bellinzago Lombardo | 015016 | 20060 | — | — |
| Bellusco | 108006 | 20882 | — | — |
| Bema | 014006 | 23010 | — | — |
| Bene Lario | 013021 | 22010 | — | — |
| Berbenno | 016023 | 24030 | — | — |
| Berbenno di Valtellina | 014007 | 23010 | — | — |
| Beregazzo con Figliaro | 013022 | 22070 | — | — |
| Bereguardo | 018014 | 27021 | — | — |
| Bergamo | 016024 | 24121, 24122, 24123, 24124, 24125, 24126, 24127, 24128, 24129 | — | — |
| Berlingo | 017015 | 25030 | — | — |
| Bernareggio | 108007 | 20881 | — | — |
| Bernate Ticino | 015019 | 20010 | — | — |
| Bertonico | 098002 | 26821 | — | — |
| Berzo Demo | 017016 | 25040 | — | — |
| Berzo Inferiore | 017017 | 25040 | — | — |
| Berzo San Fermo | 016025 | 24060 | — | — |
| Besana in Brianza | 108008 | 20842 | — | — |
| Besano | 012011 | 21050 | — | — |
| Besate | 015022 | 20080 | — | — |
| Besnate | 012012 | 21010 | — | — |
| Besozzo | 012013 | 21023 | — | — |
| Biandronno | 012014 | 21024 | — | — |
| Bianzano | 016026 | 24060 | — | — |
| Bianzone | 014008 | 23030 | — | — |
| Biassono | 108009 | 20853 | — | — |
| Bienno | 017018 | 25040 | — | — |
| Binago | 013023 | 22070 | — | — |
| Binasco | 015024 | 20082 | — | — |
| Bione | 017019 | 25070 | — | — |
| Bisuschio | 012015 | 21050 | — | — |
| Bizzarone | 013024 | 22020 | — | — |
| Blello | 016027 | 24010 | — | — |
| Blessagno | 013025 | 22028 | — | — |
| Blevio | 013026 | 22020 | — | — |
| Bodio Lomnago | 012016 | 21020 | — | — |
| Boffalora d'Adda | 098003 | 26811 | — | — |
| Boffalora sopra Ticino | 015026 | 20010 | — | — |
| Bolgare | 016028 | 24060 | — | — |
| Bollate | 015027 | 20021 | — | — |
| Boltiere | 016029 | 24040 | — | — |
| Bonate Sopra | 016030 | 24040 | — | — |
| Bonate Sotto | 016031 | 24040 | — | — |
| Bonemerse | 019006 | 26040 | — | — |
| Bordolano | 019007 | 26020 | — | — |
| Borgarello | 018015 | 27010 | — | — |
| Borghetto Lodigiano | 098004 | 26812 | — | — |
| Borgo di Terzo | 016032 | 24060 | — | — |
| Borgo Mantovano | 020072 | 46036 | — | — |
| Borgo Priolo | 018016 | 27040 | — | — |
| Borgo San Giacomo | 017020 | 25022 | — | — |
| Borgo San Giovanni | 098005 | 26851 | — | — |
| Borgo San Siro | 018018 | 27020 | — | — |
| Borgo Virgilio | 020071 | 46034 | — | — |
| Borgocarbonara | 020073 | 46021 | — | — |
| Borgoratto Mormorolo | 018017 | 27040 | — | — |
| Borgosatollo | 017021 | 25010 | — | — |
| Bormio | 014009 | 23032 | — | — |
| Bornasco | 018019 | 27010 | — | — |
| Borno | 017022 | 25042 | — | — |
| Bosisio Parini | 097009 | 23842 | — | — |
| Bosnasco | 018020 | 27040 | — | — |
| Bossico | 016033 | 24060 | — | — |
| Bottanuco | 016034 | 24040 | — | — |
| Botticino | 017023 | 25082 | — | — |
| Bovegno | 017024 | 25061 | — | — |
| Bovezzo | 017025 | 25073 | — | — |
| Bovisio-Masciago | 108010 | 20813 | — | — |
| Bozzolo | 020007 | 46012 | — | — |
| Bracca | 016035 | 24010 | — | — |
| Brallo di Pregola | 018021 | 27050 | — | — |
| Brandico | 017026 | 25030 | — | — |
| Branzi | 016036 | 24010 | — | — |
| Braone | 017027 | 25040 | — | — |
| Brebbia | 012017 | 21020 | — | — |
| Bregnano | 013028 | 22070 | — | — |
| Brembate | 016037 | 24041 | — | — |
| Brembate di Sopra | 016038 | 24030 | — | — |
| Brembio | 098006 | 26822 | — | — |
| Breme | 018022 | 27020 | — | — |
| Brenna | 013029 | 22040 | — | — |
| Breno | 017028 | 25043 | — | — |
| Brenta | 012019 | 21030 | — | — |
| Brescia | 017029 | 25121, 25122, 25123, 25124, 25125, 25126, 25127, 25128, 25129, 25131, 25132, 25133, 25134, 25135, 25136 | — | — |
| Bressana Bottarone | 018023 | 27042 | — | — |
| Bresso | 015032 | 20091 | — | — |
| Brezzo di Bedero | 012020 | 21010 | — | — |
| Brienno | 013030 | 22010 | — | — |
| Brignano Gera d'Adda | 016040 | 24053 | — | — |
| Brinzio | 012021 | 21030 | — | — |
| Brione | 017030 | 25060 | — | — |
| Briosco | 108011 | 20836 | — | — |
| Brissago-Valtravaglia | 012022 | 21030 | — | — |
| Brivio | 097010 | 23883 | — | — |
| Broni | 018024 | 27043 | — | — |
| Brugherio | 108012 | 20861 | — | — |
| Brumano | 016041 | 24037 | — | — |
| Brunate | 013032 | 22034 | — | — |
| Brunello | 012023 | 21020 | — | — |
| Brusaporto | 016042 | 24060 | — | — |
| Brusimpiano | 012024 | 21050 | — | — |
| Bubbiano | 015035 | 20080 | — | — |
| Buccinasco | 015036 | 20090 | — | — |
| Buglio in Monte | 014010 | 23010 | — | — |
| Buguggiate | 012025 | 21020 | — | — |
| Bulciago | 097011 | 23892 | — | — |
| Bulgarograsso | 013034 | 22070 | — | — |
| Burago di Molgora | 108013 | 20875 | — | — |
| Buscate | 015038 | 20010 | — | — |
| Busnago | 108051 | 20874 | — | — |
| Bussero | 015040 | 20041 | — | — |
| Busto Arsizio | 012026 | 21052 | — | — |
| Busto Garolfo | 015041 | 20038 | — | — |
| Cabiate | 013035 | 22060 | — | — |
| Cadegliano-Viconago | 012027 | 21031 | — | — |
| Cadorago | 013036 | 22071 | — | — |
| Cadrezzate con Osmate | 012143 | 21062 | — | — |
| Caglio | 013037 | 22030 | — | — |
| Caino | 017031 | 25070 | — | — |
| Caiolo | 014011 | 23010 | — | — |
| Cairate | 012029 | 21050 | — | — |
| Calcinate | 016043 | 24050 | — | — |
| Calcinato | 017032 | 25011 | — | — |
| Calcio | 016044 | 24054 | — | — |
| Calco | 097012 | 23885 | — | — |
| Calolziocorte | 097013 | 23801 | — | — |
| Calusco d'Adda | 016046 | 24033 | — | — |
| Calvagese della Riviera | 017033 | 25080 | — | — |
| Calvatone | 019009 | 26030 | — | — |
| Calvenzano | 016047 | 24040 | — | — |
| Calvignano | 018025 | 27040 | — | — |
| Calvignasco | 015042 | 20080 | — | — |
| Calvisano | 017034 | 25012 | — | — |
| Cambiago | 015044 | 20040 | — | — |
| Camerata Cornello | 016048 | 24010 | — | — |
| Camisano | 019010 | 26010 | — | — |
| Campagnola Cremasca | 019011 | 26010 | — | — |
| Camparada | 108014 | 20857 | — | — |
| Campione d'Italia | 013040 | 22061 | — | — |
| Campodolcino | 014012 | 23021 | — | — |
| Campospinoso Albaredo | 018026 | 27062 | — | — |
| Candia Lomellina | 018027 | 27031 | — | — |
| Canegrate | 015046 | 20039 | — | — |
| Canneto Pavese | 018029 | 27044 | — | — |
| Canneto sull'Oglio | 020008 | 46013 | — | — |
| Canonica d'Adda | 016049 | 24040 | — | — |
| Cantello | 012030 | 21050 | — | — |
| Cantù | 013041 | 22063 | — | — |
| Canzo | 013042 | 22035 | — | — |
| Capergnanica | 019012 | 26010 | — | — |
| Capiago Intimiano | 013043 | 22070 | — | — |
| Capizzone | 016050 | 24030 | — | — |
| Capo di Ponte | 017035 | 25044 | — | — |
| Caponago | 108052 | 20867 | — | — |
| Capovalle | 017036 | 25070 | — | — |
| Cappella Cantone | 019013 | 26020 | — | — |
| Cappella de' Picenardi | 019014 | 26030 | — | — |
| Capralba | 019015 | 26010 | — | — |
| Capriano del Colle | 017037 | 25020 | — | — |
| Capriate San Gervasio | 016051 | 24042 | — | — |
| Caprino Bergamasco | 016052 | 24030 | — | — |
| Capriolo | 017038 | 25031 | — | — |
| Carate Brianza | 108015 | 20841 | — | — |
| Carate Urio | 013044 | 22010 | — | — |
| Caravaggio | 016053 | 24043 | — | — |
| Caravate | 012031 | 21032 | — | — |
| Carbonara al Ticino | 018030 | 27020 | — | — |
| Carbonate | 013045 | 22070 | — | — |
| Cardano al Campo | 012032 | 21010 | — | — |
| Carenno | 097014 | 23802 | — | — |
| Carimate | 013046 | 22060 | — | — |
| Carlazzo | 013047 | 22010 | — | — |
| Carnago | 012033 | 21040 | — | — |
| Carnate | 108016 | 20866 | — | — |
| Carobbio degli Angeli | 016055 | 24060 | — | — |
| Carona | 016056 | 24010 | — | — |
| Caronno Pertusella | 012034 | 21042 | — | — |
| Caronno Varesino | 012035 | 21040 | — | — |
| Carpenedolo | 017039 | 25013 | — | — |
| Carpiano | 015050 | 20074 | — | — |
| Carugate | 015051 | 20061 | — | — |
| Carugo | 013048 | 22060 | — | — |
| Carvico | 016057 | 24030 | — | — |
| Casalbuttano ed Uniti | 019016 | 26011 | — | — |
| Casale Cremasco-Vidolasco | 019017 | 26010 | — | — |
| Casale Litta | 012036 | 21020 | — | — |
| Casaletto Ceredano | 019018 | 26010 | — | — |
| Casaletto di Sopra | 019019 | 26014 | — | — |
| Casaletto Lodigiano | 098008 | 26852 | — | — |
| Casaletto Vaprio | 019020 | 26010 | — | — |
| Casalmaggiore | 019021 | 26041 | — | — |
| Casalmaiocco | 098009 | 26831 | — | — |
| Casalmorano | 019022 | 26020 | — | — |
| Casalmoro | 020010 | 46040 | — | — |
| Casaloldo | 020011 | 46040 | — | — |
| Casalpusterlengo | 098010 | 26841 | — | — |
| Casalromano | 020012 | 46040 | — | — |
| Casalzuigno | 012037 | 21030 | — | — |
| Casanova Lonati | 018031 | 27041 | — | — |
| Casargo | 097015 | 23831 | — | — |
| Casarile | 015055 | 20059 | — | — |
| Casatenovo | 097016 | 23880 | — | — |
| Casatisma | 018032 | 27040 | — | — |
| Casazza | 016058 | 24060 | — | — |
| Casciago | 012038 | 21020 | — | — |
| Casei Gerola | 018033 | 27050 | — | — |
| Caselle Landi | 098011 | 26842 | — | — |
| Caselle Lurani | 098012 | 26853 | — | — |
| Casirate d'Adda | 016059 | 24040 | — | — |
| Caslino d'Erba | 013052 | 22030 | — | — |
| Casnate con Bernate | 013053 | 22070 | — | — |
| Casnigo | 016060 | 24020 | — | — |
| Casorate Primo | 018034 | 27022 | — | — |
| Casorate Sempione | 012039 | 21011 | — | — |
| Casorezzo | 015058 | 20003 | — | — |
| Caspoggio | 014013 | 23020 | — | — |
| Cassago Brianza | 097017 | 23893 | — | — |
| Cassano d'Adda | 015059 | 20062 | — | — |
| Cassano Magnago | 012040 | 21012 | — | — |
| Cassano Valcuvia | 012041 | 21030 | — | — |
| Cassiglio | 016061 | 24010 | — | — |
| Cassina de' Pecchi | 015060 | 20051 | — | — |
| Cassina Rizzardi | 013055 | 22070 | — | — |
| Cassina Valsassina | 097018 | 23817 | — | — |
| Cassinetta di Lugagnano | 015061 | 20081 | — | — |
| Cassolnovo | 018035 | 27023 | — | — |
| Castana | 018036 | 27040 | — | — |
| Castano Primo | 015062 | 20022 | — | — |
| Casteggio | 018037 | 27045 | — | — |
| Castegnato | 017040 | 25045 | — | — |
| Castel d'Ario | 020014 | 46033 | — | — |
| Castel Gabbiano | 019024 | 26010 | — | — |
| Castel Goffredo | 020015 | 46042 | — | — |
| Castel Mella | 017042 | 25030 | — | — |
| Castel Rozzone | 016063 | 24040 | — | — |
| Castelbelforte | 020013 | 46032 | — | — |
| Castelcovati | 017041 | 25030 | — | — |
| Casteldidone | 019023 | 26030 | — | — |
| Castelgerundo | 098062 | 26844 | — | — |
| Castellanza | 012042 | 21053 | — | — |
| Castelleone | 019025 | 26012 | — | — |
| Castelletto di Branduzzo | 018038 | 27040 | — | — |
| Castelli Calepio | 016062 | 24060 | — | — |
| Castello Cabiaglio | 012043 | 21030 | — | — |
| Castello d'Agogna | 018039 | 27030 | — | — |
| Castello dell'Acqua | 014014 | 23030 | — | — |
| Castello di Brianza | 097019 | 23884 | — | — |
| Castellucchio | 020016 | 46014 | — | — |
| Castelmarte | 013058 | 22030 | — | — |
| Castelnovetto | 018040 | 27030 | — | — |
| Castelnuovo Bocca d'Adda | 098013 | 26843 | — | — |
| Castelnuovo Bozzente | 013059 | 22070 | — | — |
| Castelseprio | 012044 | 21050 | — | — |
| Castelveccana | 012045 | 21010 | — | — |
| Castelverde | 019026 | 26022 | — | — |
| Castelvisconti | 019027 | 26010 | — | — |
| Castenedolo | 017043 | 25014 | — | — |
| Castiglione d'Adda | 098014 | 26823 | — | — |
| Castiglione delle Stiviere | 020017 | 46043 | — | — |
| Castiglione Olona | 012046 | 21043 | — | — |
| Castione Andevenno | 014015 | 23012 | — | — |
| Castione della Presolana | 016064 | 24020 | — | — |
| Castiraga Vidardo | 098015 | 26866 | — | — |
| Casto | 017044 | 25070 | — | — |
| Castrezzato | 017045 | 25030 | — | — |
| Castro | 016065 | 24063 | — | — |
| Castronno | 012047 | 21040 | — | — |
| Cava Manara | 018041 | 27051 | — | — |
| Cavargna | 013062 | 22010 | — | — |
| Cavaria con Premezzo | 012048 | 21044 | — | — |
| Cavenago d'Adda | 098017 | 26824 | — | — |
| Cavenago di Brianza | 108017 | 20873 | — | — |
| Cavernago | 016066 | 24050 | — | — |
| Cavriana | 020018 | 46040 | — | — |
| Cazzago Brabbia | 012049 | 21020 | — | — |
| Cazzago San Martino | 017046 | 25046 | — | — |
| Cazzano Sant'Andrea | 016067 | 24026 | — | — |
| Cecima | 018042 | 27050 | — | — |
| Cedegolo | 017047 | 25051 | — | — |
| Cedrasco | 014016 | 23010 | — | — |
| Cella Dati | 019028 | 26040 | — | — |
| Cellatica | 017048 | 25060 | — | — |
| Cenate Sopra | 016068 | 24060 | — | — |
| Cenate Sotto | 016069 | 24069 | — | — |
| Cene | 016070 | 24020 | — | — |
| Centro Valle Intelvi | 013254 | 22023 | — | — |
| Cerano d'Intelvi | 013063 | 22020 | — | — |
| Ceranova | 018043 | 27010 | — | — |
| Cercino | 014017 | 23016 | — | — |
| Ceresara | 020019 | 46040 | — | — |
| Cerete | 016071 | 24020 | — | — |
| Ceretto Lomellina | 018044 | 27030 | — | — |
| Cergnago | 018045 | 27020 | — | — |
| Ceriano Laghetto | 108018 | 20816 | — | — |
| Cermenate | 013064 | 22072 | — | — |
| Cernobbio | 013065 | 22012 | — | — |
| Cernusco Lombardone | 097020 | 23870 | — | — |
| Cernusco sul Naviglio | 015070 | 20063 | — | — |
| Cerro al Lambro | 015071 | 20070 | — | — |
| Cerro Maggiore | 015072 | 20023 | — | — |
| Certosa di Pavia | 018046 | 27012 | — | — |
| Cerveno | 017049 | 25040 | — | — |
| Cervesina | 018047 | 27050 | — | — |
| Cervignano d'Adda | 098018 | 26832 | — | — |
| Cesana Brianza | 097021 | 23861 | — | — |
| Cesano Boscone | 015074 | 20090 | — | — |
| Cesano Maderno | 108019 | 20811 | — | — |
| Cesate | 015076 | 20031 | — | — |
| Ceto | 017050 | 25040 | — | — |
| Cevo | 017051 | 25040 | — | — |
| Chiari | 017052 | 25032 | — | — |
| Chiavenna | 014018 | 23022 | — | — |
| Chiesa in Valmalenco | 014019 | 23023 | — | — |
| Chieve | 019029 | 26010 | — | — |
| Chignolo d'Isola | 016072 | 24040 | — | — |
| Chignolo Po | 018048 | 27013 | — | — |
| Chiuduno | 016073 | 24060 | — | — |
| Chiuro | 014020 | 23030 | — | — |
| Cicognolo | 019030 | 26030 | — | — |
| Cigognola | 018049 | 27040 | — | — |
| Cigole | 017053 | 25020 | — | — |
| Cilavegna | 018050 | 27024 | — | — |
| Cimbergo | 017054 | 25050 | — | — |
| Cingia de' Botti | 019031 | 26042 | — | — |
| Cinisello Balsamo | 015077 | 20092 | — | — |
| Cino | 014021 | 23010 | — | — |
| Cirimido | 013068 | 22070 | — | — |
| Cisano Bergamasco | 016074 | 24034 | — | — |
| Ciserano | 016075 | 24040 | — | — |
| Cislago | 012050 | 21040 | — | — |
| Cisliano | 015078 | 20046 | — | — |
| Cittiglio | 012051 | 21033 | — | — |
| Civate | 097022 | 23862 | — | — |
| Cividate al Piano | 016076 | 24050 | — | — |
| Cividate Camuno | 017055 | 25040 | — | — |
| Civo | 014022 | 23010 | — | — |
| Claino con Osteno | 013071 | 22010 | — | — |
| Clivio | 012052 | 21050 | — | — |
| Clusone | 016077 | 24023 | — | — |
| Coccaglio | 017056 | 25030 | — | — |
| Cocquio-Trevisago | 012053 | 21034 | — | — |
| Codevilla | 018051 | 27050 | — | — |
| Codogno | 098019 | 26845 | — | — |
| Cogliate | 108020 | 20815 | — | — |
| Colere | 016078 | 24020 | — | — |
| Colico | 097023 | 23823 | — | — |
| Colle Brianza | 097024 | 23886 | — | — |
| Collebeato | 017057 | 25060 | — | — |
| Colli Verdi | 018193 | 27061 | — | — |
| Collio | 017058 | 25060 | — | — |
| Cologne | 017059 | 25033 | — | — |
| Cologno al Serio | 016079 | 24055 | — | — |
| Cologno Monzese | 015081 | 20093 | — | — |
| Colonno | 013074 | 22010 | — | — |
| Colorina | 014023 | 23010 | — | — |
| Colturano | 015082 | 20075 | — | — |
| Colverde | 013251 | 22041 | — | — |
| Colzate | 016080 | 24020 | — | — |
| Comabbio | 012054 | 21020 | — | — |
| Comazzo | 098020 | 26833 | — | — |
| Comerio | 012055 | 21025 | — | — |
| Comezzano-Cizzago | 017060 | 25030 | — | — |
| Commessaggio | 020020 | 46010 | — | — |
| Como | 013075 | 22100 | — | — |
| Comun Nuovo | 016081 | 24040 | — | — |
| Concesio | 017061 | 25062 | — | — |
| Concorezzo | 108021 | 20863 | — | — |
| Confienza | 018052 | 27030 | — | — |
| Copiano | 018053 | 27010 | — | — |
| Corana | 018054 | 27050 | — | — |
| Corbetta | 015085 | 20011 | — | — |
| Cormano | 015086 | 20032 | — | — |
| Corna Imagna | 016082 | 24030 | — | — |
| Cornalba | 016249 | 24017 | — | — |
| Cornale e Bastida | 018191 | 27056 | — | — |
| Cornaredo | 015087 | 20007 | — | — |
| Cornate d'Adda | 108053 | 20872 | — | — |
| Cornegliano Laudense | 098021 | 26854 | — | — |
| Corno Giovine | 098022 | 26846 | — | — |
| Cornovecchio | 098023 | 26842 | — | — |
| Correzzana | 108022 | 20856 | — | — |
| Corrido | 013077 | 22010 | — | — |
| Corsico | 015093 | 20094 | — | — |
| Corte de' Cortesi con Cignone | 019032 | 26020 | — | — |
| Corte de' Frati | 019033 | 26010 | — | — |
| Corte Franca | 017062 | 25040 | — | — |
| Corte Palasio | 098024 | 26834 | — | — |
| Corteno Golgi | 017063 | 25040 | — | — |
| Cortenova | 097025 | 23813 | — | — |
| Cortenuova | 016083 | 24050 | — | — |
| Corteolona e Genzone | 018192 | 27014 | — | — |
| Corvino San Quirico | 018057 | 27050 | — | — |
| Corzano | 017064 | 25030 | — | — |
| Cosio Valtellino | 014024 | 23013 | — | — |
| Costa de' Nobili | 018058 | 27010 | — | — |
| Costa di Mezzate | 016084 | 24060 | — | — |
| Costa Masnaga | 097026 | 23845 | — | — |
| Costa Serina | 016247 | 24010 | — | — |
| Costa Valle Imagna | 016085 | 24030 | — | — |
| Costa Volpino | 016086 | 24062 | — | — |
| Covo | 016087 | 24050 | — | — |
| Cozzo | 018059 | 27030 | — | — |
| Crandola Valsassina | 097027 | 23832 | — | — |
| Credaro | 016088 | 24060 | — | — |
| Credera Rubbiano | 019034 | 26010 | — | — |
| Crema | 019035 | 26013 | — | — |
| Cremella | 097028 | 23894 | — | — |
| Cremenaga | 012056 | 21030 | — | — |
| Cremeno | 097029 | 23814 | — | — |
| Cremia | 013083 | 22010 | — | — |
| Cremona | 019036 | 26100 | — | — |
| Cremosano | 019037 | 26010 | — | — |
| Crespiatica | 098025 | 26835 | — | — |
| Crosio della Valle | 012057 | 21020 | — | — |
| Crotta d'Adda | 019038 | 26020 | — | — |
| Cuasso al Monte | 012058 | 21050 | — | — |
| Cucciago | 013084 | 22060 | — | — |
| Cuggiono | 015096 | 20012 | — | — |
| Cugliate-Fabiasco | 012059 | 21030 | — | — |
| Cumignano sul Naviglio | 019039 | 26020 | — | — |
| Cunardo | 012060 | 21035 | — | — |
| Cura Carpignano | 018060 | 27010 | — | — |
| Curiglia con Monteviasco | 012061 | 21010 | — | — |
| Curno | 016089 | 24035 | — | — |
| Curtatone | 020021 | 46010 | — | — |
| Cusago | 015097 | 20047 | — | — |
| Cusano Milanino | 015098 | 20095 | — | — |
| Cusino | 013085 | 22010 | — | — |
| Cusio | 016090 | 24010 | — | — |
| Cuveglio | 012062 | 21030 | — | — |
| Cuvio | 012063 | 21030 | — | — |
| Dairago | 015099 | 20036 | — | — |
| Dalmine | 016091 | 24044 | — | — |
| Darfo Boario Terme | 017065 | 25047 | — | — |
| Daverio | 012064 | 21020 | — | — |
| Dazio | 014025 | 23010 | — | — |
| Delebio | 014026 | 23014 | — | — |
| Dello | 017066 | 25020 | — | — |
| Derovere | 019040 | 26040 | — | — |
| Dervio | 097030 | 23824 | — | — |
| Desenzano del Garda | 017067 | 25015 | — | — |
| Desio | 108023 | 20832 | — | — |
| Dizzasco | 013087 | 22020 | — | — |
| Dolzago | 097031 | 23843 | — | — |
| Domaso | 013089 | 22013 | — | — |
| Dongo | 013090 | 22014 | — | — |
| Dorio | 097032 | 23824 | — | — |
| Dorno | 018061 | 27020 | — | — |
| Dosolo | 020022 | 46030 | — | — |
| Dossena | 016092 | 24010 | — | — |
| Dosso del Liro | 013092 | 22010 | — | — |
| Dovera | 019041 | 26010 | — | — |
| Dresano | 015101 | 20070 | — | — |
| Dubino | 014027 | 23015 | — | — |
| Dumenza | 012065 | 21010 | — | — |
| Duno | 012066 | 21030 | — | — |
| Edolo | 017068 | 25048 | — | — |
| Ello | 097033 | 23848 | — | — |
| Endine Gaiano | 016093 | 24060 | — | — |
| Entratico | 016094 | 24060 | — | — |
| Erba | 013095 | 22036 | — | — |
| Erbusco | 017069 | 25030 | — | — |
| Erve | 097034 | 23805 | — | — |
| Esine | 017070 | 25040 | — | — |
| Esino Lario | 097035 | 23825 | — | — |
| Eupilio | 013097 | 22030 | — | — |
| Faedo Valtellino | 014028 | 23020 | — | — |
| Faggeto Lario | 013098 | 22020 | — | — |
| Fagnano Olona | 012067 | 21054 | — | — |
| Faloppio | 013099 | 22020 | — | — |
| Fara Gera d'Adda | 016096 | 24045 | — | — |
| Fara Olivana con Sola | 016097 | 24058 | — | — |
| Fenegrò | 013100 | 22070 | — | — |
| Ferno | 012068 | 21010 | — | — |
| Ferrera di Varese | 012069 | 21030 | — | — |
| Ferrera Erbognone | 018062 | 27032 | — | — |
| Fiesco | 019043 | 26010 | — | — |
| Fiesse | 017071 | 25020 | — | — |
| Figino Serenza | 013101 | 22060 | — | — |
| Filago | 016098 | 24040 | — | — |
| Filighera | 018063 | 27010 | — | — |
| Fino del Monte | 016099 | 24020 | — | — |
| Fino Mornasco | 013102 | 22073 | — | — |
| Fiorano al Serio | 016100 | 24020 | — | — |
| Flero | 017072 | 25020 | — | — |
| Fombio | 098026 | 26861 | — | — |
| Fontanella | 016101 | 24056 | — | — |
| Fonteno | 016102 | 24060 | — | — |
| Foppolo | 016103 | 24010 | — | — |
| Forcola | 014029 | 23010 | — | — |
| Foresto Sparso | 016104 | 24060 | — | — |
| Formigara | 019044 | 26020 | — | — |
| Fornovo San Giovanni | 016105 | 24040 | — | — |
| Fortunago | 018064 | 27040 | — | — |
| Frascarolo | 018065 | 27030 | — | — |
| Fuipiano Valle Imagna | 016106 | 24030 | — | — |
| Fusine | 014030 | 23010 | — | — |
| Gabbioneta-Binanuova | 019045 | 26030 | — | — |
| Gadesco-Pieve Delmona | 019046 | 26030 | — | — |
| Gaggiano | 015103 | 20083 | — | — |
| Galbiate | 097036 | 23851 | — | — |
| Galgagnano | 098027 | 26832 | — | — |
| Gallarate | 012070 | 21013 | — | — |
| Galliate Lombardo | 012071 | 21020 | — | — |
| Galliavola | 018066 | 27034 | — | — |
| Gambara | 017073 | 25020 | — | — |
| Gambarana | 018067 | 27030 | — | — |
| Gambolò | 018068 | 27025 | — | — |
| Gandellino | 016107 | 24020 | — | — |
| Gandino | 016108 | 24024 | — | — |
| Gandosso | 016109 | 24060 | — | — |
| Garbagnate Milanese | 015105 | 20024 | — | — |
| Garbagnate Monastero | 097037 | 23846 | — | — |
| Gardone Riviera | 017074 | 25083 | — | — |
| Gardone Val Trompia | 017075 | 25063 | — | — |
| Gargnano | 017076 | 25084 | — | — |
| Garlasco | 018069 | 27026 | — | — |
| Garlate | 097038 | 23852 | — | — |
| Garzeno | 013106 | 22010 | — | — |
| Gavardo | 017077 | 25085 | — | — |
| Gaverina Terme | 016110 | 24060 | — | — |
| Gavirate | 012072 | 21026 | — | — |
| Gazoldo degli Ippoliti | 020024 | 46040 | — | — |
| Gazzada Schianno | 012073 | 21045 | — | — |
| Gazzaniga | 016111 | 24025 | — | — |
| Gazzuolo | 020025 | 46010 | — | — |
| Gemonio | 012074 | 21036 | — | — |
| Genivolta | 019047 | 26020 | — | — |
| Gera Lario | 013107 | 22010 | — | — |
| Gerenzago | 018071 | 27010 | — | — |
| Gerenzano | 012075 | 21040 | — | — |
| Germignaga | 012076 | 21010 | — | — |
| Gerola Alta | 014031 | 23010 | — | — |
| Gerre de' Caprioli | 019048 | 26040 | — | — |
| Gessate | 015106 | 20060 | — | — |
| Ghedi | 017078 | 25016 | — | — |
| Ghisalba | 016113 | 24050 | — | — |
| Gianico | 017079 | 25040 | — | — |
| Giussago | 018072 | 27010 | — | — |
| Giussano | 108024 | 20833 | — | — |
| Godiasco Salice Terme | 018073 | 27052 | — | — |
| Goito | 020026 | 46044 | — | — |
| Golasecca | 012077 | 21010 | — | — |
| Golferenzo | 018074 | 27047 | — | — |
| Gombito | 019049 | 26020 | — | — |
| Gonzaga | 020027 | 46023 | — | — |
| Gordona | 014032 | 23020 | — | — |
| Gorgonzola | 015108 | 20064 | — | — |
| Gorla Maggiore | 012078 | 21050 | — | — |
| Gorla Minore | 012079 | 21055 | — | — |
| Gorlago | 016114 | 24060 | — | — |
| Gorle | 016115 | 24020 | — | — |
| Gornate Olona | 012080 | 21040 | — | — |
| Gorno | 016116 | 24020 | — | — |
| Gottolengo | 017080 | 25023 | — | — |
| Graffignana | 098028 | 26813 | — | — |
| Grandate | 013110 | 22070 | — | — |
| Grandola ed Uniti | 013111 | 22010 | — | — |
| Grantola | 012081 | 21030 | — | — |
| Grassobbio | 016117 | 24050 | — | — |
| Gravedona ed Uniti | 013249 | 22015 | — | — |
| Gravellona Lomellina | 018075 | 27020 | — | — |
| Grezzago | 015110 | 20056 | — | — |
| Griante | 013113 | 22011 | — | — |
| Gromo | 016118 | 24020 | — | — |
| Grone | 016119 | 24060 | — | — |
| Grontardo | 019050 | 26044 | — | — |
| Gropello Cairoli | 018076 | 27027 | — | — |
| Grosio | 014033 | 23033 | — | — |
| Grosotto | 014034 | 23034 | — | — |
| Grumello Cremonese ed Uniti | 019051 | 26023 | — | — |
| Grumello del Monte | 016120 | 24064 | — | — |
| Guanzate | 013114 | 22070 | — | — |
| Guardamiglio | 098029 | 26862 | — | — |
| Gudo Visconti | 015112 | 20088 | — | — |
| Guidizzolo | 020028 | 46040 | — | — |
| Gussago | 017081 | 25064 | — | — |
| Gussola | 019052 | 26040 | — | — |
| Idro | 017082 | 25074 | — | — |
| Imbersago | 097039 | 23898 | — | — |
| Inarzo | 012082 | 21020 | — | — |
| Incudine | 017083 | 25040 | — | — |
| Induno Olona | 012083 | 21056 | — | — |
| Introbio | 097040 | 23815 | — | — |
| Inverigo | 013118 | 22044 | — | — |
| Inverno e Monteleone | 018077 | 27010 | — | — |
| Inveruno | 015113 | 20001 | — | — |
| Inzago | 015114 | 20065 | — | — |
| Irma | 017084 | 25061 | — | — |
| Iseo | 017085 | 25049 | — | — |
| Isola di Fondra | 016121 | 24010 | — | — |
| Isola Dovarese | 019053 | 26031 | — | — |
| Isorella | 017086 | 25010 | — | — |
| Ispra | 012084 | 21027 | — | — |
| Isso | 016122 | 24040 | — | — |
| Izano | 019054 | 26010 | — | — |
| Jerago con Orago | 012085 | 21040 | — | — |
| La Valletta Brianza | 097092 | 23888 | — | — |
| Lacchiarella | 015115 | 20084 | — | — |
| Laglio | 013119 | 22010 | — | — |
| Lainate | 015116 | 20045 | — | — |
| Laino | 013120 | 22020 | — | — |
| Lallio | 016123 | 24040 | — | — |
| Lambrugo | 013121 | 22045 | — | — |
| Landriano | 018078 | 27015 | — | — |
| Langosco | 018079 | 27030 | — | — |
| Lanzada | 014036 | 23020 | — | — |
| Lardirago | 018080 | 27016 | — | — |
| Lasnigo | 013123 | 22030 | — | — |
| Lavena Ponte Tresa | 012086 | 21037 | — | — |
| Laveno-Mombello | 012087 | 21014 | — | — |
| Lavenone | 017087 | 25074 | — | — |
| Lazzate | 108025 | 20824 | — | — |
| Lecco | 097042 | 23900 | — | — |
| Leffe | 016124 | 24026 | — | — |
| Leggiuno | 012088 | 21038 | — | — |
| Legnano | 015118 | 20025 | — | — |
| Lenna | 016125 | 24010 | — | — |
| Leno | 017088 | 25024 | — | — |
| Lentate sul Seveso | 108054 | 20823 | — | — |
| Lesmo | 108026 | 20855 | — | — |
| Levate | 016126 | 24040 | — | — |
| Lezzeno | 013126 | 22025 | — | — |
| Lierna | 097043 | 23827 | — | — |
| Limbiate | 108027 | 20812 | — | — |
| Limido Comasco | 013128 | 22070 | — | — |
| Limone sul Garda | 017089 | 25010 | — | — |
| Linarolo | 018081 | 27010 | — | — |
| Lipomo | 013129 | 22030 | — | — |
| Liscate | 015122 | 20050 | — | — |
| Lissone | 108028 | 20851 | — | — |
| Livigno | 014037 | 23041 | — | — |
| Livo | 013130 | 22010 | — | — |
| Livraga | 098030 | 26814 | — | — |
| Locate di Triulzi | 015125 | 20085 | — | — |
| Locate Varesino | 013131 | 22070 | — | — |
| Locatello | 016127 | 24030 | — | — |
| Lodi | 098031 | 26900 | — | — |
| Lodi Vecchio | 098032 | 26855 | — | — |
| Lodrino | 017090 | 25060 | — | — |
| Lograto | 017091 | 25030 | — | — |
| Lomagna | 097044 | 23871 | — | — |
| Lomazzo | 013133 | 22074 | — | — |
| Lomello | 018083 | 27034 | — | — |
| Lonate Ceppino | 012089 | 21050 | — | — |
| Lonate Pozzolo | 012090 | 21015 | — | — |
| Lonato del Garda | 017092 | 25017 | — | — |
| Longhena | 017093 | 25030 | — | — |
| Longone al Segrino | 013134 | 22030 | — | — |
| Losine | 017094 | 25050 | — | — |
| Lovere | 016128 | 24065 | — | — |
| Lovero | 014038 | 23030 | — | — |
| Lozio | 017095 | 25040 | — | — |
| Lozza | 012091 | 21040 | — | — |
| Luino | 012092 | 21016 | — | — |
| Luisago | 013135 | 22070 | — | — |
| Lumezzane | 017096 | 25065 | — | — |
| Lungavilla | 018084 | 27053 | — | — |
| Lurago d'Erba | 013136 | 22040 | — | — |
| Lurago Marinone | 013137 | 22070 | — | — |
| Lurano | 016129 | 24050 | — | — |
| Lurate Caccivio | 013138 | 22075 | — | — |
| Luvinate | 012093 | 21020 | — | — |
| Luzzana | 016130 | 24069 | — | — |
| Maccagno con Pino e Veddasca | 012142 | 21061 | — | — |
| Maccastorna | 098033 | 26843 | — | — |
| Macherio | 108029 | 20846 | — | — |
| Maclodio | 017097 | 25030 | — | — |
| Madesimo | 014035 | 23024 | — | — |
| Madignano | 019055 | 26020 | — | — |
| Madone | 016131 | 24040 | — | — |
| Magasa | 017098 | 25080 | — | — |
| Magenta | 015130 | 20013 | — | — |
| Magherno | 018085 | 27010 | — | — |
| Magnacavallo | 020029 | 46020 | — | — |
| Magnago | 015131 | 20020 | — | — |
| Magreglio | 013139 | 22030 | — | — |
| Mairago | 098034 | 26825 | — | — |
| Mairano | 017099 | 25030 | — | — |
| Malagnino | 019056 | 26030 | — | — |
| Malegno | 017100 | 25053 | — | — |
| Maleo | 098035 | 26847 | — | — |
| Malgrate | 097045 | 23864 | — | — |
| Malnate | 012096 | 21046 | — | — |
| Malonno | 017101 | 25040 | — | — |
| Mandello del Lario | 097046 | 23826 | — | — |
| Manerba del Garda | 017102 | 25080 | — | — |
| Manerbio | 017103 | 25025 | — | — |
| Mantello | 014039 | 23016 | — | — |
| Mantova | 020030 | 46100 | — | — |
| Mapello | 016132 | 24030 | — | — |
| Marcallo con Casone | 015134 | 20010 | — | — |
| Marcaria | 020031 | 46010 | — | — |
| Marcheno | 017104 | 25060 | — | — |
| Marchirolo | 012097 | 21030 | — | — |
| Marcignago | 018086 | 27020 | — | — |
| Margno | 097047 | 23832 | — | — |
| Mariana Mantovana | 020032 | 46010 | — | — |
| Mariano Comense | 013143 | 22066 | — | — |
| Marmentino | 017105 | 25060 | — | — |
| Marmirolo | 020033 | 46045 | — | — |
| Marnate | 012098 | 21050 | — | — |
| Marone | 017106 | 25054 | — | — |
| Martignana di Po | 019057 | 26040 | — | — |
| Martinengo | 016133 | 24057 | — | — |
| Marudo | 098036 | 26866 | — | — |
| Marzano | 018087 | 27010 | — | — |
| Marzio | 012099 | 21030 | — | — |
| Masate | 015136 | 20060 | — | — |
| Masciago Primo | 012100 | 21030 | — | — |
| Maslianico | 013144 | 22026 | — | — |
| Massalengo | 098037 | 26815 | — | — |
| Mazzano | 017107 | 25080 | — | — |
| Mazzo di Valtellina | 014040 | 23030 | — | — |
| Meda | 108030 | 20821 | — | — |
| Mede | 018088 | 27035 | — | — |
| Mediglia | 015139 | 20076 | — | — |
| Medolago | 016250 | 24030 | — | — |
| Medole | 020034 | 46046 | — | — |
| Melegnano | 015140 | 20077 | — | — |
| Meleti | 098038 | 26843 | — | — |
| Mello | 014041 | 23010 | — | — |
| Melzo | 015142 | 20066 | — | — |
| Menaggio | 013145 | 22017 | — | — |
| Menconico | 018089 | 27050 | — | — |
| Merate | 097048 | 23807 | — | — |
| Mercallo | 012101 | 21020 | — | — |
| Merlino | 098039 | 26833 | — | — |
| Merone | 013147 | 22046 | — | — |
| Mese | 014043 | 23020 | — | — |
| Mesenzana | 012102 | 21030 | — | — |
| Mesero | 015144 | 20010 | — | — |
| Mezzago | 108031 | 20883 | — | — |
| Mezzana Bigli | 018090 | 27030 | — | — |
| Mezzana Rabattone | 018091 | 27030 | — | — |
| Mezzanino | 018092 | 27040 | — | — |
| Mezzoldo | 016134 | 24010 | — | — |
| Milano | 015146 | 20121, 20122, 20123, 20124, 20125, 20126, 20127, 20128, 20129, 20131, 20132, 20133, 20134, 20135, 20136, 20137, 20138, 20139, 20141, 20142, 20143, 20144, 20145, 20146, 20147, 20148, 20149, 20151, 20152, 20153, 20154, 20155, 20156, 20157, 20158, 20159, 20161, 20162 | — | — |
| Milzano | 017108 | 25020 | — | — |
| Miradolo Terme | 018093 | 27010 | — | — |
| Misano di Gera d'Adda | 016135 | 24040 | — | — |
| Misinto | 108032 | 20826 | — | — |
| Missaglia | 097049 | 23873 | — | — |
| Moggio | 097050 | 23817 | — | — |
| Moglia | 020035 | 46024 | — | — |
| Moio de' Calvi | 016136 | 24010 | — | — |
| Molteno | 097051 | 23847 | — | — |
| Moltrasio | 013152 | 22010 | — | — |
| Monasterolo del Castello | 016137 | 24060 | — | — |
| Monguzzo | 013153 | 22040 | — | — |
| Moniga del Garda | 017109 | 25080 | — | — |
| Monno | 017110 | 25040 | — | — |
| Montagna in Valtellina | 014044 | 23020 | — | — |
| Montalto Pavese | 018094 | 27040 | — | — |
| Montanaso Lombardo | 098040 | 26836 | — | — |
| Montano Lucino | 013154 | 22070 | — | — |
| Monte Cremasco | 019058 | 26010 | — | — |
| Monte Isola | 017111 | 25050 | — | — |
| Monte Marenzo | 097052 | 23804 | — | — |
| Montebello della Battaglia | 018095 | 27054 | — | — |
| Montecalvo Versiggia | 018096 | 27047 | — | — |
| Montegrino Valtravaglia | 012103 | 21010 | — | — |
| Montello | 016139 | 24060 | — | — |
| Montemezzo | 013155 | 22010 | — | — |
| Montescano | 018097 | 27040 | — | — |
| Montesegale | 018098 | 27052 | — | — |
| Montevecchia | 097053 | 23874 | — | — |
| Monticelli Brusati | 017112 | 25040 | — | — |
| Monticelli Pavese | 018099 | 27010 | — | — |
| Monticello Brianza | 097054 | 23876 | — | — |
| Montichiari | 017113 | 25018 | — | — |
| Montirone | 017114 | 25010 | — | — |
| Montodine | 019059 | 26010 | — | — |
| Montorfano | 013157 | 22030 | — | — |
| Montù Beccaria | 018100 | 27040 | — | — |
| Monvalle | 012104 | 21020 | — | — |
| Monza | 108033 | 20900 | — | — |
| Monzambano | 020036 | 46040 | — | — |
| Morazzone | 012105 | 21040 | — | — |
| Morbegno | 014045 | 23017 | — | — |
| Morengo | 016140 | 24050 | — | — |
| Morimondo | 015150 | 20081 | — | — |
| Mornago | 012106 | 21020 | — | — |
| Mornico al Serio | 016141 | 24050 | — | — |
| Mornico Losana | 018101 | 27040 | — | — |
| Mortara | 018102 | 27036 | — | — |
| Morterone | 097055 | 23811 | — | — |
| Moscazzano | 019060 | 26010 | — | — |
| Motta Baluffi | 019061 | 26045 | — | — |
| Motta Visconti | 015151 | 20086 | — | — |
| Motteggiana | 020037 | 46020 | — | — |
| Mozzanica | 016142 | 24050 | — | — |
| Mozzate | 013159 | 22076 | — | — |
| Mozzo | 016143 | 24030 | — | — |
| Muggiò | 108034 | 20835 | — | — |
| Mulazzano | 098041 | 26837 | — | — |
| Mura | 017115 | 25070 | — | — |
| Muscoline | 017116 | 25080 | — | — |
| Musso | 013160 | 22010 | — | — |
| Nave | 017117 | 25075 | — | — |
| Nembro | 016144 | 24027 | — | — |
| Nerviano | 015154 | 20014 | — | — |
| Nesso | 013161 | 22020 | — | — |
| Niardo | 017118 | 25050 | — | — |
| Nibionno | 097056 | 23895 | — | — |
| Nicorvo | 018103 | 27020 | — | — |
| Nosate | 015155 | 20020 | — | — |
| Nova Milanese | 108035 | 20834 | — | — |
| Novate Mezzola | 014046 | 23025 | — | — |
| Novate Milanese | 015157 | 20026 | — | — |
| Novedrate | 013163 | 22060 | — | — |
| Noviglio | 015158 | 20082 | — | — |
| Nuvolento | 017119 | 25080 | — | — |
| Nuvolera | 017120 | 25080 | — | — |
| Odolo | 017121 | 25076 | — | — |
| Offanengo | 019062 | 26010 | — | — |
| Offlaga | 017122 | 25020 | — | — |
| Oggiona con Santo Stefano | 012107 | 21040 | — | — |
| Oggiono | 097057 | 23848 | — | — |
| Olevano di Lomellina | 018104 | 27020 | — | — |
| Olgiate Comasco | 013165 | 22077 | — | — |
| Olgiate Molgora | 097058 | 23887 | — | — |
| Olgiate Olona | 012108 | 21057 | — | — |
| Olginate | 097059 | 23854 | — | — |
| Oliva Gessi | 018105 | 27050 | — | — |
| Oliveto Lario | 097060 | 23865 | — | — |
| Olmeneta | 019063 | 26010 | — | — |
| Olmo al Brembo | 016145 | 24010 | — | — |
| Oltre il Colle | 016146 | 24013 | — | — |
| Oltressenda Alta | 016147 | 24020 | — | — |
| Oltrona di San Mamette | 013169 | 22070 | — | — |
| Ome | 017123 | 25050 | — | — |
| Oneta | 016148 | 24020 | — | — |
| Ono San Pietro | 017124 | 25040 | — | — |
| Onore | 016149 | 24020 | — | — |
| Opera | 015159 | 20073 | — | — |
| Origgio | 012109 | 21040 | — | — |
| Orino | 012110 | 21030 | — | — |
| Orio al Serio | 016150 | 24050 | — | — |
| Orio Litta | 098042 | 26863 | — | — |
| Ornago | 108036 | 20876 | — | — |
| Ornica | 016151 | 24010 | — | — |
| Orsenigo | 013170 | 22030 | — | — |
| Orzinuovi | 017125 | 25034 | — | — |
| Orzivecchi | 017126 | 25030 | — | — |
| Osio Sopra | 016152 | 24040 | — | — |
| Osio Sotto | 016153 | 24046 | — | — |
| Osnago | 097061 | 23875 | — | — |
| Ospedaletto Lodigiano | 098043 | 26864 | — | — |
| Ospitaletto | 017127 | 25035 | — | — |
| Ossago Lodigiano | 098044 | 26816 | — | — |
| Ossimo | 017128 | 25050 | — | — |
| Ossona | 015164 | 20002 | — | — |
| Ostiano | 019064 | 26032 | — | — |
| Ostiglia | 020038 | 46035 | — | — |
| Ottobiano | 018106 | 27030 | — | — |
| Ozzero | 015165 | 20080 | — | — |
| Padenghe sul Garda | 017129 | 25080 | — | — |
| Paderno d'Adda | 097062 | 23877 | — | — |
| Paderno Dugnano | 015166 | 20037 | — | — |
| Paderno Franciacorta | 017130 | 25050 | — | — |
| Paderno Ponchielli | 019065 | 26024 | — | — |
| Pagazzano | 016154 | 24040 | — | — |
| Pagnona | 097063 | 23833 | — | — |
| Paisco Loveno | 017131 | 25050 | — | — |
| Paitone | 017132 | 25080 | — | — |
| Paladina | 016155 | 24030 | — | — |
| Palazzago | 016156 | 24030 | — | — |
| Palazzo Pignano | 019066 | 26020 | — | — |
| Palazzolo sull'Oglio | 017133 | 25036 | — | — |
| Palestro | 018107 | 27030 | — | — |
| Palosco | 016157 | 24050 | — | — |
| Pancarana | 018108 | 27050 | — | — |
| Pandino | 019067 | 26025 | — | — |
| Pantigliate | 015167 | 20048 | — | — |
| Parabiago | 015168 | 20015 | — | — |
| Paratico | 017134 | 25030 | — | — |
| Parlasco | 097064 | 23837 | — | — |
| Parona | 018109 | 27020 | — | — |
| Parre | 016158 | 24020 | — | — |
| Parzanica | 016159 | 24060 | — | — |
| Paspardo | 017135 | 25050 | — | — |
| Passirano | 017136 | 25050 | — | — |
| Pasturo | 097065 | 23818 | — | — |
| Paullo | 015169 | 20067 | — | — |
| Pavia | 018110 | 27100 | — | — |
| Pavone del Mella | 017137 | 25020 | — | — |
| Pedesina | 014047 | 23010 | — | — |
| Pedrengo | 016160 | 24066 | — | — |
| Peglio | 013178 | 22010 | — | — |
| Pegognaga | 020039 | 46020 | — | — |
| Peia | 016161 | 24020 | — | — |
| Perledo | 097067 | 23828 | — | — |
| Pero | 015170 | 20016 | — | — |
| Persico Dosimo | 019068 | 26043 | — | — |
| Pertica Alta | 017139 | 25070 | — | — |
| Pertica Bassa | 017140 | 25078 | — | — |
| Pescarolo ed Uniti | 019069 | 26033 | — | — |
| Pescate | 097068 | 23855 | — | — |
| Peschiera Borromeo | 015171 | 20068 | — | — |
| Pessano con Bornago | 015172 | 20042 | — | — |
| Pessina Cremonese | 019070 | 26030 | — | — |
| Pezzaze | 017141 | 25060 | — | — |
| Piadena Drizzona | 019116 | 26034 | — | — |
| Pian Camuno | 017142 | 25050 | — | — |
| Piancogno | 017206 | 25052 | — | — |
| Pianello del Lario | 013183 | 22010 | — | — |
| Pianengo | 019072 | 26010 | — | — |
| Pianico | 016162 | 24060 | — | — |
| Piantedo | 014048 | 23010 | — | — |
| Piario | 016163 | 24020 | — | — |
| Piateda | 014049 | 23020 | — | — |
| Piazza Brembana | 016164 | 24014 | — | — |
| Piazzatorre | 016165 | 24010 | — | — |
| Piazzolo | 016166 | 24010 | — | — |
| Pieranica | 019073 | 26017 | — | — |
| Pietra de' Giorgi | 018111 | 27040 | — | — |
| Pieve Albignola | 018112 | 27030 | — | — |
| Pieve d'Olmi | 019074 | 26040 | — | — |
| Pieve del Cairo | 018113 | 27037 | — | — |
| Pieve Emanuele | 015173 | 20072 | — | — |
| Pieve Fissiraga | 098045 | 26854 | — | — |
| Pieve Porto Morone | 018114 | 27017 | — | — |
| Pieve San Giacomo | 019075 | 26035 | — | — |
| Pigra | 013184 | 22020 | — | — |
| Pinarolo Po | 018115 | 27040 | — | — |
| Pioltello | 015175 | 20096 | — | — |
| Pisogne | 017143 | 25055 | — | — |
| Piubega | 020041 | 46040 | — | — |
| Piuro | 014050 | 23020 | — | — |
| Pizzale | 018116 | 27050 | — | — |
| Pizzighettone | 019076 | 26026 | — | — |
| Plesio | 013185 | 22010 | — | — |
| Poggio Rusco | 020042 | 46025 | — | — |
| Poggiridenti | 014051 | 23020 | — | — |
| Pogliano Milanese | 015176 | 20005 | — | — |
| Pognana Lario | 013186 | 22020 | — | — |
| Pognano | 016167 | 24040 | — | — |
| Polaveno | 017144 | 25060 | — | — |
| Polpenazze del Garda | 017145 | 25080 | — | — |
| Pompiano | 017146 | 25030 | — | — |
| Pomponesco | 020043 | 46030 | — | — |
| Poncarale | 017147 | 25020 | — | — |
| Ponna | 013187 | 22020 | — | — |
| Ponte di Legno | 017148 | 25056 | — | — |
| Ponte in Valtellina | 014052 | 23026 | — | — |
| Ponte Lambro | 013188 | 22037 | — | — |
| Ponte Nizza | 018117 | 27050 | — | — |
| Ponte Nossa | 016168 | 24028 | — | — |
| Ponte San Pietro | 016170 | 24036 | — | — |
| Ponteranica | 016169 | 24010 | — | — |
| Pontevico | 017149 | 25026 | — | — |
| Ponti sul Mincio | 020044 | 46040 | — | — |
| Pontida | 016171 | 24030 | — | — |
| Pontirolo Nuovo | 016172 | 24040 | — | — |
| Pontoglio | 017150 | 25037 | — | — |
| Porlezza | 013189 | 22018 | — | — |
| Portalbera | 018118 | 27040 | — | — |
| Porto Ceresio | 012113 | 21050 | — | — |
| Porto Mantovano | 020045 | 46047 | — | — |
| Porto Valtravaglia | 012114 | 21010 | — | — |
| Postalesio | 014053 | 23010 | — | — |
| Pozzaglio ed Uniti | 019077 | 26010 | — | — |
| Pozzo d'Adda | 015177 | 20060 | — | — |
| Pozzolengo | 017151 | 25010 | — | — |
| Pozzuolo Martesana | 015178 | 20060 | — | — |
| Pradalunga | 016173 | 24020 | — | — |
| Pralboino | 017152 | 25020 | — | — |
| Prata Camportaccio | 014054 | 23020 | — | — |
| Predore | 016174 | 24060 | — | — |
| Pregnana Milanese | 015179 | 20006 | — | — |
| Premana | 097069 | 23834 | — | — |
| Premolo | 016175 | 24020 | — | — |
| Preseglie | 017153 | 25070 | — | — |
| Presezzo | 016176 | 24030 | — | — |
| Prevalle | 017155 | 25080 | — | — |
| Primaluna | 097070 | 23819 | — | — |
| Proserpio | 013192 | 22030 | — | — |
| Provaglio d'Iseo | 017156 | 25050 | — | — |
| Provaglio Val Sabbia | 017157 | 25070 | — | — |
| Puegnago del Garda | 017158 | 25080 | — | — |
| Pumenengo | 016177 | 24050 | — | — |
| Pusiano | 013193 | 22030 | — | — |
| Quingentole | 020046 | 46020 | — | — |
| Quintano | 019078 | 26017 | — | — |
| Quinzano d'Oglio | 017159 | 25027 | — | — |
| Quistello | 020047 | 46026 | — | — |
| Rancio Valcuvia | 012115 | 21030 | — | — |
| Ranco | 012116 | 21020 | — | — |
| Ranica | 016178 | 24020 | — | — |
| Ranzanico | 016179 | 24060 | — | — |
| Rasura | 014055 | 23010 | — | — |
| Rea | 018119 | 27040 | — | — |
| Redavalle | 018120 | 27050 | — | — |
| Redondesco | 020048 | 46010 | — | — |
| Remedello | 017160 | 25010 | — | — |
| Renate | 108037 | 20838 | — | — |
| Rescaldina | 015181 | 20027 | — | — |
| Retorbido | 018121 | 27050 | — | — |
| Rezzago | 013195 | 22030 | — | — |
| Rezzato | 017161 | 25086 | — | — |
| Rho | 015182 | 20017 | — | — |
| Ricengo | 019079 | 26010 | — | — |
| Ripalta Arpina | 019080 | 26010 | — | — |
| Ripalta Cremasca | 019081 | 26010 | — | — |
| Ripalta Guerina | 019082 | 26010 | — | — |
| Riva di Solto | 016180 | 24060 | — | — |
| Rivanazzano Terme | 018122 | 27055 | — | — |
| Rivarolo del Re ed Uniti | 019083 | 26036 | — | — |
| Rivarolo Mantovano | 020050 | 46017 | — | — |
| Rivolta d'Adda | 019084 | 26027 | — | — |
| Robbiate | 097071 | 23899 | — | — |
| Robbio | 018123 | 27038 | — | — |
| Robecchetto con Induno | 015183 | 20020 | — | — |
| Robecco d'Oglio | 019085 | 26010 | — | — |
| Robecco Pavese | 018124 | 27042 | — | — |
| Robecco sul Naviglio | 015184 | 20087 | — | — |
| Rocca de' Giorgi | 018125 | 27040 | — | — |
| Rocca Susella | 018126 | 27052 | — | — |
| Roccafranca | 017162 | 25030 | — | — |
| Rodano | 015185 | 20053 | — | — |
| Rodengo Saiano | 017163 | 25050 | — | — |
| Rodero | 013197 | 22070 | — | — |
| Rodigo | 020051 | 46040 | — | — |
| Roè Volciano | 017164 | 25077 | — | — |
| Rogeno | 097072 | 23849 | — | — |
| Rognano | 018127 | 27010 | — | — |
| Rogno | 016182 | 24060 | — | — |
| Rogolo | 014056 | 23010 | — | — |
| Romagnese | 018128 | 27050 | — | — |
| Romanengo | 019086 | 26014 | — | — |
| Romano di Lombardia | 016183 | 24058 | — | — |
| Roncadelle | 017165 | 25030 | — | — |
| Roncaro | 018129 | 27010 | — | — |
| Roncello | 108055 | 20877 | — | — |
| Ronco Briantino | 108038 | 20885 | — | — |
| Roncobello | 016184 | 24010 | — | — |
| Roncoferraro | 020052 | 46037 | — | — |
| Roncola | 016185 | 24030 | — | — |
| Rosasco | 018130 | 27030 | — | — |
| Rosate | 015188 | 20088 | — | — |
| Rota d'Imagna | 016186 | 24037 | — | — |
| Rovato | 017166 | 25038 | — | — |
| Rovellasca | 013201 | 22069 | — | — |
| Rovello Porro | 013202 | 22070 | — | — |
| Roverbella | 020053 | 46048 | — | — |
| Rovescala | 018131 | 27040 | — | — |
| Rovetta | 016187 | 24020 | — | — |
| Rozzano | 015189 | 20089 | — | — |
| Rudiano | 017167 | 25030 | — | — |
| Sabbio Chiese | 017168 | 25070 | — | — |
| Sabbioneta | 020054 | 46018 | — | — |
| Sala Comacina | 013203 | 22010 | — | — |
| Sale Marasino | 017169 | 25057 | — | — |
| Salerano sul Lambro | 098046 | 26857 | — | — |
| Salò | 017170 | 25087 | — | — |
| Saltrio | 012117 | 21050 | — | — |
| Salvirola | 019087 | 26010 | — | — |
| Samarate | 012118 | 21017 | — | — |
| Samolaco | 014057 | 23027 | — | — |
| San Bartolomeo Val Cavargna | 013204 | 22010 | — | — |
| San Bassano | 019088 | 26020 | — | — |
| San Benedetto Po | 020055 | 46027 | — | — |
| San Cipriano Po | 018133 | 27043 | — | — |
| San Colombano al Lambro | 015191 | 20078 | — | — |
| San Damiano al Colle | 018134 | 27040 | — | — |
| San Daniele Po | 019089 | 26046 | — | — |
| San Donato Milanese | 015192 | 20097 | — | — |
| San Felice del Benaco | 017171 | 25010 | — | — |
| San Fermo della Battaglia | 013206 | 22042 | — | — |
| San Fiorano | 098047 | 26848 | — | — |
| San Genesio ed Uniti | 018135 | 27010 | — | — |
| San Gervasio Bresciano | 017172 | 25020 | — | — |
| San Giacomo delle Segnate | 020056 | 46020 | — | — |
| San Giacomo Filippo | 014058 | 23020 | — | — |
| San Giorgio Bigarello | 020057 | 46051 | — | — |
| San Giorgio di Lomellina | 018136 | 27020 | — | — |
| San Giorgio su Legnano | 015194 | 20034 | — | — |
| San Giovanni Bianco | 016188 | 24015 | — | — |
| San Giovanni del Dosso | 020058 | 46020 | — | — |
| San Giovanni in Croce | 019090 | 26037 | — | — |
| San Giuliano Milanese | 015195 | 20098 | — | — |
| San Martino dall'Argine | 020059 | 46010 | — | — |
| San Martino del Lago | 019091 | 26040 | — | — |
| San Martino in Strada | 098048 | 26817 | — | — |
| San Martino Siccomario | 018137 | 27028 | — | — |
| San Nazzaro Val Cavargna | 013207 | 22010 | — | — |
| San Paolo | 017138 | 25020 | — | — |
| San Paolo d'Argon | 016189 | 24060 | — | — |
| San Pellegrino Terme | 016190 | 24016 | — | — |
| San Rocco al Porto | 098049 | 26865 | — | — |
| San Siro | 013248 | 22010 | — | — |
| San Vittore Olona | 015201 | 20028 | — | — |
| San Zeno Naviglio | 017173 | 25010 | — | — |
| San Zenone al Lambro | 015202 | 20070 | — | — |
| San Zenone al Po | 018145 | 27010 | — | — |
| Sangiano | 012141 | 21038 | — | — |
| Sannazzaro de' Burgondi | 018138 | 27039 | — | — |
| Sant'Alessio con Vialone | 018141 | 27016 | — | — |
| Sant'Angelo Lodigiano | 098050 | 26866 | — | — |
| Sant'Angelo Lomellina | 018144 | 27030 | — | — |
| Sant'Omobono Terme | 016252 | 24038 | — | — |
| Santa Brigida | 016191 | 24010 | — | — |
| Santa Cristina e Bissone | 018139 | 27010 | — | — |
| Santa Giuletta | 018140 | 27046 | — | — |
| Santa Margherita di Staffora | 018142 | 27050 | — | — |
| Santa Maria della Versa | 018143 | 27047 | — | — |
| Santa Maria Hoè | 097074 | 23889 | — | — |
| Santo Stefano Lodigiano | 098051 | 26849 | — | — |
| Santo Stefano Ticino | 015200 | 20010 | — | — |
| Sarezzo | 017174 | 25068 | — | — |
| Sarnico | 016193 | 24067 | — | — |
| Saronno | 012119 | 21047 | — | — |
| Sartirana Lomellina | 018146 | 27020 | — | — |
| Saviore dell'Adamello | 017175 | 25040 | — | — |
| Scaldasole | 018147 | 27020 | — | — |
| Scandolara Ravara | 019092 | 26040 | — | — |
| Scandolara Ripa d'Oglio | 019093 | 26047 | — | — |
| Scanzorosciate | 016194 | 24020 | — | — |
| Schignano | 013211 | 22020 | — | — |
| Schilpario | 016195 | 24020 | — | — |
| Schivenoglia | 020060 | 46020 | — | — |
| Secugnago | 098052 | 26826 | — | — |
| Sedriano | 015204 | 20018 | — | — |
| Sedrina | 016196 | 24010 | — | — |
| Segrate | 015205 | 20054 | — | — |
| Sellero | 017176 | 25050 | — | — |
| Selvino | 016197 | 24020 | — | — |
| Semiana | 018148 | 27020 | — | — |
| Senago | 015206 | 20030 | — | — |
| Seniga | 017177 | 25020 | — | — |
| Senna Comasco | 013212 | 22070 | — | — |
| Senna Lodigiana | 098053 | 26856 | — | — |
| Seregno | 108039 | 20831 | — | — |
| Sergnano | 019094 | 26010 | — | — |
| Seriate | 016198 | 24068 | — | — |
| Serina | 016199 | 24017 | — | — |
| Serle | 017178 | 25080 | — | — |
| Sermide e Felonica | 020061 | 46028 | — | — |
| Sernio | 014059 | 23030 | — | — |
| Serravalle a Po | 020062 | 46030 | — | — |
| Sesto Calende | 012120 | 21018 | — | — |
| Sesto ed Uniti | 019095 | 26028 | — | — |
| Sesto San Giovanni | 015209 | 20099 | — | — |
| Settala | 015210 | 20049 | — | — |
| Settimo Milanese | 015211 | 20019 | — | — |
| Seveso | 108040 | 20822 | — | — |
| Silvano Pietra | 018149 | 27050 | — | — |
| Sirmione | 017179 | 25019 | — | — |
| Sirone | 097075 | 23844 | — | — |
| Sirtori | 097076 | 23896 | — | — |
| Siziano | 018150 | 27010 | — | — |
| Soiano del Lago | 017180 | 25080 | — | — |
| Solaro | 015213 | 20033 | — | — |
| Solarolo Rainerio | 019096 | 26030 | — | — |
| Solbiate Arno | 012121 | 21048 | — | — |
| Solbiate con Cagno | 013255 | 22043 | — | — |
| Solbiate Olona | 012122 | 21058 | — | — |
| Solferino | 020063 | 46040 | — | — |
| Solto Collina | 016200 | 24060 | — | — |
| Solza | 016251 | 24030 | — | — |
| Somaglia | 098054 | 26867 | — | — |
| Somma Lombardo | 012123 | 21019 | — | — |
| Sommo | 018151 | 27048 | — | — |
| Soncino | 019097 | 26029 | — | — |
| Sondalo | 014060 | 23035 | — | — |
| Sondrio | 014061 | 23100 | — | — |
| Songavazzo | 016201 | 24020 | — | — |
| Sonico | 017181 | 25048 | — | — |
| Sordio | 098055 | 26858 | — | — |
| Soresina | 019098 | 26015 | — | — |
| Sorico | 013216 | 22010 | — | — |
| Sorisole | 016202 | 24010 | — | — |
| Sormano | 013217 | 22030 | — | — |
| Sospiro | 019099 | 26048 | — | — |
| Sotto il Monte Giovanni XXIII | 016203 | 24039 | — | — |
| Sovere | 016204 | 24060 | — | — |
| Sovico | 108041 | 20845 | — | — |
| Spessa | 018152 | 27010 | — | — |
| Spinadesco | 019100 | 26020 | — | — |
| Spineda | 019101 | 26030 | — | — |
| Spino d'Adda | 019102 | 26016 | — | — |
| Spinone al Lago | 016205 | 24060 | — | — |
| Spirano | 016206 | 24050 | — | — |
| Spriana | 014062 | 23020 | — | — |
| Stagno Lombardo | 019103 | 26049 | — | — |
| Stazzona | 013218 | 22010 | — | — |
| Stezzano | 016207 | 24040 | — | — |
| Stradella | 018153 | 27049 | — | — |
| Strozza | 016208 | 24030 | — | — |
| Suardi | 018154 | 27030 | — | — |
| Sueglio | 097077 | 23835 | — | — |
| Suello | 097078 | 23867 | — | — |
| Suisio | 016209 | 24040 | — | — |
| Sulbiate | 108042 | 20884 | — | — |
| Sulzano | 017182 | 25058 | — | — |
| Sumirago | 012124 | 21040 | — | — |
| Sustinente | 020064 | 46030 | — | — |
| Suzzara | 020065 | 46029 | — | — |
| Taceno | 097079 | 23837 | — | — |
| Taino | 012125 | 21020 | — | — |
| Talamona | 014063 | 23018 | — | — |
| Taleggio | 016210 | 24010 | — | — |
| Tartano | 014064 | 23010 | — | — |
| Tavazzano con Villavesco | 098056 | 26838 | — | — |
| Tavernerio | 013222 | 22038 | — | — |
| Tavernola Bergamasca | 016211 | 24060 | — | — |
| Tavernole sul Mella | 017183 | 25060 | — | — |
| Teglio | 014065 | 23036 | — | — |
| Telgate | 016212 | 24060 | — | — |
| Temù | 017184 | 25050 | — | — |
| Ternate | 012126 | 21020 | — | — |
| Terno d'Isola | 016213 | 24030 | — | — |
| Terranova dei Passerini | 098057 | 26827 | — | — |
| Ticengo | 019104 | 26020 | — | — |
| Tignale | 017185 | 25080 | — | — |
| Tirano | 014066 | 23037 | — | — |
| Torbole Casaglia | 017186 | 25030 | — | — |
| Torlino Vimercati | 019105 | 26017 | — | — |
| Tornata | 019106 | 26030 | — | — |
| Torno | 013223 | 22020 | — | — |
| Torrazza Coste | 018155 | 27050 | — | — |
| Torre Beretti e Castellaro | 018156 | 27030 | — | — |
| Torre Boldone | 016214 | 24020 | — | — |
| Torre d'Arese | 018157 | 27010 | — | — |
| Torre d'Isola | 018159 | 27020 | — | — |
| Torre de' Busi | 016215 | 24032 | — | — |
| Torre de' Negri | 018158 | 27011 | — | — |
| Torre de' Picenardi | 019107 | 26038 | — | — |
| Torre de' Roveri | 016216 | 24060 | — | — |
| Torre di Santa Maria | 014067 | 23020 | — | — |
| Torre Pallavicina | 016217 | 24050 | — | — |
| Torrevecchia Pia | 018160 | 27010 | — | — |
| Torricella del Pizzo | 019108 | 26040 | — | — |
| Torricella Verzate | 018161 | 27050 | — | — |
| Toscolano-Maderno | 017187 | 25088 | — | — |
| Tovo di Sant'Agata | 014068 | 23030 | — | — |
| Tradate | 012127 | 21049 | — | — |
| Traona | 014069 | 23019 | — | — |
| Travacò Siccomario | 018162 | 27020 | — | — |
| Travagliato | 017188 | 25039 | — | — |
| Travedona-Monate | 012128 | 21028 | — | — |
| Tremezzina | 013252 | 22016 | — | — |
| Tremosine sul Garda | 017189 | 25010 | — | — |
| Trenzano | 017190 | 25030 | — | — |
| Trescore Balneario | 016218 | 24069 | — | — |
| Trescore Cremasco | 019109 | 26017 | — | — |
| Tresivio | 014070 | 23020 | — | — |
| Treviglio | 016219 | 24047 | — | — |
| Treviolo | 016220 | 24048 | — | — |
| Treviso Bresciano | 017191 | 25070 | — | — |
| Trezzano Rosa | 015219 | 20060 | — | — |
| Trezzano sul Naviglio | 015220 | 20090 | — | — |
| Trezzo sull'Adda | 015221 | 20056 | — | — |
| Trezzone | 013226 | 22010 | — | — |
| Tribiano | 015222 | 20067 | — | — |
| Trigolo | 019110 | 26018 | — | — |
| Triuggio | 108043 | 20844 | — | — |
| Trivolzio | 018163 | 27020 | — | — |
| Tromello | 018164 | 27020 | — | — |
| Tronzano Lago Maggiore | 012129 | 21010 | — | — |
| Trovo | 018165 | 27020 | — | — |
| Truccazzano | 015224 | 20060 | — | — |
| Turano Lodigiano | 098058 | 26828 | — | — |
| Turate | 013227 | 22078 | — | — |
| Turbigo | 015226 | 20029 | — | — |
| Ubiale Clanezzo | 016221 | 24010 | — | — |
| Uboldo | 012130 | 21040 | — | — |
| Uggiate con Ronago | 013256 | 22029 | — | — |
| Urago d'Oglio | 017192 | 25030 | — | — |
| Urgnano | 016222 | 24059 | — | — |
| Usmate Velate | 108044 | 20865 | — | — |
| Vaiano Cremasco | 019111 | 26010 | — | — |
| Vailate | 019112 | 26019 | — | — |
| Val Brembilla | 016253 | 24012 | — | — |
| Val di Nizza | 018166 | 27050 | — | — |
| Val Masino | 014074 | 23010 | — | — |
| Val Rezzo | 013233 | 22010 | — | — |
| Valbondione | 016223 | 24020 | — | — |
| Valbrembo | 016224 | 24030 | — | — |
| Valbrona | 013229 | 22039 | — | — |
| Valdidentro | 014071 | 23038 | — | — |
| Valdisotto | 014072 | 23030 | — | — |
| Valeggio | 018167 | 27020 | — | — |
| Valera Fratta | 098059 | 26859 | — | — |
| Valfurva | 014073 | 23030 | — | — |
| Valganna | 012131 | 21039 | — | — |
| Valgoglio | 016225 | 24020 | — | — |
| Valgreghentino | 097082 | 23857 | — | — |
| Valle Lomellina | 018168 | 27020 | — | — |
| Valle Salimbene | 018169 | 27010 | — | — |
| Valleve | 016226 | 24010 | — | — |
| Vallio Terme | 017193 | 25080 | — | — |
| Valmadrera | 097083 | 23868 | — | — |
| Valmorea | 013232 | 22070 | — | — |
| Valnegra | 016227 | 24010 | — | — |
| Valsolda | 013234 | 22010 | — | — |
| Valtorta | 016229 | 24010 | — | — |
| Valvarrone | 097093 | 23836 | — | — |
| Valvestino | 017194 | 25080 | — | — |
| Vanzaghello | 015249 | 20020 | — | — |
| Vanzago | 015229 | 20043 | — | — |
| Vaprio d'Adda | 015230 | 20069 | — | — |
| Varano Borghi | 012132 | 21020 | — | — |
| Varedo | 108045 | 20814 | — | — |
| Varenna | 097084 | 23829 | — | — |
| Varese | 012133 | 21100 | — | — |
| Varzi | 018171 | 27057 | — | — |
| Vedano al Lambro | 108046 | 20854 | — | — |
| Vedano Olona | 012134 | 21040 | — | — |
| Vedeseta | 016230 | 24010 | — | — |
| Veduggio con Colzano | 108047 | 20837 | — | — |
| Veleso | 013236 | 22020 | — | — |
| Velezzo Lomellina | 018172 | 27020 | — | — |
| Vellezzo Bellini | 018173 | 27010 | — | — |
| Venegono Inferiore | 012136 | 21040 | — | — |
| Venegono Superiore | 012137 | 21040 | — | — |
| Veniano | 013238 | 22070 | — | — |
| Verano Brianza | 108048 | 20843 | — | — |
| Vercana | 013239 | 22013 | — | — |
| Verceia | 014075 | 23020 | — | — |
| Vercurago | 097086 | 23808 | — | — |
| Verdellino | 016232 | 24040 | — | — |
| Verdello | 016233 | 24049 | — | — |
| Verderio | 097091 | 23879 | — | — |
| Vergiate | 012138 | 21029 | — | — |
| Vermezzo con Zelo | 015251 | 20071 | — | — |
| Vernate | 015236 | 20080 | — | — |
| Verolanuova | 017195 | 25028 | — | — |
| Verolavecchia | 017196 | 25029 | — | — |
| Verretto | 018174 | 27053 | — | — |
| Verrua Po | 018175 | 27040 | — | — |
| Vertemate con Minoprio | 013242 | 22070 | — | — |
| Vertova | 016234 | 24029 | — | — |
| Vervio | 014076 | 23030 | — | — |
| Vescovato | 019113 | 26039 | — | — |
| Vestone | 017197 | 25078 | — | — |
| Vezza d'Oglio | 017198 | 25059 | — | — |
| Viadana | 020066 | 46019 | — | — |
| Viadanica | 016235 | 24060 | — | — |
| Vidigulfo | 018176 | 27018 | — | — |
| Viganò | 097090 | 23897 | — | — |
| Vigano San Martino | 016236 | 24060 | — | — |
| Vigevano | 018177 | 27029 | — | — |
| Viggiù | 012139 | 21059 | — | — |
| Vignate | 015237 | 20052 | — | — |
| Vigolo | 016237 | 24060 | — | — |
| Villa Biscossi | 018178 | 27035 | — | — |
| Villa Carcina | 017199 | 25069 | — | — |
| Villa Cortese | 015248 | 20035 | — | — |
| Villa d'Adda | 016238 | 24030 | — | — |
| Villa d'Almè | 016239 | 24018 | — | — |
| Villa d'Ogna | 016241 | 24020 | — | — |
| Villa di Chiavenna | 014077 | 23029 | — | — |
| Villa di Serio | 016240 | 24020 | — | — |
| Villa di Tirano | 014078 | 23030 | — | — |
| Villa Guardia | 013245 | 22079 | — | — |
| Villachiara | 017200 | 25030 | — | — |
| Villanova d'Ardenghi | 018179 | 27030 | — | — |
| Villanova del Sillaro | 098060 | 26818 | — | — |
| Villanterio | 018180 | 27019 | — | — |
| Villanuova sul Clisi | 017201 | 25089 | — | — |
| Villasanta | 108049 | 20852 | — | — |
| Villimpenta | 020068 | 46039 | — | — |
| Villongo | 016242 | 24060 | — | — |
| Vilminore di Scalve | 016243 | 24020 | — | — |
| Vimercate | 108050 | 20871 | — | — |
| Vimodrone | 015242 | 20055 | — | — |
| Vione | 017202 | 25050 | — | — |
| Visano | 017203 | 25010 | — | — |
| Vistarino | 018181 | 27010 | — | — |
| Vittuone | 015243 | 20009 | — | — |
| Vizzola Ticino | 012140 | 21010 | — | — |
| Vizzolo Predabissi | 015244 | 20070 | — | — |
| Vobarno | 017204 | 25079 | — | — |
| Voghera | 018182 | 27058 | — | — |
| Volongo | 019114 | 26030 | — | — |
| Volpara | 018183 | 27047 | — | — |
| Volta Mantovana | 020070 | 46049 | — | — |
| Voltido | 019115 | 26030 | — | — |
| Zandobbio | 016244 | 24060 | — | — |
| Zanica | 016245 | 24050 | — | — |
| Zavattarello | 018184 | 27059 | — | — |
| Zeccone | 018185 | 27010 | — | — |
| Zelbio | 013246 | 22020 | — | — |
| Zelo Buon Persico | 098061 | 26839 | — | — |
| Zeme | 018186 | 27030 | — | — |
| Zenevredo | 018187 | 27049 | — | — |
| Zerbo | 018188 | 27017 | — | — |
| Zerbolò | 018189 | 27020 | — | — |
| Zibido San Giacomo | 015247 | 20058 | — | — |
| Zinasco | 018190 | 27030 | — | — |
| Zogno | 016246 | 24019 | — | — |
| Zone | 017205 | 25050 | — | — |

<a id="region-11"></a>

### Marche (225 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Marche — region** | 11 | — | Not started | Regional sources not checked |
| Acqualagna | 041001 | 61041 | — | — |
| Acquasanta Terme | 044001 | 63095 | — | — |
| Acquaviva Picena | 044002 | 63075 | — | — |
| Agugliano | 042001 | 60020 | — | — |
| Altidona | 109001 | 63824 | — | — |
| Amandola | 109002 | 63857 | — | — |
| Ancona | 042002 | 60121, 60122, 60123, 60124, 60125, 60126, 60127, 60128, 60129, 60131 | — | — |
| Apecchio | 041002 | 61042 | — | — |
| Apiro | 043002 | 62021 | — | — |
| Appignano | 043003 | 62010 | — | — |
| Appignano del Tronto | 044005 | 63083 | — | — |
| Arcevia | 042003 | 60011 | — | — |
| Arquata del Tronto | 044006 | 63096 | — | — |
| Ascoli Piceno | 044007 | 63100 | — | — |
| Barbara | 042004 | 60010 | — | — |
| Belforte all'Isauro | 041005 | 61026 | — | — |
| Belforte del Chienti | 043004 | 62020 | — | — |
| Belmonte Piceno | 109003 | 63838 | — | — |
| Belvedere Ostrense | 042005 | 60030 | — | — |
| Bolognola | 043005 | 62035 | — | — |
| Borgo Pace | 041006 | 61040 | — | — |
| Cagli | 041007 | 61043 | — | — |
| Caldarola | 043006 | 62020 | — | — |
| Camerano | 042006 | 60021 | — | — |
| Camerata Picena | 042007 | 60020 | — | — |
| Camerino | 043007 | 62032 | — | — |
| Campofilone | 109004 | 63828 | — | — |
| Camporotondo di Fiastrone | 043008 | 62020 | — | — |
| Cantiano | 041008 | 61044 | — | — |
| Carassai | 044010 | 63063 | — | — |
| Carpegna | 041009 | 61021 | — | — |
| Cartoceto | 041010 | 61030 | — | — |
| Castel di Lama | 044011 | 63082 | — | — |
| Castelbellino | 042008 | 60030 | — | — |
| Castelfidardo | 042010 | 60022 | — | — |
| Castelleone di Suasa | 042011 | 60010 | — | — |
| Castelplanio | 042012 | 60031 | — | — |
| Castelraimondo | 043009 | 62022 | — | — |
| Castelsantangelo sul Nera | 043010 | 62039 | — | — |
| Castignano | 044012 | 63072 | — | — |
| Castorano | 044013 | 63081 | — | — |
| Cerreto d'Esi | 042013 | 60043 | — | — |
| Cessapalombo | 043011 | 62020 | — | — |
| Chiaravalle | 042014 | 60033 | — | — |
| Cingoli | 043012 | 62011 | — | — |
| Civitanova Marche | 043013 | 62012 | — | — |
| Colli al Metauro | 041069 | 61036 | — | — |
| Colli del Tronto | 044014 | 63079 | — | — |
| Colmurano | 043014 | 62020 | — | — |
| Comunanza | 044015 | 63087 | — | — |
| Corinaldo | 042015 | 60013 | — | — |
| Corridonia | 043015 | 62014 | — | — |
| Cossignano | 044016 | 63067 | — | — |
| Cupra Marittima | 044017 | 63064 | — | — |
| Cupramontana | 042016 | 60034 | — | — |
| Esanatoglia | 043016 | 62024 | — | — |
| Fabriano | 042017 | 60044 | — | — |
| Falconara Marittima | 042018 | 60015 | — | — |
| Falerone | 109005 | 63837 | — | — |
| Fano | 041013 | 61032 | — | — |
| Fermignano | 041014 | 61033 | — | — |
| Fermo | 109006 | 63900 | — | — |
| Fiastra | 043017 | 62035 | — | — |
| Filottrano | 042019 | 60024 | — | — |
| Fiuminata | 043019 | 62025 | — | — |
| Folignano | 044020 | 63084 | — | — |
| Force | 044021 | 63086 | — | — |
| Fossombrone | 041015 | 61034 | — | — |
| Francavilla d'Ete | 109007 | 63816 | — | — |
| Fratte Rosa | 041016 | 61040 | — | — |
| Frontino | 041017 | 61021 | — | — |
| Frontone | 041018 | 61040 | — | — |
| Gabicce Mare | 041019 | 61011 | — | — |
| Gagliole | 043020 | 62022 | — | — |
| Genga | 042020 | 60040 | — | — |
| Gradara | 041020 | 61012 | — | — |
| Grottammare | 044023 | 63066 | — | — |
| Grottazzolina | 109008 | 63844 | — | — |
| Gualdo | 043021 | 62020 | — | — |
| Isola del Piano | 041021 | 61030 | — | — |
| Jesi | 042021 | 60035 | — | — |
| Lapedona | 109009 | 63823 | — | — |
| Loreto | 042022 | 60025 | — | — |
| Loro Piceno | 043022 | 62020 | — | — |
| Lunano | 041022 | 61026 | — | — |
| Macerata | 043023 | 62100 | — | — |
| Macerata Feltria | 041023 | 61023 | — | — |
| Magliano di Tenna | 109010 | 63832 | — | — |
| Maiolati Spontini | 042023 | 60030 | — | — |
| Maltignano | 044027 | 63085 | — | — |
| Massa Fermana | 109011 | 63834 | — | — |
| Massignano | 044029 | 63061 | — | — |
| Matelica | 043024 | 62024 | — | — |
| Mercatello sul Metauro | 041025 | 61040 | — | — |
| Mercatino Conca | 041026 | 61013 | — | — |
| Mergo | 042024 | 60030 | — | — |
| Mogliano | 043025 | 62010 | — | — |
| Mombaroccio | 041027 | 61024 | — | — |
| Mondavio | 041028 | 61040 | — | — |
| Mondolfo | 041029 | 61037 | — | — |
| Monsampietro Morico | 109012 | 63842 | — | — |
| Monsampolo del Tronto | 044031 | 63077 | — | — |
| Monsano | 042025 | 60030 | — | — |
| Montalto delle Marche | 044032 | 63068 | — | — |
| Montappone | 109013 | 63835 | — | — |
| Monte Cavallo | 043027 | 62036 | — | — |
| Monte Cerignone | 041031 | 61010 | — | — |
| Monte Giberto | 109016 | 63846 | — | — |
| Monte Grimano Terme | 041035 | 61010 | — | — |
| Monte Porzio | 041038 | 61040 | — | — |
| Monte Rinaldo | 109021 | 63852 | — | — |
| Monte Roberto | 042029 | 60030 | — | — |
| Monte San Giusto | 043031 | 62015 | — | — |
| Monte San Martino | 043032 | 62020 | — | — |
| Monte San Pietrangeli | 109023 | 63815 | — | — |
| Monte San Vito | 042030 | 60037 | — | — |
| Monte Urano | 109024 | 63813 | — | — |
| Monte Vidon Combatte | 109025 | 63847 | — | — |
| Monte Vidon Corrado | 109026 | 63836 | — | — |
| Montecalvo in Foglia | 041030 | 61020 | — | — |
| Montecarotto | 042026 | 60036 | — | — |
| Montecassiano | 043026 | 62010 | — | — |
| Montecosaro | 043028 | 62010 | — | — |
| Montedinove | 044034 | 63069 | — | — |
| Montefalcone Appennino | 109014 | 63855 | — | — |
| Montefano | 043029 | 62010 | — | — |
| Montefelcino | 041034 | 61030 | — | — |
| Montefiore dell'Aso | 044036 | 63062 | — | — |
| Montefortino | 109015 | 63858 | — | — |
| Montegallo | 044038 | 63094 | — | — |
| Montegiorgio | 109017 | 63833 | — | — |
| Montegranaro | 109018 | 63812 | — | — |
| Montelabbate | 041036 | 61025 | — | — |
| Monteleone di Fermo | 109019 | 63841 | — | — |
| Montelparo | 109020 | 63853 | — | — |
| Montelupone | 043030 | 62010 | — | — |
| Montemarciano | 042027 | 60018 | — | — |
| Montemonaco | 044044 | 63088 | — | — |
| Monteprandone | 044045 | 63076 | — | — |
| Monterubbiano | 109022 | 63825 | — | — |
| Montottone | 109027 | 63843 | — | — |
| Moresco | 109028 | 63826 | — | — |
| Morro d'Alba | 042031 | 60030 | — | — |
| Morrovalle | 043033 | 62010 | — | — |
| Muccia | 043034 | 62034 | — | — |
| Numana | 042032 | 60026 | — | — |
| Offagna | 042033 | 60020 | — | — |
| Offida | 044054 | 63073 | — | — |
| Ortezzano | 109029 | 63851 | — | — |
| Osimo | 042034 | 60027 | — | — |
| Ostra | 042035 | 60010 | — | — |
| Ostra Vetere | 042036 | 60010 | — | — |
| Palmiano | 044056 | 63092 | — | — |
| Pedaso | 109030 | 63827 | — | — |
| Peglio | 041041 | 61049 | — | — |
| Penna San Giovanni | 043035 | 62020 | — | — |
| Pergola | 041043 | 61045 | — | — |
| Pesaro | 041044 | 61121, 61122 | — | — |
| Petriano | 041045 | 61020 | — | — |
| Petriolo | 043036 | 62014 | — | — |
| Petritoli | 109031 | 63848 | — | — |
| Piandimeleto | 041047 | 61026 | — | — |
| Pietrarubbia | 041048 | 61023 | — | — |
| Pieve Torina | 043038 | 62036 | — | — |
| Piobbico | 041049 | 61046 | — | — |
| Pioraco | 043039 | 62025 | — | — |
| Poggio San Marcello | 042037 | 60030 | — | — |
| Poggio San Vicino | 043040 | 62021 | — | — |
| Pollenza | 043041 | 62010 | — | — |
| Polverigi | 042038 | 60020 | — | — |
| Ponzano di Fermo | 109032 | 63845 | — | — |
| Porto Recanati | 043042 | 62017 | — | — |
| Porto San Giorgio | 109033 | 63822 | — | — |
| Porto Sant'Elpidio | 109034 | 63821 | — | — |
| Potenza Picena | 043043 | 62018 | — | — |
| Rapagnano | 109035 | 63831 | — | — |
| Recanati | 043044 | 62019 | — | — |
| Ripatransone | 044063 | 63065 | — | — |
| Ripe San Ginesio | 043045 | 62020 | — | — |
| Roccafluvione | 044064 | 63093 | — | — |
| Rosora | 042040 | 60030 | — | — |
| Rotella | 044065 | 63071 | — | — |
| San Benedetto del Tronto | 044066 | 63074 | — | — |
| San Costanzo | 041051 | 61039 | — | — |
| San Ginesio | 043046 | 62026 | — | — |
| San Lorenzo in Campo | 041054 | 61047 | — | — |
| San Marcello | 042041 | 60030 | — | — |
| San Paolo di Jesi | 042042 | 60038 | — | — |
| San Severino Marche | 043047 | 62027 | — | — |
| Sant'Angelo in Pontano | 043048 | 62020 | — | — |
| Sant'Angelo in Vado | 041057 | 61048 | — | — |
| Sant'Elpidio a Mare | 109037 | 63811 | — | — |
| Sant'Ippolito | 041058 | 61040 | — | — |
| Santa Maria Nuova | 042043 | 60030 | — | — |
| Santa Vittoria in Matenano | 109036 | 63854 | — | — |
| Sarnano | 043049 | 62028 | — | — |
| Sassocorvaro Auditore | 041071 | 61028 | — | — |
| Sassoferrato | 042044 | 60041 | — | — |
| Sefro | 043050 | 62025 | — | — |
| Senigallia | 042045 | 60019 | — | — |
| Serra de' Conti | 042046 | 60030 | — | — |
| Serra San Quirico | 042047 | 60048 | — | — |
| Serra Sant'Abbondio | 041061 | 61040 | — | — |
| Serrapetrona | 043051 | 62020 | — | — |
| Serravalle di Chienti | 043052 | 62038 | — | — |
| Servigliano | 109038 | 63839 | — | — |
| Sirolo | 042048 | 60020 | — | — |
| Smerillo | 109039 | 63856 | — | — |
| Spinetoli | 044071 | 63078 | — | — |
| Staffolo | 042049 | 60039 | — | — |
| Tavoleto | 041064 | 61020 | — | — |
| Tavullia | 041065 | 61010 | — | — |
| Terre Roveresche | 041070 | 61038 | — | — |
| Tolentino | 043053 | 62029 | — | — |
| Torre San Patrizio | 109040 | 63814 | — | — |
| Trecastelli | 042050 | 60012 | — | — |
| Treia | 043054 | 62010 | — | — |
| Urbania | 041066 | 61049 | — | — |
| Urbino | 041067 | 61029 | — | — |
| Urbisaglia | 043055 | 62010 | — | — |
| Ussita | 043056 | 62039 | — | — |
| Valfornace | 043058 | 62031 | — | — |
| Vallefoglia | 041068 | 61022 | — | — |
| Venarotta | 044073 | 63091 | — | — |
| Visso | 043057 | 62039 | — | — |

<a id="region-14"></a>

### Molise (136 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Molise — region** | 14 | — | Not started | Regional sources not checked |
| Acquaviva Collecroce | 070001 | 86030 | — | — |
| Acquaviva d'Isernia | 094001 | 86080 | — | — |
| Agnone | 094002 | 86081 | — | — |
| Bagnoli del Trigno | 094003 | 86091 | — | — |
| Baranello | 070002 | 86011 | — | — |
| Belmonte del Sannio | 094004 | 86080 | — | — |
| Bojano | 070003 | 86021 | — | — |
| Bonefro | 070004 | 86041 | — | — |
| Busso | 070005 | 86010 | — | — |
| Campobasso | 070006 | 86100 | — | — |
| Campochiaro | 070007 | 86020 | — | — |
| Campodipietra | 070008 | 86010 | — | — |
| Campolieto | 070009 | 86040 | — | — |
| Campomarino | 070010 | 86042 | — | — |
| Cantalupo nel Sannio | 094005 | 86092 | — | — |
| Capracotta | 094006 | 86082 | — | — |
| Carovilli | 094007 | 86083 | — | — |
| Carpinone | 094008 | 86093 | — | — |
| Casacalenda | 070011 | 86043 | — | — |
| Casalciprano | 070012 | 86010 | — | — |
| Castel del Giudice | 094009 | 86080 | — | — |
| Castel San Vincenzo | 094012 | 86071 | — | — |
| Castelbottaccio | 070013 | 86030 | — | — |
| Castellino del Biferno | 070014 | 86020 | — | — |
| Castelmauro | 070015 | 86031 | — | — |
| Castelpetroso | 094010 | 86090 | — | — |
| Castelpizzuto | 094011 | 86090 | — | — |
| Castelverrino | 094013 | 86080 | — | — |
| Castropignano | 070016 | 86010 | — | — |
| Cercemaggiore | 070017 | 86012 | — | — |
| Cercepiccola | 070018 | 86010 | — | — |
| Cerro al Volturno | 094014 | 86072 | — | — |
| Chiauci | 094015 | 86097 | — | — |
| Civitacampomarano | 070019 | 86030 | — | — |
| Civitanova del Sannio | 094016 | 86094 | — | — |
| Colle d'Anchise | 070020 | 86020 | — | — |
| Colletorto | 070021 | 86044 | — | — |
| Colli a Volturno | 094017 | 86073 | — | — |
| Conca Casale | 094018 | 86070 | — | — |
| Duronia | 070022 | 86020 | — | — |
| Ferrazzano | 070023 | 86010 | — | — |
| Filignano | 094019 | 86074 | — | — |
| Forlì del Sannio | 094020 | 86084 | — | — |
| Fornelli | 094021 | 86070 | — | — |
| Fossalto | 070024 | 86020 | — | — |
| Frosolone | 094022 | 86095 | — | — |
| Gambatesa | 070025 | 86013 | — | — |
| Gildone | 070026 | 86010 | — | — |
| Guardialfiera | 070027 | 86030 | — | — |
| Guardiaregia | 070028 | 86014 | — | — |
| Guglionesi | 070029 | 86034 | — | — |
| Isernia | 094023 | 86170 | — | — |
| Jelsi | 070030 | 86015 | — | — |
| Larino | 070031 | 86035 | — | — |
| Limosano | 070032 | 86022 | — | — |
| Longano | 094024 | 86090 | — | — |
| Lucito | 070033 | 86030 | — | — |
| Lupara | 070034 | 86030 | — | — |
| Macchia d'Isernia | 094025 | 86070 | — | — |
| Macchia Valfortore | 070035 | 86040 | — | — |
| Macchiagodena | 094026 | 86096 | — | — |
| Mafalda | 070036 | 86030 | — | — |
| Matrice | 070037 | 86030 | — | — |
| Mirabello Sannitico | 070038 | 86010 | — | — |
| Miranda | 094027 | 86080 | — | — |
| Molise | 070039 | 86020 | — | — |
| Monacilioni | 070040 | 86040 | — | — |
| Montagano | 070041 | 86023 | — | — |
| Montaquila | 094028 | 86070 | — | — |
| Montecilfone | 070042 | 86032 | — | — |
| Montefalcone nel Sannio | 070043 | 86033 | — | — |
| Montelongo | 070044 | 86040 | — | — |
| Montemitro | 070045 | 86030 | — | — |
| Montenero di Bisaccia | 070046 | 86036 | — | — |
| Montenero Val Cocchiara | 094029 | 86080 | — | — |
| Monteroduni | 094030 | 86075 | — | — |
| Montorio nei Frentani | 070047 | 86040 | — | — |
| Morrone del Sannio | 070048 | 86040 | — | — |
| Oratino | 070049 | 86010 | — | — |
| Palata | 070050 | 86037 | — | — |
| Pesche | 094031 | 86090 | — | — |
| Pescolanciano | 094032 | 86097 | — | — |
| Pescopennataro | 094033 | 86080 | — | — |
| Petacciato | 070051 | 86038 | — | — |
| Petrella Tifernina | 070052 | 86024 | — | — |
| Pettoranello del Molise | 094034 | 86090 | — | — |
| Pietrabbondante | 094035 | 86085 | — | — |
| Pietracatella | 070053 | 86040 | — | — |
| Pietracupa | 070054 | 86020 | — | — |
| Pizzone | 094036 | 86071 | — | — |
| Poggio Sannita | 094037 | 86086 | — | — |
| Portocannone | 070055 | 86045 | — | — |
| Pozzilli | 094038 | 86077 | — | — |
| Provvidenti | 070056 | 86040 | — | — |
| Riccia | 070057 | 86016 | — | — |
| Rionero Sannitico | 094039 | 86087 | — | — |
| Ripabottoni | 070058 | 86040 | — | — |
| Ripalimosani | 070059 | 86025 | — | — |
| Roccamandolfi | 094040 | 86092 | — | — |
| Roccasicura | 094041 | 86080 | — | — |
| Roccavivara | 070060 | 86020 | — | — |
| Rocchetta a Volturno | 094042 | 86070 | — | — |
| Rotello | 070061 | 86040 | — | — |
| Salcito | 070062 | 86026 | — | — |
| San Biase | 070063 | 86020 | — | — |
| San Felice del Molise | 070064 | 86030 | — | — |
| San Giacomo degli Schiavoni | 070065 | 86030 | — | — |
| San Giovanni in Galdo | 070066 | 86010 | — | — |
| San Giuliano del Sannio | 070067 | 86010 | — | — |
| San Giuliano di Puglia | 070068 | 86040 | — | — |
| San Martino in Pensilis | 070069 | 86046 | — | — |
| San Massimo | 070070 | 86027 | — | — |
| San Pietro Avellana | 094043 | 86088 | — | — |
| San Polo Matese | 070071 | 86020 | — | — |
| Sant'Agapito | 094044 | 86070 | — | — |
| Sant'Angelo del Pesco | 094046 | 86080 | — | — |
| Sant'Angelo Limosano | 070073 | 86020 | — | — |
| Sant'Elena Sannita | 094047 | 86095 | — | — |
| Sant'Elia a Pianisi | 070074 | 86048 | — | — |
| Santa Croce di Magliano | 070072 | 86047 | — | — |
| Santa Maria del Molise | 094045 | 86096 | — | — |
| Scapoli | 094048 | 86070 | — | — |
| Sepino | 070075 | 86017 | — | — |
| Sessano del Molise | 094049 | 86097 | — | — |
| Sesto Campano | 094050 | 86078 | — | — |
| Spinete | 070076 | 86020 | — | — |
| Tavenna | 070077 | 86030 | — | — |
| Termoli | 070078 | 86039 | — | — |
| Torella del Sannio | 070079 | 86028 | — | — |
| Toro | 070080 | 86018 | — | — |
| Trivento | 070081 | 86029 | — | — |
| Tufara | 070082 | 86010 | — | — |
| Ururi | 070083 | 86049 | — | — |
| Vastogirardi | 094051 | 86089 | — | — |
| Venafro | 094052 | 86079 | — | — |
| Vinchiaturo | 070084 | 86019 | — | — |

<a id="region-01"></a>

### Piemonte (1,180 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Piemonte — region** | 01 | — | Not started | Regional sources not checked |
| Acceglio | 004001 | 12021 | — | — |
| Acqui Terme | 006001 | 15011 | — | — |
| Agliano Terme | 005001 | 14041 | — | — |
| Agliè | 001001 | 10011 | — | — |
| Agrate Conturbia | 003001 | 28010 | — | — |
| Ailoche | 096001 | 13861 | — | — |
| Airasca | 001002 | 10060 | — | — |
| Aisone | 004002 | 12010 | — | — |
| Ala di Stura | 001003 | 10070 | — | — |
| Alagna Valsesia | 002002 | 13021 | — | — |
| Alba | 004003 | 12051 | — | — |
| Albano Vercellese | 002003 | 13030 | — | — |
| Albaretto della Torre | 004004 | 12050 | — | — |
| Albera Ligure | 006002 | 15060 | — | — |
| Albiano d'Ivrea | 001004 | 10010 | — | — |
| Albugnano | 005002 | 14022 | — | — |
| Alessandria | 006003 | 15121, 15122 | — | — |
| Alfiano Natta | 006004 | 15021 | — | — |
| Alice Bel Colle | 006005 | 15010 | — | — |
| Alice Castello | 002004 | 13040 | — | — |
| Alluvioni Piovera | 006192 | 15047 | — | — |
| Almese | 001006 | 10040 | — | — |
| Alpette | 001007 | 10080 | — | — |
| Alpignano | 001008 | 10091 | — | — |
| Altavilla Monferrato | 006007 | 15041 | — | — |
| Alto | 004005 | 12070 | — | — |
| Alto Sermenza | 002170 | 13029 | — | — |
| Alzano Scrivia | 006008 | 15050 | — | — |
| Ameno | 003002 | 28010 | — | — |
| Andezeno | 001009 | 10020 | — | — |
| Andorno Micca | 096002 | 13811 | — | — |
| Andrate | 001010 | 10010 | — | — |
| Angrogna | 001011 | 10060 | — | — |
| Antignano | 005003 | 14010 | — | — |
| Antrona Schieranco | 103001 | 28841 | — | — |
| Anzola d'Ossola | 103002 | 28877 | — | — |
| Aramengo | 005004 | 14020 | — | — |
| Arborio | 002006 | 13031 | — | — |
| Argentera | 004006 | 12010 | — | — |
| Arguello | 004007 | 12050 | — | — |
| Arignano | 001012 | 10020 | — | — |
| Arizzano | 103003 | 28811 | — | — |
| Armeno | 003006 | 28011 | — | — |
| Arola | 103004 | 28899 | — | — |
| Arona | 003008 | 28041 | — | — |
| Arquata Scrivia | 006009 | 15061 | — | — |
| Asigliano Vercellese | 002007 | 13032 | — | — |
| Asti | 005005 | 14100 | — | — |
| Aurano | 103005 | 28812 | — | — |
| Avigliana | 001013 | 10051 | — | — |
| Avolasca | 006010 | 15050 | — | — |
| Azeglio | 001014 | 10010 | — | — |
| Azzano d'Asti | 005006 | 14030 | — | — |
| Baceno | 103006 | 28861 | — | — |
| Bagnasco | 004008 | 12071 | — | — |
| Bagnolo Piemonte | 004009 | 12031 | — | — |
| Bairo | 001015 | 10010 | — | — |
| Balangero | 001016 | 10070 | — | — |
| Baldichieri d'Asti | 005007 | 14011 | — | — |
| Baldissero Canavese | 001017 | 10080 | — | — |
| Baldissero d'Alba | 004010 | 12040 | — | — |
| Baldissero Torinese | 001018 | 10020 | — | — |
| Balme | 001019 | 10070 | — | — |
| Balmuccia | 002008 | 13020 | — | — |
| Balocco | 002009 | 13040 | — | — |
| Balzola | 006011 | 15031 | — | — |
| Banchette | 001020 | 10010 | — | — |
| Bannio Anzino | 103007 | 28871 | — | — |
| Barbania | 001021 | 10070 | — | — |
| Barbaresco | 004011 | 12050 | — | — |
| Bardonecchia | 001022 | 10052 | — | — |
| Barengo | 003012 | 28010 | — | — |
| Barge | 004012 | 12032 | — | — |
| Barolo | 004013 | 12060 | — | — |
| Barone Canavese | 001023 | 10010 | — | — |
| Basaluzzo | 006012 | 15060 | — | — |
| Bassignana | 006013 | 15042 | — | — |
| Bastia Mondovì | 004014 | 12060 | — | — |
| Battifollo | 004015 | 12070 | — | — |
| Baveno | 103008 | 28831 | — | — |
| Bee | 103009 | 28813 | — | — |
| Beinasco | 001024 | 10092 | — | — |
| Beinette | 004016 | 12081 | — | — |
| Belforte Monferrato | 006014 | 15070 | — | — |
| Belgirate | 103010 | 28832 | — | — |
| Bellino | 004017 | 12020 | — | — |
| Bellinzago Novarese | 003016 | 28043 | — | — |
| Belvedere Langhe | 004018 | 12060 | — | — |
| Belveglio | 005008 | 14040 | — | — |
| Bene Vagienna | 004019 | 12041 | — | — |
| Benevello | 004020 | 12050 | — | — |
| Benna | 096003 | 13871 | — | — |
| Bergamasco | 006015 | 15022 | — | — |
| Bergolo | 004021 | 12074 | — | — |
| Bernezzo | 004022 | 12010 | — | — |
| Berzano di San Pietro | 005009 | 14020 | — | — |
| Berzano di Tortona | 006016 | 15050 | — | — |
| Beura-Cardezza | 103011 | 28851 | — | — |
| Biandrate | 003018 | 28061 | — | — |
| Bianzè | 002011 | 13041 | — | — |
| Bibiana | 001025 | 10060 | — | — |
| Biella | 096004 | 13900 | — | — |
| Bioglio | 096005 | 13841 | — | — |
| Bistagno | 006017 | 15012 | — | — |
| Bobbio Pellice | 001026 | 10060 | — | — |
| Boca | 003019 | 28010 | — | — |
| Boccioleto | 002014 | 13022 | — | — |
| Bognanco | 103012 | 28842 | — | — |
| Bogogno | 003021 | 28010 | — | — |
| Bollengo | 001027 | 10012 | — | — |
| Bolzano Novarese | 003022 | 28010 | — | — |
| Bonvicino | 004023 | 12060 | — | — |
| Borgaro Torinese | 001028 | 10071 | — | — |
| Borghetto di Borbera | 006018 | 15060 | — | — |
| Borgiallo | 001029 | 10080 | — | — |
| Borgo d'Ale | 002015 | 13040 | — | — |
| Borgo San Dalmazzo | 004025 | 12011 | — | — |
| Borgo San Martino | 006020 | 15032 | — | — |
| Borgo Ticino | 003025 | 28040 | — | — |
| Borgo Vercelli | 002017 | 13012 | — | — |
| Borgofranco d'Ivrea | 001030 | 10013 | — | — |
| Borgolavezzaro | 003023 | 28071 | — | — |
| Borgomale | 004024 | 12050 | — | — |
| Borgomanero | 003024 | 28021 | — | — |
| Borgomasino | 001031 | 10031 | — | — |
| Borgomezzavalle | 103078 | 28846 | — | — |
| Borgone Susa | 001032 | 10050 | — | — |
| Borgoratto Alessandrino | 006019 | 15013 | — | — |
| Borgosesia | 002016 | 13011 | — | — |
| Borriana | 096006 | 13872 | — | — |
| Bosco Marengo | 006021 | 15062 | — | — |
| Bosconero | 001033 | 10080 | — | — |
| Bosia | 004026 | 12050 | — | — |
| Bosio | 006022 | 15060 | — | — |
| Bossolasco | 004027 | 12060 | — | — |
| Boves | 004028 | 12012 | — | — |
| Bozzole | 006023 | 15040 | — | — |
| Bra | 004029 | 12042 | — | — |
| Brandizzo | 001034 | 10032 | — | — |
| Briaglia | 004030 | 12080 | — | — |
| Bricherasio | 001035 | 10060 | — | — |
| Briga Alta | 004031 | 18025 | — | — |
| Briga Novarese | 003026 | 28010 | — | — |
| Brignano-Frascata | 006024 | 15050 | — | — |
| Briona | 003027 | 28072 | — | — |
| Brondello | 004032 | 12030 | — | — |
| Brossasco | 004033 | 12020 | — | — |
| Brosso | 001036 | 10080 | — | — |
| Brovello-Carpugnino | 103013 | 28833 | — | — |
| Brozolo | 001037 | 10020 | — | — |
| Bruino | 001038 | 10090 | — | — |
| Bruno | 005010 | 14046 | — | — |
| Brusasco | 001039 | 10020 | — | — |
| Brusnengo | 096007 | 13862 | — | — |
| Bruzolo | 001040 | 10050 | — | — |
| Bubbio | 005011 | 14051 | — | — |
| Buriasco | 001041 | 10060 | — | — |
| Burolo | 001042 | 10010 | — | — |
| Buronzo | 002021 | 13040 | — | — |
| Busano | 001043 | 10080 | — | — |
| Busca | 004034 | 12022 | — | — |
| Bussoleno | 001044 | 10053 | — | — |
| Buttigliera Alta | 001045 | 10090 | — | — |
| Buttigliera d'Asti | 005012 | 14021 | — | — |
| Cabella Ligure | 006025 | 15060 | — | — |
| Cafasse | 001046 | 10070 | — | — |
| Calamandrana | 005013 | 14042 | — | — |
| Calasca-Castiglione | 103014 | 28873 | — | — |
| Callabiana | 096008 | 13821 | — | — |
| Calliano Monferrato | 005014 | 14031 | — | — |
| Calosso | 005015 | 14052 | — | — |
| Caltignaga | 003030 | 28010 | — | — |
| Caluso | 001047 | 10014 | — | — |
| Camagna Monferrato | 006026 | 15030 | — | — |
| Camandona | 096009 | 13821 | — | — |
| Cambiano | 001048 | 10020 | — | — |
| Cambiasca | 103015 | 28814 | — | — |
| Camburzano | 096010 | 13891 | — | — |
| Camerana | 004035 | 12072 | — | — |
| Camerano Casasco | 005016 | 14020 | — | — |
| Cameri | 003032 | 28062 | — | — |
| Camino | 006027 | 15020 | — | — |
| Campertogno | 002025 | 13023 | — | — |
| Campiglia Cervo | 096086 | 13812 | — | — |
| Campiglione Fenile | 001049 | 10060 | — | — |
| Canale | 004037 | 12043 | — | — |
| Candelo | 096012 | 13878 | — | — |
| Candia Canavese | 001050 | 10010 | — | — |
| Candiolo | 001051 | 10060 | — | — |
| Canelli | 005017 | 14053 | — | — |
| Canischio | 001052 | 10080 | — | — |
| Cannero Riviera | 103016 | 28821 | — | — |
| Cannobio | 103017 | 28822 | — | — |
| Canosio | 004038 | 12020 | — | — |
| Cantalupa | 001053 | 10060 | — | — |
| Cantalupo Ligure | 006028 | 15060 | — | — |
| Cantarana | 005018 | 14010 | — | — |
| Cantoira | 001054 | 10070 | — | — |
| Caprauna | 004039 | 12070 | — | — |
| Caprezzo | 103018 | 28815 | — | — |
| Capriata d'Orba | 006029 | 15060 | — | — |
| Caprie | 001055 | 10040 | — | — |
| Capriglio | 005019 | 14014 | — | — |
| Caprile | 096013 | 13864 | — | — |
| Caraglio | 004040 | 12023 | — | — |
| Caramagna Piemonte | 004041 | 12030 | — | — |
| Caravino | 001056 | 10010 | — | — |
| Carbonara Scrivia | 006030 | 15050 | — | — |
| Carcoforo | 002029 | 13026 | — | — |
| Cardè | 004042 | 12030 | — | — |
| Carema | 001057 | 10010 | — | — |
| Carentino | 006031 | 15026 | — | — |
| Caresana | 002030 | 13010 | — | — |
| Caresanablot | 002031 | 13030 | — | — |
| Carezzano | 006032 | 15051 | — | — |
| Carignano | 001058 | 10041 | — | — |
| Carisio | 002032 | 13040 | — | — |
| Carmagnola | 001059 | 10022 | — | — |
| Carpeneto | 006033 | 15071 | — | — |
| Carpignano Sesia | 003036 | 28064 | — | — |
| Carrega Ligure | 006034 | 15060 | — | — |
| Carrosio | 006035 | 15060 | — | — |
| Carrù | 004043 | 12061 | — | — |
| Cartignano | 004044 | 12020 | — | — |
| Cartosio | 006036 | 15015 | — | — |
| Casal Cermelli | 006037 | 15072 | — | — |
| Casalbeltrame | 003037 | 28060 | — | — |
| Casalborgone | 001060 | 10020 | — | — |
| Casale Corte Cerro | 103019 | 28881 | — | — |
| Casale Monferrato | 006039 | 15033 | — | — |
| Casaleggio Boiro | 006038 | 15070 | — | — |
| Casaleggio Novara | 003039 | 28060 | — | — |
| Casalgrasso | 004045 | 12030 | — | — |
| Casalino | 003040 | 28060 | — | — |
| Casalnoceto | 006040 | 15052 | — | — |
| Casalvolone | 003041 | 28060 | — | — |
| Casanova Elvo | 002033 | 13030 | — | — |
| Casapinta | 096014 | 13866 | — | — |
| Casasco | 006041 | 15050 | — | — |
| Cascinette d'Ivrea | 001061 | 10010 | — | — |
| Caselette | 001062 | 10040 | — | — |
| Caselle Torinese | 001063 | 10072 | — | — |
| Casorzo Monferrato | 005020 | 14032 | — | — |
| Cassano Spinola | 006191 | 15063 | — | — |
| Cassinasco | 005021 | 14050 | — | — |
| Cassine | 006043 | 15016 | — | — |
| Cassinelle | 006044 | 15070 | — | — |
| Castagneto Po | 001064 | 10090 | — | — |
| Castagnito | 004046 | 12050 | — | — |
| Castagnole delle Lanze | 005022 | 14054 | — | — |
| Castagnole Monferrato | 005023 | 14030 | — | — |
| Castagnole Piemonte | 001065 | 10060 | — | — |
| Castel Boglione | 005024 | 14040 | — | — |
| Castel Rocchero | 005032 | 14044 | — | — |
| Casteldelfino | 004047 | 12020 | — | — |
| Castell'Alfero | 005025 | 14033 | — | — |
| Castellamonte | 001066 | 10081 | — | — |
| Castellania Coppi | 006045 | 15051 | — | — |
| Castellar Guidobono | 006046 | 15050 | — | — |
| Castellazzo Bormida | 006047 | 15073 | — | — |
| Castellazzo Novarese | 003042 | 28060 | — | — |
| Castellero | 005026 | 14013 | — | — |
| Castelletto Cervo | 096015 | 13851 | — | — |
| Castelletto d'Erro | 006048 | 15010 | — | — |
| Castelletto d'Orba | 006049 | 15060 | — | — |
| Castelletto Merli | 006050 | 15020 | — | — |
| Castelletto Molina | 005027 | 14040 | — | — |
| Castelletto Monferrato | 006051 | 15040 | — | — |
| Castelletto sopra Ticino | 003043 | 28053 | — | — |
| Castelletto Stura | 004049 | 12040 | — | — |
| Castelletto Uzzone | 004050 | 12070 | — | — |
| Castellinaldo d'Alba | 004051 | 12050 | — | — |
| Castellino Tanaro | 004052 | 12060 | — | — |
| Castello di Annone | 005028 | 14034 | — | — |
| Castelmagno | 004053 | 12020 | — | — |
| Castelnuovo Belbo | 005029 | 14043 | — | — |
| Castelnuovo Bormida | 006052 | 15017 | — | — |
| Castelnuovo Calcea | 005030 | 14040 | — | — |
| Castelnuovo di Ceva | 004054 | 12070 | — | — |
| Castelnuovo Don Bosco | 005031 | 14022 | — | — |
| Castelnuovo Nigra | 001067 | 10080 | — | — |
| Castelnuovo Scrivia | 006053 | 15053 | — | — |
| Castelspina | 006054 | 15070 | — | — |
| Castiglione Falletto | 004055 | 12060 | — | — |
| Castiglione Tinella | 004056 | 12053 | — | — |
| Castiglione Torinese | 001068 | 10090 | — | — |
| Castino | 004057 | 12050 | — | — |
| Cavaglià | 096016 | 13881 | — | — |
| Cavaglietto | 003044 | 28010 | — | — |
| Cavaglio d'Agogna | 003045 | 28010 | — | — |
| Cavagnolo | 001069 | 10020 | — | — |
| Cavallerleone | 004058 | 12030 | — | — |
| Cavallermaggiore | 004059 | 12030 | — | — |
| Cavallirio | 003047 | 28010 | — | — |
| Cavatore | 006055 | 15010 | — | — |
| Cavour | 001070 | 10061 | — | — |
| Cella Monte | 006056 | 15034 | — | — |
| Cellarengo | 005033 | 14010 | — | — |
| Celle di Macra | 004060 | 12020 | — | — |
| Celle Enomondo | 005034 | 14010 | — | — |
| Cellio con Breia | 002171 | 13024 | — | — |
| Centallo | 004061 | 12044 | — | — |
| Ceppo Morelli | 103021 | 28875 | — | — |
| Cerano | 003049 | 28065 | — | — |
| Cercenasco | 001071 | 10060 | — | — |
| Ceres | 001072 | 10070 | — | — |
| Cereseto | 006057 | 15020 | — | — |
| Ceresole Alba | 004062 | 12040 | — | — |
| Ceresole Reale | 001073 | 10080 | — | — |
| Cerreto d'Asti | 005035 | 14020 | — | — |
| Cerreto Grue | 006058 | 15050 | — | — |
| Cerretto Langhe | 004063 | 12050 | — | — |
| Cerrina Monferrato | 006059 | 15020 | — | — |
| Cerrione | 096018 | 13882 | — | — |
| Cerro Tanaro | 005036 | 14030 | — | — |
| Cervasca | 004064 | 12010 | — | — |
| Cervatto | 002041 | 13025 | — | — |
| Cervere | 004065 | 12040 | — | — |
| Cesana Torinese | 001074 | 10054 | — | — |
| Cesara | 103022 | 28891 | — | — |
| Cessole | 005037 | 14050 | — | — |
| Ceva | 004066 | 12073 | — | — |
| Cherasco | 004067 | 12062 | — | — |
| Chialamberto | 001075 | 10070 | — | — |
| Chianocco | 001076 | 10050 | — | — |
| Chiaverano | 001077 | 10010 | — | — |
| Chieri | 001078 | 10023 | — | — |
| Chiesanuova | 001079 | 10080 | — | — |
| Chiomonte | 001080 | 10050 | — | — |
| Chiusa di Pesio | 004068 | 12013 | — | — |
| Chiusa di San Michele | 001081 | 10050 | — | — |
| Chiusano d'Asti | 005038 | 14025 | — | — |
| Chivasso | 001082 | 10034 | — | — |
| Ciconio | 001083 | 10080 | — | — |
| Cigliano | 002042 | 13043 | — | — |
| Cigliè | 004069 | 12060 | — | — |
| Cinaglio | 005039 | 14020 | — | — |
| Cintano | 001084 | 10080 | — | — |
| Cinzano | 001085 | 10090 | — | — |
| Ciriè | 001086 | 10073 | — | — |
| Cissone | 004070 | 12050 | — | — |
| Cisterna d'Asti | 005040 | 14010 | — | — |
| Civiasco | 002043 | 13010 | — | — |
| Clavesana | 004071 | 12060 | — | — |
| Claviere | 001087 | 10050 | — | — |
| Coassolo Torinese | 001088 | 10070 | — | — |
| Coazze | 001089 | 10050 | — | — |
| Coazzolo | 005041 | 14054 | — | — |
| Cocconato | 005042 | 14023 | — | — |
| Coggiola | 096019 | 13863 | — | — |
| Colazza | 003051 | 28010 | — | — |
| Collegno | 001090 | 10093 | — | — |
| Colleretto Castelnuovo | 001091 | 10080 | — | — |
| Colleretto Giacosa | 001092 | 10010 | — | — |
| Collobiano | 002045 | 13030 | — | — |
| Comignago | 003052 | 28060 | — | — |
| Condove | 001093 | 10055 | — | — |
| Coniolo | 006060 | 15030 | — | — |
| Conzano | 006061 | 15030 | — | — |
| Corio | 001094 | 10070 | — | — |
| Corneliano d'Alba | 004072 | 12040 | — | — |
| Corsione | 005044 | 14020 | — | — |
| Cortandone | 005045 | 14013 | — | — |
| Cortanze | 005046 | 14020 | — | — |
| Cortazzone | 005047 | 14010 | — | — |
| Cortemilia | 004073 | 12074 | — | — |
| Cortiglione | 005048 | 14040 | — | — |
| Cossano Belbo | 004074 | 12054 | — | — |
| Cossano Canavese | 001095 | 10010 | — | — |
| Cossato | 096020 | 13836 | — | — |
| Cossogno | 103023 | 28801 | — | — |
| Cossombrato | 005049 | 14020 | — | — |
| Costa Vescovato | 006062 | 15050 | — | — |
| Costanzana | 002047 | 13033 | — | — |
| Costigliole d'Asti | 005050 | 14055 | — | — |
| Costigliole Saluzzo | 004075 | 12024 | — | — |
| Cravagliana | 002048 | 13020 | — | — |
| Cravanzana | 004076 | 12050 | — | — |
| Craveggia | 103024 | 28852 | — | — |
| Cremolino | 006063 | 15010 | — | — |
| Crescentino | 002049 | 13044 | — | — |
| Cressa | 003055 | 28012 | — | — |
| Crevacuore | 096021 | 13864 | — | — |
| Crevoladossola | 103025 | 28865 | — | — |
| Crissolo | 004077 | 12030 | — | — |
| Crodo | 103026 | 28862 | — | — |
| Crova | 002052 | 13040 | — | — |
| Cuceglio | 001096 | 10090 | — | — |
| Cumiana | 001097 | 10040 | — | — |
| Cuneo | 004078 | 12100 | — | — |
| Cunico | 005051 | 14026 | — | — |
| Cuorgnè | 001098 | 10082 | — | — |
| Cureggio | 003058 | 28060 | — | — |
| Curino | 096023 | 13865 | — | — |
| Demonte | 004079 | 12014 | — | — |
| Denice | 006065 | 15010 | — | — |
| Dernice | 006066 | 15056 | — | — |
| Desana | 002054 | 13034 | — | — |
| Diano d'Alba | 004080 | 12055 | — | — |
| Divignano | 003060 | 28010 | — | — |
| Dogliani | 004081 | 12063 | — | — |
| Domodossola | 103028 | 28845 | — | — |
| Donato | 096024 | 13893 | — | — |
| Dormelletto | 003062 | 28040 | — | — |
| Dorzano | 096025 | 13881 | — | — |
| Dronero | 004082 | 12025 | — | — |
| Druento | 001099 | 10040 | — | — |
| Druogno | 103029 | 28853 | — | — |
| Dusino San Michele | 005052 | 14010 | — | — |
| Elva | 004083 | 12020 | — | — |
| Entracque | 004084 | 12010 | — | — |
| Envie | 004085 | 12030 | — | — |
| Exilles | 001100 | 10050 | — | — |
| Fabbrica Curone | 006067 | 15054 | — | — |
| Fara Novarese | 003065 | 28073 | — | — |
| Farigliano | 004086 | 12060 | — | — |
| Faule | 004087 | 12030 | — | — |
| Favria | 001101 | 10083 | — | — |
| Feisoglio | 004088 | 12050 | — | — |
| Feletto | 001102 | 10080 | — | — |
| Felizzano | 006068 | 15023 | — | — |
| Fenestrelle | 001103 | 10060 | — | — |
| Ferrere | 005053 | 14012 | — | — |
| Fiano | 001104 | 10070 | — | — |
| Fiorano Canavese | 001105 | 10010 | — | — |
| Fobello | 002057 | 13025 | — | — |
| Foglizzo | 001106 | 10090 | — | — |
| Fontaneto d'Agogna | 003066 | 28010 | — | — |
| Fontanetto Po | 002058 | 13040 | — | — |
| Fontanile | 005054 | 14044 | — | — |
| Formazza | 103031 | 28863 | — | — |
| Formigliana | 002059 | 13030 | — | — |
| Forno Canavese | 001107 | 10084 | — | — |
| Fossano | 004089 | 12045 | — | — |
| Frabosa Soprana | 004090 | 12082 | — | — |
| Frabosa Sottana | 004091 | 12083 | — | — |
| Fraconalto | 006069 | 15060 | — | — |
| Francavilla Bisio | 006070 | 15060 | — | — |
| Frascaro | 006071 | 15010 | — | — |
| Frassinello Monferrato | 006072 | 15035 | — | — |
| Frassineto Po | 006073 | 15040 | — | — |
| Frassinetto | 001108 | 10080 | — | — |
| Frassino | 004092 | 12020 | — | — |
| Fresonara | 006074 | 15064 | — | — |
| Frinco | 005055 | 14030 | — | — |
| Front | 001109 | 10070 | — | — |
| Frossasco | 001110 | 10060 | — | — |
| Frugarolo | 006075 | 15065 | — | — |
| Fubine Monferrato | 006076 | 15043 | — | — |
| Gabiano | 006077 | 15020 | — | — |
| Gaglianico | 096026 | 13894 | — | — |
| Gaiola | 004093 | 12010 | — | — |
| Galliate | 003068 | 28066 | — | — |
| Gamalero | 006078 | 15010 | — | — |
| Gambasca | 004094 | 12030 | — | — |
| Garbagna | 006079 | 15050 | — | — |
| Garbagna Novarese | 003069 | 28070 | — | — |
| Garessio | 004095 | 12075 | — | — |
| Gargallo | 003070 | 28010 | — | — |
| Garzigliana | 001111 | 10060 | — | — |
| Gassino Torinese | 001112 | 10090 | — | — |
| Gattico-Veruno | 003166 | 28013 | — | — |
| Gattinara | 002061 | 13045 | — | — |
| Gavi | 006081 | 15066 | — | — |
| Genola | 004096 | 12040 | — | — |
| Germagnano | 001113 | 10070 | — | — |
| Germagno | 103032 | 28887 | — | — |
| Ghemme | 003073 | 28074 | — | — |
| Ghiffa | 103033 | 28823 | — | — |
| Ghislarengo | 002062 | 13030 | — | — |
| Giaglione | 001114 | 10050 | — | — |
| Giarole | 006082 | 15036 | — | — |
| Giaveno | 001115 | 10094 | — | — |
| Gifflenga | 096027 | 13874 | — | — |
| Gignese | 103034 | 28836 | — | — |
| Givoletto | 001116 | 10040 | — | — |
| Gorzegno | 004097 | 12070 | — | — |
| Gottasecca | 004098 | 12070 | — | — |
| Govone | 004099 | 12040 | — | — |
| Gozzano | 003076 | 28024 | — | — |
| Graglia | 096028 | 13895 | — | — |
| Grana Monferrato | 005056 | 14031 | — | — |
| Granozzo con Monticello | 003077 | 28060 | — | — |
| Gravellona Toce | 103035 | 28883 | — | — |
| Gravere | 001117 | 10050 | — | — |
| Grazzano Badoglio | 005057 | 14035 | — | — |
| Greggio | 002065 | 13030 | — | — |
| Gremiasco | 006083 | 15056 | — | — |
| Grignasco | 003079 | 28075 | — | — |
| Grinzane Cavour | 004100 | 12060 | — | — |
| Grognardo | 006084 | 15010 | — | — |
| Grondona | 006085 | 15060 | — | — |
| Groscavallo | 001118 | 10070 | — | — |
| Grosso | 001119 | 10070 | — | — |
| Grugliasco | 001120 | 10095 | — | — |
| Guardabosone | 002066 | 13010 | — | — |
| Guarene | 004101 | 12050 | — | — |
| Guazzora | 006086 | 15050 | — | — |
| Gurro | 103036 | 28828 | — | — |
| Igliano | 004102 | 12060 | — | — |
| Incisa Scapaccino | 005058 | 14045 | — | — |
| Ingria | 001121 | 10080 | — | — |
| Intragna | 103037 | 28816 | — | — |
| Inverso Pinasca | 001122 | 10060 | — | — |
| Invorio | 003082 | 28045 | — | — |
| Isasca | 004103 | 12020 | — | — |
| Isola d'Asti | 005059 | 14057 | — | — |
| Isola Sant'Antonio | 006087 | 15050 | — | — |
| Isolabella | 001123 | 10046 | — | — |
| Issiglio | 001124 | 10080 | — | — |
| Ivrea | 001125 | 10015 | — | — |
| La Cassa | 001126 | 10040 | — | — |
| La Loggia | 001127 | 10040 | — | — |
| La Morra | 004105 | 12064 | — | — |
| Lagnasco | 004104 | 12030 | — | — |
| Lamporo | 002067 | 13046 | — | — |
| Landiona | 003083 | 28064 | — | — |
| Lanzo Torinese | 001128 | 10074 | — | — |
| Lauriano | 001129 | 10020 | — | — |
| Leini | 001130 | 10040 | — | — |
| Lemie | 001131 | 10070 | — | — |
| Lenta | 002068 | 13035 | — | — |
| Lequio Berria | 004106 | 12050 | — | — |
| Lequio Tanaro | 004107 | 12060 | — | — |
| Lerma | 006088 | 15070 | — | — |
| Lesa | 003084 | 28040 | — | — |
| Lesegno | 004108 | 12076 | — | — |
| Lessolo | 001132 | 10010 | — | — |
| Lessona | 096085 | 13853 | — | — |
| Levice | 004109 | 12070 | — | — |
| Levone | 001133 | 10070 | — | — |
| Lignana | 002070 | 13034 | — | — |
| Limone Piemonte | 004110 | 12015 | — | — |
| Lisio | 004111 | 12070 | — | — |
| Livorno Ferraris | 002071 | 13046 | — | — |
| Loazzolo | 005060 | 14051 | — | — |
| Locana | 001134 | 10080 | — | — |
| Lombardore | 001135 | 10040 | — | — |
| Lombriasco | 001136 | 10040 | — | — |
| Loranzè | 001137 | 10010 | — | — |
| Loreglia | 103038 | 28893 | — | — |
| Lozzolo | 002072 | 13045 | — | — |
| Lu e Cuccaro Monferrato | 006193 | 15037 | — | — |
| Luserna San Giovanni | 001139 | 10062 | — | — |
| Lusernetta | 001140 | 10060 | — | — |
| Lusigliè | 001141 | 10080 | — | — |
| Macello | 001142 | 10060 | — | — |
| Macra | 004112 | 12020 | — | — |
| Macugnaga | 103039 | 28876 | — | — |
| Madonna del Sasso | 103040 | 28894 | — | — |
| Maggiora | 003088 | 28014 | — | — |
| Magliano Alfieri | 004113 | 12050 | — | — |
| Magliano Alpi | 004114 | 12060 | — | — |
| Maglione | 001143 | 10030 | — | — |
| Magnano | 096030 | 13887 | — | — |
| Malesco | 103041 | 28854 | — | — |
| Malvicino | 006090 | 15015 | — | — |
| Mandello Vitta | 003090 | 28060 | — | — |
| Mango | 004115 | 12056 | — | — |
| Manta | 004116 | 12030 | — | — |
| Mappano | 001316 | 10079 | — | — |
| Marano Ticino | 003091 | 28040 | — | — |
| Maranzana | 005061 | 14040 | — | — |
| Marene | 004117 | 12030 | — | — |
| Marentino | 001144 | 10020 | — | — |
| Maretto | 005062 | 14018 | — | — |
| Margarita | 004118 | 12040 | — | — |
| Marmora | 004119 | 12020 | — | — |
| Marsaglia | 004120 | 12060 | — | — |
| Martiniana Po | 004121 | 12030 | — | — |
| Masera | 103042 | 28855 | — | — |
| Masio | 006091 | 15024 | — | — |
| Massazza | 096031 | 13873 | — | — |
| Massello | 001145 | 10060 | — | — |
| Masserano | 096032 | 13866 | — | — |
| Massino Visconti | 003093 | 28040 | — | — |
| Massiola | 103043 | 28895 | — | — |
| Mathi | 001146 | 10075 | — | — |
| Mattie | 001147 | 10050 | — | — |
| Mazzè | 001148 | 10035 | — | — |
| Meana di Susa | 001149 | 10050 | — | — |
| Meina | 003095 | 28046 | — | — |
| Melazzo | 006092 | 15010 | — | — |
| Melle | 004122 | 12020 | — | — |
| Merana | 006093 | 15010 | — | — |
| Mercenasco | 001150 | 10010 | — | — |
| Mergozzo | 103044 | 28802 | — | — |
| Mezzana Mortigliengo | 096033 | 13831 | — | — |
| Mezzenile | 001152 | 10070 | — | — |
| Mezzomerico | 003097 | 28040 | — | — |
| Miagliano | 096034 | 13816 | — | — |
| Miasino | 003098 | 28010 | — | — |
| Miazzina | 103045 | 28817 | — | — |
| Mirabello Monferrato | 006094 | 15040 | — | — |
| Moasca | 005063 | 14050 | — | — |
| Moiola | 004123 | 12010 | — | — |
| Molare | 006095 | 15074 | — | — |
| Molino dei Torti | 006096 | 15050 | — | — |
| Mollia | 002078 | 13020 | — | — |
| Mombaldone | 005064 | 14050 | — | — |
| Mombarcaro | 004124 | 12070 | — | — |
| Mombaruzzo | 005065 | 14046 | — | — |
| Mombasiglio | 004125 | 12070 | — | — |
| Mombello di Torino | 001153 | 10020 | — | — |
| Mombello Monferrato | 006097 | 15020 | — | — |
| Mombercelli | 005066 | 14047 | — | — |
| Momo | 003100 | 28015 | — | — |
| Mompantero | 001154 | 10059 | — | — |
| Momperone | 006098 | 15050 | — | — |
| Monale | 005067 | 14013 | — | — |
| Monastero Bormida | 005068 | 14058 | — | — |
| Monastero di Lanzo | 001155 | 10070 | — | — |
| Monastero di Vasco | 004126 | 12080 | — | — |
| Monasterolo Casotto | 004127 | 12080 | — | — |
| Monasterolo di Savigliano | 004128 | 12030 | — | — |
| Moncalieri | 001156 | 10024 | — | — |
| Moncalvo | 005069 | 14036 | — | — |
| Moncenisio | 001157 | 10050 | — | — |
| Moncestino | 006099 | 15020 | — | — |
| Monchiero | 004129 | 12060 | — | — |
| Moncrivello | 002079 | 13040 | — | — |
| Moncucco Torinese | 005070 | 14024 | — | — |
| Mondovì | 004130 | 12084 | — | — |
| Monesiglio | 004131 | 12077 | — | — |
| Monforte d'Alba | 004132 | 12065 | — | — |
| Mongardino | 005071 | 14040 | — | — |
| Mongiardino Ligure | 006100 | 15060 | — | — |
| Mongrando | 096035 | 13888 | — | — |
| Monleale | 006101 | 15059 | — | — |
| Montà | 004133 | 12046 | — | — |
| Montabone | 005072 | 14040 | — | — |
| Montacuto | 006102 | 15050 | — | — |
| Montafia | 005073 | 14014 | — | — |
| Montaldeo | 006103 | 15060 | — | — |
| Montaldo Bormida | 006104 | 15010 | — | — |
| Montaldo di Mondovì | 004134 | 12080 | — | — |
| Montaldo Roero | 004135 | 12040 | — | — |
| Montaldo Scarampi | 005074 | 14048 | — | — |
| Montaldo Torinese | 001158 | 10020 | — | — |
| Montalenghe | 001159 | 10090 | — | — |
| Montalto Dora | 001160 | 10016 | — | — |
| Montanaro | 001161 | 10017 | — | — |
| Montanera | 004136 | 12040 | — | — |
| Montecastello | 006105 | 15040 | — | — |
| Montechiaro d'Acqui | 006106 | 15010 | — | — |
| Montechiaro d'Asti | 005075 | 14025 | — | — |
| Montecrestese | 103046 | 28864 | — | — |
| Montegioco | 006107 | 15050 | — | — |
| Montegrosso d'Asti | 005076 | 14048 | — | — |
| Montelupo Albese | 004137 | 12050 | — | — |
| Montemagno Monferrato | 005077 | 14030 | — | — |
| Montemale di Cuneo | 004138 | 12025 | — | — |
| Montemarzino | 006108 | 15050 | — | — |
| Monterosso Grana | 004139 | 12020 | — | — |
| Montescheno | 103047 | 28843 | — | — |
| Monteu da Po | 001162 | 10020 | — | — |
| Monteu Roero | 004140 | 12040 | — | — |
| Montezemolo | 004141 | 12070 | — | — |
| Monticello d'Alba | 004142 | 12066 | — | — |
| Montiglio Monferrato | 005121 | 14026 | — | — |
| Morano sul Po | 006109 | 15025 | — | — |
| Moransengo-Tonengo | 005122 | 14027 | — | — |
| Morbello | 006110 | 15010 | — | — |
| Moretta | 004143 | 12033 | — | — |
| Moriondo Torinese | 001163 | 10020 | — | — |
| Mornese | 006111 | 15075 | — | — |
| Morozzo | 004144 | 12040 | — | — |
| Morsasco | 006112 | 15010 | — | — |
| Motta de' Conti | 002082 | 13010 | — | — |
| Mottalciata | 096037 | 13874 | — | — |
| Murazzano | 004145 | 12060 | — | — |
| Murello | 004146 | 12030 | — | — |
| Murisengo Monferrato | 006113 | 15020 | — | — |
| Muzzano | 096038 | 13895 | — | — |
| Narzole | 004147 | 12068 | — | — |
| Nebbiuno | 003103 | 28010 | — | — |
| Neive | 004148 | 12052 | — | — |
| Netro | 096039 | 13896 | — | — |
| Neviglie | 004149 | 12050 | — | — |
| Nibbiola | 003104 | 28070 | — | — |
| Nichelino | 001164 | 10042 | — | — |
| Niella Belbo | 004150 | 12050 | — | — |
| Niella Tanaro | 004151 | 12060 | — | — |
| Nizza Monferrato | 005080 | 14049 | — | — |
| Noasca | 001165 | 10080 | — | — |
| Nole | 001166 | 10076 | — | — |
| Nomaglio | 001167 | 10010 | — | — |
| None | 001168 | 10060 | — | — |
| Nonio | 103048 | 28891 | — | — |
| Novalesa | 001169 | 10050 | — | — |
| Novara | 003106 | 28100 | — | — |
| Novello | 004152 | 12060 | — | — |
| Novi Ligure | 006114 | 15067 | — | — |
| Nucetto | 004153 | 12070 | — | — |
| Occhieppo Inferiore | 096040 | 13897 | — | — |
| Occhieppo Superiore | 096041 | 13898 | — | — |
| Occimiano | 006115 | 15040 | — | — |
| Odalengo Grande | 006116 | 15020 | — | — |
| Odalengo Piccolo | 006117 | 15020 | — | — |
| Oggebbio | 103049 | 28824 | — | — |
| Oglianico | 001170 | 10080 | — | — |
| Olcenengo | 002088 | 13047 | — | — |
| Oldenico | 002089 | 13030 | — | — |
| Oleggio | 003108 | 28047 | — | — |
| Oleggio Castello | 003109 | 28040 | — | — |
| Olivola | 006118 | 15030 | — | — |
| Olmo Gentile | 005081 | 14050 | — | — |
| Omegna | 103050 | 28887 | — | — |
| Oncino | 004154 | 12030 | — | — |
| Orbassano | 001171 | 10043 | — | — |
| Orio Canavese | 001172 | 10010 | — | — |
| Ormea | 004155 | 12078 | — | — |
| Ornavasso | 103051 | 28877 | — | — |
| Orsara Bormida | 006119 | 15010 | — | — |
| Orta San Giulio | 003112 | 28016 | — | — |
| Osasco | 001173 | 10060 | — | — |
| Osasio | 001174 | 10040 | — | — |
| Ostana | 004156 | 12030 | — | — |
| Ottiglio | 006120 | 15038 | — | — |
| Oulx | 001175 | 10056 | — | — |
| Ovada | 006121 | 15076 | — | — |
| Oviglio | 006122 | 15026 | — | — |
| Ozegna | 001176 | 10080 | — | — |
| Ozzano Monferrato | 006123 | 15039 | — | — |
| Paderna | 006124 | 15050 | — | — |
| Paesana | 004157 | 12034 | — | — |
| Pagno | 004158 | 12030 | — | — |
| Palazzo Canavese | 001177 | 10010 | — | — |
| Palazzolo Vercellese | 002090 | 13040 | — | — |
| Pallanzeno | 103052 | 28884 | — | — |
| Pamparato | 004159 | 12087 | — | — |
| Pancalieri | 001178 | 10060 | — | — |
| Parella | 001179 | 10010 | — | — |
| Pareto | 006125 | 15010 | — | — |
| Parodi Ligure | 006126 | 15060 | — | — |
| Paroldo | 004160 | 12070 | — | — |
| Paruzzaro | 003114 | 28040 | — | — |
| Passerano Marmorito | 005082 | 14020 | — | — |
| Pasturana | 006127 | 15060 | — | — |
| Pavarolo | 001180 | 10020 | — | — |
| Pavone Canavese | 001181 | 10018 | — | — |
| Pecetto di Valenza | 006128 | 15040 | — | — |
| Pecetto Torinese | 001183 | 10020 | — | — |
| Pella | 003115 | 28010 | — | — |
| Penango | 005083 | 14030 | — | — |
| Perletto | 004161 | 12070 | — | — |
| Perlo | 004162 | 12070 | — | — |
| Perosa Argentina | 001184 | 10063 | — | — |
| Perosa Canavese | 001185 | 10010 | — | — |
| Perrero | 001186 | 10060 | — | — |
| Pertengo | 002091 | 13030 | — | — |
| Pertusio | 001187 | 10080 | — | — |
| Pessinetto | 001188 | 10070 | — | — |
| Pettenasco | 003116 | 28028 | — | — |
| Pettinengo | 096042 | 13843 | — | — |
| Peveragno | 004163 | 12016 | — | — |
| Pezzana | 002093 | 13010 | — | — |
| Pezzolo Valle Uzzone | 004164 | 12070 | — | — |
| Pianezza | 001189 | 10044 | — | — |
| Pianfei | 004165 | 12080 | — | — |
| Piasco | 004166 | 12026 | — | — |
| Piatto | 096043 | 13844 | — | — |
| Piea | 005084 | 14020 | — | — |
| Piedicavallo | 096044 | 13812 | — | — |
| Piedimulera | 103053 | 28885 | — | — |
| Pietra Marazzi | 006129 | 15040 | — | — |
| Pietraporzio | 004167 | 12010 | — | — |
| Pieve Vergonte | 103054 | 28886 | — | — |
| Pila | 002096 | 13020 | — | — |
| Pinasca | 001190 | 10060 | — | — |
| Pinerolo | 001191 | 10064 | — | — |
| Pino d'Asti | 005085 | 14020 | — | — |
| Pino Torinese | 001192 | 10025 | — | — |
| Piobesi d'Alba | 004168 | 12040 | — | — |
| Piobesi Torinese | 001193 | 10040 | — | — |
| Piode | 002097 | 13020 | — | — |
| Piossasco | 001194 | 10045 | — | — |
| Piovà Massaia | 005086 | 14026 | — | — |
| Piozzo | 004169 | 12060 | — | — |
| Pisano | 003119 | 28010 | — | — |
| Piscina | 001195 | 10060 | — | — |
| Piverone | 001196 | 10010 | — | — |
| Pocapaglia | 004170 | 12060 | — | — |
| Pogno | 003120 | 28076 | — | — |
| Poirino | 001197 | 10046 | — | — |
| Pollone | 096046 | 13814 | — | — |
| Polonghera | 004171 | 12030 | — | — |
| Pomaretto | 001198 | 10063 | — | — |
| Pomaro Monferrato | 006131 | 15040 | — | — |
| Pombia | 003121 | 28050 | — | — |
| Ponderano | 096047 | 13875 | — | — |
| Pont Canavese | 001199 | 10085 | — | — |
| Pontechianale | 004172 | 12020 | — | — |
| Pontecurone | 006132 | 15055 | — | — |
| Pontestura | 006133 | 15027 | — | — |
| Ponti | 006134 | 15010 | — | — |
| Ponzano Monferrato | 006135 | 15020 | — | — |
| Ponzone | 006136 | 15010 | — | — |
| Portacomaro | 005087 | 14037 | — | — |
| Porte | 001200 | 10060 | — | — |
| Portula | 096048 | 13833 | — | — |
| Postua | 002102 | 13010 | — | — |
| Pozzol Groppo | 006137 | 15050 | — | — |
| Pozzolo Formigaro | 006138 | 15068 | — | — |
| Pradleves | 004173 | 12027 | — | — |
| Pragelato | 001201 | 10060 | — | — |
| Prali | 001202 | 10060 | — | — |
| Pralormo | 001203 | 10040 | — | — |
| Pralungo | 096049 | 13899 | — | — |
| Pramollo | 001204 | 10065 | — | — |
| Prarolo | 002104 | 13012 | — | — |
| Prarostino | 001205 | 10060 | — | — |
| Prasco | 006139 | 15010 | — | — |
| Prascorsano | 001206 | 10080 | — | — |
| Pratiglione | 001207 | 10080 | — | — |
| Prato Sesia | 003122 | 28077 | — | — |
| Pray | 096050 | 13867 | — | — |
| Prazzo | 004174 | 12028 | — | — |
| Predosa | 006140 | 15077 | — | — |
| Premeno | 103055 | 28818 | — | — |
| Premia | 103056 | 28866 | — | — |
| Premosello-Chiovenda | 103057 | 28803 | — | — |
| Priero | 004175 | 12070 | — | — |
| Priocca | 004176 | 12040 | — | — |
| Priola | 004177 | 12070 | — | — |
| Prunetto | 004178 | 12077 | — | — |
| Quagliuzzo | 001208 | 10010 | — | — |
| Quaranti | 005088 | 14040 | — | — |
| Quaregna Cerreto | 096087 | 13854 | — | — |
| Quargnento | 006141 | 15044 | — | — |
| Quarna Sopra | 103058 | 28898 | — | — |
| Quarna Sotto | 103059 | 28896 | — | — |
| Quarona | 002107 | 13017 | — | — |
| Quassolo | 001209 | 10010 | — | — |
| Quattordio | 006142 | 15028 | — | — |
| Quincinetto | 001210 | 10010 | — | — |
| Quinto Vercellese | 002108 | 13030 | — | — |
| Racconigi | 004179 | 12035 | — | — |
| Rassa | 002110 | 13020 | — | — |
| Re | 103060 | 28856 | — | — |
| Reano | 001211 | 10090 | — | — |
| Recetto | 003129 | 28060 | — | — |
| Refrancore | 005089 | 14030 | — | — |
| Revello | 004180 | 12036 | — | — |
| Revigliasco d'Asti | 005090 | 14010 | — | — |
| Ribordone | 001212 | 10080 | — | — |
| Ricaldone | 006143 | 15010 | — | — |
| Rifreddo | 004181 | 12030 | — | — |
| Rimella | 002113 | 13020 | — | — |
| Rittana | 004182 | 12010 | — | — |
| Riva presso Chieri | 001215 | 10020 | — | — |
| Rivalba | 001213 | 10090 | — | — |
| Rivalta Bormida | 006144 | 15010 | — | — |
| Rivalta di Torino | 001214 | 10040 | — | — |
| Rivara | 001216 | 10080 | — | — |
| Rivarolo Canavese | 001217 | 10086 | — | — |
| Rivarone | 006145 | 15040 | — | — |
| Rivarossa | 001218 | 10040 | — | — |
| Rive | 002115 | 13030 | — | — |
| Rivoli | 001219 | 10098 | — | — |
| Roaschia | 004183 | 12010 | — | — |
| Roascio | 004184 | 12073 | — | — |
| Roasio | 002116 | 13060 | — | — |
| Roatto | 005091 | 14018 | — | — |
| Robassomero | 001220 | 10070 | — | — |
| Robella | 005092 | 14020 | — | — |
| Robilante | 004185 | 12017 | — | — |
| Roburent | 004186 | 12080 | — | — |
| Rocca Canavese | 001221 | 10070 | — | — |
| Rocca Cigliè | 004188 | 12060 | — | — |
| Rocca d'Arazzo | 005093 | 14030 | — | — |
| Rocca de' Baldi | 004189 | 12047 | — | — |
| Rocca Grimalda | 006147 | 15078 | — | — |
| Roccabruna | 004187 | 12020 | — | — |
| Roccaforte Ligure | 006146 | 15060 | — | — |
| Roccaforte Mondovì | 004190 | 12088 | — | — |
| Roccasparvera | 004191 | 12010 | — | — |
| Roccaverano | 005094 | 14050 | — | — |
| Roccavione | 004192 | 12018 | — | — |
| Rocchetta Belbo | 004193 | 12050 | — | — |
| Rocchetta Ligure | 006148 | 15060 | — | — |
| Rocchetta Palafea | 005095 | 14042 | — | — |
| Rocchetta Tanaro | 005096 | 14030 | — | — |
| Roddi | 004194 | 12060 | — | — |
| Roddino | 004195 | 12050 | — | — |
| Rodello | 004196 | 12050 | — | — |
| Roletto | 001222 | 10060 | — | — |
| Romagnano Sesia | 003130 | 28078 | — | — |
| Romano Canavese | 001223 | 10090 | — | — |
| Romentino | 003131 | 28068 | — | — |
| Ronco Biellese | 096053 | 13845 | — | — |
| Ronco Canavese | 001224 | 10080 | — | — |
| Rondissone | 001225 | 10030 | — | — |
| Ronsecco | 002118 | 13036 | — | — |
| Roppolo | 096054 | 13883 | — | — |
| Rorà | 001226 | 10060 | — | — |
| Rosazza | 096055 | 13815 | — | — |
| Rosignano Monferrato | 006149 | 15030 | — | — |
| Rossa | 002121 | 13020 | — | — |
| Rossana | 004197 | 12020 | — | — |
| Rosta | 001228 | 10090 | — | — |
| Roure | 001227 | 10060 | — | — |
| Rovasenda | 002122 | 13040 | — | — |
| Rubiana | 001229 | 10040 | — | — |
| Rueglio | 001230 | 10010 | — | — |
| Ruffia | 004198 | 12030 | — | — |
| Sagliano Micca | 096056 | 13816 | — | — |
| Sala Biellese | 096057 | 13884 | — | — |
| Sala Monferrato | 006150 | 15030 | — | — |
| Salasco | 002126 | 13040 | — | — |
| Salassa | 001231 | 10080 | — | — |
| Salbertrand | 001232 | 10050 | — | — |
| Sale | 006151 | 15045 | — | — |
| Sale delle Langhe | 004199 | 12070 | — | — |
| Sale San Giovanni | 004200 | 12070 | — | — |
| Salerano Canavese | 001233 | 10010 | — | — |
| Sali Vercellese | 002127 | 13040 | — | — |
| Saliceto | 004201 | 12079 | — | — |
| Salmour | 004202 | 12040 | — | — |
| Saluggia | 002128 | 13040 | — | — |
| Salussola | 096058 | 13885 | — | — |
| Saluzzo | 004203 | 12037 | — | — |
| Salza di Pinerolo | 001234 | 10060 | — | — |
| Sambuco | 004204 | 12010 | — | — |
| Samone | 001235 | 10010 | — | — |
| Sampeyre | 004205 | 12020 | — | — |
| San Benedetto Belbo | 004206 | 12050 | — | — |
| San Benigno Canavese | 001236 | 10080 | — | — |
| San Bernardino Verbano | 103061 | 28804 | — | — |
| San Carlo Canavese | 001237 | 10070 | — | — |
| San Colombano Belmonte | 001238 | 10080 | — | — |
| San Cristoforo | 006152 | 15060 | — | — |
| San Damiano d'Asti | 005097 | 14015 | — | — |
| San Damiano Macra | 004207 | 12029 | — | — |
| San Didero | 001239 | 10050 | — | — |
| San Francesco al Campo | 001240 | 10070 | — | — |
| San Germano Chisone | 001242 | 10065 | — | — |
| San Germano Vercellese | 002131 | 13047 | — | — |
| San Giacomo Vercellese | 002035 | 13030 | — | — |
| San Gillio | 001243 | 10040 | — | — |
| San Giorgio Canavese | 001244 | 10090 | — | — |
| San Giorgio Monferrato | 006153 | 15020 | — | — |
| San Giorgio Scarampi | 005098 | 14059 | — | — |
| San Giorio di Susa | 001245 | 10050 | — | — |
| San Giusto Canavese | 001246 | 10090 | — | — |
| San Martino Alfieri | 005099 | 14010 | — | — |
| San Martino Canavese | 001247 | 10010 | — | — |
| San Marzano Oliveto | 005100 | 14050 | — | — |
| San Maurizio Canavese | 001248 | 10077 | — | — |
| San Maurizio d'Opaglio | 003133 | 28017 | — | — |
| San Mauro Torinese | 001249 | 10099 | — | — |
| San Michele Mondovì | 004210 | 12080 | — | — |
| San Nazzaro Sesia | 003134 | 28060 | — | — |
| San Paolo Solbrito | 005101 | 14010 | — | — |
| San Pietro Mosezzo | 003135 | 28060 | — | — |
| San Pietro Val Lemina | 001250 | 10060 | — | — |
| San Ponso | 001251 | 10080 | — | — |
| San Raffaele Cimena | 001252 | 10090 | — | — |
| San Salvatore Monferrato | 006154 | 15046 | — | — |
| San Sebastiano Curone | 006155 | 15056 | — | — |
| San Sebastiano da Po | 001253 | 10020 | — | — |
| San Secondo di Pinerolo | 001254 | 10060 | — | — |
| Sandigliano | 096059 | 13876 | — | — |
| Sanfrè | 004208 | 12040 | — | — |
| Sanfront | 004209 | 12030 | — | — |
| Sangano | 001241 | 10090 | — | — |
| Sant'Agata Fossili | 006156 | 15050 | — | — |
| Sant'Albano Stura | 004211 | 12040 | — | — |
| Sant'Ambrogio di Torino | 001255 | 10057 | — | — |
| Sant'Antonino di Susa | 001256 | 10050 | — | — |
| Santa Maria Maggiore | 103062 | 28857 | — | — |
| Santa Vittoria d'Alba | 004212 | 12069 | — | — |
| Santena | 001257 | 10026 | — | — |
| Santhià | 002133 | 13048 | — | — |
| Santo Stefano Belbo | 004213 | 12058 | — | — |
| Santo Stefano Roero | 004214 | 12040 | — | — |
| Sardigliano | 006157 | 15060 | — | — |
| Sarezzano | 006158 | 15050 | — | — |
| Sauze d'Oulx | 001259 | 10050 | — | — |
| Sauze di Cesana | 001258 | 10054 | — | — |
| Savigliano | 004215 | 12038 | — | — |
| Scagnello | 004216 | 12070 | — | — |
| Scalenghe | 001260 | 10060 | — | — |
| Scarmagno | 001261 | 10010 | — | — |
| Scarnafigi | 004217 | 12030 | — | — |
| Sciolze | 001262 | 10090 | — | — |
| Scopa | 002134 | 13027 | — | — |
| Scopello | 002135 | 13028 | — | — |
| Scurzolengo | 005103 | 14030 | — | — |
| Serole | 005104 | 14050 | — | — |
| Serralunga d'Alba | 004218 | 12050 | — | — |
| Serralunga di Crea | 006159 | 15020 | — | — |
| Serravalle Langhe | 004219 | 12050 | — | — |
| Serravalle Scrivia | 006160 | 15069 | — | — |
| Serravalle Sesia | 002137 | 13037 | — | — |
| Sessame | 005105 | 14058 | — | — |
| Sestriere | 001263 | 10058 | — | — |
| Settime | 005106 | 14020 | — | — |
| Settimo Rottaro | 001264 | 10010 | — | — |
| Settimo Torinese | 001265 | 10036 | — | — |
| Settimo Vittone | 001266 | 10010 | — | — |
| Sezzadio | 006161 | 15079 | — | — |
| Sillavengo | 003138 | 28064 | — | — |
| Silvano d'Orba | 006162 | 15060 | — | — |
| Sinio | 004220 | 12050 | — | — |
| Sizzano | 003139 | 28070 | — | — |
| Soglio | 005107 | 14020 | — | — |
| Solero | 006163 | 15029 | — | — |
| Solonghello | 006164 | 15020 | — | — |
| Somano | 004221 | 12060 | — | — |
| Sommariva del Bosco | 004222 | 12048 | — | — |
| Sommariva Perno | 004223 | 12040 | — | — |
| Sordevolo | 096063 | 13817 | — | — |
| Soriso | 003140 | 28010 | — | — |
| Sostegno | 096064 | 13868 | — | — |
| Sozzago | 003141 | 28060 | — | — |
| Sparone | 001267 | 10080 | — | — |
| Spigno Monferrato | 006165 | 15018 | — | — |
| Spineto Scrivia | 006166 | 15050 | — | — |
| Stazzano | 006167 | 15060 | — | — |
| Strambinello | 001268 | 10010 | — | — |
| Strambino | 001269 | 10019 | — | — |
| Stresa | 103064 | 28838 | — | — |
| Strevi | 006168 | 15019 | — | — |
| Strona | 096065 | 13823 | — | — |
| Stroppiana | 002142 | 13010 | — | — |
| Stroppo | 004224 | 12020 | — | — |
| Suno | 003143 | 28019 | — | — |
| Susa | 001270 | 10059 | — | — |
| Tagliolo Monferrato | 006169 | 15070 | — | — |
| Tarantasca | 004225 | 12020 | — | — |
| Tassarolo | 006170 | 15060 | — | — |
| Tavagnasco | 001271 | 10010 | — | — |
| Tavigliano | 096066 | 13811 | — | — |
| Terdobbiate | 003144 | 28070 | — | — |
| Ternengo | 096067 | 13844 | — | — |
| Terruggia | 006171 | 15030 | — | — |
| Terzo | 006172 | 15010 | — | — |
| Ticineto | 006173 | 15040 | — | — |
| Tigliole | 005108 | 14016 | — | — |
| Toceno | 103065 | 28858 | — | — |
| Tollegno | 096068 | 13818 | — | — |
| Tonco | 005109 | 14039 | — | — |
| Torino | 001272 | 10121, 10122, 10123, 10124, 10125, 10126, 10127, 10128, 10129, 10131, 10132, 10133, 10134, 10135, 10136, 10137, 10138, 10139, 10141, 10142, 10143, 10144, 10145, 10146, 10147, 10148, 10149, 10151, 10152, 10153, 10154, 10155, 10156 | — | — |
| Tornaco | 003146 | 28070 | — | — |
| Torrazza Piemonte | 001273 | 10037 | — | — |
| Torrazzo | 096069 | 13884 | — | — |
| Torre Bormida | 004226 | 12050 | — | — |
| Torre Canavese | 001274 | 10010 | — | — |
| Torre Mondovì | 004227 | 12080 | — | — |
| Torre Pellice | 001275 | 10066 | — | — |
| Torre San Giorgio | 004228 | 12030 | — | — |
| Torresina | 004229 | 12070 | — | — |
| Tortona | 006174 | 15057 | — | — |
| Trana | 001276 | 10090 | — | — |
| Trarego Viggiona | 103066 | 28826 | — | — |
| Trasquera | 103067 | 28868 | — | — |
| Traversella | 001278 | 10080 | — | — |
| Traves | 001279 | 10070 | — | — |
| Trecate | 003149 | 28069 | — | — |
| Treiso | 004230 | 12050 | — | — |
| Treville | 006175 | 15030 | — | — |
| Trezzo Tinella | 004231 | 12050 | — | — |
| Tricerro | 002147 | 13038 | — | — |
| Trinità | 004232 | 12049 | — | — |
| Trino | 002148 | 13039 | — | — |
| Trisobbio | 006176 | 15070 | — | — |
| Trofarello | 001280 | 10028 | — | — |
| Trontano | 103068 | 28859 | — | — |
| Tronzano Vercellese | 002150 | 13049 | — | — |
| Usseaux | 001281 | 10060 | — | — |
| Usseglio | 001282 | 10070 | — | — |
| Vaglio Serra | 005111 | 14049 | — | — |
| Vaie | 001283 | 10050 | — | — |
| Val della Torre | 001284 | 10040 | — | — |
| Val di Chy | 001317 | 10039 | — | — |
| Valchiusa | 001318 | 10089 | — | — |
| Valdengo | 096071 | 13855 | — | — |
| Valdieri | 004233 | 12010 | — | — |
| Valdilana | 096088 | 13835 | — | — |
| Valduggia | 002152 | 13018 | — | — |
| Valenza | 006177 | 15048 | — | — |
| Valfenera | 005112 | 14017 | — | — |
| Valgioie | 001285 | 10094 | — | — |
| Valgrana | 004234 | 12020 | — | — |
| Vallanzengo | 096072 | 13847 | — | — |
| Valle Cannobina | 103079 | 28827 | — | — |
| Valle San Nicolao | 096074 | 13847 | — | — |
| Vallo Torinese | 001286 | 10070 | — | — |
| Valloriate | 004235 | 12010 | — | — |
| Valmacca | 006178 | 15040 | — | — |
| Valperga | 001287 | 10087 | — | — |
| Valprato Soana | 001288 | 10080 | — | — |
| Valstrona | 103069 | 28897 | — | — |
| Vanzone con San Carlo | 103070 | 28879 | — | — |
| Vaprio d'Agogna | 003153 | 28010 | — | — |
| Varallo | 002156 | 13019 | — | — |
| Varallo Pombia | 003154 | 28040 | — | — |
| Varisella | 001289 | 10070 | — | — |
| Varzo | 103071 | 28868 | — | — |
| Vauda Canavese | 001290 | 10070 | — | — |
| Veglio | 096075 | 13824 | — | — |
| Venaria Reale | 001292 | 10078 | — | — |
| Venasca | 004237 | 12020 | — | — |
| Venaus | 001291 | 10050 | — | — |
| Verbania | 103072 | 28921, 28922, 28923, 28924, 28925 | — | — |
| Vercelli | 002158 | 13100 | — | — |
| Verduno | 004238 | 12060 | — | — |
| Vernante | 004239 | 12019 | — | — |
| Verolengo | 001293 | 10038 | — | — |
| Verrone | 096076 | 13871 | — | — |
| Verrua Savoia | 001294 | 10020 | — | — |
| Verzuolo | 004240 | 12039 | — | — |
| Vesime | 005113 | 14059 | — | — |
| Vespolate | 003158 | 28079 | — | — |
| Vestignè | 001295 | 10030 | — | — |
| Vezza d'Alba | 004241 | 12040 | — | — |
| Viale | 005114 | 14010 | — | — |
| Vialfrè | 001296 | 10090 | — | — |
| Viarigi | 005115 | 14030 | — | — |
| Vicoforte | 004242 | 12080 | — | — |
| Vicolungo | 003159 | 28060 | — | — |
| Vidracco | 001298 | 10080 | — | — |
| Vigliano Biellese | 096077 | 13856 | — | — |
| Vigliano d'Asti | 005116 | 14040 | — | — |
| Vignale Monferrato | 006179 | 15049 | — | — |
| Vignole Borbera | 006180 | 15060 | — | — |
| Vignolo | 004243 | 12010 | — | — |
| Vignone | 103074 | 28819 | — | — |
| Vigone | 001299 | 10067 | — | — |
| Viguzzolo | 006181 | 15058 | — | — |
| Villa del Bosco | 096078 | 13868 | — | — |
| Villa San Secondo | 005119 | 14020 | — | — |
| Villadeati | 006182 | 15020 | — | — |
| Villadossola | 103075 | 28844 | — | — |
| Villafalletto | 004244 | 12020 | — | — |
| Villafranca d'Asti | 005117 | 14018 | — | — |
| Villafranca Piemonte | 001300 | 10068 | — | — |
| Villalvernia | 006183 | 15050 | — | — |
| Villamiroglio | 006184 | 15020 | — | — |
| Villanova Biellese | 096079 | 13877 | — | — |
| Villanova Canavese | 001301 | 10070 | — | — |
| Villanova d'Asti | 005118 | 14019 | — | — |
| Villanova Mondovì | 004245 | 12089 | — | — |
| Villanova Monferrato | 006185 | 15030 | — | — |
| Villanova Solaro | 004246 | 12030 | — | — |
| Villar Dora | 001303 | 10040 | — | — |
| Villar Focchiardo | 001305 | 10050 | — | — |
| Villar Pellice | 001306 | 10060 | — | — |
| Villar Perosa | 001307 | 10069 | — | — |
| Villar San Costanzo | 004247 | 12020 | — | — |
| Villarbasse | 001302 | 10090 | — | — |
| Villarboit | 002163 | 13030 | — | — |
| Villareggia | 001304 | 10030 | — | — |
| Villaromagnano | 006186 | 15050 | — | — |
| Villastellone | 001308 | 10029 | — | — |
| Villata | 002164 | 13010 | — | — |
| Villette | 103076 | 28856 | — | — |
| Vinadio | 004248 | 12010 | — | — |
| Vinchio | 005120 | 14040 | — | — |
| Vinovo | 001309 | 10048 | — | — |
| Vinzaglio | 003164 | 28060 | — | — |
| Viola | 004249 | 12070 | — | — |
| Virle Piemonte | 001310 | 10060 | — | — |
| Vische | 001311 | 10030 | — | — |
| Visone | 006187 | 15010 | — | — |
| Vistrorio | 001312 | 10080 | — | — |
| Viù | 001313 | 10070 | — | — |
| Viverone | 096080 | 13886 | — | — |
| Vocca | 002166 | 13020 | — | — |
| Vogogna | 103077 | 28805 | — | — |
| Volpedo | 006188 | 15059 | — | — |
| Volpeglino | 006189 | 15050 | — | — |
| Volpiano | 001314 | 10088 | — | — |
| Voltaggio | 006190 | 15060 | — | — |
| Volvera | 001315 | 10040 | — | — |
| Vottignasco | 004250 | 12020 | — | — |
| Zimone | 096081 | 13887 | — | — |
| Zubiena | 096082 | 13888 | — | — |
| Zumaglia | 096083 | 13848 | — | — |

<a id="region-16"></a>

### Puglia (257 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Puglia — region** | 16 | — | Not started | Regional sources not checked |
| Accadia | 071001 | 71021 | — | — |
| Acquaviva delle Fonti | 072001 | 70021 | — | — |
| Adelfia | 072002 | 70010 | — | — |
| Alberobello | 072003 | 70011 | — | — |
| Alberona | 071002 | 71031 | — | — |
| Alessano | 075002 | 73031 | — | — |
| Alezio | 075003 | 73011 | — | — |
| Alliste | 075004 | 73040 | — | — |
| Altamura | 072004 | 70022 | — | — |
| Andrano | 075005 | 73032 | — | — |
| Andria | 110001 | 76123 | — | — |
| Anzano di Puglia | 071003 | 71020 | — | — |
| Apricena | 071004 | 71011 | — | — |
| Aradeo | 075006 | 73040 | — | — |
| Arnesano | 075007 | 73010 | — | — |
| Ascoli Satriano | 071005 | 71022 | — | — |
| Avetrana | 073001 | 74020 | — | — |
| Bagnolo del Salento | 075008 | 73020 | — | — |
| Bari | 072006 | 70121, 70122, 70123, 70124, 70125, 70126, 70127, 70128, 70129, 70131, 70132 | — | — |
| Barletta | 110002 | 76121 | — | — |
| Biccari | 071006 | 71032 | — | — |
| Binetto | 072008 | 70020 | — | — |
| Bisceglie | 110003 | 76011 | — | — |
| Bitetto | 072010 | 70020 | — | — |
| Bitonto | 072011 | 70032 | — | — |
| Bitritto | 072012 | 70020 | — | — |
| Botrugno | 075009 | 73020 | — | — |
| Bovino | 071007 | 71023 | — | — |
| Brindisi | 074001 | 72100 | — | — |
| Cagnano Varano | 071008 | 71010 | — | — |
| Calimera | 075010 | 73021 | — | — |
| Campi Salentina | 075011 | 73012 | — | — |
| Candela | 071009 | 71024 | — | — |
| Cannole | 075012 | 73020 | — | — |
| Canosa di Puglia | 110004 | 76012 | — | — |
| Caprarica di Lecce | 075013 | 73010 | — | — |
| Capurso | 072014 | 70010 | — | — |
| Carapelle | 071010 | 71041 | — | — |
| Carlantino | 071011 | 71030 | — | — |
| Carmiano | 075014 | 73041 | — | — |
| Carosino | 073002 | 74021 | — | — |
| Carovigno | 074002 | 72012 | — | — |
| Carpignano Salentino | 075015 | 73020 | — | — |
| Carpino | 071012 | 71010 | — | — |
| Casalnuovo Monterotaro | 071013 | 71033 | — | — |
| Casalvecchio di Puglia | 071014 | 71030 | — | — |
| Casamassima | 072015 | 70010 | — | — |
| Casarano | 075016 | 73042 | — | — |
| Cassano delle Murge | 072016 | 70020 | — | — |
| Castellana Grotte | 072017 | 70013 | — | — |
| Castellaneta | 073003 | 74011 | — | — |
| Castelluccio dei Sauri | 071015 | 71025 | — | — |
| Castelluccio Valmaggiore | 071016 | 71020 | — | — |
| Castelnuovo della Daunia | 071017 | 71034 | — | — |
| Castri di Lecce | 075017 | 73020 | — | — |
| Castrignano de' Greci | 075018 | 73020 | — | — |
| Castrignano del Capo | 075019 | 73040 | — | — |
| Castro | 075096 | 73030 | — | — |
| Cavallino | 075020 | 73020 | — | — |
| Ceglie Messapica | 074003 | 72013 | — | — |
| Celenza Valfortore | 071018 | 71035 | — | — |
| Cellamare | 072018 | 70010 | — | — |
| Celle di San Vito | 071019 | 71020 | — | — |
| Cellino San Marco | 074004 | 72020 | — | — |
| Cerignola | 071020 | 71042 | — | — |
| Chieuti | 071021 | 71010 | — | — |
| Cisternino | 074005 | 72014 | — | — |
| Collepasso | 075021 | 73040 | — | — |
| Conversano | 072019 | 70014 | — | — |
| Copertino | 075022 | 73043 | — | — |
| Corato | 072020 | 70033 | — | — |
| Corigliano d'Otranto | 075023 | 73022 | — | — |
| Corsano | 075024 | 73033 | — | — |
| Crispiano | 073004 | 74012 | — | — |
| Cursi | 075025 | 73020 | — | — |
| Cutrofiano | 075026 | 73020 | — | — |
| Deliceto | 071022 | 71026 | — | — |
| Diso | 075027 | 73030 | — | — |
| Erchie | 074006 | 72020 | — | — |
| Faeto | 071023 | 71020 | — | — |
| Faggiano | 073005 | 74020 | — | — |
| Fasano | 074007 | 72015 | — | — |
| Foggia | 071024 | 71121, 71122 | — | — |
| Fragagnano | 073006 | 74022 | — | — |
| Francavilla Fontana | 074008 | 72021 | — | — |
| Gagliano del Capo | 075028 | 73034 | — | — |
| Galatina | 075029 | 73013 | — | — |
| Galatone | 075030 | 73044 | — | — |
| Gallipoli | 075031 | 73014 | — | — |
| Ginosa | 073007 | 74013 | — | — |
| Gioia del Colle | 072021 | 70023 | — | — |
| Giovinazzo | 072022 | 70054 | — | — |
| Giuggianello | 075032 | 73030 | — | — |
| Giurdignano | 075033 | 73020 | — | — |
| Gravina in Puglia | 072023 | 70024 | — | — |
| Grottaglie | 073008 | 74023 | — | — |
| Grumo Appula | 072024 | 70025 | — | — |
| Guagnano | 075034 | 73010 | — | — |
| Ischitella | 071025 | 71010 | — | — |
| Isole Tremiti | 071026 | 71051 | — | — |
| Laterza | 073009 | 74014 | — | — |
| Latiano | 074009 | 72022 | — | — |
| Lecce | 075035 | 73100 | — | — |
| Leporano | 073010 | 74020 | — | — |
| Lequile | 075036 | 73010 | — | — |
| Lesina | 071027 | 71010 | — | — |
| Leverano | 075037 | 73045 | — | — |
| Lizzanello | 075038 | 73023 | — | — |
| Lizzano | 073011 | 74020 | — | — |
| Locorotondo | 072025 | 70010 | — | — |
| Lucera | 071028 | 71036 | — | — |
| Maglie | 075039 | 73024 | — | — |
| Manduria | 073012 | 74024 | — | — |
| Manfredonia | 071029 | 71043 | — | — |
| Margherita di Savoia | 110005 | 76016 | — | — |
| Martano | 075040 | 73025 | — | — |
| Martignano | 075041 | 73020 | — | — |
| Martina Franca | 073013 | 74015 | — | — |
| Maruggio | 073014 | 74020 | — | — |
| Massafra | 073015 | 74016 | — | — |
| Matino | 075042 | 73046 | — | — |
| Mattinata | 071031 | 71030 | — | — |
| Melendugno | 075043 | 73026 | — | — |
| Melissano | 075044 | 73040 | — | — |
| Melpignano | 075045 | 73020 | — | — |
| Mesagne | 074010 | 72023 | — | — |
| Miggiano | 075046 | 73035 | — | — |
| Minervino di Lecce | 075047 | 73027 | — | — |
| Minervino Murge | 110006 | 76013 | — | — |
| Modugno | 072027 | 70026 | — | — |
| Mola di Bari | 072028 | 70042 | — | — |
| Molfetta | 072029 | 70056 | — | — |
| Monopoli | 072030 | 70043 | — | — |
| Monte Sant'Angelo | 071033 | 71037 | — | — |
| Monteiasi | 073016 | 74020 | — | — |
| Monteleone di Puglia | 071032 | 71020 | — | — |
| Montemesola | 073017 | 74020 | — | — |
| Monteparano | 073018 | 74020 | — | — |
| Monteroni di Lecce | 075048 | 73047 | — | — |
| Montesano Salentino | 075049 | 73030 | — | — |
| Morciano di Leuca | 075050 | 73040 | — | — |
| Motta Montecorvino | 071034 | 71030 | — | — |
| Mottola | 073019 | 74017 | — | — |
| Muro Leccese | 075051 | 73036 | — | — |
| Nardò | 075052 | 73048 | — | — |
| Neviano | 075053 | 73040 | — | — |
| Noci | 072031 | 70015 | — | — |
| Nociglia | 075054 | 73020 | — | — |
| Noicattaro | 072032 | 70016 | — | — |
| Novoli | 075055 | 73051 | — | — |
| Ordona | 071063 | 71040 | — | — |
| Oria | 074011 | 72024 | — | — |
| Orsara di Puglia | 071035 | 71027 | — | — |
| Orta Nova | 071036 | 71045 | — | — |
| Ortelle | 075056 | 73030 | — | — |
| Ostuni | 074012 | 72017 | — | — |
| Otranto | 075057 | 73028 | — | — |
| Palagianello | 073020 | 74018 | — | — |
| Palagiano | 073021 | 74019 | — | — |
| Palmariggi | 075058 | 73020 | — | — |
| Palo del Colle | 072033 | 70027 | — | — |
| Panni | 071037 | 71020 | — | — |
| Parabita | 075059 | 73052 | — | — |
| Patù | 075060 | 73053 | — | — |
| Peschici | 071038 | 71010 | — | — |
| Pietramontecorvino | 071039 | 71038 | — | — |
| Poggiardo | 075061 | 73037 | — | — |
| Poggio Imperiale | 071040 | 71010 | — | — |
| Poggiorsini | 072034 | 70020 | — | — |
| Polignano a Mare | 072035 | 70044 | — | — |
| Porto Cesareo | 075097 | 73010 | — | — |
| Presicce-Acquarica | 075098 | 73054 | — | — |
| Pulsano | 073022 | 74026 | — | — |
| Putignano | 072036 | 70017 | — | — |
| Racale | 075063 | 73055 | — | — |
| Rignano Garganico | 071041 | 71010 | — | — |
| Roccaforzata | 073023 | 74020 | — | — |
| Rocchetta Sant'Antonio | 071042 | 71020 | — | — |
| Rodi Garganico | 071043 | 71012 | — | — |
| Roseto Valfortore | 071044 | 71039 | — | — |
| Ruffano | 075064 | 73049 | — | — |
| Rutigliano | 072037 | 70018 | — | — |
| Ruvo di Puglia | 072038 | 70037 | — | — |
| Salice Salentino | 075065 | 73015 | — | — |
| Salve | 075066 | 73050 | — | — |
| Sammichele di Bari | 072039 | 70010 | — | — |
| San Cassiano | 075095 | 73020 | — | — |
| San Cesario di Lecce | 075068 | 73016 | — | — |
| San Donaci | 074013 | 72025 | — | — |
| San Donato di Lecce | 075069 | 73010 | — | — |
| San Ferdinando di Puglia | 110007 | 76017 | — | — |
| San Giorgio Ionico | 073024 | 74027 | — | — |
| San Giovanni Rotondo | 071046 | 71013 | — | — |
| San Marco in Lamis | 071047 | 71014 | — | — |
| San Marco la Catola | 071048 | 71030 | — | — |
| San Marzano di San Giuseppe | 073025 | 74020 | — | — |
| San Michele Salentino | 074014 | 72018 | — | — |
| San Nicandro Garganico | 071049 | 71015 | — | — |
| San Pancrazio Salentino | 074015 | 72026 | — | — |
| San Paolo di Civitate | 071050 | 71010 | — | — |
| San Pietro in Lama | 075071 | 73010 | — | — |
| San Pietro Vernotico | 074016 | 72027 | — | — |
| San Severo | 071051 | 71016 | — | — |
| San Vito dei Normanni | 074017 | 72019 | — | — |
| Sanarica | 075067 | 73030 | — | — |
| Sannicandro di Bari | 072040 | 70028 | — | — |
| Sannicola | 075070 | 73017 | — | — |
| Sant'Agata di Puglia | 071052 | 71028 | — | — |
| Santa Cesarea Terme | 075072 | 73020 | — | — |
| Santeramo in Colle | 072041 | 70029 | — | — |
| Sava | 073026 | 74028 | — | — |
| Scorrano | 075073 | 73020 | — | — |
| Seclì | 075074 | 73050 | — | — |
| Serracapriola | 071053 | 71010 | — | — |
| Sogliano Cavour | 075075 | 73010 | — | — |
| Soleto | 075076 | 73010 | — | — |
| Specchia | 075077 | 73040 | — | — |
| Spinazzola | 110008 | 76014 | — | — |
| Spongano | 075078 | 73038 | — | — |
| Squinzano | 075079 | 73018 | — | — |
| Statte | 073029 | 74010 | — | — |
| Sternatia | 075080 | 73010 | — | — |
| Stornara | 071054 | 71047 | — | — |
| Stornarella | 071055 | 71048 | — | — |
| Supersano | 075081 | 73040 | — | — |
| Surano | 075082 | 73030 | — | — |
| Surbo | 075083 | 73010 | — | — |
| Taranto | 073027 | 74121, 74122, 74123 | — | — |
| Taurisano | 075084 | 73056 | — | — |
| Taviano | 075085 | 73057 | — | — |
| Terlizzi | 072043 | 70038 | — | — |
| Tiggiano | 075086 | 73030 | — | — |
| Torchiarolo | 074018 | 72020 | — | — |
| Toritto | 072044 | 70020 | — | — |
| Torre Santa Susanna | 074019 | 72028 | — | — |
| Torremaggiore | 071056 | 71017 | — | — |
| Torricella | 073028 | 74020 | — | — |
| Trani | 110009 | 76125 | — | — |
| Trepuzzi | 075087 | 73019 | — | — |
| Tricase | 075088 | 73039 | — | — |
| Triggiano | 072046 | 70019 | — | — |
| Trinitapoli | 110010 | 76015 | — | — |
| Troia | 071058 | 71029 | — | — |
| Tuglie | 075089 | 73058 | — | — |
| Turi | 072047 | 70010 | — | — |
| Ugento | 075090 | 73059 | — | — |
| Uggiano la Chiesa | 075091 | 73020 | — | — |
| Valenzano | 072048 | 70010 | — | — |
| Veglie | 075092 | 73010 | — | — |
| Vernole | 075093 | 73029 | — | — |
| Vico del Gargano | 071059 | 71018 | — | — |
| Vieste | 071060 | 71019 | — | — |
| Villa Castelli | 074020 | 72029 | — | — |
| Volturara Appula | 071061 | 71030 | — | — |
| Volturino | 071062 | 71030 | — | — |
| Zapponeta | 071064 | 71030 | — | — |
| Zollino | 075094 | 73010 | — | — |

<a id="region-20"></a>

### Sardegna (377 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Sardegna — region** | 20 | — | Not started | Regional sources not checked |
| Abbasanta | 115001 | 09071 | — | — |
| Aggius | 113001 | 07020 | — | — |
| Aglientu | 113002 | 07020 | — | — |
| Aidomaggiore | 115002 | 09070 | — | — |
| Alà dei Sardi | 113003 | 07020 | — | — |
| Albagiara | 115003 | 09090 | — | — |
| Ales | 115004 | 09091 | — | — |
| Alghero | 112001 | 07041 | — | — |
| Allai | 115005 | 09080 | — | — |
| Anela | 112002 | 07010 | — | — |
| Arborea | 115006 | 09092 | — | — |
| Arbus | 117001 | 09031 | — | — |
| Ardara | 112003 | 07010 | — | — |
| Ardauli | 115007 | 09081 | — | — |
| Aritzo | 114001 | 08031 | — | — |
| Armungia | 118001 | 09040 | — | — |
| Arzachena | 113004 | 07021 | — | — |
| Arzana | 116001 | 08040 | — | — |
| Assemini | 118002 | 09032 | — | — |
| Assolo | 115008 | 09080 | — | — |
| Asuni | 115009 | 09080 | — | — |
| Atzara | 114002 | 08030 | — | — |
| Austis | 114003 | 08030 | — | — |
| Badesi | 113005 | 07030 | — | — |
| Ballao | 118003 | 09040 | — | — |
| Banari | 112004 | 07040 | — | — |
| Baradili | 115010 | 09090 | — | — |
| Baratili San Pietro | 115011 | 09070 | — | — |
| Baressa | 115012 | 09090 | — | — |
| Bari Sardo | 116002 | 08042 | — | — |
| Barrali | 118004 | 09040 | — | — |
| Barumini | 117002 | 09021 | — | — |
| Bauladu | 115013 | 09070 | — | — |
| Baunei | 116003 | 08040 | — | — |
| Belvì | 114004 | 08030 | — | — |
| Benetutti | 112005 | 07010 | — | — |
| Berchidda | 113006 | 07022 | — | — |
| Bessude | 112006 | 07040 | — | — |
| Bidonì | 115014 | 09080 | — | — |
| Birori | 114005 | 08010 | — | — |
| Bitti | 114006 | 08021 | — | — |
| Bolotana | 114007 | 08011 | — | — |
| Bonarcado | 115015 | 09070 | — | — |
| Bonnanaro | 112007 | 07043 | — | — |
| Bono | 112008 | 07011 | — | — |
| Bonorva | 112009 | 07012 | — | — |
| Boroneddu | 115016 | 09080 | — | — |
| Borore | 114008 | 08016 | — | — |
| Bortigali | 114009 | 08012 | — | — |
| Bortigiadas | 113007 | 07030 | — | — |
| Borutta | 112010 | 07040 | — | — |
| Bosa | 115017 | 09089 | — | — |
| Bottidda | 112011 | 07010 | — | — |
| Buddusò | 113008 | 07020 | — | — |
| Budoni | 113009 | 07051 | — | — |
| Buggerru | 119001 | 09010 | — | — |
| Bultei | 112012 | 07010 | — | — |
| Bulzi | 112013 | 07030 | — | — |
| Burcei | 118005 | 09040 | — | — |
| Burgos | 112014 | 07010 | — | — |
| Busachi | 115018 | 09082 | — | — |
| Cabras | 115019 | 09072 | — | — |
| Cagliari | 118006 | 09121, 09122, 09123, 09124, 09125, 09126, 09127, 09128, 09129, 09131, 09134 | — | — |
| Calangianus | 113010 | 07023 | — | — |
| Calasetta | 119002 | 09011 | — | — |
| Capoterra | 118007 | 09012 | — | — |
| Carbonia | 119003 | 09013 | — | — |
| Cardedu | 116004 | 08040 | — | — |
| Cargeghe | 112015 | 07030 | — | — |
| Carloforte | 119004 | 09014 | — | — |
| Castelsardo | 112016 | 07031 | — | — |
| Castiadas | 118008 | 09040 | — | — |
| Cheremule | 112017 | 07040 | — | — |
| Chiaramonti | 112018 | 07030 | — | — |
| Codrongianos | 112019 | 07040 | — | — |
| Collinas | 117003 | 09020 | — | — |
| Cossoine | 112020 | 07010 | — | — |
| Cuglieri | 115020 | 09073 | — | — |
| Curcuris | 115021 | 09090 | — | — |
| Decimomannu | 118009 | 09033 | — | — |
| Decimoputzu | 118010 | 09010 | — | — |
| Desulo | 114010 | 08032 | — | — |
| Dolianova | 118011 | 09041 | — | — |
| Domus de Maria | 118012 | 09010 | — | — |
| Domusnovas | 119005 | 09015 | — | — |
| Donori | 118013 | 09040 | — | — |
| Dorgali | 114011 | 08022 | — | — |
| Dualchi | 114012 | 08010 | — | — |
| Elini | 116005 | 08040 | — | — |
| Elmas | 118014 | 09067 | — | — |
| Erula | 112021 | 07030 | — | — |
| Escalaplano | 118015 | 09051 | — | — |
| Escolca | 118016 | 09052 | — | — |
| Esporlatu | 112022 | 07010 | — | — |
| Esterzili | 118017 | 09053 | — | — |
| Florinas | 112023 | 07030 | — | — |
| Fluminimaggiore | 119006 | 09010 | — | — |
| Flussio | 115022 | 09090 | — | — |
| Fonni | 114013 | 08023 | — | — |
| Fordongianus | 115023 | 09083 | — | — |
| Furtei | 117004 | 09040 | — | — |
| Gadoni | 114014 | 08030 | — | — |
| Gairo | 116006 | 08040 | — | — |
| Galtellì | 114015 | 08020 | — | — |
| Gavoi | 114016 | 08020 | — | — |
| Genoni | 118018 | 09054 | — | — |
| Genuri | 117005 | 09020 | — | — |
| Gergei | 118019 | 09055 | — | — |
| Gesico | 118020 | 09040 | — | — |
| Gesturi | 117006 | 09020 | — | — |
| Ghilarza | 115024 | 09074 | — | — |
| Giave | 112024 | 07010 | — | — |
| Giba | 119007 | 09010 | — | — |
| Girasole | 116007 | 08040 | — | — |
| Golfo Aranci | 113011 | 07020 | — | — |
| Goni | 118021 | 09040 | — | — |
| Gonnesa | 119008 | 09010 | — | — |
| Gonnoscodina | 115025 | 09090 | — | — |
| Gonnosfanadiga | 117007 | 09035 | — | — |
| Gonnosnò | 115026 | 09090 | — | — |
| Gonnostramatza | 115027 | 09093 | — | — |
| Guamaggiore | 118022 | 09040 | — | — |
| Guasila | 118023 | 09040 | — | — |
| Guspini | 117008 | 09036 | — | — |
| Iglesias | 119009 | 09016 | — | — |
| Ilbono | 116008 | 08040 | — | — |
| Illorai | 112025 | 07010 | — | — |
| Irgoli | 114017 | 08020 | — | — |
| Isili | 118024 | 09056 | — | — |
| Ittireddu | 112026 | 07010 | — | — |
| Ittiri | 112027 | 07044 | — | — |
| Jerzu | 116009 | 08044 | — | — |
| La Maddalena | 113012 | 07024 | — | — |
| Laconi | 115028 | 09090 | — | — |
| Laerru | 112028 | 07030 | — | — |
| Lanusei | 116010 | 08045 | — | — |
| Las Plassas | 117009 | 09020 | — | — |
| Lei | 114018 | 08010 | — | — |
| Loceri | 116011 | 08040 | — | — |
| Loculi | 114019 | 08020 | — | — |
| Lodè | 114020 | 08020 | — | — |
| Lodine | 114021 | 08020 | — | — |
| Loiri Porto San Paolo | 113013 | 07020 | — | — |
| Lotzorai | 116012 | 08040 | — | — |
| Lula | 114022 | 08020 | — | — |
| Lunamatrona | 117010 | 09022 | — | — |
| Luogosanto | 113014 | 07020 | — | — |
| Luras | 113015 | 07025 | — | — |
| Macomer | 114023 | 08015 | — | — |
| Magomadas | 115029 | 09090 | — | — |
| Mamoiada | 114024 | 08024 | — | — |
| Mandas | 118025 | 09040 | — | — |
| Mara | 112029 | 07010 | — | — |
| Maracalagonis | 118026 | 09069 | — | — |
| Marrubiu | 115030 | 09094 | — | — |
| Martis | 112030 | 07030 | — | — |
| Masainas | 119010 | 09010 | — | — |
| Masullas | 115031 | 09090 | — | — |
| Meana Sardo | 114025 | 08030 | — | — |
| Milis | 115032 | 09070 | — | — |
| Modolo | 115033 | 09090 | — | — |
| Mogorella | 115034 | 09080 | — | — |
| Mogoro | 115035 | 09095 | — | — |
| Monastir | 118027 | 09023 | — | — |
| Monserrato | 118028 | 09042 | — | — |
| Monteleone Rocca Doria | 112031 | 07010 | — | — |
| Monti | 113016 | 07020 | — | — |
| Montresta | 115036 | 09090 | — | — |
| Mores | 112032 | 07013 | — | — |
| Morgongiori | 115037 | 09090 | — | — |
| Muravera | 118029 | 09043 | — | — |
| Muros | 112033 | 07030 | — | — |
| Musei | 119011 | 09010 | — | — |
| Narbolia | 115038 | 09070 | — | — |
| Narcao | 119012 | 09010 | — | — |
| Neoneli | 115039 | 09080 | — | — |
| Noragugume | 114026 | 08010 | — | — |
| Norbello | 115040 | 09070 | — | — |
| Nughedu San Nicolò | 112034 | 07010 | — | — |
| Nughedu Santa Vittoria | 115041 | 09080 | — | — |
| Nule | 112035 | 07010 | — | — |
| Nulvi | 112036 | 07032 | — | — |
| Nuoro | 114027 | 08100 | — | — |
| Nurachi | 115042 | 09070 | — | — |
| Nuragus | 118030 | 09057 | — | — |
| Nurallao | 118031 | 09058 | — | — |
| Nuraminis | 118032 | 09024 | — | — |
| Nureci | 115043 | 09080 | — | — |
| Nurri | 118033 | 09059 | — | — |
| Nuxis | 119013 | 09010 | — | — |
| Olbia | 113017 | 07026 | — | — |
| Oliena | 114028 | 08025 | — | — |
| Ollastra | 115044 | 09088 | — | — |
| Ollolai | 114029 | 08020 | — | — |
| Olmedo | 112037 | 07040 | — | — |
| Olzai | 114030 | 08020 | — | — |
| Onanì | 114031 | 08020 | — | — |
| Onifai | 114032 | 08020 | — | — |
| Oniferi | 114033 | 08020 | — | — |
| Orani | 114034 | 08026 | — | — |
| Orgosolo | 114035 | 08027 | — | — |
| Oristano | 115045 | 09170 | — | — |
| Orosei | 114036 | 08028 | — | — |
| Orotelli | 114037 | 08020 | — | — |
| Orroli | 118034 | 09061 | — | — |
| Ortacesus | 118035 | 09040 | — | — |
| Ortueri | 114038 | 08036 | — | — |
| Orune | 114039 | 08020 | — | — |
| Oschiri | 113018 | 07027 | — | — |
| Osidda | 114040 | 08020 | — | — |
| Osilo | 112038 | 07033 | — | — |
| Osini | 116013 | 08040 | — | — |
| Ossi | 112039 | 07045 | — | — |
| Ottana | 114041 | 08020 | — | — |
| Ovodda | 114042 | 08020 | — | — |
| Ozieri | 112040 | 07014 | — | — |
| Pabillonis | 117011 | 09030 | — | — |
| Padria | 112041 | 07015 | — | — |
| Padru | 113019 | 07020 | — | — |
| Palau | 113020 | 07020 | — | — |
| Palmas Arborea | 115046 | 09090 | — | — |
| Pattada | 112042 | 07016 | — | — |
| Pau | 115047 | 09090 | — | — |
| Pauli Arbarei | 117012 | 09020 | — | — |
| Paulilatino | 115048 | 09070 | — | — |
| Perdasdefogu | 116014 | 08046 | — | — |
| Perdaxius | 119014 | 09010 | — | — |
| Perfugas | 112043 | 07034 | — | — |
| Pimentel | 118036 | 09020 | — | — |
| Piscinas | 119015 | 09010 | — | — |
| Ploaghe | 112044 | 07017 | — | — |
| Pompu | 115049 | 09093 | — | — |
| Porto Torres | 112045 | 07046 | — | — |
| Portoscuso | 119016 | 09010 | — | — |
| Posada | 114043 | 08020 | — | — |
| Pozzomaggiore | 112046 | 07018 | — | — |
| Pula | 118037 | 09050 | — | — |
| Putifigari | 112047 | 07040 | — | — |
| Quartu Sant'Elena | 118038 | 09045 | — | — |
| Quartucciu | 118039 | 09044 | — | — |
| Riola Sardo | 115050 | 09070 | — | — |
| Romana | 112048 | 07010 | — | — |
| Ruinas | 115051 | 09085 | — | — |
| Sadali | 118040 | 09062 | — | — |
| Sagama | 115052 | 09090 | — | — |
| Samassi | 117013 | 09030 | — | — |
| Samatzai | 118041 | 09020 | — | — |
| Samugheo | 115053 | 09086 | — | — |
| San Basilio | 118042 | 09040 | — | — |
| San Gavino Monreale | 117014 | 09037 | — | — |
| San Giovanni Suergiu | 119017 | 09010 | — | — |
| San Nicolò d'Arcidano | 115054 | 09097 | — | — |
| San Nicolò Gerrei | 118043 | 09040 | — | — |
| San Sperate | 118044 | 09026 | — | — |
| San Teodoro | 113021 | 07052 | — | — |
| San Vero Milis | 115055 | 09070 | — | — |
| San Vito | 118045 | 09040 | — | — |
| Sanluri | 117015 | 09025 | — | — |
| Sant'Andrea Frius | 118046 | 09040 | — | — |
| Sant'Anna Arresi | 119019 | 09010 | — | — |
| Sant'Antioco | 119020 | 09017 | — | — |
| Sant'Antonio di Gallura | 113023 | 07030 | — | — |
| Santa Giusta | 115056 | 09096 | — | — |
| Santa Maria Coghinas | 112049 | 07030 | — | — |
| Santa Teresa Gallura | 113022 | 07028 | — | — |
| Santadi | 119018 | 09010 | — | — |
| Santu Lussurgiu | 115057 | 09075 | — | — |
| Sardara | 117016 | 09030 | — | — |
| Sarroch | 118047 | 09018 | — | — |
| Sarule | 114044 | 08020 | — | — |
| Sassari | 112050 | 07100 | — | — |
| Scano di Montiferro | 115058 | 09078 | — | — |
| Sedilo | 115059 | 09076 | — | — |
| Sedini | 112051 | 07035 | — | — |
| Segariu | 117017 | 09040 | — | — |
| Selargius | 118048 | 09047 | — | — |
| Selegas | 118049 | 09040 | — | — |
| Semestene | 112052 | 07010 | — | — |
| Seneghe | 115060 | 09070 | — | — |
| Senis | 115061 | 09080 | — | — |
| Sennariolo | 115062 | 09078 | — | — |
| Sennori | 112053 | 07036 | — | — |
| Senorbì | 118050 | 09040 | — | — |
| Serdiana | 118051 | 09040 | — | — |
| Serramanna | 117018 | 09038 | — | — |
| Serrenti | 117019 | 09027 | — | — |
| Serri | 118052 | 09063 | — | — |
| Sestu | 118053 | 09028 | — | — |
| Settimo San Pietro | 118054 | 09060 | — | — |
| Setzu | 117020 | 09029 | — | — |
| Seui | 116023 | 09064 | — | — |
| Seulo | 114045 | 09065 | — | — |
| Siamaggiore | 115063 | 09070 | — | — |
| Siamanna | 115064 | 09080 | — | — |
| Siapiccia | 115065 | 09080 | — | — |
| Siddi | 117021 | 09020 | — | — |
| Silanus | 114046 | 08017 | — | — |
| Siligo | 112054 | 07040 | — | — |
| Siliqua | 118056 | 09010 | — | — |
| Silius | 118057 | 09040 | — | — |
| Simala | 115066 | 09090 | — | — |
| Simaxis | 115067 | 09088 | — | — |
| Sindia | 114047 | 08018 | — | — |
| Sini | 115068 | 09090 | — | — |
| Siniscola | 114048 | 08029 | — | — |
| Sinnai | 118058 | 09048 | — | — |
| Siris | 115069 | 09090 | — | — |
| Siurgus Donigala | 118059 | 09040 | — | — |
| Soddì | 115070 | 09080 | — | — |
| Solarussa | 115071 | 09077 | — | — |
| Soleminis | 118060 | 09040 | — | — |
| Sorgono | 114049 | 08038 | — | — |
| Sorradile | 115072 | 09080 | — | — |
| Sorso | 112055 | 07037 | — | — |
| Stintino | 112056 | 07040 | — | — |
| Suelli | 118061 | 09040 | — | — |
| Suni | 115073 | 09090 | — | — |
| Tadasuni | 115074 | 09080 | — | — |
| Talana | 116015 | 08040 | — | — |
| Telti | 113024 | 07020 | — | — |
| Tempio Pausania | 113025 | 07029 | — | — |
| Tergu | 112057 | 07030 | — | — |
| Terralba | 115075 | 09098 | — | — |
| Tertenia | 116016 | 08047 | — | — |
| Teti | 114050 | 08030 | — | — |
| Teulada | 119024 | 09019 | — | — |
| Thiesi | 112058 | 07047 | — | — |
| Tiana | 114051 | 08020 | — | — |
| Tinnura | 115076 | 09090 | — | — |
| Tissi | 112059 | 07040 | — | — |
| Tonara | 114052 | 08039 | — | — |
| Torpè | 114053 | 08020 | — | — |
| Torralba | 112060 | 07048 | — | — |
| Tortolì | 116017 | 08048 | — | — |
| Tramatza | 115077 | 09070 | — | — |
| Tratalias | 119021 | 09010 | — | — |
| Tresnuraghes | 115078 | 09079 | — | — |
| Triei | 116018 | 08040 | — | — |
| Trinità d'Agultu e Vignola | 113026 | 07038 | — | — |
| Tuili | 117022 | 09029 | — | — |
| Tula | 112061 | 07010 | — | — |
| Turri | 117023 | 09020 | — | — |
| Ulà Tirso | 115079 | 09080 | — | — |
| Ulassai | 116019 | 08040 | — | — |
| Uras | 115080 | 09099 | — | — |
| Uri | 112062 | 07040 | — | — |
| Urzulei | 116020 | 08040 | — | — |
| Usellus | 115081 | 09090 | — | — |
| Usini | 112063 | 07049 | — | — |
| Ussana | 118063 | 09020 | — | — |
| Ussaramanna | 117024 | 09020 | — | — |
| Ussassai | 116021 | 08040 | — | — |
| Uta | 118064 | 09068 | — | — |
| Valledoria | 112064 | 07039 | — | — |
| Vallermosa | 118065 | 09010 | — | — |
| Viddalba | 112065 | 07030 | — | — |
| Villa San Pietro | 118066 | 09050 | — | — |
| Villa Sant'Antonio | 115082 | 09080 | — | — |
| Villa Verde | 115083 | 09090 | — | — |
| Villacidro | 117025 | 09039 | — | — |
| Villagrande Strisaili | 116022 | 08049 | — | — |
| Villamar | 117026 | 09020 | — | — |
| Villamassargia | 119022 | 09010 | — | — |
| Villanova Monteleone | 112066 | 07019 | — | — |
| Villanova Truschedu | 115084 | 09084 | — | — |
| Villanova Tulo | 118067 | 09066 | — | — |
| Villanovaforru | 117027 | 09020 | — | — |
| Villanovafranca | 117028 | 09020 | — | — |
| Villaperuccio | 119023 | 09010 | — | — |
| Villaputzu | 118068 | 09040 | — | — |
| Villasalto | 118069 | 09040 | — | — |
| Villasimius | 118070 | 09049 | — | — |
| Villasor | 118071 | 09034 | — | — |
| Villaspeciosa | 118072 | 09010 | — | — |
| Villaurbana | 115085 | 09080 | — | — |
| Zeddiani | 115086 | 09070 | — | — |
| Zerfaliu | 115087 | 09070 | — | — |

<a id="region-19"></a>

### Sicilia (391 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Sicilia — region** | 19 | — | Not started | Regional sources not checked |
| Acate | 088001 | 97011 | — | — |
| Aci Bonaccorsi | 087001 | 95020 | — | — |
| Aci Castello | 087002 | 95021 | — | — |
| Aci Catena | 087003 | 95022 | — | — |
| Aci Sant'Antonio | 087005 | 95025 | — | — |
| Acireale | 087004 | 95024 | — | — |
| Acquaviva Platani | 085001 | 93010 | — | — |
| Acquedolci | 083107 | 98070 | — | — |
| Adrano | 087006 | 95031 | — | — |
| Agira | 086001 | 94011 | — | — |
| Agrigento | 084001 | 92100 | — | — |
| Aidone | 086002 | 94020 | — | — |
| Alcamo | 081001 | 91011 | — | — |
| Alcara li Fusi | 083001 | 98070 | — | — |
| Alessandria della Rocca | 084002 | 92010 | — | — |
| Alì | 083002 | 98020 | — | — |
| Alì Terme | 083003 | 98021 | — | — |
| Alia | 082001 | 90021 | — | — |
| Alimena | 082002 | 90020 | — | — |
| Aliminusa | 082003 | 90020 | — | — |
| Altavilla Milicia | 082004 | 90010 | — | — |
| Altofonte | 082005 | 90072 | — | — |
| Antillo | 083004 | 98030 | — | — |
| Aragona | 084003 | 92021 | — | — |
| Assoro | 086003 | 94010 | — | — |
| Augusta | 089001 | 96011 | — | — |
| Avola | 089002 | 96012 | — | — |
| Bagheria | 082006 | 90011 | — | — |
| Balestrate | 082007 | 90041 | — | — |
| Barcellona Pozzo di Gotto | 083005 | 98051 | — | — |
| Barrafranca | 086004 | 94012 | — | — |
| Basicò | 083006 | 98060 | — | — |
| Baucina | 082008 | 90061 | — | — |
| Belmonte Mezzagno | 082009 | 90031 | — | — |
| Belpasso | 087007 | 95032 | — | — |
| Biancavilla | 087008 | 95033 | — | — |
| Bisacquino | 082010 | 90032 | — | — |
| Bivona | 084004 | 92010 | — | — |
| Blufi | 082082 | 90020 | — | — |
| Bolognetta | 082011 | 90030 | — | — |
| Bompensiere | 085002 | 93010 | — | — |
| Bompietro | 082012 | 90020 | — | — |
| Borgetto | 082013 | 90042 | — | — |
| Brolo | 083007 | 98061 | — | — |
| Bronte | 087009 | 95034 | — | — |
| Buccheri | 089003 | 96010 | — | — |
| Burgio | 084005 | 92010 | — | — |
| Buscemi | 089004 | 96010 | — | — |
| Buseto Palizzolo | 081002 | 91012 | — | — |
| Butera | 085003 | 93011 | — | — |
| Caccamo | 082014 | 90012 | — | — |
| Calamonaci | 084006 | 92010 | — | — |
| Calascibetta | 086005 | 94021 | — | — |
| Calatabiano | 087010 | 95011 | — | — |
| Calatafimi-Segesta | 081003 | 91013 | — | — |
| Caltabellotta | 084007 | 92010 | — | — |
| Caltagirone | 087011 | 95041 | — | — |
| Caltanissetta | 085004 | 93100 | — | — |
| Caltavuturo | 082015 | 90022 | — | — |
| Camastra | 084008 | 92020 | — | — |
| Cammarata | 084009 | 92022 | — | — |
| Campobello di Licata | 084010 | 92023 | — | — |
| Campobello di Mazara | 081004 | 91021 | — | — |
| Campofelice di Fitalia | 082016 | 90030 | — | — |
| Campofelice di Roccella | 082017 | 90010 | — | — |
| Campofiorito | 082018 | 90030 | — | — |
| Campofranco | 085005 | 93010 | — | — |
| Camporeale | 082019 | 90043 | — | — |
| Camporotondo Etneo | 087012 | 95058 | — | — |
| Canicattì | 084011 | 92024 | — | — |
| Canicattini Bagni | 089005 | 96010 | — | — |
| Capaci | 082020 | 90040 | — | — |
| Capizzi | 083008 | 98031 | — | — |
| Capo d'Orlando | 083009 | 98071 | — | — |
| Capri Leone | 083010 | 98070 | — | — |
| Carini | 082021 | 90044 | — | — |
| Carlentini | 089006 | 96013 | — | — |
| Caronia | 083011 | 98072 | — | — |
| Casalvecchio Siculo | 083012 | 98032 | — | — |
| Cassaro | 089007 | 96010 | — | — |
| Castel di Iudica | 087013 | 95040 | — | — |
| Castel di Lucio | 083013 | 98070 | — | — |
| Castelbuono | 082022 | 90013 | — | — |
| Casteldaccia | 082023 | 90014 | — | — |
| Castell'Umberto | 083014 | 98070 | — | — |
| Castellammare del Golfo | 081005 | 91014 | — | — |
| Castellana Sicula | 082024 | 90020 | — | — |
| Castelmola | 083015 | 98030 | — | — |
| Casteltermini | 084012 | 92025 | — | — |
| Castelvetrano | 081006 | 91022 | — | — |
| Castiglione di Sicilia | 087014 | 95012 | — | — |
| Castrofilippo | 084013 | 92020 | — | — |
| Castronovo di Sicilia | 082025 | 90030 | — | — |
| Castroreale | 083016 | 98053 | — | — |
| Catania | 087015 | 95121, 95122, 95123, 95124, 95125, 95126, 95127, 95128, 95129, 95131 | — | — |
| Catenanuova | 086006 | 94010 | — | — |
| Cattolica Eraclea | 084014 | 92011 | — | — |
| Cefalà Diana | 082026 | 90030 | — | — |
| Cefalù | 082027 | 90015 | — | — |
| Centuripe | 086007 | 94010 | — | — |
| Cerami | 086008 | 94010 | — | — |
| Cerda | 082028 | 90052 | — | — |
| Cesarò | 083017 | 98033 | — | — |
| Chiaramonte Gulfi | 088002 | 97012 | — | — |
| Chiusa Sclafani | 082029 | 90033 | — | — |
| Cianciana | 084015 | 92012 | — | — |
| Ciminna | 082030 | 90023 | — | — |
| Cinisi | 082031 | 90045 | — | — |
| Collesano | 082032 | 90016 | — | — |
| Comiso | 088003 | 97013 | — | — |
| Comitini | 084016 | 92020 | — | — |
| Condrò | 083018 | 98040 | — | — |
| Contessa Entellina | 082033 | 90030 | — | — |
| Corleone | 082034 | 90034 | — | — |
| Custonaci | 081007 | 91015 | — | — |
| Delia | 085006 | 93010 | — | — |
| Enna | 086009 | 94100 | — | — |
| Erice | 081008 | 91016 | — | — |
| Falcone | 083019 | 98060 | — | — |
| Favara | 084017 | 92026 | — | — |
| Favignana | 081009 | 91023 | — | — |
| Ferla | 089008 | 96010 | — | — |
| Ficarazzi | 082035 | 90010 | — | — |
| Ficarra | 083020 | 98062 | — | — |
| Fiumedinisi | 083021 | 98022 | — | — |
| Fiumefreddo di Sicilia | 087016 | 95013 | — | — |
| Floresta | 083022 | 98030 | — | — |
| Floridia | 089009 | 96014 | — | — |
| Fondachelli-Fantina | 083023 | 98050 | — | — |
| Forza d'Agrò | 083024 | 98030 | — | — |
| Francavilla di Sicilia | 083025 | 98034 | — | — |
| Francofonte | 089010 | 96015 | — | — |
| Frazzanò | 083026 | 98070 | — | — |
| Furci Siculo | 083027 | 98023 | — | — |
| Furnari | 083028 | 98054 | — | — |
| Gaggi | 083029 | 98030 | — | — |
| Gagliano Castelferrato | 086010 | 94010 | — | — |
| Galati Mamertino | 083030 | 98070 | — | — |
| Gallodoro | 083031 | 98030 | — | — |
| Gangi | 082036 | 90024 | — | — |
| Gela | 085007 | 93012 | — | — |
| Geraci Siculo | 082037 | 90054 | — | — |
| Giardinello | 082038 | 90084 | — | — |
| Giardini-Naxos | 083032 | 98035 | — | — |
| Giarratana | 088004 | 97010 | — | — |
| Giarre | 087017 | 95014 | — | — |
| Gibellina | 081010 | 91024 | — | — |
| Gioiosa Marea | 083033 | 98063 | — | — |
| Giuliana | 082039 | 90030 | — | — |
| Godrano | 082040 | 90030 | — | — |
| Grammichele | 087018 | 95042 | — | — |
| Graniti | 083034 | 98036 | — | — |
| Gratteri | 082041 | 90010 | — | — |
| Gravina di Catania | 087019 | 95030 | — | — |
| Grotte | 084018 | 92020 | — | — |
| Gualtieri Sicaminò | 083035 | 98040 | — | — |
| Isnello | 082042 | 90010 | — | — |
| Isola delle Femmine | 082043 | 90040 | — | — |
| Ispica | 088005 | 97014 | — | — |
| Itala | 083036 | 98025 | — | — |
| Joppolo Giancaxio | 084019 | 92035 | — | — |
| Lampedusa e Linosa | 084020 | 92031 | — | — |
| Lascari | 082044 | 90010 | — | — |
| Leni | 083037 | 98050 | — | — |
| Lentini | 089011 | 96016 | — | — |
| Leonforte | 086011 | 94013 | — | — |
| Lercara Friddi | 082045 | 90025 | — | — |
| Letojanni | 083038 | 98037 | — | — |
| Librizzi | 083039 | 98064 | — | — |
| Licata | 084021 | 92027 | — | — |
| Licodia Eubea | 087020 | 95059 | — | — |
| Limina | 083040 | 98030 | — | — |
| Linguaglossa | 087021 | 95015 | — | — |
| Lipari | 083041 | 98055 | — | — |
| Longi | 083042 | 98070 | — | — |
| Lucca Sicula | 084022 | 92010 | — | — |
| Maletto | 087022 | 95035 | — | — |
| Malfa | 083043 | 98050 | — | — |
| Malvagna | 083044 | 98030 | — | — |
| Mandanici | 083045 | 98020 | — | — |
| Maniace | 087057 | 95050 | — | — |
| Marianopoli | 085008 | 93010 | — | — |
| Marineo | 082046 | 90035 | — | — |
| Marsala | 081011 | 91025 | — | — |
| Mascali | 087023 | 95016 | — | — |
| Mascalucia | 087024 | 95030 | — | — |
| Mazara del Vallo | 081012 | 91026 | — | — |
| Mazzarino | 085009 | 93013 | — | — |
| Mazzarrà Sant'Andrea | 083046 | 98056 | — | — |
| Mazzarrone | 087056 | 95040 | — | — |
| Melilli | 089012 | 96010 | — | — |
| Menfi | 084023 | 92013 | — | — |
| Merì | 083047 | 98040 | — | — |
| Messina | 083048 | 98121, 98122, 98123, 98124, 98125, 98126, 98127, 98128, 98129, 98131, 98132, 98133, 98134, 98135, 98136, 98137, 98138, 98139, 98141, 98142, 98143, 98144, 98145, 98146, 98147, 98148, 98149, 98151, 98152, 98153, 98154, 98155, 98156, 98157, 98158, 98161, 98162, 98163, 98164, 98165, 98166, 98167, 98168 | — | — |
| Mezzojuso | 082047 | 90030 | — | — |
| Milazzo | 083049 | 98057 | — | — |
| Milena | 085010 | 93010 | — | — |
| Militello in Val di Catania | 087025 | 95043 | — | — |
| Militello Rosmarino | 083050 | 98070 | — | — |
| Milo | 087026 | 95010 | — | — |
| Mineo | 087027 | 95044 | — | — |
| Mirabella Imbaccari | 087028 | 95040 | — | — |
| Mirto | 083051 | 98070 | — | — |
| Misiliscemi | 081025 | 91031 | — | — |
| Misilmeri | 082048 | 90036 | — | — |
| Misterbianco | 087029 | 95045 | — | — |
| Mistretta | 083052 | 98073 | — | — |
| Modica | 088006 | 97015 | — | — |
| Moio Alcantara | 083053 | 98030 | — | — |
| Monforte San Giorgio | 083054 | 98041 | — | — |
| Mongiuffi Melia | 083055 | 98030 | — | — |
| Monreale | 082049 | 90046 | — | — |
| Montagnareale | 083056 | 98060 | — | — |
| Montalbano Elicona | 083057 | 98065 | — | — |
| Montallegro | 084024 | 92010 | — | — |
| Montedoro | 085011 | 93010 | — | — |
| Montelepre | 082050 | 90086 | — | — |
| Montemaggiore Belsito | 082051 | 90020 | — | — |
| Monterosso Almo | 088007 | 97010 | — | — |
| Montevago | 084025 | 92038 | — | — |
| Motta Camastra | 083058 | 98030 | — | — |
| Motta d'Affermo | 083059 | 98070 | — | — |
| Motta Sant'Anastasia | 087030 | 95062 | — | — |
| Mussomeli | 085012 | 93014 | — | — |
| Naro | 084026 | 92028 | — | — |
| Naso | 083060 | 98074 | — | — |
| Nicolosi | 087031 | 95052 | — | — |
| Nicosia | 086012 | 94014 | — | — |
| Niscemi | 085013 | 93015 | — | — |
| Nissoria | 086013 | 94010 | — | — |
| Nizza di Sicilia | 083061 | 98026 | — | — |
| Noto | 089013 | 96017 | — | — |
| Novara di Sicilia | 083062 | 98058 | — | — |
| Oliveri | 083063 | 98060 | — | — |
| Pace del Mela | 083064 | 98042 | — | — |
| Paceco | 081013 | 91027 | — | — |
| Pachino | 089014 | 96018 | — | — |
| Pagliara | 083065 | 98020 | — | — |
| Palagonia | 087032 | 95046 | — | — |
| Palazzo Adriano | 082052 | 90030 | — | — |
| Palazzolo Acreide | 089015 | 96010 | — | — |
| Palermo | 082053 | 90121, 90123, 90124, 90125, 90126, 90127, 90128, 90129, 90131, 90133, 90134, 90135, 90136, 90138, 90139, 90141, 90142, 90143, 90144, 90145, 90146, 90147, 90149, 90151 | — | — |
| Palma di Montechiaro | 084027 | 92044 | — | — |
| Pantelleria | 081014 | 91017 | — | — |
| Partanna | 081015 | 91028 | — | — |
| Partinico | 082054 | 90047 | — | — |
| Paternò | 087033 | 95047 | — | — |
| Patti | 083066 | 98066 | — | — |
| Pedara | 087034 | 95053 | — | — |
| Petralia Soprana | 082055 | 90026 | — | — |
| Petralia Sottana | 082056 | 90027 | — | — |
| Petrosino | 081024 | 91032 | — | — |
| Pettineo | 083067 | 98070 | — | — |
| Piana degli Albanesi | 082057 | 90037 | — | — |
| Piazza Armerina | 086014 | 94015 | — | — |
| Piedimonte Etneo | 087035 | 95017 | — | — |
| Pietraperzia | 086015 | 94016 | — | — |
| Piraino | 083068 | 98060 | — | — |
| Poggioreale | 081016 | 91020 | — | — |
| Polizzi Generosa | 082058 | 90028 | — | — |
| Pollina | 082059 | 90010 | — | — |
| Porto Empedocle | 084028 | 92014 | — | — |
| Portopalo di Capo Passero | 089020 | 96026 | — | — |
| Pozzallo | 088008 | 97016 | — | — |
| Priolo Gargallo | 089021 | 96010 | — | — |
| Prizzi | 082060 | 90038 | — | — |
| Racalmuto | 084029 | 92020 | — | — |
| Raccuja | 083069 | 98067 | — | — |
| Raddusa | 087036 | 95040 | — | — |
| Raffadali | 084030 | 92015 | — | — |
| Ragalna | 087058 | 95054 | — | — |
| Ragusa | 088009 | 97100 | — | — |
| Ramacca | 087037 | 95040 | — | — |
| Randazzo | 087038 | 95036 | — | — |
| Ravanusa | 084031 | 92029 | — | — |
| Realmonte | 084032 | 92010 | — | — |
| Regalbuto | 086016 | 94017 | — | — |
| Reitano | 083070 | 98070 | — | — |
| Resuttano | 085014 | 93010 | — | — |
| Ribera | 084033 | 92016 | — | — |
| Riesi | 085015 | 93016 | — | — |
| Riposto | 087039 | 95018 | — | — |
| Roccafiorita | 083071 | 98030 | — | — |
| Roccalumera | 083072 | 98027 | — | — |
| Roccamena | 082061 | 90040 | — | — |
| Roccapalumba | 082062 | 90066 | — | — |
| Roccavaldina | 083073 | 98040 | — | — |
| Roccella Valdemone | 083074 | 98030 | — | — |
| Rodì Milici | 083075 | 98059 | — | — |
| Rometta | 083076 | 98043 | — | — |
| Rosolini | 089016 | 96019 | — | — |
| Salaparuta | 081017 | 91020 | — | — |
| Salemi | 081018 | 91018 | — | — |
| Sambuca di Sicilia | 084034 | 92017 | — | — |
| San Biagio Platani | 084035 | 92020 | — | — |
| San Cataldo | 085016 | 93017 | — | — |
| San Cipirello | 082063 | 90088 | — | — |
| San Cono | 087040 | 95040 | — | — |
| San Filippo del Mela | 083077 | 98044 | — | — |
| San Fratello | 083078 | 98075 | — | — |
| San Giovanni Gemini | 084036 | 92020 | — | — |
| San Giovanni la Punta | 087041 | 95037 | — | — |
| San Giuseppe Jato | 082064 | 90048 | — | — |
| San Gregorio di Catania | 087042 | 95027 | — | — |
| San Marco d'Alunzio | 083079 | 98070 | — | — |
| San Mauro Castelverde | 082065 | 90010 | — | — |
| San Michele di Ganzaria | 087043 | 95040 | — | — |
| San Pier Niceto | 083080 | 98045 | — | — |
| San Piero Patti | 083081 | 98068 | — | — |
| San Pietro Clarenza | 087044 | 95055 | — | — |
| San Salvatore di Fitalia | 083082 | 98070 | — | — |
| San Teodoro | 083090 | 98030 | — | — |
| San Vito Lo Capo | 081020 | 91030 | — | — |
| Sant'Agata di Militello | 083084 | 98076 | — | — |
| Sant'Agata li Battiati | 087045 | 95056 | — | — |
| Sant'Alessio Siculo | 083085 | 98030 | — | — |
| Sant'Alfio | 087046 | 95010 | — | — |
| Sant'Angelo di Brolo | 083088 | 98060 | — | — |
| Sant'Angelo Muxaro | 084039 | 92020 | — | — |
| Santa Caterina Villarmosa | 085017 | 93018 | — | — |
| Santa Cristina Gela | 082066 | 90082 | — | — |
| Santa Croce Camerina | 088010 | 97017 | — | — |
| Santa Domenica Vittoria | 083083 | 98030 | — | — |
| Santa Elisabetta | 084037 | 92020 | — | — |
| Santa Flavia | 082067 | 90017 | — | — |
| Santa Lucia del Mela | 083086 | 98046 | — | — |
| Santa Margherita di Belice | 084038 | 92018 | — | — |
| Santa Maria di Licodia | 087047 | 95038 | — | — |
| Santa Marina Salina | 083087 | 98050 | — | — |
| Santa Ninfa | 081019 | 91029 | — | — |
| Santa Teresa di Riva | 083089 | 98028 | — | — |
| Santa Venerina | 087048 | 95010 | — | — |
| Santo Stefano di Camastra | 083091 | 98077 | — | — |
| Santo Stefano Quisquina | 084040 | 92020 | — | — |
| Saponara | 083092 | 98047 | — | — |
| Savoca | 083093 | 98038 | — | — |
| Scaletta Zanclea | 083094 | 98029 | — | — |
| Sciacca | 084041 | 92019 | — | — |
| Sciara | 082068 | 90020 | — | — |
| Scicli | 088011 | 97018 | — | — |
| Scillato | 082081 | 90020 | — | — |
| Sclafani Bagni | 082069 | 90020 | — | — |
| Scordia | 087049 | 95048 | — | — |
| Serradifalco | 085018 | 93010 | — | — |
| Siculiana | 084042 | 92010 | — | — |
| Sinagra | 083095 | 98069 | — | — |
| Siracusa | 089017 | 96100 | — | — |
| Solarino | 089018 | 96010 | — | — |
| Sommatino | 085019 | 93019 | — | — |
| Sortino | 089019 | 96010 | — | — |
| Spadafora | 083096 | 98048 | — | — |
| Sperlinga | 086017 | 94010 | — | — |
| Sutera | 085020 | 93010 | — | — |
| Taormina | 083097 | 98039 | — | — |
| Terme Vigliatore | 083106 | 98050 | — | — |
| Termini Imerese | 082070 | 90018 | — | — |
| Terrasini | 082071 | 90049 | — | — |
| Torregrotta | 083098 | 98040 | — | — |
| Torrenova | 083108 | 98070 | — | — |
| Torretta | 082072 | 90040 | — | — |
| Tortorici | 083099 | 98078 | — | — |
| Trabia | 082073 | 90019 | — | — |
| Trapani | 081021 | 91100 | — | — |
| Trappeto | 082074 | 90090 | — | — |
| Trecastagni | 087050 | 95039 | — | — |
| Tremestieri Etneo | 087051 | 95030 | — | — |
| Tripi - Abakainon | 083100 | 98060 | — | — |
| Troina | 086018 | 94018 | — | — |
| Tusa | 083101 | 98079 | — | — |
| Ucria | 083102 | 98060 | — | — |
| Ustica | 082075 | 90051 | — | — |
| Valderice | 081022 | 91019 | — | — |
| Valdina | 083103 | 98040 | — | — |
| Valguarnera Caropepe | 086019 | 94019 | — | — |
| Valledolmo | 082076 | 90029 | — | — |
| Vallelunga Pratameno | 085021 | 93010 | — | — |
| Valverde | 087052 | 95028 | — | — |
| Venetico | 083104 | 98040 | — | — |
| Ventimiglia di Sicilia | 082077 | 90070 | — | — |
| Viagrande | 087053 | 95029 | — | — |
| Vicari | 082078 | 90071 | — | — |
| Villabate | 082079 | 90039 | — | — |
| Villafranca Sicula | 084043 | 92020 | — | — |
| Villafranca Tirrena | 083105 | 98049 | — | — |
| Villafrati | 082080 | 90030 | — | — |
| Villalba | 085022 | 93010 | — | — |
| Villarosa | 086020 | 94028 | — | — |
| Vita | 081023 | 91010 | — | — |
| Vittoria | 088012 | 97019 | — | — |
| Vizzini | 087054 | 95049 | — | — |
| Zafferana Etnea | 087055 | 95019 | — | — |

<a id="region-09"></a>

### Toscana (273 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **[Toscana — region](territori/09-toscana/README.md)** | 09 | — | Regional sources in internal pilot; public off | Observation and formal acceptance pending |
| Abbadia San Salvatore | 052001 | 53021 | — | — |
| Abetone Cutigliano | 047023 | 51024 | — | — |
| Agliana | 047002 | 51031 | — | — |
| Altopascio | 046001 | 55011 | — | — |
| Anghiari | 051001 | 52031 | — | — |
| Arcidosso | 053001 | 58031 | — | — |
| Arezzo | 051002 | 52100 | — | — |
| Asciano | 052002 | 53041 | — | — |
| Aulla | 045001 | 54011 | — | — |
| Badia Tedalda | 051003 | 52032 | — | — |
| Bagni di Lucca | 046002 | 55022 | — | — |
| Bagno a Ripoli | 048001 | 50012 | — | — |
| Bagnone | 045002 | 54021 | — | — |
| Barberino di Mugello | 048002 | 50031 | — | — |
| Barberino Tavarnelle | 048054 | 50028 | — | — |
| Barga | 046003 | 55051 | — | — |
| Bibbiena | 051004 | 52011 | — | — |
| Bibbona | 049001 | 57020 | — | — |
| Bientina | 050001 | 56031 | — | Associated service identified; municipal source not verified |
| Borgo a Mozzano | 046004 | 55023 | — | — |
| Borgo San Lorenzo | 048004 | 50032 | — | — |
| Bucine | 051005 | 52021 | — | — |
| Buggiano | 047003 | 51011 | — | — |
| Buonconvento | 052003 | 53022 | — | — |
| Buti | 050002 | 56032 | — | — |
| Calci | 050003 | 56011 | — | — |
| [Calcinaia](territori/09-toscana/comuni/050004-calcinaia.md) | 050004 | 56012 | Internal pilot; public off | Trial and acceptance pending |
| Calenzano | 048005 | 50041 | — | — |
| Camaiore | 046005 | 55041 | — | — |
| Campagnatico | 053002 | 58042 | — | — |
| Campi Bisenzio | 048006 | 50013 | — | — |
| Campiglia Marittima | 049002 | 57021 | — | — |
| Campo nell'Elba | 049003 | 57034 | — | — |
| Camporgiano | 046006 | 55031 | — | — |
| Cantagallo | 100001 | 59025 | — | — |
| Capalbio | 053003 | 58011 | — | — |
| Capannoli | 050005 | 56033 | — | Associated service identified; municipal source not verified |
| Capannori | 046007 | 55012 | — | — |
| Capoliveri | 049004 | 57031 | — | — |
| Capolona | 051006 | 52010 | — | — |
| Capraia e Limite | 048008 | 50050 | — | — |
| Capraia Isola | 049005 | 57032 | — | — |
| Caprese Michelangelo | 051007 | 52033 | — | — |
| Careggine | 046008 | 55030 | — | — |
| Carmignano | 100002 | 59015 | — | — |
| Carrara | 045003 | 54033 | — | — |
| Casale Marittimo | 050006 | 56040 | — | — |
| Casciana Terme Lari | 050040 | 56035 | — | Associated service identified; municipal source not verified |
| [Cascina](territori/09-toscana/comuni/050008-cascina.md) | 050008 | 56021 | Candidate configured; public off | Preview passed; trial and acceptance pending |
| Casola in Lunigiana | 045004 | 54014 | — | — |
| Casole d'Elsa | 052004 | 53031 | — | — |
| Castagneto Carducci | 049006 | 57022 | — | — |
| Castel del Piano | 053004 | 58033 | — | — |
| Castel Focognano | 051008 | 52016 | — | — |
| Castel San Niccolò | 051010 | 52018 | — | — |
| Castelfiorentino | 048010 | 50051 | — | — |
| Castelfranco di Sotto | 050009 | 56022 | — | — |
| Castelfranco Piandiscò | 051040 | 52026 | — | — |
| Castell'Azzara | 053005 | 58034 | — | — |
| Castellina in Chianti | 052005 | 53011 | — | — |
| Castellina Marittima | 050010 | 56040 | — | — |
| Castelnuovo Berardenga | 052006 | 53019 | — | — |
| Castelnuovo di Garfagnana | 046009 | 55032 | — | — |
| Castelnuovo di Val di Cecina | 050011 | 56041 | — | — |
| Castiglion Fibocchi | 051011 | 52029 | — | — |
| Castiglion Fiorentino | 051012 | 52043 | — | — |
| Castiglione d'Orcia | 052007 | 53023 | — | — |
| Castiglione della Pescaia | 053006 | 58043 | — | — |
| Castiglione di Garfagnana | 046010 | 55033 | — | — |
| Cavriglia | 051013 | 52022 | — | — |
| Cecina | 049007 | 57023 | — | — |
| Cerreto Guidi | 048011 | 50050 | — | — |
| Certaldo | 048012 | 50052 | — | — |
| Cetona | 052008 | 53040 | — | — |
| Chianciano Terme | 052009 | 53042 | — | — |
| Chianni | 050012 | 56034 | — | — |
| Chiesina Uzzanese | 047022 | 51013 | — | — |
| Chitignano | 051014 | 52010 | — | — |
| Chiusdino | 052010 | 53012 | — | — |
| Chiusi | 052011 | 53043 | — | — |
| Chiusi della Verna | 051015 | 52010 | — | — |
| Cinigiano | 053007 | 58044 | — | — |
| Civitella in Val di Chiana | 051016 | 52041 | — | — |
| Civitella Paganico | 053008 | 58045 | — | — |
| Colle di Val d'Elsa | 052012 | 53034 | — | — |
| Collesalvetti | 049008 | 57014 | — | — |
| Comano | 045005 | 54015 | — | — |
| Coreglia Antelminelli | 046011 | 55025 | — | — |
| Cortona | 051017 | 52044 | — | — |
| Crespina Lorenzana | 050041 | 56042 | — | Associated service identified; municipal source not verified |
| Dicomano | 048013 | 50062 | — | — |
| Empoli | 048014 | 50053 | — | — |
| Fabbriche di Vergemoli | 046036 | 55021 | — | — |
| Fauglia | 050014 | 56043 | — | Associated service identified; municipal source not verified |
| Fiesole | 048015 | 50014 | — | — |
| Figline e Incisa Valdarno | 048052 | 50063 | — | — |
| Filattiera | 045006 | 54023 | — | — |
| Firenze | 048017 | 50121, 50122, 50123, 50124, 50125, 50126, 50127, 50129, 50131, 50132, 50133, 50134, 50135, 50136, 50137, 50139, 50141, 50142, 50143, 50144, 50145 | — | Partial source research |
| Firenzuola | 048018 | 50033 | — | — |
| Fivizzano | 045007 | 54013 | — | — |
| Foiano della Chiana | 051018 | 52045 | — | — |
| Follonica | 053009 | 58022 | — | — |
| Forte dei Marmi | 046013 | 55042 | — | — |
| Fosciandora | 046014 | 55020 | — | — |
| Fosdinovo | 045008 | 54035 | — | — |
| Fucecchio | 048019 | 50054 | — | — |
| Gaiole in Chianti | 052013 | 53013 | — | — |
| Gallicano | 046015 | 55027 | — | — |
| Gambassi Terme | 048020 | 50050 | — | — |
| Gavorrano | 053010 | 58023 | — | — |
| Greve in Chianti | 048021 | 50022 | — | — |
| Grosseto | 053011 | 58100 | — | — |
| Guardistallo | 050015 | 56040 | — | — |
| Impruneta | 048022 | 50023 | — | — |
| Isola del Giglio | 053012 | 58012 | — | — |
| Lajatico | 050016 | 56030 | — | — |
| Lamporecchio | 047005 | 51035 | — | — |
| Larciano | 047006 | 51036 | — | — |
| Lastra a Signa | 048024 | 50055 | — | — |
| Laterina Pergine Valdarno | 051042 | 52019 | — | — |
| Licciana Nardi | 045009 | 54016 | — | — |
| [Livorno](territori/09-toscana/comuni/049009-livorno.md) | 049009 | 57121, 57122, 57123, 57124, 57125, 57126, 57127, 57128 | Candidate configured; public off | Preview unresolved; trial and acceptance pending |
| Londa | 048025 | 50060 | — | — |
| Loro Ciuffenna | 051020 | 52024 | — | — |
| Lucca | 046017 | 55100 | — | — |
| Lucignano | 051021 | 52046 | — | — |
| Magliano in Toscana | 053013 | 58051 | — | — |
| Manciano | 053014 | 58014 | — | — |
| Marciana | 049010 | 57030 | — | — |
| Marciana Marina | 049011 | 57033 | — | — |
| Marciano della Chiana | 051022 | 52047 | — | — |
| Marliana | 047007 | 51010 | — | — |
| Marradi | 048026 | 50034 | — | — |
| Massa | 045010 | 54100 | — | — |
| Massa e Cozzile | 047008 | 51010 | — | — |
| Massa Marittima | 053015 | 58024 | — | — |
| Massarosa | 046018 | 55054 | — | — |
| Minucciano | 046019 | 55034 | — | — |
| Molazzana | 046020 | 55020 | — | — |
| Monsummano Terme | 047009 | 51015 | — | — |
| Montaione | 048027 | 50050 | — | — |
| Montalcino | 052037 | 53024 | — | — |
| Montale | 047010 | 51037 | — | — |
| Monte Argentario | 053016 | 58019 | — | — |
| Monte San Savino | 051025 | 52048 | — | — |
| Montecarlo | 046021 | 55015 | — | — |
| Montecatini Val di Cecina | 050019 | 56040 | — | — |
| Montecatini-Terme | 047011 | 51016 | — | — |
| Montelupo Fiorentino | 048028 | 50056 | — | — |
| Montemignaio | 051023 | 52010 | — | — |
| Montemurlo | 100003 | 59013 | — | — |
| Montepulciano | 052015 | 53045 | — | — |
| Monterchi | 051024 | 52035 | — | — |
| Monteriggioni | 052016 | 53035 | — | — |
| Monteroni d'Arbia | 052017 | 53014 | — | — |
| Monterotondo Marittimo | 053027 | 58025 | — | — |
| Montescudaio | 050020 | 56040 | — | — |
| Montespertoli | 048030 | 50025 | — | — |
| Montevarchi | 051026 | 52025 | — | — |
| Monteverdi Marittimo | 050021 | 56040 | — | — |
| Monticiano | 052018 | 53015 | — | — |
| Montieri | 053017 | 58026 | — | — |
| Montignoso | 045011 | 54038 | — | — |
| Montopoli in Val d'Arno | 050022 | 56020 | — | — |
| Mulazzo | 045012 | 54026 | — | — |
| Murlo | 052019 | 53016 | — | — |
| Orbetello | 053018 | 58015 | — | — |
| Orciano Pisano | 050023 | 56040 | — | — |
| Ortignano Raggiolo | 051027 | 52010 | — | — |
| Palaia | 050024 | 56036 | — | Associated service identified; municipal source not verified |
| Palazzuolo sul Senio | 048031 | 50035 | — | — |
| Peccioli | 050025 | 56037 | — | — |
| Pelago | 048032 | 50060 | — | — |
| Pescaglia | 046022 | 55064 | — | — |
| Pescia | 047012 | 51017 | — | — |
| Piancastagnaio | 052020 | 53025 | — | — |
| Piazza al Serchio | 046023 | 55035 | — | — |
| Pienza | 052021 | 53026 | — | — |
| Pietrasanta | 046024 | 55045 | — | — |
| Pieve a Nievole | 047013 | 51018 | — | — |
| Pieve Fosciana | 046025 | 55036 | — | — |
| Pieve Santo Stefano | 051030 | 52036 | — | — |
| Piombino | 049012 | 57025 | — | — |
| [Pisa](territori/09-toscana/comuni/050026-pisa.md) | 050026 | 56121, 56122, 56123, 56124, 56125, 56126, 56127, 56128 | Candidate configured; public off | Preview passed; trial and acceptance pending |
| Pistoia | 047014 | 51100 | — | — |
| Pitigliano | 053019 | 58017 | — | — |
| Podenzana | 045013 | 54010 | — | — |
| Poggibonsi | 052022 | 53036 | — | — |
| Poggio a Caiano | 100004 | 59016 | — | — |
| Pomarance | 050027 | 56045 | — | — |
| Ponsacco | 050028 | 56038 | — | — |
| Pontassieve | 048033 | 50065 | — | — |
| Ponte Buggianese | 047016 | 51019 | — | — |
| Pontedera | 050029 | 56025 | Candidate configured; public off | Preview passed; trial and acceptance pending |
| Pontremoli | 045014 | 54027 | — | — |
| Poppi | 051031 | 52014 | — | — |
| Porcari | 046026 | 55016 | — | — |
| Porto Azzurro | 049013 | 57036 | — | — |
| Portoferraio | 049014 | 57037 | — | — |
| Prato | 100005 | 59100 | — | — |
| Pratovecchio Stia | 051041 | 52015 | — | — |
| Quarrata | 047017 | 51039 | — | — |
| Radda in Chianti | 052023 | 53017 | — | — |
| Radicofani | 052024 | 53040 | — | — |
| Radicondoli | 052025 | 53030 | — | — |
| Rapolano Terme | 052026 | 53040 | — | — |
| Reggello | 048035 | 50066 | — | — |
| Rignano sull'Arno | 048036 | 50067 | — | — |
| Rio | 049021 | 57038 | — | — |
| Riparbella | 050030 | 56046 | — | — |
| Roccalbegna | 053020 | 58053 | — | — |
| Roccastrada | 053021 | 58036 | — | — |
| Rosignano Marittimo | 049017 | 57016 | — | — |
| Rufina | 048037 | 50068 | — | — |
| Sambuca Pistoiese | 047018 | 51020 | — | — |
| San Casciano dei Bagni | 052027 | 53040 | — | — |
| San Casciano in Val di Pesa | 048038 | 50026 | — | — |
| San Gimignano | 052028 | 53037 | — | — |
| San Giovanni Valdarno | 051033 | 52027 | — | — |
| San Giuliano Terme | 050031 | 56017 | — | — |
| San Godenzo | 048039 | 50060 | — | — |
| San Marcello Piteglio | 047024 | 51028 | — | — |
| San Miniato | 050032 | 56028 | — | — |
| San Quirico d'Orcia | 052030 | 53027 | — | — |
| San Romano in Garfagnana | 046027 | 55038 | — | — |
| San Vincenzo | 049018 | 57027 | — | — |
| Sansepolcro | 051034 | 52037 | — | — |
| Santa Croce sull'Arno | 050033 | 56029 | — | — |
| Santa Fiora | 053022 | 58037 | — | — |
| Santa Luce | 050034 | 56040 | — | — |
| Santa Maria a Monte | 050035 | 56020 | — | — |
| Sarteano | 052031 | 53047 | — | — |
| Sassetta | 049019 | 57020 | — | — |
| Scandicci | 048041 | 50018 | — | — |
| Scansano | 053023 | 58054 | — | — |
| Scarlino | 053024 | 58020 | — | — |
| Scarperia e San Piero | 048053 | 50038 | — | — |
| Seggiano | 053025 | 58038 | — | — |
| Semproniano | 053028 | 58055 | — | — |
| Seravezza | 046028 | 55047 | — | — |
| Serravalle Pistoiese | 047020 | 51034 | — | — |
| Sestino | 051035 | 52038 | — | — |
| Sesto Fiorentino | 048043 | 50019 | — | — |
| Siena | 052032 | 53100 | — | — |
| Signa | 048044 | 50058 | — | — |
| Sillano Giuncugnano | 046037 | 55039 | — | — |
| Sinalunga | 052033 | 53048 | — | — |
| Sorano | 053026 | 58010 | — | — |
| Sovicille | 052034 | 53018 | — | — |
| Stazzema | 046030 | 55040 | — | — |
| Subbiano | 051037 | 52010 | — | — |
| Suvereto | 049020 | 57028 | — | — |
| Talla | 051038 | 52010 | — | — |
| Terranuova Bracciolini | 051039 | 52028 | — | — |
| Terricciola | 050036 | 56030 | — | — |
| Torrita di Siena | 052035 | 53049 | — | — |
| Trequanda | 052036 | 53020 | — | — |
| Tresana | 045015 | 54012 | — | — |
| Uzzano | 047021 | 51010 | — | — |
| Vagli Sotto | 046031 | 55030 | — | — |
| Vaglia | 048046 | 50036 | — | — |
| Vaiano | 100006 | 59021 | — | — |
| Vecchiano | 050037 | 56019 | — | — |
| Vernio | 100007 | 59024 | — | — |
| Viareggio | 046033 | 55049 | — | — |
| Vicchio | 048049 | 50039 | — | — |
| Vicopisano | 050038 | 56010 | — | — |
| Villa Basilica | 046034 | 55019 | — | — |
| Villa Collemandina | 046035 | 55030 | — | — |
| Villafranca in Lunigiana | 045016 | 54028 | — | — |
| Vinci | 048050 | 50059 | — | — |
| Volterra | 050039 | 56048 | — | — |
| Zeri | 045017 | 54029 | — | — |

<a id="region-04"></a>

### Trentino-Alto Adige/Südtirol (282 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Trentino-Alto Adige/Südtirol — region** | 04 | — | Not started | Regional sources not checked |
| Ala | 022001 | 38061 | — | — |
| Albiano | 022002 | 38041 | — | — |
| Aldeno | 022003 | 38060 | — | — |
| Aldino | 021001 | 39040 | — | — |
| Altavalle | 022235 | 38092 | — | — |
| Altopiano della Vigolana | 022236 | 38049 | — | — |
| Amblar-Don | 022237 | 38011 | — | — |
| Andalo | 022005 | 38010 | — | — |
| Andriano | 021002 | 39010 | — | — |
| Anterivo | 021003 | 39040 | — | — |
| Appiano sulla strada del vino | 021004 | 39057 | — | — |
| Arco | 022006 | 38062 | — | — |
| Avelengo | 021005 | 39010 | — | — |
| Avio | 022007 | 38063 | — | — |
| Badia | 021006 | 39036 | — | — |
| Barbiano | 021007 | 39040 | — | — |
| Baselga di Pinè | 022009 | 38042 | — | — |
| Bedollo | 022011 | 38043 | — | — |
| Besenello | 022013 | 38060 | — | — |
| Bieno | 022015 | 38050 | — | — |
| Bleggio Superiore | 022017 | 38071 | — | — |
| Bocenago | 022018 | 38080 | — | — |
| Bolzano | 021008 | 39100 | — | — |
| Bondone | 022021 | 38080 | — | — |
| Borgo Chiese | 022238 | 38083 | — | — |
| Borgo d'Anaunia | 022252 | 38013 | — | — |
| Borgo Lares | 022239 | 38079 | — | — |
| Borgo Valsugana | 022022 | 38051 | — | — |
| Braies | 021009 | 39030 | — | — |
| Brennero | 021010 | 39041 | — | — |
| Brentonico | 022025 | 38060 | — | — |
| Bresimo | 022026 | 38020 | — | — |
| Bressanone | 021011 | 39042 | — | — |
| Bronzolo | 021012 | 39051 | — | — |
| Brunico | 021013 | 39031 | — | — |
| Caderzone Terme | 022029 | 38080 | — | — |
| Caines | 021014 | 39010 | — | — |
| Calceranica al Lago | 022032 | 38050 | — | — |
| Caldaro sulla strada del vino | 021015 | 39052 | — | — |
| Caldes | 022033 | 38022 | — | — |
| Caldonazzo | 022034 | 38052 | — | — |
| Calliano | 022035 | 38060 | — | — |
| Campitello di Fassa | 022036 | 38031 | — | — |
| Campo di Trens | 021016 | 39040 | — | — |
| Campo Tures | 021017 | 39032 | — | — |
| Campodenno | 022037 | 38010 | — | — |
| Canal San Bovo | 022038 | 38050 | — | — |
| Canazei | 022039 | 38032 | — | — |
| Capriana | 022040 | 38030 | — | — |
| Carisolo | 022042 | 38080 | — | — |
| Carzano | 022043 | 38050 | — | — |
| Castel Condino | 022045 | 38082 | — | — |
| Castel Ivano | 022240 | 38059 | — | — |
| Castelbello-Ciardes | 021018 | 39020 | — | — |
| Castello Tesino | 022048 | 38053 | — | — |
| Castello-Molina di Fiemme | 022047 | 38030 | — | — |
| Castelnuovo | 022049 | 38050 | — | — |
| Castelrotto | 021019 | 39040 | — | — |
| Cavalese | 022050 | 38033 | — | — |
| Cavareno | 022051 | 38011 | — | — |
| Cavedago | 022052 | 38010 | — | — |
| Cavedine | 022053 | 38073 | — | — |
| Cavizzana | 022054 | 38022 | — | — |
| Cembra Lisignago | 022241 | 38034 | — | — |
| Cermes | 021020 | 39010 | — | — |
| Chienes | 021021 | 39030 | — | — |
| Chiusa | 021022 | 39043 | — | — |
| Cimone | 022058 | 38060 | — | — |
| Cinte Tesino | 022059 | 38050 | — | — |
| Cis | 022060 | 38020 | — | — |
| Civezzano | 022061 | 38045 | — | — |
| Cles | 022062 | 38023 | — | — |
| Comano Terme | 022228 | 38077 | — | — |
| Commezzadura | 022064 | 38020 | — | — |
| Contà | 022242 | 38093 | — | — |
| Cornedo all'Isarco | 021023 | 39053 | — | — |
| Cortaccia sulla strada del vino | 021024 | 39040 | — | — |
| Cortina sulla strada del vino | 021025 | 39040 | — | — |
| Corvara in Badia | 021026 | 39033 | — | — |
| Croviana | 022068 | 38027 | — | — |
| Curon Venosta | 021027 | 39027 | — | — |
| Dambel | 022071 | 38010 | — | — |
| Denno | 022074 | 38010 | — | — |
| Dimaro Folgarida | 022233 | 38025 | — | — |
| Dobbiaco | 021028 | 39034 | — | — |
| Drena | 022078 | 38074 | — | — |
| Dro | 022079 | 38074 | — | — |
| Egna | 021029 | 39044 | — | — |
| Fai della Paganella | 022081 | 38010 | — | — |
| Falzes | 021030 | 39030 | — | — |
| Fiavè | 022083 | 38075 | — | — |
| Fiè allo Sciliar | 021031 | 39050 | — | — |
| Fierozzo | 022085 | 38050 | — | — |
| Folgaria | 022087 | 38064 | — | — |
| Fornace | 022089 | 38040 | — | — |
| Fortezza | 021032 | 39045 | — | — |
| Frassilongo | 022090 | 38050 | — | — |
| Funes | 021033 | 39040 | — | — |
| Gais | 021034 | 39030 | — | — |
| Gargazzone | 021035 | 39010 | — | — |
| Garniga Terme | 022091 | 38060 | — | — |
| Giovo | 022092 | 38030 | — | — |
| Giustino | 022093 | 38086 | — | — |
| Glorenza | 021036 | 39020 | — | — |
| Grigno | 022095 | 38055 | — | — |
| Imer | 022097 | 38050 | — | — |
| Isera | 022098 | 38060 | — | — |
| La Valle | 021117 | 39030 | — | — |
| Laces | 021037 | 39021 | — | — |
| Lagundo | 021038 | 39022 | — | — |
| Laion | 021039 | 39040 | — | — |
| Laives | 021040 | 39055 | — | — |
| Lana | 021041 | 39011 | — | — |
| Lasa | 021042 | 39023 | — | — |
| Lauregno | 021043 | 39040 | — | — |
| Lavarone | 022102 | 38046 | — | — |
| Lavis | 022103 | 38015 | — | — |
| Ledro | 022229 | 38067 | — | — |
| Levico Terme | 022104 | 38056 | — | — |
| Livo | 022106 | 38020 | — | — |
| Lona-Lases | 022108 | 38040 | — | — |
| Luserna | 022109 | 38040 | — | — |
| Luson | 021044 | 39040 | — | — |
| Madruzzo | 022243 | 38076 | — | — |
| Magrè sulla strada del vino | 021045 | 39040 | — | — |
| Malé | 022110 | 38027 | — | — |
| Malles Venosta | 021046 | 39024 | — | — |
| Marebbe | 021047 | 39030 | — | — |
| Marlengo | 021048 | 39020 | — | — |
| Martello | 021049 | 39020 | — | — |
| Massimeno | 022112 | 38086 | — | — |
| Mazzin | 022113 | 38030 | — | — |
| Meltina | 021050 | 39010 | — | — |
| Merano | 021051 | 39012 | — | — |
| Mezzana | 022114 | 38020 | — | — |
| Mezzano | 022115 | 38050 | — | — |
| Mezzocorona | 022116 | 38016 | — | — |
| Mezzolombardo | 022117 | 38017 | — | — |
| Moena | 022118 | 38035 | — | — |
| Molveno | 022120 | 38018 | — | — |
| Monguelfo-Tesido | 021052 | 39035 | — | — |
| Montagna sulla strada del vino | 021053 | 39040 | — | — |
| Mori | 022123 | 38065 | — | — |
| Moso in Passiria | 021054 | 39013 | — | — |
| Nago-Torbole | 022124 | 38069 | — | — |
| Nalles | 021055 | 39010 | — | — |
| Naturno | 021056 | 39025 | — | — |
| Naz-Sciaves | 021057 | 39040 | — | — |
| Nogaredo | 022127 | 38060 | — | — |
| Nomi | 022128 | 38060 | — | — |
| Nova Levante | 021058 | 39056 | — | — |
| Nova Ponente | 021059 | 39050 | — | — |
| Novaledo | 022129 | 38050 | — | — |
| Novella | 022253 | 38028 | — | — |
| Ora | 021060 | 39040 | — | — |
| Ortisei | 021061 | 39046 | — | — |
| Ospedaletto | 022130 | 38050 | — | — |
| Ossana | 022131 | 38026 | — | — |
| Palù del Fersina | 022133 | 38050 | — | — |
| Panchià | 022134 | 38030 | — | — |
| Parcines | 021062 | 39020 | — | — |
| Peio | 022136 | 38024 | — | — |
| Pellizzano | 022137 | 38020 | — | — |
| Pelugo | 022138 | 38079 | — | — |
| Perca | 021063 | 39030 | — | — |
| Pergine Valsugana | 022139 | 38057 | — | — |
| Pieve di Bono-Prezzo | 022234 | 38085 | — | — |
| Pieve Tesino | 022142 | 38050 | — | — |
| Pinzolo | 022143 | 38086 | — | — |
| Plaus | 021064 | 39025 | — | — |
| Pomarolo | 022144 | 38060 | — | — |
| Ponte Gardena | 021065 | 39040 | — | — |
| Porte di Rendena | 022244 | 38094 | — | — |
| Postal | 021066 | 39014 | — | — |
| Prato allo Stelvio | 021067 | 39026 | — | — |
| Predaia | 022230 | 38012 | — | — |
| Predazzo | 022147 | 38037 | — | — |
| Predoi | 021068 | 39030 | — | — |
| Primiero San Martino di Castrozza | 022245 | 38054 | — | — |
| Proves | 021069 | 39040 | — | — |
| Rabbi | 022150 | 38020 | — | — |
| Racines | 021070 | 39040 | — | — |
| Rasun-Anterselva | 021071 | 39030 | — | — |
| Renon | 021072 | 39054 | — | — |
| Rifiano | 021073 | 39010 | — | — |
| Rio di Pusteria | 021074 | 39037 | — | — |
| Riva del Garda | 022153 | 38066 | — | — |
| Rodengo | 021075 | 39037 | — | — |
| Romeno | 022155 | 38010 | — | — |
| Roncegno Terme | 022156 | 38050 | — | — |
| Ronchi Valsugana | 022157 | 38050 | — | — |
| Ronzo-Chienis | 022135 | 38060 | — | — |
| Ronzone | 022159 | 38010 | — | — |
| Roverè della Luna | 022160 | 38030 | — | — |
| Rovereto | 022161 | 38068 | — | — |
| Ruffrè-Mendola | 022162 | 38010 | — | — |
| Rumo | 022163 | 38020 | — | — |
| Sagron Mis | 022164 | 38050 | — | — |
| Salorno sulla strada del vino | 021076 | 39040 | — | — |
| Samone | 022165 | 38059 | — | — |
| San Candido | 021077 | 39038 | — | — |
| San Genesio Atesino | 021079 | 39050 | — | — |
| San Giovanni di Fassa | 022250 | 38036 | — | — |
| San Leonardo in Passiria | 021080 | 39015 | — | — |
| San Lorenzo di Sebato | 021081 | 39030 | — | — |
| San Lorenzo Dorsino | 022231 | 38078 | — | — |
| San Martino in Badia | 021082 | 39030 | — | — |
| San Martino in Passiria | 021083 | 39010 | — | — |
| San Michele all'Adige | 022167 | 38098 | — | — |
| San Pancrazio | 021084 | 39010 | — | — |
| Sant'Orsola Terme | 022168 | 38050 | — | — |
| Santa Cristina Valgardena | 021085 | 39047 | — | — |
| Sanzeno | 022169 | 38010 | — | — |
| Sarentino | 021086 | 39058 | — | — |
| Sarnonico | 022170 | 38011 | — | — |
| Scena | 021087 | 39017 | — | — |
| Scurelle | 022171 | 38050 | — | — |
| Segonzano | 022172 | 38047 | — | — |
| Sella Giudicarie | 022246 | 38087 | — | — |
| Selva dei Molini | 021088 | 39030 | — | — |
| Selva di Val Gardena | 021089 | 39048 | — | — |
| Senale-San Felice | 021118 | 39010 | — | — |
| Senales | 021091 | 39020 | — | — |
| Sesto | 021092 | 39030 | — | — |
| Sfruz | 022173 | 38010 | — | — |
| Silandro | 021093 | 39028 | — | — |
| Sluderno | 021094 | 39020 | — | — |
| Soraga di Fassa | 022176 | 38030 | — | — |
| Sover | 022177 | 38048 | — | — |
| Spiazzo | 022179 | 38088 | — | — |
| Spormaggiore | 022180 | 38010 | — | — |
| Sporminore | 022181 | 38010 | — | — |
| Stelvio | 021095 | 39029 | — | — |
| Stenico | 022182 | 38070 | — | — |
| Storo | 022183 | 38089 | — | — |
| Strembo | 022184 | 38080 | — | — |
| Telve | 022188 | 38050 | — | — |
| Telve di Sopra | 022189 | 38050 | — | — |
| Tenna | 022190 | 38050 | — | — |
| Tenno | 022191 | 38060 | — | — |
| Terento | 021096 | 39030 | — | — |
| Terlano | 021097 | 39018 | — | — |
| Termeno sulla strada del vino | 021098 | 39040 | — | — |
| Terragnolo | 022193 | 38060 | — | — |
| Terre d'Adige | 022251 | 38097 | — | — |
| Terzolas | 022195 | 38027 | — | — |
| Tesero | 022196 | 38038 | — | — |
| Tesimo | 021099 | 39010 | — | — |
| Tione di Trento | 022199 | 38079 | — | — |
| Tires | 021100 | 39050 | — | — |
| Tirolo | 021101 | 39019 | — | — |
| Ton | 022200 | 38010 | — | — |
| Torcegno | 022202 | 38050 | — | — |
| Trambileno | 022203 | 38068 | — | — |
| Tre Ville | 022247 | 38095 | — | — |
| Trento | 022205 | 38121, 38122, 38123 | — | — |
| Trodena nel parco naturale | 021102 | 39040 | — | — |
| Tubre | 021103 | 39020 | — | — |
| Ultimo | 021104 | 39016 | — | — |
| Vadena | 021105 | 39051 | — | — |
| Val di Vizze | 021107 | 39049 | — | — |
| Valdaone | 022232 | 38091 | — | — |
| Valdaora | 021106 | 39030 | — | — |
| Valfloriana | 022209 | 38040 | — | — |
| Vallarsa | 022210 | 38060 | — | — |
| Valle Aurina | 021108 | 39030 | — | — |
| Valle di Casies | 021109 | 39030 | — | — |
| Vallelaghi | 022248 | 38096 | — | — |
| Vandoies | 021110 | 39030 | — | — |
| Varna | 021111 | 39040 | — | — |
| Velturno | 021116 | 39040 | — | — |
| Verano | 021112 | 39010 | — | — |
| Vermiglio | 022213 | 38029 | — | — |
| Vignola-Falesina | 022216 | 38057 | — | — |
| Villa Lagarina | 022222 | 38060 | — | — |
| Villabassa | 021113 | 39039 | — | — |
| Villandro | 021114 | 39040 | — | — |
| Ville d'Anaunia | 022249 | 38019 | — | — |
| Ville di Fiemme | 022254 | 38099 | — | — |
| Vipiteno | 021115 | 39049 | — | — |
| Volano | 022224 | 38060 | — | — |
| Ziano di Fiemme | 022226 | 38030 | — | — |

<a id="region-10"></a>

### Umbria (92 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Umbria — region** | 10 | — | Not started | Regional sources not checked |
| Acquasparta | 055001 | 05021 | — | — |
| Allerona | 055002 | 05011 | — | — |
| Alviano | 055003 | 05020 | — | — |
| Amelia | 055004 | 05022 | — | — |
| Arrone | 055005 | 05031 | — | — |
| Assisi | 054001 | 06081 | — | — |
| Attigliano | 055006 | 05012 | — | — |
| Avigliano Umbro | 055033 | 05020 | — | — |
| Baschi | 055007 | 05023 | — | — |
| Bastia Umbra | 054002 | 06083 | — | — |
| Bettona | 054003 | 06084 | — | — |
| Bevagna | 054004 | 06031 | — | — |
| Calvi dell'Umbria | 055008 | 05032 | — | — |
| Campello sul Clitunno | 054005 | 06042 | — | — |
| Cannara | 054006 | 06033 | — | — |
| Cascia | 054007 | 06043 | — | — |
| Castel Giorgio | 055009 | 05013 | — | — |
| Castel Ritaldi | 054008 | 06044 | — | — |
| Castel Viscardo | 055010 | 05014 | — | — |
| Castiglione del Lago | 054009 | 06061 | — | — |
| Cerreto di Spoleto | 054010 | 06041 | — | — |
| Citerna | 054011 | 06010 | — | — |
| Città della Pieve | 054012 | 06062 | — | — |
| Città di Castello | 054013 | 06012 | — | — |
| Collazzone | 054014 | 06050 | — | — |
| Corciano | 054015 | 06073 | — | — |
| Costacciaro | 054016 | 06021 | — | — |
| Deruta | 054017 | 06053 | — | — |
| Fabro | 055011 | 05015 | — | — |
| Ferentillo | 055012 | 05034 | — | — |
| Ficulle | 055013 | 05016 | — | — |
| Foligno | 054018 | 06034 | — | — |
| Fossato di Vico | 054019 | 06022 | — | — |
| Fratta Todina | 054020 | 06054 | — | — |
| Giano dell'Umbria | 054021 | 06030 | — | — |
| Giove | 055014 | 05024 | — | — |
| Gualdo Cattaneo | 054022 | 06035 | — | — |
| Gualdo Tadino | 054023 | 06023 | — | — |
| Guardea | 055015 | 05025 | — | — |
| Gubbio | 054024 | 06024 | — | — |
| Lisciano Niccone | 054025 | 06060 | — | — |
| Lugnano in Teverina | 055016 | 05020 | — | — |
| Magione | 054026 | 06063 | — | — |
| Marsciano | 054027 | 06055 | — | — |
| Massa Martana | 054028 | 06056 | — | — |
| Monte Castello di Vibio | 054029 | 06057 | — | — |
| Monte Santa Maria Tiberina | 054032 | 06010 | — | — |
| Montecastrilli | 055017 | 05026 | — | — |
| Montecchio | 055018 | 05020 | — | — |
| Montefalco | 054030 | 06036 | — | — |
| Montefranco | 055019 | 05030 | — | — |
| Montegabbione | 055020 | 05010 | — | — |
| Monteleone d'Orvieto | 055021 | 05017 | — | — |
| Monteleone di Spoleto | 054031 | 06045 | — | — |
| Montone | 054033 | 06014 | — | — |
| Narni | 055022 | 05035 | — | — |
| Nocera Umbra | 054034 | 06025 | — | — |
| Norcia | 054035 | 06046 | — | — |
| Orvieto | 055023 | 05018 | — | — |
| Otricoli | 055024 | 05030 | — | — |
| Paciano | 054036 | 06060 | — | — |
| Panicale | 054037 | 06064 | — | — |
| Parrano | 055025 | 05010 | — | — |
| Passignano sul Trasimeno | 054038 | 06065 | — | — |
| Penna in Teverina | 055026 | 05028 | — | — |
| Perugia | 054039 | 06121, 06122, 06123, 06124, 06125, 06126, 06127, 06128, 06129, 06131, 06132, 06133, 06134, 06135 | — | — |
| Piegaro | 054040 | 06066 | — | — |
| Pietralunga | 054041 | 06026 | — | — |
| Poggiodomo | 054042 | 06040 | — | — |
| Polino | 055027 | 05030 | — | — |
| Porano | 055028 | 05010 | — | — |
| Preci | 054043 | 06047 | — | — |
| San Gemini | 055029 | 05029 | — | — |
| San Giustino | 054044 | 06016 | — | — |
| San Venanzo | 055030 | 05010 | — | — |
| Sant'Anatolia di Narco | 054045 | 06040 | — | — |
| Scheggia e Pascelupo | 054046 | 06027 | — | — |
| Scheggino | 054047 | 06040 | — | — |
| Sellano | 054048 | 06030 | — | — |
| Sigillo | 054049 | 06028 | — | — |
| Spello | 054050 | 06038 | — | — |
| Spoleto | 054051 | 06049 | — | — |
| Stroncone | 055031 | 05039 | — | — |
| Terni | 055032 | 05100 | — | — |
| Todi | 054052 | 06059 | — | — |
| Torgiano | 054053 | 06089 | — | — |
| Trevi | 054054 | 06039 | — | — |
| Tuoro sul Trasimeno | 054055 | 06069 | — | — |
| Umbertide | 054056 | 06019 | — | — |
| Valfabbrica | 054057 | 06029 | — | — |
| Vallo di Nera | 054058 | 06040 | — | — |
| Valtopina | 054059 | 06030 | — | — |

<a id="region-02"></a>

### Valle d'Aosta/Vallée d'Aoste (74 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Valle d'Aosta/Vallée d'Aoste — region** | 02 | — | Not started | Regional sources not checked |
| Allein | 007001 | 11010 | — | — |
| Antey-Saint-André | 007002 | 11020 | — | — |
| Aosta | 007003 | 11100 | — | — |
| Arnad | 007004 | 11020 | — | — |
| Arvier | 007005 | 11011 | — | — |
| Avise | 007006 | 11010 | — | — |
| Ayas | 007007 | 11020 | — | — |
| Aymavilles | 007008 | 11010 | — | — |
| Bard | 007009 | 11020 | — | — |
| Bionaz | 007010 | 11010 | — | — |
| Brissogne | 007011 | 11020 | — | — |
| Brusson | 007012 | 11022 | — | — |
| Challand-Saint-Anselme | 007013 | 11020 | — | — |
| Challand-Saint-Victor | 007014 | 11020 | — | — |
| Chambave | 007015 | 11023 | — | — |
| Chamois | 007016 | 11020 | — | — |
| Champdepraz | 007017 | 11020 | — | — |
| Champorcher | 007018 | 11020 | — | — |
| Charvensod | 007019 | 11020 | — | — |
| Châtillon | 007020 | 11024 | — | — |
| Cogne | 007021 | 11012 | — | — |
| Courmayeur | 007022 | 11013 | — | — |
| Donnas | 007023 | 11020 | — | — |
| Doues | 007024 | 11010 | — | — |
| Emarèse | 007025 | 11020 | — | — |
| Etroubles | 007026 | 11014 | — | — |
| Fénis | 007027 | 11020 | — | — |
| Fontainemore | 007028 | 11020 | — | — |
| Gaby | 007029 | 11020 | — | — |
| Gignod | 007030 | 11010 | — | — |
| Gressan | 007031 | 11020 | — | — |
| Gressoney-La-Trinité | 007032 | 11020 | — | — |
| Gressoney-Saint-Jean | 007033 | 11025 | — | — |
| Hône | 007034 | 11020 | — | — |
| Introd | 007035 | 11010 | — | — |
| Issime | 007036 | 11020 | — | — |
| Issogne | 007037 | 11020 | — | — |
| Jovençan | 007038 | 11020 | — | — |
| La Magdeleine | 007039 | 11020 | — | — |
| La Salle | 007040 | 11015 | — | — |
| La Thuile | 007041 | 11016 | — | — |
| Lillianes | 007042 | 11020 | — | — |
| Montjovet | 007043 | 11020 | — | — |
| Morgex | 007044 | 11017 | — | — |
| Nus | 007045 | 11020 | — | — |
| Ollomont | 007046 | 11010 | — | — |
| Oyace | 007047 | 11010 | — | — |
| Perloz | 007048 | 11020 | — | — |
| Pollein | 007049 | 11020 | — | — |
| Pont-Saint-Martin | 007052 | 11026 | — | — |
| Pontboset | 007050 | 11020 | — | — |
| Pontey | 007051 | 11024 | — | — |
| Pré-Saint-Didier | 007053 | 11010 | — | — |
| Quart | 007054 | 11020 | — | — |
| Rhêmes-Notre-Dame | 007055 | 11010 | — | — |
| Rhêmes-Saint-Georges | 007056 | 11010 | — | — |
| Roisan | 007057 | 11010 | — | — |
| Saint-Christophe | 007058 | 11020 | — | — |
| Saint-Denis | 007059 | 11023 | — | — |
| Saint-Marcel | 007060 | 11020 | — | — |
| Saint-Nicolas | 007061 | 11010 | — | — |
| Saint-Oyen | 007062 | 11014 | — | — |
| Saint-Pierre | 007063 | 11010 | — | — |
| Saint-Rhémy-en-Bosses | 007064 | 11010 | — | — |
| Saint-Vincent | 007065 | 11027 | — | — |
| Sarre | 007066 | 11010 | — | — |
| Torgnon | 007067 | 11020 | — | — |
| Valgrisenche | 007068 | 11010 | — | — |
| Valpelline | 007069 | 11010 | — | — |
| Valsavarenche | 007070 | 11010 | — | — |
| Valtournenche | 007071 | 11028 | — | — |
| Verrayes | 007072 | 11020 | — | — |
| Verrès | 007073 | 11029 | — | — |
| Villeneuve | 007074 | 11018 | — | — |

<a id="region-05"></a>

### Veneto (559 municipalities)

| Territory | ISTAT code | CAP | Implementation | Source checks |
| --- | --- | --- | --- | --- |
| **Veneto — region** | 05 | — | Not started | Regional sources not checked |
| Abano Terme | 028001 | 35031 | — | — |
| Adria | 029001 | 45011 | — | — |
| Affi | 023001 | 37010 | — | — |
| Agna | 028002 | 35021 | — | — |
| Agordo | 025001 | 32021 | — | — |
| Agugliaro | 024001 | 36020 | — | — |
| Albaredo d'Adige | 023002 | 37041 | — | — |
| Albettone | 024002 | 36020 | — | — |
| Albignasego | 028003 | 35020 | — | — |
| Alleghe | 025003 | 32022 | — | — |
| Alonte | 024003 | 36045 | — | — |
| Alpago | 025072 | 32016 | — | — |
| Altavilla Vicentina | 024004 | 36077 | — | — |
| Altissimo | 024005 | 36070 | — | — |
| Altivole | 026001 | 31030 | — | — |
| Angiari | 023003 | 37050 | — | — |
| Anguillara Veneta | 028004 | 35022 | — | — |
| Annone Veneto | 027001 | 30020 | — | — |
| Arcade | 026002 | 31030 | — | — |
| Arcole | 023004 | 37040 | — | — |
| Arcugnano | 024006 | 36057 | — | — |
| Ariano nel Polesine | 029002 | 45012 | — | — |
| Arquà Petrarca | 028005 | 35032 | — | — |
| Arquà Polesine | 029003 | 45031 | — | — |
| Arre | 028006 | 35020 | — | — |
| Arsiè | 025004 | 32030 | — | — |
| Arsiero | 024007 | 36011 | — | — |
| Arzergrande | 028007 | 35020 | — | — |
| Arzignano | 024008 | 36071 | — | — |
| Asiago | 024009 | 36012 | — | — |
| Asigliano Veneto | 024010 | 36020 | — | — |
| Asolo | 026003 | 31011 | — | — |
| Auronzo di Cadore | 025005 | 32041 | — | — |
| Badia Calavena | 023005 | 37030 | — | — |
| Badia Polesine | 029004 | 45021 | — | — |
| Bagnoli di Sopra | 028008 | 35023 | — | — |
| Bagnolo di Po | 029005 | 45022 | — | — |
| Baone | 028009 | 35030 | — | — |
| Barbarano Mossano | 024124 | 36048 | — | — |
| Barbona | 028010 | 35040 | — | — |
| Bardolino | 023006 | 37011 | — | — |
| Bassano del Grappa | 024012 | 36061 | — | — |
| Battaglia Terme | 028011 | 35041 | — | — |
| Belfiore | 023007 | 37050 | — | — |
| Belluno | 025006 | 32100 | — | — |
| Bergantino | 029006 | 45032 | — | — |
| Bevilacqua | 023008 | 37040 | — | — |
| Boara Pisani | 028012 | 35040 | — | — |
| Bolzano Vicentino | 024013 | 36050 | — | — |
| Bonavigo | 023009 | 37040 | — | — |
| Borca di Cadore | 025007 | 32040 | — | — |
| Borgo Valbelluna | 025074 | 32026 | — | — |
| Borgo Veneto | 028107 | 35046 | — | — |
| Borgoricco | 028013 | 35010 | — | — |
| Borso del Grappa | 026004 | 31030 | — | — |
| Bosaro | 029007 | 45033 | — | — |
| Boschi Sant'Anna | 023010 | 37040 | — | — |
| Bosco Chiesanuova | 023011 | 37021 | — | — |
| Bovolenta | 028014 | 35024 | — | — |
| Bovolone | 023012 | 37051 | — | — |
| Breda di Piave | 026005 | 31030 | — | — |
| Breganze | 024014 | 36042 | — | — |
| Brendola | 024015 | 36040 | — | — |
| Brentino Belluno | 023013 | 37020 | — | — |
| Brenzone sul Garda | 023014 | 37010 | — | — |
| Bressanvido | 024016 | 36050 | — | — |
| Brogliano | 024017 | 36070 | — | — |
| Brugine | 028015 | 35020 | — | — |
| Bussolengo | 023015 | 37012 | — | — |
| Buttapietra | 023016 | 37060 | — | — |
| Cadoneghe | 028016 | 35010 | — | — |
| Caerano di San Marco | 026006 | 31031 | — | — |
| Calalzo di Cadore | 025008 | 32042 | — | — |
| Caldiero | 023017 | 37042 | — | — |
| Caldogno | 024018 | 36030 | — | — |
| Calto | 029008 | 45030 | — | — |
| Caltrano | 024019 | 36030 | — | — |
| Calvene | 024020 | 36030 | — | — |
| Camisano Vicentino | 024021 | 36043 | — | — |
| Campagna Lupia | 027002 | 30010 | — | — |
| Campiglia dei Berici | 024022 | 36020 | — | — |
| Campo San Martino | 028020 | 35010 | — | — |
| Campodarsego | 028017 | 35011 | — | — |
| Campodoro | 028018 | 35010 | — | — |
| Campolongo Maggiore | 027003 | 30010 | — | — |
| Camponogara | 027004 | 30010 | — | — |
| Camposampiero | 028019 | 35012 | — | — |
| Canale d'Agordo | 025023 | 32020 | — | — |
| Canaro | 029009 | 45034 | — | — |
| Canda | 029010 | 45020 | — | — |
| Candiana | 028021 | 35020 | — | — |
| Caorle | 027005 | 30021 | — | — |
| Cappella Maggiore | 026007 | 31012 | — | — |
| Caprino Veronese | 023018 | 37013 | — | — |
| Carbonera | 026008 | 31030 | — | — |
| Carmignano di Brenta | 028023 | 35010 | — | — |
| Carrè | 024024 | 36010 | — | — |
| Cartigliano | 024025 | 36050 | — | — |
| Cartura | 028026 | 35025 | — | — |
| Casale di Scodosia | 028027 | 35040 | — | — |
| Casale sul Sile | 026009 | 31032 | — | — |
| Casaleone | 023019 | 37052 | — | — |
| Casalserugo | 028028 | 35020 | — | — |
| Casier | 026010 | 31030 | — | — |
| Cassola | 024026 | 36022 | — | — |
| Castagnaro | 023020 | 37043 | — | — |
| Castegnero Nanto | 024129 | 36020 | — | — |
| Castel d'Azzano | 023021 | 37060 | — | — |
| Castelbaldo | 028029 | 35040 | — | — |
| Castelcucco | 026011 | 31030 | — | — |
| Castelfranco Veneto | 026012 | 31033 | — | — |
| Castelgomberto | 024028 | 36070 | — | — |
| Castelguglielmo | 029011 | 45020 | — | — |
| Castello di Godego | 026013 | 31030 | — | — |
| Castelmassa | 029012 | 45035 | — | — |
| Castelnovo Bariano | 029013 | 45030 | — | — |
| Castelnuovo del Garda | 023022 | 37014 | — | — |
| Cavaion Veronese | 023023 | 37010 | — | — |
| Cavallino-Treporti | 027044 | 30013 | — | — |
| Cavarzere | 027006 | 30014 | — | — |
| Cavaso del Tomba | 026014 | 31034 | — | — |
| Cazzano di Tramigna | 023024 | 37030 | — | — |
| Ceggia | 027007 | 30022 | — | — |
| Cencenighe Agordino | 025010 | 32020 | — | — |
| Ceneselli | 029014 | 45030 | — | — |
| Cerea | 023025 | 37053 | — | — |
| Ceregnano | 029015 | 45010 | — | — |
| Cerro Veronese | 023026 | 37020 | — | — |
| Cervarese Santa Croce | 028030 | 35030 | — | — |
| Cesiomaggiore | 025011 | 32030 | — | — |
| Cessalto | 026015 | 31040 | — | — |
| Chiampo | 024029 | 36072 | — | — |
| Chiarano | 026016 | 31040 | — | — |
| Chies d'Alpago | 025012 | 32010 | — | — |
| Chioggia | 027008 | 30015 | — | — |
| Chiuppano | 024030 | 36010 | — | — |
| Cibiana di Cadore | 025013 | 32040 | — | — |
| Cimadolmo | 026017 | 31010 | — | — |
| Cinto Caomaggiore | 027009 | 30020 | — | — |
| Cinto Euganeo | 028031 | 35030 | — | — |
| Cison di Valmarino | 026018 | 31030 | — | — |
| Cittadella | 028032 | 35013 | — | — |
| Codevigo | 028033 | 35020 | — | — |
| Codognè | 026019 | 31013 | — | — |
| Cogollo del Cengio | 024032 | 36010 | — | — |
| Colceresa | 024126 | 36064 | — | — |
| Colle Santa Lucia | 025014 | 32020 | — | — |
| Colle Umberto | 026020 | 31014 | — | — |
| Cologna Veneta | 023027 | 37044 | — | — |
| Colognola ai Colli | 023028 | 37030 | — | — |
| Comelico Superiore | 025015 | 32040 | — | — |
| Cona | 027010 | 30010 | — | — |
| Concamarise | 023029 | 37050 | — | — |
| Concordia Sagittaria | 027011 | 30023 | — | — |
| Conegliano | 026021 | 31015 | — | — |
| Conselve | 028034 | 35026 | — | — |
| Corbola | 029017 | 45015 | — | — |
| Cordignano | 026022 | 31016 | — | — |
| Cornedo Vicentino | 024034 | 36073 | — | — |
| Cornuda | 026023 | 31041 | — | — |
| Correzzola | 028035 | 35020 | — | — |
| Cortina d'Ampezzo | 025016 | 32043 | — | — |
| Costa di Rovigo | 029018 | 45023 | — | — |
| Costabissara | 024035 | 36030 | — | — |
| Costermano sul Garda | 023030 | 37010 | — | — |
| Creazzo | 024036 | 36051 | — | — |
| Crespadoro | 024037 | 36070 | — | — |
| Crespino | 029019 | 45030 | — | — |
| Crocetta del Montello | 026025 | 31035 | — | — |
| Curtarolo | 028036 | 35010 | — | — |
| Danta di Cadore | 025017 | 32040 | — | — |
| Dolcè | 023031 | 37020 | — | — |
| Dolo | 027012 | 30031 | — | — |
| Domegge di Cadore | 025018 | 32040 | — | — |
| Due Carrare | 028106 | 35020 | — | — |
| Dueville | 024038 | 36031 | — | — |
| Enego | 024039 | 36052 | — | — |
| Eraclea | 027013 | 30020 | — | — |
| Erbè | 023032 | 37060 | — | — |
| Erbezzo | 023033 | 37020 | — | — |
| Este | 028037 | 35042 | — | — |
| Falcade | 025019 | 32020 | — | — |
| Fara Vicentino | 024040 | 36030 | — | — |
| Farra di Soligo | 026026 | 31010 | — | — |
| Feltre | 025021 | 32032 | — | — |
| Ferrara di Monte Baldo | 023034 | 37020 | — | — |
| Ficarolo | 029021 | 45036 | — | — |
| Fiesso d'Artico | 027014 | 30032 | — | — |
| Fiesso Umbertiano | 029022 | 45024 | — | — |
| Follina | 026027 | 31051 | — | — |
| Fontanelle | 026028 | 31043 | — | — |
| Fontaniva | 028038 | 35014 | — | — |
| Fonte | 026029 | 31010 | — | — |
| Fonzaso | 025022 | 32030 | — | — |
| Fossalta di Piave | 027015 | 30020 | — | — |
| Fossalta di Portogruaro | 027016 | 30025 | — | — |
| Fossò | 027017 | 30030 | — | — |
| Foza | 024041 | 36010 | — | — |
| Frassinelle Polesine | 029023 | 45030 | — | — |
| Fratta Polesine | 029024 | 45025 | — | — |
| Fregona | 026030 | 31010 | — | — |
| Fumane | 023035 | 37022 | — | — |
| Gaiarine | 026031 | 31018 | — | — |
| Gaiba | 029025 | 45030 | — | — |
| Galliera Veneta | 028039 | 35015 | — | — |
| Gallio | 024042 | 36032 | — | — |
| Galzignano Terme | 028040 | 35030 | — | — |
| Gambellara | 024043 | 36053 | — | — |
| Garda | 023036 | 37016 | — | — |
| Gavello | 029026 | 45010 | — | — |
| Gazzo | 028041 | 35010 | — | — |
| Gazzo Veronese | 023037 | 37060 | — | — |
| Giacciano con Baruchella | 029027 | 45020 | — | — |
| Giavera del Montello | 026032 | 31040 | — | — |
| Godega di Sant'Urbano | 026033 | 31010 | — | — |
| Gorgo al Monticano | 026034 | 31040 | — | — |
| Gosaldo | 025025 | 32020 | — | — |
| Grantorto | 028042 | 35010 | — | — |
| Granze | 028043 | 35040 | — | — |
| Grezzana | 023038 | 37023 | — | — |
| Grisignano di Zocco | 024046 | 36040 | — | — |
| Gruaro | 027018 | 30020 | — | — |
| Grumolo delle Abbadesse | 024047 | 36040 | — | — |
| Guarda Veneta | 029028 | 45030 | — | — |
| Illasi | 023039 | 37031 | — | — |
| Isola della Scala | 023040 | 37063 | — | — |
| Isola Rizza | 023041 | 37050 | — | — |
| Isola Vicentina | 024048 | 36033 | — | — |
| Istrana | 026035 | 31036 | — | — |
| Jesolo | 027019 | 30016 | — | — |
| La Valle Agordina | 025027 | 32020 | — | — |
| Laghi | 024049 | 36010 | — | — |
| Lamon | 025026 | 32033 | — | — |
| Lastebasse | 024050 | 36040 | — | — |
| Lavagno | 023042 | 37030 | — | — |
| Lazise | 023043 | 37017 | — | — |
| Legnago | 023044 | 37045 | — | — |
| Legnaro | 028044 | 35020 | — | — |
| Lendinara | 029029 | 45026 | — | — |
| Limana | 025029 | 32020 | — | — |
| Limena | 028045 | 35010 | — | — |
| Livinallongo del Col di Lana | 025030 | 32020 | — | — |
| Longare | 024051 | 36023 | — | — |
| Longarone | 025071 | 32013 | — | — |
| Lonigo | 024052 | 36045 | — | — |
| Loreggia | 028046 | 35010 | — | — |
| Lorenzago di Cadore | 025032 | 32040 | — | — |
| Loreo | 029030 | 45017 | — | — |
| Loria | 026036 | 31037 | — | — |
| Lozzo Atestino | 028047 | 35034 | — | — |
| Lozzo di Cadore | 025033 | 32040 | — | — |
| Lugo di Vicenza | 024053 | 36030 | — | — |
| Lusia | 029031 | 45020 | — | — |
| Lusiana Conco | 024127 | 36046 | — | — |
| Malcesine | 023045 | 37018 | — | — |
| Malo | 024055 | 36034 | — | — |
| Mansuè | 026037 | 31040 | — | — |
| Marano di Valpolicella | 023046 | 37020 | — | — |
| Marano Vicentino | 024056 | 36035 | — | — |
| Marcon | 027020 | 30020 | — | — |
| Mareno di Piave | 026038 | 31010 | — | — |
| Marostica | 024057 | 36063 | — | — |
| Martellago | 027021 | 30030 | — | — |
| Maser | 026039 | 31010 | — | — |
| Maserà di Padova | 028048 | 35020 | — | — |
| Maserada sul Piave | 026040 | 31052 | — | — |
| Masi | 028049 | 35040 | — | — |
| Massanzago | 028050 | 35010 | — | — |
| Meduna di Livenza | 026041 | 31040 | — | — |
| Megliadino San Vitale | 028052 | 35040 | — | — |
| Melara | 029032 | 45037 | — | — |
| Meolo | 027022 | 30020 | — | — |
| Merlara | 028053 | 35040 | — | — |
| Mestrino | 028054 | 35035 | — | — |
| Mezzane di Sotto | 023047 | 37030 | — | — |
| Miane | 026042 | 31050 | — | — |
| Minerbe | 023048 | 37046 | — | — |
| Mira | 027023 | 30034 | — | — |
| Mirano | 027024 | 30035 | — | — |
| Mogliano Veneto | 026043 | 31021 | — | — |
| Monastier di Treviso | 026044 | 31050 | — | — |
| Monfumo | 026045 | 31010 | — | — |
| Monselice | 028055 | 35043 | — | — |
| Montagnana | 028056 | 35044 | — | — |
| Monte di Malo | 024063 | 36030 | — | — |
| Montebello Vicentino | 024060 | 36054 | — | — |
| Montebelluna | 026046 | 31044 | — | — |
| Montecchia di Crosara | 023049 | 37030 | — | — |
| Montecchio Maggiore | 024061 | 36075 | — | — |
| Montecchio Precalcino | 024062 | 36030 | — | — |
| Monteforte d'Alpone | 023050 | 37032 | — | — |
| Montegalda | 024064 | 36047 | — | — |
| Montegaldella | 024065 | 36047 | — | — |
| Montegrotto Terme | 028057 | 35036 | — | — |
| Monteviale | 024066 | 36050 | — | — |
| Monticello Conte Otto | 024067 | 36010 | — | — |
| Montorso Vicentino | 024068 | 36050 | — | — |
| Morgano | 026047 | 31050 | — | — |
| Moriago della Battaglia | 026048 | 31010 | — | — |
| Motta di Livenza | 026049 | 31045 | — | — |
| Mozzecane | 023051 | 37060 | — | — |
| Musile di Piave | 027025 | 30024 | — | — |
| Mussolente | 024070 | 36065 | — | — |
| Negrar di Valpolicella | 023052 | 37024 | — | — |
| Nervesa della Battaglia | 026050 | 31040 | — | — |
| Noale | 027026 | 30033 | — | — |
| Nogara | 023053 | 37054 | — | — |
| Nogarole Rocca | 023054 | 37060 | — | — |
| Nogarole Vicentino | 024072 | 36070 | — | — |
| Nove | 024073 | 36055 | — | — |
| Noventa di Piave | 027027 | 30020 | — | — |
| Noventa Padovana | 028058 | 35027 | — | — |
| Noventa Vicentina | 024074 | 36025 | — | — |
| Occhiobello | 029033 | 45030 | — | — |
| Oderzo | 026051 | 31046 | — | — |
| Oppeano | 023055 | 37050 | — | — |
| Orgiano | 024075 | 36040 | — | — |
| Ormelle | 026052 | 31024 | — | — |
| Orsago | 026053 | 31010 | — | — |
| Ospedaletto Euganeo | 028059 | 35045 | — | — |
| Ospitale di Cadore | 025035 | 32010 | — | — |
| Padova | 028060 | 35121, 35122, 35123, 35124, 35125, 35126, 35127, 35128, 35129, 35131, 35132, 35133, 35134, 35135, 35136, 35137, 35138, 35139, 35141, 35142, 35143 | — | — |
| Paese | 026055 | 31038 | — | — |
| Palù | 023056 | 37050 | — | — |
| Papozze | 029034 | 45010 | — | — |
| Pastrengo | 023057 | 37010 | — | — |
| Pedavena | 025036 | 32034 | — | — |
| Pedemonte | 024076 | 36040 | — | — |
| Pederobba | 026056 | 31040 | — | — |
| Perarolo di Cadore | 025037 | 32010 | — | — |
| Pernumia | 028061 | 35020 | — | — |
| Pescantina | 023058 | 37026 | — | — |
| Peschiera del Garda | 023059 | 37019 | — | — |
| Pettorazza Grimani | 029035 | 45010 | — | — |
| Piacenza d'Adige | 028062 | 35040 | — | — |
| Pianezze | 024077 | 36060 | — | — |
| Pianiga | 027028 | 30030 | — | — |
| Piazzola sul Brenta | 028063 | 35016 | — | — |
| Pieve del Grappa | 026096 | 31017 | — | — |
| Pieve di Cadore | 025039 | 32044 | — | — |
| Pieve di Soligo | 026057 | 31053 | — | — |
| Pincara | 029036 | 45020 | — | — |
| Piombino Dese | 028064 | 35017 | — | — |
| Piove di Sacco | 028065 | 35028 | — | — |
| Piovene Rocchette | 024078 | 36013 | — | — |
| Pojana Maggiore | 024079 | 36026 | — | — |
| Polesella | 029037 | 45038 | — | — |
| Polverara | 028066 | 35020 | — | — |
| Ponso | 028067 | 35040 | — | — |
| Ponte di Piave | 026058 | 31047 | — | — |
| Ponte nelle Alpi | 025040 | 32014 | — | — |
| Ponte San Nicolò | 028069 | 35020 | — | — |
| Pontecchio Polesine | 029038 | 45030 | — | — |
| Pontelongo | 028068 | 35029 | — | — |
| Ponzano Veneto | 026059 | 31050 | — | — |
| Porto Tolle | 029039 | 45018 | — | — |
| Porto Viro | 029052 | 45014 | — | — |
| Portobuffolè | 026060 | 31040 | — | — |
| Portogruaro | 027029 | 30026 | — | — |
| Posina | 024080 | 36010 | — | — |
| Possagno | 026061 | 31054 | — | — |
| Pove del Grappa | 024081 | 36020 | — | — |
| Povegliano | 026062 | 31050 | — | — |
| Povegliano Veronese | 023060 | 37064 | — | — |
| Pozzoleone | 024082 | 36050 | — | — |
| Pozzonovo | 028070 | 35020 | — | — |
| Pramaggiore | 027030 | 30020 | — | — |
| Preganziol | 026063 | 31022 | — | — |
| Pressana | 023061 | 37040 | — | — |
| Quarto d'Altino | 027031 | 30020 | — | — |
| Quinto di Treviso | 026064 | 31055 | — | — |
| Quinto Vicentino | 024083 | 36050 | — | — |
| Recoaro Terme | 024084 | 36076 | — | — |
| Refrontolo | 026065 | 31020 | — | — |
| Resana | 026066 | 31023 | — | — |
| Revine Lago | 026067 | 31020 | — | — |
| Riese Pio X | 026068 | 31039 | — | — |
| Rivamonte Agordino | 025043 | 32020 | — | — |
| Rivoli Veronese | 023062 | 37010 | — | — |
| Roana | 024085 | 36010 | — | — |
| Rocca Pietore | 025044 | 32023 | — | — |
| Romano d'Ezzelino | 024086 | 36060 | — | — |
| Roncà | 023063 | 37030 | — | — |
| Roncade | 026069 | 31056 | — | — |
| Ronco all'Adige | 023064 | 37055 | — | — |
| Rosà | 024087 | 36027 | — | — |
| Rosolina | 029040 | 45010 | — | — |
| Rossano Veneto | 024088 | 36028 | — | — |
| Rotzo | 024089 | 36010 | — | — |
| Roverchiara | 023065 | 37050 | — | — |
| Roverè Veronese | 023067 | 37028 | — | — |
| Roveredo di Guà | 023066 | 37040 | — | — |
| Rovigo | 029041 | 45100 | — | — |
| Rovolon | 028071 | 35030 | — | — |
| Rubano | 028072 | 35030 | — | — |
| Saccolongo | 028073 | 35030 | — | — |
| Salara | 029042 | 45030 | — | — |
| Salcedo | 024090 | 36040 | — | — |
| Salgareda | 026070 | 31040 | — | — |
| Salizzole | 023068 | 37056 | — | — |
| Salzano | 027032 | 30030 | — | — |
| San Bellino | 029043 | 45020 | — | — |
| San Biagio di Callalta | 026071 | 31048 | — | — |
| San Bonifacio | 023069 | 37047 | — | — |
| San Donà di Piave | 027033 | 30027 | — | — |
| San Fior | 026072 | 31020 | — | — |
| San Giorgio delle Pertiche | 028075 | 35010 | — | — |
| San Giorgio in Bosco | 028076 | 35010 | — | — |
| San Giovanni Ilarione | 023070 | 37035 | — | — |
| San Giovanni Lupatoto | 023071 | 37057 | — | — |
| San Gregorio nelle Alpi | 025045 | 32030 | — | — |
| San Martino Buon Albergo | 023073 | 37036 | — | — |
| San Martino di Lupari | 028077 | 35018 | — | — |
| San Martino di Venezze | 029044 | 45030 | — | — |
| San Mauro di Saline | 023074 | 37030 | — | — |
| San Michele al Tagliamento | 027034 | 30028 | — | — |
| San Nicolò di Comelico | 025046 | 32040 | — | — |
| San Pietro di Cadore | 025047 | 32040 | — | — |
| San Pietro di Feletto | 026073 | 31020 | — | — |
| San Pietro di Morubio | 023075 | 37050 | — | — |
| San Pietro in Cariano | 023076 | 37029 | — | — |
| San Pietro in Gu | 028078 | 35010 | — | — |
| San Pietro Mussolino | 024094 | 36070 | — | — |
| San Pietro Viminario | 028079 | 35020 | — | — |
| San Polo di Piave | 026074 | 31020 | — | — |
| San Stino di Livenza | 027036 | 30029 | — | — |
| San Tomaso Agordino | 025049 | 32020 | — | — |
| San Vendemiano | 026076 | 31020 | — | — |
| San Vito di Cadore | 025051 | 32046 | — | — |
| San Vito di Leguzzano | 024096 | 36030 | — | — |
| San Zeno di Montagna | 023079 | 37010 | — | — |
| San Zenone degli Ezzelini | 026077 | 31020 | — | — |
| Sandrigo | 024091 | 36066 | — | — |
| Sanguinetto | 023072 | 37058 | — | — |
| Sant'Ambrogio di Valpolicella | 023077 | 37015 | — | — |
| Sant'Angelo di Piove di Sacco | 028082 | 35020 | — | — |
| Sant'Anna d'Alfaedo | 023078 | 37020 | — | — |
| Sant'Elena | 028083 | 35040 | — | — |
| Sant'Urbano | 028084 | 35040 | — | — |
| Santa Caterina d'Este | 028108 | 35049 | — | — |
| Santa Giustina | 025048 | 32035 | — | — |
| Santa Giustina in Colle | 028080 | 35010 | — | — |
| Santa Lucia di Piave | 026075 | 31025 | — | — |
| Santa Maria di Sala | 027035 | 30036 | — | — |
| Santo Stefano di Cadore | 025050 | 32045 | — | — |
| Santorso | 024095 | 36014 | — | — |
| Saonara | 028085 | 35020 | — | — |
| Sarcedo | 024097 | 36030 | — | — |
| Sarego | 024098 | 36040 | — | — |
| Sarmede | 026078 | 31026 | — | — |
| Schiavon | 024099 | 36060 | — | — |
| Schio | 024100 | 36015 | — | — |
| Scorzè | 027037 | 30037 | — | — |
| Sedico | 025053 | 32036 | — | — |
| Segusino | 026079 | 31040 | — | — |
| Selva di Cadore | 025054 | 32020 | — | — |
| Selva di Progno | 023080 | 37030 | — | — |
| Selvazzano Dentro | 028086 | 35030 | — | — |
| Seren del Grappa | 025055 | 32030 | — | — |
| Sernaglia della Battaglia | 026080 | 31020 | — | — |
| Setteville | 025075 | 32038 | — | — |
| Silea | 026081 | 31057 | — | — |
| Soave | 023081 | 37038 | — | — |
| Solagna | 024101 | 36020 | — | — |
| Solesino | 028087 | 35047 | — | — |
| Sommacampagna | 023082 | 37066 | — | — |
| Sona | 023083 | 37060 | — | — |
| Sorgà | 023084 | 37060 | — | — |
| Sospirolo | 025056 | 32037 | — | — |
| Sossano | 024102 | 36040 | — | — |
| Soverzene | 025057 | 32010 | — | — |
| Sovizzo | 024128 | 36049 | — | — |
| Sovramonte | 025058 | 32030 | — | — |
| Spinea | 027038 | 30038 | — | — |
| Spresiano | 026082 | 31027 | — | — |
| Stanghella | 028088 | 35048 | — | — |
| Stienta | 029045 | 45039 | — | — |
| Stra | 027039 | 30039 | — | — |
| Susegana | 026083 | 31058 | — | — |
| Taglio di Po | 029046 | 45019 | — | — |
| Taibon Agordino | 025059 | 32027 | — | — |
| Tambre | 025060 | 32010 | — | — |
| Tarzo | 026084 | 31020 | — | — |
| Teglio Veneto | 027040 | 30025 | — | — |
| Teolo | 028089 | 35037 | — | — |
| Terrassa Padovana | 028090 | 35020 | — | — |
| Terrazzo | 023085 | 37040 | — | — |
| Tezze sul Brenta | 024104 | 36056 | — | — |
| Thiene | 024105 | 36016 | — | — |
| Tombolo | 028091 | 35019 | — | — |
| Tonezza del Cimone | 024106 | 36040 | — | — |
| Torre di Mosto | 027041 | 30020 | — | — |
| Torrebelvicino | 024107 | 36036 | — | — |
| Torreglia | 028092 | 35038 | — | — |
| Torri del Benaco | 023086 | 37010 | — | — |
| Torri di Quartesolo | 024108 | 36040 | — | — |
| Trebaseleghe | 028093 | 35010 | — | — |
| Trecenta | 029047 | 45027 | — | — |
| Tregnago | 023087 | 37039 | — | — |
| Trevenzuolo | 023088 | 37060 | — | — |
| Trevignano | 026085 | 31040 | — | — |
| Treviso | 026086 | 31100 | — | — |
| Tribano | 028094 | 35020 | — | — |
| Trissino | 024110 | 36070 | — | — |
| Urbana | 028095 | 35040 | — | — |
| Val di Zoldo | 025073 | 32012 | — | — |
| Val Liona | 024123 | 36044 | — | — |
| Valbrenta | 024125 | 36029 | — | — |
| Valdagno | 024111 | 36078 | — | — |
| Valdastico | 024112 | 36040 | — | — |
| Valdobbiadene | 026087 | 31049 | — | — |
| Valeggio sul Mincio | 023089 | 37067 | — | — |
| Vallada Agordina | 025062 | 32020 | — | — |
| Valle di Cadore | 025063 | 32040 | — | — |
| Valli del Pasubio | 024113 | 36030 | — | — |
| Vazzola | 026088 | 31028 | — | — |
| Vedelago | 026089 | 31050 | — | — |
| Veggiano | 028096 | 35030 | — | — |
| Velo d'Astico | 024115 | 36010 | — | — |
| Velo Veronese | 023090 | 37030 | — | — |
| Venezia | 027042 | 30121, 30122, 30123, 30124, 30125, 30126, 30132, 30133, 30135, 30141, 30142, 30171, 30172, 30173, 30174, 30175, 30176 | — | — |
| Verona | 023091 | 37121, 37122, 37123, 37124, 37125, 37126, 37127, 37128, 37129, 37131, 37132, 37133, 37134, 37135, 37136, 37137, 37138, 37139, 37141, 37142 | — | — |
| Veronella | 023092 | 37040 | — | — |
| Vescovana | 028097 | 35040 | — | — |
| Vestenanova | 023093 | 37030 | — | — |
| Vicenza | 024116 | 36100 | — | — |
| Vidor | 026090 | 31020 | — | — |
| Vigasio | 023094 | 37068 | — | — |
| Vigo di Cadore | 025065 | 32040 | — | — |
| Vigodarzere | 028099 | 35010 | — | — |
| Vigonovo | 027043 | 30030 | — | — |
| Vigonza | 028100 | 35010 | — | — |
| Villa Bartolomea | 023095 | 37049 | — | — |
| Villa del Conte | 028101 | 35010 | — | — |
| Villa Estense | 028102 | 35040 | — | — |
| Villadose | 029048 | 45010 | — | — |
| Villafranca di Verona | 023096 | 37069 | — | — |
| Villafranca Padovana | 028103 | 35010 | — | — |
| Villaga | 024117 | 36021 | — | — |
| Villamarzana | 029049 | 45030 | — | — |
| Villanova del Ghebbo | 029050 | 45020 | — | — |
| Villanova di Camposampiero | 028104 | 35010 | — | — |
| Villanova Marchesana | 029051 | 45030 | — | — |
| Villaverla | 024118 | 36030 | — | — |
| Villorba | 026091 | 31020 | — | — |
| Vittorio Veneto | 026092 | 31029 | — | — |
| Vo' | 028105 | 35030 | — | — |
| Vodo Cadore | 025066 | 32040 | — | — |
| Volpago del Montello | 026093 | 31040 | — | — |
| Voltago Agordino | 025067 | 32020 | — | — |
| Zanè | 024119 | 36010 | — | — |
| Zenson di Piave | 026094 | 31050 | — | — |
| Zermeghedo | 024120 | 36050 | — | — |
| Zero Branco | 026095 | 31059 | — | — |
| Zevio | 023097 | 37059 | — | — |
| Zimella | 023098 | 37040 | — | — |
| Zoppè di Cadore | 025069 | 32010 | — | — |
| Zovencedo | 024121 | 36020 | — | — |
| Zugliano | 024122 | 36030 | — | — |

## CAP attribution and reuse

CAP associations are derived from **Garda Informatica, Database Comuni Italiani**, version 11 September 2026. Its original `README.txt` is preserved in the linked ZIP and identifies Garda Informatica as author and the database as MIT-licensed. This document joins its CAP associations to the retained ISTAT-derived municipality registry and sorts municipalities by name. The [MIT license](https://opensource.org/license/mit/) governs reuse of the CAP dataset; the publisher provides it without warranty.

### MIT license notice for the CAP dataset

The retained publisher notice identifies Garda Informatica as the author of *Database Comuni Italiani* and declares the database MIT-licensed. The following license text accompanies the CAP associations reproduced here.

Copyright (c) Garda Informatica

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

### Development municipality publication — 2026-10-02

[Manual development publication](operations/development-publication.md) is a separate,
revocable municipality choice for API/MCP consultation of municipal data and applicable
CFR products. It does not alter the dated source acceptance assessment above or complete
missing reviews, observational cases or regressions. Responses retain pending source
states and identify development publication explicitly. Staging and production ignore
these choices and retain their acceptance and release gates.
