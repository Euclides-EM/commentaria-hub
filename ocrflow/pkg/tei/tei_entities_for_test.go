package tei

var EntitiesByTest = map[string][]EntityItem{
	"lines_single_entity_translation": {
		{
			Ref:   "ent_ibn_rushd",
			Start: EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 0},
			End:   EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: len("Ibn Rushd")},
			Ana:   "#feat_person",
		},
		{
			Ref:   "ent_ibn_rushd",
			Type:  "latinized_name",
			Value: "Averroes",
		},
	},
	"entity_with_fact": {
		{
			Ref:   "ent_john",
			Start: EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 0},
			End:   EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 4},
		},
		{
			Ref:   "ent_oxford",
			Start: EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 16},
			End:   EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 22},
		},
		{
			Ref:  "ent_john",
			Type: "modern name",
			Cert: 0.92,
		},
	},
	"txt_overlapping_entities": {
		// Overlapping: "John Smith" (0,10), "John" (0,4), "Smith" (5,10). End is exclusive.
		// Sorted by Start then End desc: (0,10), (0,4), (5,10) → m_1, m_2, m_3. Overlap filter keeps only m_1 in body.
		{
			Ref:   "ent_john_smith",
			Start: EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 0},
			End:   EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 10},
			Ana:   "#feat_person",
		},
		{
			Ref:   "ent_john",
			Start: EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 0},
			End:   EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 4},
			Ana:   "#feat_person",
		},
		{
			Ref:   "ent_smith",
			Start: EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 5},
			End:   EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 10},
			Ana:   "#feat_person",
		},
	},
	"txt_adjacent_entities": {
		// No overlap: "John" (0,4), "Smith" (5,10). Both rendered. End is exclusive.
		{
			Ref:   "ent_john",
			Start: EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 0},
			End:   EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 4},
			Ana:   "#feat_person",
		},
		{
			Ref:   "ent_smith",
			Start: EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 5},
			End:   EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 10},
			Ana:   "#feat_person",
		},
	},
	"alto_with_entity": {
		{
			Ref:   "ent_aristotle",
			Start: EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 0},
			End:   EntityLocationIndex{PageID: "page1", BlockID: "b1", LineID: "l0001", ByteOffset: 9},
		},
		{
			Ref:  "ent_aristotle",
			Type: "educated_at",
			Cert: 0.85,
		},
	},
}
