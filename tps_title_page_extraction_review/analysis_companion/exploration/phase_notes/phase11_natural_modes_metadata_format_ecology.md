# Phase 11 Natural Elements Modes, Metadata, Format, And Ecology

This pass follows the corrected agenda from 2026-06-11: keep the metadata-defined Elements corpus at the center, map the Phase 7 natural modes against each other and against metadata fields, and add bibliographic format as a control.

Inputs:

- `derived_data/metadata_elements_natural_modes_matrix.csv`
- `ocrflow/store/items_metadata/items_print.csv` for `format`, `volumes`, and `has_diagrams`
- `derived_data/representative_analysis_matrix_rich_social_arenas.csv` for broader-corpus comparison

New enriched matrix:

- `derived_data/metadata_elements_natural_modes_matrix_with_format.csv`

## Mode Counts

There are eight operational mode flags. The earlier phrase "six-ish modes" should be read as shorthand; the current implementation uses these eight.

| mode | column | count | pct_of_metadata_elements |
| --- | --- | --- | --- |
| humanist/ancient | mode_humanist_ancient | 243 | 85.0 |
| institutional/authority | mode_institutional_authority | 219 | 76.6 |
| pedagogical/method | mode_pedagogical_method | 201 | 70.3 |
| composite/apparatus | mode_composite_apparatus | 196 | 68.5 |
| corrected/updated | mode_corrected_updated | 144 | 50.3 |
| vernacular/transfer | mode_vernacular_transfer | 106 | 37.1 |
| practical/public | mode_practical_public | 81 | 28.3 |
| sparse/canonical | mode_sparse_canonical | 44 | 15.4 |

## Main Overlap Result

The modes are not separate subcorpora. They form overlapping bundles.

The dense center is:

- humanist/ancient;
- institutional/authority;
- pedagogical/method;
- composite/apparatus.

These four repeatedly overlap with each other. Practical/public is smaller, but it usually sits on top of pedagogical/method and canonical/humanist authority rather than outside the Elements tradition. Sparse/canonical is the most separate mode and therefore the one most vulnerable to fashion/silence controls.

Strongest pair overlaps:

| mode_a | mode_b | count_both | pct_all | jaccard | pct_of_mode_a | pct_of_mode_b |
| --- | --- | --- | --- | --- | --- | --- |
| humanist/ancient | institutional/authority | 195 | 68.2 | 0.73 | 80.2 | 89.0 |
| humanist/ancient | pedagogical/method | 187 | 65.4 | 0.728 | 77.0 | 93.0 |
| institutional/authority | pedagogical/method | 176 | 61.5 | 0.721 | 80.4 | 87.6 |
| composite/apparatus | pedagogical/method | 162 | 56.6 | 0.689 | 82.7 | 80.6 |
| composite/apparatus | institutional/authority | 169 | 59.1 | 0.687 | 86.2 | 77.2 |
| composite/apparatus | humanist/ancient | 176 | 61.5 | 0.669 | 89.8 | 72.4 |
| composite/apparatus | corrected/updated | 121 | 42.3 | 0.553 | 61.7 | 84.0 |
| corrected/updated | pedagogical/method | 121 | 42.3 | 0.54 | 84.0 | 60.2 |
| corrected/updated | institutional/authority | 126 | 44.1 | 0.532 | 87.5 | 57.5 |
| corrected/updated | humanist/ancient | 133 | 46.5 | 0.524 | 92.4 | 54.7 |
| humanist/ancient | vernacular/transfer | 99 | 34.6 | 0.396 | 40.7 | 93.4 |
| composite/apparatus | vernacular/transfer | 84 | 29.4 | 0.385 | 42.9 | 79.2 |
| institutional/authority | vernacular/transfer | 87 | 30.4 | 0.366 | 39.7 | 82.1 |
| corrected/updated | vernacular/transfer | 64 | 22.4 | 0.344 | 44.4 | 60.4 |
| pedagogical/method | practical/public | 72 | 25.2 | 0.343 | 35.8 | 88.9 |
| pedagogical/method | vernacular/transfer | 78 | 27.3 | 0.341 | 38.8 | 73.6 |
| humanist/ancient | practical/public | 75 | 26.2 | 0.301 | 30.9 | 92.6 |
| institutional/authority | practical/public | 69 | 24.1 | 0.299 | 31.5 | 85.2 |
| composite/apparatus | practical/public | 62 | 21.7 | 0.288 | 31.6 | 76.5 |
| corrected/updated | practical/public | 47 | 16.4 | 0.264 | 32.6 | 58.0 |

Interpretation:

- 93.0% of pedagogical/method cases are also humanist/ancient.
- 87.6% of pedagogical/method cases are also institutional/authority.
- 82.7% of composite/apparatus cases are also pedagogical/method.
- 88.9% of practical/public cases are also pedagogical/method.
- Sparse/canonical overlaps weakly with institutional/composite/pedagogical modes, but 88.6% of sparse/canonical cases still carry humanist/ancient authority.

So the Elements corpus is not divided into clean species. It has a large institutional-humanist-pedagogical-composite core, with practical/public and sparse/canonical as differently positioned edges.

## Book Coverage Relationships

| value | n | composite/apparatus | corrected/updated | humanist/ancient | institutional/authority | pedagogical/method | practical/public | sparse/canonical | vernacular/transfer |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| books_1_13 | 5 | 20.0 | 80.0 | 100.0 | 60.0 | 60.0 | 60.0 | 40.0 | 20.0 |
| books_1_6 | 81 | 63.0 | 45.7 | 82.7 | 76.5 | 67.9 | 23.5 | 23.5 | 37.0 |
| books_1_6_plus_solids | 39 | 59.0 | 59.0 | 97.4 | 89.7 | 79.5 | 59.0 | 10.3 | 23.1 |
| mixed_or_other_books | 1 | 100.0 | 100.0 | 100.0 | 100.0 | 100.0 | 100.0 | 0.0 | 100.0 |
| near_complete_or_expanded | 67 | 83.6 | 61.2 | 94.0 | 77.6 | 88.1 | 17.9 | 9.0 | 50.7 |
| near_complete_or_expanded_enunciations | 7 | 100.0 | 28.6 | 85.7 | 57.1 | 85.7 | 57.1 | 0.0 | 57.1 |
| partial_from_book_1 | 8 | 50.0 | 50.0 | 62.5 | 75.0 | 50.0 | 37.5 | 0.0 | 12.5 |
| selected_later_books | 21 | 61.9 | 33.3 | 90.5 | 71.4 | 57.1 | 14.3 | 14.3 | 28.6 |
| selected_later_books_enunciations | 1 | 0.0 | 0.0 | 100.0 | 100.0 | 100.0 | 0.0 | 0.0 | 0.0 |
| unknown_books | 56 | 71.4 | 44.6 | 67.9 | 71.4 | 51.8 | 23.2 | 17.9 | 35.7 |

Reading:

- `books_1_6_plus_solids` is the most practical/public among the large book-coverage groups: 59.0% practical/public.
- `near_complete_or_expanded` is highly composite/apparatus and pedagogical/method.
- plain `books_1_6` is mixed: more sparse/canonical than `1-6 + 11-12`, but still significantly pedagogical/method and composite/apparatus.
- `near_complete_or_expanded_enunciations` is small but rhetorically loud: composite, pedagogical, practical, and vernacular/transfer.

## Format Relationships

| value | n | composite/apparatus | corrected/updated | humanist/ancient | institutional/authority | pedagogical/method | practical/public | sparse/canonical | vernacular/transfer |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 18 | 1 | 0.0 | 0.0 | 100.0 | 100.0 | 0.0 | 0.0 | 100.0 | 0.0 |
| 6 | 5 | 40.0 | 40.0 | 100.0 | 60.0 | 40.0 | 20.0 | 40.0 | 20.0 |
| duodecimo | 30 | 60.0 | 60.0 | 90.0 | 90.0 | 76.7 | 40.0 | 16.7 | 20.0 |
| folio | 54 | 85.2 | 55.6 | 81.5 | 75.9 | 72.2 | 22.2 | 9.3 | 38.9 |
| missing/unknown | 18 | 50.0 | 50.0 | 72.2 | 55.6 | 55.6 | 27.8 | 33.3 | 22.2 |
| octavo | 111 | 69.4 | 50.5 | 91.0 | 78.4 | 79.3 | 27.9 | 15.3 | 34.2 |
| quarto | 67 | 65.7 | 43.3 | 77.6 | 74.6 | 58.2 | 29.9 | 11.9 | 53.7 |

Reading:

- Folios over-index composite/apparatus: 85.2% versus 68.5% overall.
- Quartos over-index vernacular/transfer: 53.7% versus 37.1% overall.
- Duodecimos are highly institutional/authority at 90.0%, pedagogical/method at 76.7%, and practical/public at 40.0%.
- Octavos are close to the corpus center but slightly high in pedagogical/method.
- Sparse/canonical is not explained by format alone. It appears in several formats and is especially high in missing/unknown and small-n formats, so it needs period/place/language controls before interpretation.

## Language And Period Relationships

Languages with at least six Elements representatives:

| value | n | composite/apparatus | corrected/updated | humanist/ancient | institutional/authority | pedagogical/method | practical/public | sparse/canonical | vernacular/transfer |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| DUTCH | 21 | 76.2 | 57.1 | 90.5 | 95.2 | 81.0 | 47.6 | 9.5 | 52.4 |
| ENGLISH | 20 | 70.0 | 55.0 | 85.0 | 90.0 | 90.0 | 70.0 | 10.0 | 50.0 |
| FRENCH | 53 | 73.6 | 54.7 | 79.2 | 88.7 | 75.5 | 41.5 | 5.7 | 37.7 |
| GERMAN | 13 | 76.9 | 30.8 | 84.6 | 76.9 | 61.5 | 38.5 | 15.4 | 84.6 |
| GREEK | 16 | 68.8 | 50.0 | 100.0 | 50.0 | 50.0 | 0.0 | 31.2 | 50.0 |
| ITALIAN | 12 | 33.3 | 41.7 | 100.0 | 66.7 | 41.7 | 25.0 | 58.3 | 58.3 |
| LATIN | 139 | 69.1 | 51.8 | 83.5 | 71.9 | 71.2 | 17.3 | 13.7 | 24.5 |
| SPANISH | 6 | 83.3 | 16.7 | 100.0 | 100.0 | 66.7 | 50.0 | 16.7 | 50.0 |

Periods:

| value | n | composite/apparatus | corrected/updated | humanist/ancient | institutional/authority | pedagogical/method | practical/public | sparse/canonical | vernacular/transfer |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1550-1599 | 55 | 81.8 | 47.3 | 87.3 | 70.9 | 63.6 | 20.0 | 18.2 | 63.6 |
| 1600-1649 | 75 | 70.7 | 52.0 | 85.3 | 80.0 | 80.0 | 29.3 | 6.7 | 38.7 |
| 1650-1699 | 83 | 60.2 | 54.2 | 83.1 | 85.5 | 73.5 | 32.5 | 14.5 | 25.3 |
| 1700+ | 40 | 65.0 | 55.0 | 87.5 | 80.0 | 75.0 | 40.0 | 20.0 | 30.0 |
| pre-1550 | 32 | 65.6 | 37.5 | 81.2 | 50.0 | 46.9 | 15.6 | 25.0 | 25.0 |
| unknown | 1 | 100.0 | 0.0 | 100.0 | 100.0 | 0.0 | 0.0 | 100.0 | 100.0 |

Reading:

- English Elements are strongly practical/public and pedagogical/method.
- Dutch Elements are institutional, pedagogical, practical/public, and vernacular/transfer; this supports the idea of a practical-vernacular route parallel to the later `1-6 + 11-12` route.
- German Elements are strongly vernacular/transfer.
- Italian Elements show high sparse/canonical and vernacular/transfer, but the n is only 12 and needs close reading.
- Latin is large and central but less practical/public than English/French/Dutch/Spanish.
- Practical/public rises over time: 15.6% pre-1550, 20.0% in 1550-1599, 29.3% in 1600-1649, 32.5% in 1650-1699, 40.0% after 1700.
- Vernacular/transfer is especially strong in 1550-1599, then declines.

## Strong Metadata Signals

Strong over-indexed mode-field combinations, using groups with at least five cases:

| field | value | n | mode | count | pct | overall_pct | delta_vs_overall |
| --- | --- | --- | --- | --- | --- | --- | --- |
| city | Bologna | 6 | sparse/canonical | 5 | 83.3 | 15.4 | 67.9 |
| language_first | GERMAN | 13 | vernacular/transfer | 11 | 84.6 | 37.1 | 47.6 |
| city | Leipzig | 5 | sparse/canonical | 3 | 60.0 | 15.4 | 44.6 |
| language_first | ITALIAN | 12 | sparse/canonical | 7 | 58.3 | 15.4 | 42.9 |
| language_first | ENGLISH | 20 | practical/public | 14 | 70.0 | 28.3 | 41.7 |
| city | Basel | 8 | vernacular/transfer | 6 | 75.0 | 37.1 | 37.9 |
| metadata_additional_optics_catoptrics | True | 14 | vernacular/transfer | 10 | 71.4 | 37.1 | 34.4 |
| elements_books_group | books_1_13 | 5 | practical/public | 3 | 60.0 | 28.3 | 31.7 |
| metadata_has_additional_content | True | 55 | composite/apparatus | 55 | 100.0 | 68.5 | 31.5 |
| metadata_additional_data | True | 20 | composite/apparatus | 20 | 100.0 | 68.5 | 31.5 |
| metadata_additional_optics_catoptrics | True | 14 | composite/apparatus | 14 | 100.0 | 68.5 | 31.5 |
| metadata_additional_music_harmonics | True | 9 | composite/apparatus | 9 | 100.0 | 68.5 | 31.5 |
| city | Basel | 8 | composite/apparatus | 8 | 100.0 | 68.5 | 31.5 |
| elements_books_group | near_complete_or_expanded_enunciations | 7 | composite/apparatus | 7 | 100.0 | 68.5 | 31.5 |
| has_diagrams_metadata | True | 7 | composite/apparatus | 7 | 100.0 | 68.5 | 31.5 |
| elements_books_group | books_1_6_plus_solids | 39 | practical/public | 23 | 59.0 | 28.3 | 30.7 |
| elements_books_group | books_1_13 | 5 | corrected/updated | 4 | 80.0 | 50.3 | 29.7 |
| elements_books_group | near_complete_or_expanded_enunciations | 7 | practical/public | 4 | 57.1 | 28.3 | 28.8 |
| city | London | 20 | practical/public | 11 | 55.0 | 28.3 | 26.7 |
| period | 1550-1599 | 55 | vernacular/transfer | 35 | 63.6 | 37.1 | 26.6 |
| metadata_additional_data | True | 20 | pedagogical/method | 19 | 95.0 | 70.3 | 24.7 |
| city | Lyon | 10 | sparse/canonical | 4 | 40.0 | 15.4 | 24.6 |
| city | Strasbourg | 13 | vernacular/transfer | 8 | 61.5 | 37.1 | 24.5 |
| city | Lyon | 10 | institutional/authority | 10 | 100.0 | 76.6 | 23.4 |
| language_first | SPANISH | 6 | institutional/authority | 6 | 100.0 | 76.6 | 23.4 |
| city | Leiden | 6 | institutional/authority | 6 | 100.0 | 76.6 | 23.4 |
| city | Amsterdam | 15 | corrected/updated | 11 | 73.3 | 50.3 | 23.0 |
| language_first | SPANISH | 6 | practical/public | 3 | 50.0 | 28.3 | 21.7 |
| language_first | ITALIAN | 12 | vernacular/transfer | 7 | 58.3 | 37.1 | 21.3 |
| elements_books_group | near_complete_or_expanded_enunciations | 7 | vernacular/transfer | 4 | 57.1 | 37.1 | 20.1 |
| has_diagrams_metadata | True | 7 | vernacular/transfer | 4 | 57.1 | 37.1 | 20.1 |
| language_first | ENGLISH | 20 | pedagogical/method | 18 | 90.0 | 70.3 | 19.7 |
| city | London | 20 | pedagogical/method | 18 | 90.0 | 70.3 | 19.7 |
| language_first | DUTCH | 21 | practical/public | 10 | 47.6 | 28.3 | 19.3 |
| language_first | DUTCH | 21 | institutional/authority | 20 | 95.2 | 76.6 | 18.7 |
| city | Cologne | 9 | pedagogical/method | 8 | 88.9 | 70.3 | 18.6 |
| city | Amsterdam | 15 | practical/public | 7 | 46.7 | 28.3 | 18.3 |
| format_group | missing/unknown | 18 | sparse/canonical | 6 | 33.3 | 15.4 | 17.9 |
| format_first | missing/unknown | 18 | sparse/canonical | 6 | 33.3 | 15.4 | 17.9 |
| elements_books_group | near_complete_or_expanded | 67 | pedagogical/method | 59 | 88.1 | 70.3 | 17.8 |
| city | Amsterdam | 15 | institutional/authority | 14 | 93.3 | 76.6 | 16.8 |
| format_group | folio | 54 | composite/apparatus | 46 | 85.2 | 68.5 | 16.7 |
| format_first | 2 | 54 | composite/apparatus | 46 | 85.2 | 68.5 | 16.7 |
| format_group | quarto | 67 | vernacular/transfer | 36 | 53.7 | 37.1 | 16.7 |
| format_first | 4 | 67 | vernacular/transfer | 36 | 53.7 | 37.1 | 16.7 |

Reading:

- Additional content is almost mechanically tied to composite/apparatus, as expected.
- Additional Data is also strongly pedagogical/method.
- Additional optics/catoptrics strongly marks vernacular/transfer as well as composite/apparatus.
- Books `1-6 + 11-12` strongly mark practical/public.
- English and London strongly mark practical/public and pedagogical/method.
- German strongly marks vernacular/transfer.
- Folio marks composite/apparatus; quarto marks vernacular/transfer.

## Relationship To The Broader Corpus

For each Elements mode, I compared the mode's cases with all non-metadata-Elements representatives. This asks: what makes this Elements subset distinctive relative to the wider mathematical ecology?

Top mode-specific contrasts:

| mode | feature | mode_n | mode_pct | non_elements_pct | delta_vs_non_elements |
| --- | --- | --- | --- | --- | --- |
| sparse/canonical | claim_ancient_authority_restoration | 44 | 88.6 | 12.0 | 76.6 |
| sparse/canonical | claim_canonical_textual_identity | 44 | 100.0 | 26.6 | 73.4 |
| sparse/canonical | sparse_canonical_identity | 44 | 43.2 | 6.3 | 36.9 |
| sparse/canonical | no_visible_social_arena | 44 | 90.9 | 60.5 | 30.4 |
| pedagogical/method | claim_ancient_authority_restoration | 201 | 92.5 | 12.0 | 80.5 |
| pedagogical/method | procedural_pedagogical_identity | 201 | 81.1 | 10.6 | 70.5 |
| pedagogical/method | claim_canonical_textual_identity | 201 | 96.0 | 26.6 | 69.4 |
| pedagogical/method | claim_method_demonstration_order | 201 | 59.7 | 16.3 | 43.4 |
| vernacular/transfer | claim_ancient_authority_restoration | 106 | 93.4 | 12.0 | 81.4 |
| vernacular/transfer | claim_canonical_textual_identity | 106 | 94.3 | 26.6 | 67.8 |
| vernacular/transfer | humanist_transfer_book | 106 | 67.0 | 1.6 | 65.4 |
| vernacular/transfer | claim_translation_vernacularization_transfer | 106 | 70.8 | 8.4 | 62.3 |
| institutional/authority | claim_ancient_authority_restoration | 219 | 88.1 | 12.0 | 76.1 |
| institutional/authority | claim_canonical_textual_identity | 219 | 91.8 | 26.6 | 65.2 |
| institutional/authority | procedural_pedagogical_identity | 219 | 64.4 | 10.6 | 53.8 |
| institutional/authority | claim_method_demonstration_order | 219 | 47.9 | 16.3 | 31.6 |
| composite/apparatus | claim_ancient_authority_restoration | 196 | 88.3 | 12.0 | 76.2 |
| composite/apparatus | claim_canonical_textual_identity | 196 | 90.3 | 26.6 | 63.7 |
| composite/apparatus | procedural_pedagogical_identity | 196 | 69.9 | 10.6 | 59.3 |
| composite/apparatus | claim_method_demonstration_order | 196 | 48.0 | 16.3 | 31.6 |
| practical/public | claim_ancient_authority_restoration | 81 | 92.6 | 12.0 | 80.6 |
| practical/public | procedural_pedagogical_identity | 81 | 85.2 | 10.6 | 74.6 |
| practical/public | claim_canonical_textual_identity | 81 | 97.5 | 26.6 | 71.0 |
| practical/public | claim_accessibility_clarity_pedagogy | 81 | 55.6 | 9.3 | 46.2 |
| corrected/updated | claim_ancient_authority_restoration | 144 | 90.3 | 12.0 | 78.2 |
| corrected/updated | claim_canonical_textual_identity | 144 | 92.4 | 26.6 | 65.8 |
| corrected/updated | procedural_pedagogical_identity | 144 | 75.7 | 10.6 | 65.1 |
| corrected/updated | claim_method_demonstration_order | 144 | 50.7 | 16.3 | 34.4 |
| humanist/ancient | claim_ancient_authority_restoration | 243 | 97.9 | 12.0 | 85.9 |
| humanist/ancient | claim_canonical_textual_identity | 243 | 97.5 | 26.6 | 71.0 |
| humanist/ancient | procedural_pedagogical_identity | 243 | 66.3 | 10.6 | 55.7 |
| humanist/ancient | claim_method_demonstration_order | 243 | 45.7 | 16.3 | 29.3 |

Reading:

The Elements subset is distinctive not because it alone has method, utility, or pedagogy, but because it binds those values to ancient/canonical authority. Non-Elements books also talk about method, usefulness, pedagogy, and practice. The special thing about Elements editions is that these values are attached to Euclid/Elements as an ancient, canonical authority. So "useful Euclid" is not the same as a generic useful mathematics book.

Mode-specific notes:

- Sparse/canonical Elements: title pages that mostly say "Euclid / Elements / book identity" and little else. But we must check whether that quietness is meaningful or just title-page fashion.
- Pedagogical/method Elements: editions that present Euclid as something taught through order, method, demonstration, explanation, or ease.
- Vernacular/transfer Elements: editions where the big action is moving Euclid across language, audience, or cultural setting.
- Institutional/authority Elements: editions authorized by schools, colleges, religious orders, professors, academies, or learned credentials.
- Composite/apparatus Elements: editions that build Euclid into a larger teaching or research apparatus: additions, notes, figures, tables, other texts, Archimedes, Data, Optics, and similar material.
- Practical/public Elements: editions that make Euclid useful for broader publics or practical settings, but still keep Euclid's canonical authority.

## Fashion / Density Controls: Quick Diagnostic

This is not a final regression model. It is a warning system for the sparse/dense problem.

Top associations between density/silence measures and period/language/city/format/subject:

| subset | x_field | y_metric | association_type | association | n |
| --- | --- | --- | --- | --- | --- |
| metadata_elements | city | no_visible_social_arena | cramers_v_for_binary | 0.32 | 286 |
| non_elements | city | no_visible_social_arena | cramers_v_for_binary | 0.315 | 557 |
| non_elements | primary_subject_family | sparse_canonical_identity | cramers_v_for_binary | 0.302 | 557 |
| all_representatives | city | no_visible_social_arena | cramers_v_for_binary | 0.3 | 843 |
| non_elements | primary_subject_family | no_visible_social_arena | cramers_v_for_binary | 0.299 | 557 |
| metadata_elements | city | sparse_canonical_identity | cramers_v_for_binary | 0.276 | 286 |
| all_representatives | primary_subject_family | no_visible_social_arena | cramers_v_for_binary | 0.242 | 843 |
| metadata_elements | language | no_visible_social_arena | cramers_v_for_binary | 0.232 | 286 |
| all_representatives | primary_subject_family | sparse_canonical_identity | cramers_v_for_binary | 0.226 | 843 |
| metadata_elements | period | sparse_canonical_identity | cramers_v_for_binary | 0.214 | 286 |
| non_elements | city | sparse_canonical_identity | cramers_v_for_binary | 0.199 | 557 |
| all_representatives | city | sparse_canonical_identity | cramers_v_for_binary | 0.198 | 843 |
| metadata_elements | primary_subject_family | no_visible_social_arena | cramers_v_for_binary | 0.19 | 286 |
| metadata_elements | format_group | sparse_canonical_identity | cramers_v_for_binary | 0.187 | 286 |
| metadata_elements | format_group | no_visible_social_arena | cramers_v_for_binary | 0.181 | 286 |
| non_elements | language | no_visible_social_arena | cramers_v_for_binary | 0.181 | 557 |
| non_elements | language | sparse_canonical_identity | cramers_v_for_binary | 0.151 | 557 |
| metadata_elements | language | sparse_canonical_identity | cramers_v_for_binary | 0.148 | 286 |
| all_representatives | language | no_visible_social_arena | cramers_v_for_binary | 0.146 | 843 |
| all_representatives | format_group | no_visible_social_arena | cramers_v_for_binary | 0.143 | 843 |

Reading:

- City has the strongest association with `no_visible_social_arena` in both metadata Elements and non-Elements books.
- Format has a moderate association with sparse-canonical and no-visible-social behavior inside metadata Elements.
- Period and language also matter, especially for sparse-canonical inside metadata Elements.
- Therefore, sparse-canonical cannot be used as a pure intellectual signal without controls. It may be canonical silence, but it may also be local/period/format/title-page fashion.

## Preliminary Author / Editor Portfolio Signals

This is only a first portfolio table, not a finished analysis. It shows where a later author/editor trajectory pass should begin.

| author_or_editor | metadata_elements_reps | non_elements_reps | cities | languages | formats | elements_bookgroups | claim_method_demonstration_order_elements_pct | claim_utility_practice_application_elements_pct | claim_accessibility_clarity_pedagogy_elements_pct | religious_institutional_arena_elements_pct | military_fortification_arena_elements_pct |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Claude-François Milliet Dechales | 13 | 1 | Amsterdam \| Lausanne \| London \| Lyon \| Oxford \| Oxford, London \| Paris | ENGLISH \| FRENCH \| LATIN | 12 \| 18 \| 4 \| 6 \| 8 | books_1_6_plus_solids | 69.2 | 61.5 | 69.2 | 76.9 | 0.0 |
| Jan Pietersz Dou | 10 | 1 | Amsterdam \| Dordrecht \| Leiden \| Rotterdam | DUTCH | 4 \| 8 | books_1_6 | 100.0 | 40.0 | 50.0 | 0.0 | 0.0 |
| Denis Henrion | 9 | 17 | Lyon \| Paris \| Rouen \| Rouen, Paris | FRENCH | 4 \| 8 | books_1_6 \| near_complete_or_expanded | 0.0 | 0.0 | 33.3 | 0.0 | 0.0 |
| Conrad Dasypodius | 9 | 0 | Strasbourg | GREEK, LATIN \| LATIN | 4 \| 8 | near_complete_or_expanded \| selected_later_books \| unknown_books | 55.6 | 0.0 | 0.0 | 0.0 | 0.0 |
| Isaac Barrow | 9 | 0 | Cambridge \| London \| Osnabrück | ENGLISH \| LATIN | 12 \| 6 \| 8 | near_complete_or_expanded \| selected_later_books \| unknown_books | 100.0 | 0.0 | 11.1 | 0.0 | 0.0 |
| André Tacquet | 7 | 0 | Amsterdam \| Antwerp \| Leuven \| Padua | LATIN | 8 | books_1_6_plus_solids \| unknown_books | 85.7 | 0.0 | 0.0 | 100.0 | 0.0 |
| Jean Magnien, Étienne Gracile | 7 | 0 | Cologne \| Paris | LATIN | 8 | near_complete_or_expanded \| near_complete_or_expanded_enunciations \| unknown_books | 0.0 | 0.0 | 57.1 | 0.0 | 0.0 |
| Johannes de Sacrobosco | 7 | 0 | Paris | LATIN | 2 | unknown_books | 0.0 | 14.3 | 0.0 | 0.0 | 0.0 |
| Christopher Clavius | 6 | 4 | Cologne \| Frankfurt \| Mainz \| Rome | LATIN | 2 \| 4 \| 8 | near_complete_or_expanded | 83.3 | 0.0 | 0.0 | 100.0 | 0.0 |
| Georges Fournier | 5 | 1 | Cambridge \| London \| Paris | FRENCH \| LATIN | 12 \| 8 | books_1_6 | 80.0 | 0.0 | 0.0 | 80.0 | 0.0 |
| Pierre Forcadel | 4 | 13 | Paris | FRENCH | 4 \| 8 | books_1_6 \| selected_later_books \| unknown_books | 0.0 | 0.0 | 0.0 | 0.0 | 0.0 |
| Pierre Hérigone | 4 | 1 | Paris | FRENCH \| FRENCH, LATIN \| LATIN | 8 | books_1_6 \| near_complete_or_expanded | 100.0 | 0.0 | 0.0 | 0.0 | 0.0 |
| Christoffer Dybvad | 4 | 0 | Arnhem \| Lyon | LATIN | 4 | books_1_6 \| selected_later_books | 75.0 | 0.0 | 0.0 | 0.0 | 0.0 |
| Jean de la Pène | 4 | 0 | Leiden \| Paris | GREEK, LATIN \| LATIN | 4 | unknown_books | 0.0 | 0.0 | 0.0 | 100.0 | 0.0 |
| Pierre Le Mardelé | 4 | 0 | Lyon \| Paris | FRENCH \| FRENCH, LATIN | 8 | near_complete_or_expanded | 100.0 | 0.0 | 0.0 | 100.0 | 0.0 |
| Jean Errard | 3 | 4 | Paris | FRENCH | 4 \| 8 | books_1_6 \| partial_from_book_1 | 0.0 | 0.0 | 0.0 | 0.0 | 0.0 |
| Jacques Ozanam | 3 | 3 | Amsterdam \| Paris | FRENCH \| FRENCH, LATIN | 8 | books_1_6_plus_solids | 0.0 | 0.0 | 0.0 | 0.0 | 100.0 |
| Antonio Possevino | 3 | 0 | Cologne \| Rome \| Venice | LATIN | 2 | unknown_books | 0.0 | 0.0 | 0.0 | 100.0 | 0.0 |

Early reading:

- Dechales is concentrated in `1-6 + 11-12`, across several cities/languages/formats, with high method/ease/utility and religious-institutional framing.
- Dou is concentrated in plain `1-6`, Dutch, Amsterdam/Leiden/Rotterdam/Dordrecht, highly method/composite/corrected and practical/public.
- Denis Henrion has both Elements and many non-Elements works; his portfolio is ideal for testing whether Elements/non-Elements title pages share ideals or diverge by genre.
- Clavius, Tacquet, Barrow, Dasypodius, and Fournier are strong candidates for controlled author/editor micro-studies.

## Provisional Interpretation

The Elements corpus has no clean natural subcorpora. Instead, it has overlapping modes arranged around a dense center:

**ancient/canonical authority + institution + pedagogy/method + apparatus.**

The edges matter:

- `practical/public` shows how the Elements becomes usable without ceasing to be canonical;
- `vernacular/transfer` shows how the Elements travels across language and publics;
- `sparse/canonical` shows possible reliance on canonical identity, but must be controlled for title-page fashion;
- `1-6 + 11-12` is one especially strong practical/public package;
- Dutch plain `1-6` looks like a different practical-vernacular route.

## Next Questions

1. Run a focused density/fashion-control analysis for sparse-canonical: period + language + city + format + school/institution markers.
2. Run the author/editor portfolio analysis: compare Elements and non-Elements works by the same author/editor across city, language, format, and claims.
3. Deepen the Dutch plain `1-6` route: compare it with `1-6 + 11-12` and with non-Elements practical geometry.
4. Build a close-reading set for the central overlap cases versus edge cases.
