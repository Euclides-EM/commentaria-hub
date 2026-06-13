# Analysis Plan

This plan is intentionally open-ended. It is designed to help us discover the shape of the corpus rather than fit it into the older argument.

## Phase 0: Finish Data Preparation

Wait for the full subject-classification run to complete.

Then:

- parse the packed `Subject::value` results;
- merge them with the old 155-edition handoff where appropriate;
- apply the reprint representative map;
- produce analysis-ready files:
  - full classification long table;
  - positive classification long table;
  - one-row-per-representative subject matrix;
  - one-row-per-title-page edition with inherited representative classification;
  - TPS feature summary joined to classifications.

## Phase 1: Build The Corpus Atlas

Goal: understand what is actually in the corpus before making an argument.

Questions:

- How many title-page editions are we analyzing?
- How many representative works after reprint deduping?
- What is the chronological distribution?
- Which cities, languages, and publishers dominate?
- What are the main subject profiles?
- Which subjects overlap?
- How many editions have multiple primary subjects?
- Which parts of the corpus have sparse or missing title-page extractions?

Output:

- corpus count table;
- timeline by subject/language/city;
- subject co-occurrence table;
- reprint/representative accounting.

## Phase 2: Study Title-Page Grammar

Goal: understand what title pages do across the mathematical corpus.

Feature families:

- Identity: `base_content`, `elements_designation`
- Authority: `references_to_euclid`, `educational_authorities_references`, `editor_name`, `editor_description`
- Transformation: `action_verbs`, `edition_details`
- Addition: `enriched_with`, `bound_with`, `bound_with_minimal`
- Community: `audience`, `institutions`, `dedicatee_name`
- Circulation: `date_in_imprint`, `location_in_imprint`, `publisher_in_imprint`, language fields

Questions:

- Which features are common, rare, or subject-specific?
- Which title pages are dense with claims?
- Which title pages are minimal?
- What verbs recur: translated, corrected, augmented, demonstrated, explained, reduced, invented?
- What forms of authority appear: ancient names, editors, institutions, offices, orders, academies?

Output:

- feature prevalence table;
- feature co-occurrence table;
- common verbs/claims;
- dense-title-page shortlist.

## Phase 3: Compare Subjects

Goal: ask how different mathematical subjects present themselves.

Questions:

- How do arithmetic books title themselves compared with practical geometry or theoretical mathematics?
- Which subjects foreground audience and use?
- Which subjects foreground institutional authority?
- Which subjects foreground correction, translation, enrichment, or bound-with material?
- Are instrument books and practical geometry books especially audience-heavy?
- Are Euclidean/theoretical works especially authority-heavy?
- Do "elements" titles cluster around Euclid, geometry, pedagogy, or something broader?

Output:

- subject-feature profiles;
- subject comparison memos;
- subject-specific example lists.

## Phase 4: Re-Enter Euclid

Goal: locate Euclid inside the larger mathematical-book ecosystem.

Questions:

- When does Euclid appear as title/author?
- When does Euclid appear as authority or method?
- What is the relationship between Euclid and `elements_designation`?
- How often does "elements" language appear without Euclid?
- How often does Euclid appear without "Elements"?
- Are Euclid title pages exceptional, or do they share formulas with other mathematical books?

Output:

- Euclid/Elements typology;
- contradiction list;
- comparison with non-Euclidean elements books;
- revised version of the "canonical text in motion" claim.

## Phase 5: Build The Casebook

Goal: collect presentation-worthy examples.

Case buckets:

- classic Euclid/Elements cases;
- "Elements" without Euclid;
- Euclid without "Elements";
- dense enrichment or bound-with cases;
- institutional cases;
- professional/audience-heavy cases;
- cases that contradict an earlier assumption;
- visually strong title pages to use in slides.

Output:

- 20-30 candidate cases;
- 8-12 strongest presentation cases;
- notes on why each case matters.

## Phase 6: Draft Possible Arguments

Goal: generate competing arguments before choosing one.

Possible argument families:

1. Title pages as maps of mathematical communities.
2. "Elements" as a broader early modern pedagogical language, not only Euclid.
3. Euclid as one anchor in a larger ecosystem of elementary and practical mathematical print.
4. Mathematical title pages as instruments of authority, utility, and circulation.
5. The early modern mathematical book as a composite object: translated, corrected, enriched, bound with, dedicated, institutionally situated.

Output:

- 2-3 candidate theses;
- evidence tables for each;
- counter-evidence for each;
- preferred conference structure.
