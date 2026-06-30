# Analysis Workspace

This is the place for question-by-question exploration of the larger title-page corpus.

## Main Data To Look At Now

- Latest larger-corpus TPS raw extraction: `data/latest_larger_corpus/printed_missing_tps_v8_preview.csv`
- Older reviewed Elements-oriented TPS extraction: `data/reviewed_elements_v8/v8_results_preview_reviewed.csv`
- Bibliographic metadata: `../ocrflow/store/items_metadata/items_print.csv`

## Subject Classification Status

The existing subject classification is partial. It covers only 155 editions, not the whole title-page corpus.

The partial files are archived here:

`data/subject_classification/partial_155_archive/`

Do not use that archive for the main analysis.

The full title-page-corpus classification is prepared but not yet run:

- Key lists: `data/subject_classification/full_run/`
- Commands: `data/subject_classification/full_run/RUN_COMMANDS.md`
- Recommended key file: `data/subject_classification/full_run/all_printed_title_page_corpus_909_keys.txt`

Once the full run finishes, we should parse the output CSV and create the real analysis files:

- full long classification table;
- positive-only long table;
- one-row-per-edition subject matrix;
- subject matrix joined to TPS feature summaries.

## Suggested Question Sequence After Classification

1. What is the corpus actually made of?
   Look at years, cities, languages, subjects, and which editions have non-empty TPS extraction rows.

2. What does "Elements" mean outside Euclid?
   Compare `elements_designation`, `base_content`, and `references_to_euclid` across subject profiles.

3. When Euclid appears, what role does he play?
   Separate Euclid as author/title, Euclid as authority, Euclid as method, Euclid as brand, and Euclid as part of a bound-with scholarly package.

4. What kinds of books use similar title-page formulas?
   Compare Euclid/Elements title pages against arithmetic, practical geometry, perspective, instruments, surveying, architecture, military engineering, and other classified subjects.

5. How do title pages define communities of use?
   Inspect `audience`, `institutions`, `editor_description`, `dedicatee_name`, and imprint fields.

6. What does "text in motion" look like now?
   Use `enriched_with`, `bound_with`, `bound_with_minimal`, `edition_details`, `action_verbs`, and references to other authorities.

7. Which examples are presentation-worthy?
   Build a shortlist of cases that are historically surprising, visually strong, or useful because they contradict an earlier assumption.

## Current Working Caution

The older presentation framed the Elements as a stable ancient canon with active adaptation around it. The larger corpus may suggest a more complicated story, but we should wait for the full subject classification before making genre-level claims.
