# TPS Title-Page Extraction Review

This folder preserves the human/Codex context for the V8 TPS title-page extraction work. It is no longer a run workspace; transient logs, checkpoints, binaries, and intermediate preview outputs were removed.

## Final Artifacts

- `v8_results_preview_reviewed.csv`: final reviewed V8 output used to generate the migration.
- `v8_analysis_report.md`: final V8 coverage/quality summary.
- `v8_feature_summary.csv`: final feature-level coverage summary.
- `v8_policy_comparison.csv`: targeted policy comparison against reviewed examples.
- `v8_policy_decisions.md`: extraction policy decisions made during interactive review.
- `v8_interactive_review_decisions.md`: final manual review notes, including the two dropped Enriched With rows.
- `v8_suspicious_rows.csv`: the rows reviewed and intentionally removed from the final CSV.

## Historical Context Kept

- `v6_evaluation_report.md` and `v6_diagnostic_review.csv`: baseline failure analysis.
- `v7_analysis_report.md`, `v7_feature_summary.csv`, `v7_policy_comparison.csv`, and `v7_policy_review.csv`: V7 comparison and policy-review context.
- `working_final_dataset_orig_rows.csv`: broader source review dataset.
- `CODEX_CONTEXT.md`: compact handoff notes for future work.

## Production Shape

The old production/paper snapshot remains in:

- `/Users/mia/dev/personal/elements-dh/ocrflow/internal/migrations/ocrflow/1774207517_tps_feature_results_opt.sql`

That migration still contains the original `ann_1` title-page classification definitions and results used for the paper.

The final reviewed V8 work is collapsed into:

- `/Users/mia/dev/personal/elements-dh/ocrflow/internal/migrations/ocrflow/1774300009_tps_title_page_latest_revisions.sql`
- `/Users/mia/dev/personal/elements-dh/ocrflow/internal/migrations/ocrflow/1774300011_tps_title_page_v8_reviewed_results.sql`

Intermediate V6/V7/V8 prompt migrations were removed because this branch is local/recreatable and the final state is clearer as a latest revision plus reviewed results.

## Offline Runner

Use only the DB-free runner for future long TPS extraction runs. It reads metadata CSVs directly and does not initialize the app, run migrations, warm caches, touch SQLite, or store DB results.

```sh
cd /Users/mia/dev/personal/elements-dh/ocrflow

go run ./cmd/title-page-extraction-offline \
  -output-csv ../tps_title_page_extraction_review/v8_offline_results_preview.csv
```

Useful options:

- `-keys Amsterdam_1616,Amsterdam_1626` to run selected keys.
- `-keys-file path/to/keys.csv` to run keys from a CSV or one-key-per-line file.
- `-features base_content,enriched_with` to run selected features.
- `-checkpoint-file path/to/file.done` to choose a checkpoint file.
- `-resume=false` to restart an output file from scratch.

If interrupted, rerun the same command. Completed keys are skipped from the checkpoint file.
