# Codex Context: Edition Classification V7

Use this file to continue the edition subject-classification work in future Codex sessions.

## Version Map

- Source comparison file:
  `edition_classification_review/features-comparison-May-23 - editions_results_with_v6.csv`
- Original comparison file:
  `edition_classification_review/features-comparison-May-23 - editions_results.csv`
- Prompt columns:
  - `llm_Value_1`: `main`
  - `llm_Value_3`: `7e62f6a`
  - `llm_Value_4`: `f932d03`
  - `llm_Value_6`: latest/current behavioral run
- V6 prompt revision:
  `6f4aafde-a8d9-4a50-8cae-7947b470c6f6`
- V7 prompt revision:
  `7a3f3e5a-8f8a-4c47-b1e7-56c31c9ab7d0`
- V7 migration:
  `ocrflow/internal/migrations/ocrflow/1774300010_edition_classification_v7.sql`

## Provenance Note

The latest `llm_Value_6` run was executed on another machine and merged into the CSV. Treat the CSV as the behavioral source of truth. Local DB provenance may not match the intended revision id.

For V7, make provenance clean: generated results should point to revision `7a3f3e5a-8f8a-4c47-b1e7-56c31c9ab7d0`.

## V6 Baseline

From `analysis_report.md`, latest run vs V4:

| Metric | V4 | V6/latest | Delta |
| --- | ---: | ---: | ---: |
| Exact | 77.7% | 78.2% | +0.5 pp |
| Covered exact | 86.0% | 89.1% | +3.1 pp |
| Unknown | 9.6% | 12.3% | +2.6 pp |
| Related precision | 40.2% | 47.9% | +7.7 pp |
| Related recall | 89.1% | 87.0% | -2.1 pp |
| False related | 317 | 226 | -91 |

V6 direction is good: higher precision and fewer false-related labels. Do not revert to broad, eager related classification.

## Human Review Signal

Mia reviewed 42 rows in `v6_diagnostic_review.csv`.

- `Orig_Value` matched 12/42 reviewed values.
- `llm_Value_6` matched 29/42 reviewed values.
- Manual `unrelated` was often too conservative: 28 reviewed manual-unrelated rows were marked `primary` or `secondary`.

Good reviewed categories:

- `Architecture`: 4/4
- `Construction`: 3/3
- `Instrument Use`: 11/12
- `Practical Geometry`: 3/3

Needs focused tightening:

- `Trigonometry`
- `Instrument Construction`
- `Commercial Mathematics`
- `Theoretical Mathematics`
- selected `Cartography` false positives

Read `v6_human_review_summary.md` for the detailed category table and Mia's notes.

## V7 Prompt Policy

V7 keeps V6 intact except for focused category wording:

- `Trigonometry`: require explicit trigonometric methods/tables/vocabulary. Sundials, astronomy, triangles, proportional geometry, or practical geometry alone are not enough.
- `Instrument Construction`: require making/designing/fabricating/calibrating real instruments. Instrument use, theory, depiction, naming, or speculation alone is unrelated.
- `Commercial Mathematics`: require commercial/accounting/trade/merchant/money/interest/exchange/commercial-measures evidence.
- `Theoretical Mathematics`: require a clear theoretical/speculative mathematical aim beyond practical, mixed-math, school, instrument, construction, military, or applied Euclidean material.
- `Cartography`: require maps/charts/mapmaking/chartmaking or map/chart content. Geography, place names, routes, diagrams, and surveying alone are not enough.

## Analyzer Behavior

`analyze_review.py` automatically picks the newest available input:

1. `features-comparison-May-23 - editions_results_with_v7.csv`
2. `features-comparison-May-23 - editions_results_with_v6.csv`
3. `features-comparison-May-23 - editions_results.csv`

It includes `llm_Value_7` when present. When run on V6 data, it writes `v6_diagnostic_review_regenerated.csv` so it does not overwrite the reviewed `v6_diagnostic_review.csv`.

## Suggested V7 Evaluation

1. Run V7 on the same edition batch.
2. Merge/export results into `features-comparison-May-23 - editions_results_with_v7.csv` with column `llm_Value_7`.
3. Run `python3 edition_classification_review/analyze_review.py`.
4. Compare V7 to V6 on:
   - related precision;
   - related recall;
   - false-related count;
   - unknown rate;
   - reviewed-row agreement in the five focus categories.
5. Review only new V7 disagreement rows in the focus categories.
