# Current Status

Last updated: 2026-06-13

## Current Deliverable

The current printable report packet is ready for reading and annotation:

`report/REPORT_PRINT_PACKET.html`

Editable report sources:

- `report/REPORT_DRAFT.md`
- `report/REPORT_APPENDICES.md`

Rebuild the printable packet after edits:

```bash
python3 tps_title_page_extraction_review/analysis_companion/report/scripts/build_print_packet.py
```

## Current Historical Frame

The analysis is centered on the **metadata-defined Elements corpus**. The broader title-page corpus is the surrounding mathematical ecology, not the final object of the argument.

Leading report question:

**What place did the metadata-defined Elements corpus occupy in the broader ecology of early modern mathematical print, and how did title pages construct that place through social address, intellectual values, and acts of textual/mathematical mediation?**

Current synthesis:

The Elements functioned as a canonical proof-corpus whose authority made it unusually available for mediation. Title pages make Euclid valuable by advertising acts such as correction, translation, commentary, explanation, demonstration, selection, augmentation, reordering, reform, and practical adaptation. The surrounding mathematical ecology more often advertises direct practical or procedural application: operations, instruments, examples, problems, visual/material work, professional use, and applied procedure.

## Most Important Current Results

- The Elements/non-Elements distinction is not simply theoretical versus practical. It is better described as authoritative textual mediation versus more direct practical/procedural application.
- Practical or pedagogical Euclid is usually not anti-canonical. Many practical/pedagogical Elements editions retain strong Euclid/book identity and ancient authority.
- The metadata-defined Elements corpus is internally diverse: learned apparatus, pedagogical/methodized Euclid, practical-pedagogical Euclid, humanist/vernacular transfer, sparse authoritative presentation, and reconstructive/reformed Euclid.
- Book coverage matters historically. First-six-books editions, first-six-plus-solid-geometry editions, near-complete/expanded editions, and selected later-book editions advertise different kinds of Euclid.
- Named mathematical parts matter. Elements title pages foreground demonstrations/proofs, scholia/commentary, principles, propositions, theorems, and enunciations more than non-Elements title pages.
- Figures and diagrams are not Elements-specific by themselves. Their function changes by context: proof apparatus, learned apparatus, edition furnishing, pedagogy, or practical operation.
- Format is historical evidence, not only a caveat. Smaller Elements formats are not simply thinner, more silent, or more popular; octavos and duodecimos have high average rich-claim counts, and duodecimos look especially pedagogical-institutional.
- Sparse title pages must be interpreted with controls for city, language, period, format, book coverage, and local title-page fashion.

## Directory Map

| Path | Purpose |
|---|---|
| `README.md` | Main workspace map and companion launch prompt. |
| `orientation/` | Companion brief, analysis plan, data map, terms, corrections, question log, and working argument/case notes. |
| `exploration/phase_notes/` | Chronological exploratory analysis trail. Useful for provenance and reopening specific questions. |
| `derived_data/` | Generated analysis-ready CSVs and matrices. |
| `report/` | Polished report, appendices, print packet, report figures, report tables, and report-specific scripts. |
| `scripts/` | Broader analysis scripts that generated derived-data files. |

## Data Sources

The subject-classification run completed and wrote preview + DB results:

`../data/subject_classification/full_run/full_subject_classification_representatives_preview.csv`

Main analysis-ready files:

- `derived_data/DERIVED_DATA_SUMMARY.md`
- `derived_data/title_page_analysis_matrix.csv`
- `derived_data/metadata_elements_corpus_ecology_matrix.csv`
- `derived_data/metadata_elements_natural_modes_matrix_with_format.csv`
- `derived_data/deductive_parts_cases.csv`

Core bibliographic and corpus metadata lives in:

`../../ocrflow/store/items_metadata/`

Important metadata files:

- `items_print.csv`: general print metadata, including format.
- `metadata_elements_print.csv`: metadata-defined Elements corpus membership and Elements-specific fields.
- `cluster_items.csv` and `clusters.csv`: reprint cluster information.
- `paratext_transcriptions.csv`: title-page and paratext transcriptions.
- `translations.csv`: translated snippets for title-page checking.

QA note: `bib-96` is missing one subject decision, `Cosmography`, in the classifier output. See `derived_data/classification_qa_warnings.csv`.

## Current Best Next Research Moves

1. Read and annotate `report/REPORT_PRINT_PACKET.html`.
2. Turn annotations into targeted follow-up questions.
3. Prioritize controlled local comparisons by city/region, language, format, period/date window, subject zone, and intersections of these controls.
4. Close-read major cited cases against title-page transcriptions/images before using them as paper evidence.

## Guardrails

- The central corpus is metadata-defined Elements, not title pages that happen to mention Euclid or "Elements."
- Reprints should inherit representative classification unless reprint variation is intentionally being studied.
- Same-author comparisons are biased because the corpus does not contain complete bibliographies.
- Rich tags are useful for navigation and hypothesis-building, but major claims should be supported by close reading.
- Do not organize the argument around legacy textual-family labels.

## Do Not Use As Main Analysis Data

`../data/subject_classification/partial_155_archive/`

This is only the old partial classification archive.
