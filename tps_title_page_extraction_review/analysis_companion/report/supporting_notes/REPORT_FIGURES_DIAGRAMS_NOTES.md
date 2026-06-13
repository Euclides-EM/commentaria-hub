# Figures, Diagrams, And Visual Mathematical Work

Date: 2026-06-12

Purpose:

This note separates title-page figure/diagram language into different historical functions. The aim is to avoid treating "figures" as automatically Euclidean, automatically practical, or automatically pedagogical.

Script:

- `report/scripts/build_figures_diagrams_deep_dive.py`

Main outputs:

- `tables/report_figures_diagrams_by_corpus_matrix.csv`
- `tables/report_figures_diagrams_by_elements_bookgroup_matrix.csv`
- `tables/report_figures_diagrams_by_elements_mode_matrix.csv`
- `tables/report_figures_diagrams_by_format_matrix.csv`
- `tables/report_figures_diagrams_by_period_matrix.csv`
- `tables/report_figures_diagrams_by_language_matrix.csv`
- `tables/report_figures_diagrams_cases.csv`
- `tables/report_figures_diagrams_close_reading_shortlist.csv`
- `figures/heatmap_figures_diagrams_by_corpus.png`
- `figures/heatmap_figures_diagrams_by_elements_bookgroup.png`
- `figures/heatmap_figures_diagrams_by_elements_mode.png`
- `figures/heatmap_figures_diagrams_by_format.png`
- `figures/heatmap_figures_diagrams_by_subject.png`

## Headline Result

Figures and diagrams are not an Elements-specific title-page signal. Visual title-page claims appear at almost the same rate in metadata Elements and non-Elements:

| Corpus | Visual Title Claim | Metadata Diagrams | Visual Claim But No Metadata Diagram Tag | Metadata Diagram Tag But No Visual Claim |
|---|---:|---:|---:|---:|
| metadata Elements | 17.8% | 3.1% | 17.5% | 2.8% |
| non-Elements | 16.7% | 24.8% | 12.0% | 20.1% |

Interpretation:

The Elements does not own visuality. Many non-Elements books contain diagrams or are tagged as diagrammatic in metadata without advertising diagrams on the title page. Elements editions, by contrast, more often make figures visible when those figures are part of a title-page claim about proof, apparatus, pedagogy, or the furnished state of Euclid.

So the report should not say "Elements is visual" as a distinctive claim. It should say:

Figures become distinctive only when joined to other Elements-specific operations: demonstration, propositions, scholia, ancient apparatus, correction, translation, or pedagogical/practical repackaging.

## Functional Split

The analysis separates visuality into five functions:

1. **Proof apparatus**: figures tied to demonstrations, propositions, theorems, corollaries, or demonstrative method.
2. **Practical operation**: figures tied to measuring, operations, construction, instruments, surveying, or practical use.
3. **Visual pedagogy**: figures tied to explanation, ease, clarity, learners, readers, or intelligibility.
4. **Edition furnishing**: figures as added, augmented, engraved, newly supplied, or part of a fuller apparatus.
5. **Ancient/learned apparatus**: figures tied to scholia, Theon/Proclus, Greek/Latin mediation, ancient restoration, or learned commentary.

These functions overlap. A figure can be proof apparatus and edition furnishing at once, or visual pedagogy and practical operation at once.

## Elements Book Groups

| Book Group | Visual Claim | Proof Apparatus | Practical Operation | Visual Pedagogy | Edition Furnishing | Ancient/Learned Apparatus |
|---|---:|---:|---:|---:|---:|---:|
| books 1-6 | 19.8% | 14.8% | 8.6% | 11.1% | 14.8% | 19.8% |
| books 1-6 + 11-12 | 10.3% | 7.7% | 0.0% | 7.7% | 5.1% | 10.3% |
| near-complete/expanded | 25.4% | 16.4% | 0.0% | 3.0% | 22.4% | 25.4% |
| selected later books | 9.5% | 4.8% | 0.0% | 0.0% | 4.8% | 9.5% |

Interpretation:

Near-complete/expanded Elements editions are the main visual-apparatus zone. Figures here tend to belong to a learned, furnished, ancient/commentarial Euclid.

Plain `1-6` is more mixed and historically important. It is the only major Elements book group where figures register as practical operation as well as proof apparatus and pedagogy. This fits the Dutch/Dou route: figures are not merely diagrams of propositions; they are connected to foundations, utility, and geometrical operations.

`1-6 + 11-12` has less figure language than expected. Its practical-pedagogical force is more often carried by method, explanation, propositions, use, and solid geometry than by figures as such.

## Natural Modes

The strongest visual mode is `euclid_composite_workshop`:

- visual title claim: 45.7%;
- edition furnishing: 45.7%;
- ancient/learned apparatus: 45.7%;
- proof apparatus: 23.4%;
- visual pedagogy: 13.8%.

Interpretation:

Composite-workshop Euclid is where figures are most heavily advertised. This is the mode of supplied, corrected, augmented, illustrated, apparatus-rich Euclid.

Other routes behave differently:

- `euclid_procedural_pedagogical`: visual claim 20.5%, proof apparatus 13.6%, ancient/learned apparatus 20.5%.
- `euclid_utility_public`: visual claim 10.5%, practical operation 10.5%, ancient/learned apparatus 10.5%.
- `elements_language_no_euclid`: visual claim 29.0%, edition furnishing 25.8%, visual pedagogy 16.1%, practical operation 12.9%.
- `not_euclid_elements`: visual claim 14.5%, metadata diagrams 22.5%, edition furnishing 10.2%, practical operation 6.5%.

Interpretation:

Visuality does not map neatly onto Euclid/non-Euclid. Within Euclid, it is strongest in composite/apparatus editions. In the surrounding ecology, diagrams are often materially present even when not foregrounded as an intellectual title-page value.

## Format And Title-Page Fashion

Format matters:

| Format | Visual Title Claim | Metadata Diagrams | Proof Apparatus | Practical Operation | Edition Furnishing |
|---|---:|---:|---:|---:|---:|
| folio | 23.4% | 23.4% | 3.2% | 3.2% | 17.0% |
| quarto | 12.3% | 19.8% | 6.6% | 0.9% | 7.5% |
| octavo | 21.0% | 1.6% | 16.1% | 5.6% | 17.7% |
| duodecimo | 7.9% | 2.6% | 5.3% | 0.0% | 7.9% |

Interpretation:

This partially supports the user's caution about format. Folios and quartos more often have metadata diagram tags, which may reflect larger books, learned apparatus, or more visually furnished publication formats. Octavos, however, have strong title-page visual claims but low metadata diagram tags; in the Elements subset, this likely reflects small-format pedagogical or portable Euclid where figures are rhetorically important even when not captured as rich diagram metadata.

Therefore, silence or density claims about figures should always be controlled by format.

## Language And Chronology

Language:

- Dutch has high visual title claims: 25.5%, with proof apparatus 21.6%, practical operation 15.7%, visual pedagogy 17.6%, edition furnishing 23.5%.
- English has the highest visual title claim among major languages: 29.9%, with edition furnishing 29.9%, practical operation 19.4%, visual pedagogy 17.9%.
- Latin is lower overall: visual title claim 14.4%, proof apparatus 5.9%, practical operation 1.6%.
- Italian has high metadata diagrams, 42.2%, but lower visual title claims, 20.3%.

Interpretation:

Dutch and English are the most visibly pedagogical/practical visual languages in this corpus. Latin and Italian may contain or imply visual apparatus, but title pages less often turn it into a practical/pedagogical selling point.

Chronology:

- visual title claims stay around 16-18% from pre-1550 through 1649;
- they dip in 1650-1699 to 13.8%;
- they rise in 1700+ to 27.3%;
- practical-operation visuality rises most clearly in 1700+ at 16.4%.

Interpretation:

This may signal a later title-page fashion of advertising figures, plates, and visual aids more explicitly. Because the 1700+ corpus is smaller and uneven, use this as a hypothesis, not a standalone claim.

## Cases For Close Reading

### Figures As Learned/Ancient Apparatus

Useful cases:

- `Urbino_1575`;
- `Cologne_1591`;
- `Rome_1603`;
- `Frankfurt_1607`;
- `Paris_1622`;
- `Pesaro_1572`;
- `Rome_1591`;
- `Rome_1609`.

Use:

These cases support figures as part of a restored, corrected, scholia-rich, Greek/Latin, or institutionally furnished Euclid.

### Figures As Practical/Pedagogical Euclid

Useful cases:

- `Leiden_1607`;
- `Amsterdam_1616`;
- `Amsterdam_1626`;
- `Rotterdam_1632`;
- `Rotterdam_1647`;
- `Rotterdam_1661`;
- `Rotterdam_1681`;
- `Amsterdam_1700b`;
- `Amsterdam_1701`;
- `London_1685a`.

Use:

These cases are strongest for the argument that figures can help turn Euclid into practical or pedagogical work. The Dutch/Dou route is especially important because figure language is tied to operations, utility, foundations, explanation, and public mathematical readership.

### Figures Outside The Elements

Useful cases:

- `bib-5` / `bib-90` (1585 Paris);
- `bib-71` (1567 Venice);
- `YB32U5` (1604 Leiden);
- `bib-133` (1650 London);
- `bib-135` (1667 Nuremberg);
- `bib-238` (1716 Paris).

Use:

These support the ecology argument: visuality belongs broadly to practical geometry, instruments, architecture, perspective, commerce, and material mathematical arts.

## Report Claim To Use

Figures and diagrams should be treated as a shared technology of mathematical print, not as a stable marker of the Elements. In the Elements corpus, figures become historically meaningful when they are folded into specific acts of mediation: restoring ancient apparatus, furnishing a complete or corrected edition, explaining propositions, supporting demonstrations, or enabling practical geometrical operations. In the broader ecology, diagrams often belong to instrumental, visual, commercial, architectural, surveying, and material mathematical work. The difference is not visual versus non-visual mathematics, but what title pages ask visuality to do.

This supports the report's social-intellectual balance. The same visual object can serve different social and intellectual programs: learned recovery, institutional authority, pedagogical clarity, professional operation, or practical utility.

## Cautions

- Metadata `has_diagrams` and title-page figure claims are different kinds of evidence. A book can contain diagrams without advertising them.
- Missing metadata diagram tags should not be interpreted as proof that a book lacks figures.
- The figure-function split is pattern-based and should guide close reading.
- Format matters: folios/quartos and octavos advertise or contain visual material differently.
- Figure language is not uniquely Elements. Do not use it as a standalone Elements identity marker.
