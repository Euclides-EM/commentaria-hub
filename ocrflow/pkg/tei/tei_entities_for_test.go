package tei

var EntitiesByTest = map[string][]EntityItem{
	"lines_single_entity_translation": {
		{
			Ref:     "ent_ibn_rushd",
			PageID:  "page1",
			BlockID: "b1",
			LineID:  "l0001",
			Start:   0,
			End:     len("Ibn Rushd"),
			Ana:     "#feat_person",
		},
		{
			Ref:   "ent_ibn_rushd",
			Type:  "latinized_name",
			Value: "Averroes",
		},
	},
	"entity_with_fact": {
		{
			Ref:     "ent_john",
			PageID:  "page1",
			BlockID: "b1",
			LineID:  "l0001",
			Start:   0,
			End:     4,
			Element: "persName",
		},
		{
			Ref:     "ent_oxford",
			PageID:  "page1",
			BlockID: "b1",
			LineID:  "l0001",
			Start:   16,
			End:     22,
			Element: "orgName",
		},
		{
			Ref:       "ent_john",
			Type:      "educated_at",
			ObjectRef: "ent_oxford",
			Cert:      0.92,
		},
	},
	"txt_overlapping_entities": {
		// Overlapping: "John Smith" (0,10), "John" (0,4), "Smith" (5,10). End is exclusive.
		// Sorted by Start then End desc: (0,10), (0,4), (5,10) → m_1, m_2, m_3. Overlap filter keeps only m_1 in body.
		{
			Ref:     "ent_john_smith",
			PageID:  "page1",
			BlockID: "b1",
			LineID:  "l0001",
			Start:   0,
			End:     10,
			Element: "persName",
			Ana:     "#feat_person",
		},
		{
			Ref:     "ent_john",
			PageID:  "page1",
			BlockID: "b1",
			LineID:  "l0001",
			Start:   0,
			End:     4,
			Element: "persName",
			Ana:     "#feat_person",
		},
		{
			Ref:     "ent_smith",
			PageID:  "page1",
			BlockID: "b1",
			LineID:  "l0001",
			Start:   5,
			End:     10,
			Element: "persName",
			Ana:     "#feat_person",
		},
	},
	"txt_adjacent_entities": {
		// No overlap: "John" (0,4), "Smith" (5,10). Both rendered. End is exclusive.
		{
			Ref:     "ent_john",
			PageID:  "page1",
			BlockID: "b1",
			LineID:  "l0001",
			Start:   0,
			End:     4,
			Element: "persName",
			Ana:     "#feat_person",
		},
		{
			Ref:     "ent_smith",
			PageID:  "page1",
			BlockID: "b1",
			LineID:  "l0001",
			Start:   5,
			End:     10,
			Element: "persName",
			Ana:     "#feat_person",
		},
	},
	"alto_with_entity": {
		{
			Ref:     "ent_aristotle",
			PageID:  "page1",
			BlockID: "b1",
			LineID:  "l0001",
			Start:   0,
			End:     9,
			Element: "persName",
		},
		{
			Ref:       "ent_aristotle",
			Type:      "educated_at",
			ObjectRef: "ent_lyceum",
			Cert:      0.85,
		},
	},
}
