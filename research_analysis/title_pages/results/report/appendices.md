# Appendices

**The Place of Euclid's Elements in Early Modern Mathematical Print: A Title-Page Ecology**

These appendices document the data sources, operational definitions, generated tables, figures, case clusters, and caveats behind the report. They are meant as reading support, not as a second narrative argument.

## Appendix A: Corpus And Data Construction

The report counts representative title-page records. The main object is the metadata-defined Elements corpus: editions manually identified in `ocrflow/store/items_metadata` as primarily editions of Euclid's *Elements*. The comparison field is the remaining representative mathematical title-page corpus.

Important construction rules:

- representative-row logic;
- metadata-defined Elements corpus definition;
- non-Elements comparison corpus definition;
- reprints and representative keys;
- subject classification source;
- title-page feature extraction source;
- metadata source: `ocrflow/store/items_metadata`;
- known exclusions and old files not used.

Core count used in the report:

| Unit | Count |
|---|---:|
| Representative title-page rows | 843 |
| Metadata-defined Elements representatives | 286 |
| Non-Elements representatives | 557 |
| Rows with at least one named deductive/mathematical part | 359 |

Key files for checking the corpus:

- `tables/corpus_accounting.csv`
- `../../data/analysis_ready/elements_ecology.csv`
- `../../data/analysis_ready/title_page_corpus.csv`

## Appendix B: Terms And Feature Definitions

The report uses human-readable terms in the prose, but each term corresponds to one or more feature fields. The table below is the quick audit map.

| Report term | Data source / feature | Working definition |
|---|---|---|
| Metadata-defined Elements corpus | metadata fields from `ocrflow/store/items_metadata` | Editions manually identified in the metadata as primarily Euclid's *Elements*. |
| Explicit Euclid/book identity | `claim_canonical_textual_identity` | Title-page language presenting the work through Euclid, the *Elements*, numbered books, textual corpus identity, edition identity, or author/title authority. |
| Ancient authority/restoration | `claim_ancient_authority_restoration`; related value markers | Ancient, Greek/Latin, recovered, corrected tradition, humanist authority, restoration of an older text. |
| Mediation | multiple claim/action fields | Advertised acts performed on knowledge: translating, correcting, explaining, demonstrating, selecting, augmenting, furnishing, reforming, adapting. |
| Utility/practice/application | `claim_utility_practice_application` | Use, practice, application, professional operation, measurement, construction, or applied procedure. |
| Named mathematical/deductive part | `deductive_parts.csv` | Propositions, demonstrations, scholia, figures, corollaries, problems, operations, examples, definitions, principles, etc. |
| Sparse authoritative presentation | sparse/canonical mode and related controls | Title pages relying heavily on Euclid/book identity with few explicit social/intellectual claims; use only with fashion controls. |

This table is the retained canonical terminology map. Extend it here when the report introduces a new analytical term.

## Appendix C: Core Tables

The report's main quantitative tables are generated as CSV files in `report/tables/`. The most important tables for rechecking the main claims are:

- `tables/elements_vs_non_elements_core_contrasts.csv`
- `tables/subject_intellectual_rates_matrix.csv`
- `tables/subject_social_rates_matrix.csv`
- `tables/elements_mode_marker_rates_matrix.csv`
- `tables/elements_book_group_marker_rates_matrix.csv`
- `tables/deductive_parts_by_corpus_matrix.csv`
- `tables/deductive_parts_by_book_group_matrix.csv`
- `tables/commentary_split_by_corpus_matrix.csv`
- `tables/commentary_by_elements_book_group_matrix.csv`
- `tables/proposition_demonstration_motifs_by_elements_book_group.csv`
- `tables/figures_diagrams_by_corpus_matrix.csv`
- `tables/bridge_case_route_marker_rates_matrix.csv`
- `tables/bridge_case_route_overlap_matrix.csv`
- `tables/format_elements_marker_rates_matrix.csv`
- `tables/format_elements_mode_rates_matrix.csv`
- `tables/format_elements_book_group_distribution.csv`
- `tables/format_elements_density_summary.csv`
- `tables/format_full_corpus_subject_rates.csv`

## Appendix D: Figures And Visualizations

Figures are generated PNG files in `report/figures/`. The recommended main-report figures are:

| Figure | File | Best section |
|---|---|---|
| Top Elements vs non-Elements contrasts | `figures/bar_elements_vs_non_elements_top_contrasts.png` | Executive argument / Section 3 |
| Subject zones by intellectual values | `figures/heatmap_subject_intellectual_rates.png` | Section 2 |
| Subject zones by social markers | `figures/heatmap_subject_social_rates.png` | Section 2, with caution |
| Bridge routes | `figures/heatmap_bridge_route_marker_rates.png` | Section 3 / Section 11 |
| Elements postures/modes | `figures/heatmap_elements_mode_marker_rates.png` | Section 4 |
| Elements book groups | `figures/heatmap_elements_book_group_marker_rates.png` | Section 5 |
| Deductive parts by corpus | `figures/heatmap_deductive_parts_by_corpus.png` | Section 9 |
| Figure/diagram functions by corpus | `figures/heatmap_figures_diagrams_by_corpus.png` | Section 9 |
| Figure/diagram functions by format | `figures/heatmap_figures_diagrams_by_format.png` | Section 6 |
| Elements modes by format | `figures/heatmap_format_elements_modes.png` | Section 6 |
| Elements social/intellectual markers by format | `figures/heatmap_format_elements_markers.png` | Section 6 |
| Commentary by book group | `figures/heatmap_commentary_by_elements_book_group.png` | Section 8 |

Diagnostic figures to use with caution:

- `figures/pca_full_corpus_features.png`
- `figures/pca_elements_features.png`

The PCA plots are useful diagnostics for gradients, but they should not carry a proof claim by themselves because explained variance is modest.

## Appendix E: Casebook

The report uses cases to anchor quantitative claims. Each case should have one main argumentative job, so that examples do not become decorative.

Primary source:

- `supporting_notes/REPORT_CASEBOOK_SHORTLIST.md`

Case clusters:

| Cluster | Cases | Main job |
|---|---|---|
| Vernacular/practical first-six-books Euclid | `Rotterdam_1647`, `Amsterdam_1616`, `Amsterdam_1700b`, `Amsterdam_1701` | Dutch/Dou route: practical Euclid without books 11-12. |
| Proposition-use practical pedagogy | `London_1685a`, `London_1703`, `Oxford_London_1700` | Uses of propositions, easy method, translation, correction. |
| Humanist/learned apparatus | `Urbino_1575`, `Rome_1591`, `Rome_1609`, `Paris_1622` | Translation, ancient scholia, figures, correction, demonstrations. |
| Jesuit/institutional apparatus | `Cologne_1591`, `Frankfurt_1607`, `Rome_1603` | Euclid as furnished institutional curriculum. |
| Reconstruction/logical restoration | `Paris_1667`, `Livorno_1709`, `Paris_1640`, `Cologne_1556` | New order, reform, reduced propositions, easier demonstrations. |
| Euclidean practical geometry outside Elements | `bib-9`, `bib-135` | Euclid as foundation/method outside metadata Elements corpus. |

## Appendix F: Methodological Caveats

These caveats should travel with the report's claims:

- Title pages are evidence of public framing, not full book contents.
- Metadata Elements is the report's corpus definition; title-page mentions of Euclid/Elements are comparison evidence, not corpus definition.
- Subject classification and title-page feature tags must be read together.
- Rich tags are useful for navigation and aggregation but should be verified by close reading for major claims.
- Sparse or silent title pages cannot be interpreted without period/place/language/format controls.
- Format is historical evidence and a control variable.
- Patronage is not readership.
- Author/editor portfolio evidence is corpus-internal, not full bibliography.

## Appendix G: Next-Direction Analysis Plan

The strongest future analysis would refine the report through controlled local comparisons.

Controlled local comparisons by:

- city or region;
- language;
- format;
- period/date window;
- subject zone;
- intersections such as city + language + format + period.

This would test whether the Elements/non-Elements contrasts survive local title-page fashion and material book conventions better than same-author analysis alone.
