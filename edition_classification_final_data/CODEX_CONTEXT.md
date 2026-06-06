# Codex Context: Edition Classification Final Data

This folder is the organized handoff area for edition subject classification.

## Root Files

- `final_hybrid_classifications.csv`: final reviewed/hybrid classifications for the 155-edition comparison batch.
- `README.md`: human-facing instructions.
- `CODEX_CONTEXT.md`: this future-session context.

## Subfolders

- `inputs/`: canonical inputs for the printed-edition expansion.
  - `printed_missing_non_reprints_v9_keys.txt`: 692 printed, missing-classification, non-reprint edition keys.
  - `printed_reprint_classification_map.csv`: 69 printed reprints omitted from the LLM run and their source editions. These are not materialized by the current SQL migration.
- `runs/edition_v9/`: raw offline V9 execution artifacts for those 692 editions.
  - `printed_missing_non_reprints_v9_preview.csv`
  - `printed_missing_non_reprints_v9_preview.csv.done`
  - `printed_missing_non_reprints_v9_run.log`
- `analysis/edition_v9/`: concise review artifacts for the V9 run.
  - `README.md`: summary, counts, and review guidance.
  - `category_summary.csv`: category-level counts.
  - `wide_classifications.csv`: one row per edition with 20 category columns.
  - `priority_review_queue.csv`: compact worksheet for unknowns and non-EIP Theoretical suspects. The first two tiny interactive review batches from 2026-06-06 are recorded in `review_decision` / `corrected_value`.
  - `accepted_eip_theoretical_overrides.csv`: Mia's accepted rule override for the 72 EIP false negatives.

## Final CSV

`final_hybrid_classifications.csv` has 3,100 rows: 155 editions * 20 subject categories.

It preserves compact provenance:

- `edition_id`, `language`, `year`, `city`, and `category`;
- final selected value;
- chosen source;
- reason/concern flags;
- review notes where Mia manually filled a value;
- original, V6, V7, and V8 values.

## Migrations

Use these optional migrations:

- `ocrflow/internal/migrations/ocrflow/1774300010_edition_classification_final_hybrid_opt.sql`
  - installs `m_classifier`, V6/V7/V8/V9 prompt revisions, and final hybrid result rows for the 155 existing editions;
  - source revision: `final_hybrid_v1_2026_06_03`.
- `ocrflow/internal/migrations/ocrflow/1774300012_edition_classification_printed_v9_opt.sql`
  - installs latest V9 result rows for 692 printed non-reprint editions;
  - includes accepted EIP Theoretical overrides and the short 2026-06-06 interactive review decisions;
  - source revision: `printed_non_reprints_v9_reviewed_2026_06_06`.

Both files are optional; include their filenames in `OPT_MIGRATIONS` when applying through the app.

## Future/New Editions

For new editions, run the latest prepared single prompt, V9:

`6a0d47e3-f472-4b63-a6f5-67c693a0adf9`

Use the offline runner for previews/new batches. It reads edition metadata CSVs directly and does not initialize the app, run migrations, touch SQLite, or write DB results:

```sh
cd /Users/mia/dev/personal/elements-dh/ocrflow

go run ./cmd/edition-classification-offline \
  -keys-file ../edition_classification_final_data/inputs/printed_missing_non_reprints_v9_keys.txt \
  -output-csv ../edition_classification_final_data/runs/edition_v9/printed_missing_non_reprints_v9_preview.csv
```

Resume is on by default and uses `<output-csv>.done`. Reviewed results are installed through optional SQL migrations, not by a DB-writing runner.

Spot-check `Commercial Mathematics`, `Construction`, `Practical Geometry`, `Instrument Use`, `Instrument Construction`, and any `unknown` value first.
