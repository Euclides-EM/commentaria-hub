# Phase 17 New / Reorganized Elements

Question:

When, where, and how often do title pages present the Elements as new, newly ordered, newly demonstrated, reduced, contracted, symbolically retooled, or otherwise reconstructed?

This follows the user's suggestion that `Paris_1667`, `Nouveaux Elemens de Geometrie`, may mark an opposite pole from fidelity-to-the-ancients or philological-restoration ideals.

## Data Produced

Script:

`scripts/build_new_reorganized_elements.py`

Outputs:

- `derived_data/new_reorganized_elements_cases.csv`
- `derived_data/new_reorganized_elements_summary.csv`

The finder is intentionally broad. It captures strong "new Elements" cases, but also weaker cases: new/easy method, new demonstrations, contraction, selected propositions, algebraic/symbolic retooling, and rearranged pedagogical forms.

## Counts

Broad candidates:

80

| reconstructive pole | cases | claim novelty | method/order | ancient restoration | hybrid with correction/restoration |
|---|---:|---:|---:|---:|---:|
| strong new Elements | 2 | 100.0% | 100.0% | 50.0% | 0 |
| new order and demonstration | 3 | 100.0% | 100.0% | 100.0% | 2 |
| reordered or new demonstrations | 18 | 33.3% | 77.8% | 83.3% | 4 |
| selected/reduced/contracted | 28 | 10.7% | 71.4% | 96.4% | 9 |
| method/ease/retooling | 29 | 3.4% | 65.5% | 93.1% | 11 |

Immediate conclusion:

The anti-philological pole exists, but it is small if defined strictly. Most reconstructive rhetoric is hybrid. Early modern title pages often claim to reorganize, simplify, demonstrate, contract, or adapt Euclid while still presenting the work as Euclidean, corrected, restored, or anciently authorized.

## The Strict Core

### `Paris_1667`: Arnauld, `Nouveaux Elemens de Geometrie`

This is the cleanest case.

Title-page signals:

- `Nouveaux Elemens de Geometrie`;
- `un ordre tout nouveau`;
- new demonstrations of common propositions;
- new means for showing incommensurable lines;
- new measures of angles;
- new ways of finding and demonstrating line proportions.

Feature profile:

- novelty/modernity: yes;
- method/demonstration/order: yes;
- ancient-restoration: no;
- practical/public: no;
- natural dominant mode: pedagogical/method.

Interpretation:

This is not a title page about fidelity to Euclid as an ancient text. It claims the right to remake the Elements as an intellectual and pedagogical structure. Its authority is not primarily ancientness, but better order, better demonstrations, and new ways of handling mathematical objects.

This is the strongest "anti-philological" case in the current pass.

### `Livorno_1709`: `Euclides Reformatus`

Title-page signals:

- `Euclides reformatus`;
- plane and solid geometry;
- new method;
- clearer, easier, firmer, more evident exposition/demonstration than before.

Interpretation:

This is another strong reconstructive case. It still keeps Euclid's name, but the operative word is reform. The title page claims improvement of method and demonstration.

Difference from Arnauld:

Arnauld says "new Elements" and "new order." Marchetti says "Euclid reformed." Arnauld sounds like constructing new Elements; Marchetti sounds like reforming Euclid.

## New Method / New Demonstration Without Fully New Elements

### Dechales / Reeve Williams: `London_1685a`, `London_1703`

Signals:

- Elements of Euclid explained and demonstrated in a new and easy method;
- uses of each proposition in all parts of mathematics;
- corrected and augmented;
- translated from French;
- often Jesuit identity.

Interpretation:

These are not anti-philological in the strict sense. They are hybrids.

They claim new method and pedagogical ease, but they still belong to an apparatus of correction, augmentation, translation, and institutional authority. This is "new method for Euclid," not "new Elements against Euclid."

### Henry Hill, `473J72`

Signals:

- first six plus eleventh and twelfth books;
- demonstrated after a new, plain, and easy method.

Interpretation:

This is a pedagogical method claim. It belongs near Dechales but may be less institutionally/philologically loaded.

## Reordered / Retooled Euclid

### Naboth, `Cologne_1556`

Signals:

- first book of geometrical Elements;
- selected propositions from following books;
- ordered so they can be demonstrated according to the method of geometers.

Interpretation:

This is a very important early case. It does not claim "new Elements," but it takes liberties with sequence: propositions from later books are selected and ordered for demonstrability.

This is reconstruction by selection and order.

### Le Tenneur, `Paris_1640`

Signals:

- tenth book of Euclid;
- new demonstrations easier and more succinct than ordinary ones;
- reduced to 62 propositions;
- discourse on explaining sciences in French.

Interpretation:

This is a strong pedagogical reconstruction case, especially because it combines:

- new easier demonstrations;
- reduction of propositions;
- French scientific explanation.

This is close to the user's question about reorganizing Elements for pedagogical reasons.

### Reyher, `Kiel/Hamburg/Leipzig_1697` and `Kiel/Leipzig_1699`

Signals:

- first six books of Euclid;
- special/easy manner;
- algebraic signs;
- signs borrowed from the newest analytic art;
- proof usable in other languages.

Interpretation:

This is not philological restoration. It is symbolic retooling. Euclid is made portable through algebraic notation and analytic technique.

This may be one of the best cases for "reconstructing the Elements for pedagogical/technical transfer."

### Burckhard von Pirckenstein, `Vienna_1694` and `Lübeck_and_Frankfurt_1699`

Signals:

- eight books of the beginnings of measurement;
- new and very easy manner;
- directed to generals, engineers, natural/truth seekers, builders, artists, and craftsmen.

Interpretation:

This is the practical/public version of new-order Euclid. It is not a restoration text; it is a German practical-mathematical adaptation for technical users.

### König / Kuypers, `The_Hague_1758`

Signals:

- first six books of Euclid;
- put in a new order;
- revised;
- for youth;
- professor König.

Interpretation:

This is a late pedagogical reorder case. It directly links new order to youth/teaching.

## Selection, Contraction, Reduction

This group is large and more hybrid. It includes:

- contracted Euclid;
- selected propositions;
- abridged demonstrations;
- shorter, clearer, more convenient forms;
- propositions or theorems selected from Archimedes;
- Euclid extracted into usable packages.

Examples:

- `London_1651`, Rudd: first six books contracted and demonstrated.
- `Rome_1629`, Grienberger/Clavius: first six books contracted from Clavius's larger commentaries into a more convenient form.
- `Paris_1639` / `Paris_1644`, Herigone: demonstrated by notes in a very brief and intelligible method.
- `Antwerp_1654` and Tacquet line: plane and solid geometry with selected Archimedean theorems.
- `Cambridge_1703`: corollaries fitted to illustrate the Elements and uses of propositions.
- `Lisbon_1735`: geometry according to Euclid's order, with useful additions.

Interpretation:

These are not anti-philological in a simple way. They show a widespread willingness to reshape the Elements for usefulness, brevity, convenience, illustration, or institutional teaching while retaining Euclidean authority.

## Relationship To Fidelity / Philology

The user's intuition is right, but the opposition should be modeled as a spectrum:

1. **Philological-restorative pole**: restore, correct, purge errors, return to Greek/ancient authority, remove Theon or corrupted transmission.
2. **Hybrid pedagogical-mediation pole**: correct/translate/augment Euclid while making him easier, shorter, clearer, more useful, or more demonstrable.
3. **Reconstructive pole**: new Elements, reformed Euclid, new order, new demonstrations, symbolic/algebraic retooling.

`Paris_1667` belongs near pole 3.

Dechales, Tacquet, Rudd, Herigone, Clavius/Grienberger often sit in pole 2.

Simson-like restoration cases sit near pole 1.

## How This Changes The Argument

The Elements was not only transmitted. It was also redesigned.

But redesign usually did not mean rejecting Euclid. Most title pages claim liberty inside a Euclidean frame:

- make Euclid easier;
- make Euclid shorter;
- rearrange Euclid;
- select from Euclid;
- demonstrate Euclid differently;
- translate Euclid into notation, vernacular, or institutional curriculum.

Only a few cases strongly imply "new Elements" or "Euclid reformed" as the central claim.

## Best Slide-Level Claim

**Early modern title pages show a spectrum of permissible intervention in Euclid: from restoring the ancient text, through correcting and explaining it, to remaking its order, demonstrations, and pedagogical form.**

## Best Cases

1. `Paris_1667`: strongest "new Elements" / new order / new demonstrations case.
2. `Livorno_1709`: `Euclides reformatus`.
3. `Cologne_1556`: selected propositions from later books newly ordered for demonstration.
4. `Paris_1640`: tenth book reduced to 62 propositions with new easier demonstrations.
5. `Kiel_and_Leipzig_1699`: Euclid retooled through algebraic signs from newest analytic art.
6. `London_1685a`: hybrid new/easy method plus correction/augmentation/translation.
7. `The_Hague_1758`: first six books put in new order for youth.

## Relation To Dutch Dou

Dou operationalizes Euclid through translation, explanation, correction, added utilities, and geometrical operations. But he does not appear to claim a new Elements in the Arnauld sense.

So Dou belongs to practical-vernacular canon-work, not the strong reconstructive pole.

## Future Link To Deductive Parts

This analysis points directly to the user's next question about title-page mentions of mathematical parts:

- propositions;
- demonstrations;
- definitions;
- axioms;
- corollaries;
- scholia;
- figures;
- theorems;
- notes;
- diagrams.

The reconstructive pole often appears through exactly these parts: new demonstrations, selected propositions, reduced proposition counts, scholia, corollaries, notes, figures, algebraic signs.
