# TPS V6 Prep Review

This folder keeps the compact context needed for the V6 TPS feature-extraction run.

## What We Learned

- V6 should not use a global "minimal span" rule.
- Some features need fuller spans: `Base Content`, `Elements Designation`, `Adapter Attribution`, `Verbs`, `Enriched With`.
- Some features benefit from cleaned spans: `Euclid References`, `Date in Imprint`, some `Institutions`, some `Adapter Description`, and some imprint fields.
- V5 had a likely wrapper bug: imprint and non-imprint prompt descriptions were reversed. The V6 code patch fixes that.
- Adapter names should include initials when they are name initials, but not professional honorifics. A title-like `P.` for Professor belongs in adapter description; honorifics in `Other Educational Authorities` can stay because there is no separate field for them.
- `Place in Imprint` should preserve internal punctuation such as commas in a place phrase, but drop terminal punctuation such as a final dot.
- `Verbs` should split distinct verbs/verbal expressions into separate list values.

## Files

- `v6_diagnostic_review.csv`: the filled diagnostic review used to infer V6 behavior.
- `v6_feedback_summary.md`: concise interpretation of the filled review.
- `v6_feature_rules.csv`: feature-level span/prompt rules for V6.
- `working_final_dataset_orig_rows.csv`: working final dataset for rows that already had `Value_orig`, with decision tiers used as KPI buckets.
- `CODEX_CONTEXT.md`: implementation notes for the next Codex pass, especially when comparing V6 results.

## Next Step

Run V6 using the patched code and the V6 feature revisions. After results are available, compare V6 against:

1. the 90 rows in `v6_diagnostic_review.csv`;
2. the decision tiers in `working_final_dataset_orig_rows.csv`;
3. the feature-level expectations in `v6_feature_rules.csv`.
