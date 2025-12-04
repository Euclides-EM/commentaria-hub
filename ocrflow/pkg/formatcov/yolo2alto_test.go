package formatcov

import "testing"

func TestYolo2Alto(t *testing.T) {
	if err := Yolo2Alto("/Users/mia/dev/personal/elements-dh/ocrflow/pkg/formatcov/testdata/3rtoer/yolo", "/Users/mia/dev/personal/elements-dh/ocrflow/pkg/formatcov/testdata/3rtoer/alto"); err != nil {
		t.Fatalf("Yolo2Alto failed: %v", err)
	}

}
