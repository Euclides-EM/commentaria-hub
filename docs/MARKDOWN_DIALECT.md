# Transcription Markdown Dialect

This document defines the canonical Markdown representation for page-level transcriptions of early modern printed editions of Euclid's *Elements*.

The specification is normative and is used as part of the prompt for LLM-based transcription and transcription correction.

A transcription MUST use only the constructs defined in this specification. Do not invent additional annotation syntax.

<!-- BEGIN LLM CONTRACT -->
## LLM operational contract

This bounded section is the compact normative contract embedded in automated LLM
transcription and correction prompts. The remainder of this document provides the
full specification, rationale and examples for people and development tools.

### Output and evidence

* Return only the complete Markdown transcription for the current page. Do not
  return commentary, explanations, confidence statements, correction summaries,
  boundary labels or a surrounding code fence.
* Treat the current page image as authoritative. Use candidate transcriptions only
  as evidence. Do not copy text from previous-page context into the current page.
* Transcribe only content visible on the current page. Do not infer content from
  adjacent pages, other copies or editions, bibliographical data or prior knowledge.
* Preserve historical spelling, capitalization, punctuation, Greek, diacritics,
  macrons, `u/v`, `i/j`, ligatures, long `ſ`, abbreviations, printer's errors and
  reliably representable Unicode characters. Do not modernize, expand, translate,
  normalize or silently correct them.
* Preserve paragraph boundaries, not ordinary printed line breaks. Join lines in
  the same paragraph and join words divided solely by line wrapping, omitting only
  the line-end hyphen introduced by that wrapping. Retain lexical hyphens.
* Separate distinct blocks with one blank line.

### Textual structure and page furniture

Ordinary body text is plain Markdown. Printed main-text headings use `# TEXT` for
book-level headings, `## TEXT` for section or proposition headings, `### TEXT` for
subordinate headings, and deeper levels only when the edition's structure requires
them. Use consistent levels within an edition. Never use headings for page furniture.

Page furniture uses exactly one HTML comment per visible item:

```markdown
<!-- Running title: TEXT -->
<!-- Page number: TEXT -->
<!-- Signature: TEXT -->
<!-- Catchword: TEXT -->
```

Transcribe each value diplomatically. Do not calculate or infer absent page numbers
or signatures. Keep distinct furniture items in separate comments and preserve
their page order. Put catchwords after main text and page objects.

### Inline annotations

A drop capital is immediately adjacent to the word it begins:

```text
{dropcap:TEXT|lines=N|style=STYLE}
```

`TEXT`, `lines` and `style` are required. `lines` is the visible number of lines or
`?`; `style` is `plain`, `decorated` or `unknown`. Only a decorated drop capital may
add `|decoration="DESCRIPTION"`. Do not infer `plain` when decoration is uncertain.

Retain a printer's erroneous reading. When a correction is sufficiently certain and
useful, append it immediately after that reading:

```text
PRINTED{printer-error-correction:CORRECTION}
```

Do not mark unfamiliar historical spelling or abbreviations as printer's errors.

### Unreadable text, zones and page objects

Use `[illegible]` when visible text has no reliable reading. Add a reliable extent
as `[illegible: N chars]` or `[illegible: N words]`. Use `[unclear: TEXT]` when a
specific reading is possible but uncertain. Never conjecturally supply unreadable
text.

Textual zones use paired, case-sensitive annotations:

```markdown
[Margin]
TEXT
[/Margin]

[Footnote]
TEXT
[/Footnote]

[Handwritten]
TEXT
[/Handwritten]

[Other]
TEXT
[/Other]

[Other type="TYPE"]
TEXT
[/Other]
```

Do not preserve ordinary printed lineation inside textual zones. `[Margin]` is for
printed marginal text; `[Handwritten]` is for handwritten text. Use a specific zone
whenever possible. `[Other]` is only for a distinct relevant zone with no defined
type, may be empty, and must not replace ordinary text or a difficult classification.

Non-textual page objects occur on their own line at the corresponding position.
Each object may optionally include a concise description after a colon:

```markdown
[Diagram]
[Diagram: six horizontal lines labelled A, a, B, b, C and c]
[Figure]
[Figure: armillary sphere]
[Illustration]
[Illustration: Euclid teaching two students]
[Ornament]
[Ornament: floral headpiece]
```

Descriptions record visible features of the object and may include transcribed
labels. Do not reconstruct diagrams as text or mathematical notation. A decorated
drop capital uses drop-cap syntax rather than `[Ornament]`.

Spatial calculations use a paired block whose internal spacing and line breaks may
be preserved:

```markdown
[Calculation]
CONTENT
[/Calculation]
```

Use an empty calculation block when the calculation is visible but cannot be
represented reliably. Do not modernize it or convert it to LaTeX. Transcribe other
mathematical notation diplomatically with Unicode or plain text when reliable;
geometrical labels such as `A`, `AB` and `ABC` are ordinary text.

Use `[Blank page]` only when no transcribable printed or handwritten content is
present. Do not transcribe scanner overlays, digital watermarks, automatically added
page numbers or other digitization artefacts.

### Syntax constraints

* Do not invent annotation names or vary their capitalization or spelling.
* Curly braces are only for defined inline features attached to text.
* Square brackets are only for defined zones, page objects, unreadable text and the
  blank-page marker.
* HTML comments are only for the four defined page-furniture types above.
* Prefer a defined specific type over `[Other]`; use `[Other]` rather than inventing
  syntax when a distinct zone has no defined type.
* Printed tables use Markdown pipe-table syntax. Preserve their textual content,
  row order and column order; do not infer missing values.
* Any feature without defined syntax receives no invented syntax; use
  `[Other]...[/Other]` only when it constitutes a distinct relevant zone.
<!-- END LLM CONTRACT -->

## 1. General principles

### 1.1 Scope

Transcribe only content visible on the current page.

Do not infer or supply text from adjacent pages, other copies, other editions, bibliographical records or prior knowledge of the work.

### 1.2 Diplomatic transcription

Transcribe textual content diplomatically.

Preserve as printed:

* historical spelling;
* capitalization;
* punctuation;
* Greek characters;
* diacritics;
* macrons;
* `u/v` and `i/j`;
* ligatures such as `æ` and `œ`;
* long `ſ`;
* abbreviation marks and abbreviated forms, such as `q;`, `cõ-` and `-ũ`;
* printer's errors;
* other special characters that can be represented reliably in Unicode.

Do not silently expand abbreviations, modernize spelling, normalize characters or correct the printed text.

### 1.3 Paragraphs and line breaks

Preserve paragraph boundaries, but do NOT preserve ordinary printed line breaks.

Text belonging to the same paragraph MUST be transcribed as a continuous Markdown paragraph, regardless of its lineation on the printed page.

For example, a paragraph printed as:

```text
Soit le triangle ABC, duquel
les deux costez AB, AC sont
égaux entr'eux.
```

MUST be represented as:

```markdown
Soit le triangle ABC, duquel les deux costez AB, AC sont égaux entr'eux.
```

Separate distinct paragraphs with one blank line.

Do not insert a paragraph break merely because of vertical spacing or a printed line break.

When a word is divided across a printed line solely because of line wrapping, join the divided parts and omit the line-end hyphen used for that division.

For example, printed:

```text
La demonstra-
tion precedente
```

becomes:

```markdown
La demonstration precedente
```

Do not remove a hyphen that belongs lexically or semantically to the printed word rather than functioning only as a line-end division mark.

### 1.4 Editorial intervention

The transcription MUST NOT contain commentary, explanations, observations, confidence statements, correction summaries or other material that is not part of the transcription or explicitly permitted by this specification.

Do not wrap the complete transcription in a Markdown code fence.

## 2. Syntax model

The dialect uses different syntactic forms for different kinds of information.

### 2.1 Markdown syntax: textual structure

Standard Markdown syntax represents the logical structure of the printed text.

Examples include headings and paragraphs.

### 2.2 HTML comments: page furniture

HTML comments represent textual elements belonging to the physical organization of the page but excluded from the main textual reading.

```markdown
<!-- Running title: LES ELEMENTS -->
<!-- Page number: 86 -->
<!-- Signature: B2 -->
<!-- Catchword: baid -->
```

Only the page-furniture annotation types defined in this specification may be used.

### 2.3 Curly-brace annotations: inline textual features

Curly-brace annotations represent features attached directly to particular transcribed text.

```markdown
{dropcap:P|lines=2|style=plain}Ropositis
```

```markdown
triangulun{printer-error-correction:triangulum}
```

Inline annotations MUST occur immediately adjacent to the text to which they apply.

### 2.4 Square-bracket annotations: zones and objects

Square-bracket annotations represent distinct page zones or non-textual objects.

A non-textual object may use a single annotation, optionally with a concise
description after a colon:

```markdown
[Diagram]
[Diagram: circle ABC with diameter AC]
```

A textual zone uses paired opening and closing annotations:

```markdown
[Margin]
Scholium.
[/Margin]
```

When a zone cannot be classified using one of the specific zone types defined in this specification, use an `Other` zone as defined below.

## 3. Main text

Represent ordinary body text as plain Markdown paragraphs.

Do not preserve ordinary printed line breaks. Join consecutive printed lines belonging to the same paragraph into one continuous Markdown paragraph.

Separate distinct paragraphs with one blank line.

```markdown
Soit donné le triangle ABC duquel les costez AB, AC sont égaux.

Il faut demonstrer que les angles ABC, ACB sont aussi égaux.
```

## 4. Headings

Printed headings belonging to the main text use Markdown headings.

Use:

* level 1 (`#`) for book-level headings;
* level 2 (`##`) for section or proposition headings;
* level 3 (`###`) for subordinate headings;
* subsequent levels only when required by the structure of the edition.

Use the same heading level for headings of the same structural type within an edition.

```markdown
# PREMIER LIVRE DES ELEMENTS D'EVCLIDE.

## THEOR. I. PROPOS. I.
```

Do not use Markdown headings for running titles or other page furniture.

## 5. Page furniture

Page furniture is represented using HTML comments.

Transcribe the text of page furniture diplomatically.

### 5.1 Running titles

```markdown
<!-- Running title: LES ELEMENTS -->
```

### 5.2 Page and folio numbers

```markdown
<!-- Page number: 86 -->
```

Transcribe the printed value. Do not calculate, normalize or infer a page number that is not printed.

### 5.3 Signatures

```markdown
<!-- Signature: B2 -->
```

Transcribe only the signature actually visible on the page. Do not infer a missing signature from the structure of the gathering.

### 5.4 Catchwords

```markdown
<!-- Catchword: baid -->
```

Preserve the catchword as printed.

### 5.5 Combined page furniture

Represent each page-furniture element separately.

```markdown
<!-- Running title: Des dritt buͤchs Euclidis -->

<!-- Page number: 86 -->
```

Do not combine multiple page-furniture types into one annotation.

## 6. Drop capitals

Represent a drop capital using:

```text
{dropcap:TEXT|lines=N|style=STYLE}
```

The permitted fields are:

* `dropcap`: REQUIRED. The character or characters printed as the drop capital.
* `lines`: REQUIRED. The number of printed lines occupied by the drop capital, or `?` when it cannot be determined reliably.
* `style`: REQUIRED. One of `plain`, `decorated` or `unknown`.
* `decoration`: OPTIONAL. A short description of the decoration. It may be used only when `style=decorated`.

Examples:

```markdown
{dropcap:P|lines=2|style=plain}Ropositis duabus lineis inæqualibus
```

```markdown
{dropcap:T|lines=2|style=decorated|decoration="floral"}HE writing of Characters
```

```markdown
{dropcap:O|lines=?|style=plain}Mnium duorum triangulorū
```

```markdown
{dropcap:P|lines=3|style=unknown}Ropositio
```

Do not infer that a drop capital is plain merely because no decoration can be identified.

## 7. Printer's errors and corrections

Printer's errors MUST be retained exactly as printed.

When a printer's error has a sufficiently certain correction and recording that correction is useful, append:

```text
{printer-error-correction:CORRECTION}
```

immediately after the erroneous printed text.

```markdown
Soit le triangulun{printer-error-correction:triangulum} ABC.
```

The text outside the annotation is the printed reading. The value inside the annotation is the correction.

Printer-error annotations are optional. Do not introduce a correction merely because a historical spelling, abbreviation or expression appears unfamiliar.

## 8. Illegible and uncertain text

### 8.1 Illegible text

When textual content is visible but cannot be read:

```markdown
[illegible]
```

When its extent can be established reliably:

```markdown
[illegible: 3 chars]
```

```markdown
[illegible: 2 words]
```

Do not conjecturally supply unreadable text.

### 8.2 Uncertain readings

When a specific reading can be proposed but remains uncertain:

```markdown
[unclear: triangulum]
```

Use `[illegible]` when no reliable reading can be proposed. Use `[unclear: TEXT]` when a specific reading is possible but uncertain.

## 9. Marginal text

Printed marginal text is represented as:

```markdown
[Margin]
Scholium.
[/Margin]
```

Do not preserve ordinary printed lineation inside the margin. Preserve paragraph boundaries where present.

Do not use `[Margin]` for handwritten annotations.

## 10. Footnotes

Represent a printed footnote as:

```markdown
[Footnote]
Hoc est triangulum æquilaterum.
[/Footnote]
```

Do not preserve ordinary printed lineation within the footnote.

[Syntax for footnote markers and links between markers and footnotes to be determined.]

## 11. Handwritten annotations

Represent handwritten textual material as:

```markdown
[Handwritten]
transcribed handwritten text
[/Handwritten]
```

Do not incorporate handwritten additions silently into the printed text.

[Additional syntax for placement and other handwritten features to be determined.]

## 12. Other zones

Use an `Other` zone for a distinct textual or page zone that is visible and relevant to the transcription but cannot be represented using a more specific zone type defined by this specification.

For textual content:

```markdown
[Other]
transcribed content
[/Other]
```

When useful, provide a short type or description:

```markdown
[Other type="privilege"]
transcribed content
[/Other]
```

```markdown
[Other type="label"]
transcribed content
[/Other]
```

An `Other` zone may also be empty when the presence of the zone should be recorded but its contents cannot or should not be transcribed:

```markdown
[Other]
[/Other]
```

Use an existing specific zone type whenever one applies. For example, do not use `[Other]` for a margin, footnote, diagram, ornament or calculation.

Do not use `Other` merely because the nature of ordinary text is uncertain. It represents a distinct zone or object whose type is not otherwise covered by the dialect.

## 13. Figures, diagrams, illustrations and ornaments

Place non-textual object annotations on their own line at the corresponding position in the transcription.

Each annotation MAY contain a concise description using the form
`[TYPE: DESCRIPTION]`. The description records visible features of the object and
may include transcribed labels. It MUST remain on the same line as the annotation
and MUST NOT contain a closing square bracket (`]`). Omit the description when a
reliable or useful description cannot be given.

### 13.1 Geometrical diagrams

```markdown
[Diagram]
```

```markdown
[Diagram: sechs waagerechte Linien, bezeichnet A (getheilet in c, d), a, C (getheilet in e, f, g), B (getheilet in h, k), b, D (getheilet in l, m, n); daneben die Ziffern I. II. V. III. IV. VI.]
```

Do not reconstruct the geometry as text or convert it into mathematical notation.

### 13.2 Figures

```markdown
[Figure]
```

```markdown
[Figure: armillary sphere on a pedestal]
```

### 13.3 Illustrations

```markdown
[Illustration]
```

```markdown
[Illustration: Euclid teaching two students]
```

### 13.4 Ornaments

```markdown
[Ornament]
```

```markdown
[Ornament: floral headpiece]
```

Decorative drop capitals use the drop-cap syntax rather than a separate `[Ornament]` annotation.

## 14. Mathematical notation

Transcribe mathematical notation diplomatically whenever it can be represented reliably using Unicode or plain text.

Do not convert historical mathematical notation into modern notation merely to express its mathematical meaning.

Do not automatically convert mathematical expressions into LaTeX.

Letters used as geometrical labels, such as `A`, `B`, `C`, `AB` or `ABC`, are ordinary transcribed text unless another rule applies.

[Additional rules for mathematical symbols, superscripts, fractions and other notation to be determined.]

## 15. Mathematical calculations

Represent a spatially structured calculation as:

```markdown
[Calculation]
12
 6
──
18
[/Calculation]
```

Unlike ordinary textual content, line breaks and spacing inside a `[Calculation]` block MAY be preserved when required to represent the calculation.

When a calculation cannot be represented reliably, leave the block empty:

```markdown
[Calculation]
[/Calculation]
```

An empty calculation block means that a calculation is visibly present but its internal content has not been transcribed.

Do not reconstruct, normalize or convert a calculation into modern mathematical notation merely to make it easier to represent.

## 16. Tables

Represent a printed table using Markdown pipe-table syntax:

```markdown
| Quantity | Value |
|---|---|
| A | 3a |
| B | 3b |
```

Preserve the table's textual content, row order and column order. Empty printed
cells remain empty. Do not infer missing values, add headings that are not printed,
or convert a table into a `[Calculation]` block merely because it contains
mathematical expressions or proof steps.

The Markdown separator row is structural syntax and is not transcribed content.

## 17. Blank pages

When a page contains no transcribable printed or handwritten content:

```markdown
[Blank page]
```

Do not infer whether the page was intentionally left blank.

[Rules concerning page furniture or other marks on otherwise blank pages to be determined.]

## 18. Digitization artefacts

Do not transcribe material introduced by digitization rather than present on the physical page.

Examples include:

* `Digitized by Google`;
* scanner-generated identifiers;
* interface overlays;
* digital watermarks;
* automatically added digital page numbers.

Do not confuse digitization artefacts with stamps, shelfmarks, annotations or other marks physically present on the digitized copy.

## 19. Unclassified phenomena

Do not invent new syntax when the page contains a feature not covered by this specification.

Use `[Other]...[/Other]` when the feature is a distinct zone that should be represented but has no defined zone type.

Do not use `[Other]` to replace an existing specific annotation merely because classification is difficult.

## 20. Phenomena inventory

This section records phenomena that require an explicit decision about whether and how they should be represented.

### 20.1 Textual structure

To be determined.

### 20.2 Page furniture

To be determined.

### 20.3 Typography and character-level features

To be determined.

### 20.4 Mathematical notation and layout

To be determined.

### 20.5 Page zones and spatial structures

To be determined.

### 20.6 Non-textual elements

To be determined.

### 20.7 Copy-specific and handwritten material

To be determined.

### 20.8 Physical damage, obscuration and reproduction problems

To be determined.

## 21. Syntax reference

| Feature                     | Canonical syntax                                                  |
|-----------------------------|-------------------------------------------------------------------|
| Book-level heading          | `# TEXT`                                                          |
| Proposition/section heading | `## TEXT`                                                         |
| Running title               | `<!-- Running title: TEXT -->`                                    |
| Page/folio number           | `<!-- Page number: TEXT -->`                                      |
| Signature                   | `<!-- Signature: TEXT -->`                                        |
| Catchword                   | `<!-- Catchword: TEXT -->`                                        |
| Drop capital                | `{dropcap:X\|lines=N\|style=STYLE}`                               |
| Decorated drop capital      | `{dropcap:X\|lines=N\|style=decorated\|decoration="DESCRIPTION"}` |
| Printer's error correction  | `PRINTED{printer-error-correction:CORRECTION}`                    |
| Illegible text              | `[illegible]`                                                     |
| Illegible text with extent  | `[illegible: N chars]`                                            |
| Uncertain reading           | `[unclear: TEXT]`                                                 |
| Marginal text               | `[Margin]...[/Margin]`                                            |
| Footnote                    | `[Footnote]...[/Footnote]`                                        |
| Handwritten text            | `[Handwritten]...[/Handwritten]`                                  |
| Other zone                  | `[Other]...[/Other]`                                              |
| Typed other zone            | `[Other type="TYPE"]...[/Other]`                                  |
| Diagram                     | `[Diagram]` or `[Diagram: DESCRIPTION]`                           |
| Figure                      | `[Figure]` or `[Figure: DESCRIPTION]`                             |
| Illustration                | `[Illustration]` or `[Illustration: DESCRIPTION]`                 |
| Ornament                    | `[Ornament]` or `[Ornament: DESCRIPTION]`                         |
| Calculation                 | `[Calculation]...[/Calculation]`                                  |
| Table                       | Markdown pipe-table syntax                                        |
| Blank page                  | `[Blank page]`                                                    |

## 22. Core syntax constraints

1. Preserve paragraph boundaries but do not preserve ordinary printed line breaks.
2. Join words divided solely because of printed line wrapping.
3. Do not invent annotation names.
4. Do not vary the capitalization or spelling of annotation names.
5. Use curly-brace annotations for inline features attached to particular transcribed text.
6. Use square-bracket annotations for zones and page objects.
7. Use HTML comments only for defined page-furniture types.
8. Preserve printed text diplomatically except for the normalization of line wrapping defined above.
9. Do not modernize or silently correct printed text.
10. Do not infer content that is not visible on the current page.
11. Prefer a defined specific zone type over `[Other]`.
12. When the existing syntax cannot represent a distinct zone, use `[Other]` rather than inventing new syntax.
