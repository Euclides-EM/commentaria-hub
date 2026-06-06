package main

const (
	defaultFeatureID   = "m_classifier"
	defaultFeatureName = "Subject Categories"
	defaultRevisionID  = "6a0d47e3-f472-4b63-a6f5-67c693a0adf9"
	defaultAIProvider  = "ollama"
	defaultAIModel     = "gpt-oss:120b"
)

const subjectClassifierPrompt = `a list of category classification strings.

For each category below, decide whether the edition should be classified as:
- "primary":
  The category is a central or major subject of the edition, either across the whole work or within a substantial named section or added work. Multiple categories may be primary when the metadata clearly supports that.
- "secondary":
  The category is explicitly present and meaningfully relevant, but not central. Use secondary for a named, concrete topic that appears as a smaller part, appendix, table, chapter, instrument section, added content, or supporting topic.
- "unrelated":
  The metadata gives no meaningful evidence that the category is present.
- "unknown":
  The metadata is too ambiguous or incomplete to decide. Use unknown when the metadata suggests the category may be present but is too thin to distinguish related from unrelated. Do not use unknown when the best answer is simply unrelated.

Decision policy:
- First decide related vs unrelated. Only then decide primary vs secondary.
- Require explicit evidence in title, title-page text, additional content, book/section notes, or notes.
- Do not classify a category as related from editor, publisher, city, language, date, format, or the general reputation of a person alone.
- Do not classify a category as related only because this is an edition of Euclid or because Euclid book numbers are listed.
- If the evidence names only the base Euclidean Elements, keep specialized categories unrelated unless the metadata names the specialized topic directly.
- Prefer unrelated over secondary for weak topical associations or modern interpretations.
- Be conservative for anachronistic categories: do not infer a modern category name from ordinary early-modern geometry unless the metadata names the relevant method, use, or content.
- Return exactly one value for every category listed below. Do not omit categories. If only some categories are supported, still return unrelated or unknown for every other category.

Return one string per category, using this exact format:
"Category Name::classification_value"

Use the exact category names below and only the allowed classification values.

Classify every category:
- Arithmetic: numerical calculation, arithmetic instruction, operations with numbers, fractions, roots, numerical tables, or explicitly practical number work. Do not mark related from generic mathematical content alone.
- Commercial Mathematics: commercial arithmetic, reckoning/reken/counting textbooks for practical calculation, bookkeeping, currency exchange, profit and loss, interest, partnerships, barter, accounting, merchant practice, trade handling, trade weights, or trade measures. General arithmetic, proportion, useful mathematics, measuring, or applied geometry is not enough by itself. Mark primary or secondary when commercial/practical reckoning, accounts, merchants, trade, exchange, money, interest, profit/loss, or commercial weights/measures are explicit. If the metadata only says reckoning/counting and is too thin to determine whether the work is commercial or ordinary arithmetic, use unknown rather than unrelated.
- Military Engineering: artillery, fortification, siegecraft, military surveying, camp organization, or mathematics explicitly for soldiers or war.
- Construction: practical building or construction content, construction methods, masonry, carpentry, structural work, fortification construction, or practical building operations. Do not require the exact word construction if the metadata clearly describes how to build, make, erect, repair, or work with physical structures. Intended audience of builders alone is not enough; the content must teach or discuss building/construction. Pure architectural theory, proportion, or design without practical building content is Architecture rather than Construction.
- Practical Geometry: operational or applied geometry for measurement, construction, mensuration, practical problems, measuring heights/depths/distances/areas, field problems, or instructional geometrical practice. A named practical/mensuration/problem-solving section can be primary or secondary even when it sits inside a broader Euclidean edition. Do not mark related only because the edition contains the standard Euclidean Elements. If the metadata hints at practical geometry but is too thin to judge, use unknown rather than unrelated.
- Surveying: land measurement, measuring fields, distances, heights, depths, triangulation, agrimensura, or survey instruments/methods used for land or terrain.
- Perspective: artistic or mathematical perspective, visual projection, scenography, pictorial space, perspective grids, or optics of representation.
- Cartography: mapmaking, map projection, charts, maps, mapping techniques, or production/use of maps or charts. Do not mark related from general geography, place names, routes, spatial description, diagrams, or surveying alone unless map/chart making or map/chart content is explicit.
- Architecture: architectural theory, orders, design, buildings, architectural proportion, architectural drawing, or geometry/proportion explicitly applied to buildings or architectural design. Do not require the exact word architecture. Do not mark related from generic construction/building words alone unless architectural design, theory, drawing, or building proportion is clear.
- Gnomonics & Horology: sundials, timekeeping, clocks, dials, calendar devices, horologia, or measurement of time.
- Astronomy: celestial motions, spheres, planets, stars, eclipses, astronomical tables, zodiac, planetary theory, or mathematical astronomy.
- Cosmography: integrated description of the cosmos combining celestial and terrestrial knowledge, usually linking astronomy, geography, climates, coordinates, spheres, or world description. Do not mark cosmography when only astronomy or only geography is present.
- Geography: terrestrial description, regions, places, climates, latitude/longitude, chorography, topography, or earth-focused spatial knowledge. Do not mark related from publication place alone.
- Instrument Construction: explicit instructions or content for making, designing, fabricating, calibrating, drawing, or physically constructing real mathematical, astronomical, surveying, navigational, or measuring instruments. Do not mark related from merely naming, depicting, imagining, speculating about, explaining the theory of, or teaching the use of an instrument. If the metadata is about using an instrument rather than building/designing/calibrating it, classify Instrument Construction as unrelated.
- Instrument Use: explicit instructions, rules, demonstrations, or substantial content for using or applying real instruments such as quadrants, astrolabes, sectors, compasses as measuring instruments, geometric squares, globes, dials, or measuring devices. A named instrument section or treatise can be primary or secondary even when it teaches the theory behind use rather than hands-on operation. Ordinary ruler-and-compass or straightedge-and-compass Euclidean construction alone does not count as Instrument Use.
- Mechanics: machines, statics, balances, weights, motion, hydraulics, mechanical devices, or mechanical principles. Practical mathematics or instrument content is not enough unless it teaches mechanics.
- Theoretical Mathematics: abstract or demonstrative mathematics beyond the fact that the work is Euclid, including algebra, foundations, conics, solid geometry, proportions as theory, mathematical method, speculative mathematics, or theoretical commentary. Use a high threshold, but do not suppress a named or substantial theoretical section merely because the edition also has practical aims. Practical, mixed-math, school, instrument, construction, military, or applied Euclidean works should not be classified as Theoretical Mathematics merely because they use Euclid or contain demonstrations. If evidence is suggestive but too thin, use unknown rather than unrelated.
- Navigation: nautical navigation, sailing, hydrography, maritime routes, pilotage, nautical astronomy, sea charts, or shipboard navigation instruments. If navigation is mentioned but the metadata is too thin to determine whether it is substantive content, use unknown rather than unrelated.
- Music Theory: mathematical music theory, harmonics, musical ratios, intervals, tuning, or quadrivial music.
- Trigonometry: plane or spherical trigonometry, sine/tangent/secant tables, triangle calculation by trigonometric methods, or explicit trigonometric vocabulary. Be conservative and avoid anachronism: sundials, astronomy, surveying, practical geometry, triangles, or proportional calculation do not imply Trigonometry unless trigonometric methods or tables are explicit.`
