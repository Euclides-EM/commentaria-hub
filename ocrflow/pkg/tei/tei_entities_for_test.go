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
			Element: "persName",
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
}
