# V6 Human Review Summary

This summarizes the filled rows in `v6_diagnostic_review.csv` after Mia's latest review pass.

## Review Coverage

- Reviewed diagnostic rows: 42/159
- Human final values:
  - `primary`: 14
  - `secondary`: 14
  - `unrelated`: 14
- Reviewed statuses:
  - `majority_llm_related_manual_unrelated`: 34
  - `split_llm_related_manual_unrelated`: 6
  - `review_manual_positive`: 2

## Agreement With Reviewed Values

| Column | Matches | Rate |
| --- | ---: | ---: |
| `Orig_Value` | 12/42 | 28.6% |
| `llm_Value_1` | 29/42 | 69.0% |
| `llm_Value_3` | 29/42 | 69.0% |
| `llm_Value_4` | 29/42 | 69.0% |
| `llm_Value_6` | 29/42 | 69.0% |
| `suggested_final_value` | 29/42 | 69.0% |

Interpretation: the old manual labels are very noisy in these reviewed conflict rows. `llm_Value_6` is often right, but the remaining misses are concentrated enough to justify a focused next prompt pass rather than broad rewrites.

## Category Signal

| Classification | V6 matches | Reviewed targets | Notes |
| --- | ---: | ---: | --- |
| Architecture | 4/4 | all `secondary` | Looks good in reviewed rows. |
| Construction | 3/3 | 2 `primary`, 1 `secondary` | Looks good; some primary rows may be theoretically framed but still construction-relevant. |
| Instrument Use | 11/12 | 6 `primary`, 4 `secondary`, 2 `unrelated` | Mostly good. Keep suppressing ordinary ruler/compass geometry. |
| Practical Geometry | 3/3 | 2 `primary`, 1 `secondary` | Looks good in reviewed rows. |
| Instrument Construction | 4/6 | 2 `primary`, 1 `secondary`, 3 `unrelated` | Needs stricter distinction from instrument use. |
| Cartography | 2/4 | 1 `primary`, 1 `secondary`, 2 `unrelated` | Mixed; review evidence patterns before prompt change. |
| Commercial Mathematics | 0/2 | both `unrelated` | V6 still overclassifies weak commercial signals. |
| Theoretical Mathematics | 0/2 | 1 `secondary`, 1 `unrelated` | Too eager for practical Euclidean/mixed-math works. |
| Trigonometry | 2/6 | 1 `primary`, 1 `secondary`, 4 `unrelated` | Needs more conservative evidence threshold; sundials/triangles alone are not enough. |

## Policy Notes From Mia

- Instrument Construction: do not infer construction from a book that discusses how to use instruments. Making/designing/fabricating/calibrating must be explicit.
- Instrument Use: ordinary ruler-and-compass geometry is not instrument use.
- Theoretical Mathematics: practical or mixed-math Euclidean works should not be classified as theoretical merely because they use Euclid; theoretical/speculative aim needs to be clear.
- Trigonometry: be conservative and avoid anachronism. Sundials or borderline triangle material do not imply trigonometry unless trigonometric method/table vocabulary is explicit.
- Commercial Mathematics: weak applied-math or practical-use evidence is not enough; require actual commercial, accounting, trade, merchant, interest, exchange, or measures evidence.

## Recommendation

Do one focused V7 classification prompt pass only if needed. Do not revert the V6 direction: it improved precision and reduced false-related classifications overall. The next prompt should target:

1. `Trigonometry`
2. `Instrument Construction`
3. `Commercial Mathematics`
4. `Theoretical Mathematics`
5. selected `Cartography` false positives

Before making another prompt revision, review a few more rows from those same categories if time allows.
