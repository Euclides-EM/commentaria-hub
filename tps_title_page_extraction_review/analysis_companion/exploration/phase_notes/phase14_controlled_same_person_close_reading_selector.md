# Phase 14 Controlled Same-Person Close-Reading Selector

Question:

Can the Phase 13 corpus-internal portfolio pattern survive a stricter comparison: same named person, same city, same first language, and within five years?

This phase does not use full bibliographies. It only selects controlled comparison pairs inside the represented title-page corpus.

## Data Produced

Script:

`scripts/build_controlled_portfolio_close_reading.py`

Outputs:

- `derived_data/controlled_author_editor_close_pairs.csv`
- `derived_data/controlled_author_editor_close_pair_shortlist.csv`
- `derived_data/controlled_author_editor_close_pair_summary.csv`

Selection rule:

- same split person;
- one metadata-defined Elements case and one non-Elements case;
- same city;
- same first language;
- year gap <= 5.

Strict same-format is tracked, but not required.

## Counts

Controlled pairs:

89

People represented:

13

Strict same-format controlled pairs:

15

Top dense clusters:

| person | controlled pairs | Elements cases | non-Elements cases | note |
|---|---:|---:|---:|---|
| Pierre Forcadel | 38 | 5 | 12 | only cluster with strict same-format pairs |
| Denis Henrion | 30 | 4 | 17 | richest same-city/same-language cluster |
| Jacques Ozanam | 4 | 2 | 3 | late French practical/public comparison |
| Jean Errard | 4 | 2 | 2 | professional/practical comparison |
| Marin Mersenne | 3 | 1 | 3 | useful negative/quiet comparison |
| Christopher Clavius | 2 | 2 | 2 | Latin/Rome institutional comparison |
| Claude-Francois Milliet Dechales | 2 | 2 | 1 | French/Paris comparison |

Several others appear as one-pair anchors: Federico Commandino, Honorat de Meynier, Oronce Fine, Pierre Bourdin, Pierre de la Ramee, Thomas Rudd.

## Main Result

The controlled comparison preserves the most important Phase 13 signal, but makes the claim narrower.

Preserved:

- Elements cases often add ancient authority/restoration.
- Elements cases often add correction, translation, augmentation, selection, or method.
- Utility/practice is often stronger in neighboring non-Elements works.

Narrowed:

- The social pattern varies by cluster. Elements is not always more socially explicit.
- Format control is still weak except for Forcadel, because many non-Elements comparison rows lack format metadata.
- The evidence is strongest as a close-reading selector, not as a standalone statistical proof.

## Cluster Signals

### Denis Henrion

Controlled pool:

- 30 pairs;
- 4 Elements cases;
- 17 non-Elements cases;
- Paris, French, within five years.

Large deltas in controlled pairs:

- ancient authority/restoration: +93 pp;
- translation/transfer: +67 pp;
- correction/revision: +60 pp;
- augmentation/enrichment: +43 pp;
- utility/practice: -30 pp;
- no visible social arena: -37 pp.

Interpretation:

Henrion is the best dense cluster for testing "Elements as canon-work." In close Paris/French comparisons, the Elements side looks more ancient-authority, correction, translation, and augmentation oriented. Nearby non-Elements works more often carry utility/practice and visual-material claims.

Good comparison pairs:

- `Paris_1623` against same-year Paris non-Elements cases;
- `Paris_1630` against `MorisS-16-11`;
- `Paris_1615` against Paris 1613/1618/1620 neighboring works.

### Pierre Forcadel

Controlled pool:

- 38 pairs;
- 5 Elements cases;
- 12 non-Elements cases;
- Paris, French, within five years;
- 15 strict same-format pairs.

Large deltas:

- translation/transfer: +50 pp;
- ancient authority/restoration: +37 pp;
- method/order: -42 pp;
- no visible social arena: +32 pp;
- court/state service arena: -32 pp.

Interpretation:

Forcadel is the strongest format-control cluster, but he complicates the claim. Elements adds translation and ancient authority, yet method/order appears more strongly in nearby non-Elements works. This makes Forcadel a good corrective case: not all Elements canon-work is method-heavy.

Good comparison pairs:

- `Paris_1565` against `UEIZFO` in 1565, both Paris/French/quarto;
- `Paris_1564`, `Paris_1566a`, `Paris_1566b` against nearby Paris/French applied, arithmetic, and instrument works.

### Christopher Clavius

Controlled pool:

- 2 pairs;
- Rome, Latin, within three years.

Signal:

- Elements adds method/order, correction, ancient authority, and visual/material aids.
- Religious institutional arena is shared.

Interpretation:

This is a clean institutional canon-work example. The social setting remains stable, while the Elements title page adds textual/methodological labor.

Good pairs:

- `Rome_1603` vs `bib-6` Rome 1604;
- `Rome_1589` vs `X2T7PY` Rome 1586.

### Jacques Ozanam

Controlled pool:

- 4 pairs;
- Paris, French, within three years.

Signal:

- ancient authority/restoration: +100 pp;
- military/fortification arena: +100 pp;
- general public/lovers arena: +100 pp;
- utility/practice: -50 pp;
- augmentation/enrichment: -50 pp;
- no visible social arena: -75 pp.

Interpretation:

Ozanam is useful because the Elements does not simply become "more practical" than the surrounding mathematical works. Instead, the Elements title pages connect ancient authority to military/public social arenas, while non-Elements neighboring works carry more generic utility.

### Jean Errard

Controlled pool:

- 4 pairs;
- Paris, French, within three years.

Signal:

- ancient authority/restoration: +100 pp;
- augmentation/enrichment: +100 pp;
- selection/extraction/abridgment: +100 pp;
- professional/practical and court/state arenas are shared.

Interpretation:

Errard is a good case for practical/professional continuity plus Elements-specific canon-work. The same practical social world remains visible, but the Elements side adds ancient authority and textual construction.

### Thomas Rudd

Controlled pool:

- 1 pair;
- London, English, one-year gap.

Signal:

- Elements adds ancient authority and military/fortification arena.
- non-Elements comparison adds utility/practice, visual/material aids, learned/scholarly and school arenas.
- both share method/order, augmentation, selection, court/state, and professional/practical arenas.

Interpretation:

A promising English case, but only one controlled pair. It should be used as an illustrative close reading, not a statistical anchor.

## Cases That Weaken Or Complicate The Claim

### Marin Mersenne

The controlled Mersenne pair set is mostly quiet: the Elements case does not strongly add the expected canon-work flags. This is useful because it prevents a universal claim.

### Honorat de Meynier

The controlled pair shares ancient authority, method, and selection; the non-Elements side adds utility/practice. This supports the idea that practical usefulness may live more strongly outside the Elements even when the Elements is nearby.

### Forcadel

Forcadel supports ancient authority and translation, but not method/order. This matters because the canon-work argument should not collapse into "Elements always means method."

## Revised Claim After Controls

Within close same-person, same-city, same-language windows, Elements title pages often add canonical/textual mediation: ancient authority, correction, translation, augmentation, selection, or method.

But the exact mixture changes by person and local setting. Utility/practice is often stronger in nearby non-Elements works, and method/order is not always an Elements-side signal.

The safest formulation:

**In the represented title-page corpus, the Elements often functions as canon-work: not necessarily the most practical book in a portfolio, and not always the most socially explicit, but a place where mathematical knowledge is anchored in ancient authority and re-mediated through correction, translation, selection, augmentation, and sometimes method.**

## Best Close-Reading Anchors For The Talk

1. Henrion: strongest dense evidence for Elements as correction/translation/ancient-authority canon-work.
2. Forcadel: strongest format-control case and best complication.
3. Clavius: clean institutional canon-work case.
4. Errard or Ozanam: practical/public worlds meeting Elements-specific canonical mediation.
5. Rudd: promising English single-pair case.

## What To Do Next

Read the title pages for:

- `Paris_1623`, `Paris_1630`, and their Henrion comparisons;
- `Paris_1565` and `UEIZFO` for Forcadel;
- `Rome_1603` and `bib-6` for Clavius;
- `Paris_1693` or `Paris_1697` for Ozanam;
- `Paris_1598a` or `Paris_1605` for Errard.

The goal should be to decide which two or three clusters can carry the argument in slides.
