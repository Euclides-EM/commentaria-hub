UPDATE features
SET is_list = 1,
    updated_at = datetime('now')
WHERE id IN (
    'dedicatee_name',
    'editor_name',
    'editor_in_imprint',
    'greek_text'
);

UPDATE features
SET name = 'Dedication',
    updated_at = datetime('now')
WHERE id = 'dedicatee_name';

UPDATE features
SET name = 'Bound With - Description',
    description = 'Descriptions of additional treatises or works physically bound together with the main work in the volume.',
    updated_at = datetime('now')
WHERE id = 'bound_with';

-- v1 prompt: Mentions of other works that are included in the work, in addition to Euclid's Elements, such as 'Optics', 'Data', theorems by Archimedes. Mentions of additions ingrained in the core text and written by the adapter/translator of the text, such as examples or explanations, should not be included here. Try to mark the minimal unit of the bound work and break down the bound works into their components when possible.
INSERT INTO features (
    id, name, description, created_at, updated_at,
    dataset_id, is_default, is_list, color, properties
) VALUES (
    'bound_with_minimal',
    'Bound With - Minimal',
    'Names or titles of additional treatises or works physically bound together with the main work in the volume.',
    datetime('now'),
    datetime('now'),
    'tps',
    1,
    1,
    '#FFB6C1',
    '[]'
) ON CONFLICT(id) DO UPDATE SET
    name = excluded.name,
    description = excluded.description,
    created_at = excluded.created_at,
    updated_at = excluded.updated_at,
    dataset_id = excluded.dataset_id,
    is_default = excluded.is_default,
    is_list = excluded.is_list,
    color = excluded.color,
    properties = excluded.properties;

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    'b58a2f9b-dc7d-459f-8953-4c3cf08aa114',
    'v1',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'bound_with_minimal',
    -- v1 prompt: Mentions of other works that are included in the work, in addition to Euclid's Elements, such as 'Optics', 'Data', theorems by Archimedes. Mentions of additions ingrained in the core text and written by the adapter/translator of the text, such as examples or explanations, should not be included here. Try to mark the minimal unit of the bound work and break down the bound works into their components when possible.
    'Mentions of other treatises or works physically bound together with the main work, in addition to Euclid''s Elements, such as ''Optics'', ''Data'', or works by Archimedes. Extract only the title of the additional bound treatise or work itself, excluding surrounding descriptive text. Do not include additions integrated into the main text by the editor, adapter, or translator, such as examples, explanations, annotations, or commentary. When possible, extract each bound work as a separate minimal unit.',
    '',
    'ollama',
    'gpt-oss:120b'
) ON CONFLICT(id) DO UPDATE SET
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    '26b1bd16-333e-416f-904d-44a9d35c4f8e',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'dedicatee_name',
    -- v1 prompt: Mentions of patrons or dedication.
    'Mentions of patrons or dedications, including cases where the dedicatee is referred to by name, title, honorific, role, or other descriptive designation.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    '46fee541-0f06-4d0b-9b19-775abe81ae84',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'dedication_in_imprint',
    -- v1 prompt: Mentions of patrons or dedications.
    'Mentions of patrons or dedications, including cases where the dedicatee is referred to by name, title, honorific, role, or other descriptive designation.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    '17990f23-7917-490f-917e-96427a004ff7',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'elements_designation',
    -- v1 prompt: The designation of the Elements, such as 'Elements of Geometry' or 'Euclid’s Elements', as it appears on the title page.
    'The designation of the Euclid''s Elements in any form (e.g., Elements, Elementa, Elementorum, Elemens, etc.), and any explicit indication of which books are included by their numbers.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    'e4d1ac46-54a8-4f86-8c7a-2c1d5cf9a101',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'destination_language',
    -- v1 prompt: Mentions of the target language of the edition or translation (e.g., "en François", "in English").
    'Mentions of the target language of the edition or translation (e.g., "François", "English"). Extract only the language name itself, without surrounding words such as "in", "en", or other descriptive phrasing.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    '71d1be08-4bb6-47d0-b31b-0c599353e775',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'date_in_imprint',
    -- v1 prompt: Mentions of the date, usually in the form of a year, when the book was printed or published.
    'Mentions of the year the book was printed or published. The year may appear in standard decimal numerals (e.g. 1604) but is more commonly written in Roman numerals (e.g. MDCIV). Identify and extract only the year itself, not surrounding words, dates, printers, or publication details.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    '475f6bcb-b233-4fe1-b290-f15ce52e1f62',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'location_in_imprint',
    -- v1 prompt: Mentions of the city or town where the book was printed or published. Do not include full addresses, just the city or town name.
    'Mentions of the city or town where the book was printed or published. The location name may appear in early modern, Latinized, archaic, or historical spellings/forms (e.g. ARGENTORATI for Strasbourg). Identify and extract only the city or town name itself, not printers, publishers, countries, regions, or full addresses.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    '26d7bc25-4ed8-4b2f-b8f7-6ca2a4f7d102',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'audience',
    -- v1 prompt: Explicit mentions of the work's intended recipients or audience.
    'Explicit mentions of the work''s intended recipients or audience, including descriptions of their abilities, level of learning, social role, or educational background. Extract references to the characteristics or qualifications of the intended readership, not dedications, dedicatees, patrons, or descriptions of the adapter, editor, or publisher themselves.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    '6c092fad-2448-4d50-bc1d-84d97542d103',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'enriched_with',
    -- v1 prompt: Mentions of additional content that is not part of the core text, such as illustrations, diagrams, explanations, expositions, examples or other supplementary material that enriches the text. Try to mark the minimal unit of enrichment and break down the enrichment into components when possible.
'Mentions of content that supplements the main text rather than constituting the core text itself, including illustrations, diagrams, figures, explanations, expositions, examples, annotations, scholia, notes, or similar additions. Exclude separately titled works, treatises, appendices, or other independent texts merely bound together with the volume. When possible, identify each distinct supplementary component as a separate minimal unit rather than grouping multiple enrichments together.',
          '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    '66fb4fbe-3140-4e4d-9fd9-8dd44c511104',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'educational_authorities_references',
    -- v1 prompt: Mentions of other scholars, either ancients, such as Theon of Alexandria, or contemporary, like Simon Stevin.
    'Mentions of scholars, mathematicians, commentators, or authorities other than the primary author of the work, including both ancient and contemporary figures, such as Theon of Alexandria or Simon Stevin. Do not include the author, translator, adapter, editor, publisher, printer, dedicatee, patron, or other individuals associated primarily with the production or dedication of the edition itself.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    'f8f833bf-6d05-46d6-8468-e2f6fe52f105',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'description_of_euclid',
    -- v1 prompt: Any descriptors found alongside Euclid's name, such as mentioning him being a mathematician.
    'Any descriptors or characterizations attached to Euclid''s name that describe him as a person, such as his profession, status, origin, expertise, or intellectual qualities. Do not include descriptors referring to the text, edition, translation, commentary, or other bibliographic aspects of the work.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    '0b9ec650-2f7b-42ef-b95b-c902d5879106',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'greek_text',
    -- v1 prompt: Greek designation of the book in non-Greek books.
    'Any Greek words, phrases, or textual snippets appearing within an otherwise non-Greek book or title page. Extract any explicitly Greek-language material regardless of its function or meaning.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    'f2b43567-c62f-493d-b854-df22cd07d107',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'origin_language',
    -- v1 prompt: Mentions of the source language (e.g., Latin or Greek) and/or the target language.
    'Explicit mentions of the source language and/or target language of the work, translation, or edition, including both specific language names (e.g., Latin, Greek, English) and broader or vernacular designations (e.g., “volgare”, “vvulgare”, “teutsch”), but not language abbreviations ("Fr"). Extract only the language name or linguistic designation itself, without surrounding words, prepositions, or descriptive phrasing.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    '2f2f3e8b-cb54-4c74-a796-8471fdcd6108',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'action_verbs',
    -- v1 prompt: Action verbs such as traduit (translated), commenté (commented), augmenté (expanded) that describe the role the contemporary scholar played in bringing about the work.
    'Action verbs describing the role a contemporary scholar, translator, editor, or adapter played in producing or modifying the work, such as translating, commenting, augmenting, correcting, or compiling. Extract only the verbal expression itself, without surrounding adjectives or descriptive phrases. If the verb appears in a compound or passive construction, include the full verbal form necessary to preserve its grammatical meaning.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    '64ce32ee-f356-4530-af48-79b51aa0e109',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'bound_with',
    -- v1 prompt: Mentions of other works that are included in the work, in addition to Euclid's Elements, such as 'Optics', 'Data', theorems by Archimedes. Mentions of additions ingrained in the core text and written by the adapter/translator of the text, such as examples or explanations, should not be included here. Try to mark the minimal unit of the bound work and break down the bound works into their components when possible.
    'Mentions or descriptions of additional material included alongside Euclid''s Elements, such as supplementary treatises, appended works, theorems, demonstrations, or other mathematical content. Extract descriptive references to the additional content itself rather than only the names of the works. Do not include additions integrated into the main text by the editor, adapter, or translator, such as examples, explanations, annotations, or commentary. When possible, separate distinct pieces of additional content into minimal individual units.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    '1c6d34fb-9ef0-4e5e-aa80-61700b402110',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'editor_description',
    -- v1 prompt: Any descriptors found alongside the adapter name, such as academic titles, ranks, or affiliations.
    'Any descriptors found alongside the adapter''s name, such as academic titles, ranks, professions, offices, or institutional affiliations. Do not treat initials, abbreviations of given names, or name components themselves as descriptors unless they explicitly denote a title or role.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    '5dcc56ef-6ab8-49db-ba9e-94f82f88e111',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'editor_description_in_imprint',
    -- v1 prompt: Any descriptors found alongside the adapter name, such as academic titles, ranks, or affiliations. Do not include printer or publisher.
    'Any descriptors found alongside the adapter''s name, such as academic titles, ranks, professions, offices, or institutional affiliations. Do not treat initials, abbreviations of given names, or name components themselves as descriptors unless they explicitly denote a title or role.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    'db1547eb-428a-44dc-93df-9ec46eaf3112',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'editor_name',
    -- v1 prompt: The name of the contemporary adapter (author, editor, translator, commentator, etc.) as it appears on the title page.
    'The full name of the contemporary adapter (author, editor, translator, commentator, etc.) as it appears on the title page. Extract the complete form of the name, including given names and surnames, rather than only initials, abbreviated forms, or partial names when the full name is present.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    '8f4ef6cf-f22b-4e99-8a4c-fbe6d6171113',
    'v2',
    'Updated revision after llm change',
    datetime('now'),
    datetime('now'),
    'tps',
    'dataset',
    'editor_in_imprint',
    -- v1 prompt: The name of the contemporary adapter (author, editor, translator, commentator, etc.) as it appears on the title page. Do not include descriptors and do not include printer or publisher.
    'The full name of the contemporary adapter (author, editor, translator, commentator, etc.) as it appears on the title page. Extract the complete form of the name, including given names and surnames, rather than only initials, abbreviated forms, or partial names when the full name is present.',
    '',
    'ollama',
    'gpt-oss:120b'
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

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
             '86e6ad11-7f39-4672-aaa9-3ae11092b479',
             'v2',
             'Updated revision after llm change',
             datetime('now'),
             datetime('now'),
             NULL,
             'editions',
             'm_classifier',
             'A list of category classification strings.

For each category below, determine whether the edition should be classified as:

- **"primary"**
  The category is a central or major subject of the edition, either across the work as a whole or within a substantial portion of it. A work may have multiple primary categories.

- **"secondary"**
  The category is clearly present and meaningfully relevant, but is not among the principal subjects of the edition. The work contains substantive material related to the category, even if the category is used in support of another topic.

- **"unrelated"**
  The category is not meaningfully relevant to the edition based on the provided metadata.

- **"unknown"**
  The metadata provides insufficient, ambiguous, or unclear evidence to determine whether the category is relevant.

## Classification Principles

### 1. Use evidence from the metadata

Classify categories based on the subjects, activities, instruments, methods, applications, or domains clearly indicated by the metadata.

Do not require that a category be the sole or explicit focus of the work in order to assign **secondary**.

### 2. Distinguish primary from secondary

Assign **primary** when the category appears to be a major topic that the work substantially teaches, discusses, demonstrates, or develops.

Assign **secondary** when the category is clearly present and relevant but appears subordinate, supportive, or more limited in scope.

### 3. Allow multiple categories

Historical works often combine several subjects.

Do not force a single primary category. Multiple categories may be classified as **primary** when supported by the metadata.

Likewise, a work may contain several **secondary** categories.

### 4. Avoid automatic inference

Do not assign a category solely because it is commonly used by another subject.

Examples:

- A military engineering work is not automatically **Arithmetic** merely because calculations are required.
- An astronomy work is not automatically **Trigonometry** merely because trigonometric calculations may be used.
- A surveying work is not automatically **Instrument Construction** merely because instruments are mentioned.

However, if the metadata indicates actual treatment, instruction, discussion, application, or substantial use of the category, classify it accordingly.

### 5. Prefer evidence over hierarchy

Related categories are not mutually exclusive.

When metadata supports both a broad and a specific category, both may be assigned.

Example:

- A surveying manual may be both **Surveying** and **Practical Geometry**.
- A cosmography may also be **Astronomy** and **Geography**.
- A navigation text may also include **Instrument Use**.

Do not suppress a category merely because another related category is present.

### 6. Use unknown sparingly

Use **unknown** only when the metadata genuinely does not provide enough information.

When there is reasonable evidence that a category is relevant but not central, prefer **secondary** over **unknown**.

## Output Format

Return exactly one line for every category using this format:

"Category Name::classification_value"

Use only the exact category names below and only the allowed classification values.

## Categories

- Arithmetic: numerical calculation, arithmetic instruction, operations with numbers, fractions, roots, numerical methods, or numerical tables.
- Commercial Mathematics: commercial arithmetic, bookkeeping, currency exchange, profit and loss, interest, partnerships, barter, or trade-related weights and measures.
- Military Engineering: artillery, fortification, siegecraft, military surveying, camp organization, or mathematics applied to military practice.
- Construction: building practice, construction methods, masonry, carpentry, structural work, or practical building operations.
- Practical Geometry: operational geometry for measurement, mensuration, construction, geometric procedures, applied geometrical problems, or practical geometric instruction.
- Surveying: land measurement, fields, boundaries, triangulation, agrimensura, terrestrial distances, heights, depths, or related measurement practices.
- Perspective: artistic or mathematical perspective, visual projection, scenography, pictorial space, or perspective methods.
- Cartography: mapmaking, map projection, chart production, topographical mapping, or mapping techniques.
- Architecture: architectural theory, orders, building design, architectural proportion, architectural drawing, or buildings.
- Gnomonics & Horology: sundials, clocks, dials, calendar devices, or methods of measuring time.
- Astronomy: celestial motions, planets, stars, eclipses, astronomical tables, zodiac, planetary theory, spheres, or mathematical astronomy.
- Cosmography: integrated descriptions of the cosmos combining celestial and terrestrial knowledge, often linking astronomy, geography, climates, coordinates, spheres, and world description.
- Geography: terrestrial description, regions, places, climates, chorography, topography, latitude/longitude, or earth-focused spatial knowledge.
- Instrument Construction: construction, design, fabrication, or making of mathematical, astronomical, surveying, navigational, or measuring instruments.
- **Instrument Use**: instructions for operating or applying instruments such as astrolabes, quadrants, sectors, compasses, globes, geometric squares, or other measuring devices.
- Mechanics: machines, statics, balances, weights, motion, hydraulics, mechanical devices, or mechanical principles.
- Theoretical Mathematics: abstract, speculative, demonstrative, foundational, or theoretical mathematics, including Euclidean geometry, algebra, proportions, conics, solid geometry, proofs, or mathematical method.
- Navigation: nautical navigation, sailing, hydrography, pilotage, maritime routes, nautical astronomy, or sea charts.
- Music Theory: mathematical music theory, harmonics, musical ratios, intervals, or quadrivial music.
- Trigonometry: plane or spherical trigonometry, trigonometric functions, trigonometric tables, triangle calculation, or trigonometric methods.',
          '',
             'ollama',
             'gpt-oss:120b'
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
