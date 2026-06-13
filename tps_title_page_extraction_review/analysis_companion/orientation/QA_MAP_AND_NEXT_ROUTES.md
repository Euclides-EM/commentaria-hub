# Q&A Map And Next Routes

This is a compact map of the analysis so far. It is meant to help us choose the next question deliberately.

## Where We Started

The old presentation treated Euclid/Elements title-page identity as more stable than it probably is. The new corpus is much larger, includes richer subject classification, and forces a wider question, but the historical argument should still come back to the Elements:

**What is the place of the metadata-defined Elements corpus in the broader ecology of early modern mathematical print, and how do title pages construct that place through subject identity, social legitimacy, and intellectual/pedagogical values?**

The broad mathematical corpus is the ecology we need in order to understand the Elements. It is not the final object of the talk.

## Branch 1: Corpus And Subject Terrain

Status note:

Branches 1-3 were created before the analysis fully centered the metadata-defined Elements corpus. They remain useful as ecological background and comparative baselines, but their conclusions should be re-read through Branch 4 and not treated as the final Elements argument.

### Q001: What Subject Families Define The Corpus?

Core finding:

The corpus is not a simple pure/applied hierarchy. It is a set of overlapping work-zones: Geometry/Theory, Practical Geometry, Instruments/Measurement, Arithmetic/Commerce, Visual/Spatial Arts, Cosmos/Earth, Applied Mechanics/Military, and Music.

Important consequence:

Subject labels alone cannot explain Euclid/Elements, because standard Euclidean editions may have no positive subject label by classifier design.

Main files:

- `phase1b_subject_families_and_calibration.md`
- `derived_data/subject_family_counts.csv`
- `derived_data/primary_subject_pair_counts.csv`
- `derived_data/no_primary_but_geometry_euclid_like_title_pages.csv`

### Q002: How Do Subject Families Present Themselves?

Core finding:

Different subject families have different title-page grammars. Geometry/Theory has strong identity signals; Instruments/Measurement and Applied Mechanics/Military have stronger community/audience signals; Cosmos/Earth and some learned books lean toward additions and bound-with material.

Important consequence:

Title pages are not just labels. They sort books into communities, authorities, and uses.

Main files:

- `phase2_feature_by_subject_profiles.md`
- `derived_data/subject_family_feature_prevalence.csv`
- `derived_data/subject_family_feature_contrasts.csv`

## Branch 2: Social Worlds

### Q005: How Do Social Groups And Intellectual Values Interact?

Core finding:

Military, surveying/engineering, institutional, Jesuit/religious, and reader/lovers publics are associated with different intellectual values: utility, method, correction, enrichment, ancient authority, or ease.

Important consequence:

The social story is not merely context. Social address helps explain why a title page foregrounds certain mathematical virtues.

Main files:

- `phase3b_social_intellectual_crosswalk.md`
- `derived_data/social_group_intellectual_value_crosswalk.csv`
- `derived_data/social_group_examples.csv`

### Q006: What Social Evidence Types Are Actually Distinct?

Core finding:

Audience, institutions, editor credentials, and patronage behave differently. Patronage is not readership. Editor credentials often make expertise legible. Institutions authorize and teach. Audience marks imagined use.

Important consequence:

We must not collapse every named person or institution into one generic "social" category.

Main files:

- `phase4a_deep_social_topology.md`
- `derived_data/social_taxonomy_counts.csv`
- `derived_data/social_taxonomy_examples.csv`

### Q008 Social Half: What Broader Social Arenas Appear?

Core finding:

Social arenas include learned/scholarly authority, school pedagogy, court/state service, religious institutions, patronage/prestige, professional practice, public readers/lovers, and military settings.

Important consequence:

Mathematical authority is socially distributed. It can come from ancient authors, offices, institutions, professions, patrons, schools, or publics.

Main files:

- `phase4d_deeper_social_arenas.md`
- `derived_data/rich_social_arena_counts.csv`
- `derived_data/rich_social_arena_claim_crosswalk.csv`

## Branch 3: Intellectual And Pedagogical Values

### Q003: Which Values Are Promoted?

Core finding:

Title pages promote usefulness, ease, correction, novelty, translation/language, demonstration/method, enrichment/addition, restoration/ancient authority, and professional/community value.

Important consequence:

The intellectual story is not just subject matter. Title pages advertise what kind of mathematical knowledge is valuable.

Main files:

- `phase3_intellectual_pedagogical_values.md`
- `derived_data/intellectual_values_by_subject_family.csv`
- `derived_data/intellectual_value_examples.csv`

### Q007: What Intellectual Claim Modes Appear?

Core finding:

Title pages often describe acts performed on knowledge: translating, correcting, adding, demonstrating, extracting, publishing, and furnishing with apparatus.

Important consequence:

Mathematical books are presented as products of transmission and transformation, not static containers.

Main files:

- `phase4b_deep_intellectual_topology.md`
- `derived_data/intellectual_value_counts_deep.csv`
- `derived_data/action_verb_counts.csv`

### Q008 Intellectual Half: What Richer Claim Modes Appear?

Core finding:

Richer exploratory tags show canonical identity, ancient authority, augmentation/composition, method/order, correction, visual aids, translation, utility, completeness, novelty, access, and extraction.

Important consequence:

The emerging umbrella argument is not "subject X has feature Y." It is that title pages stage mathematics as socially authorized procedures of transmission.

Main files:

- `phase4c_deeper_social_intellectual_paths.md`
- `derived_data/rich_intellectual_claim_mode_counts.csv`
- `derived_data/rich_title_page_archetype_counts.csv`

## Branch 4: Cases And Euclid/Elements

### Q004: Which Cases Show Social Positioning And Intellectual Values Together?

Core finding:

The best cases are not only the most famous ones. Useful cases include professional/military publics, institutional pedagogy, utility/ease, composite books, Euclid identity, broader Elements language, and thin/suppressed social cases.

Important consequence:

We need ordinary cases and weird edge cases together.

Main files:

- `CASEBOOK.md`
- `derived_data/social_intellectual_casebook_candidates.csv`

### Q009: Is Euclid/Elements Stable Or Mobile?

Core finding:

Euclid/Elements is not a single stable designation. It is a mobile authority: recognizable enough to anchor title-page identity, flexible enough to be recomposed for pedagogy, utility, translation, correction, composite apparatus, institutions, and non-Euclidean elemental vocabularies.

Key numbers:

- 332/843 representative works have Euclid/Elements evidence.
- 295/843 mention Euclid explicitly or through normalized Euclid-reference fields.
- 207/843 use elements/principles/fundamentals language.
- 31/843 use elements/principles language without explicit Euclid evidence.
- 92/332 Euclid/Elements representatives have no primary subject classification.
- 211/332 have Geometry/Theory as primary.
- 40/332 have Practical Geometry as primary.

Main files:

- `phase5_euclid_elements_mini_atlas.md`
- `phase5b_euclid_elements_close_reading_shortlist.md`
- `derived_data/euclid_elements_representative_atlas.csv`
- `derived_data/euclid_elements_close_reading_shortlist.csv`

### Q010: How Does The Metadata Elements Corpus Live In The Broader Ecology?

Core finding:

The metadata-defined Elements corpus is not the same thing as the set of title pages using Euclid/Elements wording. It has 321 edition keys, 286 representatives in the title-page analysis corpus, and overlapping internal boundaries: book coverage, language/region, pedagogy, apparatus, institutional setting, format, social authorization, and sparse canonical identity. External textual-lineage labels should not organize the argument.

Important consequence:

The Elements corpus is distinctive, but not uniform. It lives in the broader ecology as a canonical anchor that can become schoolbook, humanist recovery, vernacular practical geometry, institutional textbook, apparatus-rich learned object, or public useful mathematics.

Main files:

- `phase6_metadata_elements_corpus_ecology.md`
- `derived_data/metadata_elements_corpus_ecology_matrix.csv`
- `derived_data/metadata_elements_family_profiles.csv`
- `derived_data/metadata_elements_bookgroup_profiles.csv`
- `derived_data/metadata_elements_fuzzy_boundary_cases.csv`

### Q011: What Natural Modes Appear Without External Lineage Labels?

Core finding:

When we demote external lineage labels and use only our own evidence, the metadata Elements corpus still does not divide into clean sub-corpora. Instead, it has overlapping modes: canonical/sparse, pedagogical/method, vernacular/transfer, institutional, composite/apparatus, and practical/public.

Important consequence:

The center of the Elements corpus is not sparse canonical identity. It is a dense overlap of ancient/canonical authority, institutional authorization, pedagogy/method, and composite apparatus. Practical/public Elements is smaller, but it usually overlaps with pedagogy rather than standing outside the Elements tradition.

Main files:

- `phase7_metadata_elements_natural_modes.md`
- `derived_data/metadata_elements_natural_modes_matrix.csv`
- `derived_data/metadata_elements_natural_mode_counts.csv`
- `derived_data/metadata_elements_natural_mode_overlaps.csv`
- `derived_data/metadata_elements_unsupervised_cluster_k5_summary.csv`

### Q012: Is `1-6 + 11-12` A Distinct Practical-Pedagogical Elements Package?

Core finding:

Yes, provisionally. The `1-6 + 11-12` metadata Elements package is not simply a shortened Elements. It behaves like a flexible pedagogical package: elementary plane geometry plus enough solid geometry to preserve Euclidean authority while making the book especially available for institutional teaching, practical mathematics, military/technical settings, and useful/easy method.

Important consequence:

Book coverage itself can be a historical signal. The Elements does not only vary by language, editor, or title-page designation; it can also be repackaged by selecting mathematical contents that serve particular teaching and use contexts.

Main files:

- `phase8_elements_1_6_plus_solids_package.md`
- `derived_data/metadata_elements_1_6_plus_solids_all_cases.csv`
- `derived_data/metadata_elements_1_6_plus_solids_contrasts.csv`
- `derived_data/metadata_elements_1_6_plus_solids_examples.csv`
- `derived_data/metadata_elements_1_6_plus_solids_language_contrast.csv`

### Q013: How Does Plain `1-6` Compare With `1-6 + 11-12`?

Core finding:

Plain `1-6` and `1-6 + 11-12` are both partial Elements packages, but they do different work. Plain `1-6` is older, broader, more Latin/Dutch, and more tied to elementary Euclid, translation, composite framing, sparse canonical identity, and vernacular access. `1-6 + 11-12` is later, more English/French/Spanish, and much stronger in practical/public, utility, ease, method/access, religious-institutional, and military/academy signals.

Important consequence:

Different abridgments perform different intellectual and social work. The presence of books 11-12 changes what a partial Elements can promise: not only first principles, but usable solid geometry for institutions, academies, military settings, and practical mathematical publics.

Main files:

- `phase9_elements_1_6_vs_1_6_plus_solids.md`
- `derived_data/metadata_elements_1_6_vs_1_6_plus_solids_metric_contrasts.csv`
- `derived_data/metadata_elements_1_6_vs_1_6_plus_solids_metric_contrasts_1650plus.csv`
- `derived_data/metadata_elements_1_6_vs_1_6_plus_solids_metric_contrasts_no_dechales_tacquet.csv`
- `derived_data/metadata_elements_1_6_vs_1_6_plus_solids_examples.csv`

### Q014: Is `1-6 + 11-12` A Bridge To Non-Elements Practical Geometry?

Core finding:

`1-6 + 11-12` is a bridge form, but it bridges from the side of Euclidean authority. It does not become ordinary practical geometry. Compared with non-Elements practical geometry, it is more method/institution/theory-heavy, while sharing utility and practical-public territory.

Important consequence:

The boundary is not Elements versus practical geometry. It is a gradient: canonical Euclid -> usable Euclid -> Euclidean practical geometry -> professional/material practical arts.

Main files:

- `phase10_plus_solids_and_nonmetadata_practical_geometry.md`
- `derived_data/metadata_elements_plus_solids_vs_nonmetadata_practical_contrasts.csv`
- `derived_data/metadata_elements_plus_solids_neighboring_nonmetadata_cases.csv`

### Q015: How Do Natural Elements Modes Overlap With Metadata And Format?

Core finding:

The Phase 7 natural modes are overlapping bundles, not separate subcorpora. The dense center is humanist/ancient + institutional/authority + pedagogical/method + composite/apparatus. Practical/public is an edge that usually remains canonical and pedagogical. Sparse/canonical is another edge, but must be controlled for title-page fashion.

Important consequence:

The Elements corpus is best described as a canonical-institutional-pedagogical-apparatus center with several flexible edges: practical/public, vernacular/transfer, sparse/canonical, and specific book-coverage packages such as `1-6 + 11-12`.

Main files:

- `phase11_natural_modes_metadata_format_ecology.md`
- `derived_data/metadata_elements_natural_modes_matrix_with_format.csv`
- `derived_data/metadata_elements_natural_mode_overlap_pairs.csv`
- `derived_data/metadata_elements_natural_modes_metadata_field_strong_signals.csv`
- `derived_data/titlepage_density_fashion_association_diagnostics.csv`

### Q016: Does Sparse-Canonical Survive Title-Page Fashion Controls?

Core finding:

Sparse-canonical is a real title-page posture, but it is too entangled with city, language, format, period, and institutional visibility to use as a pure intellectual signal. City is the strongest control for the natural sparse/canonical mode; language is also meaningful; format and period matter more moderately.

Important consequence:

Sparse-canonical should be a contrast mode and close-reading category, not a standalone thesis. The safer formulation is: canonical silence is one possible strategy inside the Elements corpus, and we can only identify it after subtracting local title-page fashion.

Main files:

- `phase12_sparse_canonical_fashion_controls.md`
- `derived_data/sparse_fashion_control_associations.csv`
- `derived_data/sparse_fashion_control_rates_long.csv`
- `derived_data/sparse_canonical_cases_with_fashion_controls.csv`
- `derived_data/quiet_non_sparse_cases_for_comparison.csv`

### Q017: Do Author / Editor Portfolios Change When They Enter The Elements?

Status:

Answered provisionally in `phase13_author_editor_portfolios.md`.

Short answer:

Yes, often, within the represented title-page corpus. In same-person represented portfolios, Elements title pages frequently intensify ancient authority, translation/transfer, correction/revision, method/order, and augmentation. Utility/practice is less central than in non-Elements works, but Elements title pages are not simply socially silent.

Limitation:

This is not full-bibliography evidence. It cannot show whether Euclid was central or marginal to a person's whole career.

Key result:

Among 13 bridge portfolios with at least two Elements and two non-Elements cases, ancient authority/restoration is higher in Elements for all 13. Average Elements-minus-non-Elements deltas include ancient authority/restoration +86.3 pp, translation/transfer +35.2 pp, correction/revision +23.8 pp, method/order +19.4 pp, and no visible social arena -30.3 pp.

Interpretation:

Within this title-page corpus, the Elements behaves as canon-work: a recurring site where mathematical workers present themselves as restorers, translators, correctors, organizers, demonstrators, or mediators of an ancient/canonical corpus.

Output files:

- `phase13_author_editor_portfolios.md`
- `derived_data/author_editor_portfolio_person_summary.csv`
- `derived_data/author_editor_portfolio_interesting_people.csv`
- `derived_data/author_editor_portfolio_elements_non_elements_pairs.csv`

### Q018-Q019: What Do Controlled Same-Person Close Readings Show?

Status:

Answered in `phase14_controlled_same_person_close_reading_selector.md` and `phase15_first_controlled_close_readings.md`.

Short answer:

The controlled close readings preserve the canon-work claim but make it more precise. The contrast is not theory versus practice. It is canonical mediation versus more direct practical/procedural application.

Key result:

Elements title pages repeatedly translate, comment, correct, augment, select, demonstrate, and socially relocate Euclid. Neighboring non-Elements works more directly advertise utility, instruments, military technique, surveying, arithmetic procedure, and professional application.

Best cases:

- Henrion;
- Forcadel;
- Clavius;
- Errard;
- Rudd.

### Q020: Is Dutch Plain `1-6` A Distinct Practical-Vernacular Route?

Status:

Answered in `phase16_dutch_plain_1_6_practical_vernacular_route.md`.

Short answer:

Yes. Dutch plain `1-6`, especially Dou, is a separate practical-vernacular Elements route. It reaches almost the same practical/public level as `1-6 + 11-12`, but without adding books 11-12.

Key result:

Dutch plain `1-6` makes plane geometry usable through Dutch translation, explanation, correction, added utilities, geometrical operations, public lovers of the free art, and land-surveyor/wine-gauger authority.

Interpretation:

There are at least two practical Elements pathways:

1. Dutch/Dou plain `1-6`: vernacular, civic-measuring, public lovers/practitioners, operational plane geometry.
2. Later `1-6 + 11-12`: institutional/military/academy/course-oriented usable Euclid, often via solid geometry.

### Q021: New / Reorganized Elements As Anti-Philological Pole

Status:

Answered provisionally in `phase17_new_reorganized_elements.md`.

Seed case:

`Paris_1667`, Arnauld, `Nouveaux Elemens de Geometrie`.

Question:

When, where, and how often do title pages claim "new Elements," "new order," "new demonstrations," or reconstruction/reorganization of the Elements?

Why it matters:

This may be a counter-pole to fidelity-to-the-ancients, ancient-restoration, and philological-correction ideals. It treats the Elements as a structure to be remade for intellectual or pedagogical reasons, not only preserved or transmitted.

Short answer:

The pole exists, but it is a spectrum rather than a binary. Strict reconstructive cases are rare: `Paris_1667` and `Livorno_1709`. A larger hybrid field claims new/easy method, contraction, selected propositions, new demonstrations, symbolic/algebraic retooling, or new order while still invoking correction, translation, augmentation, or Euclidean authority.

### Q022: Deductive / Mathematical Parts On Title Pages

Status:

Parked for later analysis.

Question:

Which parts of mathematical/deductive content are named or highlighted on title pages: axioms, propositions, theorems, demonstrations, diagrams/figures, scholia, corollaries, notes, paradoxes, definitions, postulates, lemmas, enunciations? How are they valued or framed?

Why it matters:

This can show what title pages think mathematical knowledge is made of. It also connects directly to new/reorganized Elements because reconstruction often works by manipulating parts: new demonstrations, selected propositions, reduced proposition counts, scholia, corollaries, notes, figures, algebraic signs.

## Current Best Umbrella Argument

**Title pages as transmission machines.**

Early modern mathematical title pages do not simply label books by subject. They stage mathematical knowledge as socially authorized procedures: translating, correcting, augmenting, demonstrating, extracting, making useful, restoring ancient authority, and fitting texts to particular publics.

Euclid/Elements becomes a key test case because it reveals the logic of the whole corpus: identity can be stable enough to travel while remaining flexible enough to be recomposed.

## Open Tensions

1. Social analysis is now richer, but explicit audience categories are undercounted and need close reading.
2. Intellectual-value tags are useful but heuristic; claims need verification before citation.
3. Euclid/Elements is strong as a pillar, but we should keep two lenses separate: Elements as a metadata-defined corpus and Euclid/Elements as title-page language.
4. Wardhaugh/external lineage labels should be treated as weak reference context or ignored; they should not organize the argument.
5. "Utility/practicality" may be a social claim, an intellectual value, a subject label, or all three. We have not disentangled it yet.
6. Absence/silence may be meaningful, especially in sparse canonical cases, but it may also reflect title-page fashion by period, language, city/region, printer, format, genre, or schoolbook convention. See `ANALYSIS_TERMS.md`.

## Options For Where To Go Next

### Option A: Deductive / Mathematical Parts

Question:

Which parts of mathematical content do title pages name, and what values are attached to them?

Why this is strong:

It gets inside the intellectual texture of the title pages. Instead of only asking whether a book says "Euclid" or "useful," it asks whether title pages foreground demonstrations, propositions, figures, scholia, axioms, theorems, notes, corollaries, or diagrams as the units that make mathematical knowledge valuable.

Likely outputs:

- part-name taxonomy;
- counts by Elements/non-Elements, language, period, bookgroup, and subject family;
- examples showing how each part is valued;
- connection to reconstructive cases like `Paris_1667`, `Paris_1640`, `Cologne_1556`, and `Kiel_and_Leipzig_1699`.

Best if we want:

To deepen the intellectual half before turning to slides.

### Option B: Choose Slide-Level Case Set

Question:

Which 2-4 close-reading clusters should carry the conference argument?

Why this is strong:

We now have enough analysis to stop expanding and start composing. A tight case set will make the talk legible.

Likely outputs:

- shortlist of cases;
- what each case proves;
- which numbers support it;
- slide sequence skeleton.

Best if we want:

To move from analysis to presentation design.

### Option C: Sparse-Canonical Close Reading

Question:

Which sparse-canonical cases remain interesting after fashion controls, and which are likely local/title-page-fashion effects?

Why this is strong:

It turns the control analysis into a usable case set and prevents overreading silence.

Likely outputs:

- close-reading pairs;
- high-fashion-risk cases versus lower-fashion-propensity cases;
- revised examples for the talk.

Best if we want:

To decide whether sparse-canonical belongs in the final talk.

### Option D: Composite Book / Apparatus Deep Dive

Question:

What do additions, tables, diagrams, notes, comments, appendices, bound-with texts, and selected theorems do on title pages?

Why this is strong:

It follows the "transmission machine" argument most directly. It treats books as constructed mathematical tools rather than static texts.

Likely outputs:

- apparatus taxonomy;
- composite-book casebook;
- comparison across Euclid, cosmography, arithmetic, instruments, and institutional books.

Best if we want:

A more material/book-historical argument.

### Option E: Social Legitimacy Deep Dive

Question:

How do audience, institution, office/credential, and patronage differ as mechanisms of mathematical authority?

Why this is strong:

It develops the social half on its own terms, which the user rightly asked for. It also prevents patronage from being mistaken for readership.

Likely outputs:

- separate maps of audience, office, institution, and patronage;
- examples of stacked social legitimacy;
- close-reading cases where social layers conflict or reinforce each other.

Best if we want:

A conference argument centered on communities of learning and authority.

### Option F: Absence And Suppression Deep Dive

Question:

What title pages do not say: no audience, no utility, no method, no correction, no institution, sparse title-only authority?

Why this is strong:

It makes the analysis more subtle. It prevents us from only studying loud, rich, promotional title pages.

Likely outputs:

- sparse/silent title-page typology;
- canonical identity versus promotional richness;
- contrast pairs: sparse Euclid against procedural Euclid, sparse arithmetic against commercial arithmetic, sparse theory against institutional theory.

Best if we want:

A sharper historical argument about authority and silence.

### Option G: Chronology And Language Deep Dive

Question:

Do title-page rhetorics change over time or across Latin/French/English/Dutch/German/Italian/Portuguese/Chinese contexts?

Why this is strong:

It can turn the analysis into a historical narrative rather than only a thematic map.

Likely outputs:

- periodized trend tables;
- language-specific profiles;
- caution map for corpus imbalance and OCR/transcription differences.

Best if we want:

A diachronic paper structure.

## My Recommendation

Go next to **Option A: Deductive / Mathematical Parts** unless we need to start composing slides immediately.

Reason:

Phase 17 showed that reconstruction often happens through named parts: propositions, demonstrations, notes, scholia, figures, algebraic signs, and reduced proposition counts. Q022 is the natural next intellectual layer.

After that, choose between:

- Option B to choose the final slide-level case set;
- Option C to decide whether sparse-canonical belongs in the final talk;
- Option E if the talk should emphasize communities and institutions;
- Option D if the apparatus/composite-book thread becomes the center.
