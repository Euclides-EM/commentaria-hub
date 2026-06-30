# TPS Title-Page Extraction Review

This folder is now organized as a research workspace for looking at the title-page extraction results and rebuilding the historical argument.

## Analysis Companion

The open-ended analysis workspace is:

`analysis_companion/`

Start there when working on the conference argument. It contains the high-level plan, data map, companion launch prompt, question log, casebook, and candidate-argument scratchpad.

## Start Here

Open this file first:

`data/latest_larger_corpus/printed_missing_tps_v8_preview.csv`

That is the latest larger-corpus TPS extraction output. It contains 7,169 non-empty feature rows for 650 printed editions. The original target run had 690 keys with title-page transcriptions; 650 produced at least one extracted feature row.

For the older reviewed Elements-oriented dataset, use:

`data/reviewed_elements_v8/v8_results_preview_reviewed.csv`

That file contains the manually reviewed V8 output for 217 editions.

For genre-aware analysis, first run the full subject classification prepared here:

`data/subject_classification/full_run/RUN_COMMANDS.md`

The existing subject classification is only a partial 155-edition archive, so it should not be used as the main basis for the analysis.

## Folder Map

### `data/latest_larger_corpus/`

Current raw data for the broader printed-book corpus.

- `printed_missing_tps_v8_preview.csv`: latest larger-corpus extraction results; this is the main CSV to inspect.
- `printed_missing_tps_v8_keys.txt`: 690 keys targeted by the run.
- `printed_missing_tps_without_transcription_keys.txt`: 2 target keys without title-page transcription.
- `printed_tps_target_counts.txt`: quick accounting of targets, missing values, and run status.

### `data/reviewed_elements_v8/`

Older reviewed TPS output for the Elements-oriented batch.

- `v8_results_preview_reviewed.csv`: final reviewed V8 results for 217 editions.
- `v8_feature_summary.csv`: feature coverage summary for the reviewed V8 batch.
- `v7_feature_summary.csv`: previous V7 feature coverage summary, kept for comparison.

### `data/source_review_snapshot/`

Source material from the earlier human/Codex review cycle.

- `working_final_dataset_orig_rows.csv`: broad source review table with original/manual/LLM values and decision tiers.

### `data/subject_classification/`

Subject classification workspace.

- `full_run/`: prepared input key lists and commands for rerunning the classifier over the title-page corpus.
- `partial_155_archive/`: old partial classification files covering only 155 editions. Keep for provenance; do not use as the main analysis dataset.
- `README.md`: explanation of the current classification state.

### `reports/`

Narrative reports and working interpretive notes.

- `larger_corpus_reassessment.md`: first-pass reassessment of the old presentation claims against the larger corpus.
- `v8_analysis_report.md`: final V8 coverage/quality summary for the reviewed Elements batch.
- `v7_analysis_report.md`: previous V7 analysis.
- `v6_evaluation_report.md`: older baseline evaluation.
- `v8_policy_decisions.md`: extraction policy decisions from the V8 review.
- `v8_interactive_review_decisions.md`: final interactive cleanup notes.

### `review_process/`

Policy-review and QA artifacts. These are useful when checking why an extraction rule exists, but they are not the first place to read raw data.

- `v6_diagnostic_review.csv`
- `v7_policy_comparison.csv`
- `v7_policy_review.csv`
- `v8_policy_comparison.csv`
- `v8_suspicious_rows.csv`
- `v8_targeted_review_queue.csv`

### `presentation_context/`

The two documents that frame the current interpretive problem.

- `Title Page Presentation.pdf`: older presentation based on the smaller Elements corpus.
- `USTC 2026 Abstract.pdf`: abstract for the upcoming presentation.

### `run_artifacts/`

Execution bookkeeping from the latest larger-corpus extraction.

- `printed_missing_tps_v8_preview.csv.done`: checkpoint/completion list from the run.
- `printed_missing_tps_v8_run.log`: run log with debug lines and grounding warnings.

## Working Notes

For the next phase, the goal is not to force one big conclusion immediately. We can work question by question against the raw CSV, comparing:

- the larger corpus against the older reviewed Elements batch;
- Euclid/Elements vocabulary against non-Euclidean "elements" and geometry titles;
- the full subject-classification run output once it is generated in `data/subject_classification/full_run/`;
- title-page features such as `base_content`, `elements_designation`, `enriched_with`, `bound_with_minimal`, `audience`, `institutions`, and imprint fields.

The current emerging caution is that the older "stable base designation" claim was probably too dependent on the smaller Elements-oriented corpus. The larger dataset suggests a broader ecosystem of elementary mathematical books, practical geometries, institutional textbooks, and professional manuals.
