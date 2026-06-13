# Commentary, Scholia, And Explanation: Report Notes

Date: 2026-06-12

Purpose:

This note separates "commentary" into historically different title-page functions. The goal is to avoid treating ancient scholia, Jesuit/Clavius apparatus, explanatory pedagogy, and notes/annotations as the same intellectual move.

Script:

- `report/scripts/build_commentary_split.py`

Main outputs:

- `tables/report_commentary_split_by_corpus_matrix.csv`
- `tables/report_commentary_split_by_elements_books_group_matrix.csv`
- `tables/report_commentary_split_by_natural_dominant_mode_matrix.csv`
- `tables/report_commentary_split_by_period_matrix.csv`
- `tables/report_commentary_split_by_language_first_matrix.csv`
- `tables/report_commentary_split_cooccurrence_matrix.csv`
- `tables/report_commentary_split_elements_cases.csv`
- `tables/report_commentary_split_close_reading_shortlist.csv`
- `figures/heatmap_commentary_split_by_corpus.png`
- `figures/heatmap_commentary_split_by_elements_books_group.png`
- `figures/heatmap_commentary_split_by_natural_dominant_mode.png`
- `figures/heatmap_commentary_split_by_period.png`
- `figures/heatmap_commentary_split_by_language_first.png`

## Headline Result

Commentary/explanation language appears in 92/286 metadata Elements representatives, 32.2%, compared with 77/557 non-Elements representatives, 13.8%.

But the important result is not only frequency. The subtypes divide into different historical routes:

1. **Ancient/humanist scholia**: ancient scholia, Theon/Proclus, Greek/Latin mediation, old/new mathematicians, recovery from ancient sources.
2. **Clavius/Jesuit commentary**: accurate scholia, Clavius commentary, contracted Clavius, Jesuit institutional apparatus.
3. **Pedagogical explanation**: explained, expliquez, verklaert/verclaringen, spiegata, commentés in the sense of making Euclid intelligible.
4. **Notes/annotations/observations**: notes, annotations, observations, animadversions, critical/explanatory supplements.
5. **Contracted/extracted commentary**: commentary reduced or contracted into a more convenient form.
6. **General scholia**: scholia language not otherwise classified.

Report implication:

The Elements is not just "commented." Commentary itself is a historical variable. It can restore ancient authority, institutionalize Euclid, compress learned apparatus, explain propositions to learners, or add critical/explanatory aids.

## Elements Versus Non-Elements

| Commentary Type | Elements | Non-Elements | Difference |
|---|---:|---:|---:|
| any commentary/explanation split | 32.2% | 13.8% | +18.4 pp |
| ancient/humanist scholia | 12.6% | 1.4% | +11.2 pp |
| pedagogical explanation | 12.9% | 5.6% | +7.3 pp |
| Clavius/Jesuit commentary | 3.5% | 1.1% | +2.4 pp |
| contracted/extracted commentary | 1.7% | 0.0% | +1.7 pp |
| notes/annotations/observations | 5.6% | 6.6% | -1.0 pp |

Interpretation:

The Elements strongly over-indexes commentary/explanation, but not all forms equally. The most distinctive Elements forms are ancient/humanist scholia and pedagogical explanation. Notes/observations are not especially Elements-specific; they belong to the broader mathematical ecology too.

## Book-Group Patterns

### `books_1_6_plus_solids`

Rates:

- any commentary/explanation: 38.5%;
- pedagogical explanation: 33.3%;
- ancient/humanist scholia: 2.6%;
- notes/annotations: 2.6%;
- no Clavius/Jesuit commentary in this subtype pass.

Interpretation:

`1-6 + 11-12` is overwhelmingly explanatory rather than scholia-humanist. This supports its role as a practical-pedagogical package. Its commentary mode is not primarily ancient textual apparatus; it makes Euclid intelligible and usable.

This aligns with the proposition-use finding: `1-6 + 11-12` is where explanation, proposition-use, method, and practical pedagogy cluster, especially in the Dechales/Reeve/Williams route.

### `books_1_6`

Rates:

- any commentary/explanation: 30.9%;
- pedagogical explanation: 14.8%;
- ancient/humanist scholia: 6.2%;
- contracted/extracted commentary: 6.2%;
- notes/annotations: 6.2%;
- Clavius/Jesuit commentary: 3.7%.

Interpretation:

Plain `1-6` is mixed. It can be explanatory, contracted, practical-vernacular, or institutional. This fits our broader claim that plain `1-6` is not one route: Dutch/Dou practical vernacular, German explanatory, and Latin contracted Clavius-style forms all coexist.

### `near_complete_or_expanded`

Rates:

- any commentary/explanation: 34.3%;
- ancient/humanist scholia: 17.9%;
- Clavius/Jesuit commentary: 9.0%;
- scholia general: 11.9%;
- pedagogical explanation: 1.5%.

Interpretation:

Near-complete/expanded Elements editions are the main commentarial-apparatus zone. They are not primarily "explained for ease"; they tend to present Euclid as a learned ancient/institutional corpus, often with scholia, Clavius, Greek/Latin authority, or apparatus.

### `selected_later_books`

Rates:

- any commentary/explanation: 38.1%;
- ancient/humanist scholia: 28.6%;
- pedagogical explanation: 9.5%.

Interpretation:

Selected later books often remain close to humanist/ancient scholarly mediation. Selection here does not automatically mean school simplification; it can also mean learned handling of difficult parts of the corpus.

## Natural Mode Patterns

### `practical-pedagogical`

- any commentary/explanation: 41.7%;
- pedagogical explanation: 29.2%;
- ancient/humanist scholia: 6.9%.

Interpretation:

Practical-pedagogical Elements is explanation-heavy. This is not commentary as ancient apparatus, but commentary as access, clarity, and use.

### `institutional-composite`

- any commentary/explanation: 39.8%;
- ancient/humanist scholia: 17.7%;
- Clavius/Jesuit commentary: 7.1%;
- notes/annotations: 8.0%;
- pedagogical explanation: 9.7%;
- general scholia: 10.6%.

Interpretation:

Institutional-composite Elements is the mixed apparatus zone. It can carry ancient scholia, Clavius/Jesuit scholia, notes, and some explanation. It is not simply "pedagogical"; it is furnished, authorized, and institutionally mediated.

### `humanist/vernacular transfer`

- any commentary/explanation: 30.4%;
- ancient/humanist scholia: 30.4%;
- pedagogical explanation: 4.3%.

Interpretation:

This mode is almost entirely ancient/humanist commentary rather than classroom explanation. It supports the claim that translation/transfer can be humanist restoration, not simply access for beginners.

### `pedagogical/method`

- any commentary/explanation: 27.6%;
- pedagogical explanation: 10.3%;
- ancient/humanist scholia: 10.3%;
- contracted/extracted commentary: 6.9%.

Interpretation:

Pedagogical/method is mixed, but less commentary-heavy than practical-pedagogical. Its distinctive signal remains demonstration and method rather than explanation alone.

## Chronology

Commentary/explanation changes over time:

| Period | Any Commentary | Ancient/Humanist | Pedagogical Explanation | General Scholia |
|---|---:|---:|---:|---:|
| pre-1550 | 34.4% | 21.9% | 12.5% | 0.0% |
| 1550-1599 | 49.1% | 30.9% | 9.1% | 14.5% |
| 1600-1649 | 29.3% | 8.0% | 9.3% | 5.3% |
| 1650-1699 | 22.9% | 2.4% | 16.9% | 2.4% |
| 1700+ | 30.0% | 10.0% | 15.0% | 0.0% |

Interpretation:

The sixteenth century is the strongest moment for ancient/humanist scholia and general scholia. Pedagogical explanation becomes more visible later, especially after 1650. This supports a diachronic shift in title-page commentary rhetoric: from learned recovery and scholia toward explanation, method, and intelligibility.

This is not a complete replacement. Ancient/humanist forms persist, and later learned editions still exist. But the balance changes.

## Language

Language strongly differentiates commentary types:

| Language | Any Commentary | Ancient/Humanist | Pedagogical Explanation |
|---|---:|---:|---:|
| Greek | 56.2% | 50.0% | 25.0% |
| Dutch | 47.6% | 4.8% | 38.1% |
| English | 45.0% | 5.0% | 25.0% |
| French | 35.8% | 7.5% | 18.9% |
| Latin | 28.1% | 15.8% | 4.3% |
| German | 15.4% | 0.0% | 15.4% |

Interpretation:

Latin and Greek preserve the ancient/humanist commentary route. Dutch, English, and French are where commentary most often becomes explanation. This helps us say something historically sharper about vernacularization:

Vernacular Elements is not just translation. In many cases, it changes the advertised function of commentary from scholarly apparatus to intelligibility, use, and public pedagogy.

## Important Case Routes

### Ancient / Humanist Scholia

Cases:

- `Pesaro_1572`: Euclid with ancient scholia.
- `Urbino_1575`: vernacular Euclid with ancient scholia.
- `Paris_1557`: Greek and Latin Euclid.
- `Basel_1537` / `GY5QLD`: Theon/exposition tradition.
- `Rome_1591` and `Rome_1609`: Phaenomena, ancient scholia, figures, Greek to Latin, Maurolico annotations.
- `Berlin_1824`: commentaries from ancient and recent mathematicians.

Use:

These support the learned/restorative apparatus route.

### Clavius / Jesuit Institutional Commentary

Cases:

- `Cologne_1591`;
- `Frankfurt_1607`;
- `Rome_1603`;
- `Frankfurt_1654`;
- `Rome_1629`;
- `Rome_1655`;
- `Graz_1636`;
- `Mainz_1611_1612`.

Use:

These support the institutional apparatus route: Euclid as a furnished, corrected, demonstrated, scholia-rich, Jesuit-mathematical corpus.

### Pedagogical Explanation

Cases:

- `Amsterdam_1700`, `Lausanne_1683`, `London_1685a`, `London_1696`, `London_1703`, `Oxford_1685`, `Oxford_London_1700`: Dechales/Reeve/Williams explanatory route.
- Dutch/Dou route: `Amsterdam_1700b`, `Amsterdam_1701`, `Rotterdam_1647`, etc., with translation, explanation, correction, public lovers, and foundations.
- `Ansbach_1610`, `Basel_1562`: German explanatory first-six-books route.
- `Brussels_1689`: translated/explained `1-6 + 11-12`.

Use:

These support explanation as access and pedagogy, often vernacular and sometimes practical/public.

### Notes / Annotations / Critical Aids

Cases:

- `London_1570`: scholies, annotations, inventions, Dee's preface.
- `Paris_1639` / `Paris_1644`: Herigone-style demonstration by notes, brief/intelligible method.
- `London_1789`: notes critical and explanatory.
- `Oxford_1705`: annotations and useful supplements.
- `KVSFF1`: notes in a later French geometry Elements context.

Use:

Notes are not purely ancient nor purely pedagogical. They can be critical, explanatory, symbolic, supplementary, or method-oriented. They should be treated as a flexible apparatus type.

## Report Claim To Use

The Elements was not merely "commented." Title pages advertise different kinds of commentary, and those differences map onto historical routes. Earlier and learned editions often present commentary as ancient or humanist apparatus: scholia, Greek/Latin mediation, Theon, Proclus, or learned recovery. Jesuit/Clavius editions turn commentary into institutional apparatus: accurate scholia, demonstrations, correction, contraction, and a furnished Euclidean corpus. Vernacular and practical-pedagogical editions increasingly turn commentary into explanation: Euclid is translated, explained, clarified, and made usable.

This strengthens the report's social-intellectual argument. The social world shapes the intellectual function of commentary. The same broad act, commenting on Euclid, can mean recovering ancient authority, institutionalizing a curriculum, compressing learned apparatus, or making propositions usable for learners and practitioners.

## Cautions

The subtype detection is pattern-based and should guide close reading, not replace it.

Potential noise:

- Greek/Latin signals can mark language transfer rather than scholia proper.
- "Commentary" can be ancient apparatus or explanation depending on context.
- "Illustrated" can mean figures, proof, explanation, or general enrichment.
- Some rows refer to non-Elements works bound with or adjacent to Elements material.

Before final prose, use the case table and inspect the exact title-page evidence for each case used as an example.
