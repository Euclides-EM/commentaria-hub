# Full Subject Classification Run Commands

## Stop The Old Attempt First

If the old 909-key command is still running, stop it with Ctrl+C.

That attempt used every title-page key and included `-dry-run`, so it only wrote a preview CSV and did not store database results. Its small preview/checkpoint files have been archived in:

`wrong_909_dry_run_attempt/`

## Corrected Run

Use the representative, not-yet-classified key list:

`representative_title_page_corpus_needing_classification_keys.txt`

This list has 695 keys:

- 909 title-page corpus keys
- collapsed to 843 representative keys after reprint-cluster deduping
- minus 148 representative keys already present in the 155-edition classification handoff

Run this from the repo root:

```sh
cd /Users/mia/dev/personal/elements-dh/ocrflow

go run ./cmd/edition-classification \
  -keys-file ../tps_title_page_extraction_review/data/subject_classification/full_run/representative_title_page_corpus_needing_classification_keys.txt \
  -output-csv ../tps_title_page_extraction_review/data/subject_classification/full_run/full_subject_classification_representatives_preview.csv \
  -checkpoint-file ../tps_title_page_extraction_review/data/subject_classification/full_run/full_subject_classification_representatives_preview.csv.done \
  -resume=false
```

This writes results to the database because it does **not** include `-dry-run`.

It also writes a preview CSV so we can parse/check the results after the run.

## Resume After Interruption

Use this after a partial corrected run:

```sh
cd /Users/mia/dev/personal/elements-dh/ocrflow

go run ./cmd/edition-classification \
  -keys-file ../tps_title_page_extraction_review/data/subject_classification/full_run/representative_title_page_corpus_needing_classification_keys.txt \
  -output-csv ../tps_title_page_extraction_review/data/subject_classification/full_run/full_subject_classification_representatives_preview.csv \
  -checkpoint-file ../tps_title_page_extraction_review/data/subject_classification/full_run/full_subject_classification_representatives_preview.csv.done
```

## Optional Defensive Variant

If you want the app to also skip keys that already have this exact revision in the database, add:

```sh
  -skip-existing-revision
```

The prepared 695-key list already excludes the known 155-edition handoff overlap, so this should not usually be necessary.

## Input Files

- `representative_title_page_corpus_keys.txt`: 843 representative keys after reprint deduping.
- `representative_title_page_corpus_needing_classification_keys.txt`: 695 representative keys not already classified in the old 155-edition handoff. This is the one to run.
- `representative_title_page_corpus_already_classified_155_overlap_keys.txt`: 148 representative keys already covered by the old classification.
- `title_page_to_classification_representative_map.csv`: maps each of the 909 title-page keys to the representative key whose classification should be used.

## Revision

These commands use the default feature `m_classifier` and the default latest revision in `ocrflow/cmd/edition-classification`:

`6a0d47e3-f472-4b63-a6f5-67c693a0adf9`
