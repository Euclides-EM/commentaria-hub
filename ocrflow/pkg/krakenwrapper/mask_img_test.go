package krakenwrapper

import (
	"testing"
)

func TestCreateMaskFromALTO(t *testing.T) {

	if err := DetectLines("/Users/mia/dev/personal/elements-dh/ocrflow/pkg/krakenwrapper/testdata", "/Users/mia/dev/personal/elements-dh/ocrflow/pkg/krakenwrapper/testdata"); err != nil {
		t.Fatal("error:", err)
	}

	//// adapt these lists to match your "main" and "ignored" zones
	//mainLabels := []string{
	//	"MainZone",
	//}
	//
	//ignoreLabels := []string{
	//	"DigitizationArtefactZone",
	//	"GraphicZone-Diagram",
	//}
	//
	//if err := CreateMaskFromALTO(
	//	"/Users/mia/dev/personal/elements-dh/ocrflow/pkg/krakenwrapper/testdata/page-0091.xml",
	//	"/Users/mia/dev/personal/elements-dh/ocrflow/pkg/krakenwrapper/testdata/page-0091_mask.png",
	//	mainLabels,
	//	ignoreLabels,
	//); err != nil {
	//	fmt.Println("error:", err)
	//}

}
