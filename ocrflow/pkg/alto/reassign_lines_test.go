package alto

import "testing"

func TestReassignTextLinesByToleranceSkipsMissingCategories(t *testing.T) {
	doc := &Alto{}

	moved, err := ReassignTextLinesByTolerance(doc, "MainZone", "MainZone-Head--Section", 5, 0.85)
	if err != nil {
		t.Fatalf("ReassignTextLinesByTolerance() error = %v", err)
	}
	if moved != 0 {
		t.Fatalf("ReassignTextLinesByTolerance() moved = %d, want 0", moved)
	}
}

func TestRecategorizeByAlignmentSkipsMissingTargetCategory(t *testing.T) {
	doc := &Alto{
		Tags: Tags{OtherTags: []OtherTag{
			{ID: "original", Label: "Original"},
			{ID: "relative", Label: "Relative"},
		}},
	}

	if err := RecategorizeByAlignment(doc, "Original", "Missing", "Relative", "horizontal", 5); err != nil {
		t.Fatalf("RecategorizeByAlignment() error = %v", err)
	}
}
