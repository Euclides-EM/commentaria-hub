# Edition Classification Final Data

This folder intentionally contains only the production handoff artifact for the edition subject-classification review.

## File

- `final_hybrid_classifications.csv`: final selected classification for 155 editions * 20 subject categories.

The CSV includes:

- `final_value`: selected classification value;
- `final_source`: `human_review`, `llm_Value_6`, `llm_Value_7`, or `llm_Value_8`;
- `final_reason`: why that source was chosen;
- `concerns`: audit flags such as prompt disagreement, human-reviewed, unknown, or review-sensitive category;
- `review_file`, `review_notes`, `review_error_type`: retained notes for cells Mia reviewed;
- `Orig_Value`, `llm_Value_6`, `llm_Value_7`, `llm_Value_8`: compact provenance columns.

## DB Migration

The DB installer is outside this folder:

`ocrflow/internal/migrations/ocrflow/1774300010_edition_classification_final_hybrid_opt.sql`

That optional migration preserves the V6/V7/V8/V9 prompt revisions and writes the final hybrid values with source revision:

`final_hybrid_v1_2026_06_03`

## New Editions

For new editions, use the V9 prompt revision:

`6a0d47e3-f472-4b63-a6f5-67c693a0adf9`

The runner defaults to V9:

```sh
cd ocrflow
go run ./cmd/edition-classification \
  -keys-file path/to/new_keys.txt \
  -output-csv path/to/new_results_preview.csv
```

Use `-dry-run` for previews. Remove `-dry-run` when writing results to the DB.
