-- V7 TPS prompt revisions after evaluating the V6 diagnostic run.
-- V6 fixed the wrapper scope issue and removed the global minimal-span bias,
-- but the diagnostic rows showed new over-expansion in several feature
-- families. V7 keeps the feature-specific approach and sharpens boundaries.

UPDATE features
SET is_list = 1,
    description = 'The place or full address of publication as it appears in the imprint. May contain multiple address/place values.',
    updated_at = datetime('now')
WHERE id = 'location_in_imprint';

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a701',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'editor_name',
    'The full printed name of the contemporary adapter (author, editor, translator, commentator, etc.) on the title page. Include given-name initials only when they are clearly part of the personal name; do not keep ambiguous or title-like initials before a complete given name, as in P. CLAUDE FRANÇOIS MILLET DECHALLES where the reviewed adapter name is CLAUDE FRANÇOIS MILLET DECHALLES. Include given names, surname particles, family-origin particles, and surnames. Keep geographic particles or phrases when they function as part of the family name, such as de Mans. Do not drop a surname that follows initials, as in CLAAS JANSZ. VOOGHT. Do not include birthplace, residence, offices, affiliations, roles, or descriptive adjectives after the name. If a geographic phrase identifies a setting or institution, such as of the university of Paris, put that text in Adapter Description instead.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a702',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'editor_in_imprint',
    'The full printed name of the contemporary adapter (author, editor, translator, commentator, etc.) when it appears in the imprint section. Include given-name initials when they are part of the name, along with given names, surname particles, family-origin particles, and surnames. Keep geographic particles or phrases when they function as part of the family name. Do not include printer, publisher, publication place, address, office, affiliation, professional role, or institutional setting.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a703',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'editor_description',
    'Descriptors printed with the adapter name, such as academic titles, professional titles, ranks, professions, offices, institutional affiliations, and geographic settings. Professional descriptors always belong here rather than in Adapter Attribution. Geographic phrases belong here when they identify a role, office, residence, institutional setting, or affiliation, such as of the university of Paris; do not take family-name origins such as de Mans away from Adapter Attribution. Include the complete descriptor phrase that belongs to the adapter, but do not include the adapter name itself.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a704',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'editor_description_in_imprint',
    'Descriptors printed with the adapter name in the imprint section, such as academic titles, professional titles, ranks, professions, offices, institutional affiliations, and geographic settings. Professional descriptors always belong here rather than in Adapter Attribution in Imprint. Geographic phrases belong here when they identify a role, office, residence, institutional setting, or affiliation; family-name origins stay with the adapter name. Include the complete descriptor phrase that belongs to the adapter, but do not include the adapter name itself, printer, publisher, publication place, address, or unrelated imprint details.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a705',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'base_content',
    'The main title or designation of the book''s core Euclidean content as it appears on the title page. Include Euclid references, title qualifiers, and book counts or ranges when they are part of the core title, such as TREDECIM in EVCLIDIS ELEMENTORVM GEOMETRICORVM LIBROS TREDECIM. Stop before separately bound works, appended treatises, author lists for other works, enrichments, edition statements, dedications, and imprint details. If another named work begins after the core Elements title, do not absorb it into Base Content.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a706',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'elements_designation',
    'The printed designation of Euclid''s Elements, including book counts, ranges, and title qualifiers when they are part of that designation. Do not reduce the value to only the generic word Elements when the source gives a fuller designation. Do not include descriptions of Euclid as a person, edition or revision statements, enrichment phrases, or following clauses such as statements that the work is contained in a number of books unless that wording is part of the designation itself.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a707',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'content_description',
    'A phrase that describes the Elements or core mathematical content itself: subject, structure, method, scope, or contents. Do not include descriptions of Euclid as a person. Do not absorb appended works, intended audience, adapter activity, or edition statements. Utilities, demonstrations, examples, or figures belong here only when the title page frames them as inherent to the core content rather than as added supplementary material. If the title page frames them as added, appended, enriched, supplied, or extra, put them in Enriched With instead.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a708',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'enriched_with',
    'Mentions of additions or enrichments that supplement the main text, including utilities, examples, explanations, demonstrations, figures, annotations, notes, scholia, or similar supporting material. Use this feature when the title page frames the material as added, appended, enriched, supplied, or otherwise supplementary to the core text. Include enough of the printed phrase to preserve meaning, including words that signal addition such as by, with, added, appended, or enriched when present. If demonstrations or utilities are framed by the title page as inherent core content rather than added material, use Base Content Description instead. Do not include adapter names, intended audience phrases, edition statements, or separate bound works unless the phrase explicitly describes supplementary enrichment integrated with the main text.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a709',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'action_verbs',
    'Action verbs or verbal expressions describing the role a contemporary scholar, translator, editor, commentator, or adapter played in producing or modifying the work, such as translating, explaining, revising, correcting, augmenting, compiling, publishing, setting out, distending, or demonstrating the text. Include participles and adjective-like forms when they function as edition or adapter action claims, such as revised, corrected, reveuë, corrigée, distesa, or pubblicata. Include every relevant action verb in coordinated lists, including the first verb. Return distinct verbs or verbal expressions as separate list values. Do not include nouns, personal names, or adjectives that only describe the book and do not imply an action.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a710',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'origin_language',
    'Explicit references to source or target languages of the work, translation, or edition. Extract all language references when multiple languages are mentioned, including historical or vernacular designations. Do not drop a target language because a source language was already extracted. Do not include descriptors of Euclid as Greek unless the text is explicitly referring to the language of the edition, translation, or source text.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a711',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'location_in_imprint',
    'The publication place or full publication address in the imprint. Prefer the full address phrase when the imprint gives one, including street, sign, shop, churchyard, address qualifier, city, or compound place phrase. Return multiple address/place values when the imprint gives multiple publication places or address units. Preserve internal punctuation and compound place phrases, but remove terminal punctuation and noisy prefixes such as printed at or in when they are outside the address/place value. Do not include publisher names, printers, dates, or privileges unless they are grammatically inseparable from the address.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a712',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'publisher_in_imprint',
    'The printer, publisher, bookseller, or publishing body named in the imprint. Include the full named phrase when it is part of the publisher/printer identity. Do not include dates, places, addresses, privileges, permission formulas, or unrelated approval phrases such as by permission of the superiors unless that wording is part of the publisher name.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a713',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'references_to_euclid',
    'References to Euclid''s name as printed. Return the printed name form and preserve internal line-break hyphenation, spacing, or repeated forms when they are inside the selected reference. Remove only outer articles, prepositions, or punctuation when they are not part of the name reference. If the title page prints two distinct Euclid references in the same relevant phrase, return both rather than silently choosing one.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a714',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'institutions',
    'Institutions, societies, universities, colleges, schools, religious orders, gymnasia, or similar bodies associated with the work or person. Prefer the institutional name itself, trimming articles or praise words when they are not part of the name. Keep place, house, college, school, or gymnasium qualifiers when they identify the institution. Preserve printed spacing and punctuation inside the institution name.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a715',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'audience',
    'Explicit mentions of intended readers or recipients, including their abilities, level of learning, social role, educational background, or relation to the art or science. Prefer the reader/audience phrase itself, but include the preceding purpose phrase when it is needed to identify the intended-audience construction. Do not return dedicatees, patrons, adapters, publishers, or institutions unless the text clearly frames them as the intended readers.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a716',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'educational_authorities_references',
    'Other scholars, mathematicians, commentators, or educational authorities mentioned in relation to the work. Include honorifics such as Mr. when printed as part of the name phrase. Do not include the primary adapter, publisher, printer, dedicatee, or patron unless the title page cites that person as an educational or mathematical authority.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a717',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'dedicatee_name',
    'Mentions of patrons or dedications, including names, titles, honorifics, offices, and descriptive designations of the dedicatee. Prefer the patron or dedicatee name/title phrase itself. Include surrounding dedication formula only when it is needed to identify the dedication, and do not include unrelated publication or adapter details.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a718',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'bound_with',
    'Descriptions of additional treatises or works physically bound together with the main work in the volume. Include all named bound works and meaningful descriptors that distinguish them, including a first named work when a list begins with it. Separate distinct bound works when possible. Do not include the core Euclid Elements title itself, and do not include enrichments integrated into the main Elements text.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a719',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'date_in_imprint',
    'The year or date when the book was printed or published in the imprint. Extract the date value itself, whether Arabic numerals or Roman numerals. Remove surrounding words such as anno, printed, or in the year unless they are inseparable from the date value. Do not include printers, publishers, places, addresses, or privileges.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a720',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'edition_statement',
    'Statements describing the edition, revision, correction, enlargement, or version of this publication. Include the edition statement phrase itself, including coordinated added-material wording when it is part of the edition statement. Do not reduce the value to only a generic phrase such as new edition when the printed statement includes the relevant revision, correction, enlargement, or authorial continuation.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a721',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'description_of_euclid',
    'Descriptors attached to Euclid''s name that describe Euclid as a person, such as profession, status, origin, expertise, or intellectual qualities. Do not include descriptors of the Elements, the edition, the translation, the adapter, or the publication. Include the Euclid name only when it is grammatically inseparable from the printed descriptor phrase.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a722',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'printing_privilege_in_imprint',
    'Mentions of royal, civic, ecclesiastical, or legal privileges, permissions, licenses, approbations, or superior permissions granted for printing in the imprint. Extract the privilege or permission phrase itself. Do not include printer, publisher, date, place, or address unless the wording is part of the privilege formula.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '8c74a0d1-b4d5-4b90-8b44-0a6771c2a723',
    'v7',
    'V7 targeted span-boundary revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'printing_privilege',
    'Mentions of royal, civic, ecclesiastical, or legal privileges, permissions, licenses, approbations, or superior permissions granted for printing. Extract the privilege or permission phrase itself and do not include unrelated title, adapter, date, or imprint details.',
    '', 'ollama', 'gpt-oss:120b'
)
ON CONFLICT(id) DO UPDATE SET
    name = excluded.name,
    description = excluded.description,
    created_at = excluded.created_at,
    updated_at = excluded.updated_at,
    dataset_id = excluded.dataset_id,
    scope = excluded.scope,
    feature_id = excluded.feature_id,
    prompt = excluded.prompt,
    categorizer = excluded.categorizer,
    ai_provider = excluded.ai_provider,
    ai_model = excluded.ai_model;
