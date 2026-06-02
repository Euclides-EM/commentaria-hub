INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    '6f4aafde-a8d9-4a50-8cae-7947b470c6f6',
    'v6',
    'Classification-specific prompt with stricter evidence rules and noisy-category disambiguation',
    '2026-06-02T00:00:00Z',
    '2026-06-02T00:00:00Z',
    NULL,
    'editions',
    'm_classifier',
    'a list of category classification strings.

For each category below, decide whether the edition should be classified as:
- "primary":
  The category is a central or major subject of the edition, either across the whole work or within a substantial named section or added work. Multiple categories may be primary when the metadata clearly supports that.
- "secondary":
  The category is explicitly present and meaningfully relevant, but not central. Use secondary for a named, concrete topic that appears as a smaller part, appendix, table, chapter, instrument section, or added content.
- "unrelated":
  The metadata gives no meaningful evidence that the category is present.
- "unknown":
  The metadata is too ambiguous or incomplete to decide. Do not use unknown when the best answer is simply unrelated.

Decision policy:
- First decide related vs unrelated. Only then decide primary vs secondary.
- Require explicit evidence in title, title-page text, additional content, book/section notes, or notes.
- Do not classify a category as related from editor, publisher, city, language, date, format, or the general reputation of a person alone.
- Do not classify a category as related only because this is an edition of Euclid or because Euclid book numbers are listed.
- If the evidence names only the base Euclidean Elements, keep specialized categories unrelated unless the metadata names the specialized topic directly.
- Prefer unrelated over secondary for weak topical associations or modern interpretations.
- Return one value for every category.

Return one string per category, using this exact format:
"Category Name::classification_value"

Use the exact category names below and only the allowed classification values.

Classify every category:
- Arithmetic: numerical calculation, arithmetic instruction, operations with numbers, fractions, roots, numerical tables, or explicitly practical number work. Do not mark related from generic mathematical content alone.
- Commercial Mathematics: commercial arithmetic, bookkeeping, currency exchange, profit and loss, interest, partnerships, barter, accounting, merchant practice, trade handling, trade weights, or trade measures. General arithmetic, proportion, or practical mathematics is not enough.
- Military Engineering: artillery, fortification, siegecraft, military surveying, camp organization, or mathematics explicitly for soldiers or war.
- Construction: building practice, construction methods, masonry, carpentry, structural work, fortification/construction practice, or practical building operations. Architecture or military engineering may count only when the metadata indicates practical construction, not when it is only artistic, theoretical, proportional, or strategic.
- Practical Geometry: operational geometry for measurement, construction, mensuration, applied problems, measuring heights/depths/distances/areas, or instructional geometrical practice. Do not mark related only because the edition contains Euclidean geometry.
- Surveying: land measurement, measuring fields, distances, heights, depths, triangulation, agrimensura, or survey instruments/methods used for land or terrain.
- Perspective: artistic or mathematical perspective, visual projection, scenography, pictorial space, perspective grids, or optics of representation.
- Cartography: mapmaking, map projection, topographical description as mapping, mapping techniques, maps, charts, or production of maps/charts.
- Architecture: architectural theory, orders, design, buildings, architectural proportion, or architectural drawing. Do not mark related from construction/building words alone unless architectural design/theory is explicit.
- Gnomonics & Horology: sundials, timekeeping, clocks, dials, calendar devices, horologia, or measurement of time.
- Astronomy: celestial motions, spheres, planets, stars, eclipses, astronomical tables, zodiac, planetary theory, or mathematical astronomy.
- Cosmography: integrated description of the cosmos combining celestial and terrestrial knowledge, usually linking astronomy, geography, climates, coordinates, spheres, or world description. Do not mark cosmography when only astronomy or only geography is present.
- Geography: terrestrial description, regions, places, climates, latitude/longitude, chorography, topography, or earth-focused spatial knowledge. Do not mark related from publication place alone.
- Instrument Construction: explicit instructions or content for making, designing, fabricating, calibrating, or physically constructing real mathematical, astronomical, surveying, navigational, or measuring instruments. Do not mark related from merely naming, depicting, imagining, or speculating about an instrument.
- Instrument Use: explicit instructions or content for using real instruments such as quadrants, astrolabes, sectors, compasses as instruments, geometric squares, globes, or measuring devices. Ordinary compass-and-ruler geometric construction does not count. Mention of an instrument is secondary unless use is central.
- Mechanics: machines, statics, balances, weights, motion, hydraulics, mechanical devices, or mechanical principles. Practical mathematics or instrument content is not enough unless it teaches mechanics.
- Theoretical Mathematics: abstract or demonstrative mathematics beyond the fact that the work is Euclid, including algebra, foundations, conics, solid geometry, proportions as theory, mathematical method, or theoretical commentary. Use a high threshold: the theoretical material should be significant, not merely basic school mathematics, practical/mixed arts mathematics, or a standard Euclidean base text.
- Navigation: nautical navigation, sailing, hydrography, maritime routes, pilotage, nautical astronomy, sea charts, or shipboard navigation instruments.
- Music Theory: mathematical music theory, harmonics, musical ratios, intervals, tuning, or quadrivial music.
- Trigonometry: plane or spherical trigonometry, sine/tangent/secant tables, triangle calculation, or trigonometric methods. Do not mark related from any triangle or geometry reference unless trigonometric method/table vocabulary is explicit.',
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
