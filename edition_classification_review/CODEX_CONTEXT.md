# Codex Context: Edition Classification

Use this file to resume analysis after the V6 run.

## Source Data

Current source CSV:

`edition_classification_review/features-comparison-May-23 - editions_results.csv`

Columns:

`Page/Key, language, year, city, Classification, Orig_Value, llm_Value_1, llm_Value_3, llm_Value_4`

The prompt mapping from Mia:

- `llm_Value_1`: `main`
- `llm_Value_3`: `7e62f6a`
- `llm_Value_4`: `f932d03`

## Prompt Provenance

The SQL classifier prompt did not change across `main`, `7e62f6a`, and `f932d03`.

It lives in:

`ocrflow/internal/migrations/ocrflow/1774300003_edition_subject_categories_feature.sql`

The output format is one list feature containing values like:

`Category Name::primary`

Allowed values:

`primary`, `secondary`, `unrelated`, `unknown`

The wrapper did change:

- `main` / `llm_Value_1`: edition metadata includes original title, imprint, colophon, and frontispiece.
- `7e62f6a` / `llm_Value_3`: mostly same metadata, with title/imprint/colophon/frontispiece on following lines and newer parsing.
- `f932d03` / `llm_Value_4`: uses English title when available, drops imprint/colophon/frontispiece, says language is "originally in ...", and keeps hallucination checking off for edition prompts.

## V6 Code Changes

Wrapper change:

`ocrflow/internal/service/feature_exec_edition.go`

When the active edition feature is `m_classifier`, the wrapper now says this is a classification task and tells the model to return only the requested category/classification strings. This avoids the old extraction wording:

`Each field should contain the exact value(s) from the input metadata`

That old line conflicts with labels such as `primary` and `secondary`.

V6 SQL prompt:

`ocrflow/internal/migrations/ocrflow/1774300008_edition_classification_v6.sql`

Main V6 prompt changes:

- First decide related vs unrelated, then primary vs secondary.
- Require explicit evidence in title, title-page text, additional content, book/section notes, or notes.
- Do not infer subject relevance from editor, publisher, city, language, date, format, or reputation.
- Do not classify a category as related only because this is an edition of Euclid or because Euclid book numbers are listed.
- Tighten noisy categories, especially Theoretical Mathematics, Practical Geometry, Instrument Use, Instrument Construction, Construction, Architecture, Cosmography, and Trigonometry.
- Incorporate Mia's first 34 reviewed rows:
  - Construction can include military/architecture-adjacent material when it is practical construction, but not if it is only artistic/theoretical/general.
  - Instrument Construction requires actual making/design/fabrication/calibration, not just an instrument being named or depicted.
  - Instrument Use requires real instrument-use instructions; ordinary compass-and-ruler geometry and imaginary/speculative devices do not count.
  - Theoretical Mathematics has a high threshold; basic school mathematics, practical mathematics, mixed arts, or a generic Euclid base text should not automatically count.
  - Commercial Mathematics requires trade/accounting/merchant evidence.
  - Mechanics requires mechanics content, not just practical mathematics.

## Current Baseline KPIs

From `analysis_report.md`:

| Prompt | Exact | Covered exact | Unknown | Related precision | Related recall | False related |
| --- | --- | --- | --- | --- | --- | --- |
| v1 | 76.7% | 86.4% | 11.2% | 41.7% | 88.3% | 295 |
| v3 | 75.2% | 85.0% | 11.4% | 39.2% | 89.1% | 330 |
| v4 | 77.7% | 86.0% | 9.6% | 40.2% | 89.1% | 317 |

Decision buckets:

- `stable_unrelated`: 2,142
- `majority_llm_related_manual_unrelated`: 286
- `keep_manual_positive`: 223
- `single_llm_related_manual_unrelated`: 173
- `blank_with_llm_unrelated_consensus`: 162
- `blank_needs_evidence`: 55
- `blank_with_llm_related_consensus`: 36
- `review_manual_positive`: 16
- `blank_no_majority`: 7

## How To Analyze V6 Later

Add the V6 column to the source comparison CSV or make a new CSV with the same row keys and a `llm_Value_6` column.

Then update `LLM_COLS` in `analyze_review.py` to include `llm_Value_6` and rerun:

`python3 edition_classification_review/analyze_review.py`

Compare V6 on:

- related precision, especially in Theoretical Mathematics and Practical Geometry;
- false-related count;
- unknown rate;
- recall on manual positives;
- how many `majority_llm_related_manual_unrelated` rows remain.
