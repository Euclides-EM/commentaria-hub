# The Paratext Observatory

Analysis companion workspace for the report:

**The Place of Euclid's Elements in Early Modern Mathematical Print: A Title-Page Ecology**

This workspace supports a title-page ecology analysis centered on the metadata-defined Euclid's *Elements* corpus. The broader mathematical title-page corpus is used as the surrounding ecology: comparison field, control, and source of boundary cases.

## Start Here

For the current state of the work, read:

1. `CURRENT_STATUS.md`
2. `report/REPORT_PRINT_PACKET.html` or `report/REPORT_PRINT_PACKET.md`
3. `orientation/DATA_MAP.md`
4. `orientation/ANALYSIS_TERMS.md`

The printable report packet is:

`report/REPORT_PRINT_PACKET.html`

The editable report sources are:

- `report/REPORT_DRAFT.md`
- `report/REPORT_APPENDICES.md`

## Launch The Companion

When starting a new analysis session, paste this:

```text
Please act as my analysis companion for The Paratext Observatory.

Read:
- tps_title_page_extraction_review/analysis_companion/CURRENT_STATUS.md
- tps_title_page_extraction_review/analysis_companion/README.md
- tps_title_page_extraction_review/analysis_companion/orientation/COMPANION_BRIEF.md
- tps_title_page_extraction_review/analysis_companion/orientation/CORRECTIONS_AND_NEXT_ANALYSIS_AGENDA.md
- tps_title_page_extraction_review/analysis_companion/orientation/ANALYSIS_TERMS.md
- tps_title_page_extraction_review/analysis_companion/orientation/ANALYSIS_PLAN.md
- tps_title_page_extraction_review/analysis_companion/orientation/DATA_MAP.md

Then help me continue the title-page corpus analysis question by question. Start from the metadata-defined Elements corpus and use the broader mathematical corpus as ecology/context. Do not force the old presentation argument onto the new corpus.
```

## Directory Map

| Path | Purpose |
|---|---|
| `CURRENT_STATUS.md` | Current handoff, latest synthesis, and next moves. |
| `orientation/` | Companion brief, analysis plan, terms, data map, question log, corrections, and working argument notes. |
| `exploration/phase_notes/` | Chronological research trail from the exploratory phase. Useful for provenance, not the polished argument. |
| `derived_data/` | Generated analysis-ready CSVs and matrices used by the exploratory notes and report. |
| `report/` | Polished report, appendices, print packet, report figures, report tables, and report-specific scripts. |
| `scripts/` | Broader analysis scripts that generated derived-data files outside the report-specific pipeline. |

## Core Data Sources

Core metadata lives in:

`../../ocrflow/store/items_metadata/`

Use this directory for bibliographic metadata, reprint clusters, corpus membership, transcriptions, translations, shelfmarks/images, and Elements-specific metadata.

Most important metadata files:

- `items_print.csv`: edition key, title, year, city, language, author/editor, publisher, format, volumes, USTC ID, notes, diagram flag.
- `metadata_elements_print.csv`: manual metadata-defined Elements corpus membership and Elements-specific fields.
- `cluster_items.csv` and `clusters.csv`: reprint cluster information.
- `paratext_transcriptions.csv`: title-page and paratext transcriptions.
- `translations.csv`: translated title/imprint/paratext snippets.

## Current Guardrails

- The main historical object is the metadata-defined Elements corpus.
- Title pages that mention Euclid or "elements" outside that corpus are comparison evidence, not corpus-definition evidence.
- Reprints should inherit the representative classification unless reprint variation is the object of study.
- Sparse or silent title pages must be controlled against place, language, period, format, and local title-page fashion.
- Same-author comparisons are useful but biased because the corpus does not contain full author bibliographies; controlled city/region/language/format/period comparisons are stronger next steps.
