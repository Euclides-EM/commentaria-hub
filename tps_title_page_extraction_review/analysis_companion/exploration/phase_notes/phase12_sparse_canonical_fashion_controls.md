# Phase 12 Sparse-Canonical And Title-Page Fashion Controls

This pass tests whether sparse-canonical rhetoric in the metadata-defined Elements corpus is a meaningful intellectual/social signal or partly an artifact of title-page fashion.

The analysis uses three related outcomes:

1. `mode_sparse_canonical`: the Phase 7 natural-mode flag for sparse/canonical Elements.
2. `sparse_canonical_identity`: the broader archetype flag from the rich-title-page analysis.
3. `no_visible_social_arena`: absence of detected social arena.

These are not identical. The first is Elements-specific, the second is a broader title-page archetype, and the third only measures social visibility.

## Main Answer

Sparse-canonical cannot be treated as a pure intellectual signal.

It is strongly associated with city and language, moderately associated with format and period, and strongly affected by explicit school/institution markers. That means some sparse-canonical cases may represent canonical confidence, but some are probably local title-page fashion, bibliographic format, genre convention, or simply a quiet local style.

The safest claim is:

**Sparse-canonical is a real title-page posture, but it needs controls before it can become a historical argument about silence or authority.**

## Strongest Associations

Top associations between controls and sparse/density outcomes:

| field | metric | label | association_type | association | n |
| --- | --- | --- | --- | --- | --- |
| any_school_or_institution | no_visible_social_arena | No visible social arena | cramers_v | 0.685 | 286 |
| city | mode_sparse_canonical | Natural sparse/canonical mode | cramers_v | 0.487 | 286 |
| pedagogical_school_arena | no_visible_social_arena | No visible social arena | cramers_v | 0.487 | 286 |
| any_school_or_institution | social_arena_count | social_arena_count | eta_squared | 0.482 | 286 |
| religious_institutional_arena | no_visible_social_arena | No visible social arena | cramers_v | 0.454 | 286 |
| pedagogical_school_arena | social_arena_count | social_arena_count | eta_squared | 0.446 | 286 |
| city | no_visible_social_arena | No visible social arena | cramers_v | 0.32 | 286 |
| language_first | mode_sparse_canonical | Natural sparse/canonical mode | cramers_v | 0.306 | 286 |
| any_school_or_institution | density_score | density_score | eta_squared | 0.278 | 286 |
| city | sparse_canonical_identity | Sparse-canonical archetype | cramers_v | 0.276 | 286 |
| any_school_or_institution | mode_sparse_canonical | Natural sparse/canonical mode | cramers_v | 0.273 | 286 |
| language_first | no_visible_social_arena | No visible social arena | cramers_v | 0.248 | 286 |
| elements_books_group | no_visible_social_arena | No visible social arena | cramers_v | 0.234 | 286 |
| period | sparse_canonical_identity | Sparse-canonical archetype | cramers_v | 0.215 | 286 |
| elements_books_group | mode_sparse_canonical | Natural sparse/canonical mode | cramers_v | 0.208 | 286 |
| religious_institutional_arena | mode_sparse_canonical | Natural sparse/canonical mode | cramers_v | 0.196 | 286 |
| primary_subject_family | no_visible_social_arena | No visible social arena | cramers_v | 0.19 | 286 |
| pedagogical_school_arena | mode_sparse_canonical | Natural sparse/canonical mode | cramers_v | 0.19 | 286 |
| format_group | sparse_canonical_identity | Sparse-canonical archetype | cramers_v | 0.187 | 286 |
| format_group | no_visible_social_arena | No visible social arena | cramers_v | 0.181 | 286 |
| format_group | mode_sparse_canonical | Natural sparse/canonical mode | cramers_v | 0.179 | 286 |
| period | mode_sparse_canonical | Natural sparse/canonical mode | cramers_v | 0.165 | 286 |
| pedagogical_school_arena | density_score | density_score | eta_squared | 0.162 | 286 |
| language_first | sparse_canonical_identity | Sparse-canonical archetype | cramers_v | 0.148 | 286 |
| period | no_visible_social_arena | No visible social arena | cramers_v | 0.137 | 286 |

Reading:

- City is the strongest control for the Elements sparse/canonical natural mode: Cramer's V = 0.487.
- Language is also meaningful for sparse/canonical: Cramer's V = 0.306.
- Format matters, but less strongly: Cramer's V = 0.179 for natural sparse/canonical and 0.187 for the broader sparse-canonical archetype.
- Period matters, but less than city/language.
- School/institution markers strongly reduce social invisibility because they introduce social evidence onto the title page.

## Sparse-Canonical By Control Field

Highest groups for the Elements sparse/canonical natural mode:

| field | value | n | count | pct | overall_pct | delta_vs_overall | avg_density_score |
| --- | --- | --- | --- | --- | --- | --- | --- |
| city | Bologna | 6 | 5 | 83.3 | 15.4 | 67.9 | 5.13 |
| city | Leipzig | 5 | 3 | 60.0 | 15.4 | 44.6 | 4.92 |
| language_first | ITALIAN | 12 | 7 | 58.3 | 15.4 | 42.9 | 6.0 |
| format_group | 6 | 5 | 2 | 40.0 | 15.4 | 24.6 | 4.92 |
| elements_books_group | books_1_13 | 5 | 2 | 40.0 | 15.4 | 24.6 | 7.04 |
| city | Lyon | 10 | 4 | 40.0 | 15.4 | 24.6 | 7.24 |
| format_group | missing/unknown | 18 | 6 | 33.3 | 15.4 | 17.9 | 5.62 |
| language_first | GREEK | 16 | 5 | 31.2 | 15.4 | 15.9 | 5.98 |
| city | Strasbourg | 13 | 4 | 30.8 | 15.4 | 15.4 | 5.69 |
| period | pre-1550 | 32 | 8 | 25.0 | 15.4 | 9.6 | 5.57 |
| any_school_or_institution | 0 | 170 | 40 | 23.5 | 15.4 | 8.1 | 5.85 |
| elements_books_group | books_1_6 | 81 | 19 | 23.5 | 15.4 | 8.1 | 7.18 |
| city | Venice | 13 | 3 | 23.1 | 15.4 | 7.7 | 5.31 |
| city | Rome | 14 | 3 | 21.4 | 15.4 | 6.0 | 7.11 |
| period | 1700+ | 40 | 8 | 20.0 | 15.4 | 4.6 | 7.87 |
| pedagogical_school_arena | 0 | 219 | 42 | 19.2 | 15.4 | 3.8 | 6.5 |
| religious_institutional_arena | 0 | 226 | 43 | 19.0 | 15.4 | 3.6 | 6.71 |
| period | 1550-1599 | 55 | 10 | 18.2 | 15.4 | 2.8 | 7.02 |
| elements_books_group | unknown_books | 56 | 10 | 17.9 | 15.4 | 2.5 | 6.11 |
| language_first | SPANISH | 6 | 1 | 16.7 | 15.4 | 1.3 | 7.8 |
| format_group | duodecimo | 30 | 5 | 16.7 | 15.4 | 1.3 | 7.61 |
| primary_subject_family | Geometry/Theory | 157 | 25 | 15.9 | 15.4 | 0.5 | 7.5 |
| language_first | GERMAN | 13 | 2 | 15.4 | 15.4 | 0.0 | 6.05 |
| format_group | octavo | 111 | 17 | 15.3 | 15.4 | -0.1 | 7.69 |
| primary_subject_family | No primary family | 88 | 13 | 14.8 | 15.4 | -0.6 | 6.98 |

Highest groups for the broader sparse-canonical archetype:

| field | value | n | count | pct | overall_pct | delta_vs_overall | avg_density_score |
| --- | --- | --- | --- | --- | --- | --- | --- |
| format_group | missing/unknown | 18 | 4 | 22.2 | 6.6 | 15.6 | 5.62 |
| format_group | 6 | 5 | 1 | 20.0 | 6.6 | 13.4 | 4.92 |
| city | Lyon | 10 | 2 | 20.0 | 6.6 | 13.4 | 7.24 |
| city | Leipzig | 5 | 1 | 20.0 | 6.6 | 13.4 | 4.92 |
| elements_books_group | books_1_13 | 5 | 1 | 20.0 | 6.6 | 13.4 | 7.04 |
| language_first | ITALIAN | 12 | 2 | 16.7 | 6.6 | 10.0 | 6.0 |
| city | Bologna | 6 | 1 | 16.7 | 6.6 | 10.0 | 5.13 |
| period | pre-1550 | 32 | 5 | 15.6 | 6.6 | 9.0 | 5.57 |
| city | Venice | 13 | 2 | 15.4 | 6.6 | 8.7 | 5.31 |
| period | 1700+ | 40 | 6 | 15.0 | 6.6 | 8.4 | 7.87 |
| primary_subject_family | Geometry/Theory \| Arithmetic/Commerce | 7 | 1 | 14.3 | 6.6 | 7.6 | 6.43 |
| elements_books_group | unknown_books | 56 | 6 | 10.7 | 6.6 | 4.1 | 6.11 |
| any_school_or_institution | 0 | 170 | 16 | 9.4 | 6.6 | 2.8 | 5.85 |
| religious_institutional_arena | 0 | 226 | 18 | 8.0 | 6.6 | 1.3 | 6.71 |
| pedagogical_school_arena | 0 | 219 | 17 | 7.8 | 6.6 | 1.1 | 6.5 |

Highest groups for no visible social arena:

| field | value | n | count | pct | overall_pct | delta_vs_overall | avg_density_score |
| --- | --- | --- | --- | --- | --- | --- | --- |
| city | Bologna | 6 | 6 | 100.0 | 43.7 | 56.3 | 5.13 |
| primary_subject_family | Geometry/Theory \| Cosmos/Earth | 6 | 5 | 83.3 | 43.7 | 39.6 | 5.07 |
| format_group | 6 | 5 | 4 | 80.0 | 43.7 | 36.3 | 4.92 |
| city | Cologne | 9 | 7 | 77.8 | 43.7 | 34.1 | 7.16 |
| language_first | ITALIAN | 12 | 9 | 75.0 | 43.7 | 31.3 | 6.0 |
| any_school_or_institution | 0 | 170 | 122 | 71.8 | 43.7 | 28.1 | 5.85 |
| elements_books_group | near_complete_or_expanded_enunciations | 7 | 5 | 71.4 | 43.7 | 27.7 | 6.97 |
| format_group | missing/unknown | 18 | 12 | 66.7 | 43.7 | 23.0 | 5.62 |
| language_first | GREEK | 16 | 10 | 62.5 | 43.7 | 18.8 | 5.98 |
| city | Basel | 8 | 5 | 62.5 | 43.7 | 18.8 | 6.98 |
| city | Strasbourg | 13 | 8 | 61.5 | 43.7 | 17.8 | 5.69 |
| elements_books_group | books_1_13 | 5 | 3 | 60.0 | 43.7 | 16.3 | 7.04 |
| city | Leipzig | 5 | 3 | 60.0 | 43.7 | 16.3 | 4.92 |
| period | pre-1550 | 32 | 19 | 59.4 | 43.7 | 15.7 | 5.57 |
| pedagogical_school_arena | 0 | 219 | 125 | 57.1 | 43.7 | 13.4 | 6.5 |
| religious_institutional_arena | 0 | 226 | 125 | 55.3 | 43.7 | 11.6 | 6.71 |
| language_first | GERMAN | 13 | 7 | 53.8 | 43.7 | 10.1 | 6.05 |
| city | Venice | 13 | 7 | 53.8 | 43.7 | 10.1 | 5.31 |
| elements_books_group | near_complete_or_expanded | 67 | 35 | 52.2 | 43.7 | 8.5 | 7.79 |
| elements_books_group | partial_from_book_1 | 8 | 4 | 50.0 | 43.7 | 6.3 | 6.22 |

Interpretation:

- Bologna, Leipzig, and Italian-language cases are strong fashion-risk zones for sparse/canonical behavior.
- Pre-1550 cases are somewhat more sparse/canonical, but period is not the strongest explanation.
- Books `1-6` are more sparse/canonical than `1-6 + 11-12`, which fits the earlier result that `1-6 + 11-12` is a louder practical-pedagogical package.
- Missing/unknown format is high in sparse-canonical, which means some format-linked conclusions need caution.
- Duodecimo and octavo are not automatically sparse. In fact, duodecimos often have high institutional/pedagogical/practical signals in Phase 11.

## Combined Strata

High sparse rates inside combined time/language/format/city/bookgroup strata:

| stratum_type | stratum | n | avg_density_score | mode_sparse_canonical_count | mode_sparse_canonical_pct | sparse_canonical_identity_count | sparse_canonical_identity_pct | no_visible_social_arena_count | no_visible_social_arena_pct |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| period_format | period=pre-1550 \| format_group=octavo | 5 | 3.96 | 5 | 100.0 | 2 | 40.0 | 5 | 100.0 |
| period_language_format | period=pre-1550 \| language_first=LATIN \| format_group=octavo | 4 | 3.95 | 4 | 100.0 | 2 | 50.0 | 4 | 100.0 |
| period_language_bookgroup | period=1650-1699 \| language_first=ITALIAN \| elements_books_group=books_1_6 | 4 | 4.6 | 4 | 100.0 | 0 | 0.0 | 4 | 100.0 |
| language_format | language_first=ITALIAN \| format_group=octavo | 3 | 4.93 | 3 | 100.0 | 0 | 0.0 | 3 | 100.0 |
| period_language_format | period=1650-1699 \| language_first=ITALIAN \| format_group=octavo | 3 | 4.93 | 3 | 100.0 | 0 | 0.0 | 3 | 100.0 |
| period_language | period=1650-1699 \| language_first=ITALIAN | 6 | 5.03 | 5 | 83.3 | 0 | 0.0 | 6 | 100.0 |
| city_format | city=Bologna \| format_group=octavo | 4 | 5.55 | 3 | 75.0 | 0 | 0.0 | 4 | 100.0 |
| city_format | city=Naples \| format_group=octavo | 3 | 5.0 | 2 | 66.7 | 2 | 66.7 | 3 | 100.0 |
| city_format | city=Strasbourg \| format_group=quarto | 3 | 4.0 | 2 | 66.7 | 0 | 0.0 | 3 | 100.0 |
| period_language_format | period=1550-1599 \| language_first=GREEK \| format_group=quarto | 3 | 6.8 | 2 | 66.7 | 0 | 0.0 | 2 | 66.7 |
| bookgroup_format | elements_books_group=books_1_6 \| format_group=missing/unknown | 8 | 3.9 | 4 | 50.0 | 3 | 37.5 | 7 | 87.5 |
| city_format | city=Leipzig \| format_group=octavo | 4 | 5.05 | 2 | 50.0 | 1 | 25.0 | 2 | 50.0 |
| city_format | city=Rome \| format_group=octavo | 4 | 7.35 | 2 | 50.0 | 0 | 0.0 | 2 | 50.0 |
| period_language_bookgroup | period=1700+ \| language_first=LATIN \| elements_books_group=books_1_6 | 4 | 7.65 | 2 | 50.0 | 1 | 25.0 | 2 | 50.0 |
| period_format | period=1700+ \| format_group=missing/unknown | 12 | 6.2 | 5 | 41.7 | 3 | 25.0 | 7 | 58.3 |
| language_format | language_first=GREEK \| format_group=quarto | 5 | 6.2 | 2 | 40.0 | 0 | 0.0 | 3 | 60.0 |
| bookgroup_format | elements_books_group=unknown_books \| format_group=octavo | 11 | 5.27 | 4 | 36.4 | 2 | 18.2 | 6 | 54.5 |
| period_language | period=1550-1599 \| language_first=GREEK | 9 | 6.67 | 3 | 33.3 | 0 | 0.0 | 4 | 44.4 |
| period_language_bookgroup | period=pre-1550 \| language_first=LATIN \| elements_books_group=near_complete_or_expanded | 9 | 5.73 | 3 | 33.3 | 2 | 22.2 | 7 | 77.8 |
| period_language_format | period=1700+ \| language_first=LATIN \| format_group=octavo | 6 | 6.97 | 2 | 33.3 | 2 | 33.3 | 3 | 50.0 |
| period_language | period=1700+ \| language_first=GERMAN | 3 | 3.73 | 1 | 33.3 | 1 | 33.3 | 2 | 66.7 |
| period_language | period=1700+ \| language_first=ITALIAN | 3 | 6.93 | 1 | 33.3 | 1 | 33.3 | 2 | 66.7 |
| period_language | period=pre-1550 \| language_first=GREEK | 3 | 2.73 | 1 | 33.3 | 0 | 0.0 | 3 | 100.0 |
| language_format | language_first=ENGLISH \| format_group=folio | 3 | 7.8 | 1 | 33.3 | 0 | 0.0 | 2 | 66.7 |
| language_format | language_first=ITALIAN \| format_group=missing/unknown | 3 | 6.67 | 1 | 33.3 | 1 | 33.3 | 2 | 66.7 |

Reading:

Several high-sparse strata are small, but they are still useful warning signs:

- pre-1550 Latin octavos are very sparse in this corpus;
- Italian `books_1_6` in 1650-1699 are extremely sparse/socially quiet;
- Bologna octavos are high sparse/socially quiet;
- some Greek/quarto and Strasbourg/quarto strata also show high sparse behavior.

This supports the user's caution: sparse-canonical may be a local/period/language/format style in some zones, not only an intellectual posture.

## School And Institution Markers

| marker | value | n | mode_sparse_canonical_pct | mode_sparse_canonical_count | sparse_canonical_identity_pct | sparse_canonical_identity_count | no_visible_social_arena_pct | no_visible_social_arena_count | avg_rich_claim_count | avg_social_arena_count | avg_tps_feature_has_count | avg_density_score |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| any_school_or_institution | 0 | 170 | 23.5 | 40 | 9.4 | 16 | 71.8 | 122 | 3.39 | 0.42 | 10.21 | 5.85 |
| any_school_or_institution | 1 | 116 | 3.4 | 4 | 2.6 | 3 | 2.6 | 3 | 4.53 | 2.11 | 12.45 | 9.14 |
| pedagogical_school_arena | 0 | 219 | 19.2 | 42 | 7.8 | 17 | 57.1 | 125 | 3.71 | 0.66 | 10.65 | 6.5 |
| pedagogical_school_arena | 1 | 67 | 3.0 | 2 | 3.0 | 2 | 0.0 | 0 | 4.33 | 2.55 | 12.64 | 9.41 |
| religious_institutional_arena | 0 | 226 | 19.0 | 43 | 8.0 | 18 | 55.3 | 125 | 3.67 | 0.88 | 10.79 | 6.71 |
| religious_institutional_arena | 1 | 60 | 1.7 | 1 | 1.7 | 1 | 0.0 | 0 | 4.55 | 1.95 | 12.37 | 8.97 |
| aud_students_learners | 0 | 276 | 15.9 | 44 | 6.9 | 19 | 44.6 | 123 | 3.82 | 1.08 | 11.08 | 7.12 |
| aud_students_learners | 1 | 10 | 0.0 | 0 | 0.0 | 0 | 20.0 | 2 | 4.7 | 1.7 | 12.3 | 8.86 |

Reading:

In the current feature set, explicit school/institution markers usually make title pages less sparse, not more sparse.

- Cases with any school/institution marker: 3.4% natural sparse/canonical and 2.6% no visible social arena.
- Cases without school/institution marker: 23.5% natural sparse/canonical and 71.8% no visible social arena.
- Pedagogical-school arena cases have 0.0% no visible social arena because the arena itself is social evidence.

This does not prove that schoolbooks are never materially sparse. It means that when the title page explicitly signals school or institution, it becomes socially legible in our data.

## Case-Level Risk Sorting

The case-level file is exploratory:

`derived_data/sparse_canonical_cases_with_fashion_controls.csv`

It estimates a rough fashion-propensity score from marginal rates by period, language, city, format, book coverage, and primary subject family. This is not a model of causation; it is a triage tool for close reading.

Cases flagged as higher fashion risk:

| classification_key | year | city | language_first | period | format_group | elements_books_group | fashion_propensity_sparse_pct | period_language_format_n | period_language_format_sparse_pct | city_format_n | city_format_sparse_pct | density_score | interpretive_risk |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 8DP4C1 | 1613.0 | Bologna | ITALIAN | 1600-1649 | missing/unknown | books_1_6 | 36.8 | 1 |  | 2 |  | 2.8 | high_fashion_risk |
| Bologna_1651a | 1651.0 | Bologna | ITALIAN | 1650-1699 | octavo | books_1_6 | 34.9 | 3 | 100.0 | 4 | 75.0 | 4.8 | high_fashion_risk |
| Bologna_1651 | 1651.0 | Bologna | ITALIAN | 1650-1699 | octavo | books_1_6 | 34.9 | 3 | 100.0 | 4 | 75.0 | 5.0 | high_fashion_risk |
| Bologna_1686 | 1686.0 | Bologna | ITALIAN | 1650-1699 | octavo | books_1_6 | 34.9 | 3 | 100.0 | 4 | 75.0 | 5.0 | high_fashion_risk |
| Leipzig_1883 | 1883.0 | Leipzig | GREEK | 1700+ | missing/unknown | books_1_13 | 33.4 | 1 |  | 1 |  | 4.4 | high_fashion_risk |
| 2L8L55 | 1719.0 | Bologna | LATIN | 1700+ | missing/unknown | books_1_6 | 31.6 | 1 |  | 2 |  | 5.8 | high_fashion_risk |
| Milan_1702 | 1702.0 | Milan | ITALIAN | 1700+ | 6 | books_1_6 | 31.3 | 1 |  | 1 |  | 3.6 | high_fashion_risk |
| Milan_1671 | 1671.0 | Milan | ITALIAN | 1650-1699 | duodecimo | books_1_6 | 25.8 | 2 |  | 1 |  | 3.6 | high_fashion_risk |
| Leipzig_1549 | 1549.0 | Leipzig | LATIN | pre-1550 | octavo | books_1_6 | 25.4 | 4 | 100.0 | 4 | 50.0 | 5.0 | high_fashion_risk |

Possible stronger candidates for meaningful sparse/canonical rhetoric, because they are sparse in lower-propensity contexts or unusually quiet:

| classification_key | year | city | language_first | period | format_group | elements_books_group | fashion_propensity_sparse_pct | period_language_format_n | period_language_format_sparse_pct | city_format_n | city_format_sparse_pct | density_score | interpretive_risk |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Paris_1545 | 1545.0 | Paris | LATIN | pre-1550 | octavo | near_complete_or_expanded | 13.9 | 4 | 100.0 | 27 | 14.8 | 2.8 | stronger_candidate_meaningful_silence |
| FA6MQ0 | 1551.0 | London | ENGLISH | 1550-1599 | quarto | unknown_books | 13.6 | 1 |  | 3 | 33.3 | 1.8 | stronger_candidate_meaningful_silence |
| Paris_1654a | 1654.0 | Paris | FRENCH | 1650-1699 | octavo | books_1_6 | 13.2 | 9 | 11.1 | 27 | 14.8 | 4.4 | stronger_candidate_meaningful_silence |
| Arnhem_1605 | 1605.0 | Arnhem | LATIN | 1600-1649 | quarto | selected_later_books | 11.6 | 7 | 28.6 | 3 | 33.3 | 4.0 | stronger_candidate_meaningful_silence |

Medium/ambiguous cases, where close reading is needed:

| classification_key | year | city | language_first | period | format_group | elements_books_group | fashion_propensity_sparse_pct | period_language_format_n | period_language_format_sparse_pct | city_format_n | city_format_sparse_pct | density_score | interpretive_risk |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Uppsala_1744 | 1744.0 | Uppsala | SWEDISH | 1700+ | missing/unknown | books_1_6 | 23.2 | 2 |  | 2 |  | 6.6 | medium |
| Florence_1690 | 1690.0 | Florence | ITALIAN | 1650-1699 | duodecimo | books_1_6_plus_solids | 23.1 | 2 |  | 1 |  | 5.6 | medium |
| Strasbourg_1566 | 1566.0 | Strasbourg | GREEK | 1550-1599 | quarto | books_1_6 | 21.7 | 3 | 66.7 | 3 | 66.7 | 5.4 | medium |
| Halle_1758 | 1773.0 | Halle | GERMAN | 1700+ | missing/unknown | books_1_6 | 21.6 | 2 |  | 1 |  | 3.6 | medium |
| Leipzig_1607 | 1607.0 | Leipzig | LATIN | 1600-1649 | octavo | unknown_books | 21.6 | 13 | 7.7 | 4 | 50.0 | 5.8 | medium |
| Leiden_and_Amsterdam_1673 | 1673.0 | Leiden, Amsterdam | LATIN | 1650-1699 | 6 | books_1_6 | 21.5 | 3 | 33.3 | 1 |  | 5.2 | medium |
| Rome_1545 | 1545.0 | Rome | GREEK | pre-1550 | octavo | unknown_books | 21.1 | 1 |  | 4 | 50.0 | 4.0 | medium |
| Strasbourg_1557 | 1557.0 | Strasbourg | GREEK | 1550-1599 | quarto | unknown_books | 21.0 | 3 | 66.7 | 3 | 66.7 | 5.2 | medium |
| Lund_1855 | 1855.0 | Lund | SWEDISH | 1700+ | missing/unknown | selected_later_books | 20.9 | 2 |  | 1 |  | 2.8 | medium |
| Strasbourg_1570a | 1570.0 | Strasbourg | GREEK | 1550-1599 | octavo | near_complete_or_expanded | 20.9 | 6 | 16.7 | 10 | 20.0 | 4.8 | medium |
| Rome_1594 | 1594.0 | Rome | ARABIC | 1550-1599 | folio | books_1_13 | 20.7 | 1 |  | 3 | 33.3 | 4.2 | medium |
| Lyon_1669 | 1669.0 | Lyon | LATIN | 1650-1699 | duodecimo | unknown_books | 19.8 | 7 | 14.3 | 2 |  | 3.6 | medium |
| Lyon_1660 | 1660.0 | Lyon | LATIN | 1650-1699 | 18 | books_1_6_plus_solids | 18.9 | 1 |  | 1 |  | 4.0 | medium |
| Lyon_1603 | 1603.0 | Lyon | LATIN | 1600-1649 | quarto | books_1_6 | 18.4 | 7 | 28.6 | 4 | 25.0 | 5.0 | medium |
| Strasbourg_1571 | 1571.0 | Strasbourg | LATIN | 1550-1599 | octavo | selected_later_books | 18.0 | 14 | 7.1 | 10 | 20.0 | 5.2 | medium |
| Naples_1702 | 1702.0 | Naples | LATIN | 1700+ | octavo | books_1_6 | 17.7 | 6 | 33.3 | 3 | 66.7 | 3.2 | medium |
| Venice_1498 | 1498.0 | Venice | LATIN | pre-1550 | folio | unknown_books | 17.5 | 20 | 10.0 | 9 | 22.2 | 3.8 | medium |
| Lyon_1672 | 1672.0 | Lyon | FRENCH | 1650-1699 | duodecimo | books_1_6_plus_solids | 17.2 | 8 | 12.5 | 2 |  | 4.0 | medium |

Quiet but not sparse-mode comparison cases:

| classification_key | year | city | language_first | period | format_group | elements_books_group | fashion_propensity_sparse_pct | density_score | rich_claim_text_raw | rich_social_text_raw |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 4ZSED7 | 1589 | Leipzig | LATIN | 1550-1599 | octavo | unknown_books | 23.3 | 0.0 |  |  |
| BQ96WF | 1758 | Göttingen | GERMAN | 1700+ | missing/unknown | unknown_books | 21.6 | 0.6 |  | Königl. Gesellschaft der Wissenschaften |
| Strasbourg_1538 | 1538 | Strasbourg | GREEK | pre-1550 | quarto | selected_later_books | 21.3 | 1.4 | ΕΥΚΛΕΙΔΟΥ ΤΩΝ ΠΕΝΤΕ ΚΑΙ ΔΕΚΑ ΣΤΟΙΧΕΙΩΝ \| Θέωνος \| ΕΥΚΛΕΙΔΟΥ ΤΩΝ ΠΕΝΤΕ ΚΑΙ ΔΕΚΑ ΣΤΟΙΧΕΙΩΝ \| ΕΥΚΛΕΙΔΟΥ | Θέωνος |
| Strasbourg_1559 | 1559 | Strasbourg | GREEK | 1550-1599 | octavo | selected_later_books | 20.8 | 1.8 | συνουσιῶν \| ΕΥΚΛΕΙΔΟΥ ΤΩΝ ΠΕΝΤΕ ΚΑΙ ΔΕΚΑ ΣΤΟΙΧΕΙΩΝ \| τὸ πρῶτον \| Θέωνος \| ΕΥΚΛΕΙΔΟΥ ΤΩΝ ΠΕΝΤΕ ΚΑΙ ΔΕΚΑ ΣΤΟΙΧΕΙΩΝ \| ΕΥΚΛΕΙΔΟΥ | Θέωνος |
| Beijing_1629 | 1629 | Beijing | CHINESE | 1600-1649 | missing/unknown | books_1_6 | 19.8 | 0.0 |  |  |
| Beijing_1607 | 1607 | Beijing | CHINESE | 1600-1649 | missing/unknown | books_1_6 | 19.6 | 0.4 | 幾何原本 \| 幾何原本 |  |
| Venice_1491 | 1491 | Venice | LATIN | pre-1550 | folio | unknown_books | 17.3 | 0.0 |  |  |
| Basel_1533 | 1533 | Basel | GREEK | pre-1550 | folio | near_complete_or_expanded | 17.0 | 2.8 | Adiecta \| ΕΥΚΛΕΙΔΟΥ ΣΤΟΙΧΕΙΩΝ ΒΙΒΛ. ΙΕ \| Πρόκλου \| ΕΚ ΤΩΝ ΘΕΩΝΟΣ ΣΥΝΟΥΣΙΩΝ \| ἐξηγημάτων Πρόκλου βιβ. δ΄ \| ΘΕΩΝΟΣ ΣΥΝΟΥΣΙΩΝ \| Πρόκλου βιβ. δ΄ \| ΕΥΚΛΕΙΔΟΥ | Πρόκλου |
| s.l_1650 | 1650 | missing/unknown | LATIN | 1650-1699 | octavo | books_1_6 | 16.6 | 0.0 |  |  |
| Venice_1647 | 1647 | Venice | LATIN | 1600-1649 | octavo | books_1_6 | 16.4 | 0.6 |  |  |
| Venice_1575 | 1575 | Venice | LATIN | 1550-1599 | quarto | selected_later_books | 16.2 | 2.8 | ædita \| Opuscula Mathematica \| cum rerum omnium notatu dignarum \| Nunc primùm in lucem ædita \| INDICE LOCVPLETISSIMO | D. FRANCISCI MAVROLYCI, ABBATIS MESSANENSIS |
| The_Hague_1690 | 1690 | The Hague | FRENCH | 1650-1699 | duodecimo | books_1_6 | 15.0 | 1.2 | Suivant \| Suivant la Copie de Paris |  |

Interpretation:

- The high-risk cases cluster around Italian/Bologna/Milan/Leipzig and related sparse-prone contexts. Treat these as likely fashion-contaminated until close-read.
- The stronger-candidate list is small and should be handled carefully. It includes cases that are sparse relative to marginal controls, but some still have social phrases that did not trigger social-arena categories. They are candidates, not proof.
- Quiet non-sparse cases show why no-visible-social alone is too blunt: many title pages can be socially quiet without belonging to the sparse-canonical mode.

## Revised Interpretation Of Sparse-Canonical

Sparse-canonical should not be a major thesis by itself yet.

Better use:

- as a **contrast mode** against procedural/pedagogical, practical/public, and composite/apparatus Elements;
- as a **problem of title-page fashion** that lets us show methodological care;
- as a **close-reading category** only after comparing nearby books by period, language, city, format, and book coverage.

Bad use:

- as immediate evidence that a book had no audience;
- as immediate evidence that canonical authority alone was sufficient;
- as a general claim about French, Latin, schoolbooks, or early/late periods without controls.

## Provisional Claim

Some Elements title pages rely on canonical identity with little explicit social or practical explanation. But in the current corpus, sparse-canonical behavior is entangled with city, language, format, period, and institutional visibility. The safest historical claim is not "sparse means canonical confidence," but:

**Canonical silence is one possible strategy inside the Elements corpus, and we can only identify it after subtracting local title-page fashion.**

## Next Questions

1. Close-read the stronger-candidate sparse cases and high-fashion-risk cases side by side.
2. Run the author/editor portfolio analysis, because it can show whether sparse/dense rhetoric changes within the same author across language, city, format, and genre.
3. Deepen the Dutch plain `1-6` route, since it is practical/public without being the `1-6 + 11-12` package.
