# Edition Classification V6 Prep

This folder is for the edition subject-classification review only. It uses:

- `features-comparison-May-23 - editions_results.csv`: source comparison file.
- `analysis_report.md`: human-readable KPI/report summary.
- `v6_diagnostic_review.csv`: small review queue for Mia.
- `full_row_suggestions.csv`: row-level suggestions for all 3,100 rows.
- `kpi_summary.csv`: prompt KPIs overall, by language, and by classification.
- `analyze_review.py`: reproducible generator for the CSV/report files.
- `CODEX_CONTEXT.md`: prompt provenance and V6 run context.

## What To Review

Open `v6_diagnostic_review.csv`. You do not need to review all 3,100 rows.

Fill only these columns:

- `your_final_value`: use `primary`, `secondary`, `unrelated`, `unknown`, or `unsure`.
- `your_error_type`: use `manual_missed_related`, `llm_overclassified`, `primary_secondary_wrong`, `needs_more_metadata`, `definition_unclear`, or `other`.
- `your_notes_for_v6`: short note only when it helps improve the prompt, for example "Euclid alone should not count as theoretical math" or "instrument is named but no use/construction instructions".

Highest-value rows are at the top. The key question is whether LLM-majority related rows are real manual misses or prompt over-classification.

As of the first review pass, 34 rows were filled. That is enough to move to V6. Continue only if you want more examples for a specific category; the main prompt changes are already clear.

## KPIs To Watch

For the old runs, exact accuracy is not enough because most rows are `unrelated` and manual labels are imperfect.

The main KPIs are:

- Related precision: among LLM `primary`/`secondary`, how many are manually related.
- Related recall: among manual related rows, how many the LLM catches.
- False-related count: likely over-classification plus manual misses.
- Review burden: number of manual-unrelated rows where 2-3 LLMs say related.
- Unknown rate: how often the prompt abstains.

## V6 Plan

V6 changes two things:

- Go wrapper: subject classification now gets a classification-specific wrapper instead of the old extraction wrapper.
- SQL prompt: stricter evidence rules and clearer disambiguation for noisy categories.
- Mia review notes folded into V6: high threshold for Theoretical Mathematics; compass/ruler geometry is not Instrument Use; imaginary/speculative devices are not Instrument Use; Instrument Construction needs actual fabrication/design instructions; Commercial Mathematics needs actual trade/accounting/merchant evidence; Mechanics needs mechanics, not just practical math.

Run V6 on the same edition batch, then compare V6 against `llm_Value_1`, `llm_Value_3`, and `llm_Value_4` using this folder as the baseline.
