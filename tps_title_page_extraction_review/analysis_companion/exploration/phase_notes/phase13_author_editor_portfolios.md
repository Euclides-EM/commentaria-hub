# Phase 13 Author / Editor Portfolios

Question:

For named authors/editors who appear in both the metadata-defined Elements corpus and the wider mathematical title-page corpus, do Elements editions continue the same intellectual/social program, or do they change the title-page rhetoric?

This is exploratory and corpus-internal. The metadata field is `author_or_editor`, so it mixes authors, editors, translators, commentators, and sometimes multi-person title-page attributions. I split comma-separated names into individual people, but this is not a fully disambiguated authority file.

Major limitation:

We do not have the full bibliography of every named person. Therefore this phase cannot support claims about a person's whole career, full intellectual profile, or whether Euclid was central or marginal to that person's life work. It only shows how named people appear within this title-page corpus.

Use this phase to compare represented title pages, not complete bibliographies.

## Data Produced

Script:

`scripts/build_author_editor_portfolios.py`

Outputs:

- `derived_data/author_editor_portfolio_case_matrix.csv`
- `derived_data/author_editor_portfolio_person_summary.csv`
- `derived_data/author_editor_portfolio_raw_combo_summary.csv`
- `derived_data/author_editor_portfolio_interesting_people.csv`
- `derived_data/author_editor_portfolio_elements_non_elements_pairs.csv`

The rebuilt tables use the corrected metadata-defined Elements corpus and merge back `format`, `volumes`, and diagram metadata from `title_page_analysis_matrix.csv`.

## Basic Counts

Representative title-page rows in the rich matrix:

843

Metadata-defined Elements representative rows:

284

Named people with at least one represented Elements case:

160

Named people with both Elements and non-Elements represented:

38

Portfolio categories among people with Elements cases:

| category | people | meaning |
|---|---:|---|
| elements_only_low_count | 105 | one or two Elements cases, no non-Elements cases in this corpus |
| small_crossover | 22 | at least one Elements and at least one non-Elements case, but low counts |
| elements_heavy_within_sample | 17 | at least three Elements cases and no non-Elements cases in this corpus sample |
| bridge_portfolio | 13 | at least two Elements and at least two non-Elements cases |
| few_elements_many_non_elements_within_sample | 3 | one or few Elements cases and at least five non-Elements cases in this corpus sample |

Important caution:

`elements_heavy_within_sample` does not mean the historical person only worked on Euclid. It means our title-page corpus currently represents them mostly or entirely through Elements editions.

## First Result

The Elements corpus is not just a subject subset. In author/editor portfolios, it often behaves like a special kind of work: a canon-work.

By canon-work I mean title-page work around textual authority, ancient restoration, correction, translation, augmentation, selection, demonstration, and pedagogical packaging. This can overlap with practical or institutional aims, but it is not reducible to them.

This matters because it answers one of the user's major worries: the Elements should not be treated merely as a title designation or as one subject among others. It appears as a corpus that invites a particular kind of editorial and intellectual labor.

## Global Contrast

Across all representative cases in this phase:

| feature | Elements | non-Elements |
|---|---:|---:|
| ancient authority / restoration | 83.5% | 12.2% |
| method / demonstration / order | 41.9% | 16.5% |
| augmentation / enrichment / composition | 40.8% | 28.8% |
| translation / vernacularization / transfer | 26.4% | 8.4% |
| correction / revision / accuracy | 23.2% | 11.4% |
| utility / practice / application | 7.7% | 17.7% |
| pedagogical / school arena | 23.2% | 12.0% |
| religious / institutional arena | 21.1% | 7.3% |
| no visible social arena | 44.0% | 60.3% |

This suggests a useful correction to earlier instincts:

Elements title pages are not simply more "theoretical" or less social. They are often more explicitly framed through authority, method, textual intervention, and institutional/pedagogical location.

The main exception is utility/practice/application, which is more frequent in the wider non-Elements corpus. Practicality is not absent from Elements, but it is not the default Elements claim.

## Same-Person Bridge Portfolios

The strongest test is the 13 bridge portfolios with at least two Elements and two non-Elements cases.

Across those 13 bridge people, average Elements-minus-non-Elements deltas are:

| feature | average delta |
|---|---:|
| ancient authority / restoration | +86.3 pp |
| translation / vernacularization / transfer | +35.2 pp |
| correction / revision / accuracy | +23.8 pp |
| method / demonstration / order | +19.4 pp |
| selection / extraction / abridgment | +18.9 pp |
| augmentation / enrichment / composition | +14.7 pp |
| no visible social arena | -30.3 pp |
| utility / practice / application | -13.2 pp |

The ancient-authority shift is the important one: all 13 bridge portfolios have higher ancient/restoration signaling in Elements than in their non-Elements works.

This is not yet a final proof, because title-page format, city, language, and genre still matter. But it is too consistent to ignore.

## Portfolio Types

### 1. Canon-Work Bridge Figures

These people have both Elements and non-Elements title pages, and the Elements editions often intensify ancient authority, translation, correction, or augmentation.

Examples:

| person | Elements | non-Elements | profile |
|---|---:|---:|---|
| Denis Henrion | 9 | 18 | French bridge figure; Elements heavily stress ancient authority, correction, translation, augmentation |
| Pierre Forcadel | 5 | 16 | Paris/French bridge; Elements add translation and ancient authority more than his non-Elements works |
| Christopher Clavius | 9 | 4 | Latin/Jesuit institutional bridge; Elements intensify correction, augmentation, ancient authority |
| Oronce Fine | 2 | 7 | broader mathematical portfolio; Elements intensify method, correction, translation, ancient authority |
| Federico Commandino | 3 | 2 | Elements strongly marked as augmentation, translation, ancient authority |
| Jacques Peletier | 2 | 3 | Elements marked by method, correction, augmentation, translation, ancient authority |
| Niccolo Tartaglia | 2 | 3 | Elements carry translation and ancient authority, plus school arena in this tagging |

Interpretation:

For these figures, Elements editions are not simply another book in the portfolio. They are places where a mathematical worker appears as restorer, translator, corrector, organizer, or mediator of an ancient/canonical corpus.

### 2. Practical Or Public Bridge Figures

These figures connect Elements to practical or public mathematical worlds, but not always through the same rhetoric as their non-Elements books.

Examples:

| person | Elements | non-Elements | notable pattern |
|---|---:|---:|---|
| Jean Errard | 3 | 5 | professional/practical arena appears in both Elements and non-Elements; Elements still add ancient authority |
| Jacques Ozanam | 3 | 4 | non-Elements have more utility; Elements are military-fortification tagged and add correction/ancient authority |
| Jan Pietersz Dou | 12 | 3 | Dutch/German practical-pedagogical Elements route; Elements carry method, correction, augmentation, ancient authority, and some professional/school signals |

Interpretation:

This is where the social and intellectual paths meet most clearly. Practical publics can attach to Elements, but the Elements usually become practical through canon-handling: methodizing, correcting, selecting, adding, translating, or adapting.

### 3. Elements-Heavy Figures Within This Sample

Examples:

| person | Elements cases | formats/languages | profile |
|---|---:|---|---|
| Conrad Dasypodius | 10 | quarto/octavo; Greek/Latin | Elements corpus represented through repeated editorial/educational work |
| Isaac Barrow | 9 | duodecimo/sexto/octavo; English/Latin | Elements travels through language and format variation |
| Andre Tacquet | 8 | octavo; Latin | strongly institutional/pedagogical Elements tradition |
| Jean Magnien / Etienne Gracile | 7-8 | octavo; Latin | repeated Latin Elements apparatus/enunciation tradition |
| Johannes de Sacrobosco | 7 | folio; Latin | canon-adjacent learned/theoretical packaging |
| Pierre Le Mardele | 4 | octavo; French | French near-complete Elements route |

Interpretation:

These are useful for internal Elements ecology, but they should not be treated as "Elements specialists" in a biographical sense. They are less useful for comparing the same person's Elements and non-Elements rhetoric because the non-Elements side is absent from this title-page sample, not necessarily absent from their bibliography.

### 4. Few Elements Cases Inside A Larger Sample Presence

Examples:

| person | Elements | non-Elements | profile |
|---|---:|---:|---|
| Marin Mersenne | 1 | 16 | one Elements/enunciations case inside a large mathematical-philosophical title-page presence |
| Claude Hardy | 1 | 8 | one Elements case that is unusually rich in augmentation/translation/ancient authority compared with his non-Elements cases |
| Jacques Lefevre d'Etaples | 1 | 5 | one Elements case inside a larger learned corpus |

Interpretation:

These cases are good for asking when Euclid is rare within the represented corpus, but not whether Euclid was actually marginal in the person's full bibliography. They should be close-read rather than averaged.

## City / Language / Format Movement

Among the 38 people with both Elements and non-Elements cases:

| pattern | count |
|---|---:|
| single city | 15 / 38 |
| multiple first languages | 17 / 38 |
| multiple formats | 17 / 38 |

Among the 13 bridge portfolios:

| pattern | count |
|---|---:|
| single city | 2 / 13 |
| multiple first languages | 7 / 13 |
| multiple formats | 8 / 13 |

Interpretation:

For major bridge figures, author/editor identity is not enough. Their portfolios often move across city, language, and format. This means we need to keep fashion controls alive even in portfolio analysis.

But the repeated ancient-authority shift in Elements remains strong even across this mobility.

## Close Comparison Candidates

The pair table identifies Elements/non-Elements cases by the same person, sorted by year gap. Especially useful candidates are same-city and same-language comparisons within five years.

Examples:

| person | Elements case | non-Elements comparison | why useful |
|---|---|---|---|
| Denis Henrion | `Paris_1623`, `Paris_1630` | Paris non-Elements cases from the same years | same author, city, language, and year; strong local control |
| Pierre Forcadel | `Paris_1564`, `Paris_1565`, `Paris_1566a/b` | Paris arithmetic / applied-military / instrument works around 1565-1570 | excellent for a Paris/French portfolio comparison |
| Christopher Clavius | `Rome_1603` | `bib-6` Rome 1604 | close Latin/Rome institutional comparison |
| Jacques Ozanam | `Paris_1693`, `Paris_1697` | Paris mathematical/instrumental works 1691-1700 | useful for late French practical/public comparison |
| Thomas Rudd | `London_1651` | `bib-133` London 1650 | close English comparison |
| Federico Commandino | `Pesaro_1572` | `bib-42` Pesaro 1570 | close Latin/Pesaro comparison |

These are likely better for the talk than abstract examples because they control for author, place, language, and approximate moment.

## What This Does To The Argument

Earlier frame:

Elements editions are a subcorpus within mathematical print.

Better frame:

Within this title-page corpus, Elements editions are a recurring site where mathematical workers perform canon-work: they restore, correct, translate, augment, select, demonstrate, and relocate an ancient corpus for particular social worlds.

This lets us hold the social and intellectual analysis together:

- social: schools, religious orders, professional publics, military/practical readers, printers/publishers, cities, formats;
- intellectual: authority, method, correction, translation, augmentation, selection, utility;
- bridge: the same person can present non-Elements mathematics as useful, instrumental, commercial, or topical, while presenting Elements as a canon that must be mediated.

## Provisional Claim

In the represented title-page ecology, the Elements is not just one subject category and not just a stable title designation. It functions as a canonical corpus that repeatedly asks for editorial mediation. Among people who have both Elements and non-Elements works represented in this corpus, title-page rhetoric often changes when they enter the Elements: ancient authority, correction, translation, augmentation, and method become more prominent, while generic utility is less central.

This does not mean the Elements is socially detached. On the contrary, Elements title pages are often less socially silent than non-Elements title pages in this feature set. The social worlds of the Elements are not always loud practical publics; they are often schools, religious institutions, learned apparatuses, and localized translation/correction projects.

## Questions Opened

1. Which bridge figures should become close-reading anchors for the talk?
2. Do same-author, same-city, same-language pairs confirm the aggregate portfolio pattern?
3. Does the Dutch Dou route behave differently from Latin/French canon-work?
4. Do author/editor portfolios align with Elements bookgroup choices, especially plain `1-6`, `1-6 + 11-12`, and near-complete editions?
5. Are format shifts driving some of the portfolio differences, or do the same-format pairs preserve the same pattern?
6. Can we identify "Elements as canon-work" without relying on Wardhaugh/textual family labels?

## Recommended Next Move

Do a close comparison of a small number of same-person clusters:

1. Henrion: French/Paris Elements versus French/Paris non-Elements geometry/instruments.
2. Forcadel: Paris/French Elements versus arithmetic/applied works.
3. Dou: Dutch plain `1-6` Elements route and its social/practical pedagogy.
4. Clavius or Tacquet: Latin institutional Elements canon-work.
5. Ozanam or Errard: practical/military public mathematics touching Elements.

This would turn the statistics into historical texture and help decide which cases belong in the conference argument.
