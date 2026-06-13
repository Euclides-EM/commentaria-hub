# Analysis Terms

This file defines recurring shorthand used in the analysis memos.

## Sparse-Canonical

A **sparse-canonical** title page relies on the authority of a canonical text, author, or book identity while saying relatively little else.

For Euclid/Elements, this usually means the title page foregrounds signals such as Euclid, Elements, book numbers, editor, publisher, or imprint, but gives little or no explicit audience, utility, method, correction, institutional use, social scene, or practical application.

This is not a negative category. Sparse-canonical title pages may be powerful precisely because they do not need to explain why the book matters. They can let canonical identity do the work.

Use with caution:

- It describes title-page rhetoric, not the whole book.
- Silence is not absence of use; it is absence of advertised use on the title page.
- It should be contrasted with louder title-page forms: procedural/pedagogical, utility-public, composite/apparatus, humanist-transfer, and institutional-authority.

### Important Confound: Title-Page Fashion

Sparse-canonical is an interpretive category, not a neutral measurement. It can be produced by several different forces:

- **Canonical confidence:** the title page does not need to explain the work because the author/title already carries authority.
- **Local title-page fashion:** some places, printers, languages, or regions may prefer shorter or longer title-page formulas.
- **Chronology:** title-page verbosity may vary by period. For example, some sixteenth-century German titles may be longer and more elaborative, while some French or France-printed Latin titles may be more minimal. This is a hypothesis to test, not an established conclusion.
- **Bibliographic format:** format may affect title-page design and audience. In `items_print.csv`, `format` uses bibliographic shorthand such as `2` for folio, `4` for quarto, `8` for octavo, `12` for duodecimo. Larger formats may have more space and may be associated with institutional or prestige uses; smaller formats may be more portable, student-facing, or classroom-friendly. This is a control variable, not an assumption.
- **Genre or use context:** schoolbooks, institutional textbooks, or standard course texts may be more sparse because their use is already understood.
- **Material/cataloging/OCR effects:** extraction loss, damaged pages, abbreviated catalog records, or imperfect segmentation can make a title page look quieter than it was. In this project, the user has manually checked many cases, so OCR should be treated as a lower-priority concern than period/place/language/format.

Therefore, do not read sparse-canonical as a direct intellectual claim unless it survives comparison against period, language, place, format, and genre.

Better phrasing:

- Strong: "This title page has sparse-canonical rhetoric."
- Cautious: "This may indicate reliance on canonical identity, but it may also reflect title-page fashion or schoolbook convention."
- Avoid: "This book had no audience/use/method."

Test before citing:

- Compare sparse-canonical rates by `period`, `language_first`, `city`, `format`, and, where possible, school/institution markers.
- Compare within the same language/place/period before making intellectual claims.
- Use close reading of title-page images or full transcriptions for key cases.

## Evidence Layers

The analysis uses several related but distinct layers. They should not be collapsed.

### Claim

A **claim** is an intellectual or textual value advertised by the title page. Claims describe what the book says it does to knowledge or what kind of knowledge it offers: method, demonstration, correction, utility, translation, completeness, selection, visual aids, ancient authority, and so on.

Main source columns:

| Column | Meaning |
| --- | --- |
| `rich_claim_text_raw` | Raw title-page phrases used as evidence for rich claim tagging. |
| `rich_claim_text` | Normalized or joined evidence text. |
| `rich_claim_count` | Number of rich claim modes detected. |

Rich claim columns:

| Column | Shorthand |
| --- | --- |
| `claim_canonical_textual_identity` | Explicit Euclid/book identity: title-page language that presents the work through Euclid, the Elements, numbered books, textual corpus identity, edition identity, or author/title authority, rather than primarily through a practical problem, instrument, profession, or use-case. Older notes call this "canonical/textual identity." |
| `claim_method_demonstration_order` | Method, demonstration, order, proof, procedure. |
| `claim_accessibility_clarity_pedagogy` | Ease, clarity, access, learnability. |
| `claim_utility_practice_application` | Use, practice, application. |
| `claim_correction_revision_accuracy` | Corrected, revised, more accurate. |
| `claim_augmentation_enrichment_composition` | Added, augmented, enriched, composite contents. |
| `claim_translation_vernacularization_transfer` | Translation, vernacularization, language transfer. |
| `claim_ancient_authority_restoration` | Ancient authority, restoration, recovery, humanist authority. |
| `claim_novelty_modernity_invention` | Newness, invention, modernity. |
| `claim_visual_material_aids` | Figures, diagrams, tables, visual/material aids. |
| `claim_completeness_totality_system` | Complete, total, systematic coverage. |
| `claim_selection_extraction_abridgment` | Selection, extraction, abridgment, epitome. |

Older/coarser value columns:

| Column | Shorthand |
| --- | --- |
| `clarity_ease` | Ease/clarity. |
| `utility_use` | Utility/use. |
| `correction_revision` | Correction/revision. |
| `novelty_invention` | Novelty/invention. |
| `translation_language` | Translation/language. |
| `demonstration_method` | Demonstration/method. |
| `enrichment_addition` | Enrichment/addition. |
| `restoration_ancient_authority` | Ancient authority/restoration. |
| `community_professional` | Community/professional value. |

Use with caution:

- Claim tags are title-page rhetoric, not proof of actual contents.
- Claims are sensitive to local title-page fashion: a short title may suppress claims that a longer title would advertise.
- Compare claims within period/language/place when possible.

### Audience

An **audience** is an explicit or implied group of intended readers/users named or strongly signaled on the title page. Audience is narrower than social context: a professor's credential, a dedicatee, or a publisher is not automatically an audience.

Main source columns:

| Column | Meaning |
| --- | --- |
| `audience` | Original extracted audience field. |
| `audience_has` | Whether an audience field was detected. |
| `audience_text` | Evidence text for audience tags. |
| `social_text_raw` | Raw broader social evidence. |
| `rich_social_text_raw` | Raw evidence for richer social-arena tagging. |

Audience columns:

| Column | Shorthand |
| --- | --- |
| `aud_students_learners` | Students, learners, youth, beginners. |
| `aud_general_readers_lovers` | General readers, curious readers, lovers of mathematics. |
| `aud_mathematicians_scholars` | Mathematicians, scholars, learned readers. |
| `aud_artisans_visual_trades` | Painters, engravers, sculptors, visual artisans. |
| `aud_architects_builders` | Architects, builders, masons, construction users. |
| `aud_surveyors_geometers_engineers` | Surveyors, geometers, engineers. |
| `aud_military_users` | Soldiers, officers, artillery, fortification users. |
| `aud_merchants_commercial_users` | Merchants, accountants, commercial users. |
| `aud_navigators_pilots` | Navigators, pilots, maritime users. |
| `aud_nobility_court_as_audience` | Nobility or court named as intended audience. |

Related but not identical social evidence:

| Prefix/Column | Meaning |
| --- | --- |
| `inst_*` | Institutions: universities, academies, Jesuit colleges, royal/civic institutions, military schools. |
| `role_*` | Roles or credentials of authors/editors: professor, royal official, religious identity, engineer/practitioner. |
| `pat_*` | Patronage or dedication structures. |
| `*_arena` | Broader social arenas used for analysis, not necessarily explicit audience. |

Use with caution:

- Patronage is not readership.
- Office or credential is not automatically audience.
- A missing audience field may reflect local fashion or genre convention, not absence of imagined readers.

### Archetype

An **archetype** is a composite analytical label built from combinations of claims, social evidence, and identity signals. It is not a primary extracted feature. It is a browsing and argument-building aid.

Archetype columns:

| Column | Meaning |
| --- | --- |
| `sparse_canonical_identity` | Strong canonical/title identity with low advertised social/intellectual density. |
| `procedural_pedagogical_identity` | Identity plus method, demonstration, explanation, order, teaching, or learnability. |
| `composite_workshop_book` | Title page advertises additions, apparatus, bound-with texts, tables, notes, figures, or compiled material. |
| `utility_public_book` | Utility/application claims align with explicit publics, professions, or use contexts. |
| `humanist_transfer_book` | Ancient/canonical authority is moved through translation, restoration, edition, or learned apparatus. |
| `method_access_book` | Method/order and ease/accessibility are both prominent. |

Related Elements natural-mode columns:

| Column | Meaning |
| --- | --- |
| `mode_sparse_canonical` | Metadata Elements mode: sparse canonical profile. |
| `mode_pedagogical_method` | Metadata Elements mode: pedagogical or method-oriented profile. |
| `mode_vernacular_transfer` | Metadata Elements mode: language transfer or vernacularization. |
| `mode_institutional_authority` | Metadata Elements mode: institutional, credentialed, or authorized profile. |
| `mode_composite_apparatus` | Metadata Elements mode: apparatus, additions, bound-with, or composite form. |
| `mode_practical_public` | Metadata Elements mode: practical/public/use-oriented profile. |
| `mode_corrected_updated` | Metadata Elements mode: correction, revision, updating. |
| `mode_humanist_ancient` | Metadata Elements mode: ancient/canonical/humanist authority. |
| `natural_dominant_mode` | Browsing label for the strongest or most legible mode. |

Use with caution:

- Archetypes are heuristic composites, not historical species.
- They can overlap; a book may be both composite and pedagogical, or both sparse and institutional.
- They are vulnerable to title-page fashion. A short local style may increase sparse-canonical; a verbose local style may inflate claims and archetypes.
- Before making a historical argument from archetypes, check whether the result survives controls by period, language, place, and genre.

## Minimum Control Checks For Claims About Silence Or Density

Any argument about sparse, dense, loud, quiet, promotional, or minimal title pages should be checked against:

| Control | Why it matters |
| --- | --- |
| `period` | Title-page verbosity may change over time. |
| `language_first` | Languages and vernacular traditions may have different title-page conventions. |
| `city` or region | Local printing/title-page fashions may differ. |
| `publisher` or printer, where useful | House style may affect title-page rhetoric. |
| `format` | Folio/quarto/octavo/duodecimo may correlate with space, price, portability, institutional use, or student use. |
| `primary_subject_family` | Some subjects may conventionally advertise use or audience more than others. |
| School/institution markers | Schoolbooks may be sparse because use is assumed, or verbose because pedagogy is advertised. |
| OCR/transcription quality | Sparse evidence can be created by extraction loss, but this is lower priority when cases have been manually checked. |

Working rule:

Do not treat a quiet title page as meaningful silence until it has been compared with nearby books: same period, language, place, format, and subject where possible.
