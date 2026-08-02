<!-- # Transcription Conventions -->

This directory contains a page-by-page transcription of Pierre Hérigone, *Les six premiers livres des Elements d'Euclide, demonstrez par Notes, d'vne methode tres-brieve & intelligible* (Paris, chez l'Autheur & Henry Le Gras, 1639) — the French-only volume giving Euclid I–VI in Hérigone's symbolic notation, followed by practical treatises "sans Notes" (arithmetic, trigonometry, practical geometry, fortifications, gnomonics) and a dictionary of mathematical terms. The source is the Gallica/BnF scan (482 PDF pages). Files are named `<N>.md` where N is the **PDF page number** (7–475). PDF 1–6 (Gallica title cards, front cover, pastedown, flyleaves) and 476–482 (back flyleaf, pastedown, cover, spine and edge photographs) are not transcribed. The printed page number is N − 10 (main text runs PDF 11 = printed 1 through PDF 473 = printed 463; PDF 474 is the unnumbered errata leaf and 475 its blank verso). The transcription was produced by visually reading each page image (150 DPI renders, with 300 DPI zoomed crops to verify doubtful readings) — no OCR software was used.

Cross-references of the form `t. N. p. NNN.` (and `alg.` for the algebra) cite the tome and page of Hérigone's *Cursus mathematicus*; the propositions of the Gnomonique carry heads like `Propos. 1. pag. 750. du 5`, citing tome 5.

## File structure

- `# Page <N>` — PDF page number.
- `# <running head>` — exactly as printed (`# LES ELEMENTS` versos / `# D'EVCLIDE, LIV. N.` rectos; `# ARITHMETIQVE` / `# PRACTIQVE.`; `# TRIGONOMETRIE.`; `# GEOMETRIE` / `# PRACTIQVE.`; `# DES FORTIFICATIONS.`; `# DE LA GNOMONIQVE.`; `# ETYMOLOGIE.`), spelling variants preserved.
- `## ...` is reserved for the few large section headings (`## DE L'ART D'ASSAILLIR.`, `## DE LA DEFENSE.`, `## DES QVADRANS ITALIQVES, Babyloniques & antiques.`, `## Methode vniuerselle & facile de descrire vn quadrant Italique en tout plan qui ne soit parallele à l'horizon.`, `## Etymologie & explication des noms & termes plus obscurs des Mathematiques.`, `## Fautes à corriger dans l'impression.`).
- Proposition headings are plain lines as printed (`THEOR. I. PROPOS. I.`, `PROBL. II. PROPOS. XIV.`), followed by the French enunciation.
- Proof-section keywords in italics as printed: `*Hypoth.*`, `*Constr.*`, `*Preparation.*`, `#### Req. à demonſtr.`, `*Demonstr.*`, `### SCHOLIE.` lines.
- Symbolic proof lines are rendered `margin citation | body`, one printed line per transcript line, e.g. `α.47.1 | □.ac 2|2 □.ad + □.dc,`. Brace-grouped multi-line statements and stacked double margin citations are flattened onto one line with ` / ` between the printed lines (e.g. `α | □ak, □kf, / ſ.46.1 | □ci, □hg, *ſnt* 2|2 đe.`).
- Diagrams are rendered `*[Figure]*` (label letters are not transcribed); printer's ornaments `*[Ornament]*`; blank pages `*[blank page]*`.
- Printed tables (trigonometric examples, fortification calculations, the Maurolicus hour table on PDF 457, the errata leaf) are rendered as Markdown tables with the printed column heads.
- Fractions are transcribed `N M/D` (`5 1/2`, `84 944/1000`); sexagesimal/decimal primes kept (`36 deg. 37″.`, `57009792″`); the rule-of-three chains and logarithm columns follow the printed line breaks.
- Italic type is wrapped in `*asterisks*`. In the Etymologie each entry is one paragraph; the Greek/Latin etyma printed in italics are asterisked, roman glosses are not, and the closing tome/page citation is set off with a double space (`...vne mesure.  t. 3. p. 99.`).
- Greek letters used in figures and proofs (`α β γ δ`, `ε` as a figure point) are transcribed as printed.
- Catchwords and signature marks (`Dd ij`, `Ee iij`, …) are omitted.

## Typography

- Long s (ſ) is normalized to `s` in ordinary French prose, but **preserved** in the symbolic proof lines, in the margin citations (`ſ.46.1`, `3.ſ.1.d.2`), and in the Explication des notes abbreviations (`demonſtr.`, `propoſ.`, `mſur:`, `ſnt`).
- Accents, spellings, and abbreviation periods (including double periods like `part..`, `multipl..`, `raõ..`) are kept exactly as printed.
- Printed tildes and macrons are kept (`grādeur`, `prōpose`, `purgeoiẽt`, `Septẽtrion`, `contiennét`).
- Printing errors are **not** corrected: wrong sorts and misprints are kept as printed (e.g. `Trigonomerrie`, `Spqeroïde`, `Icnograghie`, `conienctiue`, `hydranlique` with turned u, `grapha`, `Iuilet`, `siguifie`, `petcer`, `faudta`, `foretresse`, `Ayanr`, `posirion`, `table suiuanre`, `qni`, `dan s la ville`), as are stray or missing punctuation marks (`de. laquelle`, `t. 5. p 801`, `t 1. p. 649.`, `def 33`, `t. 2, p. 75. alg.`, a raised dot `vne mesure·`) and wrong numerals in tables.
- Under-inked, broken, or battered letters are transcribed as the letter **intended** (a battered `n` sort makes `vient` print like `vieat` on PDF 472–473; an unprinted r-arm affects `derniere`/`Grec` on PDF 470).

## Hérigone's special symbols

The book's own legend is on PDF 14–16 (*Explication des Notes*); the citation system is explained on PDF 17 (*Explication des Citations*).

| Symbol | Meaning |
|---|---|
| `π` | "à" — separates the terms of a ratio (`a π b 2|2 c π d` = A : B :: C : D) |
| `2\|2` | is equal to |
| `3\|2` | is greater than |
| `2\|3` | is less than |
| `+` | plus |
| `~` | minus; `.~:` difference |
| `γ.` | racine ou costé (root or side) |
| `∠`, `<` | angle (`∠bag`); with a numeral, a polygon: `5∠`/`5<` pentagon, `6∠` hexagon |
| `⌐` | right angle |
| `△` | triangle |
| `□` | square (`□.ac` = square on AC) |
| `▭` | rectangle |
| `◊` | parallelogram |
| `⊙` | circle |
| `Ɔ`, `∪` | circumference |
| `∩`, `⌒` | arc |
| `⌓` | segment of a circle |
| `==` | is parallel to |
| `⊥` | is perpendicular to |
| `•` | is a point |
| `——` | is a straight line |
| `Ⅱ` | vel / ou |
| `đe`, `ꝗe` | entr'elles, entr'eux (mutually) |
| `;` | marks the plural |
| `..` | de (of) — `multd..part.. a` |
| `D.` | donné (the given) |
| `req.` | the thing required |
| `α β γ δ` | mark intermediate conclusions, cited later in the margin (`d. α`) |

## Margin citations

- `47. 1` — Proposition 47 of Book I.
- `15. d. 1` — Definition 15 of Book I (`d` = definition); `2. a. 1` — Axiom 2 (`a`), added axioms lettered (`1. a. f`); `3. p. 1` — Postulate 3 (`p`).
- `c. 15. 4` — Corollary of Prop. 15 Book IV; `ſ.46.1` — Scholium of Prop. 46 Book I.
- `hyp.` / `ſuppoſ.` — by hypothesis; `conſtr.` — by construction; `arbitr.` — arbitrarily chosen; `ſymp.` — symperasma; `d. α` — by the earlier conclusion marked α; `nota α.27.3` and compound cites joined as printed; stacked cites joined with ` / `.
- The practical treatises cite the *Cursus* (`t. 3. p. 131.`, `au 2 & 3 chapitre de la Geometrie practique du 3 tome`) and the Elements (`par la 6 du 1 des elem.`).

## Layout of the files

7 title page; 8 blank; 9 *Annotation sur l'Altimetrie*; 10 *Annotation sur la Gnomonique*; 11–13 De la division des Mathematiques & Prolegomenes; 14–16 Explication des Notes; 17 Explication des Citations; 18–101 Elements Book I; 102–130 Book II; 131–174 Book III; 175–201 Book IV; 202–245 Book V; 246–290 Book VI (290 closes with Hérigone's note introducing the practical treatises); 291–365 Brief traicté de l'Arithmetique practique; 366–375 De la Trigonometrie (De l'Altimetrie begins mid-page 375); 376–396 Geometrie practique (381 Regle generale pour l'vsage; 383 Annotations; 386 De l'Epipedometrie ou Planimetrie; 393 De la Stereometrie; 396 Fin de la Geometrie); 397–442 Des Fortifications (429 De l'art d'assaillir; 439 De la defense); 443–459 De la Gnomonique, ou Horologeographie (453 Des quadrans Italiques, Babyloniques & antiques; 456 Methode vniuerselle; 457 hour table; 459 *Des quadrans Antiques.*); 460–473 Etymologie & explication des noms & termes plus obscurs des Mathematiques (ends `F I N.`); 474 Fautes à corriger dans l'impression (errata table); 475 blank.
