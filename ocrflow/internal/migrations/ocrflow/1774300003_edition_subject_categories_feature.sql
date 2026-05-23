INSERT INTO features (
    id, name, description, created_at, updated_at,
    dataset_id, scope, is_default, is_list, color, properties
) VALUES (
    'm_classifier',
    'Subject Categories',
    'Classifies each edition against a fixed list of mathematical and related subject categories.',
    '2026-05-22T14:19:49Z',
    '2026-05-22T14:19:49Z',
    NULL,
    'editions',
    1,
    1,
    '#4E7A6A',
    '[]'
) ON CONFLICT(id) DO UPDATE SET
    name = excluded.name,
    description = excluded.description,
    created_at = excluded.created_at,
    updated_at = excluded.updated_at,
    dataset_id = excluded.dataset_id,
    scope = excluded.scope,
    is_default = excluded.is_default,
    is_list = excluded.is_list,
    color = excluded.color,
    properties = excluded.properties;

INSERT INTO feature_revisions (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
) VALUES (
    'f96afd86-79f0-4736-a91a-d58e37b6db65',
    'v1',
    'Initial seeded revision',
    '2026-05-22T14:19:49Z',
    '2026-05-22T14:19:49Z',
    NULL,
    'editions',
    'm_classifier',
    'a list of category classification strings.

For each category below, decide whether the edition should be classified as:
- "primary":
  The category is a central or major subject of the edition, whether across the entire work or within a substantial section of it. A category may be marked as "primary" even if other categories are also primary. Works may therefore contain multiple primary categories.
- "secondary":
  The category is clearly present and meaningfully relevant, but not central to the edition. The work contains material related to the category without being primarily devoted to it.
- "unrelated":
  The category is not meaningfully relevant to the edition based on the provided metadata.
- "unknown":
  The metadata provides insufficient, ambiguous or unclear evidence to determine whether the category is relevant.

Return one string per category, using this exact format:
"Category Name::classification_value"

Use the exact category names below and only the allowed classification values.

Classify every category:
- Arithmetic: numerical calculation, number theory, arithmetic instruction, operations with numbers, fractions, roots or numerical tables.
- Commercial Mathematics: commercial arithmetic, bookkeeping, currency exchange, profit and loss, interest, partnerships, barter, weights and measures used in trade.
- Military Engineering: artillery, fortification, siegecraft, military surveying, camp organization or mathematics for soldiers.
- Construction: building practice, construction methods, masonry, carpentry, structural work or practical building operations.
- Practical Geometry: operational geometry for measurement, construction, mensuration, applied problems or instructional geometrical practice.
- Surveying: land measurement, measuring fields, distances, heights, depths, triangulation or agrimensura.
- Perspective: artistic or mathematical perspective, visual projection, scenography, pictorial space or perspective grids.
- Cartography: mapmaking, map projection, topographical description, mapping techniques or production of maps/charts.
- Architecture: architectural theory, orders, design, buildings, architectural proportion or architectural drawing.
- Gnomonics & Horology: sundials, timekeeping, clocks, dials, calendar devices or measurement of time.
- Astronomy: celestial motions, spheres, planets, stars, eclipses, astronomical tables, zodiac, planetary theory or mathematical astronomy.
- Cosmography: integrated description of the cosmos combining celestial and terrestrial knowledge, usually linking astronomy, geography, climates, coordinates, spheres or world description.
- Geography: terrestrial description, regions, places, climates, latitude/longitude, chorography, topography or earth-focused spatial knowledge.
- Instrument Construction: instructions for constructing mathematical, astronomical, surveying, navigational or measuring instruments.
- Instrument Use: instructions for using instruments such as quadrants, astrolabes, sectors, compasses, geometric squares, globes or measuring devices.
- Mechanics: machines, statics, balances, weights, motion, hydraulics, mechanical devices or mechanical principles.
- Theoretical Mathematics: abstract or demonstrative mathematics, including Euclidean geometry, algebra, theoretical arithmetic, proportions, foundations, conics, solid geometry or mathematical method.
- Navigation: nautical navigation, sailing, hydrography, maritime routes, pilotage, nautical astronomy or sea charts.
- Music Theory: mathematical music theory, harmonics, musical ratios, intervals or quadrivial music.
- Trigonometry: plane or spherical trigonometry, sine/tangent/secant tables, triangle calculation or trigonometric methods.',
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
