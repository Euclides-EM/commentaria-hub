# Phase 6 Metadata Elements Corpus Ecology

This phase switches from **Euclid/Elements as title-page wording** to the user-defined/metadata-defined corpus of books that are primarily *the Elements*.

Membership source: `ocrflow/store/items_metadata/metadata_elements_print.csv`.

This is not the same as the corpus of title pages that contain Elements-like wording. The distinction matters.

## Set Overlap

| category | count | pct_of_representatives | pct_of_metadata_elements_representatives |
| --- | --- | --- | --- |
| all representatives in title-page analysis corpus | 843 | 100.0 |  |
| metadata Elements edition keys total | 321 |  |  |
| metadata Elements edition keys present in title-page corpus | 320 |  |  |
| metadata Elements edition keys not in title-page corpus | 1 |  |  |
| metadata Elements representatives in title-page corpus | 286 | 33.9 | 100.0 |
| title-page Euclid/Elements signal | 332 | 39.4 |  |
| both metadata Elements and title-page signal | 254 | 30.1 | 88.8 |
| metadata Elements without title-page signal | 32 | 3.8 | 11.2 |
| title-page signal but not metadata Elements | 78 | 9.3 |  |
| metadata Elements with no primary subject | 88 | 10.4 | 30.8 |
| metadata Elements with Geometry/Theory primary | 179 | 21.2 | 62.6 |
| metadata Elements with Practical Geometry primary | 18 | 2.1 | 6.3 |

Interpretation:

- The metadata Elements corpus has 321 edition keys, 320 of which are present in the current title-page analysis corpus.
- After reprint/representative grouping, this becomes 286 representative Elements works, or 33.9% of the 843 representative works.
- The metadata Elements corpus is smaller and more historically specific than the broad title-page Euclid/Elements signal.
- Most metadata Elements representatives have a title-page Euclid/Elements signal, but 32 do not under the current feature extraction.
- Conversely, 78 representatives have title-page Euclid/Elements language but are not metadata Elements works. These belong to the wider ecology, not the core Elements corpus.

## Internal Metadata Divisions

### Wardhaugh/Textual Families

Status note:

This section is retained for provenance only. The user considers the Wardhaugh/textual-family labeling unreliable or too coarse for the present argument. Do not use these labels as the organizing taxonomy for the talk. At most, use them lightly as legacy bibliographic hints after checking the case itself.

| wardhaugh_classification | count | pct |
| --- | --- | --- |
| MISSING | 88 | 30.8 |
| CLAVIUS | 61 | 21.3 |
| MAGNIENUS/GRACILIS and DASYPODIUS | 38 | 13.3 |
| ZAMBERTI/GRYNAEUS (THEON) | 20 | 7.0 |
| COMMANDINO | 17 | 5.9 |
| DOU | 13 | 4.5 |
| HÉRIGONE/BARROW | 13 | 4.5 |
| TACQUET | 11 | 3.8 |
| FOIX | 9 | 3.1 |
| D’ÉTAPLES | 7 | 2.4 |
| CAMPANUS | 4 | 1.4 |
| MONTDORÉ | 2 | 0.7 |
| GREGORY | 1 | 0.3 |
| ps-TUSI | 1 | 0.3 |
| MAUROLICUS | 1 | 0.3 |

### Book Coverage Groups

| elements_books_group | count | pct |
| --- | --- | --- |
| books_1_6 | 81 | 28.3 |
| near_complete_or_expanded | 67 | 23.4 |
| unknown_books | 56 | 19.6 |
| books_1_6_plus_solids | 39 | 13.6 |
| selected_later_books | 21 | 7.3 |
| partial_from_book_1 | 8 | 2.8 |
| near_complete_or_expanded_enunciations | 7 | 2.4 |
| books_1_13 | 5 | 1.7 |
| mixed_or_other_books | 1 | 0.3 |
| selected_later_books_enunciations | 1 | 0.3 |

### Periods

| period | count | pct |
| --- | --- | --- |
| 1650-1699 | 83 | 29.0 |
| 1600-1649 | 75 | 26.2 |
| 1550-1599 | 55 | 19.2 |
| 1700+ | 40 | 14.0 |
| pre-1550 | 32 | 11.2 |
| unknown | 1 | 0.3 |

### Languages

| language_first | count | pct |
| --- | --- | --- |
| LATIN | 139 | 48.6 |
| FRENCH | 53 | 18.5 |
| DUTCH | 21 | 7.3 |
| ENGLISH | 20 | 7.0 |
| GREEK | 16 | 5.6 |
| GERMAN | 13 | 4.5 |
| ITALIAN | 12 | 4.2 |
| SPANISH | 6 | 2.1 |
| CHINESE | 2 | 0.7 |
| SWEDISH | 2 | 0.7 |
| PORTUGUESE | 1 | 0.3 |
| ARABIC | 1 | 0.3 |

Interpretation:

The Elements corpus does have some native divisions, especially through textual/editorial families and book coverage. But these divisions are not the whole story. They do not perfectly align with audience, intellectual values, apparatus, language, or title-page rhetoric.

## Elements Corpus Versus Non-Elements Ecology

Strongest positive contrasts for metadata Elements representatives:

| family | metric | elements_count | elements_pct | non_elements_pct | all_pct | delta_vs_non_elements |
| --- | --- | --- | --- | --- | --- | --- |
| tps_feature_presence | references_to_euclid_has | 238 | 83.2 | 9.9 | 34.8 | 73.3 |
| tps_feature_presence | elements_designation_has | 241 | 84.3 | 11.3 | 36.1 | 73.0 |
| rich_claim_mode | claim_ancient_authority_restoration | 238 | 83.2 | 12.0 | 36.2 | 71.2 |
| rich_claim_mode | claim_canonical_textual_identity | 248 | 86.7 | 26.6 | 47.0 | 60.1 |
| archetype | procedural_pedagogical_identity | 163 | 57.0 | 10.6 | 26.3 | 46.4 |
| tps_feature_presence | base_content_has | 258 | 90.2 | 51.9 | 64.9 | 38.3 |
| rich_claim_mode | claim_method_demonstration_order | 120 | 42.0 | 16.3 | 25.0 | 25.6 |
| archetype | humanist_transfer_book | 71 | 24.8 | 1.6 | 9.5 | 23.2 |
| tps_feature_presence | institutions_has | 127 | 44.4 | 21.9 | 29.5 | 22.5 |
| primary_subject_family | Geometry/Theory | 179 | 62.6 | 44.0 | 50.3 | 18.6 |
| tps_feature_presence | edition_details_has | 131 | 45.8 | 27.3 | 33.6 | 18.5 |
| tps_feature_presence | destination_language_has | 72 | 25.2 | 7.4 | 13.4 | 17.8 |
| conservative_intellectual_value | ival_ancient_restoration_humanist | 58 | 20.3 | 2.5 | 8.5 | 17.8 |
| rich_claim_mode | claim_translation_vernacularization_transfer | 75 | 26.2 | 8.4 | 14.5 | 17.8 |
| tps_feature_presence | action_verbs_has | 223 | 78.0 | 60.5 | 66.4 | 17.5 |
| tps_feature_presence | bound_with_has | 87 | 30.4 | 14.4 | 19.8 | 16.1 |
| tps_feature_presence | bound_with_minimal_has | 86 | 30.1 | 14.2 | 19.6 | 15.9 |
| tps_feature_presence | editor_description_has | 193 | 67.5 | 52.4 | 57.5 | 15.1 |
| tps_feature_presence | origin_language_has | 56 | 19.6 | 4.8 | 9.8 | 14.7 |
| conservative_intellectual_value | ival_translation_vernacularization | 48 | 16.8 | 2.7 | 7.5 | 14.1 |
| social_arena | religious_institutional_arena | 60 | 21.0 | 7.4 | 12.0 | 13.6 |
| social_arena | learned_scholarly_arena | 92 | 32.2 | 19.7 | 24.0 | 12.4 |
| tps_feature_presence | description_of_euclid_has | 39 | 13.6 | 1.3 | 5.5 | 12.4 |
| rich_claim_mode | claim_augmentation_enrichment_composition | 117 | 40.9 | 28.7 | 32.9 | 12.2 |

Strongest negative contrasts:

| family | metric | elements_count | elements_pct | non_elements_pct | all_pct | delta_vs_non_elements |
| --- | --- | --- | --- | --- | --- | --- |
| primary_subject_family | Arithmetic/Commerce | 15 | 5.2 | 23.2 | 17.1 | -17.9 |
| social_arena | no_visible_social_arena | 125 | 43.7 | 60.5 | 54.8 | -16.8 |
| primary_subject_family | Visual/Spatial Arts | 9 | 3.1 | 18.1 | 13.0 | -15.0 |
| rich_claim_mode | claim_utility_practice_application | 22 | 7.7 | 17.8 | 14.4 | -10.1 |
| primary_subject_family | Cosmos/Earth | 17 | 5.9 | 11.3 | 9.5 | -5.4 |
| tps_feature_presence | printing_privilege_has | 5 | 1.7 | 5.9 | 4.5 | -4.2 |
| conservative_intellectual_value | ival_utility_application_practice | 12 | 4.2 | 8.3 | 6.9 | -4.1 |
| rich_claim_mode | claim_novelty_modernity_invention | 27 | 9.4 | 13.5 | 12.1 | -4.0 |
| tps_feature_presence | audience_has | 62 | 21.7 | 25.7 | 24.3 | -4.0 |
| archetype | utility_public_book | 13 | 4.5 | 8.4 | 7.1 | -3.9 |
| social_arena | professional_practical_arena | 12 | 4.2 | 8.1 | 6.8 | -3.9 |
| social_arena | court_state_service_arena | 32 | 11.2 | 12.9 | 12.3 | -1.7 |
| social_arena | general_public_lovers_arena | 18 | 6.3 | 6.8 | 6.6 | -0.5 |
| primary_subject_family | Music | 6 | 2.1 | 2.2 | 2.1 | -0.1 |
| social_arena | patronage_prestige_arena | 25 | 8.7 | 8.4 | 8.5 | 0.3 |
| archetype | sparse_canonical_identity | 19 | 6.6 | 6.3 | 6.4 | 0.4 |

Interpretation:

The metadata Elements corpus is clearly distinctive in identity, ancient authority, Geometry/Theory, and procedural/pedagogical framing. But it is not socially or intellectually uniform. It contains sparse canonical books, school books, corrected/augmented books, vernacular practical books, institutional books, and composite apparatus books.

## Internal Family Profiles

These profiles make the answer to the sub-corpus question more nuanced. The metadata Elements corpus has real internal families, but they do not map cleanly onto one social or intellectual axis.

### Wardhaugh/Textual Family Profiles

Status note:

Treat this subsection as exploratory/provenance material, not as an analytical frame. Later phases deliberately replace this with natural modes based on our own evidence: book coverage, language, title-page claims, social arenas, format, and corpus position.

| wardhaugh_classification | n | pedagogical_procedural_pct | composite_workshop_pct | utility_public_pct | humanist_transfer_pct | religious_institution_pct | general_public_pct | no_visible_social_pct | metadata_additional_content_pct |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| MISSING | 88 | 47.7 | 35.2 | 2.3 | 21.6 | 15.9 | 1.1 | 46.6 | 37.5 |
| CLAVIUS | 61 | 59.0 | 26.2 | 11.5 | 0.0 | 50.8 | 8.2 | 29.5 | 1.6 |
| MAGNIENUS/GRACILIS and DASYPODIUS | 38 | 60.5 | 39.5 | 0.0 | 57.9 | 0.0 | 7.9 | 55.3 | 21.1 |
| ZAMBERTI/GRYNAEUS (THEON) | 20 | 40.0 | 20.0 | 5.0 | 45.0 | 0.0 | 0.0 | 60.0 | 5.0 |
| COMMANDINO | 17 | 58.8 | 41.2 | 5.9 | 70.6 | 29.4 | 0.0 | 41.2 | 0.0 |
| DOU | 13 | 100.0 | 76.9 | 0.0 | 15.4 | 0.0 | 69.2 | 23.1 | 0.0 |
| HÉRIGONE/BARROW | 13 | 84.6 | 38.5 | 7.7 | 15.4 | 0.0 | 0.0 | 69.2 | 46.2 |
| TACQUET | 11 | 54.5 | 18.2 | 9.1 | 0.0 | 54.5 | 0.0 | 0.0 | 0.0 |
| FOIX | 9 | 77.8 | 66.7 | 0.0 | 11.1 | 0.0 | 0.0 | 66.7 | 22.2 |
| D’ÉTAPLES | 7 | 85.7 | 57.1 | 0.0 | 57.1 | 14.3 | 0.0 | 42.9 | 42.9 |

Reading:

- `DOU` is the clearest public/procedural/composite family in the current title-page data: very high pedagogical procedure, high general-reader/lovers signal, high composite rhetoric.
- `CLAVIUS` is strongly institutional/religious compared with the corpus average, but not only institutional; it also carries procedural pedagogy and utility in some branches.
- `TACQUET` is learned/school/religious and relatively non-silent socially.
- `COMMANDINO`, `MAGNIENUS/GRACILIS and DASYPODIUS`, and `D’ÉTAPLES` lean toward humanist transfer, translation, and older learned textual authority.
- `HÉRIGONE/BARROW` and `FOIX` look more procedural/composite than sparse, but often without broad visible social address.

### Book-Coverage Profiles

| elements_books_group | n | pedagogical_procedural_pct | composite_workshop_pct | utility_public_pct | humanist_transfer_pct | method_access_pct | metadata_additional_content_pct | no_visible_social_pct |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| books_1_6 | 81 | 54.3 | 35.8 | 0.0 | 21.0 | 9.9 | 0.0 | 45.7 |
| near_complete_or_expanded | 67 | 77.6 | 50.7 | 1.5 | 41.8 | 7.5 | 29.9 | 52.2 |
| unknown_books | 56 | 35.7 | 35.7 | 0.0 | 25.0 | 5.4 | 57.1 | 42.9 |
| books_1_6_plus_solids | 39 | 66.7 | 17.9 | 25.6 | 2.6 | 28.2 | 0.0 | 17.9 |
| selected_later_books | 21 | 38.1 | 23.8 | 4.8 | 23.8 | 0.0 | 9.5 | 47.6 |
| partial_from_book_1 | 8 | 50.0 | 37.5 | 0.0 | 12.5 | 12.5 | 0.0 | 50.0 |
| near_complete_or_expanded_enunciations | 7 | 71.4 | 42.9 | 0.0 | 57.1 | 0.0 | 14.3 | 71.4 |
| books_1_13 | 5 | 60.0 | 20.0 | 0.0 | 0.0 | 0.0 | 0.0 | 60.0 |

Reading:

- Six-book editions are a large elementary zone, but not one rhetoric.
- `1-6 + 11-12` editions are especially important for method/access, utility, and institutional settings; they may be a pedagogical-solid-geometry package rather than simply an abbreviated Elements.
- Near-complete or expanded editions often carry humanist/completeness and apparatus signals.
- Unknown-book entries are not analytically useless: many are rich in additional content and may represent cataloging ambiguity around composite Euclidean material.

## Internal Signals By Wardhaugh/Textual Family

Status note:

Do not build the argument from this table. Use Phase 7 onward for internal Elements modes without external lineage labels.

| wardhaugh_classification | n | metric | count | pct | elements_overall_pct | delta |
| --- | --- | --- | --- | --- | --- | --- |
| TACQUET | 11 | learned_scholarly_arena | 11 | 100.0 | 32.2 | 67.8 |
| DOU | 13 | general_public_lovers_arena | 9 | 69.2 | 6.3 | 62.9 |
| TACQUET | 11 | pedagogical_school_arena | 9 | 81.8 | 23.4 | 58.4 |
| DOU | 13 | claim_method_demonstration_order | 13 | 100.0 | 42.0 | 58.0 |
| HÉRIGONE/BARROW | 13 | claim_method_demonstration_order | 13 | 100.0 | 42.0 | 58.0 |
| DOU | 13 | claim_novelty_modernity_invention | 8 | 61.5 | 9.4 | 52.1 |
| DOU | 13 | claim_selection_extraction_abridgment | 9 | 69.2 | 17.8 | 51.4 |
| FOIX | 9 | claim_selection_extraction_abridgment | 6 | 66.7 | 17.8 | 48.8 |
| COMMANDINO | 17 | humanist_transfer_book | 12 | 70.6 | 24.8 | 45.8 |
| DOU | 13 | claim_visual_material_aids | 8 | 61.5 | 15.7 | 45.8 |
| COMMANDINO | 17 | claim_translation_vernacularization_transfer | 12 | 70.6 | 26.2 | 44.4 |
| DOU | 13 | procedural_pedagogical_identity | 13 | 100.0 | 57.0 | 43.0 |
| DOU | 13 | composite_workshop_book | 10 | 76.9 | 35.7 | 41.3 |
| HÉRIGONE/BARROW | 13 | metadata_additional_data | 6 | 46.2 | 7.0 | 39.2 |
| DOU | 13 | claim_correction_revision_accuracy | 8 | 61.5 | 23.1 | 38.5 |
| HÉRIGONE/BARROW | 13 | claim_completeness_totality_system | 7 | 53.8 | 16.8 | 37.1 |
| FOIX | 9 | claim_augmentation_enrichment_composition | 7 | 77.8 | 40.9 | 36.9 |
| DOU | 13 | claim_augmentation_enrichment_composition | 10 | 76.9 | 40.9 | 36.0 |
| DOU | 13 | professional_practical_arena | 5 | 38.5 | 4.2 | 34.3 |
| TACQUET | 11 | religious_institutional_arena | 6 | 54.5 | 21.0 | 33.6 |
| MAGNIENUS/GRACILIS and DASYPODIUS | 38 | humanist_transfer_book | 22 | 57.9 | 24.8 | 33.1 |
| D’ÉTAPLES | 7 | humanist_transfer_book | 4 | 57.1 | 24.8 | 32.3 |
| MAGNIENUS/GRACILIS and DASYPODIUS | 38 | claim_translation_vernacularization_transfer | 22 | 57.9 | 26.2 | 31.7 |
| FOIX | 9 | composite_workshop_book | 6 | 66.7 | 35.7 | 31.0 |

Interpretation:

Wardhaugh-style textual-family labels may point toward related edition clusters, but they are too unreliable or too coarse to organize this argument. Title-page behavior often cuts across those labels, and the safer analytical frame is to use our own evidence: book coverage, language/place/format, social arenas, apparatus, and intellectual claims.

## Internal Signals By Book Coverage

| elements_books_group | n | metric | count | pct | elements_overall_pct | delta |
| --- | --- | --- | --- | --- | --- | --- |
| near_complete_or_expanded_enunciations | 7 | claim_accessibility_clarity_pedagogy | 4 | 57.1 | 15.7 | 41.4 |
| near_complete_or_expanded_enunciations | 7 | claim_completeness_totality_system | 4 | 57.1 | 16.8 | 40.4 |
| unknown_books | 56 | metadata_has_additional_content | 32 | 57.1 | 19.2 | 37.9 |
| near_complete_or_expanded_enunciations | 7 | humanist_transfer_book | 4 | 57.1 | 24.8 | 32.3 |
| near_complete_or_expanded_enunciations | 7 | claim_translation_vernacularization_transfer | 4 | 57.1 | 26.2 | 30.9 |
| books_1_6_plus_solids | 39 | religious_institutional_arena | 19 | 48.7 | 21.0 | 27.7 |
| near_complete_or_expanded_enunciations | 7 | no_visible_social_arena | 5 | 71.4 | 43.7 | 27.7 |
| partial_from_book_1 | 8 | court_state_service_arena | 3 | 37.5 | 11.2 | 26.3 |
| books_1_6_plus_solids | 39 | claim_method_demonstration_order | 25 | 64.1 | 42.0 | 22.1 |
| books_1_6_plus_solids | 39 | utility_public_book | 10 | 25.6 | 4.5 | 21.1 |
| near_complete_or_expanded | 67 | procedural_pedagogical_identity | 52 | 77.6 | 57.0 | 20.6 |
| books_1_6_plus_solids | 39 | claim_utility_practice_application | 11 | 28.2 | 7.7 | 20.5 |
| books_1_6_plus_solids | 39 | claim_accessibility_clarity_pedagogy | 14 | 35.9 | 15.7 | 20.2 |
| books_1_6_plus_solids | 39 | method_access_book | 11 | 28.2 | 9.8 | 18.4 |
| near_complete_or_expanded | 67 | claim_completeness_totality_system | 23 | 34.3 | 16.8 | 17.5 |
| books_1_13 | 5 | Geometry/Theory | 4 | 80.0 | 62.6 | 17.4 |
| near_complete_or_expanded | 67 | humanist_transfer_book | 28 | 41.8 | 24.8 | 17.0 |
| books_1_6_plus_solids | 39 | military_fortification_arena | 8 | 20.5 | 3.5 | 17.0 |
| near_complete_or_expanded | 67 | metadata_additional_data | 16 | 23.9 | 7.0 | 16.9 |
| books_1_13 | 5 | claim_ancient_authority_restoration | 5 | 100.0 | 83.2 | 16.8 |
| books_1_13 | 5 | no_visible_social_arena | 3 | 60.0 | 43.7 | 16.3 |
| near_complete_or_expanded_enunciations | 7 | claim_augmentation_enrichment_composition | 4 | 57.1 | 40.9 | 16.2 |
| near_complete_or_expanded | 67 | claim_translation_vernacularization_transfer | 28 | 41.8 | 26.2 | 15.6 |
| unknown_books | 56 | Cosmos/Earth | 12 | 21.4 | 5.9 | 15.5 |

Interpretation:

Book coverage creates another flexible boundary. Six-book editions, near-complete/expanded editions, selected later-book editions, and 1-6 + 11-12 configurations can signal different uses of the Elements: elementary pedagogy, solid geometry, scholarly completeness, or extracted utility.

## Language/Period Contrasts Inside The Ecology

| group_type | group | n_elements | n_non_elements | metric | elements_count | elements_pct | non_elements_pct | delta |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| language_first | DUTCH | 21 | 30 | claim_ancient_authority_restoration | 18 | 85.7 | 3.3 | 82.4 |
| language_first | ITALIAN | 12 | 52 | claim_ancient_authority_restoration | 12 | 100.0 | 21.2 | 78.8 |
| period | 1600-1649 | 75 | 248 | claim_ancient_authority_restoration | 64 | 85.3 | 8.9 | 76.5 |
| language_first | ENGLISH | 20 | 47 | claim_ancient_authority_restoration | 17 | 85.0 | 8.5 | 76.5 |
| period | 1700+ | 40 | 15 | claim_canonical_textual_identity | 38 | 95.0 | 20.0 | 75.0 |
| period | 1700+ | 40 | 15 | claim_ancient_authority_restoration | 35 | 87.5 | 13.3 | 74.2 |
| period | 1550-1599 | 55 | 125 | claim_ancient_authority_restoration | 47 | 85.5 | 12.0 | 73.5 |
| language_first | FRENCH | 53 | 209 | claim_ancient_authority_restoration | 42 | 79.2 | 6.7 | 72.5 |
| language_first | ENGLISH | 20 | 47 | claim_canonical_textual_identity | 20 | 100.0 | 27.7 | 72.3 |
| language_first | GERMAN | 13 | 45 | claim_ancient_authority_restoration | 11 | 84.6 | 15.6 | 69.1 |
| language_first | DUTCH | 21 | 30 | claim_canonical_textual_identity | 18 | 85.7 | 16.7 | 69.0 |
| language_first | DUTCH | 21 | 30 | claim_method_demonstration_order | 15 | 71.4 | 3.3 | 68.1 |
| language_first | ENGLISH | 20 | 47 | procedural_pedagogical_identity | 17 | 85.0 | 17.0 | 68.0 |
| language_first | LATIN | 139 | 167 | claim_ancient_authority_restoration | 115 | 82.7 | 16.8 | 66.0 |
| period | 1650-1699 | 83 | 120 | claim_ancient_authority_restoration | 69 | 83.1 | 18.3 | 64.8 |
| language_first | FRENCH | 53 | 209 | claim_canonical_textual_identity | 44 | 83.0 | 18.7 | 64.4 |
| period | 1600-1649 | 75 | 248 | claim_canonical_textual_identity | 64 | 85.3 | 22.2 | 63.2 |
| language_first | FRENCH | 53 | 209 | procedural_pedagogical_identity | 36 | 67.9 | 4.8 | 63.1 |
| period | 1550-1599 | 55 | 125 | claim_canonical_textual_identity | 48 | 87.3 | 24.8 | 62.5 |
| language_first | ITALIAN | 12 | 52 | claim_canonical_textual_identity | 12 | 100.0 | 38.5 | 61.5 |
| language_first | GERMAN | 13 | 45 | claim_canonical_textual_identity | 12 | 92.3 | 31.1 | 61.2 |
| period | 1600-1649 | 75 | 248 | procedural_pedagogical_identity | 52 | 69.3 | 11.3 | 58.0 |
| period | pre-1550 | 32 | 49 | claim_ancient_authority_restoration | 22 | 68.8 | 12.2 | 56.5 |
| period | pre-1550 | 32 | 49 | claim_canonical_textual_identity | 24 | 75.0 | 20.4 | 54.6 |

Interpretation:

Language and time matter, but probably not as clean period boxes. Vernacular and later editions often make usefulness, accessibility, correction, and pedagogical procedure more explicit. Latin and mixed-language editions often preserve scholarly, humanist, or institutional authority, but can also be intensely pedagogical and composite.

## Close-Reading Buckets

| bucket | classification_key | year | city | language | wardhaugh_classification | elements_books | primary_classes | rich_claim_text_raw | rich_social_text_raw |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| sparse_or_low_social_elements | TTORPR | 1528 | Vienna | LATIN |  |  |  | decerp- tum / Continet / ad omnium Mathemati- ces candidatorum / ELEMENTA- LE GEOMETRICVM, EX EV- clidis Geometria / Haylpronnẽsi Colle- ga ciuilis collegij ... | ad omnium Mathemati- ces candidatorum / collegij Viennẽsis / Haylpronnẽsi Colle- ga ciuilis collegij Viennẽsis |
| sparse_or_low_social_elements | FA6MQ0 | 1551 | London | ENGLISH |  |  | Instrument Use / Practical Geometry | CONTAINING / for all sortes of men / THE FIRST PRINciples of Geometrie / as they may moste aptly be applied vnto practise, bothe for vse of instrumentes Geom... | for all sortes of men |
| sparse_or_low_social_elements | Halle_1758 | 1773 | Halle | GERMAN |  | 1-6 | Theoretical Mathematics | übersetzt / aus dem Griechischen übersetzt / zum Gebrauch der Schulen / die sechs ersten Bücher der Elemente / nebst dem eilften und zwölften / Euklids Geome... | zum Gebrauch der Schulen |
| sparse_or_low_social_elements | Venice_1482 | 1482 | Venice | LATIN | CAMPANUS | 1-15 | Theoretical Mathematics | Elementorum Euclidis megarensis / in artem geometrie incipit / Elementorum Euclidis megarensis / Euclidis megarensis / Euclidis |  |
| procedural_pedagogical_elements | Amsterdam_1616 | 1616 | Amsterdam | DUTCH | DOU | 1–6 | Practical Geometry / Theoretical Mathematics | ghevoecht / Overgeset / verclaert / uytgeleyt / oversien / verbetert / vermeerdert / van alle leergierighe, liefhebbers der selver vryer Conste / De ses eers... | van alle leergierighe, liefhebbers der selver vryer Conste / der stadt Leyden Landtmeter ende Wijnroeyer |
| procedural_pedagogical_elements | Rotterdam_1647 | 1647 | Rotterdam | DUTCH | DOU | 1–6 |  | bygevoeght / Overgeset / verklaert / uytgeleyt / oversien / verbetert / vermeerdert / breeder verklaert / bygevoeght / Overgeset / verklaert / uytgeleyt / ov... | alle leergierige liefhebbers der selver vrye Konste / der stadt Leyden Lant-meter ende Wijnroyer |
| procedural_pedagogical_elements | Amsterdam_1626 | 1626 | Amsterdam | DUTCH | DOU | 1–6 |  | ghevoecht / Overgheset / verclaert / uytgheleyt / over-sien / verbetert / vermeerdert / alle leergierighe liefhebbers der selver vryer Conste / De ses eerste... | alle leergierighe liefhebbers der selver vryer Conste / der stadt Leyden Landtmeter ende Wijnroeyer |
| procedural_pedagogical_elements | Rotterdam_1632 | 1632 | Rotterdam | DUTCH | DOU | 1–6 | Practical Geometry | ghevoeght / Overgheset / verk-laert / uytgheleyt / oversien / verbetert / byghevoeghde / verklaert / vermeerdert / van alle leergierighe lief hebbers der sel... | van alle leergierighe lief hebbers der selver vryer Konste / der stadt Leyden Lant-meter ende Wijnroeyer |
| composite_elements | Amsterdam_1616 | 1616 | Amsterdam | DUTCH | DOU | 1–6 | Practical Geometry / Theoretical Mathematics | ghevoecht / Overgeset / verclaert / uytgeleyt / oversien / verbetert / vermeerdert / van alle leergierighe, liefhebbers der selver vryer Conste / De ses eers... | van alle leergierighe, liefhebbers der selver vryer Conste / der stadt Leyden Landtmeter ende Wijnroeyer |
| composite_elements | Rotterdam_1647 | 1647 | Rotterdam | DUTCH | DOU | 1–6 |  | bygevoeght / Overgeset / verklaert / uytgeleyt / oversien / verbetert / vermeerdert / breeder verklaert / bygevoeght / Overgeset / verklaert / uytgeleyt / ov... | alle leergierige liefhebbers der selver vrye Konste / der stadt Leyden Lant-meter ende Wijnroyer |
| composite_elements | Amsterdam_1626 | 1626 | Amsterdam | DUTCH | DOU | 1–6 |  | ghevoecht / Overgheset / verclaert / uytgheleyt / over-sien / verbetert / vermeerdert / alle leergierighe liefhebbers der selver vryer Conste / De ses eerste... | alle leergierighe liefhebbers der selver vryer Conste / der stadt Leyden Landtmeter ende Wijnroeyer |
| composite_elements | Rotterdam_1632 | 1632 | Rotterdam | DUTCH | DOU | 1–6 | Practical Geometry | ghevoeght / Overgheset / verk-laert / uytgheleyt / oversien / verbetert / byghevoeghde / verklaert / vermeerdert / van alle leergierighe lief hebbers der sel... | van alle leergierighe lief hebbers der selver vryer Konste / der stadt Leyden Lant-meter ende Wijnroeyer |
| utility_public_elements | Oxford_London_1700 | 1700 | Oxford, London | ENGLISH | CLAVIUS | 1–6, 11–12 |  | Explain’d / Written in French / Now made English / Errors Corrected / THE ELEMENTS OF EUCLID / Explain’d, In a New, but most Easie Method: Together with The ... | Society of JESUS / that Excellent Mathematician / of the Society of JESUS |
| utility_public_elements | London_1789 | 1789 | London | ENGLISH |  | 1-6, 11-12 | Theoretical Mathematics | CONTAINING / ELEMENTS OF   GEOMETRY / CONTAINING THE PRINCIPAL PROPOSITIONS IN THE FIRST SIX, AND THE ELEVENTH AND TWELFTH BOOKS OF EUCLID / OF THE ROYAL MIL... | THE ROYAL MILITARY ACADEMY, WOOLWICH / OF THE ROYAL MILITARY ACADEMY, WOOLWICH |
| utility_public_elements | London_1680–81 | 1680/1681 | London | ENGLISH | HÉRIGONE/BARROW | 1–6, 11–12 | Arithmetic / Astronomy / Cartography / Cosmography / Geography / Navigation / Practical Geometry / Theoretical Mathematics / Trigonometry | CONTAIN-ING / Composed / demonstrated / the first Six Books of Euclid’s ELEMENTS / as also the Eleventh and Twelfth, symbolically demonstrated / Late Surveyo... | Royal Foundation of the Mathematical SCHOOL / ROYAL SOCIETY / Late Surveyor General of His MAJESTY’S Ordnance, and Fellow of the ROYAL SOCIETY |
| utility_public_elements | Paris_1690 | 1690 | Paris | FRENCH | CLAVIUS | 1–6, 11–12 |  | Reveuë / corrigée / LES ELEMENS D’EUCLIDE / EXPLIQUEZ D’UNE MANIERE nouvelle & tres-facile / AVEC L’USAGE DE CHAQUE Proposition pour toutes les parties des M... | Compagnie de JESUS / de la Compagnie de JESUS |
| institutional_elements | Amsterdam_1616 | 1616 | Amsterdam | DUTCH | DOU | 1–6 | Practical Geometry / Theoretical Mathematics | ghevoecht / Overgeset / verclaert / uytgeleyt / oversien / verbetert / vermeerdert / van alle leergierighe, liefhebbers der selver vryer Conste / De ses eers... | van alle leergierighe, liefhebbers der selver vryer Conste / der stadt Leyden Landtmeter ende Wijnroeyer |
| institutional_elements | Rotterdam_1647 | 1647 | Rotterdam | DUTCH | DOU | 1–6 |  | bygevoeght / Overgeset / verklaert / uytgeleyt / oversien / verbetert / vermeerdert / breeder verklaert / bygevoeght / Overgeset / verklaert / uytgeleyt / ov... | alle leergierige liefhebbers der selver vrye Konste / der stadt Leyden Lant-meter ende Wijnroyer |
| institutional_elements | Amsterdam_1626 | 1626 | Amsterdam | DUTCH | DOU | 1–6 |  | ghevoecht / Overgheset / verclaert / uytgheleyt / over-sien / verbetert / vermeerdert / alle leergierighe liefhebbers der selver vryer Conste / De ses eerste... | alle leergierighe liefhebbers der selver vryer Conste / der stadt Leyden Landtmeter ende Wijnroeyer |
| institutional_elements | Rotterdam_1632 | 1632 | Rotterdam | DUTCH | DOU | 1–6 | Practical Geometry | ghevoeght / Overgheset / verk-laert / uytgheleyt / oversien / verbetert / byghevoeghde / verklaert / vermeerdert / van alle leergierighe lief hebbers der sel... | van alle leergierighe lief hebbers der selver vryer Konste / der stadt Leyden Lant-meter ende Wijnroeyer |
| elements_without_title_signal | Basel_1570 | 1570 | Basel | DUTCH, LATIN |  |  | Arithmetic / Music Theory / Theoretical Mathematics | recognita / emendatiss / prodeunt / ACCESSERVNT / restituit / illustrauit / descripsit / ANITII MANLII SEVERINI BOETHI, PHILOSOPHORVM ET THEOLOGORVM PRINCIPI... | monasterio S. Georgij / Episcopi Pictauiensis / Ioannis Murmelij / Rodolphi Agricolæ / Gilberti Porretæ / HENRICVS LORITVS GLAREANVS / MAR-TIANVS ROTA / Plat... |
| elements_without_title_signal | Würzburg_1661 | 1661 | Würzburg | LATIN | CLAVIUS | 1-6 |  | digesta / disposita / Accesserunt / quivis, vel mediocri præditus ingenio / REGISCURIANI E SOCIETATE JESU Olim in Panormitano Siciliæ, nunc in Herbipolitano ... | quivis, vel mediocri præditus ingenio / SOCIETATE JESU / SOCIETATIS JESU / Gymnasio Matheseos / REGISCURIANI E SOCIETATE JESU Olim in Panormitano Siciliæ, nu... |
| elements_without_title_signal | Frankfurt_1674 | 1674 | Frankfurt | LATIN | CLAVIUS | 1–6 |  | digesta / disposita / Additis / Accesserunt / quivis, vel mediocri præditus ingenio / Libros XXVIII / SIVE ABSOLUTA OMNIUM MATHEMATICARUM DISCIPLINARUM ENCYC... | quivis, vel mediocri præditus ingenio / SOCIETATIS JESU Gymnasio Matheseos / REGISCURIANI E SOCIET. JESU Olim in Panormitano Siciliæ, nunc in Herbipolitano F... |
| elements_without_title_signal | Cologne_1607a | 1607 | Cologne | LATIN |  |  | Theoretical Mathematics | RECOGNITA / AVCTA / distributa / ADDITVS / EDITA / RECOGNITA NOVISSIME AB EODEM, ET AVCTA, & in duos Tomos distributa / TRIPLEX ADDITVS INDEX / PERMISSV AVCT... | SOCIETATIS IESV BIBLIOTHECA SELECTA / SOCIETATIS IESV BIBLIO-THECAE SELECTÆ |

## Preliminary Answer To The User's Question

The metadata Elements corpus does **not** divide into one set of clean, hard-edged sub-corpora. It has several partially overlapping boundary systems:

1. **Book coverage**: first six books, near-complete Elements, 1-6 + 11-12, selected later books, enunciations/excerpts.
2. **Language, place, and format**: Latin learned editions, French/English/Dutch/German vernacular adaptations, Iberian/Portuguese/Spanish Jesuit and royal settings, folio/quarto/octavo/duodecimo formats, etc.
3. **Pedagogical rhetoric**: easy method, brevity, order, demonstrations, use of propositions, school/academy audiences.
4. **Apparatus/composite logic**: Data, Optics/Catoptrics, Archimedes, tables, figures, corollaries, appendices, added uses.
5. **Social authorization**: Jesuit colleges, royal classrooms, professors, surveyors, practitioners, readers/lovers, patrons.
6. **Sparse canonical identity**: editions where author/title/textual authority carry more weight than explicit audience or use, after controlling for title-page fashion.
7. **Legacy textual/editorial labels**: Wardhaugh-style labels can be mentioned only as weak context, not as a main boundary system.

The lines are therefore **flexible rather than clean**. The Elements lives in the broader ecology as a canonical anchor that can become schoolbook, humanist recovery, vernacular practical geometry, institutional textbook, apparatus-rich learned object, or public useful mathematics.

## Questions This Opens

1. Which internal boundary is most historically important for the talk: book coverage, language/place/format, social setting, pedagogical rhetoric, or apparatus?
2. Do vernacular Elements editions behave more like practical geometry and instrument books than like Latin complete Elements editions?
3. Are six-book editions the main zone where Elements becomes pedagogical/practical, while complete editions carry textual/canonical authority?
4. Which non-Elements books look most Elements-like, and what does that say about the spread of elemental pedagogy beyond Euclid?
5. Do title-page density and silence survive controls for time, place, language, format, and school/institutional context?
6. Do authors/editors who publish both Elements and non-Elements works carry the same intellectual ideals across genres, or do they change ideals by language, city, format, and audience?
