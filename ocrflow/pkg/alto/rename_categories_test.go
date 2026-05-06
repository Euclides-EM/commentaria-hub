package alto

import "testing"

func TestApplyRenameCategoriesALTORepointsToExistingTag(t *testing.T) {
	doc := &Alto{
		Tags: Tags{
			OtherTags: []OtherTag{
				{ID: "BT1", Label: "DropCapitalZone-Plane"},
				{ID: "BT2", Label: "DropCapitalZone-Plain"},
				{ID: "BT3", Label: "GraphicZone-Diagram"},
			},
		},
		Layout: Layout{
			Page: []Page{
				{
					PrintSpace: PrintSpace{
						TextBlocks: []TextBlock{
							{ID: "tb1", TagRefs: "BT1"},
							{ID: "tb2", TagRefs: "BT1 BT3"},
							{ID: "tb3", TagRefs: "BT2 BT1"},
						},
					},
				},
			},
		},
	}

	if err := ApplyRenameCategoriesALTO(doc, map[string]string{
		"DropCapitalZone-Plane": "DropCapitalZone-Plain",
	}); err != nil {
		t.Fatalf("ApplyRenameCategoriesALTO returned error: %v", err)
	}

	blocks := doc.Layout.Page[0].PrintSpace.TextBlocks
	if got := blocks[0].TagRefs; got != "BT2" {
		t.Fatalf("block 0 tagrefs = %q, want %q", got, "BT2")
	}
	if got := blocks[1].TagRefs; got != "BT2 BT3" {
		t.Fatalf("block 1 tagrefs = %q, want %q", got, "BT2 BT3")
	}
	if got := blocks[2].TagRefs; got != "BT2" {
		t.Fatalf("block 2 tagrefs = %q, want %q", got, "BT2")
	}
}

func TestApplyRenameCategoriesALTORenamesTagLabelWhenTargetMissing(t *testing.T) {
	doc := &Alto{
		Tags: Tags{
			OtherTags: []OtherTag{
				{ID: "BT1", Label: "DropCapitalZone-Plane"},
			},
		},
	}

	if err := ApplyRenameCategoriesALTO(doc, map[string]string{
		"DropCapitalZone-Plane": "DropCapitalZone-Plain",
	}); err != nil {
		t.Fatalf("ApplyRenameCategoriesALTO returned error: %v", err)
	}

	if got := doc.Tags.OtherTags[0].Label; got != "DropCapitalZone-Plain" {
		t.Fatalf("tag label = %q, want %q", got, "DropCapitalZone-Plain")
	}
}
