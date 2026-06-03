# Codex Context: TPS V7 Run

Use this file to continue the TPS feature-extraction work in future Codex sessions.

## Version Map

- V1 comparison column: `Value_llm`, commit `7583b49`.
- V4 was dropped from decision-making because it behaved similarly to V5.
- V5 comparison column: `Value_llm5`, branch `origin/liri/features-v2`, commit `f932d03`.
- V6 evaluated DB annotation: `ann_4kp7fc` (`Title Page Experiment Reviewed`).
- V7 prepared migration:
  `ocrflow/internal/migrations/ocrflow/1774300009_tps_v7_targeted_feature_revisions.sql`

## V6 Baseline

V6 is not ready as the default, but it is the baseline for V7.

- Human-targeted diagnostic rows: 52
- V6 exact matches: 19/52
- Missing targeted values: 1
- Weak clusters: `Verbs`, `Base Content Description`, `Enriched With`, adapter fields, `Elements Designation`, `Bound With`, `Publisher in Imprint`
- Non-V6 revision coverage to fix/measure in V7: `Date in Imprint`, `Edition Statement`, `Euclid Description`, privilege fields

Read `v6_evaluation_report.md` for the detailed V6 mismatch table.

## Clean Folder Files

- `README.md`: high-level handoff.
- `CODEX_CONTEXT.md`: this future-session guide.
- `v6_evaluation_report.md`: V6 diagnostic/KPI baseline.
- `v6_diagnostic_review.csv`: original diagnostic review rows.
- `working_final_dataset_orig_rows.csv`: KPI baseline rows with decision tiers.
- `v7_policy_review.csv`: targeted pre-V7 review sheet and latest human corrections.

## Target Selection For V7 Evaluation

When evaluating V7 diagnostic rows, derive the target in this order:

1. If `v7_policy_review.csv` has `reviewer_v7_value`, use it.
2. Else if `v7_policy_review.csv` has `human_final_value`, use it.
3. Else if `v6_diagnostic_review.csv` has `human_final_value`, use it.
4. Else if `preferred_source = orig`, use `Value_orig`.
5. Else if `preferred_source = v1`, use `Value_llm`.
6. Else if `preferred_source = v5`, use `Value_llm5`.
7. Else if `preferred_source = working`, use `working_final_value`.
8. Else if `preferred_source` looks like a custom text value, use that text.

Treat `EMPTY` as an explicit target for no extracted value.

## Latest Human Corrections

The user filled `human_final_value` in `v7_policy_review.csv` for these rows:

- `R0139`, Adapter Attribution, `Amsterdam_1695`:
  `CLAAS JANSZ. VOOGHT`
- `R0171`, Adapter Attribution, `Amsterdam_1700`:
  `CLAUDE FRANÇOIS MILLET DECHALLES`
- `R0202`, Adapter Attribution, `Ansbach_1610`:
  `SIMONEM MARIUM`
- `R0019`, Adapter Description, `Amsterdam_1616`:
  `der stadt Leyden Landtmeter ende Wijnroeyer`
- `R0104`, Adapter Description, `Amsterdam_1662`:
  `Professer Matheseos der Hooge Schoole tot Leyden`
- `R0868`, Adapter Description, `Frankfurt_1674`:
  `REGISCURIANI E SOCIET. JESU, Gymnasio Matheseos Professoris CURSUS MATHEMATICUS`
- `R0222`, Base Content, `Antwerp_1645`:
  `EVCLIDIS ELEMENTORVM GEOMETRICORVM LIBROS TREDECIM`
- `R0009`, Enriched With, `Alcala_1637`:
  `comentado`

Important correction: `R0222` must include `TREDECIM`. The earlier shorter value without `TREDECIM` was a mistake.

## V7 Prompt Policy

- `Base Content`: include Euclid references, title qualifiers, and book counts/ranges when part of the core title; stop before separately bound works.
- `Elements Designation`: keep book counts/ranges when part of the designation, but avoid edition/enrichment clauses.
- `Base Content Description` vs `Enriched With`: follow title-page framing. Added/supplementary material goes to `Enriched With`; material framed as inherent core content may belong to `Base Content Description`.
- `Adapter Attribution`: include personal names and family/name-origin geography; exclude professional descriptors, offices, residences, and institutional settings.
- `Adapter Description`: professional descriptors always belong here; institutional/role/geographic settings belong here.
- `Place in Imprint`: V7 sets `location_in_imprint` to list-valued and should prefer full publication address/place phrases.
- `Verbs`: include coordinated action verbs and action participles; do not include surrounding non-verbal phrases.

## Suggested V7 Evaluation

1. Run V7 using the migration above.
2. Join V7 results to `v6_diagnostic_review.csv` by `Page/Key` and `Feature Name`.
3. Apply the target-selection order above.
4. Report exact diagnostic matches against the V6 baseline of 19/52.
5. Report per-feature results for the V6 weak clusters.
6. Run a KPI continuity check against `working_final_dataset_orig_rows.csv`, grouped by `decision_tier`.
7. Only propose V8 if V7 clearly fixes one family but breaks another.
