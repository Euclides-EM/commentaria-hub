# Basel_1537 — Transcription Overview

## Source

- File: `Basel_1537.pdf` (73,198,383 bytes, 601 pages)
- Work: *Euclidis Megarensis Mathematici Clarissimi Elementorum Geometricorum Lib. XV* — Latin Euclid in fifteen books, with the exposition of Theon in the first thirteen books (in the Latin of Bartholomaeus Venetus/Zamberti), of Campanus in all fifteen, and of Hypsicles of Alexandria in the last two
- Appended works: Phaenomena, Catoptrica (Specularia) & Optica (Perspectiva), Protheoria of Marinus, Data, and the *Opusculum de Levi et Ponderoso* ("hactenus non visum")
- Imprint: Basileae apud Iohannem Hervagium, mense Augusto, anno M.D.XXXVII, cum privilegio Caesareo
- Text collated for this edition against the Greek by Christannus Herlinus (Christian Herlin, professor of mathematics at Strasbourg), on the basis of the Paris edition produced under the direction of Jacques Lefèvre d'Étaples (Iacobus Faber Stapulensis)
- Prefaces: Johannes Hervagius to the reader; Philipp Melanchthon to young students ("ἀγεωμέτρητος οὐδεὶς εἰσίτω")
- Copy: Staatliche Bibliothek Bamberg (stamp on PDF page 3); Google Books digitization

## Contents by PDF page

| PDF pages | Content |
|---|---|
| 1 | Google Books notice sheet |
| 2 | Title page |
| 3 | Hervagius, letter to the reader |
| 4–8 | Melanchthon, epistle to students (Greek quotations incl. Homer) |
| 9 | Blank leaf |
| 10–15 | Front matter, definitions, opening of Book I (printed pagination begins: PDF page ≈ printed page + 9 in the early quires) |
| 16–52 | Book I (each proposition given twice: "Eucli. ex Camp." with Campanus's commentary, then "Eucli. ex Zamb." with Theon's proof in small italic and Greek point-letters) |
| 53–61 | Book II |
| 62–94 | Book III (opening heading misprinted "LIBER SECVNDVS" on printed p. 53) |
| 95–122 | Book IV |
| 123–160 | Book V (begins folio 114) |
| 161–211 | Book VI–VII region (foldout table leaf with blank verso at PDF 137) |
| 212–232 | Book VIII |
| 233–251 | Book IX |
| 252–310 | Book X, first part (running head misprinted "LIBER DECIMVS" already on fol. 241) |
| 311–398 | Book X, continued (irrational lines; propositions misnumbered in places, e.g. "75" for 73 on fol. 308) |
| 399–433 | Book XII (props from fol. 388) — Book XI precedes in the 350–398 range |
| 434–461 | Book XIII |
| 462–494 | Books XIV–XV (Hypsicles), ending with Zamberti's dated subscription (fol. 483) |
| 495–516 | Phaenomena; then Catoptrica/Specularia (fols 506–515) |
| 527–546 | Optica/Perspectiva with Zamberti's dedication (fols 516–535, Theoremata 1–57) |
| 547–556 | Protheoria of Marinus and prefatory letters to the Data (PDF 555–556 are duplicate scans of fols 542–543 = PDF 553–554) |
| 557–598 | Data (Theoremata/Propositiones 1–end; "Finis Datorum" mid-page on PDF 598) |
| 598–599 | *Euclidis de Levi et Ponderoso fragmentum* (definitions and theorems); Campanus additions |
| 600 | "Libri quarti additamentum" (nonagon construction), REGESTVM (collation register: "Omnes sunt terniones præter † qui est duernio"), colophon |
| 601 | Hervagius printer's device (three-faced Hermes term with caduceus) |

## Methodology

1. Every page was read visually by Claude (model: Claude Fable 5) directly from the PDF page images (rendered via the Read tool, max 5 pages per call) and transcribed by hand from the image. No OCR software (tesseract or similar), no text-extraction scripts, and no PDF text layer were used at any point.
2. Pages 1–5 were transcribed in the main session; pages 6–601 were transcribed by parallel Claude agents in batches of 5–13 consecutive pages, all following identical instructions. For the very small italic type of the Theon/Zamberti proofs, agents were permitted to render zoomed crops (sips/pdftoppm at 200–600%) purely as a visual reading aid.
3. Work spanned many sessions (18 July – 14 August 2026) across usage-limit resets; after each interruption the missing page ranges were detected from the directory contents and relaunched.

## Conventions

- One file per PDF page: `1.md` = PDF page 1 … `601.md` = PDF page 601.
- Original Latin orthography preserved exactly as printed: u/v and i/j as set, æ/œ ligatures, & for "et"; abbreviation marks rendered with combining marks (ā ē ī ō ū, q̃ for -que) where possible.
- Words hyphenated across line breaks joined; paragraphs written as continuous flowing text.
- Greek passages and Greek point-letters in Greek script. The italic Greek fount uses a "6"-shaped alternate beta, an ι-like epsilon and ξ-like zeta, and nearly identical η/κ sorts; letters were normalized by identity established from the diagrams.
- Printer's errors retained as printed; sufficiently certain and useful corrections use the adjacent `{printer-error-correction:CORRECTION}` syntax.
- The transcription uses the canonical Markdown dialect for page furniture, paired textual zones, described page objects, drop capitals, unreadable text, and blank pages.
- The "Digitized by Google" scan watermark was not transcribed.

## Known caveats

- In the small-italic Theon/Zamberti proofs, individual Greek point-letters at the limit of the scan resolution were resolved from the geometric context; residual letter-level uncertainty remains on some pages (flagged in the batch reports and, where material, in the files).
- The book's printed foliation drifts against the PDF page numbers (offsets of +9 to +13 across the volume) and is itself occasionally misprinted (e.g. "46" for 94, "180" for 106, "522" for 552); printed numbers are recorded as seen with editorial notes.
- PDF 555–556 duplicate the scans of PDF 553–554 (fols 542–543); both copies were transcribed with a duplicate-scan note.
