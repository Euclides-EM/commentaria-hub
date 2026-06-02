-- V6 TPS prompt revisions based on the v1/v5 diagnostic review.
-- The main shift from v5 is feature-specific span guidance instead of a
-- universal "minimal span" rule.

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a01',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'editor_name',
    'The full name of the contemporary adapter (author, editor, translator, commentator, etc.) as it appears on the title page. Include initials when they stand for a given name or second name, along with given names, surname particles, and surnames. Do not include professional honorifics or titles such as abbreviations for Professor; those belong in Adapter Description. Do not include following place names, offices, affiliations, roles, or descriptors unless they are inseparable from the printed name.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a02',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'editor_in_imprint',
    'The full name of the contemporary adapter (author, editor, translator, commentator, etc.) when it appears in the imprint section. Include initials when they stand for a given name or second name, along with given names, surname particles, and surnames. Do not include professional honorifics or titles such as abbreviations for Professor; those belong in Adapter Description in Imprint. Do not include printer, publisher, place, office, affiliation, or role descriptors.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a03',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'editor_description',
    'Descriptors found alongside the adapter name, such as academic titles, professional honorifics, ranks, professions, offices, institutional affiliations, or geographic/role descriptors. Include abbreviations such as "P." here when they stand for a professional title such as Professor, not for a given-name initial. Include the full adjacent descriptor phrase when needed for meaning. Do not include the adapter name itself, and do not absorb unrelated work titles or publication details.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a04',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'editor_description_in_imprint',
    'Descriptors found alongside the adapter name in the imprint section, such as academic titles, professional honorifics, ranks, professions, offices, institutional affiliations, or geographic/role descriptors. Include abbreviations such as "P." here when they stand for a professional title such as Professor, not for a given-name initial. Include the full adjacent descriptor phrase when needed for meaning. Do not include the adapter name itself, printer, publisher, place, or unrelated publication details.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a05',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'base_content',
    'The main title or designation of the book''s core content as it appears on the title page. Do not cut off book counts, Euclid references, title qualifiers, or phrases that are part of the title/designation. Exclude separate additions, enrichments, edition statements, dedications, and publisher/imprint details.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a06',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'elements_designation',
    'The designation of Euclid''s Elements as it appears on the title page, including book counts, ranges, and title qualifiers when they are part of the designation (for example references to the first six books, thirteen books, or a specific book). Do not reduce the designation to only the generic word "Elements" when the source gives a fuller designation.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a07',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'content_description',
    'Description of the Elements or core content itself, such as subject, structure, method, or scope. Do not include descriptions of Euclid as a person. Do not absorb additions, enrichments, appended works, or edition statements; those belong to other features. Extract the source phrase that describes the content, not a surrounding paragraph.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a08',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'enriched_with',
    'Mentions of additions or enrichments that supplement the main text, including examples, explanations, demonstrations, figures, annotations, notes, utilities, or similar supporting material. Include fuller phrases when isolated keywords lose meaning. Do not include the core title itself, and do not include separate bound works unless the phrase explicitly describes supplementary enrichment.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a09',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'action_verbs',
    'Action verbs describing the role a contemporary scholar, translator, editor, or adapter played in producing or modifying the work, such as translating, explaining, revising, correcting, augmenting, compiling, or publishing. Include every relevant action verb in coordinated lists, and return distinct verbs or verbal expressions as separate list values. Extract only the verbal expression itself; do not include non-verbal surrounding phrases.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a10',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'origin_language',
    'Explicit references to source or target languages of the work, translation, or edition. Extract all language references when multiple languages are mentioned. Include vernacular or historical language designations. Do not drop a second language reference merely because another language was already extracted.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a11',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'location_in_imprint',
    'The city, town, or compound publication-place phrase in the imprint. Preserve punctuation inside the place phrase, such as commas in "Paris, France", but remove terminal punctuation such as a final period when it only closes the phrase. Keep compound place phrases when they name the publication place. Do not return street addresses, publishers, printers, countries, regions, or institutional names as the place.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a12',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'publisher_in_imprint',
    'The printer, publisher, bookseller, or publishing body named in the imprint. Include the full named phrase when it is part of the publisher/printer identity. Do not include dates, places, privileges, or unrelated descriptors such as "Superiorum" unless they are part of the publisher/printer name.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a13',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'references_to_euclid',
    'References to Euclid''s name as printed. Return the clean name form when the source only adds outer punctuation, article, or preposition. Preserve internal punctuation, line-break hyphenation, or spelling when it is part of the printed name form.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a14',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'institutions',
    'Institutions, societies, universities, colleges, schools, religious orders, or similar bodies associated with the work or person. Prefer the institutional name itself, trimming articles or praise words when they are not part of the name. Keep place qualifiers when they identify the institution.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a15',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'audience',
    'Explicit mentions of intended readers or recipients, including their abilities, level of learning, social role, educational background, or relation to the art/science. Prefer the reader/audience phrase itself, not the preceding purpose phrase, unless the purpose phrase is needed for meaning.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a16',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'educational_authorities_references',
    'Other scholars, mathematicians, commentators, or educational authorities mentioned in relation to the work. Include honorifics such as "Mr." when they are printed as part of the name phrase. Do not include the primary adapter, publisher, printer, dedicatee, or patron unless they are cited as an educational authority.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a17',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'dedicatee_name',
    'Mentions of patrons or dedications, including names, titles, honorifics, offices, and descriptive designations of the dedicatee. Prefer the patron/dedicatee phrase itself; include surrounding command or dedication formula only when it is needed to identify the dedication.',
    '', 'ollama', 'gpt-oss:120b'
),
(
    '7d8fdc01-b3ec-43de-a654-4cba5ffb6a18',
    'v6',
    'V6 span-guided revision',
    datetime('now'), datetime('now'),
    'tps', 'dataset', 'bound_with',
    'Descriptions of additional treatises or works physically bound together with the main work in the volume. Include all named bound works and meaningful descriptors that distinguish them. Do not omit a bound work because it is not in the older manual value. Do not include enrichments integrated into the main Elements text.',
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
