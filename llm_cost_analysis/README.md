# LLM correction cost analysis

Updated: 2026-11-14.

- 2026-10-18_costs.csv: original report, moved without changing its contents. Its original serialization contains literal backslash-r/backslash-n separators; the current CSV uses real CRLF line endings.
- llm-curation-costs.csv: current analysis, 28 editions. Analysis dated 2026-11-14.

Sources: ../docs/TRANSCRIPTION_STATUS.md; public Huma-Num dataset image metadata (uniqueOnly=true), retrieved 2026-11-14; https://developers.openai.com/api/docs/pricing (checked 2026-11-14).

All costs are USD. OpenAI costs are API-equivalent estimates using Standard GPT-5.6 Sol short-context rates: $4 per million uncached input tokens, $0.40 per million cached input tokens, and $20 per million output tokens. Cached input is included in the input total, so it is subtracted before charging uncached input. Cache-creation tokens are zero in these logs. Actual subscription charges and per-request context/tier premiums cannot be reconstructed from the aggregated logs.

Both $/facsimile page columns divide cost by the total facsimile page count, not the provider's corrected-page count. Blank cells mean unknown or not applicable, not zero. Provider page counts now include explicitly labelled estimates. The main CSV includes OpenAI/Fable Count status and Count basis, plus Dataset ID. See page-count-evidence.csv for the detailed provider evidence and page-count-investigation.md for the code/API investigation. Request-derived estimates assume one round and no repeated pages across runs; provider totals need not reconcile to full facsimile pages. Fable is the model label supplied by the document; no underlying model/version is inferred.

Scope retains the original report's correction editions and adds six new editions with correction usage. Direct OCR runs labelled “Skipped Kraken” (Venice_1505, Cambridge_1655, Basel_1550, Venice_1543, Würzburg_1661) are excluded. Rome_1574_vol2 has only Claude Opus usage and is excluded from the requested OpenAI/Fable columns. In-progress correction entries without provider usage are excluded.

Edition notes and provider count qualifications are recorded in page-count-evidence.csv and page-count-investigation.md.

Page-count API sources:

- Antwerp_1654: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_9ni7al/images?uniqueOnly=true
- Basel_1537: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_yp3x8a/images?uniqueOnly=true
- Basel_1562: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_emvd28/images?uniqueOnly=true
- Glasgow_1756: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_ii8nbl/images?uniqueOnly=true
- Kiel_and_Leipzig_1699: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_l2ztyi/images?uniqueOnly=true
- London_1570: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_9lpdf8/images?uniqueOnly=true
- Lyon_1557: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_wqmfll/images?uniqueOnly=true
- Lyon_1603: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_69jfii/images?uniqueOnly=true
- Lyon_1672: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_71gygq/images?uniqueOnly=true
- Nuremberg_1821: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_tjokpg/images?uniqueOnly=true
- Oxford_1703: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_ej5frn/images?uniqueOnly=true
- Paris_1566: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_sfmbfc/images?uniqueOnly=true
- Paris_1615: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_57vh5j/images?uniqueOnly=true
- Paris_1634: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_w5j2g1/images?uniqueOnly=true
- Paris_1639: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_lysuca/images?uniqueOnly=true
- Paris_1667: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_j0l55w/images?uniqueOnly=true
- Paris_1685: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_fa95xr/images?uniqueOnly=true
- Paris_1794: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_fcnxho/images?uniqueOnly=true
- Paris_1814_vol1: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_mz3oph/images?uniqueOnly=true
- Paris_1814_vol2: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_a2q3yy/images?uniqueOnly=true
- Paris_1814_vol3: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_9s7hg3/images?uniqueOnly=true
- Pesaro_1572: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_g17zx1/images?uniqueOnly=true
- Rome_1574_vol1: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_0m73by/images?uniqueOnly=true
- Seville_1576: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_uwclpr/images?uniqueOnly=true
- Strasbourg_1564: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_ypncw4/images?uniqueOnly=true
- Strasbourg_1566: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_ilx0l7/images?uniqueOnly=true
- The_Hague_1758: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_9m13nh/images?uniqueOnly=true
- Venice_1482: https://euclides.huma-num.fr/commentaria/api/v1/datasets/ds_6vegrr/images?uniqueOnly=true

## Edition list page counts

The page-count columns in items_print_elements_or_selected_studies.csv were added on 2026-11-14. All 356 rows and the original 15 columns are preserved. Page counts include the complete PDF scans, including front matter and blank pages. Multi-volume counts are summed only when every identified/expected volume has a count. Alternative scans of a volume are not added together. The largest identified volume number is used alongside the existing volumes column, which understates some multi-volume editions.

Sources: Huma-Num /commentaria/api/v1/datasets, /facsimilies, /shelfmarks, and /datasets/{id}/images?uniqueOnly=true. Open rights are checked against linked shelfmark copyright or the existing local facsimile manifest. For open PDFs without image datasets, counts are matched by exact PDF filename to https://raw.githubusercontent.com/Euclides-EM/elements-facsimile/refs/heads/main/page_counts.csv. Matching local PDFs are used when the filename and byte size match the server metadata; those rows identify this source explicitly. Huma-Num's PDF download endpoint requires authentication, so some counts remain unavailable.

Results: 150 open rights not confirmed; 143 no hosted facsimile; 52 available; 10 page count unavailable; 1 incomplete volume coverage. Blank page counts are not zero. “Open rights not confirmed” means an edition has a hosted scan but the linked metadata does not establish an open scan. It does not establish that the scan is restricted.
