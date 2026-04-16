package alto

import (
	"fmt"
	"image"
	"math"
	"strings"
)

type CropRegion struct {
	Index int
	Label string
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

func ExtractCropRegionsByCategoryPrefix(a *Alto, categoryPrefix string, bounds image.Rectangle) ([]CropRegion, error) {
	if a == nil {
		return nil, fmt.Errorf("alto is nil")
	}

	labelByID := findTagLabelsByPrefix(a, categoryPrefix)
	regions := make([]CropRegion, 0)
	running := make(map[string]int)

	for _, page := range a.Layout.Page {
		for _, block := range page.PrintSpace.TextBlocks {
			label, ok := firstMatchingLabel(block.TagRefs, labelByID)
			if !ok {
				continue
			}
			rect, ok := blockCropRect(&block, bounds)
			if !ok {
				continue
			}
			running[label]++
			regions = append(regions, CropRegion{
				Index: running[label],
				Label: label,
				Rect:  rect,
			})
		}
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

func findTagLabelsByPrefix(a *Alto, prefix string) map[string]string {
	labels := make(map[string]string)
	for _, t := range a.Tags.OtherTags {
		if t.ID == "" || !strings.HasPrefix(t.Label, prefix) {
			continue
		}
		labels[t.ID] = strings.TrimPrefix(t.Label, prefix)
	}
	return labels
}

func firstMatchingLabel(tagrefs string, labelsByID map[string]string) (string, bool) {
	if tagrefs == "" || len(labelsByID) == 0 {
		return "", false
	}
	for _, tok := range strings.Fields(tagrefs) {
		if label, ok := labelsByID[tok]; ok {
			return label, true
		}
	}
	return "", false
}
