# Title-Page Data Guide

This guide describes the retained datasets at the level needed to inspect, join, and audit them without reconstructing the old exploratory workflow.

The frozen inclusion lists are in `data/base_keys/`. In particular, title-page analyses use 909 title pages / 843 representatives, while print geography uses its frozen 320-edition Elements list even if live metadata later contains more keys.

## Which File Should I Open?

| Need | File | Rows | Grain / key |
|---|---|---:|---|
| Inspect extracted title-page phrases across the broad corpus | `data/raw/title_page_features_larger_corpus.csv` | 7,169 | One non-empty extracted feature per `edition_id` + `feature_id` |
| Audit the manually reviewed Elements-oriented extraction | `data/raw/title_page_features_reviewed_elements.csv` | 5,573 | One extracted feature decision per `edition_id` + `feature_id` |
| Inspect newly generated subject decisions | `data/raw/subject_classifications_new_representatives.csv` | 13,899 | One subject decision per representative edition + subject `feature_id` |
| Map title pages and reprints to subject representatives | `data/raw/title_page_to_subject_representative_map.csv` | 909 | One row per `title_page_key` |
| Work with the complete title-page corpus | `data/analysis_ready/title_page_corpus.csv` | 909 | One row per title page; primary key `title_page_key` |
| Compare representative editions and subject/title-page features | `data/analysis_ready/elements_ecology.csv` | 843 | One row per `classification_key` |
| Analyze metadata-defined Elements representatives and their modes | `data/analysis_ready/elements_modes.csv` | 286 | One row per Elements `classification_key` |
| Analyze social arenas across all representatives | `data/analysis_ready/social_arena_representatives.csv` | 843 | One row per `classification_key` |
| Analyze named mathematical parts | `data/analysis_ready/deductive_parts.csv` | 359 | One row per representative with at least one named part |

The earlier reviewed subject decisions for 155 editions are in `../edition_classification/data/edition_subject_classifications_reviewed.csv`. They complement `subject_classifications_new_representatives.csv`; the latter does not contain the earlier overlap.

Known QA exception: the new-classification file covers 695 representative editions and should contain 20 subject decisions per edition. `bib-96` has 19 because its `Cosmography` decision is missing. The report matrices retain this as an explicit unknown/missing-data caveat; do not silently interpret it as `unrelated`.

## Join Logic

- `title_page_key` identifies a specific title page or edition record.
- `classification_key` identifies the representative edition whose multi-label subject classification is inherited by reprints.
- Use `title_page_to_subject_representative_map.csv` to move between those levels.
- In raw extraction files, `edition_id` corresponds to the title-page/edition key.
- Subject classifications are multi-label. Values such as `primary`, `secondary`, `unrelated`, and `unknown` must not be collapsed silently into one genre.

## Raw Versus Analysis-Ready

The four files in `data/raw/` are retained extraction, classification, and mapping evidence. The five files in `data/analysis_ready/` contain joins and historically meaningful flags used by the report scripts.

The old exploration did not preserve a single complete program that recreates every analysis-ready join from the raw files. For that reason, the analysis-ready files are the reproducible input boundary for the report. Do not delete them as if they were disposable generated output.

## Important Column Families

The wide matrices intentionally preserve machine-usable evidence:

- bibliographic fields: `year`, `city`, `language`, `author_or_editor`, `publisher`, `format`;
- 20 subject columns, plus `primary_classes`, `secondary_classes`, and `unknown_classes`;
- extracted title-page features such as `audience`, `elements_designation`, `enriched_with`, and `references_to_euclid`;
- boolean marker families beginning `aud_`, `inst_`, `role_`, `pat_`, `ival_`, `claim_`, `part_`, or `mode_`;
- corpus controls such as `is_metadata_elements_representative`, `elements_books_group`, `period`, `format_group`, and `natural_dominant_mode`.

For exact interpretations and historical cautions, read `results/report/appendices.md` before treating a boolean marker as a historical fact.
