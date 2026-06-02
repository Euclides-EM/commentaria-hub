# Codex Context For V6 Results

## Branch And Version Mapping

- V1/current-main comparison column: `Value_llm`, commit `7583b49`.
- V4 comparison column was intentionally dropped from decision-making because it behaved similarly to V5.
- V5/current-PR comparison column: `Value_llm5`, branch `origin/liri/features-v2`, commit `f932d03`.
- Current prep branch: `codex/v6-prep`.

## V6 Code Changes Made

- `ocrflow/internal/service/feature_exec_dataset.go`
  - Fixed prompt text scope labels:
    - imprint features now receive `the imprint section of a title page`;
    - non-imprint features now receive `a title page excluding the imprint section`.
  - Replaced the global "minimal text span" rule with feature-definition-driven span guidance.
  - Added explicit instructions not to over-trim titles, names, book counts, language references, or descriptors that are part of the requested feature.
  - Kept V5-style source fidelity, JSON-only output, and hallucination checking.

- `ocrflow/internal/migrations/ocrflow/1774300007_tps_v6_feature_revisions.sql`
  - Adds V6 feature revisions for the features where the diagnostic review gave useful signal.
  - The V6 revisions encode feature-specific span behavior from `v6_feature_rules.csv`.

## Review Signal

The user filled 52 of 90 diagnostic rows. High-level signal:

- `orig` preferred: 27 rows.
- working/LLM candidate preferred: 19 rows.
- `v1` preferred: 1 row.
- `v5` preferred: 3 rows.
- custom target in `preferred_source`: 2 rows.

Interpretation: V6 should be feature-specific. Do not make it universally more minimal or universally fuller.

Additional user clarifications before V6:

- `Base Content` and `Elements Designation` should include Euclid and book counts/ranges when they are part of the title/designation.
- `Verbs` should split distinct verbs/verbal expressions into separate list values.
- `Base Content Description` vs `Enriched With` distinction is acceptable as documented.
- `Adapter Attribution` should include initials only when they are name initials. Professional honorifics/titles such as `P.` for Professor should move to `Adapter Description`. In `Other Educational Authorities`, honorifics can stay because there is no separate field to store them.
- `Place in Imprint` should preserve internal punctuation such as commas in `Paris, France`, but remove terminal punctuation such as a final period.

## Files To Use After V6 Runs

- `v6_diagnostic_review.csv`
  - Compare V6 output on these rows first.
  - Treat `preferred_source`, `error_type`, and `human_final_value` as the user's signal even if formatting is informal.

- `v6_feature_rules.csv`
  - Use to interpret whether V6 followed the intended feature-level span policy.

- `working_final_dataset_orig_rows.csv`
  - Use the `decision_tier` column as KPI buckets:
    - `A_confirmed`
    - `B_keep_manual_no_llm`
    - `C_llm_agreement_overrides_manual`
    - `D_prompt_policy_choice`
    - `E_single_llm_over_manual`

## Suggested V6 Evaluation

1. Join V6 results to `v6_diagnostic_review.csv` by `Page/Key` and `Feature Name`.
2. Count how often V6 matches the user's preferred target:
   - if `human_final_value` is filled, use it;
   - else if `preferred_source` is `orig`, use `Value_orig`;
   - else if `preferred_source` is `v1`, use `Value_llm`;
   - else if `preferred_source` is `v5`, use `Value_llm5`;
   - else if `preferred_source` is `working`, use `working_final_value`;
   - else if `preferred_source` looks like a custom value, use that text.
3. Separately check whether V6 improves the risky buckets in `working_final_dataset_orig_rows.csv`.
4. Only consider V7 if V6 clearly fixes one feature family but breaks another.
