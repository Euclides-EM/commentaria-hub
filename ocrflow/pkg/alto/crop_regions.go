package alto

import (
	"fmt"
	"image"
	"math"
)

type CropRegion struct {
	Index int
	Rect  image.Rectangle
}

func ExtractCropRegionsByCategory(a *Alto, category string, bounds image.Rectangle) ([]CropRegion, error) {
	if a == nil {
		return nil, fmt.Errorf("alto is nil")
	}

	blocks, err := ExtractBlocksByCategory(a, category)
	if err != nil {
		return nil, err
	}

	regions := make([]CropRegion, 0, len(blocks))
	for i, block := range blocks {
		rect, ok := blockCropRect(block, bounds)
		if !ok {
			continue
		}
		regions = append(regions, CropRegion{
			Index: i + 1,
			Rect:  rect,
		})
	}

	return regions, nil
}

func blockCropRect(block *TextBlock, bounds image.Rectangle) (image.Rectangle, bool) {
	if block == nil || block.Width <= 0 || block.Height <= 0 {
		return image.Rectangle{}, false
	}

	minX := int(math.Floor(block.HPOS))
	minY := int(math.Floor(block.VPOS))
	maxX := int(math.Ceil(block.HPOS + block.Width))
	maxY := int(math.Ceil(block.VPOS + block.Height))

	rect := image.Rect(minX, minY, maxX, maxY).Intersect(bounds)
	if rect.Empty() {
		return image.Rectangle{}, false
	}

	return rect, true
}
