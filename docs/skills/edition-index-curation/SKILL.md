---
name: edition-index-curation
description: Report suggested fixes for the headers of an OCR transcription (ocrflow/store/data/transcriptions/<book>/page-NNNN/original.md) - numbering gaps in proposition/problem/theorem/chapter sequences, wrong header levels, OCR misreads, printer's errors, inconsistent case and stray spaces. Report only - never edits transcriptions. Use when asked to curate, check or review headers or header numbering, with or without a generated index.json.
---

# Edition index curation (report only)

This skill produces a report of suggested header fixes. It never changes a transcription: do not edit, write, or run any command that modifies files under `ocrflow/store/data/transcriptions/`, even for fixes that look certain. The user applies the suggestions they accept.

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

## Step 3: Classify each suggestion

The transcription must reflect the printed book, which has its own errors. Separate what is plausibly an OCR mistake (a character misread as a visually similar one) from what could be in the print. Grep the line before deciding.

**Likely OCR error:**

1. Character misreads in a number: `S`/`1S`→8/18, `1I`/`I`/`l`→1, `D`→12, a dropped repeated digit (`1.` for 11, `3.` for 31), stray letters glued to a number (`59n.`→`59.`), a misplaced dot inside a number (`2.7`→`27.`). In roman numerals: `H` for `I`/`II` (`VHI`, `XHII`, `IHII`), `I` for `L` (`IXXX`→LXXX), `E` for `C`, lowercase `x`, and `IHE0` for `THEO`.
2. Character misreads in words: split or doubled letters (`LE M ME`, `ROLLAAIRE`), stray lowercase in an uppercase header (`cO`), a glyph read as a similar one.
3. Header present as plain text: the header line exists but lacks `#`, or is split across lines (`THEOR. 7.` / blank / `PROP. IX.`).
4. Wrong level: the header text is right but its `#` count does not match its siblings.

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
- Printer's error: keep the printed reading and append the correction to the wrong token: `## PROBL{printer-error-correction:THEOR}. 20. PROP. XXII.`, `PROP. CXIII{printer-error-correction:XCIII}.`. If something was left out in print, attach the correction to the token before the gap: `## THEOR.{printer-error-correction:THEOR. 25.} PROP. XXV.`
- Header not in print: a curated heading in the `missing-headers-in-print` layer, placed just before the statement with a blank line before and after it: `[Curated heading level=2 type=missing-headers-in-print: THEOR. 32. PROP. XXXIIII.]`
- Page transcription empty: say so; do not transcribe it.

Page images for checking are at `ocrflow/store/data/<dataset_id>/imgs/page-NNNN.png` (`dataset_id` is in `index.json`). Mention the image path for every possible printer's error.

## Step 5: Report

Always write the report to a dedicated markdown file, `ocrflow/store/data/curation_reports/<book>.md` (one file per book, overwritten on a re-run), unless the user names another path. Never write it inside the transcription dir. In the reply, give the file path and a short summary of the counts per category; do not paste the full report into the conversation.

The report is grouped by section, then by category (sequence gaps, levels, OCR misreads, possible printer's errors, formatting). Each row: page, current line, suggested line, reason, and confidence (likely OCR / possible printer's error / formatting). For sequence flags, include what the sequence expects. List flags you judged false positives at the end, with a one-line reason, so a re-run can be read quickly.
