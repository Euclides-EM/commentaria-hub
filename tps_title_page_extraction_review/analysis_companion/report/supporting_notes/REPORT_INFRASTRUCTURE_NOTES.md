# Report Infrastructure Notes

Date: 2026-06-12

This note summarizes the first report-ready tables and visualization prototypes generated from the approved skeleton.

Script:

- `report/scripts/build_report_infrastructure.py`

Outputs:

- `report/tables/`
- `report/figures/`

## What Was Built

### Core Tables

- `report_corpus_accounting.csv`
- `report_subject_zone_counts.csv`
- `report_subject_social_rates_matrix.csv`
- `report_subject_intellectual_rates_matrix.csv`
- `report_elements_vs_non_elements_core_contrasts.csv`
- `report_elements_mode_marker_rates_matrix.csv`
- `report_elements_bookgroup_marker_rates_matrix.csv`
- `report_deductive_parts_by_corpus_matrix.csv`
- `report_deductive_parts_by_bookgroup_matrix.csv`
- period/language/format/book-group long tables for Elements-only markers
- PCA loading tables for full corpus and Elements-only plots

### First-Pass Figures

- `heatmap_subject_social_rates.png`
- `heatmap_subject_intellectual_rates.png`
- `bar_elements_vs_non_elements_top_contrasts.png`
- `heatmap_elements_mode_marker_rates.png`
- `heatmap_elements_bookgroup_marker_rates.png`
- `heatmap_deductive_parts_by_corpus.png`
- `heatmap_deductive_parts_by_bookgroup.png`
- `pca_full_corpus_report_features.png`
- `pca_elements_only_report_features.png`

## Immediate Historical Payoff

### 1. Elements vs Non-Elements Contrast Is Stronger And Cleaner Than Before

The strongest report-worthy contrast is not generic "theory versus practice." It is canonical mediation.

Top Elements-over-non-Elements contrasts:

| Marker | Elements | Non-Elements | Difference |
|---|---:|---:|---:|
| Ancient authority/restoration claim | 83.2% | 12.0% | +71.2 pp |
| Canonical/textual identity claim | 86.7% | 26.6% | +60.1 pp |
| Method/demonstration/order claim | 42.0% | 16.3% | +25.6 pp |
| Any named deductive part | 57.7% | 34.8% | +22.9 pp |
| Ancient restoration value | 20.3% | 2.5% | +17.8 pp |
| Translation/transfer claim | 26.2% | 8.4% | +17.8 pp |
| Translation/vernacular value | 16.8% | 2.7% | +14.1 pp |
| Augmentation/composition claim | 40.9% | 28.7% | +12.2 pp |
| Selection/extraction claim | 17.8% | 7.0% | +10.8 pp |

Non-Elements over-index utility/practice more directly:

| Marker | Elements | Non-Elements | Difference |
|---|---:|---:|---:|
| Utility/practice/application claim | 7.7% | 17.8% | -10.1 pp |
| Utility/application value | 4.2% | 8.3% | -4.1 pp |

Interpretation:

This supports the report's central distinction: Elements title pages present Euclid through canonical mediation, while neighboring mathematical books more often advertise direct practical/procedural usefulness.

### 2. The Subject-Ecology Section Needs To Emphasize Specific Grammars

The subject x intellectual-value heatmap shows that subject zones have different title-page grammars.

Useful trends:

- Instruments/Measurement has high augmentation/enrichment, demonstration/method, utility/application, and visual materiality.
- Arithmetic/Commerce has strong augmentation/enrichment, demonstration/method, correction/revision, and utility/application.
- Geometry/Theory has stronger ancient restoration and translation/vernacularization than most non-Elements zones, but not uniquely strong utility.
- Applied Mechanics/Military has demonstration/method, augmentation, correction, novelty, and visual materiality, though explicit military-user tags are rarer than the subject label might imply.
- Visual/Spatial Arts is comparatively lower on most intellectual-value markers, which may indicate either different title-page rhetoric or under-detection.

Interpretation:

The ecology section should not simply list subjects. It should show that title pages make different mathematical zones legible by different mixtures of method, utility, apparatus, correction, visuality, and authority.

### 3. Elements Book Groups Give Real Historical Texture

The book-group heatmap is one of the most useful report outputs.

Strong patterns:

- `books_1_6_plus_solids`:
  - method/demonstration/order: 64.1%;
  - propositions: 35.9%;
  - access/clarity/pedagogy: 35.9%;
  - utility/practice/application: 28.2%;
  - Jesuit: 33.3%.

- `books_1_6`:
  - method/demonstration/order: 48.1%;
  - any named deductive part: 63.0%;
  - translation/transfer: 21.0%;
  - selection/extraction: 32.1%;
  - visual aids: 13.6%.

- `near_complete_or_expanded`:
  - augmentation/composition: 55.2%;
  - translation/transfer: 41.8%;
  - demonstrations/proofs: 34.3%;
  - correction/revision: 34.3%;
  - scholia/commentary: 23.9%;
  - visual aids: 25.4%.

Interpretation:

Book coverage should remain a major report section. It is not just bibliographic description. It tracks different modes of recomposing Euclid:

- `1-6`: elementary/portable/foundational and sometimes practical-vernacular;
- `1-6 + 11-12`: usable institutional/practical-pedagogical Euclid, with propositions and method strongly visible;
- near-complete/expanded: learned, apparatus-rich, translated/corrected/demonstrated Euclid.

### 4. Elements Natural Modes Are Supported By The Heatmap

The mode heatmap reinforces the overlapping-mode model.

Useful trends:

- `practical-pedagogical` is the richest mode:
  - access/clarity/pedagogy: 62.5%;
  - any named deductive part: 79.2%;
  - method/demonstration/order: 56.9%;
  - utility/practice/application: 30.6%;
  - propositions: 19.4%;
  - scholia/commentary: 18.1%;
  - visual aids: 20.8%.

- `pedagogical/method` is especially demonstrative:
  - method/demonstration/order: 79.3%;
  - demonstrations/proofs: 44.8%;
  - propositions: 24.1%.

- `institutional-composite` has strong apparatus and learned mediation:
  - augmentation/composition: 56.6%;
  - demonstrations/proofs: 26.5%;
  - scholia/commentary: 23.0%;
  - translation/transfer: 32.7%;
  - visual aids: 26.5%.

- `humanist/vernacular transfer` is dominated by ancient/canonical identity and translation/transfer.

- `sparse/canonical` has canonical identity but very low deductive, method, apparatus, and social markers.

Interpretation:

The modes should be presented as historically meaningful postures, not as clean subcorpora. Their overlap and gradients are part of the argument.

### 5. Deductive Parts Sharpen The Intellectual Argument

The deductive-parts heatmap confirms:

- Metadata Elements are much more likely to name deductive parts overall: 57.7% vs 34.8%.
- Elements over-index demonstrations/proofs, propositions, scholia/commentary, principles, theorems, and enunciations.
- Non-Elements over-index problems, operations/constructions, examples, and notes/observations.
- Figures/diagrams are not Elements-specific.

Interpretation:

The report should say that Elements title pages construct Euclid as a demonstrative-commentarial corpus, not just as geometry. Practicality appears through transformations of that corpus, especially proposition-use, easier demonstrations, selection, figures, and method.

## Visual Assessment

### Strong / Likely Useful

- `heatmap_elements_bookgroup_marker_rates.png`
- `heatmap_elements_mode_marker_rates.png`
- `heatmap_deductive_parts_by_corpus.png`
- `bar_elements_vs_non_elements_top_contrasts.png`
- `heatmap_subject_intellectual_rates.png`

### Useful With Caution

- `heatmap_subject_social_rates.png`

Reason:

Many explicit social categories have low counts. The heatmap is still useful for showing sparseness and differentiating authority types, but it should not carry a heavy argument alone.

### Exploratory Only

- `pca_full_corpus_report_features.png`
- `pca_elements_only_report_features.png`

The PCA plots are interpretable but should be treated as diagnostics, not proof.

Full corpus PCA:

- PC1 appears to mark general title-page mediation density: deductive parts, method/order, visual aids, augmentation, pedagogy.
- PC2 separates ancient/canonical restoration from more instrumental, visual, utility, and operation-oriented directions.
- Explained variance is low: PC1 8.9%, PC2 5.5%.

Elements-only PCA:

- PC1 separates dense mediated Elements from sparse/low-signal edges.
- PC2 separates humanist/vernacular/ancient authority from practical-pedagogical/proposition-use directions.
- Explained variance is better but still modest: PC1 16.2%, PC2 8.4%.

Use PCA only if the report needs a visual showing gradients rather than categorical separation.

## What This Means For Next Work

The report can now start with Sections 1-3 or continue tying loose corners. The most useful remaining analyses are still:

1. proposition-use and demonstration-use by book group, language, period, and major edition clusters;
2. commentary split: ancient/humanist scholia versus pedagogical explanation/commentary;
3. clean bridge-case table for Elements-to-non-Elements gradients;
4. final casebook shortlist with one job per case.

Recommended next step:

Run the proposition/demonstration deepening, because it directly strengthens Sections 5, 8, and 9.
