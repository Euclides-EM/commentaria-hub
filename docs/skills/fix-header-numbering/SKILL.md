---
name: fix-header-numbering
description: Find and fix gaps, misreads and missing headers in the numbered proposition/problem/theorem/chapter headers of an OCR transcription (ocrflow/store/data/transcriptions/<book>/page-NNNN/original.md). Use when asked to check header numbering, find missing props/theorems/problems/lemmas/chapters, or fix sequence gaps, with or without a generated index.json.
---

# Fix header numbering in transcriptions

Books like Euclid editions number every proposition twice: a running proposition number (`PROP. XXV.`, roman) and a per-kind counter (`THEOR. 22.` / `PROBL. 5.`, arabic), both restarting at each section (`# ELEMENT ...`). OCR breaks these in predictable ways. Goal: every header in a section forms an unbroken sequence, and nothing that should be a header is left as plain text.

## Inputs

- Transcription dir: `ocrflow/store/data/transcriptions/<book>/` with `page-NNNN/original.md`. Index page numbers equal the `NNNN` folder number.
- Optional `index.json` (the generated header tree: `nodes[]` with `category` `header1/2/3`, `content`, `location.page`, `children`). If the user mentions an index but did not give a path, ask for it. If none exists, work from the transcription dir directly; the script supports both and the dir scan is the source of truth after edits.

## Step 1: Detect

```
python3 docs/skills/fix-header-numbering/scripts/find_gaps.py <index.json | transcription dir>
```

It splits by level-1 header, then checks the PROP sequence, the THEOR and PROB counters separately, and `CHAP.` roman numbers. When scanning a dir it also reports lines that look like `THEOR. n. PROP. X.` but are not markdown headers. A flag after a bad value can be a knock-on effect: the script resyncs, but read each flag with its neighbours.

Also eyeball the level-2 outline for the section for headers the regexes skip (LEMME, COROLLAIRE numbering, definitions lists).

## Step 2: Classify each flag

The transcription must reflect the printed book, which has its own errors. Fix only what is plausibly an OCR mistake: a character misread as a visually similar one. Anything that could be a printer's error stays as printed and is reported instead. Look at the page text before deciding.

**Fix (likely OCR):**

1. **Character misreads in a number**: `S`/`1S`→8/18, `1I`/`I`/`l`→1, `D`→12, a dropped repeated digit (`1.` for 11, `3.` for 31), stray letters glued to a number (`59n.`→`59.`), a misplaced dot inside a number (`2.7`→`27.`). In roman numerals: `H` for `I`/`II` (`VHI`, `XHII`, `IHII`), `I` for `L` (`IXXX`→LXXX), `E` for `C`, lowercase `x`, and `IHE0` for `THEO`.
2. **Header present as plain text**: the header line exists but lacks `##`, or is split across lines (`THEOR. 7.` / blank / `PROP. IX.`). Join it onto one line and add the `##` prefix. Do not change its wording.

**Do not fix; report instead (could be in the original):**

- Wrong kind: `PROBL.` where the sequence expects `THEOR.`, or the reverse. OCR does not turn one word into the other.
- A number that is missing, rather than misread (`THEOR. PROP. XXV.`).
- Roman numerals that need a character added, or are transposed (`XLVII` where XLVIII is expected, `CXIII` vs `XCIII`). The book may repeat or misprint a number.
- A header that is missing entirely. Do not insert header text that isn't in the transcription.
- A missing section header: numbering restarts at 1–2 mid-section. Report it, and check whether the preceding page's `original.md` is empty.
- Punctuation and spacing (`THEOR22.`, `PROP..`, `PROP,`, `THEOR.15.` with no space). Keep them as they are.

## Step 2b: Apply the user's verdicts on report-only items

Once the user has checked a flagged item against the scan (page images are at `ocrflow/store/data/<dataset_id>/imgs/page-NNNN.png`; `dataset_id` is in `index.json`), mark it according to `docs/MARKDOWN_DIALECT.md`:

- **OCR error**: correct the text in place.
- **Printer's error**: keep the printed reading and append the correction to the wrong token: `## PROBL{printer-error-correction:THEOR}. 20. PROP. XXII.`, `PROP. CXIII{printer-error-correction:XCIII}.`. If something was left out in print, attach the correction to the token before the gap: `## THEOR.{printer-error-correction:THEOR. 25.} PROP. XXV.`
- **Header not in print**: insert a curated heading in the `missing-headers-in-print` layer, with a blank line before and after it, just before the statement: `[Curated heading level=2 type=missing-headers-in-print: THEOR. 32. PROP. XXXIIII.]`
- **Page transcription empty**: transcribe it from the page image, following `docs/MARKDOWN_DIALECT.md`. Do it only when the user asks.

The script resolves `{printer-error-correction:...}` to the corrected value and counts curated headings, so marked items no longer show up as flags.

## Step 3: Fix

- Change only the misread characters. Keep the surrounding punctuation, spacing and label spelling exactly as they are. Keep other text, whitespace and blank lines untouched (project CLAUDE.md rules). Use exact-string replacements that assert one match per file.
- Files contain combining characters (`a` + U+0300 for `à`, `e` + U+0301 for `é`) and `ſ`; prefer match keys that avoid accented letters, or copy them from the file.
- Do not touch running titles (`<!-- Running title -->`) or page-number comments.

## Step 4: Verify

Re-run the script on the transcription dir. Remaining flags must each be explained (e.g. first chapter has no number, empty source page). Report to the user: fixes grouped by type with page numbers. List the report-only items separately, with what the sequence expects, so someone can check them against the scans.
