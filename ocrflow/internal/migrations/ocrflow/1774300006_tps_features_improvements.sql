UPDATE features
SET is_list = 1
WHERE id IN (
    'dedicatee_name',
    'editor_name',
    'editor_in_imprint',
    'greek_text'
);

UPDATE features
SET name = 'Bound With - Description'
WHERE id = 'bound_with';

-- v1 prompt: Mentions of other works that are included in the work, in addition to Euclid's Elements, such as 'Optics', 'Data', theorems by Archimedes. Mentions of additions ingrained in the core text and written by the adapter/translator of the text, such as examples or explanations, should not be included here. Try to mark the minimal unit of the bound work and break down the bound works into their components when possible.
INSERT INTO features (
    id, name, description, created_at, updated_at,
    dataset_id, is_default, is_list, color, properties
) VALUES (
    'bound_with_minimal',
    'Bound With - Minimal',
    'Mentions of other treatises or works physically bound together with the main work in the volume. Extract only the title of the additional treatise or bound work itself (e.g. ''Optics''), not surrounding descriptive text.',
    '2026-03-01T15:37:08Z',
    '2026-03-01T15:37:08Z',
    'tps',
    1,
    1,
    '#7BC8A4',
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
    '2026-03-01T15:37:08Z',
    '2026-03-01T15:37:08Z',
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
    'e4d1ac46-54a8-4f86-8c7a-2c1d5cf9a101',
    'v2',
    'Updated revision after llm change',
    '2026-03-01T15:37:08Z',
    '2026-03-01T15:37:08Z',
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
    '26d7bc25-4ed8-4b2f-b8f7-6ca2a4f7d102',
    'v2',
    'Updated revision after llm change',
    '2026-03-01T15:37:08Z',
    '2026-03-01T15:37:08Z',
    'tps',
    'dataset',
    'audience',
    -- v1 prompt: Explicit mentions of the work's intended recipients or audience.
    'Explicit mentions of the work''s intended recipients or audience, including descriptions of their abilities, level of learning, social role, or educational background. Extract references to the characteristics or qualifications of the intended readership, not dedicatees, patrons, or descriptions of the adapter, editor, or publisher themselves.',
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
    '2026-03-01T15:37:08Z',
    '2026-03-01T15:37:08Z',
    'tps',
    'dataset',
    'enriched_with',
    -- v1 prompt: Mentions of additional content that is not part of the core text, such as illustrations, diagrams, explanations, expositions, examples or other supplementary material that enriches the text. Try to mark the minimal unit of enrichment and break down the enrichment into components when possible.
    'Mentions of supplementary material explicitly presented as enriching, clarifying, or supporting Euclid''s Elements, such as illustrations, diagrams, explanations, expositions, examples, annotations, or similar additions to the core text. Do not include separate treatises, appended works, or other independently titled texts bound with the volume. When possible, extract each distinct enrichment as a separate minimal unit.',
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
    '2026-03-01T15:37:08Z',
    '2026-03-01T15:37:08Z',
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
    '2026-03-01T15:37:08Z',
    '2026-03-01T15:37:08Z',
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
    '2026-03-01T15:37:08Z',
    '2026-03-01T15:37:08Z',
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
    '2026-03-01T15:37:08Z',
    '2026-03-01T15:37:08Z',
    'tps',
    'dataset',
    'origin_language',
    -- v1 prompt: Mentions of the source language (e.g., Latin or Greek) and/or the target language.
    'Mentions of the source language and/or target language of the work, translation, or edition, including both specific language names (e.g., Latin, Greek, English) and broader or vernacular designations (e.g., “volgare”, “vvulgare”, “teutsch”). Extract only the language name or linguistic designation itself, without surrounding words, prepositions, or descriptive phrasing.',
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
    '2026-03-01T15:37:08Z',
    '2026-03-01T15:37:08Z',
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
    '2026-03-01T15:37:08Z',
    '2026-03-01T15:37:08Z',
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
    '2026-03-01T15:37:08Z',
    '2026-03-01T15:37:08Z',
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
    '2026-03-01T15:37:08Z',
    '2026-03-01T15:37:08Z',
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
    '2026-03-01T15:37:08Z',
    '2026-03-01T15:37:08Z',
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
    '2026-03-01T15:37:08Z',
    '2026-03-01T15:37:08Z',
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
