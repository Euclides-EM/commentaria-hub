# Transcription Markdown dialect

This document defines the canonical Markdown representation used for page
transcriptions. It is shared by the frontend and transcription-processing tools,
and is the normative specification included in the LLM transcription-corrector
prompt.

## General principles

- Transcribe only content visible on the current page.
- Preserve historical spelling, capitalization, punctuation, special characters,
  and textual order.
- Do not add explanations, observations, confidence statements, correction
  summaries, or other material that is not part of the transcription.
- Do not wrap the complete transcription in a Markdown code fence.
- Put a blank line between distinct blocks.

## Main text

Represent ordinary body text as plain Markdown paragraphs.

Printed headings that belong to the main text use Markdown headings. Use level 1
for book-level headings and level 2 for section or proposition headings:

```markdown
# Das dritte Buch

## Geſchicht aſo/
```

## Running titles

A running title is page furniture and must not be rendered as a visible Markdown
heading. Put its complete text in one HTML comment:

```markdown
<!-- Running title: Des dritt buͤchs Euclidis -->
```

Do not use `<!-- # Des dritt buͤchs Euclidis -->`, and do not put
`<!-- Running title -->` and its text on separate lines.

## Page numbers

Put a printed page number in one HTML comment:

```markdown
<!-- Page: 86 -->
```

If a candidate transcription combines a running title and page number, separate
them into two annotations:

```markdown
<!-- Running title: Des dritt buͤchs Euclidis -->

<!-- Page: 86 -->
```

Place running-title and page-number annotations at the beginning of the
transcription, following their order on the printed page.

## Catchwords

Put the catchword and its text in one HTML comment:

```markdown
<!-- Catchword: baid -->
```

Never put `<!-- Catchword -->` and the catchword text on separate lines. Preserve
punctuation that visibly belongs to the catchword.

Place catchword annotations at the end of the transcription, after the main text
and figures.

## Signatures and quire marks

Put a signature or quire mark in one HTML comment:

```markdown
<!-- Signature: B 2 -->
<!-- Quire marks: B2 -->
```

Use the label supplied by the source format when it reliably distinguishes a
signature from another quire mark.

## Figures, ornaments, and marginal text

Use italicized bracket annotations for non-textual figures and ornaments:

```markdown
*[Figure]*
*[Figure: Diagram]*
*[Figure: Table]*
*[Ornament]*
```

Represent marginal text as:

```markdown
*[Margin: a note]*
```

## Other zones

Represent drop capitals and digitization artefacts with HTML comments:

```markdown
<!-- Drop capital: A -->
<!-- Drop capital (plain): A -->
<!-- Digitization artefact -->
```

Include a digitization-artefact annotation only when the source data or page image
supports the presence of an artefact. It contains no invented descriptive text.
