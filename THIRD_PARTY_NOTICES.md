# Third-party data notices

These notices apply to the files listed below and to the derived municipality and CAP tables in `docs/coverage.md`. They do not change the GPL-3.0-only license of original IWA code or license other source material.

## ISTAT: municipality registry

Files: `docs/nazionale/registri/comuni-italia.csv` and its identical embedded copy `internal/domain/territorial_municipalities.csv`; municipality names and ISTAT codes in `docs/coverage.md`.

Source and attribution: Istituto nazionale di statistica (ISTAT), [Codici statistici delle unità amministrative territoriali](https://www.istat.it/classificazione/codici-dei-comuni-delle-province-e-delle-regioni/), captured 15 September 2026. [ISTAT's open-data terms](https://www.istat.it/dati/open-data/) permit redistribution with source attribution under Creative Commons Attribution 4.0. IWA selected and normalized registry fields and joined them with Garda Informatica CAP associations. This derived table is not an official ISTAT publication, and ISTAT does not endorse IWA.

## Garda Informatica: Database Comuni Italiani

File: `docs/prerequisiti-mvp/evidenze/cap-garda-2026-09-11.zip`.

Source: [Database Comuni Italiani](https://www.gardainformatica.it/database-comuni-italiani), version dated 11 September 2026 in the retained archive. Author: Garda Informatica. License: MIT, as declared by the publisher and in the archive's `README.txt`. The ZIP is reproduced as captured; IWA's municipality/CAP selection and reconciliation are described separately in `docs/coverage.md`.

Copyright (c) Garda Informatica

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
