package krakenwrapper

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
)

// ---- Mask creation ----

// CreateMaskFromALTO reads an ALTO XML and writes an inverse black/white mask PNG.
// mainLabels: labels treated as "main zones", painted black
// ignoreLabels: labels that should always be white, overriding main zones
func CreateMaskFromALTO(altoPath, maskPath string, mainLabels, ignoreLabels []string) error {
	// read ALTO XML
	a, err := alto.LoadFromFile(altoPath)
	if err != nil {
		return fmt.Errorf("load ALTO: %w", err)
	}
	if len(a.Layout.Page) == 0 {
		return fmt.Errorf("no pages in ALTO")
	}
	// here we just take the first page; adapt if you have multi-page handling
	page := a.Layout.Page[0]

	// build ID -> LABEL map
	idToLabel := make(map[string]string)
	for _, t := range a.Tags.OtherTags {
		idToLabel[t.ID] = t.Label
	}

	// build fast lookup maps for main/ignored labels
	mainSet := make(map[string]struct{}, len(mainLabels))
	for _, l := range mainLabels {
		mainSet[l] = struct{}{}
	}
	ignoreSet := make(map[string]struct{}, len(ignoreLabels))
	for _, l := range ignoreLabels {
		ignoreSet[l] = struct{}{}
	}

	// create a grayscale image
	img := image.NewGray(image.Rect(0, 0, page.Width, page.Height))

	// NEW: make background white instead of default black
	paintRect(img, 0, 0, page.Width, page.Height, color.Gray{Y: 255})

	// helper to check if any of the tagrefs matches a label in a set
	hasLabelInSet := func(tagRefs string, set map[string]struct{}) bool {
		if tagRefs == "" {
			return false
		}
		ids := strings.Fields(tagRefs) // TAGREFS can be space-separated IDs
		for _, id := range ids {
			if label, ok := idToLabel[id]; ok {
				if _, exists := set[label]; exists {
					return true
				}
			}
		}
		return false
	}

	// first pass: paint all main zones black (inverse of original white)
	for _, tb := range page.PrintSpace.TextBlocks {
		if hasLabelInSet(tb.TagRefs, mainSet) && !hasLabelInSet(tb.TagRefs, ignoreSet) {
			paintRect(img, tb.HPOS, tb.VPOS, tb.Width, tb.Height, color.Gray{Y: 0})
		}
	}

	// second pass: ignored zones become white (inverse of original black)
	for _, tb := range page.PrintSpace.TextBlocks {
		if hasLabelInSet(tb.TagRefs, ignoreSet) {
			paintRect(img, tb.HPOS, tb.VPOS, tb.Width, tb.Height, color.Gray{Y: 255})
		}
	}

	// write mask PNG
	out, err := os.Create(maskPath)
	if err != nil {
		return fmt.Errorf("create mask file: %w", err)
	}
	defer out.Close()

	if err := png.Encode(out, img); err != nil {
		return fmt.Errorf("encode PNG: %w", err)
	}

	return nil
}

// paintRect fills a rectangle (x,y,width,height) in img with the given gray color.
// Coordinates are assumed to match the ALTO page coordinate system (origin top-left).
func paintRect(img *image.Gray, x, y, w, h int, col color.Gray) {
	maxX := x + w
	maxY := y + h

	// clip to image bounds to be safe
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if maxX > img.Bounds().Max.X {
		maxX = img.Bounds().Max.X
	}
	if maxY > img.Bounds().Max.Y {
		maxY = img.Bounds().Max.Y
	}

	for yy := y; yy < maxY; yy++ {
		for xx := x; xx < maxX; xx++ {
			img.SetGray(xx, yy, col)
		}
	}
}
