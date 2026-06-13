# Proposition And Demonstration Notes

Date: 2026-06-12

Purpose:

This note deepens the report's claims about propositions and demonstrations, especially for Sections 5, 8, and 9.

Script:

- `report/scripts/build_prop_demo_deep_dive.py`

Main outputs:

- `tables/report_prop_demo_motif_cases.csv`
- `tables/report_prop_demo_motifs_by_elements_books_group_matrix.csv`
- `tables/report_prop_demo_motifs_by_natural_dominant_mode_matrix.csv`
- `tables/report_prop_demo_motifs_by_period_matrix.csv`
- `tables/report_prop_demo_motifs_by_language_first_matrix.csv`
- `tables/report_prop_demo_motifs_by_format_group_matrix.csv`
- `tables/report_prop_demo_motifs_by_major_edition_cluster_matrix.csv`
- `tables/report_prop_demo_cluster_exclusion_controls.csv`
- `tables/report_prop_demo_close_reading_shortlist.csv`

## Headline Result

The motif finder found 81 metadata-defined Elements representatives with proposition or demonstration motifs.

The strongest report use is not simply "Elements title pages mention propositions and demonstrations." We can say something more precise:

- demonstration language is relatively broad within the Elements corpus;
- explicit proposition-use language is narrower and heavily concentrated in Dechales/Reeve/Williams-style `1-6 + 11-12` editions;
- new/easy/order demonstration language becomes especially visible in the seventeenth century and later;
- proposition selection/reduction is rare but historically important, because it marks active restructuring of Euclid.

## Book-Group Patterns

### `books_1_6_plus_solids`

Rates:

- proposition_any: 35.9%;
- proposition_use_application: 20.5%;
- proposition_explanation_commentary: 23.1%;
- demonstration_any: 12.8%;
- demonstration_easy_clear: 10.3%;
- demonstration_new: 7.7%;
- demonstration_order_method: 10.3%.

Interpretation:

This confirms that `1-6 + 11-12` is strongly proposition/method oriented. But cluster controls matter.

After excluding Dechales/Reeve/Williams:

- proposition_any falls from 35.9% to 12.5%;
- proposition_use_application falls from 20.5% to 0.0%;
- proposition_explanation_commentary falls from 23.1% to 0.0%;
- demonstration_any remains 12.5%;
- demonstration_easy_clear remains 8.3%;
- demonstration_new remains 4.2%;
- demonstration_order_method remains 8.3%.

Report implication:

The exact "use of each/every proposition" formula should be treated as a Dechales/Reeve/Williams route, not as a general property of all `1-6 + 11-12` editions. The broader method/demonstration orientation of `1-6 + 11-12` survives outside that route, but the explicit proposition-use rhetoric does not.

### `books_1_6`

Rates:

- demonstration_any: 23.5%;
- proposition_any: 1.2%;
- proposition_use_application: 1.2%;
- demonstration_easy_clear: 2.5%.

Interpretation:

Plain `1-6` is less proposition-use oriented than `1-6 + 11-12`. Its title-page work is not usually "use of every proposition"; it is more foundational, translational, elementary, and in some routes practical through principles, figures, and operations rather than proposition-use formulae.

This fits the Dutch/Dou route: practical vernacular Euclid can be operational without using the Dechales-style proposition-use language.

### `near_complete_or_expanded`

Rates:

- demonstration_any: 32.8%;
- demonstration_easy_clear: 9.0%;
- proposition_any: 9.0%;
- proposition_ordering_reduction: 1.5%;
- proposition_selection_extraction: 1.5%;
- proposition_and_demonstration: 3.0%.

Interpretation:

Near-complete/expanded editions are more demonstration-heavy than proposition-use heavy. They are closer to a learned demonstrative corpus: complete/expanded Euclid, demonstrated, corrected, translated, augmented, sometimes with scholia or apparatus.

## Natural Mode Patterns

### `pedagogical/method`

- demonstration_any: 41.4%;
- demonstration_easy_clear: 13.8%;
- proposition_any: 24.1%;
- proposition_ordering_reduction: 3.4%;
- proposition_selection_extraction: 3.4%.

Interpretation:

This is the strongest mode for proof as advertised method. It supports the "logical/pedagogical restoration" language.

### `practical-pedagogical`

- demonstration_any: 19.4%;
- demonstration_easy_clear: 12.5%;
- proposition_any: 19.4%;
- proposition_use_application: 11.1%;
- proposition_explanation_commentary: 12.5%;
- demonstration_order_method: 6.9%.

Interpretation:

Practical-pedagogical Elements is where proposition-use, explanation, method, and utility intersect. This is the strongest social-intellectual bridge mode.

### `institutional-composite`

- demonstration_any: 26.5%;
- proposition_any: 4.4%;
- scholia/commentary was stronger in the broader deductive-parts analysis.

Interpretation:

Institutional-composite Elements is demonstration/commentary/apparatus oriented, not proposition-use oriented.

## Chronological Patterns

Demonstration language rises after 1600:

- pre-1550: 6.2%;
- 1550-1599: 14.5%;
- 1600-1649: 25.3%;
- 1650-1699: 27.7%;
- 1700+: 15.0%.

Proposition-use and proposition-explanation are most visible in 1650-1699 and 1700+, largely because of Dechales/Reeve/Williams and related easy-method routes.

New/order/method demonstration language:

- demonstration_new: 2.4% in 1650-1699, 5.0% in 1700+;
- demonstration_order_method: 2.4% in 1650-1699, 10.0% in 1700+.

Interpretation:

The seventeenth-century and later title-page rhetoric increasingly presents Euclid as something to be demonstrated in a clearer, easier, newer, or differently ordered way. This is consistent with the reconstruction/restoration argument.

## Language Patterns

English Elements are striking in this motif analysis:

- demonstration_any: 45.0%;
- demonstration_easy_clear: 20.0%;
- demonstration_new: 15.0%;
- demonstration_order_method: 15.0%;
- proposition_any: 35.0%;
- proposition_use_application: 15.0%.

French also shows proposition explanation/use:

- proposition_explanation_commentary: 13.2%;
- proposition_use_application: 9.4%;
- demonstration_any: 17.0%.

Latin is more demonstration-heavy than proposition-use heavy:

- demonstration_any: 25.9%;
- proposition_any: 5.8%;
- proposition_use_application: 0.7%.

Interpretation:

This supports a language-sensitive claim. Vernacular or translingual contexts, especially English and French, make proposition-use and easy/new method more title-page-visible. Latin remains important for demonstration, but often in a more learned or institutional mode.

## Important Cases

### Proposition-Use Route

Mainly Dechales/Reeve/Williams:

- `Lausanne_1683`;
- `Oxford_1685`;
- `London_1696`;
- `Oxford_London_1700`;
- `Amsterdam_1700`;
- `Paris_1682`;
- `Paris_1683`;
- `Paris_1690`;
- `London_1685a`;
- `London_1703`.

Use carefully:

These cases can support a proposition-use route, but not a general claim that all practical Elements title pages foreground proposition-use.

### Selection / Ordering / Reduction

Rare but very important:

- `Cologne_1556`: selected propositions from following books ordered for demonstration.
- `Paris_1640`: book 10 reduced to 62 propositions with new, easier, more succinct demonstrations.
- `Strasbourg_1564b`: propositions from remaining books for users lacking Euclid's volume.

Use:

These cases are excellent for showing the Elements as manipulable structure.

### New / Easy / Ordered Demonstration

Cases:

- `Paris_1667`: new order and new demonstrations.
- `473J72`: demonstrated after a new, plain, easy method.
- `London_1685a` and `London_1703`: explained and demonstrated in a new/easy method.
- `The_Hague_1758`: six first books put in a new order for youth.
- `Rome_1679`: restored/prisca geometria, shorter/easier context.
- `Naples_1701`: new method and compendious demonstration.
- Barrow editions: briefly/succinctly demonstrated.

Use:

These cases support the claim that reconstructive/restorative Euclid often works through demonstration and order rather than through axioms/postulates.

## Report Claim To Use

The title-page history of propositions and demonstrations is not uniform. Demonstration language is broadly important to Elements presentation, especially from the seventeenth century onward. Proposition-use, however, is a narrower and more historically specific route, concentrated in Dechales/Reeve/Williams-style practical-pedagogical `1-6 + 11-12` editions. Other editions manipulate propositions differently: selecting them, ordering them, reducing their number, extracting them, or using them as a portable digest of Euclidean structure.

This lets the report avoid overgeneralization while preserving the larger argument: early modern title pages present Euclid's Elements as a corpus whose internal proof-units can be demonstrated, reordered, selected, explained, and made useful.
