# Data Map

All paths are relative to:

`/Users/mia/dev/personal/elements-dh/tps_title_page_extraction_review/`

## Main TPS Data

### Latest Larger-Corpus TPS Extraction

`data/latest_larger_corpus/printed_missing_tps_v8_preview.csv`

This is the newer larger-corpus title-page segmentation output. It contains non-empty extracted feature rows.

Guide:

`data/latest_larger_corpus/README.md`

### Older Reviewed Elements-Oriented TPS Extraction

`data/reviewed_elements_v8/v8_results_preview_reviewed.csv`

This is the older manually reviewed V8 output for 217 editions.

Guide:

`data/reviewed_elements_v8/README.md`

## Subject Classification

### Pending Full Classification Run

Folder:

`data/subject_classification/full_run/`

Corrected run commands:

`data/subject_classification/full_run/RUN_COMMANDS.md`

Important files:

- `representative_title_page_corpus_needing_classification_keys.txt`: 695 keys to classify now.
- `representative_title_page_corpus_keys.txt`: 843 representative title-page corpus keys after reprint deduping.
- `representative_title_page_corpus_already_classified_155_overlap_keys.txt`: 148 representative keys already classified in the old handoff.
- `title_page_to_classification_representative_map.csv`: maps 909 title-page keys to classification representative keys.
- `wrong_909_dry_run_attempt/`: archived incorrect dry-run attempt; ignore for analysis.

Expected output after the corrected run:

`data/subject_classification/full_run/full_subject_classification_representatives_preview.csv`

### Partial Old Classification Archive

`data/subject_classification/partial_155_archive/`

This covers only 155 editions. It is provenance, not the main analysis source.

## Metadata

All core bibliographic, corpus, transcription, visual, and relationship metadata lives in:

`../ocrflow/store/items_metadata/`

Treat this directory as the metadata home base. Before inventing a join or asking where a field lives, inspect this folder.

### Print Metadata

`../ocrflow/store/items_metadata/items_print.csv`

Contains edition key, title, year, city, language, author/editor, publisher, format, volumes, USTC ID, notes, and diagram flag.

Format note:

The `format` column uses bibliographic shorthand such as `2` = folio, `4` = quarto, `8` = octavo, `12` = duodecimo. Use it as a control when analyzing title-page density, audience, and school/institutional use.

### Elements-Specific Print Metadata

`../ocrflow/store/items_metadata/metadata_elements_print.csv`

Contains Elements-related metadata such as covered books, additional content, and Wardhaugh classification. Use this for Elements-specific questions, not for the whole mathematical corpus.

Important caution:

Use `elements_books` and `additional_content` actively. Treat `wardhaugh_classification` as weak legacy context only, or ignore it. It should not organize the main analysis.

### Reprint Clusters

`../ocrflow/store/items_metadata/cluster_items.csv`

Used to dedupe reprints to representative classification keys.

`../ocrflow/store/items_metadata/clusters.csv`

Cluster metadata. Reprint clusters have type `reprint`.

### Title-Page Transcriptions

`../ocrflow/store/items_metadata/paratext_transcriptions.csv`

Contains title-page and other paratext transcriptions used by feature extraction.

### Existing Manual/LLM Title-Page Metadata

`../ocrflow/store/items_metadata/title_page.csv`

Older title-page metadata table with many manually/LLM-populated columns. Useful for comparison and spot checks, but do not confuse it with the newer TPS feature extraction CSVs.

### Translations

`../ocrflow/store/items_metadata/translations.csv`

Contains translated title/imprint/paratext snippets. Useful when inspecting non-English title pages.

### Visual And Paratext Metadata

- `../ocrflow/store/items_metadata/visual_elements.csv`
- `../ocrflow/store/items_metadata/visual_elements_examples.csv`
- `../ocrflow/store/items_metadata/frontispiece.csv`
- `../ocrflow/store/items_metadata/dotted_lines.csv`

Use these when we need visual/title-page context beyond text segmentation.

### Other Useful Metadata

- `../ocrflow/store/items_metadata/shelfmarks.csv`: scan URLs, shelfmarks, title-page images.
- `../ocrflow/store/items_metadata/bibliography.csv`: bibliography links/citations.
- `../ocrflow/store/items_metadata/corpuses.csv`: corpus/study membership.
- `../ocrflow/store/items_metadata/reviews.csv`: review/verification markers.
- `../ocrflow/store/items_metadata/cities.csv`: city metadata.
- `../ocrflow/store/items_metadata/mentions.csv` and `mentions_sources.csv`: entity/mention metadata.

## Presentation Context

Old presentation:

`presentation_context/Title Page Presentation.pdf`

USTC abstract:

`presentation_context/USTC 2026 Abstract.pdf`

## Existing Reports

First-pass larger-corpus reassessment:

`reports/larger_corpus_reassessment.md`

V8 extraction reports:

- `reports/v8_analysis_report.md`
- `reports/v8_policy_decisions.md`
- `reports/v8_interactive_review_decisions.md`

## Important Warnings

- The subject classification is multi-label. Do not collapse it into one genre unless needed for a specific chart.
- Reprints should inherit the representative classification unless we intentionally study reprint variation.
- The old 155-edition classification is not enough for the main analysis.
- The older 217-edition TPS file and the newer larger-corpus TPS file have different roles; do not blend their denominators casually.
- The richer Phase 4C/4D social and intellectual tags are exploratory. Treat them as a way to find patterns and examples, not as publication-ready classifications without spot-checking.
- Early broad-corpus analyses are ecological background. The main historical argument should return to the metadata-defined Elements corpus.
- Do not use Wardhaugh/textual-family labels as the organizing taxonomy for the Elements argument.
