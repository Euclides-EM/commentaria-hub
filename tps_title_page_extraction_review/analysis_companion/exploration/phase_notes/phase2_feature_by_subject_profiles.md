# Phase 2 Feature-By-Subject Profiles

Representative-level analysis. Reprints are aggregated to classification representatives so repeated editions do not dominate feature prevalence.

## Method

- Aggregate all non-empty TPS feature values for title-page editions mapped to the same representative classification key.
- For each representative, mark whether each TPS feature is present.
- Compare feature prevalence by subject family.

## Feature Family Prevalence By Subject Family

| subject_family | addition | authority | community | identity | imprint | transformation |
| --- | --- | --- | --- | --- | --- | --- |
| Applied Mechanics/Military | 56.1 | 84.2 | 64.9 | 35.1 | 87.7 | 70.2 |
| Arithmetic/Commerce | 54.9 | 86.1 | 46.5 | 62.5 | 89.6 | 70.8 |
| Cosmos/Earth | 60.0 | 91.2 | 53.8 | 51.2 | 85.0 | 67.5 |
| Geometry/Theory | 55.0 | 90.6 | 60.6 | 75.0 | 90.1 | 70.8 |
| Instruments/Measurement | 56.7 | 90.4 | 61.5 | 49.0 | 93.3 | 69.2 |
| Music | 38.9 | 94.4 | 50.0 | 44.4 | 77.8 | 66.7 |
| Visual/Spatial Arts | 40.0 | 84.5 | 60.0 | 39.1 | 82.7 | 55.5 |

## Strongest Positive Contrasts By Subject Family

### Arithmetic/Commerce

| feature | n_in | pct_in | pct_out | delta |
| --- | --- | --- | --- | --- |
| edition_details | 144 | 45.8 | 31.0 | 14.8 |
| audience | 144 | 30.6 | 23.0 | 7.5 |
| has_addition | 144 | 54.9 | 48.5 | 6.4 |
| action_verbs | 144 | 70.8 | 65.5 | 5.3 |
| editor_name | 144 | 79.9 | 75.8 | 4.0 |
| has_transformation | 144 | 70.8 | 68.0 | 2.9 |
| enriched_with | 144 | 43.8 | 41.1 | 2.7 |
| bound_with | 144 | 21.5 | 19.5 | 2.1 |

### Geometry/Theory

| feature | n_in | pct_in | pct_out | delta |
| --- | --- | --- | --- | --- |
| elements_designation | 424 | 46.5 | 25.5 | 20.9 |
| base_content | 424 | 74.5 | 55.1 | 19.4 |
| references_to_euclid | 424 | 44.3 | 25.1 | 19.3 |
| has_identity | 424 | 75.0 | 55.8 | 19.2 |
| content_description | 424 | 50.2 | 34.4 | 15.9 |
| bound_with | 424 | 27.1 | 12.4 | 14.7 |
| bound_with_minimal | 424 | 26.9 | 12.2 | 14.7 |
| editor_description | 424 | 63.4 | 51.6 | 11.9 |

### Visual/Spatial Arts

| feature | n_in | pct_in | pct_out | delta |
| --- | --- | --- | --- | --- |
| audience | 110 | 33.6 | 22.9 | 10.7 |
| printing_privilege | 110 | 12.7 | 3.3 | 9.5 |
| has_community | 110 | 60.0 | 55.8 | 4.2 |
| dedicatee_name | 110 | 21.8 | 19.1 | 2.7 |
| editor_name | 110 | 74.5 | 76.8 | -2.3 |
| educational_authorities_references | 110 | 11.8 | 15.1 | -3.3 |
| destination_language | 110 | 10.0 | 13.9 | -3.9 |
| origin_language | 110 | 6.4 | 10.4 | -4.0 |

### Instruments/Measurement

| feature | n_in | pct_in | pct_out | delta |
| --- | --- | --- | --- | --- |
| audience | 104 | 39.4 | 22.2 | 17.2 |
| editor_name | 104 | 86.5 | 75.1 | 11.4 |
| enriched_with | 104 | 51.0 | 40.2 | 10.8 |
| date_in_imprint | 104 | 93.3 | 83.4 | 9.9 |
| has_addition | 104 | 56.7 | 48.6 | 8.2 |
| editor_description | 104 | 64.4 | 56.6 | 7.9 |
| location_in_imprint | 104 | 90.4 | 83.1 | 7.3 |
| content_description | 104 | 48.1 | 41.5 | 6.5 |

### Cosmos/Earth

| feature | n_in | pct_in | pct_out | delta |
| --- | --- | --- | --- | --- |
| has_addition | 80 | 60.0 | 48.5 | 11.5 |
| bound_with_minimal | 80 | 28.7 | 18.6 | 10.1 |
| bound_with | 80 | 28.7 | 18.9 | 9.9 |
| enriched_with | 80 | 48.8 | 40.8 | 8.0 |
| institutions | 80 | 36.2 | 28.8 | 7.4 |
| description_of_euclid | 80 | 8.8 | 5.1 | 3.6 |
| has_authority | 80 | 91.2 | 87.7 | 3.6 |
| educational_authorities_references | 80 | 17.5 | 14.4 | 3.1 |

### Applied Mechanics/Military

| feature | n_in | pct_in | pct_out | delta |
| --- | --- | --- | --- | --- |
| audience | 57 | 33.3 | 23.7 | 9.7 |
| has_community | 57 | 64.9 | 55.7 | 9.2 |
| has_addition | 57 | 56.1 | 49.1 | 7.0 |
| printing_privilege | 57 | 10.5 | 4.1 | 6.5 |
| editor_description | 57 | 63.2 | 57.1 | 6.0 |
| bound_with_minimal | 57 | 24.6 | 19.2 | 5.4 |
| bound_with | 57 | 24.6 | 19.5 | 5.1 |
| institutions | 57 | 33.3 | 29.3 | 4.1 |

### Music

| feature | n_in | pct_in | pct_out | delta |
| --- | --- | --- | --- | --- |
| institutions | 18 | 38.9 | 29.3 | 9.6 |
| bound_with_minimal | 18 | 27.8 | 19.4 | 8.4 |
| bound_with | 18 | 27.8 | 19.6 | 8.1 |
| has_authority | 18 | 94.4 | 87.9 | 6.6 |
| educational_authorities_references | 18 | 16.7 | 14.7 | 2.0 |
| editor_name | 18 | 77.8 | 76.5 | 1.3 |
| action_verbs | 18 | 66.7 | 66.4 | 0.2 |
| has_transformation | 18 | 66.7 | 68.5 | -1.8 |

## Preliminary Reading

1. This step should tell us which title-page moves belong to which mathematical work-zones: community/audience, authority/institutions, additions/bound-with material, translation/correction, and identity language.
2. The next pass should inspect examples rather than rely only on percentages, especially where deltas are large.
3. Euclid/Elements should be treated as a feature pattern that cuts across subject families, not as a subject family by itself.
