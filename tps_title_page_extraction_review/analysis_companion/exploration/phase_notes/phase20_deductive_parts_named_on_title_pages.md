# Phase 20: Deductive Parts Named On Title Pages

Date: 2026-06-11

Question:

Which parts of mathematical or deductive content are named on title pages: demonstrations, propositions, theorems, problems, figures, scholia, notes, corollaries, definitions, axioms, postulates, lemmas, enunciations, examples, operations, constructions, principles, paradoxes?

This phase asks what title pages publicly say mathematical knowledge is made of. It does not ask what the books actually contain. A title page can contain propositions, axioms, diagrams, and scholia without advertising them. So the evidence here is about public framing and advertised value.

Data:

- Source matrix: `derived_data/metadata_elements_corpus_ecology_matrix.csv`
- Enriched mode/format source for the metadata Elements subset: `derived_data/metadata_elements_natural_modes_matrix_with_format.csv`
- Script: `scripts/build_deductive_parts_analysis.py`
- Outputs:
  - `derived_data/deductive_parts_summary.csv`
  - `derived_data/deductive_parts_cases.csv`
  - `derived_data/deductive_parts_combinations.csv`
  - `derived_data/deductive_parts_pairs.csv`
  - `derived_data/deductive_parts_by_strata.csv`
  - `derived_data/deductive_parts_interesting_cases.csv`

Method:

I searched representative title-page evidence for a multilingual set of terms around named deductive parts. The categories are intentionally broad enough for navigation, not final philological proof. Each cited case still needs close reading before being used as a decisive example.

## Headline Counts

Named deductive or mathematical parts appear in 359 of 843 representative rows, 42.6%.

The metadata-defined Elements corpus is much more likely to advertise such parts:

| Corpus | Rows With Named Parts | Total Rows | Rate |
|---|---:|---:|---:|
| All representatives | 359 | 843 | 42.6% |
| Metadata Elements | 165 | 286 | 57.7% |
| Non-Elements ecology | 194 | 557 | 34.8% |

This is already important. Elements editions are not only more likely to name Euclid or canonical books. They are more likely to advertise the internal furniture of demonstrative knowledge.

## Parts Most Associated With Elements

| Part | Elements | Non-Elements | Difference |
|---|---:|---:|---:|
| demonstrations/proofs | 60/286, 21.0% | 42/557, 7.5% | +13.4 points |
| scholia/commentary | 39/286, 13.6% | 36/557, 6.5% | +7.2 points |
| principles | 27/286, 9.4% | 17/557, 3.1% | +6.4 points |
| propositions | 29/286, 10.1% | 25/557, 4.5% | +5.7 points |
| theorems | 17/286, 5.9% | 21/557, 3.8% | +2.2 points |
| enunciations | 8/286, 2.8% | 0/557, 0.0% | +2.8 points |

The strongest Elements signal is not "figures" or "examples." It is a constellation:

- demonstration/proof;
- proposition;
- scholium/commentary;
- principle;
- theorem;
- occasionally enunciation, corollary, definition, axiom.

This suggests that advertised Elements are framed as a demonstrative and commentarial corpus: a body of ordered statements, proofs, and explanatory apparatus.

## Parts More Associated With The Wider Ecology

| Part | Elements | Non-Elements | Difference |
|---|---:|---:|---:|
| problems | 4/286, 1.4% | 21/557, 3.8% | -2.4 points |
| operations/constructions | 3/286, 1.0% | 23/557, 4.1% | -3.1 points |
| examples | 3/286, 1.0% | 12/557, 2.2% | -1.1 points |
| notes/observations | 11/286, 3.8% | 28/557, 5.0% | -1.2 points |
| figures/diagrams | 21/286, 7.3% | 43/557, 7.7% | -0.4 points |

The broader mathematical ecology is not less "mathematical." It advertises different units of mathematics: problems, operations, constructions, examples, observations, and figures. That is closer to procedural, instrumental, problem-solving, and practical knowledge.

So the contrast is not abstract mathematics versus practical mathematics. It is closer to:

- Elements: proof-bearing and commentarial units.
- Non-Elements ecology: procedural, operational, visual, problem-solving units.

## Demonstration Is The Core Elements Signal

Demonstrations/proofs are named in 21.0% of metadata Elements representatives, compared with 7.5% of non-Elements representatives.

Inside the Elements subset, this varies by period:

| Period | Elements Demonstration/Proof Rate |
|---|---:|
| pre-1550 | 2/32, 6.2% |
| 1550-1599 | 8/55, 14.5% |
| 1600-1649 | 19/75, 25.3% |
| 1650-1699 | 24/83, 28.9% |
| 1700+ | 7/40, 17.5% |

This supports the previous reconstruction/restoration analysis. The mid-to-late seventeenth century is when title pages most strongly advertise the Elements as something to be demonstrated, redemonstrated, made easier to demonstrate, or reorganized through proof.

This does not mean proofs become newly important in the books. It means proof becomes more explicit as a title-page value.

## Commentary And Scholia Are Earlier And More Institutional

Scholia/commentary are also strongly Elements-associated: 13.6% in Elements versus 6.5% in non-Elements.

Within the Elements subset:

- pre-1550: 8/32, 25.0%;
- 1550-1599: 13/55, 23.6%;
- 1600-1649: 10/75, 13.3%;
- 1650-1699: 6/83, 7.2%.

This has the opposite trajectory from demonstrations. Commentary/scholia are especially visible earlier, while demonstration/proof language grows later.

Interpretation:

The advertised Elements shifts, at least in title-page rhetoric, from a learned/commentarial ancient corpus toward a more demonstrative, pedagogically reorganizable corpus. That is not a total replacement. Both continue. But the public emphasis changes.

This aligns with the restoration clusters:

- philological/commentarial restoration is especially visible in earlier and humanist-institutional forms;
- logical/demonstrative restoration becomes more visible in the seventeenth century.

## Figures Are Not An Elements Specialty

Figures/diagrams appear at almost the same rate:

- Elements: 21/286, 7.3%;
- Non-Elements: 43/557, 7.7%.

This matters because one might expect Euclid to own the visual language of geometry. The title-page evidence says otherwise. Figures are part of the broader mathematical publishing ecology: arithmetic, perspective, instruments, practical geometry, surveying, and mixed mathematics also sell themselves through figures.

For Elements, figures tend to matter in specific routes:

- French Henrion-style full Elements with added figures and demonstrations;
- Dutch practical/public first-six-books editions, where figures are tied to foundations, usefulness, and geometrical operations;
- composite or contracted institutional editions where figures accompany commentary or apparatus.

But figures alone do not define the Elements subset. The distinctive Elements signal is not visuality by itself. It is figures joined to demonstration, proposition, principles, or commentary.

## Propositions Divide Into Several Uses

Propositions are named in:

- Elements: 29/286, 10.1%;
- Non-Elements: 25/557, 4.5%.

In Elements, propositions can mean at least four things:

1. **The internal unit of Euclidean proof**  
   Example: editions advertising propositions of the Elements or propositions from later books.

2. **The unit selected, rearranged, or reduced**  
   `Cologne_1556` selects propositions from following books and orders them for demonstration. `Paris_1640` reduces book 10 to 62 propositions with easier and more succinct demonstrations.

3. **The unit whose use is explained**  
   Dechales/Reeve Williams style title pages emphasize the use of each proposition in all parts of mathematics.

4. **The portable extract**  
   Strasbourg propositiones/enunciationes editions advertise Euclid as a set of extractable statements.

This is a crucial bridge between social and intellectual history. Propositions are not neutral pieces. They can be selected for students, ordered for proof, attached to practical uses, extracted for portability, or used to rebuild the logic of the corpus.

## Problems And Operations Belong More To The Neighboring Ecology

Problems and operations/constructions are more common outside Elements:

- problems: Elements 1.4%, non-Elements 3.8%;
- operations/constructions: Elements 1.0%, non-Elements 4.1%.

This reinforces the earlier claim that the wider mathematical ecology often advertises immediate mathematical doing: solving, constructing, operating, using tables, applying rules, handling instruments, measuring, and resolving practical cases.

Important edge cases:

- `Nuremberg_1625`: an Elements case that explicitly advertises `Elementa practica` and extracts all problems/hand-work from the 15 books. This is not typical, but it shows one way Euclid can be pulled toward practical operation.
- `London_1747` and `6XTAVU`: later English Elements/geometry cases where problems and operations connect to mensuration, maxima/minima, and school or academy contexts.
- `MorisS-16-1`, `ustc-26`, and similar non-Elements military or mathematical memoirs show the neighboring procedural world more clearly.

The Elements can become procedural, but when it does so the title page often marks it as a special transformation: practical Elements, extracted problems, uses of propositions, or applications.

## Rare Foundational Parts Are Rarely Advertised

The most foundational logical parts are almost absent as title-page selling points:

- definitions: 6 total, 4 Elements;
- axioms/common notions: 2 total, 1 Elements;
- postulates: 0;
- lemmas: 0;
- corollaries: 4 total, 3 Elements.

This absence is historically useful. Title pages rarely advertise Euclid by saying: here are axioms, postulates, definitions, and lemmas. They much more often advertise:

- demonstrations;
- propositions;
- scholia/commentary;
- figures;
- principles;
- theorems;
- uses.

So the Elements is not publicly framed mainly as a foundation-list or axiomatic skeleton. It is framed as a working demonstrative-commentarial machine: statements, proofs, explanations, additions, corrections, and applications.

This is important for the "real Euclid" question. Reconstructive editions do not usually say they are restoring axioms or postulates. They say they restore order, demonstration, ease, method, or use.

## Natural Mode And Format Controls

Within metadata Elements, the natural-mode overlap is helpful:

| Natural Mode | Strongest Deductive-Part Signals |
|---|---|
| pedagogical/method | demonstrations/proofs 13/29, 44.8%; propositions 7/29, 24.1% |
| institutional-composite | demonstrations/proofs 30/113, 26.5%; scholia/commentary 26/113, 23.0%; theorems 13/113, 11.5% |
| practical-pedagogical | demonstrations/proofs 15/72, 20.8%; propositions 14/72, 19.4%; principles 13/72, 18.1%; figures 10/72, 13.9% |

This sharpens the internal map:

- pedagogical/method Elements are most strongly proof/proposition oriented;
- institutional-composite Elements are most commentarial and theorem/apparatus oriented;
- practical-pedagogical Elements combine propositions, principles, figures, and occasional problems/operations.

Format also matters:

| Format Group | Notable Signals |
|---|---|
| folio | scholia/commentary 14/54, 25.9%; demonstrations/proofs 7/54, 13.0% |
| quarto | demonstrations/proofs 15/67, 22.4%; principles and scholia 9/67 each, 13.4% |
| octavo | demonstrations/proofs 24/111, 21.6%; figures 15/111, 13.5%; propositions 12/111, 10.8% |
| duodecimo | propositions 10/30, 33.3%; demonstrations/proofs 8/30, 26.7% |

Caution: format is not a causal explanation by itself. But it helps prevent a bad reading. Dense commentarial language is partly tied to folio/institutional forms, while proposition-heavy portable/pedagogical forms show up strongly in smaller formats. So "sparse" or "part-heavy" title pages should be read with format in mind.

## Case Families

### 1. Commentarial-Humanist Euclid

Cases:

- `Pesaro_1572`: Euclid with ancient scholia.
- `Urbino_1575`: vernacular Euclid with ancient scholia.
- `Strasbourg_1564`, `Strasbourg_1564a`, `Strasbourg_1566`: Greek/Latin, Theon, commentaries, school context, analyses.
- `London_1570`: Euclid translated with Dee's preface and scholia-like apparatus.

Interpretation:

Here Euclid is advertised as a learned ancient corpus requiring commentary, scholia, linguistic mediation, and institutional handling.

### 2. Demonstrative-Pedagogical Euclid

Cases:

- `Douai_1620`, `Douai_1625`: demonstrations adapted to easier understanding.
- `Paris_1640`: book 10 reduced and redemonstrated more easily/succinctly.
- `Paris_1667`: new order, new demonstrations, new methods.
- `London_1685a`, `London_1703`, `473J72`: new/plain/easy method and demonstrations.

Interpretation:

Here Euclid is not merely transmitted. He is made demonstrable for a target reader. This is where "logical restoration" is strongest.

### 3. Proposition-Use Euclid

Cases:

- `Lausanne_1683`: use of each proposition.
- `London_1685a`, `London_1703`: uses of each proposition in the parts of mathematics.
- `Cologne_1556`: selected propositions ordered for demonstration.
- `Amsterdam_1662`: short explanations of some propositions.

Interpretation:

This is one of the best bridges between intellectual and social analysis. The proposition is a proof unit, but also a teachable unit, a usable unit, and sometimes a portable unit. It lets Euclid move between canon, classroom, and application.

### 4. Practical-Operational Euclid

Cases:

- `Nuremberg_1625`: practical Elements, extracting problems and hand-work from the 15 books.
- Dutch/Dou first-six-books route: principles, foundations, figures, operations, utilities, public lovers and surveyor authority.
- Later English school/academy geometry cases such as `London_1747` and `6XTAVU`.

Interpretation:

This is not the dominant Elements mode, but it is historically important because it shows Euclid being converted into operative geometry. It is a transformation of the corpus, not simply the natural default of the corpus.

### 5. Composite-Apparatus Euclid

Cases:

- Clavius/Tacquet-related institutional editions with scholia, selected Archimedean theorems, figures, corollaries, or added apparatus.
- `Cambridge_1703`: propositions, theorems, figures, corollaries.
- `Antwerp_1645`: unusually dense combination of demonstrations, propositions, corollaries, definitions, and common notions/axioms.

Interpretation:

These title pages frame Euclid as a furnished corpus. The intellectual ideal is not minimal purity, but a well-equipped demonstrative text.

## Main Historical Interpretation

The Elements lives in the broader mathematical ecology as a special kind of object: a corpus whose advertised parts are disproportionately demonstrative and commentarial.

This distinguishes it from neighboring mathematical books, which more often advertise operations, problems, examples, visual aids, practical uses, tables, and applications.

But the Elements is not sealed off from that ecology. It touches it through specific transformations:

- propositions become useful;
- demonstrations become easier, newer, or clearer;
- figures become pedagogical and practical aids;
- selected books or propositions become portable curricula;
- scholia and commentary make the ancient corpus institutionally legible;
- practical editions extract problems, operations, and utilities.

So the Elements' place in the ecology is not just "canonical ancient geometry." It is better described as a canonical demonstrative corpus that different social and pedagogical settings recompose by foregrounding different parts.

## What This Adds To The Reconstructive-Restoration Question

The deductive-parts analysis supports the idea that reconstructive Euclid is another kind of restoration.

The parts most implicated in reconstruction are:

- demonstrations;
- propositions;
- order;
- selected/reduced proposition sets;
- uses of propositions;
- occasionally figures, corollaries, and symbolic/analytic aids.

The parts not usually foregrounded are:

- axioms;
- postulates;
- lemmas;
- definitions, except in a few cases.

That means the reconstructive project is rarely advertised as rebuilding Euclid from axiomatic foundations. It is advertised as restoring or improving the functioning of Euclid as a demonstrative sequence: propositions in order, demonstrations made easier or firmer, uses clarified, and the corpus adjusted to learners or practitioners.

## Provisional Argument For The Final Report

Title pages do not present the Elements simply as a book of geometry. They present it as a manipulable demonstrative corpus. Its public identity is built from propositions, demonstrations, scholia, principles, figures, and selected apparatus. Different historical settings activate different parts:

- humanist/institutional settings foreground scholia, ancient commentary, correction, and textual mediation;
- pedagogical/method settings foreground demonstrations, propositions, ease, order, and intelligibility;
- practical-pedagogical settings foreground use, principles, figures, and sometimes problems or operations;
- neighboring non-Elements works foreground procedural and application-oriented units more directly.

This gives us a stronger, more precise version of the earlier claim: the Elements is not merely one geometry book among others. It is a canonical proof-corpus whose internal parts can be advertised, rearranged, selected, explained, operationalized, and socially redirected.

## Open Questions

1. Which exact cases should become close-reading anchors for each part-family?
2. Are proposition-use cases concentrated in particular book groups, especially `1-6 + 11-12`, after controlling for Dechales/Reeve Williams?
3. How do diagrams in metadata/images relate to title-page figure claims? Figure language is not uniquely Elements, so visuality needs its own analysis.
4. Can we separate commentary as ancient/humanist apparatus from commentary as pedagogical explanation?
5. Do smaller formats disproportionately turn Euclid into a portable proposition/proof package, while folios preserve the commentarial ancient corpus?
