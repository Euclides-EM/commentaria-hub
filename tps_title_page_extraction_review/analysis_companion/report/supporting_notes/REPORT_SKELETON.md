# The Place of Euclid's Elements in Early Modern Mathematical Print: A Title-Page Ecology

Working report skeleton.

Leading question:

**What place did the metadata-defined Elements corpus occupy in the broader ecology of early modern mathematical print, and how did title pages construct that place through social address, intellectual values, and acts of textual/mathematical mediation?**

Report infrastructure:

- `REPORT_DRAFT.md`
- `REPORT_APPENDICES.md`
- `REPORT_INFRASTRUCTURE_NOTES.md`
- `REPORT_PROP_DEMO_NOTES.md`
- `REPORT_COMMENTARY_SPLIT_NOTES.md`
- `REPORT_BRIDGE_CASE_NOTES.md`
- `REPORT_FIGURES_DIAGRAMS_NOTES.md`
- `REPORT_CASEBOOK_SHORTLIST.md`
- `tables/`
- `figures/`
- `scripts/build_report_infrastructure.py`
- `scripts/build_prop_demo_deep_dive.py`
- `scripts/build_commentary_split.py`
- `scripts/build_bridge_case_table.py`
- `scripts/build_figures_diagrams_deep_dive.py`

## Report Principle

This report is not a chronology of our exploration. It is an integrated historical analysis. Exploratory paths, corrections, and dead ends should not appear in the main body unless they produce a substantive historical result. Technical details belong in appendices.

The broad mathematical title-page corpus is used as ecology and comparison. The metadata-defined Elements corpus is the central object.

Write for a reader opening the report cold months later. Do not assume memory of this conversation or of internal analysis labels. When a label is useful, translate it immediately into human language: for example, say "apparatus-rich learned Euclid" rather than relying on `euclid_composite_workshop`, and say "usable or pedagogical Euclid" rather than relying on `usable_elements`. Tables may preserve internal column names, but the main prose should explain what each pattern means historically and why the example is being used.

## 0. Executive Argument

Purpose:

State the central answer in compact form.

Core claim to defend:

The metadata-defined Elements corpus occupied a distinctive but flexible place in early modern mathematical print. Title pages present it not simply as a stable ancient text or as one branch of geometry, but as a canonical demonstrative corpus repeatedly recomposed for different publics, institutions, pedagogies, and intellectual ideals.

The report should argue that the Elements becomes visible through several forms of work:

- restoring or correcting ancient authority;
- translating and vernacularizing;
- commenting and explaining;
- demonstrating and re-demonstrating;
- selecting, contracting, and repackaging;
- augmenting with apparatus, figures, additional texts, or practical uses;
- reorganizing or reforming the sequence of mathematical knowledge.

Consequences, not starting points:

- The old impression of stable Elements designation can be revisited near the end as a consequence of the new analysis, not as the report's organizing problem.
- Euclid/Elements title-page language broader than metadata Elements is useful as a comparison, but the central object remains metadata-defined Elements editions.

Needed before writing:

- One clean "report core findings" table. First-pass supporting tables now exist in `tables/`; choose a small subset before drafting prose.

## 1. Corpus, Object, And Reading Strategy

Purpose:

Define what the report studies and how evidence is read.

Central question:

What is the corpus, what counts as an Elements edition for this report, and what can title pages tell us?

Main points:

1. The broad corpus reconstructs an ecology of mathematical print.
2. The metadata-defined Elements corpus is the home corpus.
3. Title pages are evidence for public framing: identity, authority, audience, intellectual value, and advertised acts performed on mathematical knowledge.
4. The report distinguishes between:
   - advertised title-page claims;
   - inferred social/intellectual positioning;
   - actual book contents, which require separate internal reading.

Claims to avoid:

- Do not organize this section around old mistakes, old claims, or external lineage taxonomies.
- Do not over-center Euclid/Elements wording as if title-page designation defines the corpus.

Main evidence:

- Corpus accounting: representatives, metadata Elements subset, title-page matrix.
- Short description of classification and feature families.

Appendix material:

- Full data construction.
- Reprint/representative logic.
- Classification run details.
- Feature/tag taxonomy.

Likely visual:

- Corpus accounting flowchart.

Available table:

- `tables/report_corpus_accounting.csv`

## 2. The Surrounding Mathematical Print Ecology

Purpose:

Establish the surrounding ecology so the Elements can later be located within it.

Central question:

What kinds of mathematical books surround the Elements, and how do title pages present their social worlds, intellectual values, and advertised forms of mathematical work?

This section needs concrete trends, not only a high-level subject map.

Proposed subsections:

### 2.1 Subject Zones And Overlaps

Claim:

The surrounding corpus is organized by overlapping work-zones rather than a simple pure/applied hierarchy.

Taxonomy to defend:

- Geometry/Theory;
- Practical Geometry / Surveying;
- Instruments/Measurement;
- Arithmetic/Commerce;
- Cosmos/Earth;
- Applied Mechanics/Military;
- Visual/Spatial Arts;
- Music.

What to show:

- Subject-family counts.
- Pairwise overlaps.
- Books with no clean primary family but strong Euclidean or mathematical identity.

Likely visual:

- Heatmap of subject-family co-occurrence.
- Possibly PCA/UMAP plot using subject + title-page feature vectors, colored by subject family and marked for metadata Elements.

Available tables/figures:

- `tables/report_subject_zone_counts.csv`
- `tables/report_subject_social_rates_matrix.csv`
- `tables/report_subject_intellectual_rates_matrix.csv`
- `figures/heatmap_subject_social_rates.png`
- `figures/heatmap_subject_intellectual_rates.png`
- `figures/pca_full_corpus_report_features.png`

### 2.2 Social Grammars By Subject Zone

Claim:

Different mathematical zones advertise different forms of social legitimacy.

Expected trends to test/defend:

- Instruments/measurement and applied/military books should over-index explicit users, offices, professions, and practical publics.
- Geometry/theory and Elements-adjacent books should over-index authorship, textual identity, ancient authority, institutions, and editor credentials.
- Arithmetic/commercial books should connect practical utility to merchants, reckoning, teaching, examples, and procedural clarity.
- Cosmos/earth and composite learned books may over-index additions, bound-with materials, tables, figures, and learned apparatus.

Needed analysis:

- Social category rates by subject family.
- Social authority type by subject family: audience vs institution vs editor/office vs patronage.

Likely visuals:

- Heatmap: subject families x social markers.
- PCA plot: social markers only, colored by subject and Elements status.

### 2.3 Intellectual Values By Subject Zone

Claim:

Subject zones differ not only in content, but in the virtues their title pages promote.

Expected trends to test/defend:

- Practical/military/instrumental zones: utility, application, construction, operation, use, profession.
- Geometry/theory and Elements-adjacent zones: demonstration, method, correction, translation, ancient authority, commentary.
- Arithmetic/commercial zones: rule, example, ease, practice, calculation.
- Learned composite zones: completeness, enrichment, additions, tables, apparatus.

Needed analysis:

- Intellectual-value heatmap by subject family.
- Strongest positive/negative contrasts.

Likely visuals:

- Heatmap: subject family x intellectual value.
- Clustered heatmap if patterns are clean.

### 2.4 Title Pages As Acts, Not Labels

Claim:

Across the corpus, mathematical title pages often advertise acts performed on knowledge: translating, correcting, augmenting, demonstrating, extracting, reducing, explaining, furnishing, and adapting.

Why this matters:

This prepares the central Elements argument: the Elements is not merely named; it is worked on.

Needed analysis:

- Action verb taxonomy by subject family and period.
- Co-occurrence of action verbs with social markers.

Possible visual:

- Heatmap or bipartite map: action categories x subject families.

## 3. Locating The Elements Within The Ecology

Purpose:

Make the first direct Elements-versus-ecology comparison.

Central question:

What is distinctive about metadata-defined Elements title pages compared with surrounding mathematical books?

Core claim:

The Elements is not distinctive simply because it is theoretical. It is distinctive because title pages bind method, pedagogy, correction, translation, apparatus, and sometimes utility to canonical ancient authority.

Subclaims to defend:

1. Elements title pages over-index canonical/textual mediation: ancient authority, correction, translation, commentary, method, demonstration.
2. Non-Elements neighbors more directly advertise procedure, operation, instruments, professional practice, problem-solving, and application.
3. Elements title pages are not socially empty; they often relocate Euclid into institutions, schools, religious orders, vernacular publics, or practical settings.
4. The contrast is best described as canonical mediation versus more direct practical/procedural application.

Main evidence:

- Elements vs non-Elements feature contrast table.
- Controlled same-person close readings.
- Deductive-parts contrast from Phase 20.

Needed analysis:

- Final robust Elements/non-Elements contrast table with selected metrics only.
- Possibly logistic or stratified checks by period/language/format for the most important metrics.
- Bridge-case table for Elements/non-Elements gradients.

Current result:

- See `REPORT_BRIDGE_CASE_NOTES.md`.
- The Elements has a firm corpus identity but porous functional boundaries.
- "Usable Elements" is usually not anti-canonical; 17 of 31 usable Elements cases also meet the canonical Elements route.
- The surrounding ecology has a parallel bridge: 43 cases overlap between Euclidean practical geometry and professional/material practical arts.
- This supports a gradient from Euclid as restored text, to Euclid as teachable/usable corpus, to Euclid as practical-geometrical authority, to mathematics as visual/material/professional operation.

Likely visuals:

- Heatmap: Elements vs non-Elements x selected claim/social/deductive markers.
- PCA plot of full corpus feature vectors, highlighting metadata Elements, non-Elements practical geometry, and Euclid-language-but-not-metadata cases if useful.

Available tables/figures:

- `tables/report_elements_vs_non_elements_core_contrasts.csv`
- `figures/bar_elements_vs_non_elements_top_contrasts.png`
- `figures/heatmap_deductive_parts_by_corpus.png`
- `tables/report_bridge_case_route_marker_rates_matrix.csv`
- `tables/report_bridge_case_top_cases.csv`
- `figures/heatmap_bridge_route_marker_rates.png`

## 4. The Internal Ecology Of The Elements Corpus

Purpose:

Answer whether the Elements corpus divides into clear subcorpora.

Central question:

Is the metadata-defined Elements corpus internally divided into stable subcorpora, or does it consist of overlapping modes?

Core claim:

The Elements corpus does not divide cleanly. It forms overlapping modes with a dense center and flexible edges.

Taxonomy to defend:

- canonical/ancient authority;
- institutional authority;
- pedagogical/method;
- composite/apparatus;
- vernacular/transfer;
- practical/public;
- sparse/canonical as a controlled posture, not a primary explanation.

Subclaims:

1. The dense center is ancient/canonical + institutional + pedagogical/method + apparatus.
2. Practical/public Elements usually overlaps with pedagogy/method and does not abandon canonical authority.
3. Vernacular/transfer is not merely a language fact; it is often an access and mediation strategy.
4. Sparse/canonical title pages exist, but are entangled with title-page fashion, format, language, city, and period.

Needed analysis:

- Final mode overlap table.
- Mode rates by language, period, format, and book group.
- Decide whether PCA/cluster plot of Elements-only features adds explanatory value.

Likely visuals:

- Mode co-occurrence heatmap.
- Elements-only PCA plot colored by dominant mode and shaped by format/book group.

Available tables/figures:

- `tables/report_elements_mode_marker_rates_matrix.csv`
- `figures/heatmap_elements_mode_marker_rates.png`
- `figures/pca_elements_only_report_features.png`

## 5. Book Coverage As Historical Repackaging

Purpose:

Show that which books of the Elements are included can be a social/intellectual signal.

Central question:

How do different Elements book packages behave on title pages?

Core claim:

Different partial Elements editions perform different work. Selection is not merely omission; it is historical repackaging.

Subsections:

### 5.1 Plain Books 1-6

Likely claim:

Plain `1-6` often presents elementary Euclid: foundations, translation, vernacular access, and sometimes sparse/canonical or practical-vernacular framing.

### 5.2 Books 1-6 + 11-12

Likely claim:

`1-6 + 11-12` is a later and more sharply usable package: elementary plane geometry plus solid geometry, often suited to institutional, military, academy, or practical-pedagogical contexts.

### 5.3 Dutch/Dou Plain 1-6 Route

Likely claim:

Dutch plain `1-6`, especially Dou, forms a practical-vernacular route without books 11-12: civic measuring, public lovers of mathematics, geometrical operations, and land-surveyor/wine-gauger authority.

### 5.4 Near-Complete And Expanded Elements

Likely claim:

Near-complete/expanded editions often sustain the Elements as a learned, institutional, commentarial, and apparatus-rich corpus.

Needed analysis:

- Proposition-use and demonstration-use by book group.
- Check how much `1-6 + 11-12` depends on Dechales/Reeve Williams/Tacquet clusters.
- Final concise table of book groups x social/intellectual markers.

Current result:

- See `REPORT_PROP_DEMO_NOTES.md`.
- Explicit "use of each/every proposition" rhetoric is heavily Dechales/Reeve/Williams and should not be generalized to all `1-6 + 11-12`.
- The broader method/demonstration orientation of `1-6 + 11-12` survives outside that cluster.
- Near-complete/expanded editions are more demonstration-heavy than proposition-use heavy.

Likely visuals:

- Heatmap: book group x selected features.
- Timeline by book group and language.

Available tables/figures:

- `tables/report_elements_bookgroup_marker_rates_matrix.csv`
- `figures/heatmap_elements_bookgroup_marker_rates.png`
- `tables/report_elements_period_marker_rates_long.csv`
- `tables/report_elements_language_first_marker_rates_long.csv`

## 6. Social Worlds Of The Elements

Purpose:

Give the social analysis its own serious section.

Central question:

Which social worlds do Elements title pages invoke, and how do those social worlds shape the presentation of Euclidean mathematics?

Taxonomy to defend:

- learned/humanist scholarly authority;
- university/academy/school settings;
- Jesuit/religious institutional settings;
- court/state patronage and service;
- military/technical publics;
- surveyors/geometers/engineers;
- vernacular public/lovers/readers;
- editor/office/professional credential as authority.

Subclaims:

1. Social address is not one thing: audience, institution, office/credential, and patronage must be separated.
2. Elements editions often authorize Euclid through institutions and editors rather than through explicit occupational audiences.
3. Practical Elements routes appear when institutional or professional authority is joined to utility, method, figures, or uses of propositions.
4. Patronage/prestige often frames authority but should not be mistaken for readership.

Needed analysis:

- Compact authority taxonomy table.
- Social markers by Elements mode/book group.
- Close-reading case clusters: Dou, Clavius/Tacquet, Henrion/Forcadel, Errard/Rudd/Ozanam.

Likely visuals:

- Heatmap: Elements modes/book groups x social authority types.
- Maybe network diagram: social authority types connected to intellectual values.

## 7. Intellectual Values Of The Elements

Purpose:

Give the intellectual/pedagogical analysis its own section.

Central question:

What ideals of mathematical knowledge do Elements title pages promote or suppress?

Taxonomy to defend:

- ancient authority/restoration;
- correction/revision;
- translation/vernacular transfer;
- commentary/explanation;
- demonstration/method/order;
- ease/clarity/pedagogy;
- utility/application;
- augmentation/apparatus;
- selection/contraction/portability;
- novelty/reorganization/reform.

Subclaims:

1. Elements title pages value Euclid as something to be mediated, not merely named.
2. Correction, translation, commentary, and demonstration are not separate decorations; together they make ancient authority usable.
3. Utility in Elements title pages is often mediated through proposition-use, method, institutions, or book coverage rather than direct professional procedure.
4. Some values are suppressed or rare: axioms/postulates/definitions are not usually title-page selling points.

Needed analysis:

- Clean examples for each intellectual value.

Current result:

- See `REPORT_COMMENTARY_SPLIT_NOTES.md`.
- Commentary/explanation is an Elements-specific signal, but it divides into historically different functions.
- Ancient/humanist scholia cluster especially in Greek/Latin and sixteenth-century learned routes.
- Pedagogical explanation becomes especially visible in vernacular and practical-pedagogical routes after 1650.
- Clavius/Jesuit commentary is best treated as institutional apparatus, not generic explanation.

Likely visuals:

- Heatmap: Elements modes/book groups x intellectual values.

## 8. The Mathematical Parts Of Euclid On Title Pages

Purpose:

Integrate deductive-parts analysis.

Central question:

Which mathematical parts do title pages foreground, and what does that reveal about the advertised identity of the Elements?

Core claim:

The Elements is advertised as a canonical demonstrative-commentarial corpus. Its distinctive title-page parts are demonstrations/proofs, propositions, scholia/commentary, principles, theorems, and enunciations.

Subclaims:

1. The broader non-Elements ecology more often foregrounds problems, operations/constructions, examples, and notes/observations.
2. Figures/diagrams are not uniquely Elements; they belong broadly to mathematical print.
3. Reconstructive or pedagogical Euclid usually works through propositions, demonstrations, order, method, and use, not through advertised axioms/postulates.
4. The proposition is especially important because it can be a proof unit, teaching unit, portable extract, or usable/application unit.
5. Figures become distinctive only when joined to other acts: proof, pedagogy, practical operation, edition furnishing, or ancient/learned apparatus.

Needed analysis:

- Proposition-use and demonstration-use by book group.
- Possibly close read `Lausanne_1683`, `London_1685a`, `Paris_1640`, `Cologne_1556`, `Paris_1667`.

Current result:

- See `REPORT_PROP_DEMO_NOTES.md`.
- Demonstration language is broad within the Elements corpus and rises after 1600.
- Proposition-use is narrower and historically clustered.
- Proposition selection/reduction/order cases are rare but analytically important.
- See also `REPORT_COMMENTARY_SPLIT_NOTES.md`: scholia/commentary is not one category, but splits between ancient/humanist apparatus, institutional commentary, pedagogical explanation, notes/annotations, and contracted/extracted commentary.
- See also `REPORT_FIGURES_DIAGRAMS_NOTES.md`: visual title-page claims appear at almost the same rate in metadata Elements and non-Elements, but their functions differ. Elements figures matter when attached to proof, apparatus, pedagogy, or practical Euclidean operations.

Likely visuals:

- Heatmap: part categories x Elements/non-Elements/book group.
- Timeline: demonstrations vs scholia/commentary in Elements.

Available tables/figures:

- `tables/report_deductive_parts_by_corpus_matrix.csv`
- `tables/report_deductive_parts_by_bookgroup_matrix.csv`
- `tables/report_figures_diagrams_by_corpus_matrix.csv`
- `tables/report_figures_diagrams_by_elements_bookgroup_matrix.csv`
- `tables/report_figures_diagrams_by_elements_mode_matrix.csv`
- `figures/heatmap_deductive_parts_by_corpus.png`
- `figures/heatmap_deductive_parts_by_bookgroup.png`
- `figures/heatmap_figures_diagrams_by_corpus.png`
- `figures/heatmap_figures_diagrams_by_elements_bookgroup.png`

## 9. Restoration, Reconstruction, And Competing Euclids

Purpose:

Explain what happens when title pages restore, reform, reorder, or remake Euclid.

Central question:

When title pages claim new order, new method, new demonstrations, reform, or reorganization, are they rejecting ancient Euclid, or restoring him differently?

Core claim:

Reconstructive Euclid is not simply anti-philological. It is a rival restoration ideal: restoring Euclid's demonstrative function, order, clarity, teachability, and force rather than only ancient wording.

Taxonomy to defend:

- philological/ancient-text restoration;
- corrective-pedagogical mediation;
- logical/demonstrative restoration;
- symbolic/analytic retooling;
- practical/technical refunctionalization;
- selection/contraction/portable Euclid.

Subclaims:

1. Strict "new/reformed Elements" cases are rare but important.
2. The broader reconstructive field is hybrid: new method, selected propositions, contraction, and retooling often coexist with correction, translation, augmentation, or ancient authority.
3. Strong reconstructive rhetoric becomes especially visible from the mid-seventeenth century onward.
4. Reconstructive cases are not reducible to vernacularization or practicality; pedagogy and method are more consistent.

Needed analysis:

- Close-read strict and hybrid cases.
- Check chronology/language of reconstructive clusters in final table.

Likely visuals:

- Timeline of reconstructive motifs.
- Cluster table: restoration ideal x cases x social/intellectual markers.

## 10. Boundaries And Gradients

Purpose:

Return from internal Elements analysis to the surrounding ecology.

Central question:

Where does the Elements meet, overlap with, or differ from neighboring mathematical genres?

Core claim:

The Elements/non-Elements boundary is a gradient rather than a wall:

canonical Euclid -> usable Euclid -> Euclidean practical geometry -> professional/material practical arts.

Subclaims:

1. `1-6 + 11-12` bridges toward practical geometry, but from the side of Euclidean authority.
2. Dutch/Dou plain `1-6` bridges toward civic practical geometry through vernacular operation and public mathematical users.
3. Some non-Elements works are Euclidean or elements-like without being metadata Elements editions.
4. The most interesting boundary cases are hybrids, not classification errors.

Needed analysis:

- Clean bridge-case table.
- Possibly nearest-neighbor examples already generated in Phase 10.

Likely visuals:

- Gradient diagram.
- Table of bridge cases by route.

## 11. Silence, Density, And Title-Page Fashion

Purpose:

Prevent overinterpretation.

Central question:

When Elements title pages are sparse or silent, is that meaningful?

Core claim:

Sparse-canonical presentation is real, but it is entangled with period, city, language, format, and title-page fashion. It should be used as a controlled contrast, not as a major thesis.

Subclaims:

1. Silence does not mean lack of audience, use, or pedagogical value.
2. Dense promotional title pages and sparse canonical title pages are both historically meaningful only when compared with local conventions.
3. Format matters: folio/quarto/octavo/duodecimo can affect density and the kind of claims made.

Needed analysis:

- No major new analysis unless choosing close-reading pairs.

Likely visual:

- Small control table only; detailed diagnostics in appendix.

## 12. Final Synthesis

Purpose:

Answer the leading question directly.

Final claim to develop:

The Elements occupied a special place in early modern mathematical print because it could carry ancient demonstrative authority across many social worlds while remaining open to recomposition. Title pages show this recomposition. They present the Elements as restored, corrected, translated, commented, demonstrated, selected, contracted, augmented, applied, and reorganized.

The Elements was therefore not a stable monument outside mathematical practice. It was a canonical proof-corpus whose authority made it especially available for mediation.

Possible final paragraph direction:

The title-page ecology shows not one Euclid, but a set of historically situated Euclids: ancient text, school method, institutional canon, practical geometry, portable curriculum, apparatus-rich scholarly object, and logical reconstruction.

## Possible "Against Earlier Assumptions" Section

Placement:

Either after the synthesis or as a short methodological reflection before appendices.

Purpose:

Compare the final findings with assumptions from the earlier, smaller analysis without making that comparison the report's central aim.

Possible points:

- Elements designation is less stable than it first appeared.
- The metadata Elements corpus must be distinguished from title-page Euclid/Elements language.
- Base designation does not capture the corpus's real historical mobility.
- The main story is not consistency of naming, but recomposition of canonical authority.

## Appendices

### Appendix A: Data And Corpus Construction

- source files;
- representative/reprint logic;
- metadata-defined Elements definition;
- classification run;
- title-page feature matrix.

### Appendix B: Taxonomies

- subject families;
- social authority categories;
- intellectual value categories;
- deductive-part categories;
- natural Elements modes.

### Appendix C: Core Tables

- corpus counts;
- Elements vs non-Elements contrasts;
- subject-family heatmaps;
- mode overlap tables;
- book group contrasts;
- deductive-part counts;
- format/language/period controls.

### Appendix D: Visualizations

Potential visual set:

- subject co-occurrence heatmap;
- subject x social marker heatmap;
- subject x intellectual value heatmap;
- Elements-only mode overlap heatmap;
- book group x feature heatmap;
- deductive parts x corpus/book group heatmap;
- PCA plot of full corpus title-page features;
- PCA plot of Elements-only features.

Use PCA only if interpretable. It should complement, not replace, tables.

### Appendix E: Casebook

Short entries for major cases and what each proves.

Current result:

- See `REPORT_CASEBOOK_SHORTLIST.md`.
- The casebook is organized by report section and assigns each case one main argumentative job.
- Use the human-friendly label translations in that file when drafting prose.

### Appendix F: Methodological Caveats

- title-page evidence versus book content;
- heuristic tag limits;
- incomplete author/editor bibliographies;
- OCR/transcription as minor but real noise;
- title-page fashion by period/place/language/format;
- small category counts.

## Report-Driven Analyses Still Needed

These are not open-ended exploration tasks. They exist only because the report structure needs them.

1. Build final report contrast tables:
   - Elements vs non-Elements;
   - subject family x social marker;
   - subject family x intellectual value;
   - Elements mode/book group x selected markers.

2. Build visualization prototypes:
   - heatmaps first;
   - PCA plots only if they show interpretable clusters or gradients.

3. Continue drafting the report section by section:
   - Sections 0-3 now have a first draft in `REPORT_DRAFT.md`;
   - next draft Sections 4-8;
   - keep technical tables in appendices;
   - explain internal labels in human language before using them.
