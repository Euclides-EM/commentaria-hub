# Question Log

Use this file to record each analysis question, the data used, what we found, and what question came next.

## Template

### Q000: Question Title

Date:

Question:

Data/files:

Method:

Findings:

Interpretation:

Open questions:

Follow-up:

---

## Questions

### Q001: What Subject Families Define The Corpus?

Date: 2026-06-10

Question:

Which broad subject relationships make the 843 representative works legible, and are there calibration issues before we interpret the subject labels?

Data/files:

- `derived_data/classification_representative_matrix.csv`
- `derived_data/classification_representative_full_long.csv`
- `derived_data/title_page_analysis_matrix.csv`

Method:

Grouped the 20 subject labels into broad working families; counted primary/secondary/unknown/unrelated status by family; computed common primary-subject and primary-or-secondary subject pairings; flagged title pages with Euclid/Elements/geometry-like TPS fields but no positive subject classification.

Findings:

- Geometry/Theory is the largest broad family, but standard Euclidean editions can still have no positive subject label because the classifier prompt says not to mark Practical Geometry or Theoretical Mathematics merely because a work is Euclidean.
- Practical Geometry is the major bridge subject, co-occurring often with Instrument Use, Surveying, Theoretical Mathematics, Arithmetic, Perspective, Construction, Architecture, and Military Engineering.
- The corpus looks like overlapping work-zones rather than a simple pure/applied hierarchy.
- 160 representative works have no primary subject.
- 100 title-page editions have no primary subject but geometry/Euclid/Elements-like TPS fields.

Interpretation:

Use subject classification and title-page features together. Subject labels identify topical zones; TPS features identify positioning, authority, identity, and paratextual framing. Euclid/Elements reception cannot be recovered from subject labels alone.

Open questions:

- Which title-page features distinguish subject families?
- Are Euclidean title pages exceptional in feature grammar, even when not positively classified by subject?
- Which no-primary Euclid/Elements cases are expected standard-Euclid cases versus actual classifier misses?

Follow-up:

Move to feature-by-subject profiles.

---

### Q002: How Do Subject Families Present Themselves On Title Pages?

Date: 2026-06-10

Question:

Which title-page feature families are especially associated with each subject family?

Data/files:

- `derived_data/title_page_analysis_matrix.csv`
- `derived_data/representative_analysis_matrix.csv`
- `derived_data/subject_family_feature_prevalence.csv`
- `derived_data/subject_family_feature_contrasts.csv`
- `derived_data/subject_family_feature_examples.csv`

Method:

Aggregated title-page feature presence at the representative-work level, so reprints do not dominate. Grouped features into identity, authority, transformation, addition, community, and imprint families. Compared prevalence by subject family and pulled examples for high-signal family/feature combinations.

Findings:

- Geometry/Theory has the strongest identity signature: `elements_designation`, `base_content`, `references_to_euclid`, and `content_description` are all elevated.
- Instruments/Measurement has a strong audience/community signal, especially named users such as geometers, surveyors, architects, soldiers, painters, sculptors, engineers, and students.
- Visual/Spatial Arts also shows an audience/community signal, often around painters, architects, sculptors, nobles, and royal/dedicatory contexts.
- Cosmos/Earth stands out through addition/bound-with material and institutions.
- Applied Mechanics/Military is audience-heavy and community-heavy, with soldiers, engineers, captains, and military-professional framing.
- Arithmetic/Commerce has elevated `edition_details`, `audience`, and `action_verbs`, but the signal is subtler than the practical/instrumental families.

Interpretation:

Title pages appear to function as social sorting devices. They do not only name a text; they identify the relevant community of use, the authority behind the book, and the kind of mathematical activity the book belongs to.

Open questions:

- Should the first major interpretive thread be "communities of use"?
- How does Euclid/Elements interact with this community signal?
- Are additions/bound-with materials especially important for learned/cosmographic/Euclidean books?

Follow-up:

Inspect community/audience language across subjects, then compare it to Euclid/Elements identity language.

---

### Q003: Which Intellectual And Pedagogical Values Are Promoted On Title Pages?

Date: 2026-06-10

Question:

Can we detect intellectual and pedagogical ideals in the title-page feature language, and how do those values relate to social positioning?

Data/files:

- `derived_data/representative_analysis_matrix.csv`
- `derived_data/representative_analysis_matrix_with_values.csv`
- `derived_data/intellectual_values_by_subject_family.csv`
- `derived_data/intellectual_values_by_subject_family_contrasts.csv`
- `derived_data/intellectual_values_by_social_marker_contrasts.csv`
- `derived_data/intellectual_value_examples.csv`

Method:

Built a first-pass multilingual heuristic taxonomy over title-page feature text. The categories are `clarity_ease`, `utility_use`, `correction_revision`, `novelty_invention`, `translation_language`, `demonstration_method`, `enrichment_addition`, `restoration_ancient_authority`, and `community_professional`. Counts are rough pointers, not final interpretations.

Findings:

- Social and intellectual signals are intertwined, not separate.
- Representatives with professional/community language are much more likely to also carry enrichment/addition, utility/use, demonstration/method, and correction/revision language.
- Instruments/Measurement and Applied Mechanics/Military show strong connections between social address and utility/practice.
- Geometry/Theory shows comparatively stronger restoration/ancient authority, translation/language, and enrichment signals.
- Cosmos/Earth is relatively addition/bound-with heavy.

Interpretation:

The emerging frame should be social-intellectual: title pages position books within communities while also promoting specific ideals of mathematical knowledge, such as usefulness, ease, demonstration, correction, enrichment, and ancient authority.

Open questions:

- Which value categories need better multilingual patterns?
- Which examples best show social address and intellectual value working together?
- Which title pages suppress audience/use and instead foreground textual authority or ancient restoration?

Follow-up:

Inspect examples from `intellectual_value_examples.csv`, refine the taxonomy, then build a social-intellectual casebook.

---

### Q004: Which Cases Best Show Social Positioning And Intellectual Values Together?

Date: 2026-06-10

Question:

Which title pages can serve as close-reading cases for the social-intellectual argument?

Data/files:

- `derived_data/representative_analysis_matrix_with_values.csv`
- `derived_data/intellectual_value_examples.csv`
- `derived_data/social_intellectual_casebook_candidates.csv`
- `CASEBOOK.md`

Method:

Curated candidate cases across several buckets: professional audience, military practice, institutional pedagogy, utility/ease, enrichment/composite books, Euclid authority/identity, broader Elements language, and thin/suppressed social cases.

Findings:

- Professional and military cases are especially good for showing social address and intellectual values together: named publics are linked to utility, practice, correction, enrichment, and method.
- Institutional pedagogy cases connect students, academies, Jesuit/royal classrooms, and Euclidean or astronomical authority.
- Euclid cases split into several forms: canonical identity with thin subject labels; Euclid as easy/useful pedagogical tool; Euclid as applied/composite machinery; and broader "Elements" language beyond simple Euclidean naming.
- Some cases are valuable precisely because the subject classifier suppresses standard Euclid, while the title-page segmentation preserves Euclidean identity.

Interpretation:

The casebook supports a possible argument that title pages construct mathematical knowledge by linking social publics to intellectual virtues: use, ease, correction, demonstration, enrichment, authority, and institutional order.

Open questions:

- Which 8-12 cases are strongest for slides?
- Which examples have good title-page images available in `shelfmarks.csv`?
- Which cases need manual transcription/image spot-checking before being cited?

Follow-up:

Prune the casebook into a slide shortlist and begin drafting candidate argument structures.

---

### Q005: How Do Social Groups And Intellectual Values Interact?

Date: 2026-06-10

Question:

Can we analyze social positioning and intellectual/pedagogical values together, instead of treating them as separate dimensions?

Data/files:

- `derived_data/representative_analysis_matrix_social_values.csv`
- `derived_data/social_group_intellectual_value_crosswalk.csv`
- `derived_data/social_group_subject_family_crosswalk.csv`
- `derived_data/social_group_examples.csv`
- `phase3b_social_intellectual_crosswalk.md`

Method:

Built heuristic social groups from `audience`, `institutions`, `editor_description`, and `dedicatee_name`: students/schools, universities/academies, Jesuit/religious, royal/court patronage, civic/state, artisans/visual trades, military, surveying/engineering, merchants/commercial, and general readers/lovers. Crossed these social markers with heuristic intellectual/pedagogical values.

Findings:

- Military and surveying/engineering publics are strongly associated with `utility_use`, `community_professional`, and correction/revision.
- Artisans/visual trades, though a small group, are strongly associated with utility and method/demonstration.
- Jesuit/religious and institutional settings connect community markers with novelty, enrichment, clarity/ease, restoration/ancient authority, and utility/use.
- Universities/academies show stronger restoration/ancient authority, utility/use, and enrichment than the corpus average.
- Royal/court patronage correlates with utility and demonstration/method, but should be distinguished from actual audience.
- Some Euclidean/title-identity pages have little explicit social address, suggesting a different authority mode centered on text, antiquity, or canonical identity.

Interpretation:

This is probably the central analytical bridge: title pages define mathematical knowledge by linking publics to virtues. The social terrain helps explain why certain intellectual values are promoted: usefulness for soldiers and engineers, method and authority for students and academies, enriched/corrected composite knowledge for institutional or professional use, and textual/ancient identity for Euclidean works.

Open questions:

- Which social groups are robust enough statistically, and which should be used mainly as case-study categories?
- How do Euclid/Elements title pages fit into or resist this social-intellectual model?
- Which values are suppressed or absent in sparse canonical title pages?

Follow-up:

Return to Euclid/Elements through this social-intellectual framework, not through the old "stable designation" frame.

---

### Q006: What Social Evidence Types Are Actually Distinct?

Date: 2026-06-10

Question:

Can we separate audience, institution, editor credential, and patronage instead of treating them all as one social signal?

Data/files:

- `derived_data/representative_analysis_matrix_deep_social.csv`
- `derived_data/social_taxonomy_counts.csv`
- `derived_data/social_taxonomy_subject_family_crosswalk.csv`
- `derived_data/social_taxonomy_intellectual_value_crosswalk.csv`
- `phase4a_deep_social_topology.md`

Findings:

- Social evidence is layered: explicit audience, institutions, editor credentials, and patronage behave differently.
- Editor credentials are common and make expertise legible; explicit audience is less frequent and probably undercounted by strict pattern matching.
- Patronage is common, but it should not be mistaken for readership.
- Institutions and offices often help explain why method, utility, correction, and authority are foregrounded.

Interpretation:

The social title page is not simply a list of people. It is a legitimacy machine: authority can be built through institutional affiliation, professional office, dedication, audience address, or learned credentials.

---

### Q007: What Intellectual And Pedagogical Claim Modes Appear?

Date: 2026-06-10

Question:

Can we analyze intellectual/pedagogical values on their own terms before crossing them with the social layer?

Data/files:

- `derived_data/representative_analysis_matrix_deep_intellectual.csv`
- `derived_data/intellectual_value_counts_deep.csv`
- `derived_data/intellectual_value_subject_family_crosswalk_deep.csv`
- `derived_data/action_verb_counts.csv`
- `phase4b_deep_intellectual_topology.md`

Findings:

- Title pages often describe acts performed on knowledge: translating, correcting, adding, demonstrating, extracting, publishing, and furnishing with apparatus.
- Major intellectual families include identity naming, edition transformation, addition/composition, authority reference, language transfer, and content definition.
- Sparse canonical identity is meaningful: some Euclidean or authoritative texts need few explicit claims because title/author/textual identity do the work.

Interpretation:

The intellectual story is not just subject matter. It is about procedures of transmission: how mathematical knowledge is made trustworthy, teachable, usable, corrected, augmented, or restored.

---

### Q008: What Happens When We Go Deeper Than The Conservative Tags?

Date: 2026-06-10

Question:

Can broader exploratory tags reveal additional social and intellectual patterns without forcing the corpus into the old argument?

Data/files:

- `derived_data/representative_analysis_matrix_richer_claims.csv`
- `derived_data/rich_intellectual_claim_mode_counts.csv`
- `derived_data/rich_title_page_archetype_counts.csv`
- `derived_data/rich_claims_by_period.csv`
- `derived_data/rich_claims_by_language.csv`
- `derived_data/rich_claims_by_primary_subject_family.csv`
- `derived_data/rich_claims_by_social_category.csv`
- `derived_data/rich_absence_suppression_counts.csv`
- `derived_data/representative_analysis_matrix_rich_social_arenas.csv`
- `derived_data/rich_social_arena_counts.csv`
- `derived_data/rich_social_arena_claim_crosswalk.csv`
- `derived_data/rich_social_arena_subject_crosswalk.csv`
- `phase4c_deeper_social_intellectual_paths.md`
- `phase4d_deeper_social_arenas.md`

Findings:

- Rich intellectual claim modes show a broader world of mathematical title-page rhetoric: canonical identity, ancient authority, augmentation/composition, method/order, correction, visual aids, translation, utility, completeness, novelty, access, and extraction.
- Working archetypes are more analytically useful than a flat list: composite workshop books, procedural pedagogical identity, humanist transfer books, utility public books, sparse canonical identity, and method/access books.
- Social arenas include learned/scholarly authority, school pedagogy, court/state service, religious institutions, patronage/prestige, professional practice, public readers/lovers, and military settings.
- Professional/practical and military arenas strongly intensify utility claims; religious and pedagogical arenas intensify method, canonical authority, correction, and procedural identity.
- More than half the representatives have no broad visible social arena under these exploratory tags, which may itself be meaningful.

Interpretation:

The emerging argument should not be "subject X has feature Y." A better frame is that title pages stage mathematics as socially authorized procedures of transmission. Euclid/Elements becomes especially important because it is not stable in one way: it can act as canonical identity, ancient authority, teachable method, practical geometry, humanist transfer, or composite apparatus.

Caveat:

The rich tags are deliberately broad and sometimes over-inclusive. For example, ancient authority/restoration includes Euclid references and therefore should be used to find cases, not as a final quantitative claim without close reading.

Follow-up:

Build a Euclid/Elements mini-atlas and a small close-reading shortlist contrasting sparse canonical identity against procedural/composite/pedagogical title pages.

---

### Q009: Is Euclid/Elements A Stable Designation Or A Mobile Authority?

Date: 2026-06-10

Question:

Can we test the old presentation claim that base designation is highly consistent, especially for Euclid/Elements?

Data/files:

- `derived_data/euclid_elements_representative_atlas.csv`
- `derived_data/euclid_elements_atlas_counts.csv`
- `derived_data/euclid_elements_mode_counts.csv`
- `derived_data/euclid_elements_modes_by_period.csv`
- `derived_data/euclid_elements_modes_by_language.csv`
- `derived_data/euclid_elements_modes_by_subject_family.csv`
- `derived_data/euclid_elements_mode_examples.csv`
- `derived_data/euclid_elements_edge_cases.csv`
- `phase5_euclid_elements_mini_atlas.md`
- `phase5b_euclid_elements_close_reading_shortlist.md`

Findings:

- 332 of 843 representative works, 39.4%, have Euclid/Elements evidence under the current feature-based definition.
- 295 representatives, 35.0%, mention Euclid explicitly or through normalized Euclid-reference fields.
- 207 representatives, 24.6%, use elements/principles/fundamentals language.
- 31 representatives, 3.7% of all representatives and 9.3% of the Euclid/Elements universe, use elements/principles language without explicit Euclid evidence.
- 92 Euclid/Elements representatives, 27.7% of the Euclid/Elements universe, have no primary subject classification.
- Only 211 Euclid/Elements representatives, 63.6%, have Geometry/Theory as a primary family, and 40, 12.0%, have Practical Geometry as primary.

Interpretation:

The old claim should be inverted. Euclid/Elements title-page identity is not stable in the simple sense. It is stable enough to be recognized, but flexible enough to be recomposed. It can function as sparse canonical identity, pedagogical method, practical geometry, humanist transfer, composite apparatus, institutional classroom text, professional utility, or a broader elemental vocabulary outside explicit Euclid.

Caveat:

The Euclid/Elements feature definition required regex and field normalization, including split forms like `EV- clidis`. The `elements_language_no_euclid` bucket excludes explicit `references_to_euclid` and `description_of_euclid` evidence, but every edge case still needs close reading before being used in a slide.

Follow-up:

Use the close-reading shortlist to compare ordinary and edge cases: `Venice_1482`, `TTORPR`, `FA6MQ0`, `Amsterdam_1616`, `London_1685a`, `35H84Y`, `Lisbon_1735`, `Paris_1667`, `Cambridge_1703`, `bib-190`, `Beijing_1607`, and `Lausanne_1683`.

---

### Q010: How Does The Metadata Elements Corpus Live In The Broader Ecology?

Date: 2026-06-10

Question:

Instead of studying books whose title pages are designated as Elements, can we study the metadata-defined corpus of books primarily belonging to Euclid's *Elements*? What are its audiences, intellectual values, social forms, and internal boundaries?

Data/files:

- `ocrflow/store/items_metadata/metadata_elements_print.csv`
- `derived_data/metadata_elements_corpus_ecology_matrix.csv`
- `derived_data/metadata_elements_corpus_overlap_counts.csv`
- `derived_data/metadata_elements_vs_non_elements_contrasts.csv`
- `derived_data/metadata_elements_family_profiles.csv`
- `derived_data/metadata_elements_bookgroup_profiles.csv`
- `derived_data/metadata_elements_wardhaugh_internal_signals.csv`
- `derived_data/metadata_elements_bookgroup_internal_signals.csv`
- `derived_data/metadata_elements_fuzzy_boundary_cases.csv`
- `phase6_metadata_elements_corpus_ecology.md`

Findings:

- The metadata Elements corpus has 321 edition keys; 320 are present in the current title-page analysis corpus.
- After representative/reprint grouping, there are 286 metadata Elements representatives in the 843-work title-page analysis corpus.
- 254 of those 286 also have title-page Euclid/Elements signal; 32 do not under the current feature extraction.
- 78 representatives have title-page Euclid/Elements signal but are not metadata Elements works; these belong to the wider ecology, not the core Elements corpus.
- The metadata Elements corpus is distinctive in identity, ancient authority, Geometry/Theory, procedural/pedagogical framing, language transfer, institutions, editor credentials, and composite/additional apparatus.
- It is less strongly associated than non-Elements works with Arithmetic/Commerce, Visual/Spatial Arts, utility/practical rhetoric, professional/practical social arenas, and explicit audience fields.
- Internal families matter: Dou is strongly procedural/public/composite; Clavius and Tacquet are more institutional/religious/scholarly; Commandino, Magnienus/Gracilis/Dasypodius, and d'Étaples lean toward humanist transfer; Barrow/Hérigone and Foix are procedural/composite in different ways.

Interpretation:

The metadata Elements corpus has native divisions, but not clean hard-edged sub-corpora. Its boundaries are flexible and overlapping: textual/editorial lineage, book coverage, language/region, pedagogy, apparatus, social authorization, and sparse canonical identity all divide it differently. The Elements lives in the broader mathematical ecology as a canonical anchor that can become schoolbook, humanist recovery, vernacular practical geometry, institutional textbook, apparatus-rich learned object, or public useful mathematics.

Open questions:

- Which internal boundary should organize the talk: textual lineage, book coverage, language, social setting, or pedagogical rhetoric?
- Are Clavius/Tacquet/Barrow/Dou better treated as lineages, brands, or pedagogical packages?
- Do vernacular Elements editions behave more like practical geometry/instrument books than like Latin complete Elements editions?
- Are six-book editions the main zone where Elements becomes pedagogical/practical?
- Which non-Elements books look most Elements-like, and what does that tell us about elemental pedagogy beyond Euclid?

---

### Q011: What Natural Modes Appear Inside The Metadata Elements Corpus Without External Lineage Labels?

Date: 2026-06-10

Question:

If we disregard or heavily demote external lineage labels, does the metadata Elements corpus divide naturally according to our own evidence?

Data/files:

- `derived_data/metadata_elements_natural_modes_matrix.csv`
- `derived_data/metadata_elements_natural_mode_counts.csv`
- `derived_data/metadata_elements_natural_mode_overlaps.csv`
- `derived_data/metadata_elements_natural_modes_by_bookgroup.csv`
- `derived_data/metadata_elements_natural_modes_by_language.csv`
- `derived_data/metadata_elements_natural_modes_by_period.csv`
- `derived_data/metadata_elements_natural_mode_examples.csv`
- `derived_data/metadata_elements_neighboring_non_elements_cases.csv`
- `derived_data/metadata_elements_unsupervised_cluster_scores.csv`
- `derived_data/metadata_elements_unsupervised_cluster_k5_summary.csv`
- `phase7_metadata_elements_natural_modes.md`

Findings:

- Natural modes are strongly overlapping, not clean bins.
- The most common modes are humanist/ancient, institutional authority, pedagogical/method, composite apparatus, corrected/updated, vernacular/transfer, practical/public, and sparse/canonical.
- Sparse/canonical Elements is a minority mode, not the center of the corpus.
- The center of the corpus is a dense overlap of ancient/canonical authority, institutional authority, pedagogy/method, and composite apparatus.
- Practical/public Elements is a minority mode, but when it appears it usually overlaps with pedagogy/method.
- Book coverage matters: `1-6 + 11-12` editions look especially practical/method/access; near-complete/expanded editions are often composite and humanist/ancient; six-book editions are large but internally varied.
- A simple unsupervised clustering sanity check produced low silhouette scores, supporting the conclusion that boundaries are fuzzy rather than hard.

Interpretation:

The metadata Elements corpus has flexible axes rather than natural species. Useful modes are canonical/sparse, pedagogical/method, vernacular/transfer, institutional, composite/apparatus, and practical/public. These axes cross rather than align. The Elements lives as a canonical object repeatedly repackaged for school, institution, vernacular pedagogy, practical use, learned apparatus, and ancient textual authority.

Open questions:

- Is `1-6 + 11-12` a distinct practical-pedagogical package?
- Which neighboring non-Elements cases should we use to show fuzzy boundaries?
- Which close-reading cases best show overlap rather than clean division?

---

### Q012: Is `1-6 + 11-12` A Distinct Practical-Pedagogical Elements Package?

Date: 2026-06-10

Question:

Inside the metadata-defined Elements corpus, does the book-coverage pattern `1-6 + 11-12` behave like a meaningful historical package rather than merely an abbreviated Elements?

Data/files:

- `derived_data/metadata_elements_1_6_plus_solids_all_cases.csv`
- `derived_data/metadata_elements_1_6_plus_solids_examples.csv`
- `derived_data/metadata_elements_1_6_plus_solids_contrasts.csv`
- `derived_data/metadata_elements_1_6_plus_solids_language_contrast.csv`
- `derived_data/metadata_elements_1_6_plus_solids_distribution_period.csv`
- `phase8_elements_1_6_plus_solids_package.md`

Findings:

- The package contains 39 representative metadata Elements works.
- It is concentrated after 1650: 26/39 cases are from 1650-1699 and 13/39 are from 1700+.
- Compared with other metadata Elements editions, it over-indexes practical/public rhetoric, method/demonstration/order, utility/application, accessibility/ease, explicit institutions, religious/institutional settings, and military/fortification settings.
- It is more English, French, and Spanish than the rest of the Elements corpus, and less Latin-dominant, though Latin remains a substantial part of the package.
- Its main local styles include Dechales-style easy/useful Euclid, Tacquet-style institutional Euclid with Archimedean/composite apparatus, military/academy Euclid, and a smaller sparse/reduced group.

Interpretation:

`1-6 + 11-12` looks like a flexible pedagogical package: elementary plane geometry plus enough solid geometry to preserve Euclidean authority while making the book useful for institutional teaching, practical mathematics, and military/technical settings. It should not be treated as simply a damaged or shortened full Elements. It is one of the places where the Elements becomes a teachable and usable object.

Open questions:

- How does `1-6 + 11-12` compare directly with plain `1-6` editions?
- Which examples best represent the package without relying too heavily on Dechales or Tacquet clusters?
- Does this package align with diagrams, figures, or visual title-page features?
- Is `1-6 + 11-12` a bridge between metadata Elements editions and non-Elements practical geometry books?

---

### Q013: How Does Plain `1-6` Compare With `1-6 + 11-12`?

Date: 2026-06-10

Question:

Is the `1-6 + 11-12` practical-pedagogical signal really distinctive, or do all short Elements editions behave that way?

Data/files:

- `phase9_elements_1_6_vs_1_6_plus_solids.md`
- `derived_data/metadata_elements_1_6_vs_1_6_plus_solids_all_cases.csv`
- `derived_data/metadata_elements_1_6_vs_1_6_plus_solids_metric_contrasts.csv`
- `derived_data/metadata_elements_1_6_vs_1_6_plus_solids_metric_contrasts_1650plus.csv`
- `derived_data/metadata_elements_1_6_vs_1_6_plus_solids_metric_contrasts_no_dechales_tacquet.csv`
- `derived_data/metadata_elements_1_6_vs_1_6_plus_solids_examples.csv`

Findings:

- Plain `1-6` has 81 representatives; `1-6 + 11-12` has 39.
- Plain `1-6` is broader and older, with substantial Latin and Dutch components and cases from the sixteenth and early seventeenth centuries.
- `1-6 + 11-12` is essentially post-1650 in this corpus and is more English, French, and Spanish.
- Against plain `1-6`, `1-6 + 11-12` strongly over-indexes practical/public mode, religious institutional arena, Jesuit markers, utility-public books, ease/accessibility, utility/application, general readers/lovers, military/fortification, method/access, and method/demonstration/order.
- Plain `1-6` is stronger in translation/vernacular transfer, composite-workshop form, augmentation/enrichment, sparse/canonical mode, and no-visible-social-arena cases.
- The contrast survives a 1650+ chronology check.
- The Dechales/Tacquet cluster strongly shapes the ease/method/Jesuit part of the `1-6 + 11-12` signal, but after removing Dechales/Tacquet, the military/practical/public signal remains strong.

Interpretation:

Plain `1-6` and `1-6 + 11-12` are both partial Elements packages, but they do different historical work. Plain `1-6` often functions as elementary Euclid: foundation, vernacular transfer, school text, readerly text, or short canonical object. `1-6 + 11-12` more often functions as usable Euclid: plane fundamentals plus enough solid geometry to serve institutional, academy, military, and practical mathematical settings.

Open questions:

- Which non-Elements practical geometry books are closest to `1-6 + 11-12`?
- Are Dutch plain `1-6` editions a separate practical-vernacular route parallel to the later English/French/Spanish `1-6 + 11-12` route?
- Do diagrams or apparatus features reinforce the book-coverage distinction?

---

### Q014: Is `1-6 + 11-12` A Bridge To Non-Elements Practical Geometry?

Date: 2026-06-10

Question:

Does the metadata Elements `1-6 + 11-12` package sit mainly inside the Elements corpus, or does it form a fuzzy boundary with non-Elements practical geometry?

Data/files:

- `phase10_plus_solids_and_nonmetadata_practical_geometry.md`
- `derived_data/metadata_elements_plus_solids_vs_nonmetadata_practical_contrasts.csv`
- `derived_data/metadata_elements_plus_solids_neighboring_nonmetadata_cases.csv`
- `derived_data/metadata_elements_plus_solids_neighbor_bucket_summary.csv`
- `derived_data/metadata_elements_plus_solids_nearest_nonmetadata_practical_geometry.csv`
- `derived_data/metadata_elements_plus_solids_nearest_nonmetadata_euclid_language_not_metadata.csv`

Findings:

- `1-6 + 11-12` is much more method/institution/theory-heavy than ordinary non-metadata practical geometry.
- Against non-metadata practical geometry, it over-indexes procedural/pedagogical identity, method/demonstration/order, religious-institutional authority, Theoretical Mathematics, accessibility/ease, learned/scholarly framing, school framing, correction, and professor/lecturer authority.
- It is not much more utilitarian than non-metadata practical geometry; utility/practice/application is nearly equal.
- Non-metadata practical geometry is stronger in explicit Practical Geometry classification, architecture, perspective, visual/material aids, professional/practical arenas, and no-visible-social-arena cases.
- The closest non-metadata practical/military neighbors are disproportionately cases that invoke Euclid/Elements language, join theoretical and practical geometry, discuss planes and solids, or use institutional/professorial authority.

Interpretation:

`1-6 + 11-12` is a bridge form, but it bridges from the side of Euclidean authority. It does not become an ordinary trade manual. It makes Euclid usable by joining first principles, solid geometry, method, institutional authority, and selected practical publics.

Open questions:

- Which six or eight bridge cases should become the close-reading set?
- Does Dutch plain `1-6` form an earlier practical-vernacular bridge route?
- Do diagrams/apparatus/title-page visuality reinforce this bridge pattern?

---

### Method Note: Sparse-Canonical And Title-Page Fashion

Date: 2026-06-10

Issue:

The label sparse-canonical may confuse intellectual rhetoric with title-page fashion. Minimal title pages can result from canonical authority, but also from language, region, period, printer, schoolbook convention, genre, or extraction quality.

Action:

- Enriched `ANALYSIS_TERMS.md` with definitions for sparse-canonical, claim, audience, and archetype.
- Added explicit column maps for claim, audience, social evidence, and archetype fields.
- Added a rule: do not interpret quiet title pages as meaningful silence until comparing nearby books by period, language, place, and subject where possible.

Future checks:

- Test sparse/dense title-page rhetoric by `period`, `language_first`, `city`, and school/institution markers.
- Check whether German sixteenth-century titles are generally more elaborate.
- Check whether French or France-printed Latin title pages are generally more minimal.
- Check whether schoolbooks are systematically more sparse or more pedagogically explicit.

---

### Q015: How Do Natural Elements Modes Overlap With Each Other And Metadata?

Date: 2026-06-11

Question:

How do the Phase 7 natural modes overlap with each other, and how do they relate to metadata fields such as book coverage, additional content, format, language, period, city, and diagrams?

Data/files:

- `phase11_natural_modes_metadata_format_ecology.md`
- `derived_data/metadata_elements_natural_modes_matrix_with_format.csv`
- `derived_data/metadata_elements_mode_counts_with_format.csv`
- `derived_data/metadata_elements_natural_mode_overlap_pairs.csv`
- `derived_data/metadata_elements_natural_modes_by_metadata_fields_long.csv`
- `derived_data/metadata_elements_natural_modes_metadata_field_strong_signals.csv`
- `derived_data/metadata_elements_modes_vs_non_elements_feature_contrasts.csv`
- `derived_data/titlepage_density_fashion_association_diagnostics.csv`
- `derived_data/metadata_elements_author_editor_portfolios_preliminary.csv`

Findings:

- The natural modes are not separate subcorpora. They form overlapping bundles.
- The dense center of the metadata Elements corpus is humanist/ancient + institutional/authority + pedagogical/method + composite/apparatus.
- Practical/public is smaller, but it usually rides on top of pedagogical/method and canonical/humanist authority rather than standing outside the Elements tradition.
- Sparse/canonical is the most separate mode and must be treated carefully because it may reflect title-page fashion.
- Book coverage matters: `1-6 + 11-12` is the strongest large practical/public package; near-complete/expanded editions are especially composite/apparatus and pedagogical/method.
- Format matters: folios over-index composite/apparatus; quartos over-index vernacular/transfer; duodecimos are strongly institutional and fairly practical/public.
- Language and period matter: English and Dutch Elements are especially practical/public; German is strongly vernacular/transfer; practical/public rises over time; vernacular/transfer is especially strong in 1550-1599.
- Additional content is strongly tied to composite/apparatus, and Data is also tied to pedagogical/method.
- Compared with non-Elements works, Elements modes are distinctive because they bind method, pedagogy, utility, transfer, and apparatus to ancient/canonical authority.

Interpretation:

The metadata Elements corpus has a dense canonical-institutional-pedagogical-apparatus center and several edges. Practical/public Elements shows how Euclid becomes usable without ceasing to be canonical. Vernacular/transfer shows how Euclid travels across language and publics. Sparse/canonical may show reliance on canonical identity, but it must be controlled for period, place, language, format, and genre.

Open questions:

- Run a focused density/fashion-control analysis for sparse-canonical.
- Run an author/editor portfolio analysis comparing Elements and non-Elements works.
- Deepen the Dutch plain `1-6` route as a practical-vernacular pathway.
- Build close-reading sets for central-overlap cases and edge cases.

---

### Q016: Does Sparse-Canonical Survive Title-Page Fashion Controls?

Date: 2026-06-11

Question:

Can sparse-canonical Elements title pages be interpreted as meaningful canonical silence, or are they strongly shaped by title-page fashion: city, language, format, period, book coverage, and institution/school visibility?

Data/files:

- `phase12_sparse_canonical_fashion_controls.md`
- `derived_data/sparse_fashion_control_associations.csv`
- `derived_data/sparse_fashion_control_rates_long.csv`
- `derived_data/sparse_fashion_control_strong_groups.csv`
- `derived_data/sparse_fashion_control_combined_strata.csv`
- `derived_data/sparse_fashion_school_institution_comparison.csv`
- `derived_data/sparse_canonical_cases_with_fashion_controls.csv`
- `derived_data/quiet_non_sparse_cases_for_comparison.csv`

Findings:

- Sparse-canonical cannot be treated as a pure intellectual signal.
- City is the strongest control for the Elements sparse/canonical natural mode.
- Language is also meaningful; Italian-language cases are especially high in sparse/canonical behavior.
- Format matters, but less strongly than city/language.
- Period matters, but less than city/language.
- Explicit school/institution markers usually make title pages less sparse and less socially invisible in the current feature set.
- Bologna, Leipzig, Italian cases, pre-1550 Latin octavos, and Italian `books_1_6` in 1650-1699 are high fashion-risk zones.
- `1-6 + 11-12` remains a louder practical-pedagogical package; plain `1-6` is more sparse/canonical.

Interpretation:

Sparse-canonical is a real title-page posture, but it should be used as a contrast mode and close-reading category, not as a major thesis by itself. Some Elements title pages may rely on canonical identity with little explicit social/practical explanation, but we can only identify meaningful canonical silence after subtracting local title-page fashion.

Open questions:

- Close-read stronger sparse candidates against high-fashion-risk sparse cases.
- Run author/editor portfolio analysis to see whether sparse/dense rhetoric changes within the same author/editor across language, city, format, and genre.
- Continue the Dutch plain `1-6` route.

---

### Q017: Do Author / Editor Portfolios Change When They Enter The Elements?

Date: 2026-06-11

Question:

For named authors/editors who appear in both the metadata-defined Elements corpus and the broader mathematical corpus, do their title pages carry the same intellectual/social values, or does the Elements trigger a different rhetorical posture?

Data/files:

- `phase13_author_editor_portfolios.md`
- `scripts/build_author_editor_portfolios.py`
- `derived_data/author_editor_portfolio_case_matrix.csv`
- `derived_data/author_editor_portfolio_person_summary.csv`
- `derived_data/author_editor_portfolio_raw_combo_summary.csv`
- `derived_data/author_editor_portfolio_interesting_people.csv`
- `derived_data/author_editor_portfolio_elements_non_elements_pairs.csv`

Findings:

- This is corpus-internal evidence only. We do not have full bibliographies for every named person.
- 160 named people have at least one represented metadata-defined Elements case.
- 38 named people have both Elements and non-Elements cases.
- 13 bridge portfolios have at least two Elements and two non-Elements cases.
- In those 13 bridge portfolios, ancient authority/restoration is higher in Elements for all 13 people.
- Average Elements-minus-non-Elements deltas in bridge portfolios: ancient authority +86.3 pp, translation/transfer +35.2 pp, correction/revision +23.8 pp, method/order +19.4 pp, no visible social arena -30.3 pp, utility/practice -13.2 pp.
- Elements editions often make a person appear as restorer, translator, corrector, organizer, demonstrator, or mediator of an ancient/canonical corpus.
- Major bridge figures often move across city, language, and format, so fashion controls remain important.

Interpretation:

The Elements is not only a subject subset within the represented title-page corpus. In author/editor portfolios represented here, it often behaves as canon-work: a site where mathematical workers perform textual authority, correction, translation, augmentation, selection, method, and social relocation. This gives a stronger social-intellectual bridge than simply saying the Elements is theoretical or pedagogical, but it should not be converted into full-career claims without external bibliography.

Open questions:

- Close-read same-person clusters: Henrion, Forcadel, Dou, Clavius/Tacquet, Ozanam/Errard.
- Test whether same-city, same-language, same-format pairs preserve the aggregate portfolio pattern.
- Decide whether the Dutch Dou route is a separate practical-vernacular route.

---

### Q018: Do Same-Person Controlled Pairs Preserve The Canon-Work Pattern?

Date: 2026-06-11

Question:

If we restrict to same person, same city, same first language, and within five years, do Elements title pages still differ from nearby non-Elements title pages?

Data/files:

- `phase14_controlled_same_person_close_reading_selector.md`
- `scripts/build_controlled_portfolio_close_reading.py`
- `derived_data/controlled_author_editor_close_pairs.csv`
- `derived_data/controlled_author_editor_close_pair_shortlist.csv`
- `derived_data/controlled_author_editor_close_pair_summary.csv`

Findings:

- 89 controlled pairs across 13 people.
- 15 strict same-format pairs, all in the Forcadel cluster.
- Dense clusters: Forcadel and Henrion.
- Henrion strongly preserves the canon-work pattern: Elements adds ancient authority, translation, correction, augmentation; non-Elements more often add utility/practice.
- Forcadel preserves ancient authority and translation but complicates method/order, which appears stronger in nearby non-Elements works.
- Clavius is a clean institutional canon-work case.
- Ozanam and Errard connect practical/public worlds to Elements-specific canonical mediation.

Interpretation:

The controlled comparison preserves the main claim but narrows it: Elements often adds canonical/textual mediation, but the mixture varies by person and local setting. Utility/practice is often stronger in neighboring non-Elements works, and method/order is not always an Elements-side signal.

Open questions:

- Close-read Henrion, Forcadel, Clavius, Ozanam/Errard title pages manually.
- Decide whether Dou needs a separate Dutch practical-vernacular analysis, since it does not appear strongly in the five-year controlled pool.

---

### Q019: What Do The Controlled Close Readings Actually Show?

Date: 2026-06-11

Question:

When we read the title-page annotation fields for the best Phase 14 controlled clusters, what is the historical shape of the Elements/non-Elements contrast?

Data/files:

- `phase15_first_controlled_close_readings.md`
- `derived_data/controlled_author_editor_close_pair_shortlist.csv`
- `derived_data/title_page_analysis_matrix.csv`
- `derived_data/author_editor_portfolio_case_matrix.csv`

Findings:

- Henrion is the strongest dense case: Elements title pages foreground translation, revision, augmentation, extraction, and Euclid as canonical text; neighboring non-Elements works foreground military arithmetic, instruments, fortification, and practical procedure.
- Forcadel is the best complication and format-control case: Elements adds translation/commentary/institutional authority, but method/order is not necessarily stronger than in non-Elements works.
- Clavius is a clean institutional canon-work case: Jesuit identity is shared, while Elements activates edition history, demonstrations, scholia, augmentation, and ancient textual identity.
- Ozanam shows that Elements can be public/military without being the most utility-forward title page.
- Errard shows professional continuity: the same royal-engineer identity carries practical geometry/fortification and Elements textual mediation.
- Rudd is a vivid English single-pair illustration: Euclid contracted/demonstrated with Dee versus practical geometry for surveyors, engineers, and students.

Interpretation:

The Elements/non-Elements contrast is not theory versus practice. It is canonical mediation versus more direct practical/procedural application. Elements title pages repeatedly translate, comment, correct, augment, select, demonstrate, and socially relocate Euclid. Neighboring non-Elements works more directly advertise utility, instruments, military technique, surveying, arithmetic procedure, and professional application.

Open questions:

- Run the Dutch plain `1-6` route, especially Dou, to see whether it forms a distinct practical-vernacular pathway.
- Choose 2-3 close-reading clusters for the conference argument.

---

### Q020: Is Dutch Plain `1-6` A Distinct Practical-Vernacular Route?

Date: 2026-06-11

Question:

Do Dutch plain `1-6` Elements editions, especially Dou, form a practical-vernacular route distinct from the later `1-6 + 11-12` practical-pedagogical package?

Data/files:

- `phase16_dutch_plain_1_6_practical_vernacular_route.md`
- `scripts/build_dutch_plain_1_6_route.py`
- `derived_data/dutch_plain_1_6_group_profiles.csv`
- `derived_data/dutch_plain_1_6_cases.csv`
- `derived_data/dutch_non_elements_practical_neighbors_1580_1700.csv`

Findings:

- Dutch plain `1-6`: 15 cases, heavily dominated by Jan Pietersz Dou.
- Dutch plain `1-6` practical/public mode: 60.0%, almost the same as `1-6 + 11-12` at 59.0%.
- Dutch plain `1-6` is much more practical/public than non-Dutch plain `1-6` at 15.2%.
- Dutch plain `1-6` has very high method/demonstration/order claims: 86.7%.
- It also over-indexes correction/revision, augmentation, professional/practical arena, and general public/lovers arena.
- Dou repeatedly combines Dutch translation, explanation, correction, augmentation, added utilities extracted from the books, species in geometrical figures, public lovers of the free art, and land-surveyor/wine-gauger authority.

Interpretation:

Dutch plain `1-6` is a distinct practical-vernacular Euclid route. It makes the first six books usable without adding books 11-12. It does so through translation, explanation, correction, added utilities, geometrical operations, and civic/professional land-measuring authority.

Open questions:

- Decide whether Dou should be a main slide case.
- Compare Dou with the "new/reorganized Elements" question: Dou operationalizes Euclid, but does not seem to claim a new Elements in the Arnauld `Paris_1667` sense.

---

### Q021: New / Reorganized Elements As Anti-Philological Pole

Date: 2026-06-11

Question:

When, where, and how often do title pages claim something like "new Elements," "new order," "new demonstrations," or a reorganized/reconstructed Elements? Is this a pole opposite to fidelity-to-the-ancients or philological-restoration ideals?

Seed case:

- `Paris_1667`, Antoine Arnauld, `Nouveaux Elemens de Geometrie`.

Initial observation:

`Paris_1667` is tagged as pedagogical/method and novelty/modernity, not ancient-restoration. Its title page emphasizes:

- new Elements;
- a wholly new order;
- new demonstrations;
- new means for showing incommensurable lines;
- new measures of angles;
- new ways of finding and demonstrating line proportions.

Why it matters:

This may be a counter-pole to editions that present themselves through fidelity, restoration, ancient authority, philology, or corrected transmission. It could reveal a more interventionist pedagogical ideal: the Elements as a structure to be remade rather than a text to be preserved.

Status:

Answered provisionally in `phase17_new_reorganized_elements.md`.

Data/files:

- `phase17_new_reorganized_elements.md`
- `scripts/build_new_reorganized_elements.py`
- `derived_data/new_reorganized_elements_cases.csv`
- `derived_data/new_reorganized_elements_summary.csv`

Findings:

- Broad finder returns 80 candidates.
- Strict strong "new Elements" cases are rare: `Paris_1667` and `Livorno_1709`.
- New order/new demonstration/reorganized cases are broader and include Dechales/Reeve Williams, Naboth, Le Tenneur, Reyher, Burckhard von Pirckenstein, König/Kuypers, and others.
- Most reconstructive rhetoric is hybrid: new/easy method, contraction, selection, or new demonstration often coexists with correction, augmentation, translation, or ancient/Euclidean authority.
- The best model is a spectrum: philological-restorative, hybrid pedagogical mediation, reconstructive.

Open questions:

- Compare reconstructive cases directly against restoration/fidelity cases.
- Follow the user's next question about deductive/mathematical parts on title pages: propositions, demonstrations, axioms, scholia, diagrams, figures, theorems, corollaries, notes.

---

### Q022: Which Deductive Or Mathematical Parts Are Named On Title Pages?

Date: 2026-06-11

Question:

Which parts of mathematical/deductive content are named or highlighted on title pages: axioms, propositions, theorems, demonstrations, diagrams/figures, scholia, corollaries, notes, paradoxes, definitions, postulates, lemmas, enunciations? How are they valued or framed?

Why it matters:

This can show what title pages think mathematical knowledge is made of. It also connects to the new/reorganized Elements question, because reconstruction often happens by manipulating parts: new demonstrations, selected propositions, reduced proposition counts, scholia, corollaries, notes, diagrams, algebraic signs.

Status:

Answered in `phase20_deductive_parts_named_on_title_pages.md`.

Data/files:

- `phase20_deductive_parts_named_on_title_pages.md`
- `scripts/build_deductive_parts_analysis.py`
- `derived_data/deductive_parts_summary.csv`
- `derived_data/deductive_parts_cases.csv`
- `derived_data/deductive_parts_combinations.csv`
- `derived_data/deductive_parts_pairs.csv`
- `derived_data/deductive_parts_by_strata.csv`
- `derived_data/deductive_parts_interesting_cases.csv`

Findings:

- Named deductive or mathematical parts appear in 359/843 representative rows, 42.6%.
- Metadata Elements representatives name such parts much more often: 165/286, 57.7%, compared with 194/557, 34.8% for non-Elements.
- The strongest Elements-associated parts are demonstrations/proofs, scholia/commentary, principles, propositions, theorems, and enunciations.
- Non-Elements title pages more often foreground problems, operations/constructions, examples, and notes/observations; figures/diagrams occur at almost the same rate in Elements and non-Elements.
- Demonstration/proof language rises strongly in Elements title pages after 1600, especially 1650-1699.
- Scholia/commentary are especially visible earlier and in institutional/commentarial modes.
- Rare foundational parts such as axioms, postulates, lemmas, and definitions are rarely advertised. Reconstructive Euclid is therefore usually framed around order, demonstrations, propositions, method, and use, not around an axiomatic skeleton.

Interpretation:

The Elements is advertised as a canonical demonstrative-commentarial corpus. Neighboring mathematical books more often advertise procedural, operational, visual, and problem-solving units. This clarifies the Elements' place in the ecology: it is not isolated from practical or pedagogical mathematics, but it enters those worlds through transformations of its parts, especially propositions, demonstrations, figures, scholia, and uses.

Open questions:

- Deepen proposition-use and demonstration-use by book group and author/editor routes.
- Separate commentary as ancient/humanist apparatus from commentary as pedagogical explanation.
- Analyze diagrams/figures against image or metadata evidence, because figure language is not uniquely Elements.

---

### Q023: Is Reconstructive Euclid Another Kind Of Restoration?

Date: 2026-06-11

Question:

Is the reconstructive pole really opposed to ancient/philological restoration, or does it restore Euclid differently: by recovering the logic, method, order, demonstrative force, or pedagogical intelligibility behind the Elements?

Data/files:

- `phase18_reconstructive_restoration_refinement.md`
- `derived_data/new_reorganized_elements_cases.csv`
- `derived_data/new_reorganized_elements_summary.csv`

Findings:

- The better model is competing restoration ideals, not simple opposition.
- Strict reconstructive cases are mostly later: `Paris_1667` in 1667 and `Livorno_1709` in 1709.
- Broader reordered/new-demonstration cases cluster strongly in 1650-1699.
- Reconstructive rhetoric is not reducible to vernacularization: the strict cases are French and Latin, while the broader field includes English, German, French, Italian, and Latin.
- It is more consistently connected to pedagogy/method than to practical application, though practical/technical audiences appear in cases such as `Vienna_1694`, `London_1680-81`, and Dechales/Reeve Williams.
- The core idea may be "logical restoration" or "pedagogical restoration": restoring Euclid's demonstrative order and force rather than the ancient text.

Interpretation:

The reconstructive pole is not simply anti-philological. It offers a rival restoration ideal: Euclid is restored not by recovering ancient wording, but by recovering the demonstrative order, clarity, and pedagogical force that the Elements ought to have.

Open questions:

- Test this directly against the future deductive-parts analysis.
- Look for title-page language that explicitly contrasts old/ordinary order with new/better logical order.

---

### Q024: How Should We Characterize The Restoration / Reconstruction Clusters?

Date: 2026-06-11

Question:

How should we characterize the clusters around philological restoration, logical restoration, pedagogical mediation, symbolic retooling, practical refunctionalization, and selection/contraction?

Data/files:

- `phase19_restoration_reconstruction_cluster_characterization.md`
- `derived_data/new_reorganized_elements_cases.csv`
- `derived_data/new_reorganized_elements_summary.csv`

Findings:

- The clusters are competing answers to: what has to be restored in Euclid for the Elements to work?
- Six working clusters: philological/ancient-text restoration; hybrid corrective-pedagogical mediation; logical/demonstrative restoration; symbolic/analytic retooling; practical/technical refunctionalization; selection/contraction/portable Euclid.
- Strong reconstructive/logical restoration is mostly later, especially visible after the mid-seventeenth century.
- It is partly vernacular or translingual, but not reducible to vernacularization.
- It is more consistently pedagogical/methodological than practical, though practical/technical audiences are important in several routes.

Interpretation:

The reconstructive cluster is not a rejection of Euclid's authority. It is a claim to deeper Euclidean fidelity: fidelity to order, proof, and intelligibility rather than inherited textual form.

Open questions:

- Use Q022/Phase 20 to refine which restoration clusters are actually supported by deductive-part evidence.
- Decide which restoration clusters should survive into the final report.
