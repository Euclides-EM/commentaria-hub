# V6 Feedback Summary

## Review Coverage

| Metric | Value |
| --- | --- |
| Diagnostic rows | 90 |
| Rows with feedback | 52 (57.8%) |
| Preferred `orig` | 27 (51.9%) |
| Preferred working/LLM candidate | 19 (36.5%) |
| Preferred v1 | 1 |
| Preferred v5 | 3 |
| Custom target value supplied in preferred_source | 2 |

## Main Interpretation

Your feedback says v6 should not simply become more minimal. In the reviewed rows, `orig` is preferred slightly more often than the working LLM candidate, especially for title/designation/person/content fields. But the working LLM candidate is often better for punctuation cleanup, Euclid references, dates, some institutions, and some adapter descriptions. So v6 needs feature-specific span policies, not one global extraction style.

## Decision Tier Feedback

| Tier | Reviewed | orig | working | v1 | v5 | custom |
| --- | --- | --- | --- | --- | --- | --- |
| B_keep_manual_no_llm | 5 | 4 | 0 | 0 | 1 | 0 |
| C_llm_agreement_overrides_manual | 35 | 19 | 16 | 0 | 0 | 0 |
| D_prompt_policy_choice | 10 | 4 | 2 | 1 | 2 | 1 |
| E_single_llm_over_manual | 2 | 0 | 1 | 0 | 0 | 1 |

## Feature-Level V6 Rules

| Feature | Reviewed | orig | working | v1 | v5 | custom | V6 instruction |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Adapter Attribution | 4 | 3 | 1 | 0 | 0 | 0 | Include initials, honorific initials, and surname particles when they are part of the printed name; do not include following place/role descriptors. |
| Adapter Description | 5 | 1 | 3 | 0 | 1 | 0 | Include role/affiliation descriptors adjacent to the adapter, but do not absorb the adapter's name or unrelated work title. |
| Base Content | 6 | 6 | 0 | 0 | 0 | 0 | Do not cut off book counts, Euclid references, or title qualifiers that are part of the title/designation. |
| Base Content Description | 2 | 1 | 0 | 0 | 0 | 1 | Extract description of the Elements/content itself; do not include descriptions of Euclid as a person, and do not absorb additions/enrichments. |
| Bound With | 1 | 0 | 0 | 0 | 0 | 1 | Use the custom reviewed value as an example; current candidates missed the desired boundary. |
| Date in Imprint | 1 | 0 | 1 | 0 | 0 | 0 | Keep the cleaner LLM span; remove noisy prefixes/suffixes and outer punctuation. |
| Edition Statement | 1 | 0 | 1 | 0 | 0 | 0 | Keep the cleaner LLM span; remove noisy prefixes/suffixes and outer punctuation. |
| Elements Designation | 4 | 4 | 0 | 0 | 0 | 0 | Do not cut off book counts, Euclid references, or title qualifiers that are part of the title/designation. |
| Enriched With | 2 | 2 | 0 | 0 | 0 | 0 | Include additions/enrichments as fuller phrases, not just isolated keywords. |
| Euclid References | 4 | 0 | 4 | 0 | 0 | 0 | Keep the cleaner LLM span; remove noisy prefixes/suffixes and outer punctuation. |
| Explicit Language References | 3 | 2 | 0 | 0 | 1 | 0 | Extract all explicit source/target language references when multiple languages are mentioned. |
| Institutions | 4 | 0 | 3 | 1 | 0 | 0 | Keep the cleaner LLM span; remove noisy prefixes/suffixes and outer punctuation. |
| Intended Audience | 2 | 1 | 1 | 0 | 0 | 0 | Keep the cleaner LLM span; remove noisy prefixes/suffixes and outer punctuation. |
| Other Educational Authorities | 2 | 1 | 1 | 0 | 0 | 0 | Keep the cleaner LLM span; remove noisy prefixes/suffixes and outer punctuation. |
| Patronage Dedication | 1 | 0 | 1 | 0 | 0 | 0 | Keep the cleaner LLM span; remove noisy prefixes/suffixes and outer punctuation. |
| Place in Imprint | 4 | 2 | 2 | 0 | 0 | 0 | Prefer the place/city unit from the imprint; remove only outer punctuation, but keep compound place phrases when they name the publication place. |
| Publisher in Imprint | 1 | 1 | 0 | 0 | 0 | 0 | Avoid over-trimming; include the full named/title phrase when it is part of the feature. |
| Verbs | 5 | 3 | 1 | 0 | 1 | 0 | Extract all action verbs that describe edition/adapter work; do not drop the first verb in a coordinated list, and do not include non-verbal surrounding phrases. |

## Recommended V6 Changes

- Keep v5's stricter JSON/output discipline and source-span hallucination check.

- Fix the likely imprint/non-imprint text-description reversal before rerunning v6.

- Replace the global `minimal span` bias with feature-specific span instructions from `v6_feature_rules.csv`.

- Preserve internal punctuation, initials, and line-break hyphenation when they are inside the chosen span; trim only outer punctuation/noise.

- For `Verbs`, tell the model to include every relevant action verb in coordinated lists, not only later revision/correction verbs.

- For `Base Content` and `Elements Designation`, tell the model not to drop book counts, Euclid references, or title qualifiers when they are part of the designation.

- For `Base Content Description` vs `Enriched With`, add explicit negative examples so additions do not get swallowed into content description.

## What Now

Make v6 as one focused run. Do not plan v8. If v6 improves the reviewed patterns and preserves strong imprint fields, use it as the next-batch default with feature-specific fallbacks. If v6 only fixes one family and breaks another, do one small v7 only for that family.

## Files

- `v6_feature_rules.csv`: machine-readable feature-level v6 rules.

- `v6_feedback_summary.md`: this summary.
