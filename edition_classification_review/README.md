# Edition Classification V7 Handoff

This folder is the clean handoff packet for the edition subject-classification V7 run.

## Current Status

- Latest behavioral comparison column: `llm_Value_6`
- V6 improved precision and reduced false-related classifications, but still misses reviewed policy boundaries in a few categories.
- V7 is prepared as a focused prompt revision in:
  `ocrflow/internal/migrations/ocrflow/1774300010_edition_classification_v7.sql`
- Do not broad-rewrite the classifier. V7 should only tighten the categories identified by human review.

## Files

- `CODEX_CONTEXT.md`: start here in future Codex sessions.
- `analysis_report.md`: latest V6/current-run KPI baseline.
- `v6_human_review_summary.md`: summary of Mia's 42 reviewed diagnostic rows and policy implications.
- `v6_diagnostic_review.csv`: reviewed V6 diagnostic queue; preserve the filled `your_*` columns.
- `features-comparison-May-23 - editions_results.csv`: original comparison source.
- `features-comparison-May-23 - editions_results_with_v6.csv`: latest comparison source with `llm_Value_6`.
- `analyze_review.py`: reproducible analyzer. It will automatically use `features-comparison-May-23 - editions_results_with_v7.csv` if that file exists.

## V7 Focus

Keep V6 behavior broadly intact. V7 only tightens:

- `Trigonometry`: be conservative; sundials, astronomy, triangles, and proportional geometry are not enough without trigonometric method/table vocabulary.
- `Instrument Construction`: require actual making/designing/fabricating/calibrating; instrument use or theory alone is not construction.
- `Commercial Mathematics`: require trade/accounting/merchant/money/interest/exchange/commercial-measures evidence.
- `Theoretical Mathematics`: practical or mixed-math Euclidean works are not theoretical unless a theoretical/speculative aim is clear.
- `Cartography`: require map/chart making or map/chart content, not geography/place/route/diagram evidence alone.

## Next Steps

1. Apply/run V7 revision `7a3f3e5a-8f8a-4c47-b1e7-56c31c9ab7d0`.
2. Export/merge V7 results as `features-comparison-May-23 - editions_results_with_v7.csv` with a column named `llm_Value_7`.
3. Run `python3 edition_classification_review/analyze_review.py`.
4. Compare V7 against the V6 baseline in `analysis_report.md` and `v6_human_review_summary.md`.
5. Review only the new V7 disagreement rows in the five V7 focus categories.

## Important Warning

Do not overwrite `v6_diagnostic_review.csv`; it contains Mia's reviewed values. The analyzer writes a separate regenerated V6 diagnostic file when run without V7 input.
