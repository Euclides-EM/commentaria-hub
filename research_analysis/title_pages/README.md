# Title-Page Analysis

## Start Here

Read [`results/report/report.md`](results/report/report.md). It is the canonical analysis: *The Place of Euclid's Elements in Early Modern Mathematical Print: A Title-Page Ecology*.

Its central conclusion is that the metadata-defined *Elements* corpus functioned as an authoritative demonstrative corpus repeatedly made usable through translation, correction, commentary, selection, explanation, new demonstrations, visual furnishing, institutional authority, and practical adaptation. The strongest contrast is not theoretical versus practical mathematics, but mediation of an authoritative text versus more direct advertisement of practical or procedural application.

The report's appendices, cited tables, and cited figures are beside it. The final audit also incorporated controlled same-person comparisons, Jesuit-marked corpus contrasts, and print-geography findings directly into the report.

Returning after a long break: use [`DATA_GUIDE.md`](DATA_GUIDE.md) to choose the correct dataset and [`RUNBOOK.md`](RUNBOOK.md) to rebuild or extend the analysis.

## Evidence

`data/raw/` contains the retained extraction/classification evidence:

- `title_page_features_larger_corpus.csv`: larger-corpus title-page extraction;
- `title_page_features_reviewed_elements.csv`: manually reviewed Elements-oriented extraction;
- `subject_classifications_new_representatives.csv`: newly classified representative rows (13,899 decisions; the earlier reviewed overlap is retained in the sibling `edition_classification` area);
- `title_page_to_subject_representative_map.csv`: reprint/representative mapping.

`data/analysis_ready/` contains the five joined matrices consumed by the retained scripts. They are retained because the original exploratory workspace did not preserve one complete script that rebuilds every join from raw inputs. They are therefore the reproducible input boundary for this analysis, not disposable output.

Canonical repository metadata used by the scripts remains in `ocrflow/store/items_metadata/`, especially `items_print.csv`, `metadata_elements_print.csv`, `cluster_items.csv`, `clusters.csv`, and `title_page.csv`.

## Reproduction

From the repository root, install the small analysis environment listed in `scripts/requirements.txt`, then run the scripts in `scripts/report/` to rebuild report tables and figures, including controlled same-person comparisons, print-geography tables, and Jesuit comparisons.

Generated scripts may emit additional diagnostic tables; only outputs cited by the report belong in this curated folder.

## Guardrails

- The central corpus is metadata-defined *Elements*, not every title page mentioning Euclid or “elements.”
- Subject classification is multi-label.
- Reprints inherit representative classification unless reprint variation is the object of study.
- Title pages show public presentation, not the full intellectual contents of a book.
- Sparse title pages require controls for place, language, period, format, book coverage, and local title-page fashion.
- Rich social/intellectual tags are navigational classifications; major claims still require close reading.
