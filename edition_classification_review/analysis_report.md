# Edition Classification Analysis Report

This uses `edition_classification_review/features-comparison-May-23 - editions_results_with_v6.csv`. `Orig_Value` is treated as an imperfect reference, not ground truth.

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
| v6 | 78.2% | 89.1% | 12.3% | 47.9% | 87.0% | 226 |

Interpretation: exact accuracy is limited by noisy manual labels and the dominance of `unrelated`. Related precision, recall, false-related count, and unknown rate are the more useful prompt-comparison signals.

## Current DB Run Delta vs V4

| Metric | v4 | v6 | Delta |
| --- | --- | --- | --- |
| Exact | 77.7% | 78.2% | +0.5 pp |
| Covered exact | 86.0% | 89.1% | +3.1 pp |
| Unknown | 9.6% | 12.3% | +2.6 pp |
| Related precision | 40.2% | 47.9% | +7.7 pp |
| Related recall | 89.1% | 87.0% | -2.1 pp |
| False related | 317 | 226 | -91 |

## Adjudication Buckets

| Bucket | Rows |
| --- | --- |
| stable_unrelated | 2129 |
| majority_llm_related_manual_unrelated | 225 |
| keep_manual_positive | 223 |
| single_llm_related_manual_unrelated | 160 |
| blank_with_llm_unrelated_consensus | 157 |
| split_llm_related_manual_unrelated | 87 |
| blank_needs_evidence | 58 |
| blank_with_llm_related_consensus | 25 |
| blank_no_majority | 20 |
| review_manual_positive | 16 |

## What To Review Now

`v6_diagnostic_review.csv` contains 159 rows. Fill `your_final_value`, `your_error_type`, and `your_notes_for_v6`.

Suggested `your_final_value`: `primary`, `secondary`, `unrelated`, `unknown`, or `unsure`.

Suggested `your_error_type`: `manual_missed_related`, `llm_overclassified`, `primary_secondary_wrong`, `needs_more_metadata`, `definition_unclear`, or `other`.

## Current DB Run Provenance

- `llm_Value_6` was merged from the current `ocrflow/store/ocrflow.db` values for `scope='editions'` and `feature_id='m_classifier'`.
- The active DB result metadata reports `source_revision='f96afd86-79f0-4736-a91a-d58e37b6db65'`, which is the stored `v1` revision.
- The intended V6 prompt revision in migrations is `6f4aafde-a8d9-4a50-8cae-7947b470c6f6`; the DB values are therefore a complete current DB run comparison, but they are not provenance-confirmed as V6.
