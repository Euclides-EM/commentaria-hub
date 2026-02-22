insert into facsimiles (edition_id, id, created_at, updated_at, name, description, url, main_text_pages)
values ('', 'tps_facsimiles', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'Title Pages Facsimiles', 'Facsimiles for the title pages dataset', '', '');

INSERT INTO datasets (id, name, description, created_at, updated_at, edition_id, facsimile_id, dpi, deskewed) VALUES
('tps', 'Title Pages', 'Title pages dataset', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '', 'tps_facsimiles', -1, FALSE);

INSERT INTO annotations (id, name, description, created_at, updated_at, pages, segmented, ground_truth, ocred, dataset_id) VALUES
('ann_1', 'Title Page Annotation', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '', FALSE, FALSE, TRUE, 'tps');

-- Seed data from: features (from features.py), default feature_revisions (prompt from features.py), feature_results (from title_page.csv).
-- Features: all is_root=1; is_default from FeaturesNotSelectedByDefault (index.ts). Revisions: one per feature, prompt from prompt(), type=annotation, execution_strategy=prompt. Results: source_resp=human, note/name indicate migration.

-- Features: all is_root=1; is_default from FeaturesNotSelectedByDefault (index.ts)
INSERT INTO features (dataset_id, id, created_at, updated_at, name, description, is_root, is_default) VALUES
('tps', 'base_content', datetime('now'), datetime('now'), 'Base Content', 'The minimal designation of the book''s main content, typically appearing at the beginning of the title page, without elaboration.', 1, 1),
('tps', 'content_description', datetime('now'), datetime('now'), 'Content Description', 'Further description of the book’s content, often elaborating on the subject matter, scope, or structure beyond the base content line.', 1, 1),
('tps', 'destination_language', datetime('now'), datetime('now'), 'Destination Language', 'Mentions of the target language of the edition or translation (e.g., “en François”, “in English”).', 1, 1),
('tps', 'editor_name', datetime('now'), datetime('now'), 'Editor Name', 'The name of the contemporary adapter (author, editor, translator, commentator, etc.) as it appears on the title page.', 1, 1),
('tps', 'editor_description', datetime('now'), datetime('now'), 'Editor Description', 'Any descriptors found alongside the adapter name, such as academic titles, ranks, or affiliations.', 1, 1),
('tps', 'dedicatee_name', datetime('now'), datetime('now'), 'Dedicatee Name', 'Mentions of patrons or dedication.', 1, 1),
('tps', 'edition_details', datetime('now'), datetime('now'), 'Edition Details', 'Any information that is highlighted as relevant for this specific edition such as claims regarding the corrections and revisions introduced in it.', 1, 1),
('tps', 'printing_privilege', datetime('now'), datetime('now'), 'Printing Privilege', 'Mentions of royal privileges or legal permissions granted for printing.', 1, 1),
('tps', 'action_verbs', datetime('now'), datetime('now'), 'Action Verbs', 'Action verbs such as traduit (translated), commenté (commented), augmenté (expanded) that describe the role the contemporary scholar played in bringing about the work.', 1, 1),
('tps', 'origin_language', datetime('now'), datetime('now'), 'Origin Language', 'Mentions of the source language (e.g., Latin or Greek) and/or the target language.', 1, 1),
('tps', 'educational_authorities_references', datetime('now'), datetime('now'), 'Educational Authorities References', 'Mentions of other scholars, either ancients, such as Theon of Alexandria, or contemporary, like Simon Stevin.', 1, 1),
('tps', 'references_to_euclid', datetime('now'), datetime('now'), 'References to Euclid', 'Euclid''s name as it appears on the title page.', 1, 1),
('tps', 'description_of_euclid', datetime('now'), datetime('now'), 'Description of Euclid', 'Any descriptors found alongside Euclid''s name, such as mentioning him being a mathematician.', 1, 1),
('tps', 'audience', datetime('now'), datetime('now'), 'Audience', 'Explicit mentions of the work''s intended recipients or audience.', 1, 1),
('tps', 'elements_designation', datetime('now'), datetime('now'), 'Elements Designation', 'The designation of the Elements, such as ''Elements of Geometry'' or ''Euclid''s Elements'', as it appears on the title page.', 1, 0),
('tps', 'greek_text', datetime('now'), datetime('now'), 'Greek Text', 'Greek designation of the book in non-Greek books.', 1, 0),
('tps', 'institutions', datetime('now'), datetime('now'), 'Institutions', 'Mentions of institutions, such as societies or universities, associated with the book.', 1, 1),
('tps', 'bound_with', datetime('now'), datetime('now'), 'Bound With', 'Mentions of other works that are included in the work, in addition to Euclid''s Elements, such as ''Optics'', ''Data'', theorems by Archimedes. Mentions of additions ingrained in the core text and written by the adapter/translator of the text, such as examples or explanations, should not be included here. Try to mark the minimal unit of the bound work. For example, instead of marking ''cum expositione Theonis.'', mark ''expositione Theonis'' as the bound work. Try to also break down the bound works into their components. For example, instead of marking ''Phaenomena, Optics, and Catoptrics'', mark ''Phaenomena'', ''Optics'', and ''Catoptrics'' as the bound works.', 1, 1),
('tps', 'enriched_with', datetime('now'), datetime('now'), 'Enriched With', 'Mentions of additional content that is not part of the core text, such as illustrations, diagrams, explanations, expositions, examples or other supplementary material that enriches the text and makes it more understandable, accurate, or useful. Mentions of other distinct works that are included in the book, in addition to Euclid''s Elements, such as ''Optics'', ''Data'', theorems by Archimedes, should not be included here. Try to mark the minimal unit of enrichment. For example, instead of marking ''To which are added some utilities'', mark ''utilities'' as the enriched content. Try to also break down the enrichment into its components, such as ''explanations'', ''examples'', ''diagrams'', etc.', 1, 1),
('tps', 'date_in_imprint', datetime('now'), datetime('now'), 'Date in Imprint', 'Mentions of the date, usually in the form of a year, when the book was printed or published.', 1, 0),
('tps', 'publisher_in_imprint', datetime('now'), datetime('now'), 'Publisher in Imprint', 'Mentions of the publisher or printer, such as the name of the printing house or the person responsible for the publication. Try to include the minimal unit of the publisher''s name. For example, instead of marking ''Apud Vincentium Accoltum'', mark ''Vincentium Accoltum'', and instead of marking ''Chez GVILLAVME AVVRAY, au haut de la ruë sainct Iean de Beauuais, au Bellerophon couronné'', mark ''GVILLAVME AVVRAY''.', 1, 0),
('tps', 'location_in_imprint', datetime('now'), datetime('now'), 'Location in Imprint', 'Mentions of the city or town where the book was printed or published. Do not include full addresses, just the city or town name. For example, instead of marking ''A Paris, chez GVILLAVME AVVRAY, au haut de la ruë sainct Iean de Beauuais, au Bellerophon couronné'', mark ''Paris''.', 1, 0),
('tps', 'printing_privilege_in_imprint', datetime('now'), datetime('now'), 'Printing Privilege in Imprint', 'Mentions of royal privileges or legal permissions granted for printing.', 1, 0),
('tps', 'dedication_in_imprint', datetime('now'), datetime('now'), 'Dedication in Imprint', 'Mentions of patrons or dedications.', 1, 0),
('tps', 'editor_in_imprint', datetime('now'), datetime('now'), 'Editor in Imprint', 'The name of the contemporary adapter (author, editor, translator, commentator, etc.) as it appears on the title page. Do not include any descriptors, just the name itself. Do not include the printer or publisher''s name, just the adapter''s name, if exists.', 1, 0),
('tps', 'editor_description_in_imprint', datetime('now'), datetime('now'), 'Editor Description in Imprint', 'Any descriptors found alongside the adapter name, such as academic titles, ranks, or affiliations. Do not include the printer or publisher''s name, just the adapter''s description, if exists.', 1, 0)
;

-- Default feature revision per feature (prompt from features.py prompt(), type=annotation, execution_strategy=prompt)
INSERT INTO feature_revisions (dataset_id, id, feature_id, created_at, updated_at, name, description, prompt, regex, execution_strategy, note, type) VALUES
('tps', 'base_content__default', 'base_content', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "base_content": "...", // a single quote or empty if not applicable
}

Definitions: 
- base_content: The minimal designation of the book''s main content, typically appearing at the beginning of the title page, without elaboration.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'editor_name__default', 'editor_name', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "editor_name": "...", // a single quote or empty if not applicable
}

Definitions: 
- editor_name: The name of the contemporary adapter (author, editor, translator, commentator, etc.) as it appears on the title page.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'editor_description__default', 'editor_description', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "editor_description": "...", // a single quote or empty if not applicable
}

Definitions: 
- editor_description: Any descriptors found alongside the adapter name, such as academic titles, ranks, or affiliations.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'dedicatee_name__default', 'dedicatee_name', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "dedicatee_name": "...", // a single quote or empty if not applicable
}

Definitions: 
- dedicatee_name: Mentions of patrons or dedication.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'edition_details__default', 'edition_details', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "edition_details": "...", // a single quote or empty if not applicable
}

Definitions: 
- edition_details: Any information that is highlighted as relevant for this specific edition such as claims regarding the corrections and revisions introduced in it.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'printing_privilege__default', 'printing_privilege', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "printing_privilege": "...", // a single quote or empty if not applicable
}

Definitions: 
- printing_privilege: Mentions of royal privileges or legal permissions granted for printing.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'action_verbs__default', 'action_verbs', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "action_verbs": [...], // zero or more quotes
}

Definitions: 
- action_verbs: Action verbs such as traduit (translated), commenté (commented), augmenté (expanded) that describe the role the contemporary scholar played in bringing about the work.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'origin_language__default', 'origin_language', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "origin_language": [...], // zero or more quotes
}

Definitions: 
- origin_language: Mentions of the source language (e.g., Latin or Greek) and/or the target language.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'educational_authorities_references__default', 'educational_authorities_references', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "educational_authorities_references": [...], // zero or more quotes
}

Definitions: 
- educational_authorities_references: Mentions of other scholars, either ancients, such as Theon of Alexandria, or contemporary, like Simon Stevin.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'references_to_euclid__default', 'references_to_euclid', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "references_to_euclid": [...], // zero or more quotes
}

Definitions: 
- references_to_euclid: Euclid''s name as it appears on the title page.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'description_of_euclid__default', 'description_of_euclid', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "description_of_euclid": [...], // zero or more quotes
}

Definitions: 
- description_of_euclid: Any descriptors found alongside Euclid''s name, such as mentioning him being a mathematician.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'audience__default', 'audience', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "audience": [...], // zero or more quotes
}

Definitions: 
- audience: Explicit mentions of the work''s intended recipients or audience.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'elements_designation__default', 'elements_designation', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "elements_designation": "...", // a single quote or empty if not applicable
}

Definitions: 
- elements_designation: The designation of the Elements, such as ''Elements of Geometry'' or ''Euclid''s Elements'', as it appears on the title page.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'greek_text__default', 'greek_text', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "greek_text": "...", // a single quote or empty if not applicable
}

Definitions: 
- greek_text: Greek designation of the book in non-Greek books.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'institutions__default', 'institutions', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "institutions": [...], // zero or more quotes
}

Definitions: 
- institutions: Mentions of institutions, such as societies or universities, associated with the book.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'bound_with__default', 'bound_with', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "bound_with": [...], // zero or more quotes
}

Definitions: 
- bound_with: Mentions of other works that are included in the work, in addition to Euclid''s Elements, such as ''Optics'', ''Data'', theorems by Archimedes. Mentions of additions ingrained in the core text and written by the adapter/translator of the text, such as examples or explanations, should not be included here. Try to mark the minimal unit of the bound work. For example, instead of marking ''cum expositione Theonis.'', mark ''expositione Theonis'' as the bound work. Try to also break down the bound works into their components. For example, instead of marking ''Phaenomena, Optics, and Catoptrics'', mark ''Phaenomena'', ''Optics'', and ''Catoptrics'' as the bound works.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'enriched_with__default', 'enriched_with', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "enriched_with": [...], // zero or more quotes
}

Definitions: 
- enriched_with: Mentions of additional content that is not part of the core text, such as illustrations, diagrams, explanations, expositions, examples or other supplementary material that enriches the text and makes it more understandable, accurate, or useful. Mentions of other distinct works that are included in the book, in addition to Euclid''s Elements, such as ''Optics'', ''Data'', theorems by Archimedes, should not be included here. Try to mark the minimal unit of enrichment. For example, instead of marking ''To which are added some utilities'', mark ''utilities'' as the enriched content. Try to also break down the enrichment into its components, such as ''explanations'', ''examples'', ''diagrams'', etc.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'date_in_imprint__default', 'date_in_imprint', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page''s imprint.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "date_in_imprint": "...", // a single quote or empty if not applicable
}

Definitions: 
- date_in_imprint: Mentions of the date, usually in the form of a year, when the book was printed or published.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'publisher_in_imprint__default', 'publisher_in_imprint', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page''s imprint.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "publisher_in_imprint": [...], // zero or more quotes
}

Definitions: 
- publisher_in_imprint: Mentions of the publisher or printer, such as the name of the printing house or the person responsible for the publication. Try to include the minimal unit of the publisher''s name. For example, instead of marking ''Apud Vincentium Accoltum'', mark ''Vincentium Accoltum'', and instead of marking ''Chez GVILLAVME AVVRAY, au haut de la ruë sainct Iean de Beauuais, au Bellerophon couronné'', mark ''GVILLAVME AVVRAY''.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'location_in_imprint__default', 'location_in_imprint', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page''s imprint.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "location_in_imprint": [...], // zero or more quotes
}

Definitions: 
- location_in_imprint: Mentions of the city or town where the book was printed or published. Do not include full addresses, just the city or town name. For example, instead of marking ''A Paris, chez GVILLAVME AVVRAY, au haut de la ruë sainct Iean de Beauuais, au Bellerophon couronné'', mark ''Paris''.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'printing_privilege_in_imprint__default', 'printing_privilege_in_imprint', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page''s imprint.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "printing_privilege_in_imprint": [...], // zero or more quotes
}

Definitions: 
- printing_privilege_in_imprint: Mentions of royal privileges or legal permissions granted for printing.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'dedication_in_imprint__default', 'dedication_in_imprint', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page''s imprint.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "dedication_in_imprint": [...], // zero or more quotes
}

Definitions: 
- dedication_in_imprint: Mentions of patrons or dedications.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'editor_in_imprint__default', 'editor_in_imprint', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page''s imprint.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "editor_in_imprint": "...", // a single quote or empty if not applicable
}

Definitions: 
- editor_in_imprint: The name of the contemporary adapter (author, editor, translator, commentator, etc.) as it appears on the title page. Do not include any descriptors, just the name itself. Do not include the printer or publisher''s name, just the adapter''s name, if exists.
', '', 'prompt', 'initial migration', 'annotation'),
('tps', 'editor_description_in_imprint__default', 'editor_description_in_imprint', datetime('now'), datetime('now'), 'Initial Migration', '', 'You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid''s Elements.

You will be given:
- The transcribed text of a title page''s imprint.
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "editor_description_in_imprint": [...], // zero or more quotes
}

Definitions: 
- editor_description_in_imprint: Any descriptors found alongside the adapter name, such as academic titles, ranks, or affiliations. Do not include the printer or publisher''s name, just the adapter''s description, if exists.
', '', 'prompt', 'initial migration', 'annotation')
;
