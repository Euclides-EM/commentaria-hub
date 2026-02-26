package tei

var EntitiesByTest = map[string][]EntityItem{
	"lines_single_entity_translation": {
		{
			Start:    EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 0},
			End:      EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: len("Ibn Rushd")},
			Category: "person",
			Properties: map[string]string{
				"latinized_name": "Averroes",
				"educated_at":    "University of Oxford",
				"born_in":        "Cordoba",
				"died_in":        "Marrakesh",
				"era":            "Medieval",
			},
		},
	},
	"entity_with_fact": {
		{
			Start:    EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 0},
			End:      EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 4},
			Category: "person",
			Properties: map[string]string{
				"nickname": "Johnny",
			},
		},
		{
			Start:    EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 16},
			End:      EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 22},
			Category: "place",
		},
	},
	"txt_overlapping_entities": {
		{
			Start:    EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 2},
			End:      EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 10},
			Category: "axiom drift",
		},
		{
			Start:    EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 0},
			End:      EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 3},
			Category: "lemma cache",
			Properties: map[string]string{
				"stability": "questionable",
			},
		},
		{
			Start:    EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 5},
			End:      EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 8},
			Category: "proof builder",
			Properties: map[string]string{
				"reliability": "unproven",
			},
		},
		{
			Start:    EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 8},
			End:      EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 12},
			Category: "proof builder",
			Properties: map[string]string{
				"reliability": "unknown",
			},
		},
	},
	"txt_adjacent_entities": {
		// No overlap: "John" (0,4), "Smith" (5,10). Both rendered. End is exclusive.
		{
			Start:    EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 0},
			End:      EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 4},
			Category: "first_name",
		},
		{
			Start:    EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 5},
			End:      EntityLocationIndex{BlockID: "1", LineID: "0", ByteOffset: 10},
			Category: "last_name",
			Properties: map[string]string{
				"name_origin": "English",
				"commonality": "common",
				"popularity":  "high",
			},
		},
	},
	"alto_with_entity": {
		// ALTO input has TextLine ID="l1", TextBlock ID="b1"
		{
			Start:    EntityLocationIndex{BlockID: "b1", LineID: "l1", ByteOffset: 0},
			End:      EntityLocationIndex{BlockID: "b1", LineID: "l1", ByteOffset: 9},
			Category: "philosopher",
		},
	},
}
