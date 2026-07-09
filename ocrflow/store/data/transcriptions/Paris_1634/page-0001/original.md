# Transcription Conventions

This directory contains a page-by-page transcription of Pierre Hérigone, *Cursus Mathematicus* (1634) — the bilingual Latin/French edition of Euclid's Elements in Hérigone's symbolic notation. Files are named `<N>.md` where N is the **PDF page number** (11–882). The printed page number is N − 82 (e.g. `882.md` = printed page 800). The transcription was produced by visually reading each page image (150 DPI renders; 300 DPI + deskew for skewed photos) — no OCR software was used.

## File structure

- `# Page <N>` — PDF page number.
- `# <running head>` — exactly as printed, e.g. `# ELEM.. EVCLID. LI. XV.` (double periods and spelling variants preserved).
- `## THEOR. N. PROPOS. N.` / `## PROBL. N. PROPOS. N.` — proposition headings.
- Latin enunciation in plain text; the French translation follows in italics (`*...*`).
- Section keywords as plain lines: `Hypoth.`, `Conſtr.`, `Præpar.`, `Req. π. demonſtr.`, `Demonſtr.`
- Symbolic proof lines are rendered `margin citation | body`, one printed line per transcript line, e.g.
  `47. 1 | □.bc 2|2, 2□.bi,`
- Brace-grouped multi-line statements and two-line proportion displays are flattened onto one line with ` / ` between the printed lines. A stacked double margin citation is likewise joined with ` / ` (e.g. `1. 14 / 15. 5 | ...`).
- Two-column Latin|French prose paragraphs are rendered `Latin | *French*` on one line; when a margin citation sits between the columns: `Latin α | α. cite | *French α*`.
- Diagrams are rendered `(figure)` followed by the label letters row by row, rows separated by ` / `, separate diagrams on the same page separated by ` — `. Example: `(figure) B / H G / F / A C / D / E`.
- Printer's ornaments are rendered `(ornament)`.
- Catchwords and signature marks (e.g. `Ddd iij`) are omitted.

## Typography

- Long s (ſ) is normalized to `s` in ordinary Latin/French prose, but **preserved** in the symbolic proof lines and abbreviations: `ſnt`, `ſymp.`, `ſuppoſ.`, `conſtr.`, `ſ.46.1`, `ſphær.`, `ſuperfic..`, `ſemidiamet..`, `ſectr.`, `ſemic.`, `circſcri.`, `inſcri.`, and in French italic passages.
- Period spelling, accents, and abbreviation periods (including inconsistent single/double periods like `pyram..`, `diamet.` vs `diamet..`) are kept exactly as printed.
- Latin macron abbreviations preserved: `cubū`, `cētris`, `comprēd`, `pētag..`
- The book mixes a Latin `k` and a Cyrillic-looking `к` glyph, sometimes within one line (`nk 2|2 ko` vs `hк, кi`); each occurrence is transcribed as the glyph actually printed.
- Printing errors and anomalies are **not** corrected (e.g. `docaedr.`, `ratiöne`, `parellelogramme`, `incommensubilis`, wrong margin citations, Latin/French mismatches in enunciations).

## Hérigone's special symbols

| Symbol | Meaning |
|---|---|
| `π` | "to" — separates the terms of a ratio (a π b = a : b) |
| `2\|2` | is equal to |
| `3\|2` | is greater than |
| `2\|3` | is less than |
| `+` | plus |
| `~` | minus (`eg~ef`) |
| `□.ab` | square on line ab; `□` alone = a square (figure) |
| `▭.cd, fg` | rectangle contained by cd and fg |
| `□.h, ef` | rectangle of h and ef (square sign also used for rectangles of two lines) |
| `γ.` | side (latus) of — `γ. cub.` = side of the cube |
| `<` | angle (`<bag`); also polygon with a numeral: `5<` pentagon, `10<` decagon |
| `⌐` | right angle (`eſt ⌐,`) |
| `△` | triangle |
| `⊙` | circle (`⊙abh`) |
| `∩` | arc (`∩bcd`) |
| `⌓` | segment of a circle |
| `——` | straight line(s) (drawn/joined) |
| `==` | is parallel to |
| `⊥` | is perpendicular to |
| `•` | point (`e, eſt • ꝗn ae` = e is a point on ae) |
| `Ⅱ` | vel / or |
| `ꝗe` | inter se (mutually, to one another) |
| `ꝗn` | in (`inſcri. ꝗn ſphær.`) |
| `&c.` | et cetera |
| `½ ⅓.. ⅕.` | fractions, with the book's trailing periods kept |
| `2□.ad`, `3□.ab`, `6□.bi` | numeric multiples of a square |
| `α β γ δ ε` | Greek letters marking intermediate conclusions, referenced later in the margin (`d. α` = by statement α) |
| `D.` | datum — the given magnitude/figure (`eſt cub. D.`) |
| `req.` | quæsitum — the thing required |
| `commun.` | common (shared element in a proof step) |

## Margin citations

The left margin cites the justification for each proof line:

- `47. 1` — Proposition 47 of Book I (prop. book).
- `3. p. 1` — Postulate 3 of Book I (`p` = postulatum); `1. p. 1` = Post. 1 Book I.
- `2. a. 1` — Axiom (communis notio) 2 of Book I (`a` = axioma); variants `7.a.1`, `1. a. f`.
- `29. d. 1` — Definition 29 of Book I (`d` = definitio); spacing varies: `29.d.1`, `1. d. 3`, `2.d.6`.
- `c. 15. 4` — Corollary of Prop. 15 Book IV (`c` = corollarium); variants `c.10.13`, `c.4.14`.
- `ſ.46.1` — Scholium of Prop. 46 Book I.
- `5. app.` — Proposition 5 of the Appendix (after Book VI).
- `concl. 11. 5` — the conclusion step, justified by 11. 5.
- `ſuppoſ.` — by supposition/hypothesis; `conſtr.` — by construction; `arbitr.` — arbitrarily chosen; `ſymp.` — symperasma (statement of what results/is required); `d. α` — by the earlier conclusion marked α.
- Compound cites joined as printed: `1&2.p.1`, `5.&6 12`, `7.a.1, & 4. 1`; stacked cites joined with ` / `.

## Book layout of the files

11–14 Ad Lectorem; 15–21 Prolegomena; 22–29 Explicatio Notarum; 30–33 Explicatio Citationum; 34–139 Book I; 140–182 Book II; 183–229 Book III; 230–264 Book IV; 265–332 Book V; 333–384 Book VI; 385–430 Appendix; 431–491 Book VII; 492–527 Book VIII; 528–567 Book IX; 568–726 Book X; 727–780 Book XI; 781–823 Book XII; 824–863 Book XIII; 864–874 Book XIV; 875–882 Book XV.
