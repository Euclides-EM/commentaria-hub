package krakenwrapper

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
)

// Kraken requires the mask to be bitonal (exactly two colors). We use a 2-color
// paletted image so PIL's is_bitonal() accepts it; grayscale PNGs can be rejected.

const (
	maskIdxWhite = 0
	maskIdxBlack = 1
)

var maskPalette = color.Palette{
	color.White,
	color.Black,
}

// CreateMaskFromALTO reads an ALTO XML and writes an inverse black/white mask PNG.
// The mask is bitonal (2-color paletted) so Kraken accepts it.
// mainLabels: labels treated as "main zones", painted black
// ignoreLabels: labels that should always be white, overriding main zones
//
// It returns hasRegions: true if at least one main zone was painted black. When false,
// the mask would be all white (one color); Kraken rejects that, so callers should skip
// the Kraken run for this mask.
func CreateMaskFromALTO(altoPath, maskPath string, mainLabels, ignoreLabels []string) (hasRegions bool, err error) {
	// read ALTO XML
	a, err := alto.LoadFromFile(altoPath)
	if err != nil {
		return false, fmt.Errorf("load ALTO: %w", err)
	}
	if len(a.Layout.Page) == 0 {
		return false, fmt.Errorf("no pages in ALTO")
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

	// Bitonal mask: paletted image with only white (0) and black (1)
	img := image.NewPaletted(image.Rect(0, 0, page.Width, page.Height), maskPalette)
	paintRectPaletted(img, 0, 0, float64(page.Width), float64(page.Height), maskIdxWhite)

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

	// first pass: paint all main zones black (only zones with positive area; zero-area would yield all-white mask and Kraken "Mask is not bitonal")
	for _, tb := range page.PrintSpace.TextBlocks {
		if hasLabelInSet(tb.TagRefs, mainSet) && !hasLabelInSet(tb.TagRefs, ignoreSet) && tb.Width > 0 && tb.Height > 0 {
			paintRectPaletted(img, tb.HPOS, tb.VPOS, tb.Width, tb.Height, maskIdxBlack)
			hasRegions = true
		}
	}

	// second pass: ignored zones become white
	for _, tb := range page.PrintSpace.TextBlocks {
		if hasLabelInSet(tb.TagRefs, ignoreSet) {
			paintRectPaletted(img, tb.HPOS, tb.VPOS, tb.Width, tb.Height, maskIdxWhite)
		}
	}

	// write mask PNG
	out, err := os.Create(maskPath)
	if err != nil {
		return false, fmt.Errorf("create mask file: %w", err)
	}
	defer out.Close()

	if err := png.Encode(out, img); err != nil {
		return false, fmt.Errorf("encode PNG: %w", err)
	}

	return hasRegions, nil
}

// paintRectPaletted fills a rectangle (x,y,width,height) in img with the given palette index.
// Coordinates are assumed to match the ALTO page coordinate system (origin top-left).
func paintRectPaletted(img *image.Paletted, x, y, w, h float64, idx uint8) {
	maxX := x + w
	maxY := y + h

	// clip to image bounds to be safe
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if int(maxX) > img.Bounds().Max.X {
		maxX = float64(img.Bounds().Max.X)
	}
	if int(maxY) > img.Bounds().Max.Y {
		maxY = float64(img.Bounds().Max.Y)
	}

	for yy := y; yy < maxY; yy++ {
		for xx := x; xx < maxX; xx++ {
			img.SetColorIndex(int(xx), int(yy), idx)
		}
	}
}
