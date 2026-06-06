package main

var currentFeatures = []offlineFeature{
	{
		ID: "action_verbs", Name: "Verbs", RevisionID: "9e1d8f80-7c1a-4b7d-a001-000000000006", IsList: true, IsDefault: true,
		Prompt: `Bibliographic action verbs or verbal expressions on the title page: verbs that describe how the edition, book, text, translation, commentary, correction, revision, enlargement, addition, enrichment, publication, or presentation was made or changed. Include verbs and participles for addition, enrichment, translation, explanation, setting out, reviewing, correcting, enlarging, continuing, publishing, or presenting the edition/book, such as ghevoecht, by-gevoeght, angefüget, Overgeset, Vertaelt, verclaert, verklaert, uytgeleyt, oversien, verbetert, vermeerdert, verrijckt, distesa, pubblicata, reveuë, or corrigée. Return distinct verbs or verbal expressions as separate list values. Do not include ordinary content-list verbs that merely name mathematical operations or examples inside an enrichment, such as maken, veranderen, t'samen voegen, aftrecken, vermenigvuldigen, or deelen, unless the title page explicitly uses them as a bibliographic action performed on the edition or book. Do not include nouns, personal names, institution names, or adjectives that only describe the book and do not imply an action.`,
	},
	{
		ID: "audience", Name: "Intended Audience", RevisionID: "8c74a0d1-b4d5-4b90-8b44-0a6771c2a715", IsList: true, IsDefault: true,
		Prompt: `Explicit mentions of intended readers or recipients, including their abilities, level of learning, social role, educational background, or relation to the art or science. Prefer the reader/audience phrase itself, but include the preceding purpose phrase when it is needed to identify the intended-audience construction. Do not return dedicatees, patrons, adapters, publishers, or institutions unless the text clearly frames them as the intended readers.`,
	},
	{
		ID: "base_content", Name: "Base Content", RevisionID: "9e1d8f80-7c1a-4b7d-a001-000000000001", IsList: true, IsDefault: true,
		Prompt: `The minimal title nucleus of the core Euclidean work as printed on the title page. Prefer the compact title/designation itself, not the following descriptive phrase. Include the specific book count, range, or Euclid name when it is part of the title nucleus. For example, from "De ses eerste Boecken EVCLIDIS, Van de beginselen ende fondamenten der Geometrie", extract "De ses eerste Boecken EVCLIDIS" as Base Content. Stop before descriptions of the content, additions, enrichments, edition statements, dedications, bound works, adapter details, and imprint details.`,
	},
	{
		ID: "bound_with", Name: "Bound With", RevisionID: "8c74a0d1-b4d5-4b90-8b44-0a6771c2a718", IsList: true, IsDefault: true,
		Prompt: `Descriptions of additional treatises or works physically bound together with the main work in the volume. Include all named bound works and meaningful descriptors that distinguish them, including a first named work when a list begins with it. Separate distinct bound works when possible. Do not include the core Euclid Elements title itself, and do not include enrichments integrated into the main Elements text.`,
	},
	{
		ID: "bound_with_minimal", Name: "Bound With - Minimal", RevisionID: "b58a2f9b-dc7d-459f-8953-4c3cf08aa114", IsList: true, IsDefault: true,
		Prompt: `Mentions of other treatises or works physically bound together with the main work, in addition to Euclid's Elements, such as 'Optics', 'Data', or works by Archimedes. Extract only the title of the additional bound treatise or work itself, excluding surrounding descriptive text. Do not include additions integrated into the main text by the editor, adapter, or translator, such as examples, explanations, annotations, or commentary. When possible, extract each bound work as a separate minimal unit.`,
	},
	{
		ID: "content_description", Name: "Base Content Description", RevisionID: "9e1d8f80-7c1a-4b7d-a001-000000000007", IsList: true, IsDefault: true,
		Prompt: `Descriptive wording about the core Elements or core mathematical content itself: its subject, structure, method, scope, contents, principles, or foundations. Extract only the source phrase that describes the core content. For example, from "De ses eerste Boecken EVCLIDIS, Van de beginselen ende fondamenten der Geometrie", extract only "Van de beginselen ende fondamenten der Geometrie" as Base Content Description. Do not include the title nucleus itself. Do not include descriptions of Euclid as a person, author, philosopher, mathematician, learned/famous figure, or similar biographical/honorific wording. Do not include additions or enrichments such as utilities, explanations, figures, demonstrations, appendices, or "Met korte verklaringen"; those belong to Enriched With. Do not include intended audience or purpose phrases such as for learners, lovers, advancement, practice, use, or benefit. Do not include adapter activity, edition statements, bound works, or imprint details.`,
	},
	{
		ID: "date_in_imprint", Name: "Date in Imprint", RevisionID: "8c74a0d1-b4d5-4b90-8b44-0a6771c2a719", IsDefault: true, ImprintOnly: true,
		Prompt: `The year or date when the book was printed or published in the imprint. Extract the date value itself, whether Arabic numerals or Roman numerals. Remove surrounding words such as anno, printed, or in the year unless they are inseparable from the date value. Do not include printers, publishers, places, addresses, or privileges.`,
	},
	{
		ID: "dedicatee_name", Name: "Patronage Dedication", RevisionID: "8c74a0d1-b4d5-4b90-8b44-0a6771c2a717", IsDefault: true,
		Prompt: `Mentions of patrons or dedications, including names, titles, honorifics, offices, and descriptive designations of the dedicatee. Prefer the patron or dedicatee name/title phrase itself. Include surrounding dedication formula only when it is needed to identify the dedication, and do not include unrelated publication or adapter details.`,
	},
	{
		ID: "description_of_euclid", Name: "Euclid Description", RevisionID: "8c74a0d1-b4d5-4b90-8b44-0a6771c2a721", IsList: true, IsDefault: true,
		Prompt: `Descriptors attached to Euclid's name that describe Euclid as a person, such as profession, status, origin, expertise, or intellectual qualities. Do not include descriptors of the Elements, the edition, the translation, the adapter, or the publication. Include the Euclid name only when it is grammatically inseparable from the printed descriptor phrase.`,
	},
	{
		ID: "destination_language", Name: "Destination Language", RevisionID: "e4d1ac46-54a8-4f86-8c7a-2c1d5cf9a101", IsList: true, IsDefault: true,
		Prompt: `Mentions of the target language of the edition or translation (e.g., "en François", "in English"). Extract only target-language phrases, preserving the wording from the title page.`,
	},
	{
		ID: "edition_details", Name: "Edition Statement", RevisionID: "8c74a0d1-b4d5-4b90-8b44-0a6771c2a720", IsList: true, IsDefault: true,
		Prompt: `Statements describing the edition, revision, correction, enlargement, or version of this publication. Include the edition statement phrase itself, including coordinated added-material wording when it is part of the edition statement. Do not reduce the value to only a generic phrase such as new edition when the printed statement includes the relevant revision, correction, enlargement, or authorial continuation.`,
	},
	{
		ID: "editor_description", Name: "Adapter Description", RevisionID: "8c74a0d1-b4d5-4b90-8b44-0a6771c2a703", IsList: true, IsDefault: true,
		Prompt: `Descriptors printed with the adapter name, such as academic titles, professional titles, ranks, professions, offices, institutional affiliations, and geographic settings. Professional descriptors always belong here rather than in Adapter Attribution. Geographic phrases belong here when they identify a role, office, residence, institutional setting, or affiliation, such as of the university of Paris; do not take family-name origins such as de Mans away from Adapter Attribution. Include the complete descriptor phrase that belongs to the adapter, but do not include the adapter name itself.`,
	},
	{
		ID: "editor_name", Name: "Adapter Attribution", RevisionID: "8c74a0d1-b4d5-4b90-8b44-0a6771c2a701", IsDefault: true,
		Prompt: `The full printed name of the contemporary adapter (author, editor, translator, commentator, etc.) on the title page. Include given-name initials only when they are clearly part of the personal name; do not keep ambiguous or title-like initials before a complete given name, as in P. CLAUDE FRANÇOIS MILLET DECHALLES where the reviewed adapter name is CLAUDE FRANÇOIS MILLET DECHALLES. Include given names, surname particles, family-origin particles, and surnames. Keep geographic particles or phrases when they function as part of the family name, such as de Mans. Do not drop a surname that follows initials, as in CLAAS JANSZ. VOOGHT. Do not include birthplace, residence, offices, affiliations, roles, or descriptive adjectives after the name. If a geographic phrase identifies a setting or institution, such as of the university of Paris, put that text in Adapter Description instead.`,
	},
	{
		ID: "educational_authorities_references", Name: "Other Educational Authorities", RevisionID: "8c74a0d1-b4d5-4b90-8b44-0a6771c2a716", IsList: true, IsDefault: true,
		Prompt: `Other scholars, mathematicians, commentators, or educational authorities mentioned in relation to the work. Include honorifics such as Mr. when printed as part of the name phrase. Do not include the primary adapter, publisher, printer, dedicatee, or patron unless the title page cites that person as an educational or mathematical authority.`,
	},
	{
		ID: "elements_designation", Name: "Elements Designation", RevisionID: "9e1d8f80-7c1a-4b7d-a001-000000000002", IsList: true, IsDefault: true,
		Prompt: `The printed designation of Euclid's Elements, preserving the specific book count, range, title qualifier, or Euclid reference when printed. A value may match the Base Content title nucleus when that nucleus is itself the Elements designation. It is also acceptable to omit a leading article or generic title word if the specific designation remains intact. Do not reduce the value to only a generic normalized label such as "Elements" or "Boecken" when the source gives a fuller designation such as "De ses eerste Boecken EVCLIDIS" or "ses eerste Boecken EVCLIDIS". Do not include content descriptions, enrichments, edition statements, dedications, adapter details, or imprint details.`,
	},
	{
		ID: "enriched_with", Name: "Enriched With", RevisionID: "9e1d8f80-7c1a-4b7d-a001-000000000008", IsList: true, IsDefault: true,
		Prompt: `Mentions of additions or enrichments that supplement the main text, including utilities, examples, explanations, demonstrations, figures, annotations, notes, scholia, appendices, added books, or similar supporting material. Use this feature when the title page frames the material as added, appended, enriched, supplied, joined, or otherwise supplementary to the core text. If the title page names multiple distinct additions or enrichments, return them as separate values rather than forcing one comma-joined phrase. Include enough of the printed source span to preserve the enrichment meaning; introductory addition wording is acceptable when it is part of the printed enrichment phrase. Do not include the Base Content title nucleus, Base Content Description, adapter names, edition statements, or separate bound works unless the phrase explicitly describes supplementary enrichment integrated with the main text. Do not include intended audience, reader, purpose, or benefit phrases such as "tot vorderinghe", "oeffeninghe", "leergierighe", "liefhebbers", "ad usum", "for the use of", or similar wording unless those words are grammatically part of the enrichment object itself.`,
	},
	{
		ID: "institutions", Name: "Institutions", RevisionID: "8c74a0d1-b4d5-4b90-8b44-0a6771c2a714", IsList: true, IsDefault: true,
		Prompt: `Institutions, societies, universities, colleges, schools, religious orders, gymnasia, or similar bodies associated with the work or person. Prefer the institutional name itself, trimming articles or praise words when they are not part of the name. Keep place, house, college, school, or gymnasium qualifiers when they identify the institution. Preserve printed spacing and punctuation inside the institution name.`,
	},
	{
		ID: "location_in_imprint", Name: "Place in Imprint", RevisionID: "8c74a0d1-b4d5-4b90-8b44-0a6771c2a711", IsList: true, IsDefault: true, ImprintOnly: true,
		Prompt: `The publication place or full publication address in the imprint. Prefer the full address phrase when the imprint gives one, including street, sign, shop, churchyard, address qualifier, city, or compound place phrase. Return multiple address/place values when the imprint gives multiple publication places or address units. Preserve internal punctuation and compound place phrases, but remove terminal punctuation and noisy prefixes such as printed at or in when they are outside the address/place value. Do not include publisher names, printers, dates, or privileges unless they are grammatically inseparable from the address.`,
	},
	{
		ID: "origin_language", Name: "Explicit Language References", RevisionID: "8c74a0d1-b4d5-4b90-8b44-0a6771c2a710", IsList: true, IsDefault: true,
		Prompt: `Explicit references to source or target languages of the work, translation, or edition. Extract all language references when multiple languages are mentioned, including historical or vernacular designations. Do not drop a target language because a source language was already extracted. Do not include descriptors of Euclid as Greek unless the text is explicitly referring to the language of the edition, translation, or source text.`,
	},
	{
		ID: "printing_privilege", Name: "Publishing Privileges", RevisionID: "8c74a0d1-b4d5-4b90-8b44-0a6771c2a723", IsDefault: true,
		Prompt: `Mentions of royal, civic, ecclesiastical, or legal privileges, permissions, licenses, approbations, or superior permissions granted for printing. Extract the privilege or permission phrase itself and do not include unrelated title, adapter, date, or imprint details.`,
	},
	{
		ID: "publisher_in_imprint", Name: "Publisher in Imprint", RevisionID: "8c74a0d1-b4d5-4b90-8b44-0a6771c2a712", IsList: true, IsDefault: true, ImprintOnly: true,
		Prompt: `The printer, publisher, bookseller, or publishing body named in the imprint. Include the full named phrase when it is part of the publisher/printer identity. Do not include dates, places, addresses, privileges, permission formulas, or unrelated approval phrases such as by permission of the superiors unless that wording is part of the publisher name.`,
	},
	{
		ID: "references_to_euclid", Name: "Euclid References", RevisionID: "8c74a0d1-b4d5-4b90-8b44-0a6771c2a713", IsList: true, IsDefault: true,
		Prompt: `References to Euclid's name as printed. Return the printed name form and preserve internal line-break hyphenation, spacing, or repeated forms when they are inside the selected reference. Remove only outer articles, prepositions, or punctuation when they are not part of the name reference. If the title page prints two distinct Euclid references in the same relevant phrase, return both rather than silently choosing one.`,
	},
}
