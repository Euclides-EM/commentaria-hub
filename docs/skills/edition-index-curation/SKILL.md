---
name: edition-index-curation
description: Report suggested fixes for the headers of an OCR transcription (ocrflow/store/data/transcriptions/<book>/page-NNNN/original.md) - numbering gaps in proposition/problem/theorem/chapter sequences, wrong header levels, OCR misreads, printer's errors, inconsistent case and stray spaces, and recommends curated index layers (proposition-kind layers such as propositions, theorems, problems, definitions, axioms, using only explicitly printed classification; marginal content). Report only - never edits transcriptions unless the user explicitly asks to apply approved layers. Use when asked to curate, check or review headers or header numbering, with or without a generated index.json.
---

# Edition index curation (report only)

This skill produces a report of suggested header fixes and curated index layers. It never changes a transcription: do not edit, write, or run any command that modifies files under `ocrflow/store/data/transcriptions/`, even for fixes that look certain. The user applies the suggestions they accept. The only exceptions are inserting index layers, and removing duplicate `numbered_item` entries, when the user has explicitly asked to apply them (see "Applying approved layers").

Scope is headers only: markdown headers (`#`, `##`, `###` ...) and curated headings (`[Curated heading level=N ...: TEXT]`). To save tokens, do not open whole `original.md` files. Work from the script output. When a flag needs context, grep the header line with one or two lines around it (`grep -n -B1 -A2 'PROP. XXV' <dir>/page-0123/original.md`).

Books like Euclid editions number every proposition twice: a running proposition number (`PROP. XXV.`, roman) and a per-kind counter (`THEOR. 22.` / `PROBL. 5.`, arabic). Both restart at each section (`# ELEMENT ...`). OCR breaks these in predictable ways. Each header in a section should continue an unbroken sequence, sit at the right level, and be spelled the way the book prints it.

## Inputs

- Transcription dir: `ocrflow/store/data/transcriptions/<book>/` with `page-NNNN/original.md`. Index page numbers equal the `NNNN` folder number.
- Optional `index.json` (the generated header tree: `nodes[]` with `category` `header1/2/3`, `content`, `location.page`, `children`). If the user mentions an index but gives no path, ask for it. If none exists, work from the transcription dir. Both scripts accept either.

## Step 1: Sequence gaps

```
python3 docs/skills/edition-index-curation/scripts/find_gaps.py <index.json | transcription dir>
```

It splits by level-1 header, then checks the PROP sequence, the THEOR and PROB counters separately, and `CHAP.` roman numbers. When scanning a dir it also reports lines that look like `THEOR. n. PROP. X.` but are not markdown headers. A flag that follows a bad value can be a knock-on effect. The script resyncs, but read each flag together with its neighbours.

## Step 2: Levels, characters, case and spacing

```
python3 docs/skills/edition-index-curation/scripts/lint_headers.py <index.json | transcription dir>
```

It prints flagged headers as `p<page> L<level> 'text' -> notes`:

- Level: a label (first four letters, e.g. `CORO`, `SCHO`, `THEO`) at a level it rarely has in the book, a level skip (`#` followed by `###`), or a header that is only a numeral (`### IIII.`, usually a list item or a sub-number rather than a header).
- Spelling: a label written differently from its usual form (`LE M ME` vs `LEMME`), or close to a frequent label (`cO ROLLAAIRE` vs `COROLLAIRE`).
- Characters: letters mixed into numbers (`I4`, `59n`), unexpected characters.
- Case: lowercase letters in an uppercase header, a case change inside a word.
- Spacing and punctuation: repeated spaces, space before punctuation, repeated or mixed punctuation (`PROP..`, `PROP,.`), no space between label and number (`THEO.23.`), extra spaces around the header text.

Then review the outline yourself for problems the heuristics miss: wrong nesting (a `COROLLAIRE` or `SCHOLIE` at the level of the propositions it belongs to, a section header at level 2), misreads in words (`COMMVNES` is period spelling, not an error), and headers whose level differs from their siblings:

```
python3 docs/skills/edition-index-curation/scripts/lint_headers.py <src> --outline
```

`--outline` prints every header, one line each, with any notes appended. Read it section by section rather than all at once for long books.

- **Book finish statements are not headers.** Closing formulas such as `FINIS LIBRI PRIMI`, `FIN DU PREMIER LIVRE`, `ELEMENTI SECVNDI FINIS`, `LAUS DEO` end a section; they do not open one. A finish statement marked as a header always needs a fix: suggest the same line without the `#` prefix. Never count it as a section header for sequence checks.
- **Corollaries, lemmas and scholia are never standalone.** A `COROLLAIRE`/`COROLLARIUM`/`CORO.`, `LEMME`/`LEMMA` or `SCHOLIE`/`SCHOLIUM` belongs to a theorem, problem or proposition (or another numbered item such as a definition) and must sit one level below it. Flag any of them at the same level as, or above, the propositions of its section, or nested directly under a section header with no proposition in between. Suggest the level one deeper than its parent item. A lemma that the book places before the proposition it serves still nests under the preceding item; mention the ambiguity in the reason.
- **Long headers are suspect.** A header that runs well beyond a label, number and short title (a whole sentence or enunciation, several labels joined together, a paragraph) is probably either several headers merged into one, which should be broken into a header and sub-headers, or a false positive: body text, a caption or a running head wrongly marked with `#`. Flag every long header. When several long headers share the same cause (consecutive pages, the same run of merged labels, the same misclassified block), group them into one finding that lists their pages instead of repeating the same reason. Suggest the split lines with their levels, or the line without the `#` prefix, and say which case you judge it to be.

## Step 3: Classify each suggestion

The transcription must reflect the printed book, which has its own errors. Separate what is plausibly an OCR mistake (a character misread as a visually similar one) from what could be in the print. Grep the line before deciding.

**Likely OCR error:**

1. Character misreads in a number: `S`/`1S`→8/18, `1I`/`I`/`l`→1, `D`→12, a dropped repeated digit (`1.` for 11, `3.` for 31), stray letters glued to a number (`59n.`→`59.`), a misplaced dot inside a number (`2.7`→`27.`). In roman numerals: `H` for `I`/`II` (`VHI`, `XHII`, `IHII`), `I` for `L` (`IXXX`→LXXX), `E` for `C`, lowercase `x`, and `IHE0` for `THEO`.
2. Character misreads in words: split or doubled letters (`LE M ME`, `ROLLAAIRE`), stray lowercase in an uppercase header (`cO`), a glyph read as a similar one.
3. Header present as plain text: the header line exists but lacks `#`, or is split across lines (`THEOR. 7.` / blank / `PROP. IX.`).
4. Wrong level: the header text is right but its `#` count does not match its siblings, or a corollary, lemma or scholium is not nested under its item.
5. Not a header: a book finish statement (or other closing formula) marked with `#`.

**Possible printer's error (needs a check against the scan):**

- Wrong kind: `PROBL.` where the sequence expects `THEOR.`, or the reverse. OCR does not turn one word into the other.
- A number that is missing rather than misread (`THEOR. PROP. XXV.`).
- Roman numerals that need a character added, or are transposed (`XLVII` where XLVIII is expected, `CXIII` vs `XCIII`). The book may repeat or misprint a number.
- A header that is missing entirely.
- A missing section header: numbering restarts at 1–2 mid-section. Note whether the preceding page's `original.md` is empty.

**Formatting (low confidence):** punctuation and spacing (`THEOR22.`, `PROP..`, `PROP,`, `THEOR.15.`). Report them, but the print may have them too.

## Step 4: Write the suggestion

For each item give the exact replacement line, written per `docs/MARKDOWN_DIALECT.md`, so the user can apply it as is:

- OCR error or wrong level: the corrected line, changing only the misread characters or the `#` count. Keep the surrounding punctuation, spacing and label spelling as printed.
- Header as plain text: the joined line with the `#` prefix, wording unchanged.
- Not a header: the same line without the `#` prefix, wording unchanged.
- Printer's error: keep the printed reading and append the correction to the wrong token: `## PROBL{printer-error-correction:THEOR}. 20. PROP. XXII.`, `PROP. CXIII{printer-error-correction:XCIII}.`. If something was left out in print, attach the correction to the token before the gap: `## THEOR.{printer-error-correction:THEOR. 25.} PROP. XXV.`
- Header not in print: a curated heading with no `type` (default index layer), placed just before the statement with a blank line before and after it: `[Curated heading level=2: THEOR. 32. PROP. XXXIIII.]`
- Page transcription empty: say so; do not transcribe it.

Page images for checking are at `ocrflow/store/data/<dataset_id>/imgs/page-NNNN.png` (`dataset_id` is in `index.json`). Mention the image path for every possible printer's error. Do not open or view page images yourself unless the user asks; leave the check against the scan to the user.

## Step 5: Recommend curated indices

Beyond fixing headers, recommend parallel index layers (typed curated headings, `[Curated heading level=N type=TYPE: TEXT]`, see `docs/MARKDOWN_DIALECT.md` section 4.2). First list the layers the book already has (`grep -rhoE 'type=[A-Za-z0-9_.-]+' <dir> | sort | uniq -c`) and never propose a duplicate of one.

### Proposition-kind layers

The main layers track each kind of enunciated unit separately, so a reader can browse only the theorems, only the problems, only the definitions, and so on. Use one layer per kind, with these `type` names:

- `propositions`: every proposition (or the book's equivalent unit: `Vortrag`, `Satz`, a numbered `Theorema` in a part that has no `PROPOSITIO`).
- `theorems` and `problems`: only when the book prints the kind (see the explicit-only rule below).
- `definitions`, `postulates`, `axioms` (common notions, `Grundsätze`, `COMMVNES SENTENCES`).
- `lemmas`: only when the book numbers a lemma as a standalone unit (`PROPOSITION n.` followed by `LEMME.`, or a numbered `LEMME I.` not attached to a proposition).
- `additions`, `added_propositions`: added theorems, problems or propositions the book presents as standalone units on a par with propositions (`Propoſitiones à Campano additæ`, an appendix that numbers its own propositions).

Scope is proposition level or above. Do not recommend layers for sub-proposition units: corollaries (`COROLL.`, `Poriſma`, `Zuſatz`, `Zugabe`, `Conſectarium`), scholia, lemmas attached to a proposition, remarks and notes (`Anmerkung`, `NOTE`), appendices attached to a proposition (`Anhang`), additions inside a proposition's commentary (`EX PROCLO.`, `CAMPANI additio`), cases, and alternative demonstrations (`ALITER`, `Autrement`). List them under "Considered, not recommended" with their counts.

**Explicit-only rule.** Layer membership and heading text must reflect only what the print states, or a `{printer-error-correction:...}` already in the transcription. Never infer:

- **Kind.** Put an item in `theorems` or `problems` only when its own header (or the line printed with it) names the kind: `THEOR.`/`PROBL.`, `Theorema`/`Problema`, `THÉORÊME`/`PROBLÊME`, `Lehrſatz`/`Aufgabe`, `Θεώρημα`/`Πρόβλημα`. Do not classify by the content of the enunciation, by elimination (unmarked propositions are not "theorems"), by a translator's note about a marker word, or by another edition. If the book prints only `PROP.`, recommend `propositions` and say that `theorems`/`problems` are not applicable. Where only some sections print the kind, those layers cover only those sections.
- **Numbers.** Use the number as printed (roman or arabic, as in the book). A definition or axiom with no printed number (the decorated initial takes its place, or the definitions are run together in a paragraph) gets its printed opening words as text (`Equales circuli`, `Vnité`), not a positional number. A run-in paragraph of several definitions gets only a group entry.
- **Existence.** A unit whose header is not printed (only an untyped curated heading supplies it) gets no typed entry.
- **Corrections.** Use the corrected reading only when the transcription already carries `{printer-error-correction:...}`. A suspected misprint without that markup keeps the printed reading in the curated text and goes into the findings list as a possible printer's error.

**Levels.** Each layer is a self-contained hierarchy: level 1 is the book or section (its title as printed, e.g. `LIBER PRIMVS`, `Das erſt buͦch Euclidis`), level 2 is the item. Use level 3 only for a printed sub-grouping inside a book (e.g. lettered axioms under numbered ones), never for sub-proposition material.

**Text.** Copy the printed label and number (`THEOR. 5. PROPOS. VII.`, `DEF. I.`, `Erkl. 1)`), with no markup and no `]`. Drop figure references and the enunciation. A short locator prefix is allowed when the item would be ambiguous on its own (`Lib. V, Pars II, Theorema 1.`); keep it factual and never let it imply a kind.

**Placement.** Each item entry goes immediately before the printed header or numbered line it indexes, with one blank line before and after. A level-1 entry goes right before that layer's first item in the book. When several layers attach to the same line, order them: level-1 entries first, then `propositions`, `theorems`, `problems`, `lemmas`, `added_propositions`, `additions`, `definitions`, `postulates`, `axioms`, all before any existing curated heading for the same item.

### Other layers

Also look for other structures with a clear, repeated pattern, such as `[Margin]` blocks with recurring topics, terms or figure callouts (`grep -n -A2 '\[Margin\]' <dir>/page-*/original.md | head -50`, sample rather than read everything). Recommend only layers with a clear, repeated pattern; do not list every one-off item.

### Numbered items

The `numbered_item` layer indexes printed numbered lines (`1.`, `II.`, `3)`) that no proposition-kind layer covers: sub-items inside a proposition, corollary or demonstration, numbered lists in prefaces, symbol tables, or a section with no typed layer (for example, postulates in a book with no `postulates` layer). An item that a proposition-kind layer (`definitions`, `postulates`, `axioms`, `propositions`, `theorems`, `problems`, `lemmas`, `additions`, `added_propositions`) already indexes must not also get a `numbered_item` entry. The more important sequence covers it.

- When recommending a `numbered_item` layer, leave out every line that already carries, or that you recommend for, a proposition-kind entry, and say how many you left out.
- When the book already has `numbered_item` entries, flag each one that duplicates a proposition-kind entry, and suggest removing it. A duplicate sits next to that entry, with only other curated headings in between, and has the same number (roman and arabic count as equal). It also counts as a duplicate when the proposition-kind entry has no printed number (`Partie`, `Ratio ex duabus rationibus`). `figures` and untyped curated headings never cover a numbered item. A `numbered_item` entry whose number differs from the adjacent entry is a sub-item. Keep it.
- When you recommend a new proposition-kind layer for lines that already have `numbered_item` entries, include removing those entries in the recommendation.

### What to write

For each recommended layer give the `type`, what it tracks, counts per book or section, the pages, the levels, and two to four sample lines with their page. Note numbering gaps or irregularities that affect a layer. Put layers that are not applicable or out of scope under "Considered, not recommended", with the reason and count. Do not insert anything; the user decides which layers to create.

## Step 6: Report

Always write the report to a dedicated markdown file, `ocrflow/store/data/curation_reports/<book>.md` (one file per book, overwritten on a re-run), unless the user names another path. Never write it inside the transcription dir. In the reply, give the file path and a short summary of the counts per category; do not paste the full report into the conversation.

The findings are one list per book, sorted by page in ascending order. Do not group or split them by section, category or finding type; the user applies fixes page by page. The one exception is a group of long headers (Step 2), which sits at the page of its first member. Each row: page, current line, suggested line, reason, and confidence (likely OCR / possible printer's error / formatting). For sequence flags, include what the sequence expects. Include only actionable items, whether or not they need a human check against the scan. Do not list flags you judged false positives. Put the curated index recommendations from Step 5 in the same report, in a `## Curated index recommendations` section after the findings list (summary table of layers, one subsection per layer, then "Considered, not recommended"). Do not write them to a separate file. On a re-run, replace that section rather than appending a second one.

## Applying approved layers

Only when the user explicitly asks to apply recommended layers may transcriptions change, and then only by inserting curated-heading lines:

- Insert lines only. Never change, reflow or delete existing text or whitespace; keep each file's line endings and trailing newline.
- The one exception: approved duplicate `numbered_item` removals (see "Numbered items"). Delete only the `numbered_item` line and one adjacent blank line. For those books, the deletions in `git diff --numstat` must equal twice the number of removals, and every deleted non-blank line must contain `type=numbered_item`.
- Use a script with a dry run that prints the planned insertions per layer; check the counts against the report before applying. Make it idempotent (refuse if a page already has `type=<layer>`).
- Verify afterwards: `git diff --numstat` shows 0 deletions for the book (apart from approved `numbered_item` removals); every added line matches `^\[Curated heading level=[1-9][0-9]* type=[A-Za-z0-9][A-Za-z0-9._-]*: [^\]]+\]$`; each layer's numbering runs as the report expects; and every `theorems`/`problems` entry sits before a printed line that names its kind (the explicit-only rule).
- Record what was applied, with counts and any deviations, in an `## Applied` subsection of the report's curated index section.
- When several books are applied in parallel, keep scripts in per-book folders so they do not overwrite each other.
