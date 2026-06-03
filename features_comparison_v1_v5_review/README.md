# TPS Feature Extraction V7 Handoff

This folder is the clean handoff packet for the TPS V7 feature-extraction run.

## Current Status

- V6 is useful as a diagnostic run, but not ready as the default.
- V6 matched 19 of 52 human-targeted diagnostic rows and had 1 missing targeted value.
- V7 is prepared as a targeted prompt-boundary revision in:
  `ocrflow/internal/migrations/ocrflow/1774300009_tps_v7_targeted_feature_revisions.sql`
- Do not rerun or revise V6 unless explicitly requested. The next run should use V7.

## Files

- `CODEX_CONTEXT.md`: start here in future Codex sessions.
- `v6_evaluation_report.md`: V6 baseline and failure clusters.
- `v6_diagnostic_review.csv`: original diagnostic review rows and human feedback.
- `working_final_dataset_orig_rows.csv`: KPI baseline rows with decision tiers.
- `v7_policy_review.csv`: targeted pre-V7 review sheet, including the latest reviewed `human_final_value` corrections.

## V7 Review Columns

In `v7_policy_review.csv`, treat either of these as reviewed target values:

- `human_final_value`
- `reviewer_v7_value`

If both are present, prefer `reviewer_v7_value`. Use `EMPTY` when the feature should return no value.

## Important V7 Rules

- `Base Content`: include book counts/ranges when they are part of the core title, including `TREDECIM` in `EVCLIDIS ELEMENTORVM GEOMETRICORVM LIBROS TREDECIM`; stop before separately bound works.
- `Base Content Description` vs `Enriched With`: follow the title page framing. Added/supplementary material goes to `Enriched With`; material framed as inherent core content can stay in `Base Content Description`.
- `Adapter Attribution`: keep family/name-origin geography, such as `de Mans`; do not keep role, institutional, office, or residence geography.
- `Adapter Description`: professional descriptors always belong here, as do institutional/role/geographic settings.
- `Place in Imprint`: V7 makes `location_in_imprint` list-valued and should prefer full publication address/place phrases.
- `Verbs`: include action participles when they function as edition or adapter action claims.

## Next Codex Task

1. Apply/run V7 revisions.
2. Join V7 results to `v6_diagnostic_review.csv`.
3. Use reviewed targets from `v7_policy_review.csv` where available.
4. Compare V7 against `v6_evaluation_report.md`, especially the 19/52 diagnostic baseline.
5. Check KPI movement by decision tier using `working_final_dataset_orig_rows.csv`.
