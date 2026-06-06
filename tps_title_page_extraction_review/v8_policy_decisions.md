# V8 Policy Decisions

## Verbs

Decision: prefer V7's broader bibliographic verb extraction over the old target in the reviewed verb examples, but do not include ordinary mathematical-operation verbs from content lists.

Rule: include action verbs on the title page related to modifying, translating, explaining, setting out, enlarging, correcting, reviewing, adding to, enriching, publishing, or otherwise presenting the edition/book. Do not exclude added/enriched/translated/reviewed verbs merely because the same action is also represented in another feature. Exclude ordinary content-list verbs that merely name mathematical operations or examples inside an enrichment, such as `maken`, `veranderen`, `t'samen voegen`, `aftrecken`, `vermenichvuldigen`, or `deelen`.

Examples accepted as in-scope: `ghevoecht`, `by-gevoeght`, `angefüget`, `Vertaelt`, `oversien`, `verbetert`, `vermeerdert`, `verrijckt`.

## Enriched With

Decision: V7 is preferred when it extracts distinct enrichment/addition items as separate values.

Rule: if the title page names multiple distinct additions or enrichments, return them as separate values rather than forcing one comma-joined phrase. Do not overfit introductory phrase boundaries when the extracted content object is clear; full-ish source spans are acceptable when they preserve the enrichment meaning.

Accepted example: `Met korte verk-laringen eeniger Propositien` and `een Aenhanghsel der Fondamenten vande Mathematische Namen ende Coracteren` as two separate values.

## Core Content Split

Decision: restore the three-feature split instead of collapsing everything into `Base Content`.

Features to extract separately:
- `base_content` / Base Content
- `elements_designation` / Elements Designation
- `content_description` / Base Content Description

Rule: do not evaluate old `Elements Designation` or `Base Content Description` targets by mapping them onto `Base Content`. V8 should run all three features explicitly.

Base Content decision: prefer the minimal title nucleus for the core Euclidean work, not the full descriptive phrase. For example, from `De ses eerste Boecken EVCLIDIS, Van de beginselen ende fondamenten der Geometrie`, extract `De ses eerste Boecken EVCLIDIS` as Base Content.

Elements Designation decision: A/B boundary is not important as long as the specific Elements/books designation is preserved. Either `De ses eerste Boecken EVCLIDIS` or `ses eerste Boecken EVCLIDIS` is acceptable; do not reduce it to a generic normalized label like only `Elements`.

Base Content Description decision: extract only descriptive wording about the core content itself, such as `Van de beginselen ende fondamenten der Geometrie`. Do not include additions or enrichments such as `Waer by ghevoecht zijn eenighe nut-ticheden`; those belong to `Enriched With`. Do not include Euclid/person honorifics or biographies. Do not include intended-audience/purpose phrases such as learner/use/practice/benefit wording.

Date in Imprint decision: include `date_in_imprint` as a default feature for every title-page extraction run.

Imprint place/publisher decision: include `location_in_imprint` and `publisher_in_imprint` as default features for every title-page extraction run.
