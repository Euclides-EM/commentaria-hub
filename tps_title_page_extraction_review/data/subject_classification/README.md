# Subject Classification

This folder is prepared for a fresh subject-classification run over the full title-page corpus.

## Current State

The existing subject classification is partial. It covers only 155 editions, not the whole title-page corpus.

Those old partial files have been moved to:

`partial_155_archive/`

Do not use that archive for the main historical analysis except as provenance or comparison material.

## Full-Run Inputs

Use the files in:

`full_run/`

Recommended key file:

`full_run/representative_title_page_corpus_needing_classification_keys.txt`

This contains:

- 695 representative title-page corpus keys that still need classification.
- Reprints are deduped through `ocrflow/store/items_metadata/cluster_items.csv`.
- Keys already present in the old 155-edition classification handoff are excluded.

The full accounting is:

- 909 title-page corpus keys.
- 843 representative keys after reprint-cluster deduping.
- 148 representative keys already classified in the old handoff.
- 695 representative keys to run now.

All 843 representative keys are present in `ocrflow/store/items_metadata/items_print.csv`.

Mapping file:

`full_run/title_page_to_classification_representative_map.csv`

This maps each of the 909 title-page keys to the representative classification key that should be used for analysis.

The old incorrect 909-key dry-run preview has been archived in:

`full_run/wrong_909_dry_run_attempt/`

## Recommended Command

Run this from the repo root:

```sh
cd /Users/mia/dev/personal/elements-dh/ocrflow

go run ./cmd/edition-classification \
  -keys-file ../tps_title_page_extraction_review/data/subject_classification/full_run/representative_title_page_corpus_needing_classification_keys.txt \
  -output-csv ../tps_title_page_extraction_review/data/subject_classification/full_run/full_subject_classification_representatives_preview.csv \
  -checkpoint-file ../tps_title_page_extraction_review/data/subject_classification/full_run/full_subject_classification_representatives_preview.csv.done \
  -resume=false
```

This will run the latest classifier revision:

`6a0d47e3-f472-4b63-a6f5-67c693a0adf9`

and write results to the database. It also writes a preview CSV for later analysis checks.

If the run stops midway, resume with the same command but omit `-resume=false`:

```sh
cd /Users/mia/dev/personal/elements-dh/ocrflow

go run ./cmd/edition-classification \
  -keys-file ../tps_title_page_extraction_review/data/subject_classification/full_run/representative_title_page_corpus_needing_classification_keys.txt \
  -output-csv ../tps_title_page_extraction_review/data/subject_classification/full_run/full_subject_classification_representatives_preview.csv \
  -checkpoint-file ../tps_title_page_extraction_review/data/subject_classification/full_run/full_subject_classification_representatives_preview.csv.done
```

## After The Run

The output CSV should have columns:

`edition_id,feature_id,source_id,source_revision,source_name,value`

The `value` field should contain packed subject categories, for example:

`Architecture::unrelated, Arithmetic::primary, ...`

After the run completes, parse that preview CSV into:

- full long format: one row per edition-subject pair;
- positive long format: only `primary`, `secondary`, and `unknown`;
- matrix format: one row per edition, 20 subject columns;
- joined TPS format: classification matrix plus title-page feature summaries.

Those derived files should replace the temporary partial-155 analysis files.
