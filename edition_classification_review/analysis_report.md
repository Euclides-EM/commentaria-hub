# Edition Classification Analysis Report

This uses only `edition_classification_review/features-comparison-May-23 - editions_results.csv`. `Orig_Value` is treated as an imperfect reference, not ground truth.

## Dataset

- Rows: 3,100
- Editions: 155
- Classifications per edition: [20]
- Manual labels: `unrelated` 2,601, `blank` 260, `primary` 221, `secondary` 18

## KPI Snapshot

| Prompt | Exact | Covered exact | Unknown | Related precision | Related recall | False related |
| --- | --- | --- | --- | --- | --- | --- |
| v1 | 76.7% | 86.4% | 11.2% | 41.7% | 88.3% | 295 |
| v3 | 75.2% | 85.0% | 11.4% | 39.2% | 89.1% | 330 |
| v4 | 77.7% | 86.0% | 9.6% | 40.2% | 89.1% | 317 |

Interpretation: all three runs have high related recall, but weak related precision against the manual labels. Because the manual labels are known to be sloppy, the most important KPI is not exact accuracy; it is how much review is needed to separate manual misses from LLM over-classification.

## Adjudication Buckets

| Bucket | Rows |
| --- | --- |
| stable_unrelated | 2142 |
| majority_llm_related_manual_unrelated | 286 |
| keep_manual_positive | 223 |
| single_llm_related_manual_unrelated | 173 |
| blank_with_llm_unrelated_consensus | 162 |
| blank_needs_evidence | 55 |
| blank_with_llm_related_consensus | 36 |
| review_manual_positive | 16 |
| blank_no_majority | 7 |

## What To Review Now

`v6_diagnostic_review.csv` contains 136 rows. Fill `your_final_value`, `your_error_type`, and `your_notes_for_v6`.

Suggested `your_final_value`: `primary`, `secondary`, `unrelated`, `unknown`, or `unsure`.

Suggested `your_error_type`: `manual_missed_related`, `llm_overclassified`, `primary_secondary_wrong`, `needs_more_metadata`, `definition_unclear`, or `other`.

## V6 Direction

- Use a classification-specific edition wrapper. The old wrapper says each field should contain exact metadata text, which conflicts with returning labels like `primary` and `secondary`.
- Keep the v4 metadata shape as the base because it had the best exact score and lowest unknown rate.
- Tighten the SQL prompt around the noisy categories: Practical Geometry vs Theoretical Mathematics, Instrument Use vs Instrument Construction, Construction vs Architecture, and Geography/Cosmography/Astronomy/Cartography.
- Require explicit evidence for `primary`/`secondary`; do not classify from a person's profession, publisher, city, or generic Euclid-only metadata alone.
- Treat `primary` vs `secondary` as the second decision. First decide whether the category is meaningfully related at all.
