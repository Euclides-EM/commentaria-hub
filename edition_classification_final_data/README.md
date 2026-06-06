# Edition Classification Final Data

This folder contains the production handoff artifacts for edition subject classification.

## Layout

- `final_hybrid_classifications.csv`: final selected classification for the 155-edition comparison batch.
- `inputs/`: key lists and reprint maps for the printed-edition expansion.
- `runs/edition_v9/`: raw V9 offline output, checkpoint, and run log for 692 printed non-reprint editions.
- `analysis/edition_v9/`: concise review artifacts for the 692-edition V9 run.

`final_hybrid_classifications.csv` includes:

- `edition_id`, `language`, `year`, `city`, `category`;
- `final_value`: selected classification value;
- `final_source`: `human_review`, `llm_Value_6`, `llm_Value_7`, or `llm_Value_8`;
- `final_reason`: why that source was chosen;
- `concerns`: audit flags such as prompt disagreement, human-reviewed, unknown, or review-sensitive category;
- `review_file`, `review_notes`, `review_error_type`: retained notes for cells Mia reviewed;
- `Orig_Value`, `llm_Value_6`, `llm_Value_7`, `llm_Value_8`: compact provenance columns.

## DB Migrations

The DB installers are outside this folder:

- `ocrflow/internal/migrations/ocrflow/1774300010_edition_classification_final_hybrid_opt.sql`
  - final reviewed/hybrid values for the 155-edition comparison batch;
  - source revision: `final_hybrid_v1_2026_06_03`.
- `ocrflow/internal/migrations/ocrflow/1774300012_edition_classification_printed_v9_opt.sql`
  - latest V9 values for 692 printed non-reprint editions;
  - includes the accepted EIP Theoretical override rule and the short interactive review decisions from 2026-06-06;
  - source revision: `printed_non_reprints_v9_reviewed_2026_06_06`.

To apply both optional classification migrations via the app, include both filenames in `OPT_MIGRATIONS`.

## V9 Review Artifacts

The V9 analysis folder intentionally keeps only the files needed for review or reproduction:

- `analysis/edition_v9/README.md`: concise summary and counts.
- `analysis/edition_v9/category_summary.csv`: category-level counts.
- `analysis/edition_v9/wide_classifications.csv`: one row per edition with 20 category columns.
- `analysis/edition_v9/priority_review_queue.csv`: compact worksheet for unknowns and non-EIP Theoretical suspects.
- `analysis/edition_v9/accepted_eip_theoretical_overrides.csv`: accepted EIP rule overrides.

The V9 SQL migration installs only the 692 non-reprint editions. `inputs/printed_reprint_classification_map.csv` contains 69 reprints that can be copy-forwarded from canonical editions later; they are not materialized by the current migration.

## Future Runs

For new editions, use the V9 prompt revision:

`6a0d47e3-f472-4b63-a6f5-67c693a0adf9`

Use the offline runner for previews/new batches. It reads edition metadata CSVs directly and does not initialize the app, run migrations, touch SQLite, or write DB results:

```sh
cd /Users/mia/dev/personal/elements-dh/ocrflow

go run ./cmd/edition-classification-offline \
  -keys-file ../edition_classification_final_data/inputs/printed_missing_non_reprints_v9_keys.txt \
  -output-csv ../edition_classification_final_data/runs/edition_v9/printed_missing_non_reprints_v9_preview.csv
```

If interrupted, rerun the same command. Resume is on by default and uses `<output-csv>.done`.
